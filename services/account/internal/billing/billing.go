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
	"time"
)

// FeatureAppAccess az alkalmazás korlátlan használata. A desktop app
// ingyenes kreditkerettel indul (importálásonként fogy); ha elfogyott, csak
// ennek a jogosultságnak a birtokában használható tovább. Minden fizetős
// csomag tartalmazza.
const FeatureAppAccess = "snitt-app-access"

// KindAI az AI tokencsomagok fajtája. Az AI csomag nem előfizetés a
// subscriptions tábla értelmében: nem ad entitlementet, nem foglalja el a
// felhasználó egyetlen élő előfizetésének helyét, hanem az AI tárcát
// (ai_wallets) tölti fel. A kettő függetlenül vehető.
const KindAI = "ai"

// Plan egy megvásárolható csomag. A FeatureKeys pontosan azok az
// entitlementek, amiket az aktív előfizetés a desktop appban felold.
type Plan struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PriceMinor  int64    `json:"price_minor"`
	FeatureKeys []string `json:"feature_keys"`
	// Lifetime: egyszeri vásárlás, ami soha nem jár le. Ugyanúgy egy
	// subscriptions sor képviseli (így a "legfeljebb egy élő előfizetés"
	// szabály és az entitlement-szinkron ugyanaz marad), csak a periódus vége
	// a távoli jövő, és a jogosultságnak nincs lejárata.
	Lifetime bool `json:"lifetime"`
	// SelfService: a felhasználó maga is megveheti (checkout). A többi csomagot
	// ma csak az admin adhatja ki.
	SelfService bool `json:"self_service"`
	// Kind a csomag fajtája: üres a hozzáférést adó előfizetéseknél, KindAI
	// az AI tokencsomagoknál.
	Kind string `json:"kind"`
	// AITokens AI csomagnál: ennyi Snitt AI tokent ad. Havi AI csomagnál
	// elszámolási periódusonként ennyi a keret (nem görgetjük át), egyszeri
	// tokencsomagnál ennyi kerül a soha le nem járó vásárolt egyenlegre.
	AITokens int64 `json:"ai_tokens"`
}

// IsAI igaz, ha a csomag az AI tárcát tölti, nem előfizetést ad.
func (p Plan) IsAI() bool { return p.Kind == KindAI }

// lifetimeYears: ennyi évre nyitunk periódust a lifetime csomagnak. A séma a
// periódus végét kötelezővé teszi; a "soha" helyett egy elérhetetlen dátum
// egyszerűbb, mint egy NULL-t kezelni minden lekérdezésben.
const lifetimeYears = 100

// PeriodEnd a csomag elszámolási periódusának vége, ha most indul: lifetime
// csomagnál a távoli jövő, egyébként a megadott napszám.
func (p Plan) PeriodEnd(start time.Time, days int) time.Time {
	if p.Lifetime {
		return start.AddDate(lifetimeYears, 0, 0)
	}
	return start.Add(time.Duration(days) * 24 * time.Hour)
}

// EntitlementExpiry a csomagból következő jogosultságok lejárata: lifetime
// csomagnál nincs (nil), egyébként a periódus vége.
func (p Plan) EntitlementExpiry(periodEnd time.Time) *time.Time {
	if p.Lifetime {
		return nil
	}
	return &periodEnd
}

// A csomagok kódban vannak, nem adatbázisban: kevés van belőlük, ritkán
// változnak, és így a hozzájuk tartozó entitlementek sem csúszhatnak el.
var plans = map[string]Plan{
	"monthly": {
		Key:         "monthly",
		Name:        "Előfizetés",
		Description: "Korlátlan importálás és használat, havonta megújuló előfizetéssel.",
		PriceMinor:  149000, // 1 490 Ft / hó
		FeatureKeys: []string{FeatureAppAccess},
		SelfService: true,
	},
	"lifetime": {
		Key:         "lifetime",
		Name:        "Lifetime",
		Description: "Korlátlan importálás és használat, egyszeri díjért, örökre.",
		PriceMinor:  2499000, // 24 990 Ft egyszeri
		FeatureKeys: []string{FeatureAppAccess},
		Lifetime:    true,
		SelfService: true,
	},
	"pro": {
		Key:         "pro",
		Name:        "Pro",
		Description: "Korlátlan használat, plusz szemantikus keresés a saját videótárban, AI-alapú találati rangsorral.",
		PriceMinor:  299000, // 2 990 Ft
		FeatureKeys: []string{FeatureAppAccess, "ai-semantic-search"},
	},
	"studio": {
		Key:         "studio",
		Name:        "Studio",
		Description: "A Pro mindene, plusz videótartalom-értelmezés (jelenet- és arcfelismerés).",
		PriceMinor:  799000, // 7 990 Ft
		FeatureKeys: []string{FeatureAppAccess, "ai-semantic-search", "ai-video-understanding"},
	},
	// Az AI csomagok a felhőben futó AI funkciókat (szemantikus keresés
	// beágyazással, válasz a saját videótárból) fizetik. A Lifetime itt az
	// egyszeri fizetést jelöli (mode=payment), nem örök periódust.
	"ai_monthly": {
		Key:         "ai_monthly",
		Name:        "AI csomag",
		Description: "Havi 1 000 000 Snitt AI token okos kereséshez és válaszokhoz a saját videótáradból.",
		PriceMinor:  299000, // 2 990 Ft / hó
		SelfService: true,
		Kind:        KindAI,
		AITokens:    1_000_000,
	},
	"ai_tokens": {
		Key:         "ai_tokens",
		Name:        "AI tokencsomag",
		Description: "500 000 Snitt AI token egyszeri díjért, lejárat nélkül.",
		PriceMinor:  199000, // 1 990 Ft egyszeri
		Lifetime:    true,
		SelfService: true,
		Kind:        KindAI,
		AITokens:    500_000,
	},
}

// Plans az előfizetésként kiadható (hozzáférést adó) csomagok, ár szerint
// rendezve. Az AI csomagok nincsenek benne: azokat az admin nem adhatja ki
// előfizetésként, lásd AllPlans.
func Plans() []Plan {
	out := make([]Plan, 0, len(plans))
	for _, p := range plans {
		if !p.IsAI() {
			out = append(out, p)
		}
	}
	sortPlans(out)
	return out
}

// AllPlans minden csomag, az AI csomagokkal együtt, ár szerint rendezve.
func AllPlans() []Plan {
	out := make([]Plan, 0, len(plans))
	for _, p := range plans {
		out = append(out, p)
	}
	sortPlans(out)
	return out
}

func sortPlans(out []Plan) {
	sort.Slice(out, func(i, j int) bool {
		if out[i].PriceMinor != out[j].PriceMinor {
			return out[i].PriceMinor < out[j].PriceMinor
		}
		return out[i].Key < out[j].Key
	})
}

// LookupPlan az előfizetésként kiadható csomag kulcs alapján. AI csomagra
// nem talál: így az előfizetést kezelő kódutak (admin kiadás, csomagváltás,
// webhook szinkron) sosem írhatnak AI csomagot a subscriptions táblába.
func LookupPlan(key string) (Plan, bool) {
	p, ok := plans[key]
	if !ok || p.IsAI() {
		return Plan{}, false
	}
	return p, true
}

// LookupAnyPlan bármely csomag kulcs alapján, az AI csomagokat is beleértve.
func LookupAnyPlan(key string) (Plan, bool) {
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

// CheckoutRequest a felhasználó által indított vásárlás adatai.
type CheckoutRequest struct {
	Subject string
	Email   string
	Plan    Plan
	// SuccessURL / CancelURL: ide tér vissza a böngésző a fizetés után.
	SuccessURL string
	CancelURL  string
}

// CheckoutProvider az önkiszolgáló vásárlást támogató szolgáltató: a
// felhasználót a szolgáltató hosztolt fizetési oldalára küldjük, a fizetés
// eredménye pedig webhookon érkezik vissza. A Manual nem valósítja meg -
// nála nincs mit fizetni.
type CheckoutProvider interface {
	Provider
	// CreateCheckout a hosztolt fizetési oldal címét adja vissza.
	CreateCheckout(ctx context.Context, req CheckoutRequest) (string, error)
}
