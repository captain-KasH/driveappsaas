// Package server wires modules onto one HTTP mux.
package server

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
	"github.com/captain-KasH/driveappsaas/services/api/internal/identity"
	"github.com/captain-KasH/driveappsaas/services/api/internal/tenancy"
)

func New(pool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /healthz", httpx.Handler(func(w http.ResponseWriter, r *http.Request) error {
		if err := pool.Ping(r.Context()); err != nil {
			return httpx.Errorf(http.StatusServiceUnavailable, "db_unavailable", "database unreachable")
		}
		return httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}))
	identity.Register(mux, pool)
	(&tenancy.Service{Pool: pool}).Register(mux)
	mux.Handle("/", httpx.Handler(func(http.ResponseWriter, *http.Request) error { return httpx.ErrNotFound }))
	return httpx.Middleware(mux)
}
