package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ProviderUpdate a fizetési szolgáltatótól (webhookból) érkezett állapot egy
// előfizetésről. Ez az igazság: a szolgáltató terheli a kártyát, ő tudja,
// meddig van kifizetve az időszak.
type ProviderUpdate struct {
	Status            string
	PeriodStart       time.Time
	PeriodEnd         time.Time
	CancelAtPeriodEnd bool
}

// SubscriptionByExternalID a szolgáltató oldali előfizetés-azonosító alapján
// keresi meg a sort; ha nincs, ErrNotFound.
func (s *Store) SubscriptionByExternalID(ctx context.Context, provider, externalID string) (Subscription, error) {
	const q = `SELECT ` + subscriptionCols + ` FROM subscriptions
WHERE provider = $1 AND external_subscription_id = $2 AND external_subscription_id <> ''
ORDER BY created_at DESC
LIMIT 1`
	sub, err := scanSubscription(s.pool.QueryRow(ctx, q, provider, externalID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Subscription{}, ErrNotFound
	}
	if err != nil {
		return Subscription{}, fmt.Errorf("előfizetés lekérdezés külső azonosítóval: %w", err)
	}
	return sub, nil
}

// ApplyProviderUpdate a szolgáltató által közölt állapotra állítja az
// előfizetést, és ugyanabban a tranzakcióban szinkronizálja a jogosultságokat.
func (s *Store) ApplyProviderUpdate(ctx context.Context, id string, upd ProviderUpdate, sync EntitlementSync) (Subscription, error) {
	var out Subscription
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		const q = `
UPDATE subscriptions
SET status               = $2,
    current_period_start = $3,
    current_period_end   = $4,
    cancel_at_period_end = $5,
    canceled_at          = CASE
                             WHEN $2 = 'canceled' OR $5 THEN COALESCE(canceled_at, now())
                             ELSE NULL
                           END,
    updated_at           = now()
WHERE id = $1
RETURNING ` + subscriptionCols

		sub, err := scanSubscription(tx.QueryRow(ctx, q, id, upd.Status, upd.PeriodStart, upd.PeriodEnd, upd.CancelAtPeriodEnd))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("szolgáltatói állapot rögzítése: %w", err)
		}
		out = sub
		return syncEntitlements(ctx, tx, sub.Subject, sync)
	})
	return out, err
}

// ProviderEventSeen igaz, ha ezt a webhook eseményt már feldolgoztuk. A
// szolgáltatók ugyanazt az eseményt többször is elküldhetik.
func (s *Store) ProviderEventSeen(ctx context.Context, provider, eventID string) (bool, error) {
	var seen bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM billing_events WHERE provider = $1 AND event_id = $2)`,
		provider, eventID).Scan(&seen)
	if err != nil {
		return false, fmt.Errorf("webhook esemény ellenőrzése: %w", err)
	}
	return seen, nil
}

// RecordProviderEvent feljegyzi a sikeresen feldolgozott webhook eseményt.
// Csak a feldolgozás UTÁN hívjuk: ha közben hiba történik, a szolgáltató
// újraküldi, és akkor újra nekifutunk.
func (s *Store) RecordProviderEvent(ctx context.Context, provider, eventID, eventType string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO billing_events (provider, event_id, event_type) VALUES ($1, $2, $3)
         ON CONFLICT (provider, event_id) DO NOTHING`,
		provider, eventID, eventType)
	if err != nil {
		return fmt.Errorf("webhook esemény rögzítése: %w", err)
	}
	return nil
}
