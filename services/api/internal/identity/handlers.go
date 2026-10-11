package identity

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
)

func Register(mux *http.ServeMux, pool *pgxpool.Pool) {
	mux.Handle("POST /v1/auth/login", httpx.Handler(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Client   string `json:"client"` // "web" (default) or "mobile" (180-day sliding session)
		}
		if err := httpx.DecodeJSON(r, &in); err != nil {
			return err
		}
		token, p, err := LoginAs(r.Context(), pool, in.Email, in.Password, r.UserAgent(), RemoteIP(r), in.Client)
		if err != nil {
			return err
		}
		return httpx.WriteJSON(w, http.StatusOK, map[string]any{"token": token, "userId": p.UserID, "platformRole": p.PlatformRole})
	}))
	mux.Handle("POST /v1/auth/logout", RequireAuth(pool, func(w http.ResponseWriter, r *http.Request) error {
		p, _ := FromContext(r.Context())
		if err := Logout(r.Context(), pool, p.SessionID); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}))
	mux.Handle("GET /v1/auth/sessions", RequireAuth(pool, func(w http.ResponseWriter, r *http.Request) error {
		p, _ := FromContext(r.Context())
		type session struct {
			ID        string     `json:"id"`
			Client    string     `json:"client"`
			UserAgent *string    `json:"userAgent"`
			CreatedAt time.Time  `json:"createdAt"`
			LastSeen  *time.Time `json:"lastSeenAt"`
			ExpiresAt time.Time  `json:"expiresAt"`
			Current   bool       `json:"current"`
		}
		rows, err := pool.Query(r.Context(), `SELECT id, client, user_agent, created_at, last_seen_at, expires_at, id = $2 FROM sessions
			WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > now() ORDER BY coalesce(last_seen_at, created_at) DESC`, p.UserID, p.SessionID)
		if err != nil {
			return err
		}
		list, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (session, error) {
			var s session
			return s, row.Scan(&s.ID, &s.Client, &s.UserAgent, &s.CreatedAt, &s.LastSeen, &s.ExpiresAt, &s.Current)
		})
		if err != nil {
			return err
		}
		return httpx.WriteJSON(w, http.StatusOK, map[string]any{"sessions": list})
	}))
	mux.Handle("DELETE /v1/auth/sessions/{sessionID}", RequireAuth(pool, func(w http.ResponseWriter, r *http.Request) error {
		p, _ := FromContext(r.Context())
		tag, err := pool.Exec(r.Context(), `UPDATE sessions SET revoked_at = now() WHERE id::text = $1 AND user_id = $2 AND revoked_at IS NULL`,
			strings.ToLower(r.PathValue("sessionID")), p.UserID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return httpx.ErrNotFound
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}))
	// Sign out everywhere except here (lost phone, shared computer).
	mux.Handle("POST /v1/auth/logout-all", RequireAuth(pool, func(w http.ResponseWriter, r *http.Request) error {
		p, _ := FromContext(r.Context())
		tag, err := pool.Exec(r.Context(), `UPDATE sessions SET revoked_at = now() WHERE user_id = $1 AND id <> $2 AND revoked_at IS NULL`, p.UserID, p.SessionID)
		if err != nil {
			return err
		}
		return httpx.WriteJSON(w, http.StatusOK, map[string]any{"revoked": tag.RowsAffected()})
	}))
}

// trustProxy is set only where a reverse proxy we control overwrites X-Real-IP
// (packline nginx sets it from the Cloudflare-restored client address).
var trustProxy = os.Getenv("TAXIOS_TRUST_PROXY") == "1"

func RemoteIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); trustProxy && ip != "" {
		return ip
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i > 0 {
		host = host[:i]
	}
	return strings.Trim(host, "[]")
}
