package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/crypto"
	"github.com/yoheizuho/konatsu-mailer/internal/domain"
	"github.com/yoheizuho/konatsu-mailer/internal/llm"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

type llmConfigResp struct {
	ID                 domain.UUID `json:"id"`
	Name               string      `json:"name"`
	BaseURL            string      `json:"base_url"`
	Model              string      `json:"model"`
	Temperature        float32     `json:"temperature"`
	MaxTokens          int         `json:"max_tokens"`
	SupportsJSONSchema bool        `json:"supports_json_schema"`
	RequestTimeoutMs   int         `json:"request_timeout_ms"`
	IsDefault          bool        `json:"is_default"`
	IsActive           bool        `json:"is_active"`
	HasAPIKey          bool        `json:"has_api_key"`
}

func toLLMResp(c domain.LLMConfig) llmConfigResp {
	return llmConfigResp{
		ID: c.ID, Name: c.Name, BaseURL: c.BaseURL, Model: c.Model,
		Temperature: c.Temperature, MaxTokens: c.MaxTokens,
		SupportsJSONSchema: c.SupportsJSONSchema, RequestTimeoutMs: c.RequestTimeoutMs,
		IsDefault: c.IsDefault, IsActive: c.IsActive, HasAPIKey: len(c.APIKeyEncrypted) > 0,
	}
}

type llmConfigInput struct {
	Name               string  `json:"name"`
	BaseURL            string  `json:"base_url"`
	Model              string  `json:"model"`
	APIKey             string  `json:"api_key"`
	Temperature        float32 `json:"temperature"`
	MaxTokens          int     `json:"max_tokens"`
	SupportsJSONSchema bool    `json:"supports_json_schema"`
	RequestTimeoutMs   int     `json:"request_timeout_ms"`
	IsDefault          bool    `json:"is_default"`
	IsActive           bool    `json:"is_active"`
}

func (in llmConfigInput) toDomain() domain.LLMConfig {
	c := domain.LLMConfig{
		Name: in.Name, BaseURL: in.BaseURL, Model: in.Model, Temperature: in.Temperature,
		MaxTokens: in.MaxTokens, SupportsJSONSchema: in.SupportsJSONSchema,
		RequestTimeoutMs: in.RequestTimeoutMs, IsDefault: in.IsDefault, IsActive: in.IsActive,
	}
	if c.BaseURL == "" {
		c.BaseURL = "https://api.openai.com/v1"
	}
	if c.MaxTokens == 0 {
		c.MaxTokens = 512
	}
	if c.RequestTimeoutMs == 0 {
		c.RequestTimeoutMs = 30000
	}
	return c
}

func listLLMConfigsHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := db.ListLLMConfigs(c.Request.Context(), c.GetString("userID"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to list"))
			return
		}
		out := make([]llmConfigResp, len(items))
		for i, it := range items {
			out[i] = toLLMResp(it)
		}
		c.JSON(http.StatusOK, gin.H{"items": out})
	}
}

func createLLMConfigHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in llmConfigInput
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}
		dc := in.toDomain()
		if err := validateLLMBaseURL(dc.BaseURL, cfg.LLMAllowPrivateHosts); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("invalid_base_url", err.Error()))
			return
		}
		if in.APIKey != "" {
			enc, err := encryptKey(cfg, in.APIKey)
			if err != nil {
				c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "encryption unavailable"))
				return
			}
			dc.APIKeyEncrypted = enc
		}
		created, err := db.CreateLLMConfig(c.Request.Context(), c.GetString("userID"), &dc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to create"))
			return
		}
		c.JSON(http.StatusCreated, toLLMResp(created))
	}
}

func updateLLMConfigHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in llmConfigInput
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}
		dc := in.toDomain()
		if err := validateLLMBaseURL(dc.BaseURL, cfg.LLMAllowPrivateHosts); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("invalid_base_url", err.Error()))
			return
		}
		var apiKey []byte // nil keeps the stored key
		if in.APIKey != "" {
			enc, err := encryptKey(cfg, in.APIKey)
			if err != nil {
				c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "encryption unavailable"))
				return
			}
			apiKey = enc
		}
		updated, err := db.UpdateLLMConfig(c.Request.Context(), c.GetString("userID"), c.Param("id"), &dc, apiKey)
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "config not found"))
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to update"))
			return
		}
		c.JSON(http.StatusOK, toLLMResp(updated))
	}
}

func deleteLLMConfigHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		n, err := db.DeleteLLMConfig(c.Request.Context(), c.GetString("userID"), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to delete"))
			return
		}
		if n == 0 {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "config not found"))
			return
		}
		c.JSON(http.StatusOK, gin.H{"deleted": true})
	}
}

func testLLMConfigHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		conf, err := db.GetLLMConfig(c.Request.Context(), c.GetString("userID"), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "config not found"))
			return
		}
		if err := validateLLMBaseURL(conf.BaseURL, cfg.LLMAllowPrivateHosts); err != nil {
			c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
			return
		}
		apiKey := ""
		if len(conf.APIKeyEncrypted) > 0 {
			if dec, derr := decryptKey(cfg, conf.APIKeyEncrypted); derr == nil {
				apiKey = dec
			}
		}
		provider := llm.NewProvider(llm.Config{
			BaseURL: conf.BaseURL, APIKey: apiKey, Model: conf.Model,
			Timeout: time.Duration(conf.RequestTimeoutMs) * time.Millisecond,
		})
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(conf.RequestTimeoutMs+5000)*time.Millisecond)
		defer cancel()
		models, err := provider.TestConnection(ctx)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"ok": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "models": models})
	}
}

func encryptKey(cfg *config.Config, key string) ([]byte, error) {
	enc, err := crypto.NewAES256GCM(cfg.MasterEncKey)
	if err != nil {
		return nil, err
	}
	return enc.Encrypt([]byte(key))
}

func decryptKey(cfg *config.Config, ct []byte) (string, error) {
	enc, err := crypto.NewAES256GCM(cfg.MasterEncKey)
	if err != nil {
		return "", err
	}
	pt, err := enc.Decrypt(ct)
	return string(pt), err
}
