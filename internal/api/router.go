package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
	"github.com/yoheizuho/konatsu-mailer/internal/ws"
)

// NewRouter creates the main Gin router with all handlers.
func NewRouter(cfg *config.Config, db *store.DB, hub *ws.Hub) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())
	r.Use(requestLogger())

	api := r.Group("/api")
	{
		// Public
		api.GET("/auth/config", authConfigHandler(cfg))
		api.POST("/auth/register", registerHandler(db, cfg))
		api.POST("/auth/login", loginHandler(db, cfg))
		api.POST("/auth/refresh", refreshHandler(cfg))

		// Authenticated
		auth := api.Group("")
		auth.Use(jwtAuthMiddleware(cfg.JWTSecret))
		{
			auth.GET("/folders", listFoldersHandler(db))

			auth.GET("/emails", listEmailsHandler(db))
			auth.GET("/emails/:id", getEmailHandler(db, cfg))
			auth.POST("/emails/send", sendEmailHandler(db, cfg))
			auth.PATCH("/emails/:id/read", patchReadHandler(db))
			auth.PATCH("/emails/:id/star", patchStarHandler(db))
			auth.POST("/emails/:id/labels", assignLabelsHandler(db))
			auth.POST("/emails/:id/reanalyze", reanalyzeHandler(db, cfg))

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
