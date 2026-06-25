package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

func draftHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.Status(http.StatusOK)
		// TODO: implement SSE streaming for AI draft
	}
}
