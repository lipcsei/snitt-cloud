package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrNotFound akkor jön vissza, ha a kért sor nem létezik. A HTTP réteg
// ebből csinál 404-et, hogy ne kelljen pgx-specifikus hibát ismernie.
var ErrNotFound = errors.New("a keresett rekord nem található")

// ErrConflict ütköző állapotot jelöl (pl. már van élő előfizetése).
var ErrConflict = errors.New("az állapot nem engedi a műveletet")

// Előfizetés-státuszok. Az 'active' és a 'past_due' számít élőnek.
const (
	SubActive   = "active"
	SubPastDue  = "past_due"
	SubCanceled = "canceled"
	SubExpired  = "expired"
)

// Számlastátuszok.
const (
	InvoiceOpen = "open"
	InvoicePaid = "paid"
	InvoiceVoid = "void"
)

// Subscription egy előfizetés sora. A pénzösszeg a pénznem legkisebb
// egységében (fillér/cent) van, egész számként.
type Subscription struct {
	ID                     string     `json:"id"`
	Subject                string     `json:"subject"`
	Plan                   string     `json:"plan"`
	Status                 string     `json:"status"`
	CurrentPeriodStart     time.Time  `json:"current_period_start"`
	CurrentPeriodEnd       time.Time  `json:"current_period_end"`
	PriceMinor             int64      `json:"price_minor"`
	Currency               string     `json:"currency"`
	CancelAtPeriodEnd      bool       `json:"cancel_at_period_end"`
	CanceledAt             *time.Time `json:"canceled_at"`
	Provider               string     `json:"provider"`
	ExternalCustomerID     string     `json:"external_customer_id"`
	ExternalSubscriptionID string     `json:"external_subscription_id"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// Invoice egy számla a saját nyilvántartásban.
type Invoice struct {
	ID                string     `json:"id"`
	Number            string     `json:"number"`
	Subject           string     `json:"subject"`
	SubscriptionID    *string    `json:"subscription_id"`
	AmountMinor       int64      `json:"amount_minor"`
	Currency          string     `json:"currency"`
	Status            string     `json:"status"`
	IssuedAt          time.Time  `json:"issued_at"`
	DueAt             *time.Time `json:"due_at"`
	PaidAt            *time.Time `json:"paid_at"`
	Provider          string     `json:"provider"`
	ExternalInvoiceID string     `json:"external_invoice_id"`
	Note              string     `json:"note"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// NewSubscription egy előfizetés létrehozásának bemenete.
type NewSubscription struct {
	Subject     string
	Plan        string
	PriceMinor  int64
	Currency    string
	PeriodStart time.Time
	PeriodEnd   time.Time
	Provider    string
}

// PlanChange a csomagváltás bemenete. A nil periódus a meglévőt hagyja.
type PlanChange struct {
	Plan        string
	PriceMinor  int64
	Currency    string
	PeriodStart *time.Time
	PeriodEnd   *time.Time
}

// EntitlementSync az előfizetéshez tartozó jogosultságok kívánt állapota.
// A Managed a csomagok által vezérelt összes kulcs: csak ezeket írjuk felül,
// a kézzel adott, csomagon kívüli jogosultságok érintetlenek maradnak.
type EntitlementSync struct {
	Managed   []string
	Grant     []string
	ExpiresAt *time.Time
}

// SubscriptionFilter a listázás szűrői. Az üres mező nem szűr.
type SubscriptionFilter struct {
	Subject string
	Plan    string
	Status  string
	Limit   int
	Offset  int
}

// InvoiceFilter a számlalistázás szűrői.
type InvoiceFilter struct {
	Subject string
	Status  string
	Limit   int
	Offset  int
}

// InvoiceTotals a szűrt számlahalmaz összesítése – az admin ezt nézi először.
type InvoiceTotals struct {
	Count      int   `json:"count"`
	TotalMinor int64 `json:"total_minor"`
	PaidMinor  int64 `json:"paid_minor"`
	OpenMinor  int64 `json:"open_minor"`
}

// PlanCount aktív előfizetések csomagonként.
type PlanCount struct {
	Plan  string `json:"plan"`
	Count int    `json:"count"`
}

// Overview a vezérlőpult összesített számai (a Keycloakból jövő
// felhasználószám nélkül – azt a HTTP réteg teszi hozzá).
type Overview struct {
	ProfileCount            int         `json:"profile_count"`
	ActiveSubscriptionCount int         `json:"active_subscription_count"`
	ActiveByPlan            []PlanCount `json:"active_by_plan"`
	MRRMinor                int64       `json:"mrr_minor"`
	NewProfiles30d          int         `json:"new_profiles_30d"`
	NewSubscriptions30d     int         `json:"new_subscriptions_30d"`
	UnpaidInvoiceCount      int         `json:"unpaid_invoice_count"`
	UnpaidInvoiceTotalMinor int64       `json:"unpaid_invoice_total_minor"`
}

const subscriptionCols = `id, subject, plan, status, current_period_start, current_period_end,
       price_minor, currency, cancel_at_period_end, canceled_at, provider,
       external_customer_id, external_subscription_id, created_at, updated_at`

const invoiceCols = `id, number, subject, subscription_id, amount_minor, currency, status,
       issued_at, due_at, paid_at, provider, external_invoice_id, note, created_at, updated_at`

func scanSubscription(row pgx.Row) (Subscription, error) {
	var s Subscription
	err := row.Scan(&s.ID, &s.Subject, &s.Plan, &s.Status, &s.CurrentPeriodStart, &s.CurrentPeriodEnd,
		&s.PriceMinor, &s.Currency, &s.CancelAtPeriodEnd, &s.CanceledAt, &s.Provider,
		&s.ExternalCustomerID, &s.ExternalSubscriptionID, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func scanInvoice(row pgx.Row) (Invoice, error) {
	var i Invoice
	err := row.Scan(&i.ID, &i.Number, &i.Subject, &i.SubscriptionID, &i.AmountMinor, &i.Currency,
		&i.Status, &i.IssuedAt, &i.DueAt, &i.PaidAt, &i.Provider, &i.ExternalInvoiceID, &i.Note,
		&i.CreatedAt, &i.UpdatedAt)
	return i, err
}

// GetProfile a profil sort adja vissza; ha nincs, ErrNotFound.
func (s *Store) GetProfile(ctx context.Context, subject string) (Profile, error) {
	const q = `SELECT subject, email, username, display_name, locale, created_at, updated_at
FROM profiles WHERE subject = $1`
	var p Profile
	err := s.pool.QueryRow(ctx, q, subject).Scan(
		&p.Subject, &p.Email, &p.Username, &p.DisplayName, &p.Locale, &p.CreatedAt, &p.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Profile{}, ErrNotFound
	}
	if err != nil {
		return Profile{}, fmt.Errorf("profil lekérdezés: %w", err)
	}
	return p, nil
}

// AllEntitlements a lejártakat is visszaadja: az adminnak látnia kell, hogy
// egy jogosultság mikor futott ki.
func (s *Store) AllEntitlements(ctx context.Context, subject string) ([]Entitlement, error) {
	const q = `SELECT feature_key, granted_at, expires_at FROM entitlements
WHERE subject = $1 ORDER BY feature_key`
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

// ProfilesBySubjects egy listányi felhasználó profilját adja vissza egyetlen
// körben – a táblák kirajzolása így nem indít soronkénti lekérdezést.
func (s *Store) ProfilesBySubjects(ctx context.Context, subjects []string) (map[string]Profile, error) {
	out := map[string]Profile{}
	if len(subjects) == 0 {
		return out, nil
	}
	const q = `SELECT subject, email, username, display_name, locale, created_at, updated_at
FROM profiles WHERE subject = ANY($1::text[])`
	rows, err := s.pool.Query(ctx, q, subjects)
	if err != nil {
		return nil, fmt.Errorf("profilok lekérdezése: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.Subject, &p.Email, &p.Username, &p.DisplayName, &p.Locale,
			&p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("profil sor: %w", err)
		}
		out[p.Subject] = p
	}
	return out, rows.Err()
}

// LiveSubscriptionsBySubjects a felhasználók élő előfizetését adja vissza.
// A részleges unique index garantálja, hogy subjectenként legfeljebb egy van.
func (s *Store) LiveSubscriptionsBySubjects(ctx context.Context, subjects []string) (map[string]Subscription, error) {
	out := map[string]Subscription{}
	if len(subjects) == 0 {
		return out, nil
	}
	const q = `SELECT ` + subscriptionCols + ` FROM subscriptions
WHERE subject = ANY($1::text[]) AND status IN ('active', 'past_due')`
	rows, err := s.pool.Query(ctx, q, subjects)
	if err != nil {
		return nil, fmt.Errorf("élő előfizetések lekérdezése: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, fmt.Errorf("előfizetés sor: %w", err)
		}
		out[sub.Subject] = sub
	}
	return out, rows.Err()
}

// ListSubscriptions szűrt, lapozott lista és a szűrésre illeszkedő összes elem száma.
func (s *Store) ListSubscriptions(ctx context.Context, f SubscriptionFilter) ([]Subscription, int, error) {
	const q = `
SELECT ` + subscriptionCols + `, count(*) OVER () AS total
FROM subscriptions
WHERE ($1::text = '' OR subject = $1)
  AND ($2::text = '' OR plan = $2)
  AND ($3::text = '' OR status = $3)
ORDER BY created_at DESC, id
LIMIT $4 OFFSET $5`

	rows, err := s.pool.Query(ctx, q, f.Subject, f.Plan, f.Status, f.Limit, f.Offset)
	if err != nil {
		return nil, 0, fmt.Errorf("előfizetések lekérdezése: %w", err)
	}
	defer rows.Close()

	out := []Subscription{}
	total := 0
	for rows.Next() {
		var sub Subscription
		if err := rows.Scan(&sub.ID, &sub.Subject, &sub.Plan, &sub.Status, &sub.CurrentPeriodStart,
			&sub.CurrentPeriodEnd, &sub.PriceMinor, &sub.Currency, &sub.CancelAtPeriodEnd,
			&sub.CanceledAt, &sub.Provider, &sub.ExternalCustomerID, &sub.ExternalSubscriptionID,
			&sub.CreatedAt, &sub.UpdatedAt, &total); err != nil {
			return nil, 0, fmt.Errorf("előfizetés sor: %w", err)
		}
		out = append(out, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("előfizetés sorok: %w", err)
	}
	// Üres találathalmaznál a window function nem fut le, a total 0 marad.
	return out, total, nil
}

// GetSubscription egy előfizetés azonosító alapján.
func (s *Store) GetSubscription(ctx context.Context, id string) (Subscription, error) {
	sub, err := scanSubscription(s.pool.QueryRow(ctx, `SELECT `+subscriptionCols+` FROM subscriptions WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Subscription{}, ErrNotFound
	}
	if err != nil {
		return Subscription{}, fmt.Errorf("előfizetés lekérdezés: %w", err)
	}
	return sub, nil
}

// GrantSubscription új előfizetést hoz létre, és ugyanabban a tranzakcióban
// szinkronizálja a jogosultságokat. A kettő nem csúszhat szét: az entitlements
// tábla az, amiből a desktop app az AI funkciókat feloldja.
func (s *Store) GrantSubscription(ctx context.Context, in NewSubscription, sync EntitlementSync) (Subscription, error) {
	var out Subscription
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		const q = `
INSERT INTO subscriptions (id, subject, plan, status, current_period_start, current_period_end,
                           price_minor, currency, provider)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING ` + subscriptionCols

		id := newID()
		sub, err := scanSubscription(tx.QueryRow(ctx, q, id, in.Subject, in.Plan, SubActive,
			in.PeriodStart, in.PeriodEnd, in.PriceMinor, in.Currency, in.Provider))
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return fmt.Errorf("%w: a felhasználónak már van élő előfizetése", ErrConflict)
			}
			return fmt.Errorf("előfizetés létrehozás: %w", err)
		}
		out = sub
		return syncEntitlements(ctx, tx, in.Subject, sync)
	})
	return out, err
}

// ChangeSubscriptionPlan csomagot (és árat) vált egy élő előfizetésen.
func (s *Store) ChangeSubscriptionPlan(ctx context.Context, id string, ch PlanChange, sync EntitlementSync) (Subscription, error) {
	var out Subscription
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		const q = `
UPDATE subscriptions
SET plan                 = $2,
    price_minor          = $3,
    currency             = $4,
    current_period_start = COALESCE($5::timestamptz, current_period_start),
    current_period_end   = COALESCE($6::timestamptz, current_period_end),
    updated_at           = now()
WHERE id = $1 AND status IN ('active', 'past_due')
RETURNING ` + subscriptionCols

		sub, err := scanSubscription(tx.QueryRow(ctx, q, id, ch.Plan, ch.PriceMinor, ch.Currency,
			ch.PeriodStart, ch.PeriodEnd))
		if errors.Is(err, pgx.ErrNoRows) {
			// Vagy nincs ilyen sor, vagy már nem él – a hívó a GET-tel tudja szétválasztani.
			return fmt.Errorf("%w: csak élő előfizetésen lehet csomagot váltani", ErrConflict)
		}
		if err != nil {
			return fmt.Errorf("csomagváltás: %w", err)
		}
		out = sub
		return syncEntitlements(ctx, tx, sub.Subject, sync)
	})
	return out, err
}

// CancelSubscription lemondás azonnal vagy a periódus végén. Azonnali
// lemondásnál a jogosultságok is azonnal elvesznek; periódus végi lemondásnál
// megmaradnak, mert a lejáratuk eleve a periódus vége.
func (s *Store) CancelSubscription(ctx context.Context, id string, atPeriodEnd bool, sync EntitlementSync) (Subscription, error) {
	var out Subscription
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		const q = `
UPDATE subscriptions
SET status               = CASE WHEN $2 THEN status ELSE 'canceled' END,
    cancel_at_period_end = $2,
    canceled_at          = now(),
    updated_at           = now()
WHERE id = $1 AND status IN ('active', 'past_due')
RETURNING ` + subscriptionCols

		sub, err := scanSubscription(tx.QueryRow(ctx, q, id, atPeriodEnd))
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w: csak élő előfizetés mondható le", ErrConflict)
		}
		if err != nil {
			return fmt.Errorf("lemondás: %w", err)
		}
		out = sub
		return syncEntitlements(ctx, tx, sub.Subject, sync)
	})
	return out, err
}

// ReactivateSubscription visszakapcsol egy lemondott vagy periódus végén
// lejáró előfizetést. A nil periódus a meglévőt hagyja.
func (s *Store) ReactivateSubscription(ctx context.Context, id string, periodStart, periodEnd *time.Time, sync EntitlementSync) (Subscription, error) {
	var out Subscription
	err := s.inTx(ctx, func(tx pgx.Tx) error {
		const q = `
UPDATE subscriptions
SET status               = 'active',
    cancel_at_period_end = false,
    canceled_at          = NULL,
    current_period_start = COALESCE($2::timestamptz, current_period_start),
    current_period_end   = COALESCE($3::timestamptz, current_period_end),
    updated_at           = now()
WHERE id = $1
RETURNING ` + subscriptionCols

		sub, err := scanSubscription(tx.QueryRow(ctx, q, id, periodStart, periodEnd))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return fmt.Errorf("%w: a felhasználónak már van másik élő előfizetése", ErrConflict)
			}
			return fmt.Errorf("visszakapcsolás: %w", err)
		}
		out = sub
		return syncEntitlements(ctx, tx, sub.Subject, sync)
	})
	return out, err
}

// syncEntitlements a csomagok által vezérelt jogosultságokat állítja be:
// előbb törli az összes vezérelt kulcsot, majd beszúrja a most járókat.
func syncEntitlements(ctx context.Context, tx pgx.Tx, subject string, sync EntitlementSync) error {
	if len(sync.Managed) == 0 {
		return nil
	}
	if _, err := tx.Exec(ctx,
		`DELETE FROM entitlements WHERE subject = $1 AND feature_key = ANY($2::text[])`,
		subject, sync.Managed); err != nil {
		return fmt.Errorf("jogosultságok törlése: %w", err)
	}
	for _, key := range sync.Grant {
		if _, err := tx.Exec(ctx,
			`INSERT INTO entitlements (subject, feature_key, granted_at, expires_at)
             VALUES ($1, $2, now(), $3)`,
			subject, key, sync.ExpiresAt); err != nil {
			return fmt.Errorf("jogosultság kiadása (%s): %w", key, err)
		}
	}
	return nil
}

// ListInvoices szűrt, lapozott számlalista, a szűrésre illeszkedő
// összesítéssel együtt (az admin a végösszeget nézi elsőként).
func (s *Store) ListInvoices(ctx context.Context, f InvoiceFilter) ([]Invoice, InvoiceTotals, error) {
	const q = `
SELECT ` + invoiceCols + `
FROM invoices
WHERE ($1::text = '' OR subject = $1)
  AND ($2::text = '' OR status = $2)
ORDER BY issued_at DESC, id
LIMIT $3 OFFSET $4`

	rows, err := s.pool.Query(ctx, q, f.Subject, f.Status, f.Limit, f.Offset)
	if err != nil {
		return nil, InvoiceTotals{}, fmt.Errorf("számlák lekérdezése: %w", err)
	}
	defer rows.Close()

	out := []Invoice{}
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, InvoiceTotals{}, fmt.Errorf("számla sor: %w", err)
		}
		out = append(out, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, InvoiceTotals{}, fmt.Errorf("számla sorok: %w", err)
	}

	const totalsQ = `
SELECT count(*),
       COALESCE(sum(amount_minor), 0),
       COALESCE(sum(amount_minor) FILTER (WHERE status = 'paid'), 0),
       COALESCE(sum(amount_minor) FILTER (WHERE status = 'open'), 0)
FROM invoices
WHERE ($1::text = '' OR subject = $1)
  AND ($2::text = '' OR status = $2)`

	var t InvoiceTotals
	if err := s.pool.QueryRow(ctx, totalsQ, f.Subject, f.Status).
		Scan(&t.Count, &t.TotalMinor, &t.PaidMinor, &t.OpenMinor); err != nil {
		return nil, InvoiceTotals{}, fmt.Errorf("számla összesítés: %w", err)
	}
	return out, t, nil
}

// GetInvoice egy számla azonosító alapján.
func (s *Store) GetInvoice(ctx context.Context, id string) (Invoice, error) {
	inv, err := scanInvoice(s.pool.QueryRow(ctx, `SELECT `+invoiceCols+` FROM invoices WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Invoice{}, ErrNotFound
	}
	if err != nil {
		return Invoice{}, fmt.Errorf("számla lekérdezés: %w", err)
	}
	return inv, nil
}

// NewInvoice számla kiállításának bemenete.
type NewInvoice struct {
	Subject        string
	SubscriptionID *string
	AmountMinor    int64
	Currency       string
	DueAt          *time.Time
	Provider       string
	Note           string
}

// CreateInvoice nyitott (open) számlát állít ki. Terhelés nem történik:
// a számlaszámot mi generáljuk, a pénz beérkezését az admin jelöli.
func (s *Store) CreateInvoice(ctx context.Context, in NewInvoice) (Invoice, error) {
	const q = `
INSERT INTO invoices (id, number, subject, subscription_id, amount_minor, currency, status,
                      issued_at, due_at, provider, note)
VALUES ($1,
        'SNITT-' || to_char(now(), 'YYYY') || '-' || lpad(nextval('invoice_number_seq')::text, 6, '0'),
        $2, $3, $4, $5, 'open', now(), $6, $7, $8)
RETURNING ` + invoiceCols

	inv, err := scanInvoice(s.pool.QueryRow(ctx, q, newID(), in.Subject, in.SubscriptionID,
		in.AmountMinor, in.Currency, in.DueAt, in.Provider, in.Note))
	if err != nil {
		return Invoice{}, fmt.Errorf("számla kiállítás: %w", err)
	}
	return inv, nil
}

// MarkInvoicePaid kifizetettre állít egy nyitott számlát. Ezt ma ember hívja
// az admin felületről; később ugyanezt fogja hívni a szolgáltató webhookja.
func (s *Store) MarkInvoicePaid(ctx context.Context, id string, paidAt time.Time, externalID string) (Invoice, error) {
	const q = `
UPDATE invoices
SET status              = 'paid',
    paid_at             = $2,
    external_invoice_id = CASE WHEN $3::text = '' THEN external_invoice_id ELSE $3 END,
    updated_at          = now()
WHERE id = $1 AND status = 'open'
RETURNING ` + invoiceCols

	inv, err := scanInvoice(s.pool.QueryRow(ctx, q, id, paidAt, externalID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Invoice{}, fmt.Errorf("%w: csak nyitott számla jelölhető kifizetettnek", ErrConflict)
	}
	if err != nil {
		return Invoice{}, fmt.Errorf("számla kifizetés: %w", err)
	}
	return inv, nil
}

// VoidInvoice sztornóz egy nyitott számlát. Kifizetett számla nem sztornózható:
// azt már csak jóváíró számlával lehet visszavonni.
func (s *Store) VoidInvoice(ctx context.Context, id, note string) (Invoice, error) {
	const q = `
UPDATE invoices
SET status     = 'void',
    paid_at    = NULL,
    note       = CASE WHEN $2::text = '' THEN note ELSE $2 END,
    updated_at = now()
WHERE id = $1 AND status = 'open'
RETURNING ` + invoiceCols

	inv, err := scanInvoice(s.pool.QueryRow(ctx, q, id, note))
	if errors.Is(err, pgx.ErrNoRows) {
		return Invoice{}, fmt.Errorf("%w: csak nyitott számla sztornózható", ErrConflict)
	}
	if err != nil {
		return Invoice{}, fmt.Errorf("számla sztornó: %w", err)
	}
	return inv, nil
}

// Overview a vezérlőpult összesítései egy körben.
func (s *Store) Overview(ctx context.Context) (Overview, error) {
	var o Overview

	const totals = `
SELECT (SELECT count(*) FROM profiles),
       (SELECT count(*) FROM profiles WHERE created_at > now() - interval '30 days'),
       (SELECT count(*) FROM subscriptions WHERE status IN ('active', 'past_due')),
       (SELECT COALESCE(sum(price_minor), 0) FROM subscriptions WHERE status IN ('active', 'past_due')),
       (SELECT count(*) FROM subscriptions WHERE created_at > now() - interval '30 days'),
       (SELECT count(*) FROM invoices WHERE status = 'open'),
       (SELECT COALESCE(sum(amount_minor), 0) FROM invoices WHERE status = 'open')`

	if err := s.pool.QueryRow(ctx, totals).Scan(&o.ProfileCount, &o.NewProfiles30d,
		&o.ActiveSubscriptionCount, &o.MRRMinor, &o.NewSubscriptions30d,
		&o.UnpaidInvoiceCount, &o.UnpaidInvoiceTotalMinor); err != nil {
		return Overview{}, fmt.Errorf("összesítők lekérdezése: %w", err)
	}

	rows, err := s.pool.Query(ctx, `
SELECT plan, count(*) FROM subscriptions
WHERE status IN ('active', 'past_due')
GROUP BY plan ORDER BY plan`)
	if err != nil {
		return Overview{}, fmt.Errorf("csomagonkénti bontás: %w", err)
	}
	defer rows.Close()

	o.ActiveByPlan = []PlanCount{}
	for rows.Next() {
		var pc PlanCount
		if err := rows.Scan(&pc.Plan, &pc.Count); err != nil {
			return Overview{}, fmt.Errorf("csomag sor: %w", err)
		}
		o.ActiveByPlan = append(o.ActiveByPlan, pc)
	}
	if err := rows.Err(); err != nil {
		return Overview{}, fmt.Errorf("csomag sorok: %w", err)
	}
	return o, nil
}

// inTx a callbacket tranzakcióban futtatja; hiba esetén visszagörget.
func (s *Store) inTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("tranzakció indítás: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("tranzakció commit: %w", err)
	}
	return nil
}

// newID 128 bites véletlen azonosító. Nem kell hozzá pgcrypto kiterjesztés,
// és az azonosító nem árul el sorrendet vagy darabszámot.
func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// A crypto/rand olvasása Go 1.24 óta nem hibázhat; ha mégis,
		// nincs értelme félkész azonosítóval továbbmenni.
		panic("crypto/rand olvasás sikertelen: " + err.Error())
	}
	return hex.EncodeToString(b[:])
}
