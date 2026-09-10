package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lipcsei/snitt-cloud/services/account/internal/auth"
	"github.com/lipcsei/snitt-cloud/services/account/internal/billing"
	"github.com/lipcsei/snitt-cloud/services/account/internal/keycloak"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

// UserDirectory a Keycloak Admin API interfészként, hogy az admin handlerek
// élő Keycloak nélkül is tesztelhetők legyenek.
type UserDirectory interface {
	ListUsers(ctx context.Context, search string, first, max int) ([]keycloak.User, error)
	CountUsers(ctx context.Context, search string) (int, error)
	GetUser(ctx context.Context, id string) (keycloak.User, error)
}

const (
	defaultPageSize = 25
	maxPageSize     = 100
	defaultPeriod   = 30 * 24 * time.Hour
)

// userRef a listákban megjelenő, minimális felhasználó-hivatkozás. A helyi
// profilból jön, hogy a táblák kirajzolása ne igényeljen soronkénti
// Keycloak-hívást.
type userRef struct {
	Subject     string `json:"subject"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

// adminUser a Keycloak identitás és a helyi profil összefésült nézete.
type adminUser struct {
	Subject       string     `json:"subject"`
	Email         string     `json:"email"`
	Username      string     `json:"username"`
	FullName      string     `json:"full_name"`
	Enabled       bool       `json:"enabled"`
	EmailVerified bool       `json:"email_verified"`
	RegisteredAt  *time.Time `json:"registered_at"`

	HasProfile   bool       `json:"has_profile"`
	DisplayName  string     `json:"display_name"`
	Locale       string     `json:"locale"`
	ProfileSince *time.Time `json:"profile_since"`

	Subscription *store.Subscription `json:"subscription"`
}

func mergeUser(ku keycloak.User, p *store.Profile, sub *store.Subscription) adminUser {
	u := adminUser{
		Subject:       ku.ID,
		Email:         ku.Email,
		Username:      ku.Username,
		FullName:      ku.FullName(),
		Enabled:       ku.Enabled,
		EmailVerified: ku.EmailVerified,
		Subscription:  sub,
	}
	if t := ku.CreatedAt(); !t.IsZero() {
		u.RegisteredAt = &t
	}
	if p != nil {
		u.HasProfile = true
		u.DisplayName = p.DisplayName
		u.Locale = p.Locale
		created := p.CreatedAt
		u.ProfileSince = &created
		if u.Email == "" {
			u.Email = p.Email
		}
	}
	if u.FullName == "" {
		u.FullName = u.DisplayName
	}
	return u
}

func (a *API) adminRoutes(mux *http.ServeMux) {
	// Két réteg: előbb a meglévő token-ellenőrzés, fölötte a realm szerep.
	// Külön admin hitelesítés nincs, ugyanaz a Keycloak token dönt.
	guard := func(h http.HandlerFunc) http.Handler {
		return auth.Middleware(a.verifier)(auth.RequireRealmRole(a.adminRole)(h))
	}

	mux.Handle("GET /api/v1/admin/overview", guard(a.handleOverview))
	mux.Handle("GET /api/v1/admin/plans", guard(a.handlePlans))

	mux.Handle("GET /api/v1/admin/users", guard(a.handleListUsers))
	mux.Handle("GET /api/v1/admin/users/{subject}", guard(a.handleGetUser))

	mux.Handle("GET /api/v1/admin/subscriptions", guard(a.handleListSubscriptions))
	mux.Handle("POST /api/v1/admin/subscriptions", guard(a.handleGrantSubscription))
	mux.Handle("GET /api/v1/admin/subscriptions/{id}", guard(a.handleGetSubscription))
	mux.Handle("PATCH /api/v1/admin/subscriptions/{id}", guard(a.handleChangeSubscription))
	mux.Handle("POST /api/v1/admin/subscriptions/{id}/cancel", guard(a.handleCancelSubscription))
	mux.Handle("POST /api/v1/admin/subscriptions/{id}/reactivate", guard(a.handleReactivateSubscription))

	mux.Handle("GET /api/v1/admin/invoices", guard(a.handleListInvoices))
	mux.Handle("POST /api/v1/admin/invoices", guard(a.handleCreateInvoice))
	mux.Handle("GET /api/v1/admin/invoices/{id}", guard(a.handleGetInvoice))
	mux.Handle("POST /api/v1/admin/invoices/{id}/pay", guard(a.handlePayInvoice))
	mux.Handle("POST /api/v1/admin/invoices/{id}/void", guard(a.handleVoidInvoice))

	mux.Handle("GET /api/v1/admin/audit", guard(a.handleListAudit))
}

// ---------- áttekintés ----------

type overviewResponse struct {
	store.Overview
	// A felhasználószám a Keycloakból jön; ha az nem elérhető, null marad,
	// és a warnings megmondja, miért – így az admin nem néz hibás nullát.
	UserCount *int           `json:"user_count"`
	Currency  string         `json:"currency"`
	Plans     []billing.Plan `json:"plans"`
	Warnings  []string       `json:"warnings"`
}

func (a *API) handleOverview(w http.ResponseWriter, r *http.Request) {
	ov, err := a.store.Overview(r.Context())
	if err != nil {
		a.log.ErrorContext(r.Context(), "áttekintés lekérdezése sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "az összesítők nem elérhetők")
		return
	}

	resp := overviewResponse{
		Overview: ov,
		Currency: a.currency,
		Plans:    billing.Plans(),
		Warnings: []string{},
	}
	if n, err := a.directory.CountUsers(r.Context(), ""); err != nil {
		a.log.WarnContext(r.Context(), "keycloak felhasználószám nem elérhető", "error", err)
		resp.Warnings = append(resp.Warnings, "A Keycloak nem elérhető, a regisztrált felhasználók száma hiányzik.")
	} else {
		resp.UserCount = &n
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *API) handlePlans(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"plans": billing.Plans(), "currency": a.currency})
}

// ---------- felhasználók ----------

type usersResponse struct {
	Users    []adminUser `json:"users"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int         `json:"total"`
}

func (a *API) handleListUsers(w http.ResponseWriter, r *http.Request) {
	page, size := pagination(r)
	search := strings.TrimSpace(r.URL.Query().Get("q"))

	kcUsers, err := a.directory.ListUsers(r.Context(), search, (page-1)*size, size)
	if err != nil {
		a.log.ErrorContext(r.Context(), "keycloak felhasználólista sikertelen", "error", err)
		writeError(w, http.StatusBadGateway, "a Keycloak felhasználólistája nem elérhető: "+err.Error())
		return
	}
	total, err := a.directory.CountUsers(r.Context(), search)
	if err != nil {
		a.log.ErrorContext(r.Context(), "keycloak felhasználószám sikertelen", "error", err)
		writeError(w, http.StatusBadGateway, "a Keycloak nem elérhető: "+err.Error())
		return
	}

	subjects := make([]string, 0, len(kcUsers))
	for _, u := range kcUsers {
		subjects = append(subjects, u.ID)
	}
	profiles, err := a.store.ProfilesBySubjects(r.Context(), subjects)
	if err != nil {
		a.log.ErrorContext(r.Context(), "profilok lekérdezése sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "a profilok nem elérhetők")
		return
	}
	subs, err := a.store.LiveSubscriptionsBySubjects(r.Context(), subjects)
	if err != nil {
		a.log.ErrorContext(r.Context(), "előfizetések lekérdezése sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "az előfizetések nem elérhetők")
		return
	}

	out := make([]adminUser, 0, len(kcUsers))
	for _, ku := range kcUsers {
		var p *store.Profile
		if v, ok := profiles[ku.ID]; ok {
			p = &v
		}
		var s *store.Subscription
		if v, ok := subs[ku.ID]; ok {
			s = &v
		}
		out = append(out, mergeUser(ku, p, s))
	}
	writeJSON(w, http.StatusOK, usersResponse{Users: out, Page: page, PageSize: size, Total: total})
}

type userDetailResponse struct {
	User          adminUser            `json:"user"`
	Entitlements  []store.Entitlement  `json:"entitlements"`
	Subscriptions []store.Subscription `json:"subscriptions"`
	Invoices      []store.Invoice      `json:"invoices"`
	InvoiceTotals store.InvoiceTotals  `json:"invoice_totals"`
}

func (a *API) handleGetUser(w http.ResponseWriter, r *http.Request) {
	subject := r.PathValue("subject")

	ku, err := a.directory.GetUser(r.Context(), subject)
	if errors.Is(err, keycloak.ErrNotFound) {
		writeError(w, http.StatusNotFound, "nincs ilyen felhasználó")
		return
	}
	if err != nil {
		a.log.ErrorContext(r.Context(), "keycloak felhasználó lekérdezés sikertelen", "error", err)
		writeError(w, http.StatusBadGateway, "a Keycloak nem elérhető: "+err.Error())
		return
	}

	var profile *store.Profile
	p, err := a.store.GetProfile(r.Context(), subject)
	switch {
	case err == nil:
		profile = &p
	case errors.Is(err, store.ErrNotFound):
		// Létező Keycloak-felhasználó, aki még sosem lépett be a felhőbe:
		// ez normál állapot, nem hiba.
	default:
		a.log.ErrorContext(r.Context(), "profil lekérdezés sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "a profil nem elérhető")
		return
	}

	ents, err := a.store.AllEntitlements(r.Context(), subject)
	if err != nil {
		a.log.ErrorContext(r.Context(), "jogosultságok lekérdezése sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "a jogosultságok nem elérhetők")
		return
	}
	subs, _, err := a.store.ListSubscriptions(r.Context(), store.SubscriptionFilter{Subject: subject, Limit: maxPageSize})
	if err != nil {
		a.log.ErrorContext(r.Context(), "előfizetések lekérdezése sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "az előfizetések nem elérhetők")
		return
	}
	invoices, totals, err := a.store.ListInvoices(r.Context(), store.InvoiceFilter{Subject: subject, Limit: maxPageSize})
	if err != nil {
		a.log.ErrorContext(r.Context(), "számlák lekérdezése sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "a számlák nem elérhetők")
		return
	}

	var live *store.Subscription
	for i := range subs {
		if subs[i].Status == store.SubActive || subs[i].Status == store.SubPastDue {
			live = &subs[i]
			break
		}
	}

	writeJSON(w, http.StatusOK, userDetailResponse{
		User:          mergeUser(ku, profile, live),
		Entitlements:  ents,
		Subscriptions: subs,
		Invoices:      invoices,
		InvoiceTotals: totals,
	})
}

// ---------- előfizetések ----------

// subscriptionView a beágyazott Subscription mezőit lapítva adja vissza,
// kiegészítve a felhasználó megjelenítéséhez szükséges minimummal.
type subscriptionView struct {
	store.Subscription
	User userRef `json:"user"`
}

type subscriptionsResponse struct {
	Subscriptions []subscriptionView `json:"subscriptions"`
	Page          int                `json:"page"`
	PageSize      int                `json:"page_size"`
	Total         int                `json:"total"`
}

func (a *API) handleListSubscriptions(w http.ResponseWriter, r *http.Request) {
	page, size := pagination(r)
	q := r.URL.Query()

	plan := strings.TrimSpace(q.Get("plan"))
	if plan != "" {
		if _, ok := billing.LookupPlan(plan); !ok {
			writeError(w, http.StatusBadRequest, "ismeretlen csomag: "+plan)
			return
		}
	}
	status := strings.TrimSpace(q.Get("status"))
	if status != "" && !validSubStatus(status) {
		writeError(w, http.StatusBadRequest, "ismeretlen státusz: "+status)
		return
	}

	subs, total, err := a.store.ListSubscriptions(r.Context(), store.SubscriptionFilter{
		Subject: strings.TrimSpace(q.Get("subject")),
		Plan:    plan,
		Status:  status,
		Limit:   size,
		Offset:  (page - 1) * size,
	})
	if err != nil {
		a.log.ErrorContext(r.Context(), "előfizetések lekérdezése sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "az előfizetések nem elérhetők")
		return
	}

	views, err := a.decorateSubscriptions(r.Context(), subs)
	if err != nil {
		a.log.ErrorContext(r.Context(), "profilok lekérdezése sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "a profilok nem elérhetők")
		return
	}
	writeJSON(w, http.StatusOK, subscriptionsResponse{
		Subscriptions: views, Page: page, PageSize: size, Total: total,
	})
}

func (a *API) decorateSubscriptions(ctx context.Context, subs []store.Subscription) ([]subscriptionView, error) {
	subjects := make([]string, 0, len(subs))
	for _, s := range subs {
		subjects = append(subjects, s.Subject)
	}
	profiles, err := a.store.ProfilesBySubjects(ctx, subjects)
	if err != nil {
		return nil, err
	}
	out := make([]subscriptionView, 0, len(subs))
	for _, s := range subs {
		out = append(out, subscriptionView{Subscription: s, User: refFor(s.Subject, profiles)})
	}
	return out, nil
}

func refFor(subject string, profiles map[string]store.Profile) userRef {
	ref := userRef{Subject: subject}
	if p, ok := profiles[subject]; ok {
		ref.Email, ref.DisplayName = p.Email, p.DisplayName
	}
	return ref
}

func (a *API) handleGetSubscription(w http.ResponseWriter, r *http.Request) {
	sub, err := a.store.GetSubscription(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeStoreError(w, r, err, "az előfizetés nem elérhető")
		return
	}
	views, err := a.decorateSubscriptions(r.Context(), []store.Subscription{sub})
	if err != nil {
		a.log.ErrorContext(r.Context(), "profil lekérdezés sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "a profil nem elérhető")
		return
	}
	writeJSON(w, http.StatusOK, views[0])
}

type grantSubscriptionRequest struct {
	Subject       string  `json:"subject"`
	Plan          string  `json:"plan"`
	PeriodDays    *int    `json:"period_days"`
	PriceMinor    *int64  `json:"price_minor"`
	Currency      *string `json:"currency"`
	CreateInvoice bool    `json:"create_invoice"`
}

func (a *API) handleGrantSubscription(w http.ResponseWriter, r *http.Request) {
	var req grantSubscriptionRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Subject) == "" {
		writeError(w, http.StatusBadRequest, "a subject kötelező")
		return
	}
	plan, ok := billing.LookupPlan(req.Plan)
	if !ok {
		writeError(w, http.StatusBadRequest, "ismeretlen csomag: "+req.Plan)
		return
	}
	days, err := periodDays(req.PeriodDays)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	price, err := priceOrDefault(req.PriceMinor, plan.PriceMinor)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	currency, err := a.currencyOrDefault(req.Currency)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now().UTC()
	end := now.Add(time.Duration(days) * 24 * time.Hour)

	// A varrat: valódi szolgáltatónál itt születne a külső előfizetés. A
	// Manual implementáció üres azonosítót ad, terhelés nem történik.
	if _, err := a.billing.EnsureCustomer(r.Context(), req.Subject, ""); err != nil {
		a.log.ErrorContext(r.Context(), "fizetési szolgáltató ügyfél sikertelen", "error", err)
		writeError(w, http.StatusBadGateway, "a fizetési szolgáltató nem elérhető")
		return
	}

	sub, err := a.store.GrantSubscription(r.Context(), store.NewSubscription{
		Subject:     req.Subject,
		Plan:        plan.Key,
		PriceMinor:  price,
		Currency:    currency,
		PeriodStart: now,
		PeriodEnd:   end,
		Provider:    a.billing.Name(),
	}, store.EntitlementSync{
		Managed:   billing.ManagedFeatureKeys(),
		Grant:     plan.FeatureKeys,
		ExpiresAt: &end,
	})
	if err != nil {
		a.writeStoreError(w, r, err, "az előfizetés nem hozható létre")
		return
	}
	a.audit(r, store.NewAuditEntry{
		Action:     store.AuditSubscriptionGrant,
		TargetType: "subscription",
		TargetID:   sub.ID,
		Subject:    sub.Subject,
		Summary: fmt.Sprintf("Előfizetés kiadva: %s csomag, %s, %d napra.",
			plan.Name, moneyText(price, currency), days),
		Detail: map[string]any{
			"plan": plan.Key, "price_minor": price, "currency": currency,
			"period_days": days, "period_end": end, "create_invoice": req.CreateInvoice,
		},
	})

	if req.CreateInvoice && price > 0 {
		due := now.Add(14 * 24 * time.Hour)
		inv, err := a.store.CreateInvoice(r.Context(), store.NewInvoice{
			Subject:        sub.Subject,
			SubscriptionID: &sub.ID,
			AmountMinor:    price,
			Currency:       currency,
			DueAt:          &due,
			Provider:       a.billing.Name(),
			Note:           "Kézzel kiállított számla a(z) " + plan.Name + " csomaghoz.",
		})
		if err != nil {
			// Az előfizetés már él; a számla hiánya nem indokolja a művelet
			// visszavonását, de az adminnak látnia kell.
			a.log.ErrorContext(r.Context(), "számla kiállítása sikertelen", "error", err, "subscription_id", sub.ID)
			writeJSON(w, http.StatusCreated, map[string]any{
				"subscription": sub,
				"warning":      "Az előfizetés létrejött, de a számla kiállítása nem sikerült.",
			})
			return
		}
		a.audit(r, store.NewAuditEntry{
			Action:     store.AuditInvoiceCreate,
			TargetType: "invoice",
			TargetID:   inv.ID,
			Subject:    inv.Subject,
			Summary: fmt.Sprintf("Számla kiállítva az előfizetés kiadásakor: %s, %s.",
				inv.Number, moneyText(inv.AmountMinor, inv.Currency)),
			Detail: map[string]any{
				"number": inv.Number, "amount_minor": inv.AmountMinor, "currency": inv.Currency,
				"subscription_id": sub.ID, "due_at": due,
			},
		})
	}
	writeJSON(w, http.StatusCreated, map[string]any{"subscription": sub})
}

type changeSubscriptionRequest struct {
	Plan       string  `json:"plan"`
	PriceMinor *int64  `json:"price_minor"`
	Currency   *string `json:"currency"`
	PeriodDays *int    `json:"period_days"`
}

func (a *API) handleChangeSubscription(w http.ResponseWriter, r *http.Request) {
	var req changeSubscriptionRequest
	if !decodeBody(w, r, &req) {
		return
	}
	plan, ok := billing.LookupPlan(req.Plan)
	if !ok {
		writeError(w, http.StatusBadRequest, "ismeretlen csomag: "+req.Plan)
		return
	}
	price, err := priceOrDefault(req.PriceMinor, plan.PriceMinor)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	currency, err := a.currencyOrDefault(req.Currency)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	ch := store.PlanChange{Plan: plan.Key, PriceMinor: price, Currency: currency}
	// Csomagváltásnál az admin dönthet úgy, hogy új elszámolási időszak indul.
	if req.PeriodDays != nil {
		days, err := periodDays(req.PeriodDays)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		now := time.Now().UTC()
		end := now.Add(time.Duration(days) * 24 * time.Hour)
		ch.PeriodStart, ch.PeriodEnd = &now, &end
	}

	current, err := a.store.GetSubscription(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeStoreError(w, r, err, "az előfizetés nem elérhető")
		return
	}
	expires := current.CurrentPeriodEnd
	if ch.PeriodEnd != nil {
		expires = *ch.PeriodEnd
	}

	sub, err := a.store.ChangeSubscriptionPlan(r.Context(), current.ID, ch, store.EntitlementSync{
		Managed:   billing.ManagedFeatureKeys(),
		Grant:     plan.FeatureKeys,
		ExpiresAt: &expires,
	})
	if err != nil {
		a.writeStoreError(w, r, err, "a csomag nem váltható")
		return
	}
	a.audit(r, store.NewAuditEntry{
		Action:     store.AuditSubscriptionChangePlan,
		TargetType: "subscription",
		TargetID:   sub.ID,
		Subject:    sub.Subject,
		Summary: fmt.Sprintf("Csomagváltás: %s → %s, %s.",
			current.Plan, plan.Key, moneyText(price, currency)),
		Detail: map[string]any{
			"from_plan": current.Plan, "to_plan": plan.Key,
			"from_price_minor": current.PriceMinor, "price_minor": price, "currency": currency,
			"new_period": req.PeriodDays != nil, "period_end": expires,
		},
	})
	writeJSON(w, http.StatusOK, map[string]any{"subscription": sub})
}

type cancelSubscriptionRequest struct {
	AtPeriodEnd bool `json:"at_period_end"`
}

func (a *API) handleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	var req cancelSubscriptionRequest
	if !decodeBody(w, r, &req) {
		return
	}
	current, err := a.store.GetSubscription(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeStoreError(w, r, err, "az előfizetés nem elérhető")
		return
	}

	sync := store.EntitlementSync{Managed: billing.ManagedFeatureKeys()}
	if req.AtPeriodEnd {
		// A periódus végéig jár a hozzáférés: a jogosultságok maradnak,
		// lejáratuk a periódus vége.
		if plan, ok := billing.LookupPlan(current.Plan); ok {
			end := current.CurrentPeriodEnd
			sync.Grant, sync.ExpiresAt = plan.FeatureKeys, &end
		}
	}

	if err := a.billing.CancelSubscription(r.Context(), current.ExternalSubscriptionID, req.AtPeriodEnd); err != nil {
		a.log.ErrorContext(r.Context(), "szolgáltatói lemondás sikertelen", "error", err)
		writeError(w, http.StatusBadGateway, "a fizetési szolgáltatónál nem sikerült a lemondás: "+err.Error())
		return
	}

	sub, err := a.store.CancelSubscription(r.Context(), current.ID, req.AtPeriodEnd, sync)
	if err != nil {
		a.writeStoreError(w, r, err, "az előfizetés nem mondható le")
		return
	}
	when := "azonnali hatállyal"
	if req.AtPeriodEnd {
		when = "a periódus végén (" + sub.CurrentPeriodEnd.Format(dateLayout) + ")"
	}
	a.audit(r, store.NewAuditEntry{
		Action:     store.AuditSubscriptionCancel,
		TargetType: "subscription",
		TargetID:   sub.ID,
		Subject:    sub.Subject,
		Summary:    fmt.Sprintf("Előfizetés lemondva %s: %s csomag.", when, sub.Plan),
		Detail: map[string]any{
			"plan": sub.Plan, "at_period_end": req.AtPeriodEnd,
			"period_end": sub.CurrentPeriodEnd, "status": sub.Status,
		},
	})
	writeJSON(w, http.StatusOK, map[string]any{"subscription": sub})
}

type reactivateSubscriptionRequest struct {
	PeriodDays *int `json:"period_days"`
}

func (a *API) handleReactivateSubscription(w http.ResponseWriter, r *http.Request) {
	var req reactivateSubscriptionRequest
	if !decodeBody(w, r, &req) {
		return
	}
	current, err := a.store.GetSubscription(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeStoreError(w, r, err, "az előfizetés nem elérhető")
		return
	}
	plan, ok := billing.LookupPlan(current.Plan)
	if !ok {
		writeError(w, http.StatusConflict, "a rögzített csomag ("+current.Plan+") már nem létezik, előbb csomagot kell váltani")
		return
	}

	var periodStart, periodEnd *time.Time
	expires := current.CurrentPeriodEnd
	now := time.Now().UTC()
	// Ha a régi periódus már lejárt, visszakapcsoláskor újat kell nyitni,
	// különben azonnal lejárt jogosultságot adnánk vissza.
	if req.PeriodDays != nil || !current.CurrentPeriodEnd.After(now) {
		days, err := periodDays(req.PeriodDays)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		end := now.Add(time.Duration(days) * 24 * time.Hour)
		periodStart, periodEnd, expires = &now, &end, end
	}

	sub, err := a.store.ReactivateSubscription(r.Context(), current.ID, periodStart, periodEnd, store.EntitlementSync{
		Managed:   billing.ManagedFeatureKeys(),
		Grant:     plan.FeatureKeys,
		ExpiresAt: &expires,
	})
	if err != nil {
		a.writeStoreError(w, r, err, "az előfizetés nem kapcsolható vissza")
		return
	}
	a.audit(r, store.NewAuditEntry{
		Action:     store.AuditSubscriptionReactivate,
		TargetType: "subscription",
		TargetID:   sub.ID,
		Subject:    sub.Subject,
		Summary: fmt.Sprintf("Előfizetés visszakapcsolva: %s csomag, a periódus vége %s.",
			sub.Plan, sub.CurrentPeriodEnd.Format(dateLayout)),
		Detail: map[string]any{
			"plan": sub.Plan, "new_period": periodEnd != nil,
			"period_end": sub.CurrentPeriodEnd, "status": sub.Status,
		},
	})
	writeJSON(w, http.StatusOK, map[string]any{"subscription": sub})
}

// ---------- számlázás ----------

type invoiceView struct {
	store.Invoice
	User userRef `json:"user"`
}

type invoicesResponse struct {
	Invoices []invoiceView       `json:"invoices"`
	Totals   store.InvoiceTotals `json:"totals"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Currency string              `json:"currency"`
}

func (a *API) handleListInvoices(w http.ResponseWriter, r *http.Request) {
	page, size := pagination(r)
	q := r.URL.Query()

	status := strings.TrimSpace(q.Get("status"))
	if status != "" && !validInvoiceStatus(status) {
		writeError(w, http.StatusBadRequest, "ismeretlen státusz: "+status)
		return
	}

	invoices, totals, err := a.store.ListInvoices(r.Context(), store.InvoiceFilter{
		Subject: strings.TrimSpace(q.Get("subject")),
		Status:  status,
		Limit:   size,
		Offset:  (page - 1) * size,
	})
	if err != nil {
		a.log.ErrorContext(r.Context(), "számlák lekérdezése sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "a számlák nem elérhetők")
		return
	}

	subjects := make([]string, 0, len(invoices))
	for _, inv := range invoices {
		subjects = append(subjects, inv.Subject)
	}
	profiles, err := a.store.ProfilesBySubjects(r.Context(), subjects)
	if err != nil {
		a.log.ErrorContext(r.Context(), "profilok lekérdezése sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "a profilok nem elérhetők")
		return
	}

	views := make([]invoiceView, 0, len(invoices))
	for _, inv := range invoices {
		views = append(views, invoiceView{Invoice: inv, User: refFor(inv.Subject, profiles)})
	}
	writeJSON(w, http.StatusOK, invoicesResponse{
		Invoices: views, Totals: totals, Page: page, PageSize: size, Currency: a.currency,
	})
}

func (a *API) handleGetInvoice(w http.ResponseWriter, r *http.Request) {
	inv, err := a.store.GetInvoice(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeStoreError(w, r, err, "a számla nem elérhető")
		return
	}
	profiles, err := a.store.ProfilesBySubjects(r.Context(), []string{inv.Subject})
	if err != nil {
		a.log.ErrorContext(r.Context(), "profil lekérdezés sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, "a profil nem elérhető")
		return
	}
	writeJSON(w, http.StatusOK, invoiceView{Invoice: inv, User: refFor(inv.Subject, profiles)})
}

type createInvoiceRequest struct {
	Subject        string  `json:"subject"`
	SubscriptionID *string `json:"subscription_id"`
	AmountMinor    int64   `json:"amount_minor"`
	Currency       *string `json:"currency"`
	DueDays        *int    `json:"due_days"`
	Note           string  `json:"note"`
}

func (a *API) handleCreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req createInvoiceRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Subject) == "" {
		writeError(w, http.StatusBadRequest, "a subject kötelező")
		return
	}
	if req.AmountMinor <= 0 {
		writeError(w, http.StatusBadRequest, "az összegnek pozitívnak kell lennie")
		return
	}
	currency, err := a.currencyOrDefault(req.Currency)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var due *time.Time
	days := 14
	if req.DueDays != nil {
		days = *req.DueDays
	}
	if days < 0 || days > 365 {
		writeError(w, http.StatusBadRequest, "a fizetési határidő 0 és 365 nap között lehet")
		return
	}
	d := time.Now().UTC().Add(time.Duration(days) * 24 * time.Hour)
	due = &d

	inv, err := a.store.CreateInvoice(r.Context(), store.NewInvoice{
		Subject:        req.Subject,
		SubscriptionID: req.SubscriptionID,
		AmountMinor:    req.AmountMinor,
		Currency:       currency,
		DueAt:          due,
		Provider:       a.billing.Name(),
		Note:           strings.TrimSpace(req.Note),
	})
	if err != nil {
		a.writeStoreError(w, r, err, "a számla nem állítható ki")
		return
	}
	a.audit(r, store.NewAuditEntry{
		Action:     store.AuditInvoiceCreate,
		TargetType: "invoice",
		TargetID:   inv.ID,
		Subject:    inv.Subject,
		Summary: fmt.Sprintf("Számla kiállítva: %s, %s.",
			inv.Number, moneyText(inv.AmountMinor, inv.Currency)),
		Detail: map[string]any{
			"number": inv.Number, "amount_minor": inv.AmountMinor, "currency": inv.Currency,
			"subscription_id": req.SubscriptionID, "due_at": due, "note": inv.Note,
		},
	})
	writeJSON(w, http.StatusCreated, map[string]any{"invoice": inv})
}

type payInvoiceRequest struct {
	PaidAt            *time.Time `json:"paid_at"`
	ExternalInvoiceID string     `json:"external_invoice_id"`
}

func (a *API) handlePayInvoice(w http.ResponseWriter, r *http.Request) {
	var req payInvoiceRequest
	if !decodeBody(w, r, &req) {
		return
	}
	paidAt := time.Now().UTC()
	if req.PaidAt != nil {
		paidAt = req.PaidAt.UTC()
		if paidAt.After(time.Now().Add(time.Hour)) {
			writeError(w, http.StatusBadRequest, "a fizetés dátuma nem lehet a jövőben")
			return
		}
	}

	// Itt nincs terhelés: a pénz máshol érkezett be, ez csak a könyvelés
	// vezetése. Valódi szolgáltatónál ezt a lépést a webhook váltaná ki.
	inv, err := a.store.MarkInvoicePaid(r.Context(), r.PathValue("id"), paidAt, strings.TrimSpace(req.ExternalInvoiceID))
	if err != nil {
		a.writeStoreError(w, r, err, "a számla nem jelölhető kifizetettnek")
		return
	}
	a.audit(r, store.NewAuditEntry{
		Action:     store.AuditInvoicePay,
		TargetType: "invoice",
		TargetID:   inv.ID,
		Subject:    inv.Subject,
		Summary: fmt.Sprintf("Számla kifizetettnek jelölve: %s, %s.",
			inv.Number, moneyText(inv.AmountMinor, inv.Currency)),
		Detail: map[string]any{
			"number": inv.Number, "amount_minor": inv.AmountMinor, "currency": inv.Currency,
			"paid_at": paidAt, "external_invoice_id": inv.ExternalInvoiceID,
		},
	})
	writeJSON(w, http.StatusOK, map[string]any{"invoice": inv})
}

type voidInvoiceRequest struct {
	Note string `json:"note"`
}

func (a *API) handleVoidInvoice(w http.ResponseWriter, r *http.Request) {
	var req voidInvoiceRequest
	if !decodeBody(w, r, &req) {
		return
	}
	inv, err := a.store.VoidInvoice(r.Context(), r.PathValue("id"), strings.TrimSpace(req.Note))
	if err != nil {
		a.writeStoreError(w, r, err, "a számla nem sztornózható")
		return
	}
	a.audit(r, store.NewAuditEntry{
		Action:     store.AuditInvoiceVoid,
		TargetType: "invoice",
		TargetID:   inv.ID,
		Subject:    inv.Subject,
		Summary: fmt.Sprintf("Számla sztornózva: %s, %s. Indok: %s",
			inv.Number, moneyText(inv.AmountMinor, inv.Currency), strings.TrimSpace(req.Note)),
		Detail: map[string]any{
			"number": inv.Number, "amount_minor": inv.AmountMinor,
			"currency": inv.Currency, "note": strings.TrimSpace(req.Note),
		},
	})
	writeJSON(w, http.StatusOK, map[string]any{"invoice": inv})
}

// ---------- segédek ----------

func (a *API) writeStoreError(w http.ResponseWriter, r *http.Request, err error, fallback string) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "a keresett rekord nem található")
	case errors.Is(err, store.ErrConflict):
		// Az ütközés oka az adminnak érdemi információ (pl. "már van élő
		// előfizetése"), ezért a store üzenetét visszaadjuk.
		writeError(w, http.StatusConflict, err.Error())
	default:
		a.log.ErrorContext(r.Context(), "adatbázis művelet sikertelen", "error", err)
		writeError(w, http.StatusInternalServerError, fallback)
	}
}

// decodeBody a kérés törzsét olvassa; üres törzs üres objektumnak számít,
// mert több admin művelethez nincs kötelező paraméter.
func decodeBody(w http.ResponseWriter, r *http.Request, out any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		if errors.Is(err, io.EOF) {
			return true
		}
		writeError(w, http.StatusBadRequest, "érvénytelen kérés törzs: "+err.Error())
		return false
	}
	return true
}

func pagination(r *http.Request) (page, size int) {
	page, size = 1, defaultPageSize
	if v, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && v > 0 {
		size = min(v, maxPageSize)
	}
	return page, size
}

func periodDays(v *int) (int, error) {
	if v == nil {
		return int(defaultPeriod / (24 * time.Hour)), nil
	}
	if *v < 1 || *v > 3650 {
		return 0, fmt.Errorf("az időszak hossza 1 és 3650 nap között lehet")
	}
	return *v, nil
}

func priceOrDefault(v *int64, def int64) (int64, error) {
	if v == nil {
		return def, nil
	}
	if *v < 0 {
		return 0, fmt.Errorf("az ár nem lehet negatív")
	}
	return *v, nil
}

func (a *API) currencyOrDefault(v *string) (string, error) {
	if v == nil || strings.TrimSpace(*v) == "" {
		return a.currency, nil
	}
	c := strings.ToUpper(strings.TrimSpace(*v))
	if len(c) != 3 {
		return "", fmt.Errorf("a pénznem három betűs ISO kód kell legyen")
	}
	for _, r := range c {
		if r < 'A' || r > 'Z' {
			return "", fmt.Errorf("a pénznem három betűs ISO kód kell legyen")
		}
	}
	return c, nil
}

func validSubStatus(s string) bool {
	switch s {
	case store.SubActive, store.SubPastDue, store.SubCanceled, store.SubExpired:
		return true
	}
	return false
}

func validInvoiceStatus(s string) bool {
	switch s {
	case store.InvoiceOpen, store.InvoicePaid, store.InvoiceVoid:
		return true
	}
	return false
}
