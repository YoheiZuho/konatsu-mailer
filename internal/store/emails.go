package store

import (
	"context"
	"time"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

// UpsertThread inserts or updates a thread and returns its id.
func (db *DB) UpsertThread(ctx context.Context, accountID domain.UUID, threadKey, subject string, lastDate time.Time) (domain.UUID, error) {
	var id domain.UUID
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO threads (account_id, thread_key, subject, last_date)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (account_id, thread_key)
		 DO UPDATE SET last_date = GREATEST(threads.last_date, EXCLUDED.last_date),
		               subject = COALESCE(NULLIF(EXCLUDED.subject, ''), threads.subject)
		 RETURNING id`,
		accountID, threadKey, subject, lastDate,
	).Scan(&id)
	return id, err
}

// UpsertEmail inserts a message (keyed by account/folder/UID) or refreshes its
// mutable flags. Returns the row id and whether it was newly inserted.
func (db *DB) UpsertEmail(ctx context.Context, e *domain.Email) (domain.UUID, bool, error) {
	var id domain.UUID
	var inserted bool
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO emails
		   (account_id, thread_id, folder, imap_uid, message_id, in_reply_to,
		    subject, sender_name, sender_addr, recipients, body_preview,
		    has_attachment, date_sent, is_read, category)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		 ON CONFLICT (account_id, folder, imap_uid) DO UPDATE
		   SET is_read = EXCLUDED.is_read,
		       subject = EXCLUDED.subject,
		       sender_name = EXCLUDED.sender_name,
		       sender_addr = EXCLUDED.sender_addr,
		       recipients = EXCLUDED.recipients,
		       body_preview = EXCLUDED.body_preview,
		       has_attachment = EXCLUDED.has_attachment
		 RETURNING id, (xmax = 0) AS inserted`,
		e.AccountID, e.ThreadID, e.Folder, e.IMAPUID, e.MessageID, e.InReplyTo,
		e.Subject, e.SenderName, e.SenderAddr, e.Recipients, e.BodyPreview,
		e.HasAttachment, e.DateSent, e.IsRead, categoryOrDefault(e.Category),
	).Scan(&id, &inserted)
	return id, inserted, err
}

// EmailRecord is the subset of an email needed to serve detail views and fetch
// bodies on demand.
type EmailRecord struct {
	ID         domain.UUID
	AccountID  domain.UUID
	ThreadID   *domain.UUID
	Folder     string
	IMAPUID    int64
	MessageID  string
	Subject     string
	SenderName  string
	SenderAddr  string
	Recipients  domain.Recipients
	BodyPreview string
	Category    string
	DateSent    time.Time
	IsRead      bool
	IsStarred   bool
	AISummary   *string
	AIPriority  *int16
}

const emailRecordSelect = `e.id, e.account_id, e.thread_id, e.folder, e.imap_uid,
	e.message_id, e.subject, e.sender_name, e.sender_addr, e.recipients,
	e.body_preview, e.category, e.date_sent, e.is_read, e.is_starred, e.ai_summary, e.ai_priority`

func scanEmailRecord(row interface {
	Scan(dest ...any) error
}) (EmailRecord, error) {
	var e EmailRecord
	err := row.Scan(
		&e.ID, &e.AccountID, &e.ThreadID, &e.Folder, &e.IMAPUID,
		&e.MessageID, &e.Subject, &e.SenderName, &e.SenderAddr, &e.Recipients,
		&e.BodyPreview, &e.Category, &e.DateSent, &e.IsRead, &e.IsStarred, &e.AISummary, &e.AIPriority,
	)
	return e, err
}

// GetEmailForUser loads a single email owned by the user (via its account).
func (db *DB) GetEmailForUser(ctx context.Context, id, userID string) (EmailRecord, error) {
	row := db.Pool.QueryRow(ctx,
		`SELECT `+emailRecordSelect+`
		 FROM emails e JOIN accounts a ON a.id = e.account_id
		 WHERE e.id = $1 AND a.user_id = $2`, id, userID)
	return scanEmailRecord(row)
}

// ThreadEmailsForUser returns every message in a thread, oldest first.
func (db *DB) ThreadEmailsForUser(ctx context.Context, threadID domain.UUID, userID string) ([]EmailRecord, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT `+emailRecordSelect+`
		 FROM emails e JOIN accounts a ON a.id = e.account_id
		 WHERE e.thread_id = $1 AND a.user_id = $2
		 ORDER BY e.date_sent ASC`, threadID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EmailRecord
	for rows.Next() {
		e, err := scanEmailRecord(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func categoryOrDefault(c string) string {
	if c == "" {
		return "primary"
	}
	return c
}

// DeleteEmail removes an email row (used after an IMAP move re-homes it).
func (db *DB) DeleteEmail(ctx context.Context, id domain.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM emails WHERE id=$1`, id)
	return err
}

// SetReadByID / SetStarredByID / SetCategoryByID mutate a single email by id
// (used by the filter engine during sync, where there is no user context).
func (db *DB) SetReadByID(ctx context.Context, id domain.UUID, read bool) error {
	_, err := db.Pool.Exec(ctx, `UPDATE emails SET is_read=$2 WHERE id=$1`, id, read)
	return err
}
func (db *DB) SetStarredByID(ctx context.Context, id domain.UUID, starred bool) error {
	_, err := db.Pool.Exec(ctx, `UPDATE emails SET is_starred=$2 WHERE id=$1`, id, starred)
	return err
}
func (db *DB) SetCategoryByID(ctx context.Context, id domain.UUID, category string) error {
	_, err := db.Pool.Exec(ctx, `UPDATE emails SET category=$2 WHERE id=$1`, id, category)
	return err
}

// GetOrCreateLabel returns the id and color of a label by name within an
// account, creating it if necessary. isSystem marks AI-assigned labels.
func (db *DB) GetOrCreateLabel(ctx context.Context, accountID domain.UUID, name string, isSystem bool) (domain.UUID, string, error) {
	var id domain.UUID
	var color string
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO labels (account_id, name, is_system) VALUES ($1,$2,$3)
		 ON CONFLICT (account_id, name) DO UPDATE SET name=EXCLUDED.name
		 RETURNING id, color`, accountID, name, isSystem).Scan(&id, &color)
	return id, color, err
}

// UpdateEmailAnalysis stores LLM results on an email.
func (db *DB) UpdateEmailAnalysis(ctx context.Context, id domain.UUID, summary string, priority int, status string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE emails SET ai_summary=$2, ai_priority=$3, analysis_status=$4 WHERE id=$1`,
		id, summary, priority, status)
	return err
}

// SetAnalysisStatus updates only the analysis status (e.g. skipped/error).
func (db *DB) SetAnalysisStatus(ctx context.Context, id domain.UUID, status string) error {
	_, err := db.Pool.Exec(ctx, `UPDATE emails SET analysis_status=$2 WHERE id=$1`, id, status)
	return err
}

// LinkEmailLabel attaches a label to an email (idempotent).
func (db *DB) LinkEmailLabel(ctx context.Context, emailID, labelID domain.UUID, source string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO email_labels (email_id, label_id, source) VALUES ($1,$2,$3)
		 ON CONFLICT (email_id, label_id) DO NOTHING`, emailID, labelID, source)
	return err
}

// UnreadCounts returns unread message counts per folder for a user, plus the
// unread counts for the virtual starred and important views.
func (db *DB) UnreadCounts(ctx context.Context, userID string) (perFolder map[string]int, starred, important int, err error) {
	perFolder = map[string]int{}
	rows, err := db.Pool.Query(ctx,
		`SELECT e.folder, count(*)
		 FROM emails e JOIN accounts a ON a.id = e.account_id
		 WHERE a.user_id = $1 AND e.is_read = false
		 GROUP BY e.folder`, userID)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var folder string
		var n int
		if err := rows.Scan(&folder, &n); err != nil {
			return nil, 0, 0, err
		}
		perFolder[folder] = n
	}
	if err := rows.Err(); err != nil {
		return nil, 0, 0, err
	}

	_ = db.Pool.QueryRow(ctx,
		`SELECT count(*) FROM emails e JOIN accounts a ON a.id = e.account_id
		 WHERE a.user_id = $1 AND e.is_read = false AND e.is_starred = true`, userID).Scan(&starred)
	_ = db.Pool.QueryRow(ctx,
		`SELECT count(*) FROM emails e JOIN accounts a ON a.id = e.account_id
		 WHERE a.user_id = $1 AND e.is_read = false AND e.ai_priority >= 4`, userID).Scan(&important)
	return perFolder, starred, important, nil
}

// SetReadForUser toggles the read flag for an owned email.
func (db *DB) SetReadForUser(ctx context.Context, id, userID string, read bool) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE emails e SET is_read=$3
		 FROM accounts a WHERE e.account_id=a.id AND e.id=$1 AND a.user_id=$2`,
		id, userID, read)
	return err
}

// SetStarForUser toggles the starred flag for an owned email.
func (db *DB) SetStarForUser(ctx context.Context, id, userID string, starred bool) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE emails e SET is_starred=$3
		 FROM accounts a WHERE e.account_id=a.id AND e.id=$1 AND a.user_id=$2`,
		id, userID, starred)
	return err
}
