package domain

import "time"

// Thread groups related emails together.
type Thread struct {
	ID         UUID      `db:"id" json:"id"`
	AccountID  UUID      `db:"account_id" json:"account_id"`
	ThreadKey  string    `db:"thread_key" json:"thread_key"`
	Subject    string    `db:"subject" json:"subject"`
	LastDate   time.Time `db:"last_date" json:"last_date"`
}

// Email holds metadata for a single message (body is not kept here per §7.2).
type Email struct {
	ID              UUID      `db:"id" json:"id"`
	AccountID       UUID      `db:"account_id" json:"account_id"`
	ThreadID        *UUID     `db:"thread_id" json:"thread_id,omitempty"`
	Folder          string    `db:"folder" json:"folder"`
	IMAPUID         int64     `db:"imap_uid" json:"imap_uid"`
	MessageID       string    `db:"message_id" json:"message_id"`
	InReplyTo       string    `db:"in_reply_to" json:"in_reply_to"`
	Subject         string    `db:"subject" json:"subject"`
	SenderName      string    `db:"sender_name" json:"sender_name"`
	SenderAddr      string    `db:"sender_addr" json:"sender_addr"`
	Recipients      Recipients `db:"recipients" json:"recipients"`
	BodyPreview     string    `db:"body_preview" json:"body_preview"`
	AISummary       *string   `db:"ai_summary" json:"ai_summary,omitempty"`
	AIPriority      *int16    `db:"ai_priority" json:"ai_priority,omitempty"`
	AnalysisStatus  string    `db:"analysis_status" json:"analysis_status"`
	HasAttachment   bool      `db:"has_attachment" json:"has_attachment"`
	DateSent        time.Time `db:"date_sent" json:"date_sent"`
	IsRead          bool      `db:"is_read" json:"is_read"`
	IsStarred       bool      `db:"is_starred" json:"is_starred"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}
