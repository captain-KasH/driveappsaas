package identity

import (
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
)

func Register(mux *http.ServeMux, pool *pgxpool.Pool) {
	mux.Handle("POST /v1/auth/login", httpx.Handler(func(w http.ResponseWriter, r *http.Request) error {
		var in struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := httpx.DecodeJSON(r, &in); err != nil {
			return err
		}
		token, p, err := Login(r.Context(), pool, in.Email, in.Password, r.UserAgent(), RemoteIP(r))
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
}

func RemoteIP(r *http.Request) string {
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i > 0 {
		host = host[:i]
	}
	return strings.Trim(host, "[]")
}
