package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

func subscribePushHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Endpoint  string `json:"endpoint" binding:"required"`
			P256dh    string `json:"p256dh" binding:"required"`
			Auth      string `json:"auth" binding:"required"`
			UserAgent string `json:"user_agent"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}
		if err := db.SavePushSubscription(c.Request.Context(), c.GetString("userID"),
			req.Endpoint, req.P256dh, req.Auth, req.UserAgent); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to save subscription"))
			return
		}
		c.JSON(http.StatusCreated, gin.H{"subscribed": true})
	}
}

func vapidPublicKeyHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"public_key": cfg.VapidPublicKey})
	}
}
