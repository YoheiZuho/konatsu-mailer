package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

const llmColumns = `id, user_id, name, base_url, api_key_encrypted, model, temperature,
	max_tokens, supports_json_schema, request_timeout_ms, is_default, is_active, created_at`

func scanLLMConfig(row interface{ Scan(dest ...any) error }) (domain.LLMConfig, error) {
	var c domain.LLMConfig
	err := row.Scan(
		&c.ID, &c.UserID, &c.Name, &c.BaseURL, &c.APIKeyEncrypted, &c.Model, &c.Temperature,
		&c.MaxTokens, &c.SupportsJSONSchema, &c.RequestTimeoutMs, &c.IsDefault, &c.IsActive, &c.CreatedAt,
	)
	return c, err
}

// ListLLMConfigs returns a user's LLM connections (default first).
func (db *DB) ListLLMConfigs(ctx context.Context, userID string) ([]domain.LLMConfig, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT `+llmColumns+` FROM llm_configs WHERE user_id=$1 ORDER BY is_default DESC, created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.LLMConfig{}
	for rows.Next() {
		c, err := scanLLMConfig(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// DefaultLLMConfig returns the user's default (or first active) connection.
func (db *DB) DefaultLLMConfig(ctx context.Context, userID domain.UUID) (domain.LLMConfig, error) {
	row := db.Pool.QueryRow(ctx,
		`SELECT `+llmColumns+` FROM llm_configs
		 WHERE user_id=$1 AND is_active=true
		 ORDER BY is_default DESC, created_at LIMIT 1`, userID)
	return scanLLMConfig(row)
}

func (db *DB) GetLLMConfig(ctx context.Context, userID, id string) (domain.LLMConfig, error) {
	row := db.Pool.QueryRow(ctx, `SELECT `+llmColumns+` FROM llm_configs WHERE id=$1 AND user_id=$2`, id, userID)
	return scanLLMConfig(row)
}

// CreateLLMConfig inserts a connection; if it becomes the default, others are cleared.
func (db *DB) CreateLLMConfig(ctx context.Context, userID string, c *domain.LLMConfig) (domain.LLMConfig, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return domain.LLMConfig{}, err
	}
	defer tx.Rollback(ctx)

	if c.IsDefault {
		if _, err := tx.Exec(ctx, `UPDATE llm_configs SET is_default=false WHERE user_id=$1`, userID); err != nil {
			return domain.LLMConfig{}, err
		}
	}
	row := tx.QueryRow(ctx,
		`INSERT INTO llm_configs
		   (user_id, name, base_url, api_key_encrypted, model, temperature, max_tokens,
		    supports_json_schema, request_timeout_ms, is_default, is_active)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		 RETURNING `+llmColumns,
		userID, c.Name, c.BaseURL, c.APIKeyEncrypted, c.Model, c.Temperature, c.MaxTokens,
		c.SupportsJSONSchema, c.RequestTimeoutMs, c.IsDefault, c.IsActive)
	created, err := scanLLMConfig(row)
	if err != nil {
		return domain.LLMConfig{}, err
	}
	return created, tx.Commit(ctx)
}

// UpdateLLMConfig replaces a connection. If apiKey is nil the stored key is kept.
func (db *DB) UpdateLLMConfig(ctx context.Context, userID, id string, c *domain.LLMConfig, apiKey []byte) (domain.LLMConfig, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return domain.LLMConfig{}, err
	}
	defer tx.Rollback(ctx)

	if c.IsDefault {
		if _, err := tx.Exec(ctx, `UPDATE llm_configs SET is_default=false WHERE user_id=$1 AND id<>$2`, userID, id); err != nil {
			return domain.LLMConfig{}, err
		}
	}
	row := tx.QueryRow(ctx,
		`UPDATE llm_configs SET
		   name=$3, base_url=$4, model=$5, temperature=$6, max_tokens=$7,
		   supports_json_schema=$8, request_timeout_ms=$9, is_default=$10, is_active=$11,
		   api_key_encrypted = COALESCE($12, api_key_encrypted)
		 WHERE id=$1 AND user_id=$2
		 RETURNING `+llmColumns,
		id, userID, c.Name, c.BaseURL, c.Model, c.Temperature, c.MaxTokens,
		c.SupportsJSONSchema, c.RequestTimeoutMs, c.IsDefault, c.IsActive, apiKey)
	updated, err := scanLLMConfig(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.LLMConfig{}, err
	}
	if err != nil {
		return domain.LLMConfig{}, err
	}
	return updated, tx.Commit(ctx)
}

func (db *DB) DeleteLLMConfig(ctx context.Context, userID, id string) (int64, error) {
	tag, err := db.Pool.Exec(ctx, `DELETE FROM llm_configs WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
