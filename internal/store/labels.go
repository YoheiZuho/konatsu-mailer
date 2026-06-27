package store

import (
	"context"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

// Label is a label as returned to the API/UI.
type Label struct {
	ID       domain.UUID `json:"id"`
	Name     string      `json:"name"`
	Color    string      `json:"color"`
	IsSystem bool        `json:"is_system"`
}

// ListLabelsForUser returns labels across the user's accounts (deduped by name).
func (db *DB) ListLabelsForUser(ctx context.Context, userID string) ([]Label, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT DISTINCT ON (l.name) l.id, l.name, l.color, l.is_system
		 FROM labels l JOIN accounts a ON a.id = l.account_id
		 WHERE a.user_id = $1
		 ORDER BY l.name, l.is_system`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Label{}
	for rows.Next() {
		var l Label
		if err := rows.Scan(&l.ID, &l.Name, &l.Color, &l.IsSystem); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// LabelNamesForAccount returns the label names defined on an account (used as
// LLM classification candidates).
func (db *DB) LabelNamesForAccount(ctx context.Context, accountID domain.UUID) ([]string, error) {
	rows, err := db.Pool.Query(ctx, `SELECT name FROM labels WHERE account_id=$1`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// firstAccountID returns the user's first account id (for label creation).
func (db *DB) firstAccountID(ctx context.Context, userID string) (domain.UUID, error) {
	var id domain.UUID
	err := db.Pool.QueryRow(ctx,
		`SELECT id FROM accounts WHERE user_id=$1 ORDER BY created_at LIMIT 1`, userID).Scan(&id)
	return id, err
}

// CreateLabelForUser creates a user label on their first account.
func (db *DB) CreateLabelForUser(ctx context.Context, userID, name, color string) (Label, error) {
	acc, err := db.firstAccountID(ctx, userID)
	if err != nil {
		return Label{}, err
	}
	var l Label
	if color == "" {
		color = "oklch(0.55 0.13 255)"
	}
	err = db.Pool.QueryRow(ctx,
		`INSERT INTO labels (account_id, name, color, is_system) VALUES ($1,$2,$3,false)
		 ON CONFLICT (account_id, name) DO UPDATE SET color=EXCLUDED.color
		 RETURNING id, name, color, is_system`, acc, name, color).Scan(&l.ID, &l.Name, &l.Color, &l.IsSystem)
	return l, err
}

// UpdateLabelForUser updates a label owned by the user.
func (db *DB) UpdateLabelForUser(ctx context.Context, userID, id, name, color string) (Label, error) {
	var l Label
	err := db.Pool.QueryRow(ctx,
		`UPDATE labels l SET name=COALESCE(NULLIF($3,''), l.name), color=COALESCE(NULLIF($4,''), l.color)
		 FROM accounts a WHERE l.account_id=a.id AND l.id=$1 AND a.user_id=$2
		 RETURNING l.id, l.name, l.color, l.is_system`, id, userID, name, color).
		Scan(&l.ID, &l.Name, &l.Color, &l.IsSystem)
	return l, err
}

// DeleteLabelForUser deletes a label owned by the user.
func (db *DB) DeleteLabelForUser(ctx context.Context, userID, id string) (int64, error) {
	tag, err := db.Pool.Exec(ctx,
		`DELETE FROM labels l USING accounts a WHERE l.account_id=a.id AND l.id=$1 AND a.user_id=$2`,
		id, userID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
