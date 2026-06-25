package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

func listAccountsHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"items": []any{}}) }
}

func createAccountHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusCreated, gin.H{"message": "not implemented"}) }
}

func updateAccountHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "not implemented"}) }
}

func deleteAccountHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "not implemented"}) }
}
