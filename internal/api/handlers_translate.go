package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/yoheizuho/konatsu-mailer/internal/config"
)

// translateConfigHandler reports whether translation is available and the
// default target language, so the UI can show/hide the translate controls.
func translateConfigHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"enabled":        cfg.LibreTranslateURL != "",
			"default_target": cfg.TranslateDefaultTarget,
		})
	}
}

// translateLanguagesHandler proxies LibreTranslate's GET /languages. Returns an
// empty list (rather than an error) when translation is not configured.
func translateLanguagesHandler(cfg *config.Config) gin.HandlerFunc {
	client := &http.Client{Timeout: 10 * time.Second}
	return func(c *gin.Context) {
		if cfg.LibreTranslateURL == "" {
			c.JSON(http.StatusOK, []any{})
			return
		}
		url := strings.TrimRight(cfg.LibreTranslateURL, "/") + "/languages"
		req, _ := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, url, nil)
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusOK, []any{})
			return
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json", raw)
	}
}

type translateReq struct {
	Q      string `json:"q" binding:"required"`
	Source string `json:"source"` // default "auto"
	Target string `json:"target"` // default cfg.TranslateDefaultTarget
	Format string `json:"format"` // "text" (default) | "html"
}

// translateHandler proxies a translation request to LibreTranslate's POST
// /translate, attaching the configured API key server-side so it is never
// exposed to the browser.
func translateHandler(cfg *config.Config) gin.HandlerFunc {
	client := &http.Client{Timeout: 20 * time.Second}
	return func(c *gin.Context) {
		if cfg.LibreTranslateURL == "" {
			c.JSON(http.StatusNotImplemented, errorResponse("translation_disabled", "translation is not configured"))
			return
		}

		var req translateReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}

		source := req.Source
		if source == "" {
			source = "auto"
		}
		target := req.Target
		if target == "" {
			target = cfg.TranslateDefaultTarget
		}
		format := req.Format
		if format == "" {
			format = "text"
		}

		payload := map[string]any{"q": req.Q, "source": source, "target": target, "format": format}
		if cfg.LibreTranslateAPIKey != "" {
			payload["api_key"] = cfg.LibreTranslateAPIKey
		}
		body, _ := json.Marshal(payload)

		url := strings.TrimRight(cfg.LibreTranslateURL, "/") + "/translate"
		httpReq, _ := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, url, bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(httpReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, errorResponse("translation_error", err.Error()))
			return
		}
		defer resp.Body.Close()
		raw, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusBadGateway, errorResponse("translation_error",
				fmt.Sprintf("upstream returned %d", resp.StatusCode)))
			return
		}

		var lt struct {
			TranslatedText   string `json:"translatedText"`
			DetectedLanguage *struct {
				Language   string  `json:"language"`
				Confidence float64 `json:"confidence"`
			} `json:"detectedLanguage"`
		}
		if err := json.Unmarshal(raw, &lt); err != nil {
			c.JSON(http.StatusBadGateway, errorResponse("translation_error", "invalid upstream response"))
			return
		}

		out := gin.H{"translated_text": lt.TranslatedText, "target": target}
		if lt.DetectedLanguage != nil {
			out["detected_source"] = lt.DetectedLanguage.Language
		}
		c.JSON(http.StatusOK, out)
	}
}
