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
		    has_attachment, date_sent, is_read)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		 ON CONFLICT (account_id, folder, imap_uid) DO UPDATE
		   SET is_read = EXCLUDED.is_read
		 RETURNING id, (xmax = 0) AS inserted`,
		e.AccountID, e.ThreadID, e.Folder, e.IMAPUID, e.MessageID, e.InReplyTo,
		e.Subject, e.SenderName, e.SenderAddr, e.Recipients, e.BodyPreview,
		e.HasAttachment, e.DateSent, e.IsRead,
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
	DateSent    time.Time
	IsRead      bool
	IsStarred   bool
	AISummary   *string
	AIPriority  *int16
}

const emailRecordSelect = `e.id, e.account_id, e.thread_id, e.folder, e.imap_uid,
	e.message_id, e.subject, e.sender_name, e.sender_addr, e.recipients,
	e.body_preview, e.date_sent, e.is_read, e.is_starred, e.ai_summary, e.ai_priority`

func scanEmailRecord(row interface {
	Scan(dest ...any) error
}) (EmailRecord, error) {
	var e EmailRecord
	err := row.Scan(
		&e.ID, &e.AccountID, &e.ThreadID, &e.Folder, &e.IMAPUID,
		&e.MessageID, &e.Subject, &e.SenderName, &e.SenderAddr, &e.Recipients,
		&e.BodyPreview, &e.DateSent, &e.IsRead, &e.IsStarred, &e.AISummary, &e.AIPriority,
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
