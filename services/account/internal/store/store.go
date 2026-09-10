// Package store a PostgreSQL-alapú perzisztencia.
package store

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

// Profile az alkalmazás-oldali felhasználói profil. Az Email és a Username a
// Keycloak tokenből frissül, a DisplayName és a Locale itt szerkeszthető.
type Profile struct {
	Subject     string    `json:"sub"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Locale      string    `json:"locale"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Entitlement egy megadott prémium funkció jogosultsága.
// Az ExpiresAt nil értéke lejárat nélküli (örökös) jogosultságot jelent.
type Entitlement struct {
	FeatureKey string     `json:"feature_key"`
	GrantedAt  time.Time  `json:"granted_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
}

// Store a pgx connection pool köré húzott vékony adatelérési réteg.
type Store struct {
	pool *pgxpool.Pool
}

// New felépíti a poolt és ellenőrzi, hogy az adatbázis elérhető-e.
func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("pgxpool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("adatbázis ping: %w", err)
	}
	return &Store{pool: pool}, nil
}

// Close elengedi a pool összes kapcsolatát.
func (s *Store) Close() { s.pool.Close() }

// Migrate lefuttatja a beágyazott sémát. Minden utasítás IF NOT EXISTS,
// ezért többszöri indítás mellett is biztonságos.
func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("séma alkalmazása: %w", err)
	}
	return nil
}

// Ping rövid időkorláttal ellenőrzi az adatbázis-kapcsolatot.
func (s *Store) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return s.pool.Ping(ctx)
}

// UpsertProfile első híváskor létrehozza a profilt a token claimjeiből,
// később csak a Keycloak által birtokolt mezőket (email, username) frissíti.
// A display_name-t nem írja felül, mert azt a felhasználó itt szerkesztheti.
func (s *Store) UpsertProfile(ctx context.Context, subject, email, username, name string) (Profile, error) {
	const q = `
INSERT INTO profiles (subject, email, username, display_name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (subject) DO UPDATE
SET email      = EXCLUDED.email,
    username   = EXCLUDED.username,
    updated_at = now()
RETURNING subject, email, username, display_name, locale, created_at, updated_at`

	displayName := name
	if displayName == "" {
		displayName = username
	}

	var p Profile
	err := s.pool.QueryRow(ctx, q, subject, email, username, displayName).Scan(
		&p.Subject, &p.Email, &p.Username, &p.DisplayName, &p.Locale, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return Profile{}, fmt.Errorf("profil upsert: %w", err)
	}
	return p, nil
}

// UpdateProfile a szerkeszthető mezőket módosítja. A nil értékű paraméter
// azt jelenti, hogy a hívó nem küldte az adott mezőt, tehát marad a régi.
func (s *Store) UpdateProfile(ctx context.Context, subject string, displayName, locale *string) (Profile, error) {
	const q = `
UPDATE profiles
-- Az explicit ::text cast kell, különben a NULL paraméter típusát a
-- Postgres nem tudja levezetni.
SET display_name = COALESCE($2::text, display_name),
    locale       = COALESCE($3::text, locale),
    updated_at   = now()
WHERE subject = $1
RETURNING subject, email, username, display_name, locale, created_at, updated_at`

	var p Profile
	err := s.pool.QueryRow(ctx, q, subject, displayName, locale).Scan(
		&p.Subject, &p.Email, &p.Username, &p.DisplayName, &p.Locale, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return Profile{}, fmt.Errorf("profil frissítés: %w", err)
	}
	return p, nil
}

// ActiveEntitlements a le nem járt jogosultságokat adja vissza.
func (s *Store) ActiveEntitlements(ctx context.Context, subject string) ([]Entitlement, error) {
	const q = `
SELECT feature_key, granted_at, expires_at
FROM entitlements
WHERE subject = $1 AND (expires_at IS NULL OR expires_at > now())
ORDER BY feature_key`

	rows, err := s.pool.Query(ctx, q, subject)
	if err != nil {
		return nil, fmt.Errorf("jogosultságok lekérdezése: %w", err)
	}
	defer rows.Close()

	out := []Entitlement{}
	for rows.Next() {
		var e Entitlement
		if err := rows.Scan(&e.FeatureKey, &e.GrantedAt, &e.ExpiresAt); err != nil {
			return nil, fmt.Errorf("jogosultság sor: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
