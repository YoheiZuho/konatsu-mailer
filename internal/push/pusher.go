// Package push sends Web Push notifications via VAPID (design §10).
package push

import (
	"context"
	"io"

	webpush "github.com/SherClockHolmes/webpush-go"
)

// Pusher holds the VAPID credentials used to sign push messages.
type Pusher struct {
	publicKey  string
	privateKey string
	subject    string
}

func NewPusher(publicKey, privateKey, subject string) *Pusher {
	return &Pusher{publicKey: publicKey, privateKey: privateKey, subject: subject}
}

// Enabled reports whether VAPID keys are configured.
func (p *Pusher) Enabled() bool {
	return p.publicKey != "" && p.privateKey != ""
}

// Send delivers a payload to one subscription. Returns the HTTP status code so
// the caller can prune gone subscriptions (404/410).
func (p *Pusher) Send(ctx context.Context, endpoint, p256dh, auth string, payload []byte) (int, error) {
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: endpoint,
		Keys:     webpush.Keys{P256dh: p256dh, Auth: auth},
	}, &webpush.Options{
		Subscriber:      p.subject,
		VAPIDPublicKey:  p.publicKey,
		VAPIDPrivateKey: p.privateKey,
		TTL:             30,
	})
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode, nil
}
