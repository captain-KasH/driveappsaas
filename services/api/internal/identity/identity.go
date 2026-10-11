// Package identity handles users, passwords and session tokens.
package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/captain-KasH/driveappsaas/services/api/internal/db"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
)

const (
	SessionTTL     = 30 * 24 * time.Hour  // web
	MobileTTL      = 180 * 24 * time.Hour // mobile apps stay signed in; expiry slides on use
	tokenPrefix    = "txs_"
	MinPasswordLen = 10
)

type Principal struct {
	UserID       string
	Email        string
	PlatformRole string // "", "super_admin", "support"
	SessionID    string
	Client       string // "web" | "mobile"
}

type ctxKey int

const principalKey ctxKey = 0

func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey).(Principal)
	return p, ok
}

// NewToken returns a random opaque token and the hash stored server-side.
func NewToken(prefix string) (token string, hash []byte) {
	b := make([]byte, 32)
	rand.Read(b)
	token = prefix + base64.RawURLEncoding.EncodeToString(b)
	return token, HashToken(token)
}

func HashToken(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

func HashPassword(pw string) (string, error) {
	if len(pw) < MinPasswordLen || len(pw) > 72 {
		return "", httpx.Errorf(http.StatusUnprocessableEntity, "weak_password", "password must be %d-72 characters", MinPasswordLen)
	}
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h), err
}

// dummyHash keeps login timing similar whether or not the email exists.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("timing-equaliser-not-a-password"), bcrypt.DefaultCost)

var ErrBadCredentials = httpx.Errorf(http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")

// Login verifies credentials and creates a session.
func Login(ctx context.Context, pool *pgxpool.Pool, email, password, userAgent, ip string) (string, Principal, error) {
	return LoginAs(ctx, pool, email, password, userAgent, ip, "web")
}

// LoginAs is Login for a named client ("web" or "mobile"; mobile sessions live longer).
func LoginAs(ctx context.Context, pool *pgxpool.Pool, email, password, userAgent, ip, client string) (string, Principal, error) {
	var p Principal
	var hash *string
	var disabled *time.Time
	var platformRole *string
	err := pool.QueryRow(ctx, `SELECT id, email, password_hash, platform_role, disabled_at FROM users WHERE email = $1`,
		strings.TrimSpace(email)).Scan(&p.UserID, &p.Email, &hash, &platformRole, &disabled)
	if errors.Is(err, pgx.ErrNoRows) {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return "", p, ErrBadCredentials
	}
	if err != nil {
		return "", p, err
	}
	if hash == nil || bcrypt.CompareHashAndPassword([]byte(*hash), []byte(password)) != nil || disabled != nil {
		return "", p, ErrBadCredentials
	}
	if platformRole != nil {
		p.PlatformRole = *platformRole
	}
	token, err := CreateSessionAs(ctx, pool, p.UserID, userAgent, ip, client)
	return token, p, err
}

func ttlFor(client string) time.Duration {
	if client == "mobile" {
		return MobileTTL
	}
	return SessionTTL
}

// Execer is satisfied by both the pool and a transaction.
type Execer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func CreateSession(ctx context.Context, q Execer, userID, userAgent, ip string) (string, error) {
	return CreateSessionAs(ctx, q, userID, userAgent, ip, "web")
}

func CreateSessionAs(ctx context.Context, q Execer, userID, userAgent, ip, client string) (string, error) {
	if client != "mobile" {
		client = "web"
	}
	token, hash := NewToken(tokenPrefix)
	var ipArg any
	if net.ParseIP(ip) != nil {
		ipArg = ip
	}
	_, err := q.Exec(ctx, `INSERT INTO sessions(user_id, token_hash, expires_at, user_agent, ip, client, last_seen_at) VALUES ($1,$2,$3,$4,$5,$6,now())`,
		userID, hash, time.Now().Add(ttlFor(client)), truncate(userAgent, 300), ipArg, client)
	return token, err
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func Logout(ctx context.Context, pool *pgxpool.Pool, sessionID string) error {
	_, err := pool.Exec(ctx, `UPDATE sessions SET revoked_at = now() WHERE id = $1 AND revoked_at IS NULL`, sessionID)
	return err
}

// Authenticate resolves a bearer token; it never trusts any client-sent identity.
func Authenticate(ctx context.Context, pool *pgxpool.Pool, token string) (Principal, error) {
	var p Principal
	var platformRole *string
	var expires time.Time
	var lastSeen *time.Time
	err := pool.QueryRow(ctx, `
		SELECT u.id, u.email, u.platform_role, s.id, s.client, s.expires_at, s.last_seen_at FROM sessions s JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.revoked_at IS NULL AND s.expires_at > now() AND u.disabled_at IS NULL`,
		HashToken(token)).Scan(&p.UserID, &p.Email, &platformRole, &p.SessionID, &p.Client, &expires, &lastSeen)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, httpx.ErrUnauthenticated
	}
	if err != nil {
		return p, err
	}
	if platformRole != nil {
		p.PlatformRole = *platformRole
	}
	// Sliding expiry, written at most every 5 minutes per session: a session in use never expires;
	// one left idle for its full TTL does.
	if lastSeen == nil || time.Since(*lastSeen) > 5*time.Minute {
		_, err = pool.Exec(ctx, `UPDATE sessions SET last_seen_at = now(), expires_at = greatest(expires_at, now() + $2::interval) WHERE id = $1`,
			p.SessionID, ttlFor(p.Client))
	}
	return p, err
}

// RequireAuth wraps a handler so it only runs for an authenticated principal.
func RequireAuth(pool *pgxpool.Pool, next httpx.Handler) httpx.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || !strings.HasPrefix(token, tokenPrefix) {
			return httpx.ErrUnauthenticated
		}
		p, err := Authenticate(r.Context(), pool, token)
		if err != nil {
			return err
		}
		return next(w, r.WithContext(context.WithValue(r.Context(), principalKey, p)))
	}
}

// Scope builds the RLS scope for an authenticated user with no tenant selected.
func (p Principal) Scope() db.Scope { return db.Scope{UserID: p.UserID} }

// CheckPassword compares a stored bcrypt hash with a candidate password.
func CheckPassword(hash *string, pw string) bool {
	if hash == nil {
		bcrypt.CompareHashAndPassword(dummyHash, []byte(pw))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(*hash), []byte(pw)) == nil
}
