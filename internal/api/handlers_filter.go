package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

type filterReq struct {
	Name       string                   `json:"name"`
	Enabled    *bool                    `json:"enabled"`
	MatchType  string                   `json:"match_type"`
	Conditions []domain.FilterCondition `json:"conditions"`
	Actions    []domain.FilterAction    `json:"actions"`
}

func (r filterReq) toFilter() domain.Filter {
	matchType := r.MatchType
	if matchType != "any" {
		matchType = "all"
	}
	enabled := true
	if r.Enabled != nil {
		enabled = *r.Enabled
	}
	conds := r.Conditions
	if conds == nil {
		conds = []domain.FilterCondition{}
	}
	acts := r.Actions
	if acts == nil {
		acts = []domain.FilterAction{}
	}
	return domain.Filter{
		Name:       r.Name,
		Enabled:    enabled,
		MatchType:  matchType,
		Conditions: conds,
		Actions:    acts,
	}
}

func listFiltersHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		items, err := db.ListFilters(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to list filters"))
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

func createFilterHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		var req filterReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}
		f := req.toFilter()
		created, err := db.CreateFilter(c.Request.Context(), userID, &f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to create filter"))
			return
		}
		c.JSON(http.StatusCreated, created)
	}
}

func updateFilterHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		var req filterReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}
		f := req.toFilter()
		updated, err := db.UpdateFilter(c.Request.Context(), userID, c.Param("id"), &f)
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "filter not found"))
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to update filter"))
			return
		}
		c.JSON(http.StatusOK, updated)
	}
}

func deleteFilterHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		n, err := db.DeleteFilter(c.Request.Context(), userID, c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to delete filter"))
			return
		}
		if n == 0 {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "filter not found"))
			return
		}
		c.JSON(http.StatusOK, gin.H{"deleted": true})
	}
}
