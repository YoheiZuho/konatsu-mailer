package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

func listEmailsHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		_ = userID
		// TODO: implement actual query with folder/label/q/unread/pagination
		c.JSON(http.StatusOK, gin.H{
			"items":      []any{},
			"next_cursor": nil,
		})
	}
}

func getEmailHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
	}
}

func sendEmailHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
	}
}

func patchReadHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
	}
}

func patchStarHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
	}
}

func assignLabelsHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "not implemented"})
	}
}

func reanalyzeHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusAccepted, gin.H{"message": "not implemented"})
	}
}
