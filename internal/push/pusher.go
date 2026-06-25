package push

import "context"

// Pusher sends Web Push notifications.
type Pusher struct{}

func NewPusher(publicKey, privateKey, subject string) *Pusher {
	return &Pusher{}
}

func (p *Pusher) Notify(ctx context.Context, userID string, payload []byte) error {
	return nil
}
