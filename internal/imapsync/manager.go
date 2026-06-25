// Package imapsync keeps a background connection per mail account, fetching new
// messages into the database. It supports both implicit TLS (port 993) and
// STARTTLS, selected per account via Account.IMAPUseTLS.
package imapsync

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"

	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/crypto"
	"github.com/yoheizuho/konatsu-mailer/internal/domain"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

const (
	defaultFolder  = "INBOX"
	initialFetch   = 50 // newest N messages fetched per account
	pollInterval   = 30 * time.Second
	reconcileEvery = 30 * time.Second
	maxBackoff     = 5 * time.Minute
)

// Broadcaster delivers realtime events to a user's connected clients.
type Broadcaster interface {
	Broadcast(userID string, msg any)
}

// SyncManager supervises one worker goroutine per active account.
type SyncManager struct {
	db  *store.DB
	cfg *config.Config
	hub Broadcaster
}

func NewManager(db *store.DB, cfg *config.Config, hub Broadcaster) *SyncManager {
	return &SyncManager{db: db, cfg: cfg, hub: hub}
}

// Start blocks until ctx is cancelled, periodically reconciling the set of
// running workers with the active accounts in the database (so accounts added
// after startup begin syncing without a restart).
func (m *SyncManager) Start(ctx context.Context) error {
	running := map[domain.UUID]context.CancelFunc{}
	defer func() {
		for _, cancel := range running {
			cancel()
		}
	}()

	reconcile := func() {
		accounts, err := m.db.ActiveAccounts(ctx)
		if err != nil {
			slog.Error("imapsync: list accounts", slog.Any("error", err))
			return
		}
		seen := make(map[domain.UUID]bool, len(accounts))
		for _, a := range accounts {
			seen[a.ID] = true
			if _, ok := running[a.ID]; ok {
				continue
			}
			wctx, cancel := context.WithCancel(ctx)
			running[a.ID] = cancel
			go m.runAccount(wctx, a)
		}
		for id, cancel := range running {
			if !seen[id] {
				cancel()
				delete(running, id)
			}
		}
	}

	reconcile()
	ticker := time.NewTicker(reconcileEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			reconcile()
		}
	}
}

func (m *SyncManager) Stop() {}

// runAccount maintains a session for one account, reconnecting with exponential
// backoff after failures.
func (m *SyncManager) runAccount(ctx context.Context, a domain.Account) {
	backoff := time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		err := m.session(ctx, a)
		if ctx.Err() != nil {
			return
		}
		if err != nil {
			slog.Warn("imapsync: session ended",
				slog.String("account", a.Email), slog.Any("error", err))
			m.broadcastStatus(a, "down")
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff *= 2
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
	}
}

// session connects, logs in, and polls for new mail until ctx is done or an
// error occurs.
func (m *SyncManager) session(ctx context.Context, a domain.Account) error {
	enc, err := crypto.NewAES256GCM(m.cfg.MasterEncKey)
	if err != nil {
		return fmt.Errorf("crypto: %w", err)
	}
	password, err := enc.Decrypt(a.PasswordEncrypted)
	if err != nil {
		return fmt.Errorf("decrypt password: %w", err)
	}

	c, err := dial(a)
	if err != nil {
		m.broadcastStatus(a, "reconnecting")
		return fmt.Errorf("dial: %w", err)
	}
	defer c.Close()

	if err := c.Login(a.AuthUser, string(password)).Wait(); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	m.broadcastStatus(a, "connected")

	// Discover the account's mailboxes (Inbox, Sent, Junk, Trash, ...) and cache
	// them so the UI can show real IMAP folders.
	mailboxes := listMailboxes(c)
	if len(mailboxes) > 0 {
		folders := make([]store.Folder, len(mailboxes))
		for i, mb := range mailboxes {
			folders[i] = store.Folder{Name: mb.Name, Role: mb.Role}
		}
		_ = m.db.SaveAccountFolders(ctx, a.ID, folders)
	}
	syncSet := foldersToSync(mailboxes)

	for {
		for _, folder := range syncSet {
			if ctx.Err() != nil {
				return nil
			}
			if err := m.syncFolder(ctx, c, &a, folder); err != nil {
				// Likely a dropped connection; reconnect from the supervisor.
				return fmt.Errorf("sync %s: %w", folder, err)
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(pollInterval):
		}
	}
}

// dial opens an IMAP connection using implicit TLS or STARTTLS per the account.
func dial(a domain.Account) (*imapclient.Client, error) {
	addr := fmt.Sprintf("%s:%d", a.IMAPHost, a.IMAPPort)
	if a.IMAPUseTLS {
		return imapclient.DialTLS(addr, nil)
	}
	return imapclient.DialStartTLS(addr, nil)
}

// syncFolder fetches the newest messages from a folder and upserts them. The
// ON CONFLICT upsert makes re-fetching the same window idempotent, so new mail
// is picked up on each poll without UID bookkeeping.
func (m *SyncManager) syncFolder(ctx context.Context, c *imapclient.Client, a *domain.Account, folder string) error {
	selectData, err := c.Select(folder, nil).Wait()
	if err != nil {
		return fmt.Errorf("select %s: %w", folder, err)
	}
	if selectData.NumMessages == 0 {
		return nil
	}

	start := uint32(1)
	if selectData.NumMessages > initialFetch {
		start = selectData.NumMessages - initialFetch + 1
	}
	seqSet := imap.SeqSet{}
	seqSet.AddRange(start, selectData.NumMessages)

	section := &imap.FetchItemBodySection{
		Specifier: imap.PartSpecifierText,
		Partial:   &imap.SectionPartial{Offset: 0, Size: 4096},
		Peek:      true,
	}
	opts := &imap.FetchOptions{
		UID:          true,
		Flags:        true,
		Envelope:     true,
		InternalDate: true,
		BodySection:  []*imap.FetchItemBodySection{section},
	}

	msgs, err := c.Fetch(seqSet, opts).Collect()
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}

	var newCount int
	for _, msg := range msgs {
		if ctx.Err() != nil {
			return nil
		}
		preview := previewText(msg.FindBodySection(section))
		email := buildEmail(a, folder, msg, preview)

		threadID, err := m.db.UpsertThread(ctx, a.ID, threadKey(email), email.Subject, email.DateSent)
		if err == nil {
			email.ThreadID = &threadID
		}
		id, inserted, err := m.db.UpsertEmail(ctx, email)
		if err != nil {
			slog.Warn("imapsync: upsert email", slog.Any("error", err))
			continue
		}
		if inserted {
			newCount++
			m.hub.Broadcast(string(a.UserID), event("NEW_MAIL", map[string]any{
				"account_id": a.ID,
				"email_id":   id,
			}))
		}
	}

	// Persist UID validity so a future change can trigger a full re-sync.
	a.SyncState = domain.SyncState{folder: domain.FolderSyncState{UIDValidity: selectData.UIDValidity}}
	_ = m.db.UpdateSyncState(ctx, a.ID, a.SyncState)

	if newCount > 0 {
		slog.Info("imapsync: new messages",
			slog.String("account", a.Email), slog.Int("count", newCount))
	}
	return nil
}

func (m *SyncManager) broadcastStatus(a domain.Account, state string) {
	m.hub.Broadcast(string(a.UserID), event("SYNC_STATUS", map[string]any{
		"account_id": a.ID,
		"state":      state,
	}))
}

func event(t string, payload map[string]any) map[string]any {
	return map[string]any{"type": t, "payload": payload}
}

// buildEmail maps a fetched IMAP message to a domain.Email (metadata only).
func buildEmail(a *domain.Account, folder string, msg *imapclient.FetchMessageBuffer, preview string) *domain.Email {
	env := msg.Envelope
	e := &domain.Email{
		AccountID:      a.ID,
		Folder:         folder,
		IMAPUID:        int64(msg.UID),
		BodyPreview:    preview,
		AnalysisStatus: "pending",
		Recipients:     domain.Recipients{},
		IsRead:         hasFlag(msg.Flags, imap.FlagSeen),
		DateSent:       msg.InternalDate,
	}
	if env != nil {
		e.Subject = cleanUTF8(env.Subject)
		e.MessageID = env.MessageID
		if len(env.InReplyTo) > 0 {
			e.InReplyTo = env.InReplyTo[0]
		}
		if !env.Date.IsZero() {
			e.DateSent = env.Date
		}
		if len(env.From) > 0 {
			e.SenderName = cleanUTF8(env.From[0].Name)
			e.SenderAddr = env.From[0].Addr()
		}
		e.Recipients = mapRecipients(env)
	}
	if e.SenderAddr == "" {
		e.SenderAddr = "unknown@unknown"
	}
	if e.DateSent.IsZero() {
		e.DateSent = time.Now()
	}
	return e
}

// cleanUTF8 strips invalid UTF-8 byte sequences so values are safe to store in
// Postgres text columns. IMAP previews fetched as raw partial bytes may be in a
// non-UTF-8 charset or cut mid-multibyte-character.
func cleanUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strings.ToValidUTF8(s, "")
}

func mapRecipients(env *imap.Envelope) domain.Recipients {
	out := domain.Recipients{}
	add := func(addrs []imap.Address, typ string) {
		for _, a := range addrs {
			out = append(out, domain.Recipient{Name: cleanUTF8(a.Name), Addr: a.Addr(), Type: typ})
		}
	}
	add(env.To, "to")
	add(env.Cc, "cc")
	add(env.Bcc, "bcc")
	return out
}

func hasFlag(flags []imap.Flag, want imap.Flag) bool {
	for _, f := range flags {
		if f == want {
			return true
		}
	}
	return false
}

// previewText turns the raw leading bytes of a message body into a short,
// printable preview.
func previewText(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	s := cleanUTF8(string(raw))
	s = strings.Join(strings.Fields(s), " ") // collapse whitespace/newlines
	if utf8.RuneCountInString(s) > 500 {
		s = string([]rune(s)[:500])
	}
	return s
}

// threadKey groups messages: normalized subject within the account, falling
// back to the Message-ID for subject-less mail.
func threadKey(e *domain.Email) string {
	subject := normalizeSubject(e.Subject)
	if subject == "" {
		if e.MessageID != "" {
			return e.MessageID
		}
		return fmt.Sprintf("uid:%d", e.IMAPUID)
	}
	return subject
}

func normalizeSubject(subject string) string {
	s := strings.TrimSpace(subject)
	for {
		lower := strings.ToLower(s)
		switch {
		case strings.HasPrefix(lower, "re:"):
			s = strings.TrimSpace(s[3:])
		case strings.HasPrefix(lower, "fwd:"):
			s = strings.TrimSpace(s[4:])
		case strings.HasPrefix(lower, "fw:"):
			s = strings.TrimSpace(s[3:])
		default:
			return strings.ToLower(strings.Join(strings.Fields(s), " "))
		}
	}
}
