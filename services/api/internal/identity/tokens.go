package identity

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
)

const (
	PurposeVerifyEmail   = "verify_email"
	PurposeResetPassword = "reset_password"
	VerifyTTL            = 24 * time.Hour
	ResetTTL             = 1 * time.Hour
)

var ErrBadToken = httpx.Errorf(http.StatusNotFound, "invalid_token", "this link is invalid or has expired")

// IssueToken creates a single-use token for purpose; the raw token is returned once.
// Earlier unused tokens of the same purpose are invalidated so only the latest link works.
func IssueToken(ctx context.Context, tx pgx.Tx, userID, purpose string) (string, error) {
	ttl := VerifyTTL
	prefix := "txv_"
	if purpose == PurposeResetPassword {
		ttl, prefix = ResetTTL, "txr_"
	}
	if _, err := tx.Exec(ctx, `UPDATE user_tokens SET used_at = now() WHERE user_id=$1 AND purpose=$2 AND used_at IS NULL`, userID, purpose); err != nil {
		return "", err
	}
	token, hash := NewToken(prefix)
	_, err := tx.Exec(ctx, `INSERT INTO user_tokens(user_id, purpose, token_hash, expires_at) VALUES ($1,$2,$3,$4)`,
		userID, purpose, hash, time.Now().Add(ttl))
	return token, err
}

// ConsumeToken marks the token used and returns its user. Runs in platform scope (no identity yet).
func ConsumeToken(ctx context.Context, tx pgx.Tx, token, purpose string) (userID string, err error) {
	err = tx.QueryRow(ctx, `UPDATE user_tokens SET used_at = now()
		WHERE token_hash = $1 AND purpose = $2 AND used_at IS NULL AND expires_at > now() RETURNING user_id`,
		HashToken(token), purpose).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrBadToken
	}
	return userID, err
}
