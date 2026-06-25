package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yoheizuho/konatsu-mailer/internal/config"
)

func wsUpgradeHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: upgrade to WebSocket using coder/websocket
		c.JSON(http.StatusOK, gin.H{"message": "websocket endpoint (not implemented)"})
	}
}
