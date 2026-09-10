package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// GrantEntitlement kézzel ad ki egy jogosultságot. Ugyanarra a kulcsra
// ismételve nem hibázik, hanem felülírja a lejáratot: az admin szándéka
// ilyenkor a hosszabbítás, nem az ütközés.
//
// Az expiresAt nil értéke lejárat nélküli jogosultság.
func (s *Store) GrantEntitlement(ctx context.Context, subject, featureKey string, expiresAt *time.Time) (Entitlement, error) {
	const q = `
INSERT INTO entitlements (subject, feature_key, granted_at, expires_at)
VALUES ($1, $2, now(), $3)
ON CONFLICT (subject, feature_key)
DO UPDATE SET granted_at = now(), expires_at = EXCLUDED.expires_at
RETURNING feature_key, granted_at, expires_at`

	var e Entitlement
	if err := s.pool.QueryRow(ctx, q, subject, featureKey, expiresAt).
		Scan(&e.FeatureKey, &e.GrantedAt, &e.ExpiresAt); err != nil {
		return Entitlement{}, fmt.Errorf("jogosultság kiadása: %w", err)
	}
	return e, nil
}

// RevokeEntitlement visszavon egy jogosultságot. Ha nem volt kiadva,
// ErrNotFound jön vissza – az admin így tudja, hogy nem az történt, amit várt.
func (s *Store) RevokeEntitlement(ctx context.Context, subject, featureKey string) (Entitlement, error) {
	const q = `
DELETE FROM entitlements
WHERE subject = $1 AND feature_key = $2
RETURNING feature_key, granted_at, expires_at`

	var e Entitlement
	err := s.pool.QueryRow(ctx, q, subject, featureKey).
		Scan(&e.FeatureKey, &e.GrantedAt, &e.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Entitlement{}, ErrNotFound
	}
	if err != nil {
		return Entitlement{}, fmt.Errorf("jogosultság visszavonása: %w", err)
	}
	return e, nil
}
