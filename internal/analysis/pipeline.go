// Package analysis consumes classification jobs, runs the LLM, persists results,
// and triggers push notifications (design §6).
package analysis

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/crypto"
	"github.com/yoheizuho/konatsu-mailer/internal/domain"
	"github.com/yoheizuho/konatsu-mailer/internal/llm"
	"github.com/yoheizuho/konatsu-mailer/internal/push"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

// Broadcaster delivers realtime events to a user's clients.
type Broadcaster interface {
	Broadcast(userID string, msg any)
}

// Job is a request to analyze one email.
type Job struct {
	EmailID   domain.UUID
	UserID    domain.UUID
	AccountID domain.UUID
}

// Pipeline runs a worker pool over a buffered job channel.
type Pipeline struct {
	db     *store.DB
	cfg    *config.Config
	hub    Broadcaster
	pusher *push.Pusher
	jobs   chan Job
}

// Default classification candidate labels, merged with the account's labels.
var defaultCandidates = []string{
	"仕事", "プライベート", "重要", "請求書", "通知", "ニュース", "プロモーション", "ソーシャル", "採用", "開発",
}

func New(db *store.DB, cfg *config.Config, hub Broadcaster, pusher *push.Pusher) *Pipeline {
	return &Pipeline{db: db, cfg: cfg, hub: hub, pusher: pusher, jobs: make(chan Job, 1000)}
}

// Start launches the worker pool until ctx is cancelled.
func (p *Pipeline) Start(ctx context.Context) {
	workers := p.cfg.LlmWorkers
	if workers < 1 {
		workers = 1
	}
	for i := 0; i < workers; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case job := <-p.jobs:
					p.process(ctx, job)
				}
			}
		}()
	}
}

// Enqueue submits a job (non-blocking; drops if the queue is full).
func (p *Pipeline) Enqueue(emailID, userID, accountID domain.UUID) {
	select {
	case p.jobs <- Job{EmailID: emailID, UserID: userID, AccountID: accountID}:
	default:
		slog.Warn("analysis: job queue full, dropping", slog.String("email", string(emailID)))
	}
}

func (p *Pipeline) process(parent context.Context, job Job) {
	ctx, cancel := context.WithTimeout(parent, 60*time.Second)
	defer cancel()

	rec, err := p.db.GetEmailForUser(ctx, string(job.EmailID), string(job.UserID))
	if err != nil {
		return
	}

	prefs, _ := p.db.GetUserPrefs(ctx, job.UserID)
	if skipCategory(rec.Category, prefs) {
		_ = p.db.SetAnalysisStatus(ctx, job.EmailID, "skipped")
		return
	}

	llmCfg, err := p.db.DefaultLLMConfig(ctx, job.UserID)
	if err != nil {
		// No LLM configured — leave it unanalyzed (not an error state).
		_ = p.db.SetAnalysisStatus(ctx, job.EmailID, "skipped")
		return
	}

	apiKey := ""
	if len(llmCfg.APIKeyEncrypted) > 0 {
		if enc, e := crypto.NewAES256GCM(p.cfg.MasterEncKey); e == nil {
			if pt, e := enc.Decrypt(llmCfg.APIKeyEncrypted); e == nil {
				apiKey = string(pt)
			}
		}
	}

	candidates := mergeCandidates(p.candidateLabels(ctx, job.AccountID))
	provider := llm.NewProvider(llm.Config{
		BaseURL:            llmCfg.BaseURL,
		APIKey:             apiKey,
		Model:              llmCfg.Model,
		Temperature:        llmCfg.Temperature,
		MaxTokens:          llmCfg.MaxTokens,
		Timeout:            time.Duration(llmCfg.RequestTimeoutMs) * time.Millisecond,
		SupportsJSONSchema: llmCfg.SupportsJSONSchema,
	})

	result, err := provider.Classify(ctx, llm.ClassifyInput{
		Subject:    rec.Subject,
		Body:       rec.BodyPreview,
		Sender:     rec.SenderName + " <" + rec.SenderAddr + ">",
		Candidates: candidates,
	})
	if err != nil {
		slog.Warn("analysis: classify failed", slog.Any("error", err))
		_ = p.db.SetAnalysisStatus(ctx, job.EmailID, "error")
		return
	}

	_ = p.db.UpdateEmailAnalysis(ctx, job.EmailID, result.Summary, result.Priority, "done")

	labels := make([]map[string]string, 0, len(result.Labels))
	for _, name := range result.Labels {
		labelID, color, err := p.db.GetOrCreateLabel(ctx, job.AccountID, name, true)
		if err != nil {
			continue
		}
		_ = p.db.LinkEmailLabel(ctx, job.EmailID, labelID, "ai")
		labels = append(labels, map[string]string{"name": name, "color": color})
	}

	p.hub.Broadcast(string(job.UserID), map[string]any{
		"type": "MAIL_ANALYZED",
		"payload": map[string]any{
			"email_id":    job.EmailID,
			"ai_summary":  result.Summary,
			"ai_priority": result.Priority,
			"labels":      labels,
		},
	})

	if result.Priority >= p.cfg.NotifyThreshold {
		p.notify(parent, job.UserID, rec, result, prefs)
	}
}

func (p *Pipeline) candidateLabels(ctx context.Context, accountID domain.UUID) []string {
	names, _ := p.db.LabelNamesForAccount(ctx, accountID)
	return names
}

func mergeCandidates(accountLabels []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, l := range append(append([]string{}, accountLabels...), defaultCandidates...) {
		if l != "" && !seen[l] {
			seen[l] = true
			out = append(out, l)
		}
	}
	return out
}

// skipCategory reports whether AI analysis should be skipped for this email's
// category, per the user's ai_filters preference.
func skipCategory(category string, prefs map[string]any) bool {
	filters, ok := prefs["ai_filters"].(map[string]any)
	if !ok {
		return false
	}
	key := category
	if key == "" || key == "primary" {
		return false
	}
	if v, ok := filters[key].(bool); ok {
		return v
	}
	return false
}

func (p *Pipeline) notify(parent context.Context, userID domain.UUID, rec store.EmailRecord, result llm.ClassifyResult, prefs map[string]any) {
	if p.pusher == nil || !p.pusher.Enabled() {
		return
	}
	// Restrict to selected labels, if the user chose any.
	if wanted := stringSlice(prefs["push_labels"]); len(wanted) > 0 && !intersects(result.Labels, wanted) {
		return
	}

	subs, err := p.db.ListPushSubscriptions(parent, userID)
	if err != nil || len(subs) == 0 {
		return
	}
	sender := rec.SenderName
	if sender == "" {
		sender = rec.SenderAddr
	}
	payload, _ := json.Marshal(map[string]any{
		"title": "[重要] " + sender,
		"body":  result.Summary,
		"data":  map[string]any{"email_id": rec.ID},
	})

	ctx, cancel := context.WithTimeout(parent, 15*time.Second)
	defer cancel()
	for _, s := range subs {
		code, err := p.pusher.Send(ctx, s.Endpoint, s.P256dh, s.Auth, payload)
		if err != nil {
			continue
		}
		if code == http.StatusNotFound || code == http.StatusGone {
			_ = p.db.DeletePushSubscription(ctx, s.Endpoint)
		}
	}
}

func stringSlice(v any) []string {
	switch s := v.(type) {
	case []any:
		out := make([]string, 0, len(s))
		for _, x := range s {
			if str, ok := x.(string); ok {
				out = append(out, str)
			}
		}
		return out
	case []string:
		return s
	}
	return nil
}

func intersects(a, b []string) bool {
	set := make(map[string]bool, len(b))
	for _, x := range b {
		set[x] = true
	}
	for _, x := range a {
		if set[x] {
			return true
		}
	}
	return false
}
