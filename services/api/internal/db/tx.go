package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Scope is the row-level-security context for one transaction. Settings are
// transaction-local (set_config(..., true)) so pooled connections never leak scope.
type Scope struct {
	UserID   string // authenticated user, "" for anonymous
	TenantID string // tenant whose rows are visible, "" for none
	Platform bool   // cross-tenant platform scope; only platform handlers and invite acceptance set this
}

func InTx(ctx context.Context, pool *pgxpool.Pool, s Scope, fn func(pgx.Tx) error) error {
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		platform := ""
		if s.Platform {
			platform = "on"
		}
		if _, err := tx.Exec(ctx, `SELECT set_config('app.user_id', $1, true),
			set_config('app.tenant_id', $2, true), set_config('app.platform', $3, true)`,
			s.UserID, s.TenantID, platform); err != nil {
			return err
		}
		return fn(tx)
	})
}
