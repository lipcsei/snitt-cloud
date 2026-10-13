package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// AI használat fajtái (ai_usage.kind).
const (
	AIUsageEmbed  = "embed"
	AIUsageAnswer = "answer"
)

// AIWallet egy felhasználó AI tárcája. A nulla érték a "még sosem vásárolt"
// tárca: minden egyenleg 0, nincs AI előfizetés.
type AIWallet struct {
	Subject                string
	PeriodTokens           int64
	PeriodEnd              *time.Time
	PurchasedTokens        int64
	ExternalSubscriptionID string
	// Status az AI előfizetés állapota a subscriptions státuszkészletével
	// (active, past_due, canceled, expired); üres, ha sosem volt.
	Status    string
	UpdatedAt time.Time
}

// SubscriptionActive igaz, ha a havi AI keret most költhető: az előfizetés
// aktív, és a periódus még nem járt le. Késedelmes fizetésnél (past_due) a
// keret nem költhető - a vásárolt tokenek viszont igen.
func (w AIWallet) SubscriptionActive(now time.Time) bool {
	return w.Status == SubActive && w.PeriodEnd != nil && w.PeriodEnd.After(now)
}

// ValidPeriodTokens a most költhető havi keret.
func (w AIWallet) ValidPeriodTokens(now time.Time) int64 {
	if !w.SubscriptionActive(now) || w.PeriodTokens < 0 {
		return 0
	}
	return w.PeriodTokens
}

// Total a most költhető teljes egyenleg.
func (w AIWallet) Total(now time.Time) int64 {
	purchased := w.PurchasedTokens
	if purchased < 0 {
		purchased = 0
	}
	return w.ValidPeriodTokens(now) + purchased
}

// Debit levonja az amount tokent: előbb a havi keretből (az úgyis elveszne
// a periódus végén), aztán a vásárolt egyenlegből. Nulla alá sosem megy: ha
// az utolsó kérés többe került, mint ami maradt, a különbözetet elnyeljük -
// a kérés előtti egyenleg-ellenőrzés miatt ez legfeljebb egy kérésnyi.
// Lejárt keretet nem érint (az a következő megújuláskor úgyis felülíródik).
func (w AIWallet) Debit(now time.Time, amount int64) AIWallet {
	if amount <= 0 {
		return w
	}
	if period := w.ValidPeriodTokens(now); period > 0 {
		take := min(period, amount)
		w.PeriodTokens -= take
		amount -= take
	}
	if amount > 0 && w.PurchasedTokens > 0 {
		w.PurchasedTokens -= min(w.PurchasedTokens, amount)
	}
	return w
}

// NewAIUsage egy AI szolgáltatóhívás elszámolása.
type NewAIUsage struct {
	Kind          string
	Model         string
	InputTokens   int64
	OutputTokens  int64
	ChargedTokens int64
}

// AISubscription a havi AI csomag (újra)indítása egy tárcán.
type AISubscription struct {
	ExternalSubscriptionID string
	Status                 string
	PeriodTokens           int64
	PeriodEnd              time.Time
}

// AISubscriptionUpdate a fizetési szolgáltató eseményéből jövő változás. A
// nulla értékű mező változatlanul hagyja az oszlopot.
type AISubscriptionUpdate struct {
	Status       string
	PeriodEnd    *time.Time
	PeriodTokens *int64
}

const aiWalletCols = `subject, period_tokens, period_end, purchased_tokens,
COALESCE(external_subscription_id, ''), status, updated_at`

func scanAIWallet(row pgx.Row) (AIWallet, error) {
	var w AIWallet
	err := row.Scan(&w.Subject, &w.PeriodTokens, &w.PeriodEnd, &w.PurchasedTokens,
		&w.ExternalSubscriptionID, &w.Status, &w.UpdatedAt)
	return w, err
}

// AIWallet a felhasználó AI tárcája; ha még nincs, a nulla értékű tárca
// (hiba nélkül).
func (s *Store) AIWallet(ctx context.Context, subject string) (AIWallet, error) {
	w, err := scanAIWallet(s.pool.QueryRow(ctx,
		`SELECT `+aiWalletCols+` FROM ai_wallets WHERE subject = $1`, subject))
	if errors.Is(err, pgx.ErrNoRows) {
		return AIWallet{Subject: subject}, nil
	}
	if err != nil {
		return AIWallet{}, fmt.Errorf("AI tárca lekérdezése: %w", err)
	}
	return w, nil
}

// AIWalletByExternalSubscription a Stripe előfizetés-azonosító alapján
// keresi a tárcát; ha nincs, ErrNotFound.
func (s *Store) AIWalletByExternalSubscription(ctx context.Context, externalID string) (AIWallet, error) {
	if externalID == "" {
		return AIWallet{}, ErrNotFound
	}
	w, err := scanAIWallet(s.pool.QueryRow(ctx,
		`SELECT `+aiWalletCols+` FROM ai_wallets WHERE external_subscription_id = $1`, externalID))
	if errors.Is(err, pgx.ErrNoRows) {
		return AIWallet{}, ErrNotFound
	}
	if err != nil {
		return AIWallet{}, fmt.Errorf("AI tárca lekérdezése külső azonosítóval: %w", err)
	}
	return w, nil
}

// DebitAITokens egy tranzakcióban zárolja a tárcát, levonja a felhasznált
// tokeneket (lásd AIWallet.Debit), és naplózza a használatot. A sorzár
// (FOR UPDATE) miatt két párhuzamos kérés nem költheti el kétszer ugyanazt.
func (s *Store) DebitAITokens(ctx context.Context, subject string, u NewAIUsage) (AIWallet, error) {
	var out AIWallet
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		// Ha még nincs tárca, üres sort hozunk létre, hogy legyen mit zárolni.
		if _, err := tx.Exec(ctx,
			`INSERT INTO ai_wallets (subject) VALUES ($1) ON CONFLICT (subject) DO NOTHING`, subject); err != nil {
			return fmt.Errorf("AI tárca létrehozása: %w", err)
		}
		w, err := scanAIWallet(tx.QueryRow(ctx,
			`SELECT `+aiWalletCols+` FROM ai_wallets WHERE subject = $1 FOR UPDATE`, subject))
		if err != nil {
			return fmt.Errorf("AI tárca zárolása: %w", err)
		}
		// Az időt az adatbázis és a Go oldal közt nem keverjük: a lejárat
		// vizsgálata itt, egyetlen "most" értékkel történik.
		w = w.Debit(time.Now(), u.ChargedTokens)
		out, err = scanAIWallet(tx.QueryRow(ctx, `
UPDATE ai_wallets
SET period_tokens = $2, purchased_tokens = $3, updated_at = now()
WHERE subject = $1
RETURNING `+aiWalletCols, subject, w.PeriodTokens, w.PurchasedTokens))
		if err != nil {
			return fmt.Errorf("AI tokenek levonása: %w", err)
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO ai_usage (subject, kind, model, input_tokens, output_tokens, charged_tokens)
VALUES ($1, $2, $3, $4, $5, $6)`,
			subject, u.Kind, u.Model, u.InputTokens, u.OutputTokens, u.ChargedTokens); err != nil {
			return fmt.Errorf("AI használat naplózása: %w", err)
		}
		return nil
	})
	return out, err
}

// CreditAITokens a vásárolt (soha le nem járó) egyenleghez ad tokeneket.
func (s *Store) CreditAITokens(ctx context.Context, subject string, tokens int64) (AIWallet, error) {
	w, err := scanAIWallet(s.pool.QueryRow(ctx, `
INSERT INTO ai_wallets (subject, purchased_tokens) VALUES ($1, $2)
ON CONFLICT (subject) DO UPDATE
SET purchased_tokens = ai_wallets.purchased_tokens + EXCLUDED.purchased_tokens,
    updated_at       = now()
RETURNING `+aiWalletCols, subject, tokens))
	if err != nil {
		return AIWallet{}, fmt.Errorf("AI tokenek jóváírása: %w", err)
	}
	return w, nil
}

// SetAISubscription elindítja (vagy egy új Stripe előfizetésre cseréli) a
// havi AI csomagot: a keret a csomag teljes kerete, a vásárolt egyenleg
// érintetlen marad.
func (s *Store) SetAISubscription(ctx context.Context, subject string, in AISubscription) (AIWallet, error) {
	w, err := scanAIWallet(s.pool.QueryRow(ctx, `
INSERT INTO ai_wallets (subject, external_subscription_id, status, period_tokens, period_end)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (subject) DO UPDATE
SET external_subscription_id = EXCLUDED.external_subscription_id,
    status                   = EXCLUDED.status,
    period_tokens            = EXCLUDED.period_tokens,
    period_end               = EXCLUDED.period_end,
    updated_at               = now()
RETURNING `+aiWalletCols,
		subject, in.ExternalSubscriptionID, in.Status, in.PeriodTokens, in.PeriodEnd))
	if err != nil {
		return AIWallet{}, fmt.Errorf("AI előfizetés rögzítése: %w", err)
	}
	return w, nil
}

// UpdateAISubscription a Stripe előfizetés-azonosító alapján frissíti a
// tárca AI előfizetését (állapot, periódus vége, megújuláskor a keret
// visszaállítása). Ismeretlen azonosítónál ErrNotFound.
func (s *Store) UpdateAISubscription(ctx context.Context, externalID string, upd AISubscriptionUpdate) (AIWallet, error) {
	if externalID == "" {
		return AIWallet{}, ErrNotFound
	}
	w, err := scanAIWallet(s.pool.QueryRow(ctx, `
UPDATE ai_wallets
-- Az explicit castok kellenek, különben a NULL paraméter típusát a
-- Postgres nem tudja levezetni.
SET status        = COALESCE(NULLIF($2::text, ''), status),
    period_end    = COALESCE($3::timestamptz, period_end),
    period_tokens = COALESCE($4::bigint, period_tokens),
    updated_at    = now()
WHERE external_subscription_id = $1
RETURNING `+aiWalletCols, externalID, upd.Status, upd.PeriodEnd, upd.PeriodTokens))
	if errors.Is(err, pgx.ErrNoRows) {
		return AIWallet{}, ErrNotFound
	}
	if err != nil {
		return AIWallet{}, fmt.Errorf("AI előfizetés frissítése: %w", err)
	}
	return w, nil
}
