package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/crypto"
	"github.com/yoheizuho/konatsu-mailer/internal/domain"
	"github.com/yoheizuho/konatsu-mailer/internal/imapsync"
	"github.com/yoheizuho/konatsu-mailer/internal/smtpsend"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

// listEmailsHandler serves GET /api/emails with folder/label/q/unread filters
// and keyset pagination by date_sent (design §7.1).
func listEmailsHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")

		limit := 50
		if v, err := strconv.Atoi(c.Query("limit")); err == nil && v > 0 && v <= 100 {
			limit = v
		}

		where := []string{"a.user_id = $1"}
		args := []any{userID}
		addArg := func(condFmt string, val any) {
			args = append(args, val)
			where = append(where, fmt.Sprintf(condFmt, len(args)))
		}

		switch {
		case c.Query("starred") == "true":
			where = append(where, "e.is_starred = true")
		case c.Query("important") == "true":
			where = append(where, "e.ai_priority >= 4")
		default:
			addArg("e.folder = $%d", c.DefaultQuery("folder", "INBOX"))
		}
		if c.Query("unread") == "true" {
			where = append(where, "e.is_read = false")
		}
		if cat := c.Query("category"); cat != "" {
			addArg("e.category = $%d", cat)
		}
		if label := c.Query("label"); label != "" {
			addArg("EXISTS (SELECT 1 FROM email_labels el JOIN labels l ON l.id=el.label_id "+
				"WHERE el.email_id=e.id AND l.name=$%d)", label)
		}
		// When searching, also match the cached full body (email_bodies) via a
		// LEFT JOIN; this covers messages that have been opened (body cached).
		bodyJoin := ""
		if q := strings.TrimSpace(c.Query("q")); q != "" {
			bodyJoin = " LEFT JOIN email_bodies eb ON eb.email_id = e.id"
			args = append(args, q)
			idx := len(args)
			// Full-text (tsvector, word/phrase) OR substring (ILIKE, covers CJK),
			// across subject/sender/preview (all mail) and the cached body (opened).
			where = append(where, fmt.Sprintf(
				"(e.search_tsv @@ websearch_to_tsquery('simple', $%d)"+
					" OR eb.search_tsv @@ websearch_to_tsquery('simple', $%d)"+
					" OR e.subject ILIKE '%%'||$%d||'%%'"+
					" OR e.sender_addr ILIKE '%%'||$%d||'%%'"+
					" OR e.sender_name ILIKE '%%'||$%d||'%%'"+
					" OR e.body_preview ILIKE '%%'||$%d||'%%'"+
					" OR eb.text ILIKE '%%'||$%d||'%%')",
				idx, idx, idx, idx, idx, idx, idx))
		}
		if cursor := c.Query("cursor"); cursor != "" {
			if t, err := time.Parse(time.RFC3339, cursor); err == nil {
				addArg("e.date_sent < $%d", t)
			}
		}

		args = append(args, limit+1)
		query := `SELECT e.id, e.sender_name, e.sender_addr, e.subject, e.body_preview,
			e.ai_summary, e.ai_priority, e.is_read, e.is_starred, e.has_attachment,
			e.date_sent, e.analysis_status, e.thread_id,
			COALESCE((SELECT json_agg(json_build_object('name', l.name, 'color', l.color))
			          FROM email_labels el JOIN labels l ON l.id = el.label_id
			          WHERE el.email_id = e.id), '[]') AS labels
		FROM emails e JOIN accounts a ON a.id = e.account_id` + bodyJoin + `
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY e.date_sent DESC
		LIMIT $` + strconv.Itoa(len(args))

		rows, err := db.Pool.Query(c.Request.Context(), query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to list emails"))
			return
		}
		defer rows.Close()

		items := make([]gin.H, 0, limit)
		var lastDate time.Time
		for rows.Next() {
			var (
				id, senderAddr, analysisStatus  string
				senderName, subject, preview    *string
				aiSummary                       *string
				aiPriority                      *int16
				isRead, isStarred, hasAttachment bool
				dateSent                        time.Time
				threadID                        *string
				labelsRaw                       []byte
			)
			if err := rows.Scan(&id, &senderName, &senderAddr, &subject, &preview,
				&aiSummary, &aiPriority, &isRead, &isStarred, &hasAttachment,
				&dateSent, &analysisStatus, &threadID, &labelsRaw); err != nil {
				c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to read email"))
				return
			}
			lastDate = dateSent
			items = append(items, gin.H{
				"id":              id,
				"sender_name":     deref(senderName),
				"sender_addr":     senderAddr,
				"subject":         deref(subject),
				"body_preview":    deref(preview),
				"ai_summary":      aiSummary,
				"ai_priority":     aiPriority,
				"is_read":         isRead,
				"is_starred":      isStarred,
				"has_attachment":  hasAttachment,
				"date_sent":       dateSent.Format(time.RFC3339),
				"analysis_status": analysisStatus,
				"thread_id":       threadID,
				"labels":          json.RawMessage(labelsRaw),
			})
		}

		var nextCursor any
		if len(items) > limit {
			items = items[:limit]
			nextCursor = lastDate.Format(time.RFC3339)
		}
		c.JSON(http.StatusOK, gin.H{"items": items, "next_cursor": nextCursor})
	}
}

// getEmailHandler serves GET /api/emails/:id, returning the full thread with
// bodies fetched on demand from IMAP (design §7.2).
func getEmailHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		id := c.Param("id")
		ctx := c.Request.Context()

		root, err := db.GetEmailForUser(ctx, id, userID)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "email not found"))
			return
		}

		messages := []store.EmailRecord{root}
		if root.ThreadID != nil {
			if msgs, err := db.ThreadEmailsForUser(ctx, *root.ThreadID, userID); err == nil && len(msgs) > 0 {
				messages = msgs
			}
		}

		bodies := fetchThreadBodies(ctx, db, cfg, messages)

		out := make([]gin.H, 0, len(messages))
		for _, m := range messages {
			body := bodies[m.ID]
			text := body.Text
			if text == "" && body.HTML == "" {
				text = m.BodyPreview // graceful fallback if IMAP fetch failed
			}
			out = append(out, gin.H{
				"id":          m.ID,
				"from":        gin.H{"name": m.SenderName, "addr": m.SenderAddr},
				"to":          recipientsByType(m.Recipients, "to"),
				"cc":          recipientsByType(m.Recipients, "cc"),
				"date":        m.DateSent.Format(time.RFC3339),
				"subject":     m.Subject,
				"text":        text,
				"html":        body.HTML,
				"ai_summary":  m.AISummary,
				"ai_priority": m.AIPriority,
				"is_read":     m.IsRead,
				"attachments": attachmentsJSON(body.Attachments),
			})
		}

		var threadID domain.UUID = root.ID
		if root.ThreadID != nil {
			threadID = *root.ThreadID
		}
		c.JSON(http.StatusOK, gin.H{
			"thread_id":  threadID,
			"subject":    root.Subject,
			"labels":     []any{},
			"is_starred": root.IsStarred,
			"messages":   out,
		})
	}
}

// fetchThreadBodies returns message bodies keyed by email id, reading from the
// DB cache and fetching only cache misses from IMAP (then persisting them).
// Failures degrade silently (the caller falls back to the stored preview).
func fetchThreadBodies(ctx context.Context, db *store.DB, cfg *config.Config, messages []store.EmailRecord) map[domain.UUID]store.CachedBody {
	result := make(map[domain.UUID]store.CachedBody)

	// 1. Serve from the DB cache.
	var misses []store.EmailRecord
	for _, m := range messages {
		if cb, ok, _ := db.GetEmailBody(ctx, m.ID); ok {
			result[m.ID] = cb
		} else {
			misses = append(misses, m)
		}
	}
	if len(misses) == 0 {
		return result
	}

	// 2. Fetch misses from IMAP (grouped by account+folder) and cache them.
	enc, err := crypto.NewAES256GCM(cfg.MasterEncKey)
	if err != nil {
		return result
	}
	type group struct {
		account domain.UUID
		folder  string
	}
	groups := make(map[group][]store.EmailRecord)
	for _, m := range misses {
		g := group{m.AccountID, m.Folder}
		groups[g] = append(groups[g], m)
	}

	for g, recs := range groups {
		account, err := db.GetAccount(ctx, string(g.account))
		if err != nil {
			continue
		}
		password, err := enc.Decrypt(account.PasswordEncrypted)
		if err != nil {
			continue
		}
		uids := make([]int64, len(recs))
		for i, r := range recs {
			uids[i] = r.IMAPUID
		}
		parsed, err := imapsync.FetchBodies(ctx, account, string(password), g.folder, uids)
		if err != nil {
			continue
		}
		for _, r := range recs {
			pb, ok := parsed[r.IMAPUID]
			if !ok || (pb.Text == "" && pb.HTML == "") {
				continue
			}
			cb := store.CachedBody{Text: pb.Text, HTML: pb.HTML, Attachments: toBodyAttachments(pb.Attachments)}
			_ = db.SaveEmailBody(ctx, r.ID, cb) // populate cache for next time
			result[r.ID] = cb
		}
	}
	return result
}

func toBodyAttachments(atts []imapsync.AttachmentInfo) []store.BodyAttachment {
	out := make([]store.BodyAttachment, len(atts))
	for i, a := range atts {
		out[i] = store.BodyAttachment{Filename: a.Filename, Size: a.Size}
	}
	return out
}

func patchReadHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		id := c.Param("id")
		ctx := c.Request.Context()
		var body struct {
			IsRead bool `json:"is_read"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}
		if err := db.SetReadForUser(ctx, id, userID, body.IsRead); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to update"))
			return
		}
		// Propagate the \Seen flag to IMAP in the background (best-effort) so app
		// and server read-state stay in sync.
		propagateSeen(db, cfg, userID, id, body.IsRead)
		c.JSON(http.StatusOK, gin.H{"is_read": body.IsRead})
	}
}

// propagateSeen mirrors a read/unread change to the IMAP server asynchronously.
func propagateSeen(db *store.DB, cfg *config.Config, userID, emailID string, seen bool) {
	rec, err := db.GetEmailForUser(context.Background(), emailID, userID)
	if err != nil {
		return
	}
	account, err := db.GetAccount(context.Background(), string(rec.AccountID))
	if err != nil {
		return
	}
	enc, err := crypto.NewAES256GCM(cfg.MasterEncKey)
	if err != nil {
		return
	}
	password, err := enc.Decrypt(account.PasswordEncrypted)
	if err != nil {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := imapsync.SetSeen(ctx, account, string(password), rec.Folder, rec.IMAPUID, seen); err != nil {
			slog.Warn("imap seen propagation failed", slog.String("email", emailID), slog.Any("error", err))
		}
	}()
}

func patchStarHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		var body struct {
			IsStarred bool `json:"is_starred"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}
		if err := db.SetStarForUser(c.Request.Context(), c.Param("id"), userID, body.IsStarred); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to update"))
			return
		}
		c.JSON(http.StatusOK, gin.H{"is_starred": body.IsStarred})
	}
}

type sendEmailReq struct {
	AccountID string   `json:"account_id"`
	To        []string `json:"to" binding:"required"`
	Cc        []string `json:"cc"`
	Bcc       []string `json:"bcc"`
	Subject   string   `json:"subject"`
	Text      string   `json:"text"`
	HTML      string   `json:"html"`
	InReplyTo string   `json:"in_reply_to"`
	ThreadID  string   `json:"thread_id"`
}

func sendEmailHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		ctx := c.Request.Context()

		var req sendEmailReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}

		var account domain.Account
		var err error
		if req.AccountID != "" {
			account, err = db.GetAccountForUser(ctx, req.AccountID, userID)
		} else {
			account, err = db.FirstActiveAccountForUser(ctx, userID)
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("no_account", "no sending account configured"))
			return
		}

		enc, err := crypto.NewAES256GCM(cfg.MasterEncKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "encryption unavailable"))
			return
		}
		password, err := enc.Decrypt(account.PasswordEncrypted)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to decrypt credentials"))
			return
		}

		err = smtpsend.Send(ctx, smtpsend.Params{
			Host:        account.SMTPHost,
			Port:        account.SMTPPort,
			UseStartTLS: account.SMTPUseStartTLS,
			AuthUser:    account.AuthUser,
			Password:    string(password),
			Message: smtpsend.Message{
				FromAddr:  account.Email,
				To:        req.To,
				Cc:        req.Cc,
				Bcc:       req.Bcc,
				Subject:   req.Subject,
				Text:      req.Text,
				HTML:      req.HTML,
				InReplyTo: req.InReplyTo,
			},
		})
		if err != nil {
			c.JSON(http.StatusBadGateway, errorResponse("send_failed", err.Error()))
			return
		}
		c.JSON(http.StatusOK, gin.H{"sent": true})
	}
}

func assignLabelsHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		ctx := c.Request.Context()
		var body struct {
			Add    []string `json:"add"`
			Remove []string `json:"remove"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}
		rec, err := db.GetEmailForUser(ctx, c.Param("id"), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "email not found"))
			return
		}
		for _, name := range body.Add {
			if name == "" {
				continue
			}
			if labelID, _, e := db.GetOrCreateLabel(ctx, rec.AccountID, name, false); e == nil {
				_ = db.LinkEmailLabel(ctx, rec.ID, labelID, "user")
			}
		}
		for _, name := range body.Remove {
			_ = db.UnlinkEmailLabelByName(ctx, rec.ID, rec.AccountID, name)
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// setCategoryHandler changes an email's inbox category (drag-and-drop / manual).
func setCategoryHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		ctx := c.Request.Context()
		var body struct {
			Category string `json:"category" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}
		rec, err := db.GetEmailForUser(ctx, c.Param("id"), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "email not found"))
			return
		}
		if err := db.SetCategoryByID(ctx, rec.ID, body.Category); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to set category"))
			return
		}
		c.JSON(http.StatusOK, gin.H{"category": body.Category})
	}
}

// moveEmailHandler moves an email to another IMAP folder (drag-and-drop).
func moveEmailHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		ctx := c.Request.Context()
		var body struct {
			Folder string `json:"folder" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}
		rec, err := db.GetEmailForUser(ctx, c.Param("id"), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "email not found"))
			return
		}
		if body.Folder == rec.Folder {
			c.JSON(http.StatusOK, gin.H{"moved": true})
			return
		}
		account, err := db.GetAccountForUser(ctx, string(rec.AccountID), userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("no_account", "account not found"))
			return
		}
		enc, err := crypto.NewAES256GCM(cfg.MasterEncKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "encryption unavailable"))
			return
		}
		password, err := enc.Decrypt(account.PasswordEncrypted)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to decrypt credentials"))
			return
		}
		if err := imapsync.MoveMessage(ctx, account, string(password), rec.Folder, rec.IMAPUID, body.Folder); err != nil {
			c.JSON(http.StatusBadGateway, errorResponse("move_failed", err.Error()))
			return
		}
		// Drop the source row; the target folder's next sync re-adds it.
		_ = db.DeleteEmail(ctx, rec.ID)
		c.JSON(http.StatusOK, gin.H{"moved": true})
	}
}

// Enqueuer queues an email for (re)analysis by the LLM pipeline.
type Enqueuer interface {
	Enqueue(emailID, userID, accountID domain.UUID)
}

func reanalyzeHandler(db *store.DB, analyzer Enqueuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		id := c.Param("id")
		ctx := c.Request.Context()
		rec, err := db.GetEmailForUser(ctx, id, userID)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "email not found"))
			return
		}
		_ = db.SetAnalysisStatus(ctx, rec.ID, "pending")
		if analyzer != nil {
			analyzer.Enqueue(rec.ID, domain.UUID(userID), rec.AccountID)
		}
		c.JSON(http.StatusAccepted, gin.H{"status": "pending"})
	}
}

// --- helpers ---

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func recipientsByType(rs domain.Recipients, typ string) []gin.H {
	out := []gin.H{}
	for _, r := range rs {
		if r.Type == typ {
			out = append(out, gin.H{"name": r.Name, "addr": r.Addr})
		}
	}
	return out
}

func attachmentsJSON(atts []store.BodyAttachment) []gin.H {
	out := make([]gin.H, 0, len(atts))
	for i, a := range atts {
		out = append(out, gin.H{"id": strconv.Itoa(i), "filename": a.Filename, "size": a.Size})
	}
	return out
}
