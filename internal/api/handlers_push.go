package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

func subscribePushHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusCreated, gin.H{"message": "not implemented"}) }
}

func vapidPublicKeyHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"public_key": cfg.VapidPublicKey})
	}
}
