// Package audit writes append-only audit events inside the caller's transaction,
// so an action and its audit record commit or roll back together.
package audit

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/captain-KasH/driveappsaas/services/api/internal/httpx"
)

type Event struct {
	TenantID    string // "" = platform-level event
	ActorUserID string
	Action      string // e.g. "tenant.created"
	TargetType  string
	TargetID    string
	Reason      string
	Metadata    map[string]any
}

func Record(ctx context.Context, tx pgx.Tx, e Event) error {
	if e.Metadata == nil {
		e.Metadata = map[string]any{}
	}
	_, err := tx.Exec(ctx, `INSERT INTO audit_events
		(tenant_id, actor_user_id, action, target_type, target_id, reason, request_id, metadata)
		VALUES (nullif($1,'')::uuid, nullif($2,'')::uuid, $3, $4, $5, nullif($6,''), nullif($7,''), $8)`,
		e.TenantID, e.ActorUserID, e.Action, e.TargetType, e.TargetID, e.Reason, httpx.RequestID(ctx), e.Metadata)
	return err
}
