package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

// defaultAIFilters: which mail categories are skipped by AI analysis by default
// (true = skip / no AI summary). Promotional and social mail rarely need it.
var defaultAIFilters = map[string]bool{
	"promotions":  true,
	"social":      true,
	"newsletters": true,
	"automated":   true,
}

// preferencesResponse assembles the response from the users row + prefs JSONB.
func preferencesResponse(theme, brand string, prefs map[string]any) gin.H {
	density, _ := prefs["density"].(string)
	if density != "compact" {
		density = "comfortable"
	}
	aiSummaries := true
	if v, ok := prefs["ai_summaries"].(bool); ok {
		aiSummaries = v
	}
	// The stored value may be map[string]any (decoded from JSONB) or
	// map[string]bool (just set in this request); handle both.
	aiFilters := map[string]bool{}
	for key, def := range defaultAIFilters {
		aiFilters[key] = def
		switch m := prefs["ai_filters"].(type) {
		case map[string]any:
			if b, ok := m[key].(bool); ok {
				aiFilters[key] = b
			}
		case map[string]bool:
			if b, ok := m[key]; ok {
				aiFilters[key] = b
			}
		}
	}
	return gin.H{
		"theme":        theme,
		"brand_color":  brand,
		"density":      density,
		"ai_summaries": aiSummaries,
		"ai_filters":   aiFilters,
	}
}

func loadUserPrefs(c *gin.Context, db *store.DB, userID string) (string, string, map[string]any, bool) {
	var theme, brand string
	var raw []byte
	err := db.Pool.QueryRow(c.Request.Context(),
		`SELECT theme, brand_color, prefs FROM users WHERE id=$1`, userID).Scan(&theme, &brand, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		// The token references a user that no longer exists (e.g. DB was reset).
		// Returning 401 makes the client clear its session and re-authenticate.
		c.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse("session_invalid", "please sign in again"))
		return "", "", nil, false
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to load preferences"))
		return "", "", nil, false
	}
	prefs := map[string]any{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &prefs)
	}
	return theme, brand, prefs, true
}

func getPreferencesHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		theme, brand, prefs, ok := loadUserPrefs(c, db, userID)
		if !ok {
			return
		}
		c.JSON(http.StatusOK, preferencesResponse(theme, brand, prefs))
	}
}

type prefsPatch struct {
	Theme       *string          `json:"theme"`
	BrandColor  *string          `json:"brand_color"`
	Density     *string          `json:"density"`
	AISummaries *bool            `json:"ai_summaries"`
	AIFilters   map[string]bool  `json:"ai_filters"`
}

func patchPreferencesHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")

		var p prefsPatch
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}

		theme, brand, prefs, ok := loadUserPrefs(c, db, userID)
		if !ok {
			return
		}

		if p.Theme != nil {
			theme = *p.Theme
		}
		if p.BrandColor != nil {
			brand = *p.BrandColor
		}
		if p.Density != nil {
			prefs["density"] = *p.Density
		}
		if p.AISummaries != nil {
			prefs["ai_summaries"] = *p.AISummaries
		}
		if p.AIFilters != nil {
			prefs["ai_filters"] = p.AIFilters
		}

		newPrefs, _ := json.Marshal(prefs)
		if _, err := db.Pool.Exec(c.Request.Context(),
			`UPDATE users SET theme=$2, brand_color=$3, prefs=$4 WHERE id=$1`,
			userID, theme, brand, newPrefs); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to save preferences"))
			return
		}
		c.JSON(http.StatusOK, preferencesResponse(theme, brand, prefs))
	}
}
