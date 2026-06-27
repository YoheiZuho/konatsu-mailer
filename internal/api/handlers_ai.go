package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/domain"
	"github.com/yoheizuho/konatsu-mailer/internal/llm"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

// draftHandler streams an AI-generated reply/compose draft as SSE (design §5.7).
func draftHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		ctx := c.Request.Context()

		var req struct {
			Mode        string `json:"mode"`
			ThreadID    string `json:"thread_id"`
			EmailID     string `json:"email_id"`
			Instruction string `json:"instruction"`
			Context     string `json:"context"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}

		conf, err := db.DefaultLLMConfig(ctx, domain.UUID(userID))
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("no_llm", "AI接続が設定されていません。設定→AI接続から追加してください。"))
			return
		}
		apiKey := ""
		if len(conf.APIKeyEncrypted) > 0 {
			if dec, derr := decryptKey(cfg, conf.APIKeyEncrypted); derr == nil {
				apiKey = dec
			}
		}

		threadText := req.Context
		if req.EmailID != "" {
			if rec, e := db.GetEmailForUser(ctx, req.EmailID, userID); e == nil {
				threadText = rec.Subject + "\n\n" + rec.BodyPreview
			}
		}

		provider := llm.NewProvider(llm.Config{
			BaseURL: conf.BaseURL, APIKey: apiKey, Model: conf.Model,
			Temperature: conf.Temperature, MaxTokens: conf.MaxTokens,
			Timeout: time.Duration(conf.RequestTimeoutMs) * time.Millisecond,
		})

		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		write := func(text string) {
			b, _ := json.Marshal(map[string]string{"text": text})
			_, _ = c.Writer.Write([]byte("data: " + string(b) + "\n\n"))
			c.Writer.Flush()
		}

		if err := provider.Draft(ctx, llm.DraftInput{ThreadText: threadText, UserHint: req.Instruction}, write); err != nil {
			b, _ := json.Marshal(map[string]string{"error": err.Error()})
			_, _ = c.Writer.Write([]byte("data: " + string(b) + "\n\n"))
		}
		_, _ = c.Writer.Write([]byte("data: [DONE]\n\n"))
		c.Writer.Flush()
	}
}
