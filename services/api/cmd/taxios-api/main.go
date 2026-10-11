// Command taxios-api runs the Taxios API, its migrations and admin bootstrap.
//
//	taxios-api serve                      # DATABASE_URL (app role), ADDR
//	taxios-api migrate                    # MIGRATE_DATABASE_URL (schema owner)
//	taxios-api create-platform-admin EMAIL "FULL NAME"   # password from TAXIOS_ADMIN_PASSWORD
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata" // distroless has no zoneinfo; tenant timezones are validated with LoadLocation

	"github.com/captain-KasH/driveappsaas/services/api/internal/db"
	"github.com/captain-KasH/driveappsaas/services/api/internal/identity"
	"github.com/captain-KasH/driveappsaas/services/api/internal/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: taxios-api serve|migrate|create-platform-admin")
	}
	switch args[0] {
	case "migrate":
		pool, err := db.Open(ctx, mustEnv("MIGRATE_DATABASE_URL"))
		if err != nil {
			return err
		}
		defer pool.Close()
		applied, err := db.Migrate(ctx, pool)
		fmt.Println("applied:", applied)
		return err

	case "create-platform-admin":
		if len(args) != 3 {
			return errors.New(`usage: create-platform-admin EMAIL "FULL NAME" (password in TAXIOS_ADMIN_PASSWORD)`)
		}
		hash, err := identity.HashPassword(mustEnv("TAXIOS_ADMIN_PASSWORD"))
		if err != nil {
			return err
		}
		pool, err := db.Open(ctx, mustEnv("MIGRATE_DATABASE_URL"))
		if err != nil {
			return err
		}
		defer pool.Close()
		_, err = pool.Exec(ctx, `INSERT INTO users(email, full_name, password_hash, platform_role) VALUES ($1,$2,$3,'super_admin')
			ON CONFLICT (email) DO UPDATE SET platform_role='super_admin', password_hash=EXCLUDED.password_hash`, args[1], args[2], hash)
		if err == nil {
			fmt.Println("platform super_admin ready:", args[1])
		}
		return err

	case "healthcheck": // used by the container HEALTHCHECK; distroless has no curl
		c := &http.Client{Timeout: 3 * time.Second}
		res, err := c.Get("http://127.0.0.1:8080/healthz")
		if err != nil {
			return err
		}
		res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("healthz returned %d", res.StatusCode)
		}
		return nil

	case "serve":
		pool, err := db.Open(ctx, mustEnv("DATABASE_URL"))
		if err != nil {
			return err
		}
		defer pool.Close()
		var super bool
		if err := pool.QueryRow(ctx, `SELECT rolsuper OR rolbypassrls FROM pg_roles WHERE rolname = current_user`).Scan(&super); err != nil {
			return err
		}
		if super && os.Getenv("TAXIOS_ALLOW_RLS_BYPASS") != "1" {
			return errors.New("DATABASE_URL role bypasses row-level security; connect as " + db.AppRole)
		}
		addr := os.Getenv("ADDR")
		if addr == "" {
			addr = ":8080"
		}
		bookings := server.NewBookings(pool)
		go bookings.Run(ctx, 5*time.Second) // ponytail: in-process dispatch worker; one API replica only until it takes an advisory lock
		srv := &http.Server{Addr: addr, Handler: server.New(pool, bookings), ReadHeaderTimeout: 10 * time.Second}
		go func() {
			<-ctx.Done()
			shutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			srv.Shutdown(shutdown)
		}()
		slog.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
	return fmt.Errorf("unknown command %q", args[0])
}

func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		fmt.Fprintf(os.Stderr, "error: %s is required\n", k)
		os.Exit(2)
	}
	return v
}
