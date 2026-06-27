package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/imapsync"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
	"github.com/yoheizuho/konatsu-mailer/internal/ws"
)

// NewRouter creates the main Gin router with all handlers.
func NewRouter(cfg *config.Config, db *store.DB, hub *ws.Hub, analyzer Enqueuer, pool *imapsync.Pool) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(requestLogger())

	api := r.Group("/api")
	{
		// Public. Auth endpoints are rate-limited per IP against brute force.
		authRL := newRateLimiter(20, time.Minute).middleware()
		api.GET("/auth/config", authConfigHandler(cfg))
		api.POST("/auth/register", authRL, registerHandler(db, cfg))
		api.POST("/auth/login", authRL, loginHandler(db, cfg))
		api.POST("/auth/refresh", authRL, refreshHandler(cfg))

		// Authenticated
		auth := api.Group("")
		auth.Use(jwtAuthMiddleware(cfg.JWTSecret))
		{
			auth.GET("/folders", listFoldersHandler(db))

			auth.GET("/filters", listFiltersHandler(db))
			auth.POST("/filters", createFilterHandler(db))
			auth.PATCH("/filters/:id", updateFilterHandler(db))
			auth.DELETE("/filters/:id", deleteFilterHandler(db))

			auth.GET("/emails", listEmailsHandler(db))
			auth.GET("/emails/:id", getEmailHandler(db, cfg, pool))
			auth.POST("/emails/send", sendEmailHandler(db, cfg))
			auth.PATCH("/emails/:id/read", patchReadHandler(db, cfg, pool))
			auth.PATCH("/emails/:id/star", patchStarHandler(db))
			auth.POST("/emails/:id/labels", assignLabelsHandler(db))
			auth.PATCH("/emails/:id/category", setCategoryHandler(db))
			auth.POST("/emails/:id/move", moveEmailHandler(db, cfg, pool))
			auth.POST("/emails/:id/reanalyze", reanalyzeHandler(db, analyzer))

			auth.GET("/labels", listLabelsHandler(db))
			auth.POST("/labels", createLabelHandler(db))
			auth.PATCH("/labels/:id", updateLabelHandler(db))
			auth.DELETE("/labels/:id", deleteLabelHandler(db))

			auth.GET("/accounts", listAccountsHandler(db))
			auth.POST("/accounts", createAccountHandler(db, cfg))
			auth.PATCH("/accounts/:id", updateAccountHandler(db, cfg))
			auth.DELETE("/accounts/:id", deleteAccountHandler(db))

			auth.GET("/me/preferences", getPreferencesHandler(db))
			auth.PATCH("/me/preferences", patchPreferencesHandler(db))

			auth.GET("/llm-configs", listLLMConfigsHandler(db))
			auth.POST("/llm-configs", createLLMConfigHandler(db, cfg))
			auth.PATCH("/llm-configs/:id", updateLLMConfigHandler(db, cfg))
			auth.DELETE("/llm-configs/:id", deleteLLMConfigHandler(db))
			auth.POST("/llm-configs/:id/test", testLLMConfigHandler(db, cfg))

			auth.POST("/ai/draft", draftHandler(db, cfg))

			auth.GET("/translate/config", translateConfigHandler(cfg))
			auth.GET("/translate/languages", translateLanguagesHandler(cfg))
			auth.POST("/translate", translateHandler(cfg))

			auth.POST("/push/subscribe", subscribePushHandler(db, cfg))
			auth.GET("/push/vapid-public-key", vapidPublicKeyHandler(cfg))

			auth.GET("/ws", wsUpgradeHandler(cfg, hub))
		}
	}

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
