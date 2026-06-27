package store

import (
	"context"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

// PushSubscription is a stored Web Push endpoint for a user's device.
type PushSubscription struct {
	ID       domain.UUID
	Endpoint string
	P256dh   string
	Auth     string
}

// SavePushSubscription upserts a device's push subscription.
func (db *DB) SavePushSubscription(ctx context.Context, userID, endpoint, p256dh, auth, userAgent string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO push_subscriptions (user_id, endpoint, p256dh, auth, user_agent)
		 VALUES ($1,$2,$3,$4,$5)
		 ON CONFLICT (user_id, endpoint) DO UPDATE
		   SET p256dh=EXCLUDED.p256dh, auth=EXCLUDED.auth, user_agent=EXCLUDED.user_agent`,
		userID, endpoint, p256dh, auth, userAgent)
	return err
}

// ListPushSubscriptions returns all push endpoints for a user.
func (db *DB) ListPushSubscriptions(ctx context.Context, userID domain.UUID) ([]PushSubscription, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, endpoint, p256dh, auth FROM push_subscriptions WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PushSubscription{}
	for rows.Next() {
		var s PushSubscription
		if err := rows.Scan(&s.ID, &s.Endpoint, &s.P256dh, &s.Auth); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// DeletePushSubscription removes a subscription (e.g. after a 404/410 from the
// push service).
func (db *DB) DeletePushSubscription(ctx context.Context, endpoint string) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM push_subscriptions WHERE endpoint=$1`, endpoint)
	return err
}
