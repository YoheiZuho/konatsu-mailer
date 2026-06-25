package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

func listLLMConfigsHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"items": []any{}}) }
}

func createLLMConfigHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusCreated, gin.H{"message": "not implemented"}) }
}

func updateLLMConfigHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "not implemented"}) }
}

func deleteLLMConfigHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "not implemented"}) }
}

func testLLMConfigHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": false, "error": "not implemented"}) }
}
