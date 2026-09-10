// Package httpapi a JSON HTTP felület: útvonalak, handlerek, middleware-ek.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/lipcsei/snitt-cloud/services/account/internal/auth"
	"github.com/lipcsei/snitt-cloud/services/account/internal/billing"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

// Store az adatelérési függőség interfészként, hogy a handlerek élő
// adatbázis nélkül is tesztelhetők legyenek.
type Store interface {
	Ping(ctx context.Context) error
	UpsertProfile(ctx context.Context, subject, email, username, name string) (store.Profile, error)
	UpdateProfile(ctx context.Context, subject string, displayName, locale *string) (store.Profile, error)
	ActiveEntitlements(ctx context.Context, subject string) ([]store.Entitlement, error)

	// Az admin felület által használt műveletek.
	GetProfile(ctx context.Context, subject string) (store.Profile, error)
	ProfilesBySubjects(ctx context.Context, subjects []string) (map[string]store.Profile, error)
	AllEntitlements(ctx context.Context, subject string) ([]store.Entitlement, error)

	ListSubscriptions(ctx context.Context, f store.SubscriptionFilter) ([]store.Subscription, int, error)
	LiveSubscriptionsBySubjects(ctx context.Context, subjects []string) (map[string]store.Subscription, error)
	GetSubscription(ctx context.Context, id string) (store.Subscription, error)
	GrantSubscription(ctx context.Context, in store.NewSubscription, sync store.EntitlementSync) (store.Subscription, error)
	ChangeSubscriptionPlan(ctx context.Context, id string, ch store.PlanChange, sync store.EntitlementSync) (store.Subscription, error)
	CancelSubscription(ctx context.Context, id string, atPeriodEnd bool, sync store.EntitlementSync) (store.Subscription, error)
	ReactivateSubscription(ctx context.Context, id string, periodStart, periodEnd *time.Time, sync store.EntitlementSync) (store.Subscription, error)

	ListInvoices(ctx context.Context, f store.InvoiceFilter) ([]store.Invoice, store.InvoiceTotals, error)
	GetInvoice(ctx context.Context, id string) (store.Invoice, error)
	CreateInvoice(ctx context.Context, in store.NewInvoice) (store.Invoice, error)
	MarkInvoicePaid(ctx context.Context, id string, paidAt time.Time, externalID string) (store.Invoice, error)
	VoidInvoice(ctx context.Context, id, note string) (store.Invoice, error)

	GrantEntitlement(ctx context.Context, subject, featureKey string, expiresAt *time.Time) (store.Entitlement, error)
	RevokeEntitlement(ctx context.Context, subject, featureKey string) (store.Entitlement, error)

	Overview(ctx context.Context) (store.Overview, error)

	RecordAudit(ctx context.Context, in store.NewAuditEntry) (store.AuditEntry, error)
	ListAudit(ctx context.Context, f store.AuditFilter) ([]store.AuditEntry, int, error)
}

// API a HTTP réteg összefogója.
type API struct {
	store     Store
	verifier  auth.TokenVerifier
	directory UserDirectory
	billing   billing.Provider
	origins   []string
	adminRole string
	currency  string
	log       *slog.Logger
}

// Options az API függőségei. Struct, mert a paraméterlista már túl hosszú
// lenne, és a hívási helyen így látszik, melyik érték mi.
type Options struct {
	Store           Store
	Verifier        auth.TokenVerifier
	Directory       UserDirectory
	Billing         billing.Provider
	CORSOrigins     []string
	AdminRole       string
	DefaultCurrency string
	Log             *slog.Logger
}

// New összeállítja az API-t.
func New(o Options) *API {
	if o.Billing == nil {
		o.Billing = billing.Manual{}
	}
	if o.AdminRole == "" {
		o.AdminRole = "admin"
	}
	if o.DefaultCurrency == "" {
		o.DefaultCurrency = "HUF"
	}
	return &API{
		store:     o.Store,
		verifier:  o.Verifier,
		directory: o.Directory,
		billing:   o.Billing,
		origins:   o.CORSOrigins,
		adminRole: o.AdminRole,
		currency:  o.DefaultCurrency,
		log:       o.Log,
	}
}

// Handler visszaadja a teljes, middleware-ekkel becsomagolt routert.
func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", a.handleHealthz)

	protected := auth.Middleware(a.verifier)
	mux.Handle("GET /api/v1/me", protected(http.HandlerFunc(a.handleGetMe)))
	mux.Handle("PATCH /api/v1/me", protected(http.HandlerFunc(a.handlePatchMe)))
	mux.Handle("GET /api/v1/entitlements", protected(http.HandlerFunc(a.handleEntitlements)))

	a.adminRoutes(mux)

	return a.cors(mux)
}

func (a *API) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if err := a.store.Ping(r.Context()); err != nil {
		a.log.WarnContext(r.Context(), "healthz: adatbázis nem elérhető", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "degraded", "database": "down"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "database": "up"})
}

func (a *API) handleGetMe(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "hiányzó identitás")
		return
	}
	// Első bejelentkezéskor itt jön létre a profil sor; külön regisztrációs
	// végpontra nincs szükség, a Keycloak már elvégezte a regisztrációt.
	p, err := a.store.UpsertProfile(r.Context(), id.Subject, id.Email, id.PreferredUsername, id.Name)
	if err != nil {
		a.log.ErrorContext(r.Context(), "profil upsert sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "profil nem elérhető")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// patchMeRequest a PATCH /api/v1/me törzse. A pointer mezők különböztetik meg
// a "nem küldött" és az "üresre állított" esetet.
type patchMeRequest struct {
	DisplayName *string `json:"display_name"`
	Locale      *string `json:"locale"`
}

func (a *API) handlePatchMe(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "hiányzó identitás")
		return
	}

	var req patchMeRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10))
	// A Keycloak birtokolja az email/sub mezőket: ha a kliens megpróbálja
	// küldeni őket, inkább 400-zal elutasítjuk, mint hogy csendben eldobjuk.
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "érvénytelen kérés törzs: "+err.Error())
		return
	}

	if req.DisplayName != nil {
		v := strings.TrimSpace(*req.DisplayName)
		if v == "" {
			writeError(w, http.StatusBadRequest, "a display_name nem lehet üres")
			return
		}
		if len([]rune(v)) > 100 {
			writeError(w, http.StatusBadRequest, "a display_name legfeljebb 100 karakter lehet")
			return
		}
		req.DisplayName = &v
	}
	if req.Locale != nil {
		v := strings.TrimSpace(*req.Locale)
		if v != "" && !validLocale(v) {
			writeError(w, http.StatusBadRequest, "érvénytelen locale (várt formátum: hu vagy hu-HU)")
			return
		}
		req.Locale = &v
	}
	if req.DisplayName == nil && req.Locale == nil {
		writeError(w, http.StatusBadRequest, "nincs módosítható mező a kérésben")
		return
	}

	// A profil sor lehet, hogy még nem létezik (a kliens a GET /me előtt
	// hívott PATCH-et), ezért előbb létrehozzuk a token adataiból.
	if _, err := a.store.UpsertProfile(r.Context(), id.Subject, id.Email, id.PreferredUsername, id.Name); err != nil {
		a.log.ErrorContext(r.Context(), "profil upsert sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "profil nem elérhető")
		return
	}

	p, err := a.store.UpdateProfile(r.Context(), id.Subject, req.DisplayName, req.Locale)
	if err != nil {
		a.log.ErrorContext(r.Context(), "profil frissítés sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "profil nem frissíthető")
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// entitlementsResponse mindig objektum, nem csupasz tömb: így később
// bővíthető (pl. lekérdezés ideje, plan neve) a kliensek törése nélkül.
type entitlementsResponse struct {
	Entitlements []store.Entitlement `json:"entitlements"`
}

func (a *API) handleEntitlements(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "hiányzó identitás")
		return
	}
	items, err := a.store.ActiveEntitlements(r.Context(), id.Subject)
	if err != nil {
		a.log.ErrorContext(r.Context(), "jogosultságok lekérdezése sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "jogosultságok nem elérhetők")
		return
	}
	if items == nil {
		items = []store.Entitlement{}
	}
	writeJSON(w, http.StatusOK, entitlementsResponse{Entitlements: items})
}

// cors a CORS_ORIGINS-ban engedélyezett originokat szolgálja ki és kezeli
// a preflight kéréseket.
func (a *API) cors(next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(a.origins))
	wildcard := false
	for _, o := range a.origins {
		if o == "*" {
			wildcard = true
		}
		allowed[o] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			_, ok := allowed[origin]
			if ok || wildcard {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
				w.Header().Set("Access-Control-Max-Age", "600")
			}
			// Ugyanaz az URL más választ ad originonként, ezért a cache-nek
			// az Origin fejlécet is figyelembe kell vennie.
			w.Header().Add("Vary", "Origin")
		}

		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// validLocale a BCP47 részhalmazát fogadja el: "hu" vagy "hu-HU".
func validLocale(v string) bool {
	if len(v) > 12 {
		return false
	}
	parts := strings.Split(v, "-")
	if len(parts) > 2 {
		return false
	}
	for _, p := range parts {
		if len(p) < 2 || len(p) > 4 {
			return false
		}
		for _, r := range p {
			if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
				return false
			}
		}
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil && !errors.Is(err, http.ErrHandlerTimeout) {
		return
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
