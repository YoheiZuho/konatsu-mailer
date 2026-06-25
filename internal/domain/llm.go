package domain

import "time"

// LLMConfig stores an OpenAI-compatible endpoint configuration.
type LLMConfig struct {
	ID                 UUID      `db:"id" json:"id"`
	UserID             UUID      `db:"user_id" json:"user_id"`
	Name               string    `db:"name" json:"name"`
	BaseURL            string    `db:"base_url" json:"base_url"`
	APIKeyEncrypted    []byte    `db:"api_key_encrypted" json:"-"`
	Model              string    `db:"model" json:"model"`
	Temperature        float32   `db:"temperature" json:"temperature"`
	MaxTokens          int       `db:"max_tokens" json:"max_tokens"`
	SupportsJSONSchema bool      `db:"supports_json_schema" json:"supports_json_schema"`
	RequestTimeoutMs   int       `db:"request_timeout_ms" json:"request_timeout_ms"`
	IsDefault          bool      `db:"is_default" json:"is_default"`
	IsActive           bool      `db:"is_active" json:"is_active"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
}
