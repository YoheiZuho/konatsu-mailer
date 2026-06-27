package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

// BodyAttachment is attachment metadata stored alongside a cached body.
type BodyAttachment struct {
	Filename string `json:"filename"`
	Size     int    `json:"size"`
}

// CachedBody is a message body cached in the DB.
type CachedBody struct {
	Text        string
	HTML        string
	Attachments []BodyAttachment
}

// GetEmailBody returns a cached body and whether it was present.
func (db *DB) GetEmailBody(ctx context.Context, emailID domain.UUID) (CachedBody, bool, error) {
	var b CachedBody
	var raw []byte
	err := db.Pool.QueryRow(ctx,
		`SELECT text, html, attachments FROM email_bodies WHERE email_id=$1`, emailID).
		Scan(&b.Text, &b.HTML, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return CachedBody{}, false, nil
	}
	if err != nil {
		return CachedBody{}, false, err
	}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &b.Attachments)
	}
	return b, true, nil
}

// SaveEmailBody upserts a cached body for an email.
func (db *DB) SaveEmailBody(ctx context.Context, emailID domain.UUID, b CachedBody) error {
	atts, _ := json.Marshal(b.Attachments)
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO email_bodies (email_id, text, html, attachments, fetched_at)
		 VALUES ($1,$2,$3,$4, now())
		 ON CONFLICT (email_id) DO UPDATE
		   SET text=EXCLUDED.text, html=EXCLUDED.html, attachments=EXCLUDED.attachments, fetched_at=now()`,
		emailID, b.Text, b.HTML, atts)
	return err
}
