package store

import (
	"context"
	"fmt"
	"time"
)

// Admin naplóműveletek. A konstans a napló gépi kulcsa: az adatbázisba ez
// kerül, a felület ehhez rendel magyar címkét. Ha egy művelet neve
// megváltozna, a régi sorok akkor is olvashatók maradnak a summary miatt.
const (
	AuditSubscriptionGrant      = "subscription.grant"
	AuditSubscriptionChangePlan = "subscription.change_plan"
	AuditSubscriptionCancel     = "subscription.cancel"
	AuditSubscriptionReactivate = "subscription.reactivate"
	AuditInvoiceCreate          = "invoice.create"
	AuditInvoicePay             = "invoice.pay"
	AuditInvoiceVoid            = "invoice.void"
	AuditEntitlementGrant       = "entitlement.grant"
	AuditEntitlementRevoke      = "entitlement.revoke"
	AuditUserEnable             = "user.enable"
	AuditUserDisable            = "user.disable"
)

// AuditActions a naplóban előforduló műveletek teljes listája, a felület
// szűrőjéhez. A sorrend a szűrő legördülőjének sorrendje.
func AuditActions() []string {
	return []string{
		AuditSubscriptionGrant,
		AuditSubscriptionChangePlan,
		AuditSubscriptionCancel,
		AuditSubscriptionReactivate,
		AuditInvoiceCreate,
		AuditInvoicePay,
		AuditInvoiceVoid,
		AuditEntitlementGrant,
		AuditEntitlementRevoke,
		AuditUserEnable,
		AuditUserDisable,
	}
}

// AuditEntry egy naplósor.
type AuditEntry struct {
	ID           string         `json:"id"`
	At           time.Time      `json:"at"`
	ActorSubject string         `json:"actor_subject"`
	ActorLabel   string         `json:"actor_label"`
	Action       string         `json:"action"`
	TargetType   string         `json:"target_type"`
	TargetID     string         `json:"target_id"`
	Subject      string         `json:"subject"`
	Summary      string         `json:"summary"`
	Detail       map[string]any `json:"detail"`
}

// NewAuditEntry a rögzítendő naplósor. Az időbélyeget az adatbázis adja.
type NewAuditEntry struct {
	ActorSubject string
	ActorLabel   string
	Action       string
	TargetType   string
	TargetID     string
	Subject      string
	Summary      string
	Detail       map[string]any
}

// AuditFilter a naplólista szűrése.
type AuditFilter struct {
	Actor   string
	Subject string
	Action  string
	Since   *time.Time
	Until   *time.Time
	Limit   int
	Offset  int
}

const auditCols = `id, at, actor_subject, actor_label, action, target_type, target_id, subject, summary, detail`

// RecordAudit egy naplósort ír. A napló csak nő: nincs update és nincs
// delete, ezért itt egyetlen INSERT elég.
func (s *Store) RecordAudit(ctx context.Context, in NewAuditEntry) (AuditEntry, error) {
	if in.Detail == nil {
		in.Detail = map[string]any{}
	}
	const q = `
INSERT INTO admin_audit (id, actor_subject, actor_label, action, target_type, target_id, subject, summary, detail)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING ` + auditCols

	row := s.pool.QueryRow(ctx, q, newID(), in.ActorSubject, in.ActorLabel, in.Action,
		in.TargetType, in.TargetID, in.Subject, in.Summary, in.Detail)

	var e AuditEntry
	if err := row.Scan(&e.ID, &e.At, &e.ActorSubject, &e.ActorLabel, &e.Action,
		&e.TargetType, &e.TargetID, &e.Subject, &e.Summary, &e.Detail); err != nil {
		return AuditEntry{}, fmt.Errorf("napló írása: %w", err)
	}
	return e, nil
}

// ListAudit a naplósorok szűrt, lapozott listája, legfrissebb elöl.
func (s *Store) ListAudit(ctx context.Context, f AuditFilter) ([]AuditEntry, int, error) {
	const q = `
SELECT ` + auditCols + `, count(*) OVER () AS total
FROM admin_audit
WHERE ($1::text = '' OR actor_subject = $1)
  AND ($2::text = '' OR subject = $2)
  AND ($3::text = '' OR action = $3)
  AND ($4::timestamptz IS NULL OR at >= $4)
  AND ($5::timestamptz IS NULL OR at < $5)
ORDER BY at DESC, id
LIMIT $6 OFFSET $7`

	rows, err := s.pool.Query(ctx, q, f.Actor, f.Subject, f.Action, f.Since, f.Until, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("napló lekérdezése: %w", err)
	}
	defer rows.Close()

	out := []AuditEntry{}
	total := 0
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.At, &e.ActorSubject, &e.ActorLabel, &e.Action,
			&e.TargetType, &e.TargetID, &e.Subject, &e.Summary, &e.Detail, &total); err != nil {
			return nil, 0, fmt.Errorf("napló sor: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("napló sorok: %w", err)
	}
	// Üres találathalmaznál a window function nem fut le, a total 0 marad.
	return out, total, nil
}
