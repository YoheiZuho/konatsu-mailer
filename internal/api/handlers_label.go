package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

func listLabelsHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, err := db.ListLabelsForUser(c.Request.Context(), c.GetString("userID"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to list labels"))
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

type labelInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func createLabelHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in labelInput
		if err := c.ShouldBindJSON(&in); err != nil || in.Name == "" {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", "name is required"))
			return
		}
		label, err := db.CreateLabelForUser(c.Request.Context(), c.GetString("userID"), in.Name, in.Color)
		if err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("label_error", "failed to create label (add an account first)"))
			return
		}
		c.JSON(http.StatusCreated, label)
	}
}

func updateLabelHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in labelInput
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}
		label, err := db.UpdateLabelForUser(c.Request.Context(), c.GetString("userID"), c.Param("id"), in.Name, in.Color)
		if err != nil {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "label not found"))
			return
		}
		c.JSON(http.StatusOK, label)
	}
}

func deleteLabelHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		n, err := db.DeleteLabelForUser(c.Request.Context(), c.GetString("userID"), c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to delete"))
			return
		}
		if n == 0 {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "label not found"))
			return
		}
		c.JSON(http.StatusOK, gin.H{"deleted": true})
	}
}
