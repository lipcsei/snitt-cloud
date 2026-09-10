// Package billing az előfizetési csomagok leírása és a későbbi fizetési
// szolgáltató (Stripe/Paddle) becsatlakozási pontja.
//
// # A fizetési szolgáltató illesztési pontja
//
// Ma egyetlen implementáció létezik: a Manual, ami semmilyen külső hívást nem
// végez. Az admin kézzel ad előfizetést és kézzel jelöl kifizetettnek egy
// számlát – valódi terhelés SEHOL nem történik. A subscriptions és invoices
// táblák viszont már hordozzák a provider + external_* mezőket, így egy valódi
// szolgáltató bekötése ennyi:
//
//  1. új Provider implementáció (pl. stripe.Provider) ebben a csomagban,
//  2. a httpapi.API a konfigurációtól függően ezt kapja meg a Manual helyett,
//  3. a szolgáltató webhookja egy új végponton hívja a store már meglévő
//     MarkInvoicePaid / UpdateSubscription függvényeit.
//
// A HTTP réteg szándékosan sosem hív közvetlenül szolgáltató-specifikus kódot,
// csak ezt az interfészt – ez a varrat.
package billing

import (
	"context"
	"fmt"
	"sort"
)

// Plan egy megvásárolható csomag. A FeatureKeys pontosan azok az
// entitlementek, amiket az aktív előfizetés a desktop appban felold.
type Plan struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PriceMinor  int64    `json:"price_minor"`
	FeatureKeys []string `json:"feature_keys"`
}

// A csomagok kódban vannak, nem adatbázisban: kevés van belőlük, ritkán
// változnak, és így a hozzájuk tartozó entitlementek sem csúszhatnak el.
var plans = map[string]Plan{
	"pro": {
		Key:         "pro",
		Name:        "Pro",
		Description: "Szemantikus keresés a saját videótárban, AI-alapú találati rangsorral.",
		PriceMinor:  299000, // 2 990 Ft
		FeatureKeys: []string{"ai-semantic-search"},
	},
	"studio": {
		Key:         "studio",
		Name:        "Studio",
		Description: "A Pro mindene, plusz videótartalom-értelmezés (jelenet- és arcfelismerés).",
		PriceMinor:  799000, // 7 990 Ft
		FeatureKeys: []string{"ai-semantic-search", "ai-video-understanding"},
	},
}

// Plans a választható csomagok, kulcs szerint rendezve.
func Plans() []Plan {
	out := make([]Plan, 0, len(plans))
	for _, p := range plans {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PriceMinor < out[j].PriceMinor })
	return out
}

// LookupPlan a csomag kulcs alapján.
func LookupPlan(key string) (Plan, bool) {
	p, ok := plans[key]
	return p, ok
}

// ManagedFeatureKeys az összes csomag által vezérelt entitlement kulcs.
// A szinkronizáció csak ezeket írja felül, a kézzel adott, csomagtól független
// jogosultságokat (pl. béta-hozzáférés) érintetlenül hagyja.
func ManagedFeatureKeys() []string {
	set := map[string]struct{}{}
	for _, p := range plans {
		for _, f := range p.FeatureKeys {
			set[f] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for f := range set {
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

// Charge egy tervezett terhelés leírása. Ma csak a ledger sorának adata.
type Charge struct {
	Subject     string
	PlanKey     string
	AmountMinor int64
	Currency    string
}

// Provider a fizetési szolgáltató illesztője. Amíg nincs valódi szolgáltató,
// a Manual implementáció felel meg neki.
type Provider interface {
	// Name a provider oszlopba kerülő azonosító ("manual", később "stripe").
	Name() string
	// EnsureCustomer a szolgáltató oldali ügyfélazonosítót adja vissza.
	EnsureCustomer(ctx context.Context, subject, email string) (string, error)
	// StartSubscription a szolgáltató oldali előfizetés-azonosítót adja vissza.
	StartSubscription(ctx context.Context, ch Charge) (string, error)
	// CancelSubscription lemondja a szolgáltató oldali előfizetést.
	CancelSubscription(ctx context.Context, externalID string, atPeriodEnd bool) error
}

// Manual a jelenlegi, szolgáltató nélküli működés: minden lépés kézi admin
// döntés, külső rendszer nincs. Szándékosan nem szimulál fizetést.
type Manual struct{}

// Name a "manual" azonosító.
func (Manual) Name() string { return "manual" }

// EnsureCustomer üres azonosítót ad: kézi működésnél nincs szolgáltató-oldali ügyfél.
func (Manual) EnsureCustomer(context.Context, string, string) (string, error) { return "", nil }

// StartSubscription üres azonosítót ad: kézi működésnél nincs külső előfizetés.
func (Manual) StartSubscription(context.Context, Charge) (string, error) { return "", nil }

// CancelSubscription kézi működésnél nincs mit lemondani a külső rendszerben.
func (Manual) CancelSubscription(_ context.Context, externalID string, _ bool) error {
	if externalID != "" {
		// Ha valaha kerül külső azonosító a sorba, azt már nem a kézi
		// üzemmódnak kell lezárnia – ez inkább konfigurációs hiba.
		return fmt.Errorf("külső előfizetés (%s) kézi módban nem mondható le", externalID)
	}
	return nil
}
