package domain

import "time"

// Account represents an external mail account (IMAP/SMTP).
type Account struct {
	ID              UUID      `db:"id" json:"id"`
	UserID          UUID      `db:"user_id" json:"user_id"`
	Email           string    `db:"email" json:"email"`
	IMAPHost        string    `db:"imap_host" json:"imap_host"`
	IMAPPort        int       `db:"imap_port" json:"imap_port"`
	IMAPUseTLS      bool      `db:"imap_use_tls" json:"imap_use_tls"`
	SMTPHost        string    `db:"smtp_host" json:"smtp_host"`
	SMTPPort        int       `db:"smtp_port" json:"smtp_port"`
	SMTPUseStartTLS bool      `db:"smtp_use_starttls" json:"smtp_use_starttls"`
	AuthUser        string    `db:"auth_user" json:"auth_user"`
	PasswordEncrypted []byte  `db:"password_encrypted" json:"-"`
	SyncState       SyncState `db:"sync_state" json:"sync_state"`
	IsActive        bool      `db:"is_active" json:"is_active"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}
