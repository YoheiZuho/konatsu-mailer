package store

import (
	"context"

	"github.com/yoheizuho/konatsu-mailer/internal/domain"
)

const accountSelect = `id, user_id, email, imap_host, imap_port, imap_use_tls,
	smtp_host, smtp_port, smtp_use_starttls, auth_user, password_encrypted,
	sync_state, is_active, created_at`

func scanAccount(row interface {
	Scan(dest ...any) error
}) (domain.Account, error) {
	var a domain.Account
	err := row.Scan(
		&a.ID, &a.UserID, &a.Email, &a.IMAPHost, &a.IMAPPort, &a.IMAPUseTLS,
		&a.SMTPHost, &a.SMTPPort, &a.SMTPUseStartTLS, &a.AuthUser, &a.PasswordEncrypted,
		&a.SyncState, &a.IsActive, &a.CreatedAt,
	)
	return a, err
}

// ActiveAccounts returns all active accounts across all users (for the sync
// manager).
func (db *DB) ActiveAccounts(ctx context.Context) ([]domain.Account, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT `+accountSelect+` FROM accounts WHERE is_active = true ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Account
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// GetAccount loads a single account by id (no user scoping; callers that serve
// user requests must verify ownership separately).
func (db *DB) GetAccount(ctx context.Context, id string) (domain.Account, error) {
	row := db.Pool.QueryRow(ctx, `SELECT `+accountSelect+` FROM accounts WHERE id=$1`, id)
	return scanAccount(row)
}

// GetAccountForUser loads an account owned by the given user.
func (db *DB) GetAccountForUser(ctx context.Context, id, userID string) (domain.Account, error) {
	row := db.Pool.QueryRow(ctx,
		`SELECT `+accountSelect+` FROM accounts WHERE id=$1 AND user_id=$2`, id, userID)
	return scanAccount(row)
}

// FirstActiveAccountForUser returns the user's first active account, used as a
// default sender when none is specified.
func (db *DB) FirstActiveAccountForUser(ctx context.Context, userID string) (domain.Account, error) {
	row := db.Pool.QueryRow(ctx,
		`SELECT `+accountSelect+` FROM accounts WHERE user_id=$1 AND is_active=true ORDER BY created_at LIMIT 1`, userID)
	return scanAccount(row)
}

// UpdateSyncState persists the per-folder UID sync cursor for an account.
func (db *DB) UpdateSyncState(ctx context.Context, accountID domain.UUID, state domain.SyncState) error {
	_, err := db.Pool.Exec(ctx, `UPDATE accounts SET sync_state=$2 WHERE id=$1`, accountID, state)
	return err
}
