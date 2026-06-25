package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	DatabaseURL string

	// Encryption
	MasterEncKey []byte // must be 32 bytes for AES-256

	// Auth
	JWTSecret string
	// AllowRegistration gates the public /auth/register endpoint.
	AllowRegistration bool

	// Web Push (VAPID)
	VapidPublicKey  string
	VapidPrivateKey string
	VapidSubject    string

	// LLM pipeline
	LlmWorkers       int
	NotifyThreshold  int
	LlmDefaultBaseURL string
	LlmDefaultModel   string
	LlmDefaultAPIKey  string

	// Translation (LibreTranslate). Empty URL disables the feature.
	LibreTranslateURL      string
	LibreTranslateAPIKey   string
	TranslateDefaultTarget string

	// Timeouts
	RequestTimeout time.Duration
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	dbURL := env("DATABASE_URL", "")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	encKeyStr := env("MASTER_ENC_KEY", "")
	if encKeyStr == "" {
		return nil, fmt.Errorf("MASTER_ENC_KEY is required")
	}
	encKey := []byte(encKeyStr)
	if len(encKey) != 32 {
		// If the key is base64-encoded, decode it. For simplicity in dev,
		// we accept raw 32-byte strings. Production should use base64.
		// Here we just pad or slice for safety.
		if len(encKey) < 32 {
			padded := make([]byte, 32)
			copy(padded, encKey)
			encKey = padded
		} else {
			encKey = encKey[:32]
		}
	}

	jwtSecret := env("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	workers, _ := strconv.Atoi(env("LLM_WORKERS", "4"))
	threshold, _ := strconv.Atoi(env("NOTIFY_THRESHOLD", "4"))

	return &Config{
		DatabaseURL:        dbURL,
		MasterEncKey:       encKey,
		JWTSecret:          jwtSecret,
		AllowRegistration:  envBool("ALLOW_REGISTRATION", true),
		VapidPublicKey:     env("VAPID_PUBLIC_KEY", ""),
		VapidPrivateKey:    env("VAPID_PRIVATE_KEY", ""),
		VapidSubject:       env("VAPID_SUBJECT", "mailto:admin@example.com"),
		LlmWorkers:         workers,
		NotifyThreshold:    threshold,
		LlmDefaultBaseURL:  env("LLM_DEFAULT_BASE_URL", "https://api.openai.com/v1"),
		LlmDefaultModel:    env("LLM_DEFAULT_MODEL", "gpt-4o-mini"),
		LlmDefaultAPIKey:   env("LLM_DEFAULT_API_KEY", ""),

		LibreTranslateURL:      env("LIBRETRANSLATE_URL", ""),
		LibreTranslateAPIKey:   env("LIBRETRANSLATE_API_KEY", ""),
		TranslateDefaultTarget: env("TRANSLATE_DEFAULT_TARGET", "ja"),

		RequestTimeout:     30 * time.Second,
	}, nil
}

func env(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

// envBool parses a boolean env var, accepting common truthy/falsey spellings.
func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
