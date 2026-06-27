//go:build integration

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/imapsync"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
	"github.com/yoheizuho/konatsu-mailer/internal/ws"
)

func TestAuthFlow_Integration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	abs, _ := filepath.Abs("../../migrations")
	if err := store.Migrate(url, abs); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	db, err := store.New(url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{
		MasterEncKey:      []byte("0123456789abcdef0123456789abcdef"),
		JWTSecret:         "test-secret",
		AllowRegistration: true,
	}
	gin.SetMode(gin.TestMode)
	r := NewRouter(cfg, db, ws.NewHub(), nil, imapsync.NewPool())

	email := "api-it-" + t.Name() + "@example.com"
	do := func(method, path, token string, body any) *httptest.ResponseRecorder {
		var buf bytes.Buffer
		if body != nil {
			_ = json.NewEncoder(&buf).Encode(body)
		}
		req := httptest.NewRequest(method, path, &buf)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	// Register.
	w := do("POST", "/api/auth/register", "", map[string]any{"email": email, "password": "password123"})
	if w.Code != http.StatusCreated {
		t.Fatalf("register: %d %s", w.Code, w.Body.String())
	}
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &tok)
	if tok.AccessToken == "" {
		t.Fatal("no access token")
	}

	// Authenticated preferences round-trip.
	if w := do("GET", "/api/me/preferences", tok.AccessToken, nil); w.Code != http.StatusOK {
		t.Fatalf("get prefs: %d %s", w.Code, w.Body.String())
	}
	if w := do("PATCH", "/api/me/preferences", tok.AccessToken, map[string]any{"theme": "dark"}); w.Code != http.StatusOK {
		t.Fatalf("patch prefs: %d %s", w.Code, w.Body.String())
	}

	// Unauthenticated access is rejected.
	if w := do("GET", "/api/me/preferences", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}

	// SSRF guard rejects link-local LLM base_url.
	if w := do("POST", "/api/llm-configs", tok.AccessToken, map[string]any{
		"name": "x", "base_url": "http://169.254.169.254/v1", "model": "m",
	}); w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for metadata base_url, got %d %s", w.Code, w.Body.String())
	}
}
