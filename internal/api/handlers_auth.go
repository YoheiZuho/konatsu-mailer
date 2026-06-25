package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/crypto"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

type registerReq struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name"`
	// MailAccount optionally provisions an IMAP/SMTP account at sign-up time.
	MailAccount *mailAccountReq `json:"mail_account"`
}

// mailAccountReq carries IMAP/SMTP credentials supplied during registration.
type mailAccountReq struct {
	Email           string `json:"email" binding:"required,email"`
	ImapHost        string `json:"imap_host" binding:"required"`
	ImapPort        int    `json:"imap_port"`
	ImapUseTLS      *bool  `json:"imap_use_tls"`
	SmtpHost        string `json:"smtp_host" binding:"required"`
	SmtpPort        int    `json:"smtp_port"`
	SmtpUseStarttls *bool  `json:"smtp_use_starttls"`
	AuthUser        string `json:"auth_user"`
	Password        string `json:"password" binding:"required"`
}

// boolOr returns *p when set, otherwise the default.
func boolOr(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

// intOr returns v when non-zero, otherwise the default.
func intOr(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type tokenResp struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// authConfigHandler exposes public auth settings the login UI needs, such as
// whether self-service registration is currently open.
func authConfigHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"allow_registration": cfg.AllowRegistration})
	}
}

func registerHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.AllowRegistration {
			c.JSON(http.StatusForbidden, errorResponse("registration_disabled", "registration is currently disabled"))
			return
		}

		var req registerReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to hash password"))
			return
		}

		// Encrypt the mail-account password up front (outside the tx) so a crypto
		// failure never leaves a half-created user behind.
		var encPassword []byte
		if req.MailAccount != nil {
			enc, err := crypto.NewAES256GCM(cfg.MasterEncKey)
			if err != nil {
				c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "encryption unavailable"))
				return
			}
			encPassword, err = enc.Encrypt([]byte(req.MailAccount.Password))
			if err != nil {
				c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to encrypt credentials"))
				return
			}
		}

		ctx := c.Request.Context()
		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to start transaction"))
			return
		}
		defer tx.Rollback(ctx) // no-op after a successful commit

		var userID string
		err = tx.QueryRow(ctx,
			`INSERT INTO users(email, password_hash, display_name) VALUES($1,$2,$3) RETURNING id`,
			req.Email, string(hash), req.DisplayName,
		).Scan(&userID)
		if err != nil {
			c.JSON(http.StatusConflict, errorResponse("email_exists", "email already registered"))
			return
		}

		if ma := req.MailAccount; ma != nil {
			authUser := ma.AuthUser
			if authUser == "" {
				authUser = ma.Email
			}
			_, err = tx.Exec(ctx,
				`INSERT INTO accounts
				   (user_id, email, imap_host, imap_port, imap_use_tls,
				    smtp_host, smtp_port, smtp_use_starttls, auth_user, password_encrypted)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
				userID, ma.Email,
				ma.ImapHost, intOr(ma.ImapPort, 993), boolOr(ma.ImapUseTLS, true),
				ma.SmtpHost, intOr(ma.SmtpPort, 587), boolOr(ma.SmtpUseStarttls, true),
				authUser, encPassword,
			)
			if err != nil {
				c.JSON(http.StatusBadRequest, errorResponse("account_error", "failed to create mail account"))
				return
			}
		}

		if err := tx.Commit(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to commit registration"))
			return
		}

		tok, err := issueToken(cfg.JWTSecret, userID, 24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to issue token"))
			return
		}
		refresh, _ := issueToken(cfg.JWTSecret, userID, 7*24*time.Hour)

		c.JSON(http.StatusCreated, tokenResp{
			AccessToken:  tok,
			RefreshToken: refresh,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		})
	}
}

func loginHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}

		ctx := c.Request.Context()
		var userID, hash string
		err := db.Pool.QueryRow(ctx,
			`SELECT id, password_hash FROM users WHERE email=$1`, req.Email,
		).Scan(&userID, &hash)
		if err != nil {
			c.JSON(http.StatusUnauthorized, errorResponse("invalid_credentials", "invalid email or password"))
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, errorResponse("invalid_credentials", "invalid email or password"))
			return
		}

		tok, err := issueToken(cfg.JWTSecret, userID, 24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to issue token"))
			return
		}
		refresh, _ := issueToken(cfg.JWTSecret, userID, 7*24*time.Hour)

		c.JSON(http.StatusOK, tokenResp{
			AccessToken:  tok,
			RefreshToken: refresh,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		})
	}
}

func refreshHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}

		token, err := jwtParse(body.RefreshToken, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, errorResponse("invalid_token", err.Error()))
			return
		}
		claims := token.Claims.(jwtMap)
		userID, _ := claims["sub"].(string)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, errorResponse("invalid_token", "missing sub"))
			return
		}

		tok, err := issueToken(cfg.JWTSecret, userID, 24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to issue token"))
			return
		}
		refresh, _ := issueToken(cfg.JWTSecret, userID, 7*24*time.Hour)

		c.JSON(http.StatusOK, tokenResp{
			AccessToken:  tok,
			RefreshToken: refresh,
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		})
	}
}
