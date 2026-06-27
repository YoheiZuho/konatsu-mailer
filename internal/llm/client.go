// Package llm integrates with OpenAI-compatible Chat Completions endpoints for
// email classification, summarization, and reply drafting (design §5).
package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// Config describes a resolved LLM connection (decrypted API key).
type Config struct {
	BaseURL            string
	APIKey             string
	Model              string
	Temperature        float32
	MaxTokens          int
	Timeout            time.Duration
	SupportsJSONSchema bool
}

// ClassifyInput holds email metadata for analysis.
type ClassifyInput struct {
	Subject    string
	Body       string
	Sender     string
	Candidates []string // allowed label names
}

// ClassifyResult is the structured analysis output.
type ClassifyResult struct {
	Summary  string   `json:"summary"`
	Priority int      `json:"priority"`
	Labels   []string `json:"labels"`
	IsSpam   bool     `json:"is_spam"`
}

// DraftInput is the context for generating a reply/compose draft.
type DraftInput struct {
	ThreadText string
	UserHint   string
}

// Provider talks to one OpenAI-compatible endpoint.
type Provider struct {
	client *openai.Client
	cfg    Config
}

func NewProvider(cfg Config) *Provider {
	oc := openai.DefaultConfig(cfg.APIKey) // empty key is fine for local servers
	if cfg.BaseURL != "" {
		oc.BaseURL = cfg.BaseURL
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	oc.HTTPClient = &http.Client{Timeout: timeout}
	return &Provider{client: openai.NewClientWithConfig(oc), cfg: cfg}
}

const systemPrompt = `あなたは高度なメール秘書です。入力メールを解析し、必ず次のJSON形式のみで返答してください。
{"summary": string, "priority": number, "labels": string[], "is_spam": boolean}
- summary: 日本語1文（80字以内）の要約
- priority: 重要度 1(低)〜5(緊急)
- labels: 提示された候補ラベルから0個以上を厳密に選択（候補外は禁止）
- is_spam: 明確な迷惑メールなら true
説明文・コードブロック・前置きは一切出力しないこと。`

// Classify runs analysis and returns validated, structured results.
func (p *Provider) Classify(ctx context.Context, in ClassifyInput) (ClassifyResult, error) {
	candidates, _ := json.Marshal(in.Candidates)
	userPayload, _ := json.Marshal(map[string]any{
		"sender":            in.Sender,
		"subject":           in.Subject,
		"body":              truncate(in.Body, 4000),
		"labels_candidates": json.RawMessage(candidates),
	})

	req := openai.ChatCompletionRequest{
		Model:       p.cfg.Model,
		Temperature: p.cfg.Temperature,
		MaxTokens:   p.cfg.MaxTokens,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: string(userPayload)},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	}

	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		// Retry once without response_format (some endpoints reject it).
		req.ResponseFormat = nil
		resp, err = p.client.CreateChatCompletion(ctx, req)
		if err != nil {
			return ClassifyResult{}, fmt.Errorf("chat completion: %w", err)
		}
	}
	if len(resp.Choices) == 0 {
		return ClassifyResult{}, errors.New("empty completion")
	}

	result, err := parseResult(resp.Choices[0].Message.Content)
	if err != nil {
		return ClassifyResult{}, err
	}
	return validate(result, in.Candidates), nil
}

// Draft streams a reply/compose draft, invoking onChunk for each text delta.
func (p *Provider) Draft(ctx context.Context, in DraftInput, onChunk func(string)) error {
	sys := "あなたは丁寧なビジネスメールの作成を支援するアシスタントです。返信または新規メールの本文のみを日本語で出力してください（件名・署名・前置きの説明は不要）。"
	user := strings.TrimSpace(fmt.Sprintf("以下の文脈に基づいて本文を作成してください。\n\n# 指示\n%s\n\n# 文脈\n%s",
		in.UserHint, truncate(in.ThreadText, 4000)))

	stream, err := p.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       p.cfg.Model,
		Temperature: 0.5,
		MaxTokens:   p.cfg.MaxTokens,
		Stream:      true,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: sys},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
	})
	if err != nil {
		return fmt.Errorf("stream: %w", err)
	}
	defer stream.Close()

	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		if len(resp.Choices) > 0 {
			if delta := resp.Choices[0].Delta.Content; delta != "" {
				onChunk(delta)
			}
		}
	}
}

// TestConnection verifies the endpoint is reachable, returning model ids.
func (p *Provider) TestConnection(ctx context.Context) ([]string, error) {
	models, err := p.client.ListModels(ctx)
	if err == nil {
		ids := make([]string, 0, len(models.Models))
		for _, m := range models.Models {
			ids = append(ids, m.ID)
		}
		return ids, nil
	}
	_, cerr := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:     p.cfg.Model,
		MaxTokens: 1,
		Messages:  []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "ping"}},
	})
	if cerr != nil {
		return nil, cerr
	}
	return []string{p.cfg.Model}, nil
}

var jsonObjectRe = regexp.MustCompile(`(?s)\{.*\}`)

func parseResult(content string) (ClassifyResult, error) {
	var r ClassifyResult
	if err := json.Unmarshal([]byte(content), &r); err == nil {
		return r, nil
	}
	if m := jsonObjectRe.FindString(content); m != "" {
		if err := json.Unmarshal([]byte(m), &r); err == nil {
			return r, nil
		}
	}
	return ClassifyResult{}, fmt.Errorf("could not parse LLM JSON output")
}

// validate clamps priority and whitelists labels against the candidates.
func validate(r ClassifyResult, candidates []string) ClassifyResult {
	if r.Priority < 1 {
		r.Priority = 1
	}
	if r.Priority > 5 {
		r.Priority = 5
	}
	allowed := make(map[string]bool, len(candidates))
	for _, c := range candidates {
		allowed[c] = true
	}
	filtered := make([]string, 0, len(r.Labels))
	for _, l := range r.Labels {
		if allowed[l] {
			filtered = append(filtered, l)
		}
	}
	r.Labels = filtered
	if len([]rune(r.Summary)) > 200 {
		r.Summary = string([]rune(r.Summary)[:200])
	}
	return r
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}
