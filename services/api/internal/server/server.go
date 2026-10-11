// Package server wires modules onto one HTTP mux.
package server

import (
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/captain-KasH/driveappsaas/services/api/internal/booking"
	"github.com/captain-KasH/driveappsaas/services/api/internal/fleet"
	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
	"github.com/captain-KasH/driveappsaas/services/api/internal/identity"
	"github.com/captain-KasH/driveappsaas/services/api/internal/tenancy"
	"github.com/captain-KasH/driveappsaas/services/api/internal/web"
)

// New wires every module. The booking service is passed in so main can also run its dispatch worker.
func New(pool *pgxpool.Pool, bookings *booking.Service) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /healthz", httpx.Handler(func(w http.ResponseWriter, r *http.Request) error {
		if err := pool.Ping(r.Context()); err != nil {
			return httpx.Errorf(http.StatusServiceUnavailable, "db_unavailable", "database unreachable")
		}
		return httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}))
	identity.Register(mux, pool)
	ten := &tenancy.Service{Pool: pool}
	ten.Register(mux)
	(&fleet.Service{Pool: pool, Tenancy: ten}).Register(mux)
	bookings.Register(mux)
	apiNotFound := httpx.Handler(func(http.ResponseWriter, *http.Request) error { return httpx.ErrNotFound })
	site, hasSite := web.Handler()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hasSite && !strings.HasPrefix(r.URL.Path, "/v1/") && (r.Method == http.MethodGet || r.Method == http.MethodHead) {
			site.ServeHTTP(w, r)
			return
		}
		apiNotFound.ServeHTTP(w, r)
	}))
	return httpx.Middleware(mux)
}

// NewBookings builds the booking service sharing one tenancy service.
func NewBookings(pool *pgxpool.Pool) *booking.Service {
	return &booking.Service{Pool: pool, Tenancy: &tenancy.Service{Pool: pool}}
}
