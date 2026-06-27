package store

import (
	"context"
	"encoding/json"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

func scanFilter(row interface{ Scan(dest ...any) error }) (domain.Filter, error) {
	var f domain.Filter
	var conds, acts []byte
	err := row.Scan(&f.ID, &f.UserID, &f.Name, &f.Enabled, &f.MatchType, &conds, &acts, &f.Position)
	if err != nil {
		return f, err
	}
	_ = json.Unmarshal(conds, &f.Conditions)
	_ = json.Unmarshal(acts, &f.Actions)
	if f.Conditions == nil {
		f.Conditions = []domain.FilterCondition{}
	}
	if f.Actions == nil {
		f.Actions = []domain.FilterAction{}
	}
	return f, nil
}

const filterColumns = `id, user_id, name, enabled, match_type, conditions, actions, position`

// ListFilters returns all of a user's filters in evaluation order.
func (db *DB) ListFilters(ctx context.Context, userID string) ([]domain.Filter, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT `+filterColumns+` FROM filters WHERE user_id=$1 ORDER BY position, created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Filter{}
	for rows.Next() {
		f, err := scanFilter(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// EnabledFiltersForUser returns the enabled filters for a user (used by sync).
func (db *DB) EnabledFiltersForUser(ctx context.Context, userID domain.UUID) ([]domain.Filter, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT `+filterColumns+` FROM filters WHERE user_id=$1 AND enabled=true ORDER BY position, created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Filter{}
	for rows.Next() {
		f, err := scanFilter(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// CreateFilter inserts a new filter (appended to the end of the list).
func (db *DB) CreateFilter(ctx context.Context, userID string, f *domain.Filter) (domain.Filter, error) {
	conds, _ := json.Marshal(f.Conditions)
	acts, _ := json.Marshal(f.Actions)
	row := db.Pool.QueryRow(ctx,
		`INSERT INTO filters (user_id, name, enabled, match_type, conditions, actions, position)
		 VALUES ($1,$2,$3,$4,$5,$6,
		         COALESCE((SELECT max(position)+1 FROM filters WHERE user_id=$1), 0))
		 RETURNING `+filterColumns,
		userID, f.Name, f.Enabled, f.MatchType, conds, acts)
	return scanFilter(row)
}

// UpdateFilter replaces a filter owned by the user.
func (db *DB) UpdateFilter(ctx context.Context, userID, id string, f *domain.Filter) (domain.Filter, error) {
	conds, _ := json.Marshal(f.Conditions)
	acts, _ := json.Marshal(f.Actions)
	row := db.Pool.QueryRow(ctx,
		`UPDATE filters SET name=$3, enabled=$4, match_type=$5, conditions=$6, actions=$7
		 WHERE id=$1 AND user_id=$2
		 RETURNING `+filterColumns,
		id, userID, f.Name, f.Enabled, f.MatchType, conds, acts)
	return scanFilter(row)
}

// DeleteFilter removes a filter owned by the user.
func (db *DB) DeleteFilter(ctx context.Context, userID, id string) (int64, error) {
	tag, err := db.Pool.Exec(ctx, `DELETE FROM filters WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
