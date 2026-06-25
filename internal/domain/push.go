package domain

import "time"

// PushSubscription stores a browser push subscription.
type PushSubscription struct {
	ID        UUID      `db:"id" json:"id"`
	UserID    UUID      `db:"user_id" json:"user_id"`
	Endpoint  string    `db:"endpoint" json:"endpoint"`
	P256DH    string    `db:"p256dh" json:"p256dh"`
	Auth      string    `db:"auth" json:"auth"`
	UserAgent string    `db:"user_agent" json:"user_agent"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
