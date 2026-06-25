package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

func getPreferencesHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		_ = userID
		c.JSON(http.StatusOK, gin.H{
			"theme":       "system",
			"brand_color": "#ffd20a",
			"density":     "comfortable",
			"ai_summaries": true,
		})
	}
}

func patchPreferencesHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
	}
}
