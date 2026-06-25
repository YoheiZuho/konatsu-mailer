package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

func listLabelsHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"items": []any{}}) }
}

func createLabelHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusCreated, gin.H{"message": "not implemented"}) }
}

func updateLabelHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "not implemented"}) }
}

func deleteLabelHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "not implemented"}) }
}
