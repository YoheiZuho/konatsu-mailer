package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/yoheizuho/konatsu-mailer/internal/config"
	"github.com/yoheizuho/konatsu-mailer/internal/crypto"
	"github.com/yoheizuho/konatsu-mailer/internal/store"
)

// isForeignKeyViolation reports whether err is a Postgres FK violation (23503),
// which here means the account's user_id no longer exists (stale session).
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}

// accountResp is the API representation of a mail account. The stored password
// is never returned.
type accountResp struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	ImapHost        string `json:"imap_host"`
	ImapPort        int    `json:"imap_port"`
	ImapUseTLS      bool   `json:"imap_use_tls"`
	SmtpHost        string `json:"smtp_host"`
	SmtpPort        int    `json:"smtp_port"`
	SmtpUseStarttls bool   `json:"smtp_use_starttls"`
	AuthUser        string `json:"auth_user"`
	IsActive        bool   `json:"is_active"`
}

type accountCreateReq struct {
	Email           string `json:"email" binding:"required,email"`
	ImapHost        string `json:"imap_host" binding:"required"`
	ImapPort        int    `json:"imap_port"`
	ImapUseTLS      *bool  `json:"imap_use_tls"`
	SmtpHost        string `json:"smtp_host" binding:"required"`
	SmtpPort        int    `json:"smtp_port"`
	SmtpUseStarttls *bool  `json:"smtp_use_starttls"`
	AuthUser        string `json:"auth_user"`
	Password        string `json:"password" binding:"required"`
	IsActive        *bool  `json:"is_active"`
}

const accountColumns = `id, email, imap_host, imap_port, imap_use_tls,
	smtp_host, smtp_port, smtp_use_starttls, auth_user, is_active`

func scanAccount(row pgx.Row) (accountResp, error) {
	var a accountResp
	err := row.Scan(
		&a.ID, &a.Email, &a.ImapHost, &a.ImapPort, &a.ImapUseTLS,
		&a.SmtpHost, &a.SmtpPort, &a.SmtpUseStarttls, &a.AuthUser, &a.IsActive,
	)
	return a, err
}

func listAccountsHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		rows, err := db.Pool.Query(c.Request.Context(),
			`SELECT `+accountColumns+` FROM accounts WHERE user_id=$1 ORDER BY created_at`, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to list accounts"))
			return
		}
		defer rows.Close()

		items := make([]accountResp, 0)
		for rows.Next() {
			a, err := scanAccount(rows)
			if err != nil {
				c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to read account"))
				return
			}
			items = append(items, a)
		}
		c.JSON(http.StatusOK, gin.H{"items": items})
	}
}

func createAccountHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")

		var req accountCreateReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}

		enc, err := crypto.NewAES256GCM(cfg.MasterEncKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "encryption unavailable"))
			return
		}
		encPassword, err := enc.Encrypt([]byte(req.Password))
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to encrypt credentials"))
			return
		}

		authUser := req.AuthUser
		if authUser == "" {
			authUser = req.Email
		}

		row := db.Pool.QueryRow(c.Request.Context(),
			`INSERT INTO accounts
			   (user_id, email, imap_host, imap_port, imap_use_tls,
			    smtp_host, smtp_port, smtp_use_starttls, auth_user, password_encrypted, is_active)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			 RETURNING `+accountColumns,
			userID, req.Email,
			req.ImapHost, intOr(req.ImapPort, 993), boolOr(req.ImapUseTLS, true),
			req.SmtpHost, intOr(req.SmtpPort, 587), boolOr(req.SmtpUseStarttls, true),
			authUser, encPassword, boolOr(req.IsActive, true),
		)
		a, err := scanAccount(row)
		if err != nil {
			if isForeignKeyViolation(err) {
				c.JSON(http.StatusUnauthorized, errorResponse("session_invalid", "please sign in again"))
				return
			}
			c.JSON(http.StatusBadRequest, errorResponse("account_error", "failed to create account"))
			return
		}
		c.JSON(http.StatusCreated, a)
	}
}

type accountPatchReq struct {
	Email           *string `json:"email"`
	ImapHost        *string `json:"imap_host"`
	ImapPort        *int    `json:"imap_port"`
	ImapUseTLS      *bool   `json:"imap_use_tls"`
	SmtpHost        *string `json:"smtp_host"`
	SmtpPort        *int    `json:"smtp_port"`
	SmtpUseStarttls *bool   `json:"smtp_use_starttls"`
	AuthUser        *string `json:"auth_user"`
	Password        *string `json:"password"`
	IsActive        *bool   `json:"is_active"`
}

func updateAccountHandler(db *store.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		id := c.Param("id")
		ctx := c.Request.Context()

		var req accountPatchReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse("validation_error", err.Error()))
			return
		}

		// Load the current row (also enforces ownership).
		cur, err := scanAccount(db.Pool.QueryRow(ctx,
			`SELECT `+accountColumns+` FROM accounts WHERE id=$1 AND user_id=$2`, id, userID))
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "account not found"))
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to load account"))
			return
		}

		// Overlay provided fields.
		if req.Email != nil {
			cur.Email = *req.Email
		}
		if req.ImapHost != nil {
			cur.ImapHost = *req.ImapHost
		}
		if req.ImapPort != nil {
			cur.ImapPort = *req.ImapPort
		}
		if req.ImapUseTLS != nil {
			cur.ImapUseTLS = *req.ImapUseTLS
		}
		if req.SmtpHost != nil {
			cur.SmtpHost = *req.SmtpHost
		}
		if req.SmtpPort != nil {
			cur.SmtpPort = *req.SmtpPort
		}
		if req.SmtpUseStarttls != nil {
			cur.SmtpUseStarttls = *req.SmtpUseStarttls
		}
		if req.AuthUser != nil {
			cur.AuthUser = *req.AuthUser
		}
		if req.IsActive != nil {
			cur.IsActive = *req.IsActive
		}

		// Re-encrypt the password only when a new one is supplied.
		var newEnc []byte
		if req.Password != nil && *req.Password != "" {
			enc, err := crypto.NewAES256GCM(cfg.MasterEncKey)
			if err != nil {
				c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "encryption unavailable"))
				return
			}
			newEnc, err = enc.Encrypt([]byte(*req.Password))
			if err != nil {
				c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to encrypt credentials"))
				return
			}
		}

		row := db.Pool.QueryRow(ctx,
			`UPDATE accounts SET
			   email=$3, imap_host=$4, imap_port=$5, imap_use_tls=$6,
			   smtp_host=$7, smtp_port=$8, smtp_use_starttls=$9, auth_user=$10, is_active=$11,
			   password_encrypted = COALESCE($12, password_encrypted)
			 WHERE id=$1 AND user_id=$2
			 RETURNING `+accountColumns,
			id, userID,
			cur.Email, cur.ImapHost, cur.ImapPort, cur.ImapUseTLS,
			cur.SmtpHost, cur.SmtpPort, cur.SmtpUseStarttls, cur.AuthUser, cur.IsActive,
			newEnc,
		)
		a, err := scanAccount(row)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to update account"))
			return
		}
		c.JSON(http.StatusOK, a)
	}
}

func deleteAccountHandler(db *store.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("userID")
		id := c.Param("id")
		tag, err := db.Pool.Exec(c.Request.Context(),
			`DELETE FROM accounts WHERE id=$1 AND user_id=$2`, id, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, errorResponse("internal_error", "failed to delete account"))
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, errorResponse("not_found", "account not found"))
			return
		}
		c.JSON(http.StatusOK, gin.H{"deleted": true})
	}
}
