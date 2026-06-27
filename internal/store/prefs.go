package store

import (
	"context"
	"encoding/json"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

// GetUserPrefs returns a user's prefs JSONB as a map (empty if unset).
func (db *DB) GetUserPrefs(ctx context.Context, userID domain.UUID) (map[string]any, error) {
	var raw []byte
	if err := db.Pool.QueryRow(ctx, `SELECT prefs FROM users WHERE id=$1`, userID).Scan(&raw); err != nil {
		return nil, err
	}
	prefs := map[string]any{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &prefs)
	}
	return prefs, nil
}
