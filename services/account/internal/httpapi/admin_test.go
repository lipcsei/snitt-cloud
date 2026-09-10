package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/lipcsei/snitt-cloud/services/account/internal/auth"
	"github.com/lipcsei/snitt-cloud/services/account/internal/keycloak"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

// adminIdentity a realm "admin" szereppel rendelkező hívó.
var adminIdentity = auth.Identity{
	Subject:           "admin-1",
	Email:             "admin@snitt.local",
	PreferredUsername: "admin",
	Name:              "Admin Elek",
	RealmAccess:       auth.RealmAccess{Roles: []string{"offline_access", "admin"}},
}

func newTestAPIWith(st Store, dir UserDirectory) http.Handler {
	v := fakeVerifier{valid: map[string]auth.Identity{
		"good-token":  testIdentity, // bejelentkezett, de nem admin
		"admin-token": adminIdentity,
	}}
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))
	return New(Options{
		Store:           st,
		Verifier:        v,
		Directory:       dir,
		CORSOrigins:     []string{"http://localhost:5174", "http://localhost:5175"},
		AdminRole:       "admin",
		DefaultCurrency: "HUF",
		Log:             log,
	}).Handler()
}

// ---------- kamu Keycloak címtár ----------

type fakeDirectory struct {
	users      []keycloak.User
	listErr    error
	countErr   error
	getErr     error
	lastSearch string
	lastFirst  int
	lastMax    int
}

func (d *fakeDirectory) ListUsers(_ context.Context, search string, first, max int) ([]keycloak.User, error) {
	d.lastSearch, d.lastFirst, d.lastMax = search, first, max
	if d.listErr != nil {
		return nil, d.listErr
	}
	return d.users, nil
}

func (d *fakeDirectory) CountUsers(_ context.Context, _ string) (int, error) {
	if d.countErr != nil {
		return 0, d.countErr
	}
	return len(d.users), nil
}

func (d *fakeDirectory) GetUser(_ context.Context, id string) (keycloak.User, error) {
	if d.getErr != nil {
		return keycloak.User{}, d.getErr
	}
	for _, u := range d.users {
		if u.ID == id {
			return u, nil
		}
	}
	return keycloak.User{}, keycloak.ErrNotFound
}

// ---------- kamu admin store ----------

// adminFake az admin végpontok mögötti, memóriában tartott állapot. Nem
// modellezi az adatbázist, csak annyit, amennyit a handlerek látnak belőle.
type adminFake struct {
	profiles      map[string]store.Profile
	entitlements  map[string][]store.Entitlement
	subscriptions map[string]store.Subscription
	invoices      map[string]store.Invoice
	overview      store.Overview

	err error // ha nem nil, minden admin művelet ezzel tér vissza

	lastSync        store.EntitlementSync
	lastNewSub      store.NewSubscription
	lastPlanChange  store.PlanChange
	lastCancelAtEnd bool
	lastNewInvoice  store.NewInvoice
	createdInvoices int

	// A napló külön hibát kap: a naplóírás bukása szándékosan nem bukatja
	// el az admin műveletet, ezt csak külön kapcsolóval lehet előidézni.
	audit    []store.AuditEntry
	auditErr error
}

func (f *fakeStore) GetProfile(_ context.Context, subject string) (store.Profile, error) {
	if f.admin.err != nil {
		return store.Profile{}, f.admin.err
	}
	if p, ok := f.admin.profiles[subject]; ok {
		return p, nil
	}
	return store.Profile{}, store.ErrNotFound
}

func (f *fakeStore) ProfilesBySubjects(_ context.Context, subjects []string) (map[string]store.Profile, error) {
	if f.admin.err != nil {
		return nil, f.admin.err
	}
	out := map[string]store.Profile{}
	for _, s := range subjects {
		if p, ok := f.admin.profiles[s]; ok {
			out[s] = p
		}
	}
	return out, nil
}

func (f *fakeStore) AllEntitlements(_ context.Context, subject string) ([]store.Entitlement, error) {
	if f.admin.err != nil {
		return nil, f.admin.err
	}
	out := f.admin.entitlements[subject]
	if out == nil {
		out = []store.Entitlement{}
	}
	return out, nil
}

func (f *fakeStore) ListSubscriptions(_ context.Context, filter store.SubscriptionFilter) ([]store.Subscription, int, error) {
	if f.admin.err != nil {
		return nil, 0, f.admin.err
	}
	out := []store.Subscription{}
	for _, s := range f.admin.subscriptions {
		if filter.Subject != "" && s.Subject != filter.Subject {
			continue
		}
		if filter.Plan != "" && s.Plan != filter.Plan {
			continue
		}
		if filter.Status != "" && s.Status != filter.Status {
			continue
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, len(out), nil
}

func (f *fakeStore) LiveSubscriptionsBySubjects(_ context.Context, subjects []string) (map[string]store.Subscription, error) {
	if f.admin.err != nil {
		return nil, f.admin.err
	}
	want := map[string]bool{}
	for _, s := range subjects {
		want[s] = true
	}
	out := map[string]store.Subscription{}
	for _, s := range f.admin.subscriptions {
		if want[s.Subject] && (s.Status == store.SubActive || s.Status == store.SubPastDue) {
			out[s.Subject] = s
		}
	}
	return out, nil
}

func (f *fakeStore) GetSubscription(_ context.Context, id string) (store.Subscription, error) {
	if f.admin.err != nil {
		return store.Subscription{}, f.admin.err
	}
	if s, ok := f.admin.subscriptions[id]; ok {
		return s, nil
	}
	return store.Subscription{}, store.ErrNotFound
}

func (f *fakeStore) GrantSubscription(_ context.Context, in store.NewSubscription, sync store.EntitlementSync) (store.Subscription, error) {
	f.admin.lastNewSub, f.admin.lastSync = in, sync
	if f.admin.err != nil {
		return store.Subscription{}, f.admin.err
	}
	for _, s := range f.admin.subscriptions {
		if s.Subject == in.Subject && (s.Status == store.SubActive || s.Status == store.SubPastDue) {
			return store.Subscription{}, fmt.Errorf("%w: a felhasználónak már van élő előfizetése", store.ErrConflict)
		}
	}
	sub := store.Subscription{
		ID: "sub-new", Subject: in.Subject, Plan: in.Plan, Status: store.SubActive,
		CurrentPeriodStart: in.PeriodStart, CurrentPeriodEnd: in.PeriodEnd,
		PriceMinor: in.PriceMinor, Currency: in.Currency, Provider: in.Provider,
	}
	f.admin.ensure()
	f.admin.subscriptions[sub.ID] = sub
	return sub, nil
}

func (f *fakeStore) ChangeSubscriptionPlan(_ context.Context, id string, ch store.PlanChange, sync store.EntitlementSync) (store.Subscription, error) {
	f.admin.lastPlanChange, f.admin.lastSync = ch, sync
	if f.admin.err != nil {
		return store.Subscription{}, f.admin.err
	}
	sub, ok := f.admin.subscriptions[id]
	if !ok {
		return store.Subscription{}, store.ErrNotFound
	}
	if sub.Status != store.SubActive && sub.Status != store.SubPastDue {
		return store.Subscription{}, store.ErrConflict
	}
	sub.Plan, sub.PriceMinor, sub.Currency = ch.Plan, ch.PriceMinor, ch.Currency
	if ch.PeriodEnd != nil {
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd = *ch.PeriodStart, *ch.PeriodEnd
	}
	f.admin.subscriptions[id] = sub
	return sub, nil
}

func (f *fakeStore) CancelSubscription(_ context.Context, id string, atPeriodEnd bool, sync store.EntitlementSync) (store.Subscription, error) {
	f.admin.lastCancelAtEnd, f.admin.lastSync = atPeriodEnd, sync
	if f.admin.err != nil {
		return store.Subscription{}, f.admin.err
	}
	sub, ok := f.admin.subscriptions[id]
	if !ok {
		return store.Subscription{}, store.ErrNotFound
	}
	if sub.Status != store.SubActive && sub.Status != store.SubPastDue {
		return store.Subscription{}, store.ErrConflict
	}
	now := time.Now().UTC()
	sub.CancelAtPeriodEnd, sub.CanceledAt = atPeriodEnd, &now
	if !atPeriodEnd {
		sub.Status = store.SubCanceled
	}
	f.admin.subscriptions[id] = sub
	return sub, nil
}

func (f *fakeStore) ReactivateSubscription(_ context.Context, id string, periodStart, periodEnd *time.Time, sync store.EntitlementSync) (store.Subscription, error) {
	f.admin.lastSync = sync
	if f.admin.err != nil {
		return store.Subscription{}, f.admin.err
	}
	sub, ok := f.admin.subscriptions[id]
	if !ok {
		return store.Subscription{}, store.ErrNotFound
	}
	sub.Status, sub.CancelAtPeriodEnd, sub.CanceledAt = store.SubActive, false, nil
	if periodStart != nil {
		sub.CurrentPeriodStart, sub.CurrentPeriodEnd = *periodStart, *periodEnd
	}
	f.admin.subscriptions[id] = sub
	return sub, nil
}

func (f *fakeStore) ListInvoices(_ context.Context, filter store.InvoiceFilter) ([]store.Invoice, store.InvoiceTotals, error) {
	if f.admin.err != nil {
		return nil, store.InvoiceTotals{}, f.admin.err
	}
	out := []store.Invoice{}
	var t store.InvoiceTotals
	for _, inv := range f.admin.invoices {
		if filter.Subject != "" && inv.Subject != filter.Subject {
			continue
		}
		if filter.Status != "" && inv.Status != filter.Status {
			continue
		}
		out = append(out, inv)
		t.Count++
		t.TotalMinor += inv.AmountMinor
		switch inv.Status {
		case store.InvoicePaid:
			t.PaidMinor += inv.AmountMinor
		case store.InvoiceOpen:
			t.OpenMinor += inv.AmountMinor
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, t, nil
}

func (f *fakeStore) GetInvoice(_ context.Context, id string) (store.Invoice, error) {
	if f.admin.err != nil {
		return store.Invoice{}, f.admin.err
	}
	if inv, ok := f.admin.invoices[id]; ok {
		return inv, nil
	}
	return store.Invoice{}, store.ErrNotFound
}

func (f *fakeStore) CreateInvoice(_ context.Context, in store.NewInvoice) (store.Invoice, error) {
	f.admin.lastNewInvoice = in
	if f.admin.err != nil {
		return store.Invoice{}, f.admin.err
	}
	f.admin.createdInvoices++
	f.admin.ensure()
	inv := store.Invoice{
		ID: "inv-new", Number: "SNITT-2026-000001", Subject: in.Subject,
		SubscriptionID: in.SubscriptionID, AmountMinor: in.AmountMinor,
		Currency: in.Currency, Status: store.InvoiceOpen, IssuedAt: time.Now().UTC(),
		DueAt: in.DueAt, Provider: in.Provider, Note: in.Note,
	}
	f.admin.invoices[inv.ID] = inv
	return inv, nil
}

func (f *fakeStore) MarkInvoicePaid(_ context.Context, id string, paidAt time.Time, externalID string) (store.Invoice, error) {
	if f.admin.err != nil {
		return store.Invoice{}, f.admin.err
	}
	inv, ok := f.admin.invoices[id]
	if !ok {
		return store.Invoice{}, store.ErrNotFound
	}
	if inv.Status != store.InvoiceOpen {
		return store.Invoice{}, store.ErrConflict
	}
	inv.Status, inv.PaidAt = store.InvoicePaid, &paidAt
	if externalID != "" {
		inv.ExternalInvoiceID = externalID
	}
	f.admin.invoices[id] = inv
	return inv, nil
}

func (f *fakeStore) VoidInvoice(_ context.Context, id, note string) (store.Invoice, error) {
	if f.admin.err != nil {
		return store.Invoice{}, f.admin.err
	}
	inv, ok := f.admin.invoices[id]
	if !ok {
		return store.Invoice{}, store.ErrNotFound
	}
	if inv.Status != store.InvoiceOpen {
		return store.Invoice{}, store.ErrConflict
	}
	inv.Status, inv.PaidAt = store.InvoiceVoid, nil
	if note != "" {
		inv.Note = note
	}
	f.admin.invoices[id] = inv
	return inv, nil
}

func (f *fakeStore) Overview(context.Context) (store.Overview, error) {
	if f.admin.err != nil {
		return store.Overview{}, f.admin.err
	}
	return f.admin.overview, nil
}

func (f *fakeStore) RecordAudit(_ context.Context, in store.NewAuditEntry) (store.AuditEntry, error) {
	if f.admin.auditErr != nil {
		return store.AuditEntry{}, f.admin.auditErr
	}
	e := store.AuditEntry{
		ID:           fmt.Sprintf("audit-%d", len(f.admin.audit)+1),
		At:           time.Now().UTC(),
		ActorSubject: in.ActorSubject,
		ActorLabel:   in.ActorLabel,
		Action:       in.Action,
		TargetType:   in.TargetType,
		TargetID:     in.TargetID,
		Subject:      in.Subject,
		Summary:      in.Summary,
		Detail:       in.Detail,
	}
	f.admin.audit = append(f.admin.audit, e)
	return e, nil
}

func (f *fakeStore) ListAudit(_ context.Context, filter store.AuditFilter) ([]store.AuditEntry, int, error) {
	if f.admin.auditErr != nil {
		return nil, 0, f.admin.auditErr
	}
	out := []store.AuditEntry{}
	for _, e := range f.admin.audit {
		switch {
		case filter.Actor != "" && e.ActorSubject != filter.Actor:
		case filter.Subject != "" && e.Subject != filter.Subject:
		case filter.Action != "" && e.Action != filter.Action:
		case filter.Since != nil && e.At.Before(*filter.Since):
		case filter.Until != nil && !e.At.Before(*filter.Until):
		default:
			out = append(out, e)
		}
	}
	return out, len(out), nil
}

func (a *adminFake) ensure() {
	if a.subscriptions == nil {
		a.subscriptions = map[string]store.Subscription{}
	}
	if a.invoices == nil {
		a.invoices = map[string]store.Invoice{}
	}
}

// ---------- tesztek ----------

func do(t *testing.T, h http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestAdminRoutesRequireAdminRole(t *testing.T) {
	routes := []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/v1/admin/overview", ""},
		{http.MethodGet, "/api/v1/admin/plans", ""},
		{http.MethodGet, "/api/v1/admin/users", ""},
		{http.MethodGet, "/api/v1/admin/users/u1", ""},
		{http.MethodGet, "/api/v1/admin/subscriptions", ""},
		{http.MethodPost, "/api/v1/admin/subscriptions", `{"subject":"u1","plan":"pro"}`},
		{http.MethodGet, "/api/v1/admin/subscriptions/s1", ""},
		{http.MethodPatch, "/api/v1/admin/subscriptions/s1", `{"plan":"pro"}`},
		{http.MethodPost, "/api/v1/admin/subscriptions/s1/cancel", `{}`},
		{http.MethodPost, "/api/v1/admin/subscriptions/s1/reactivate", `{}`},
		{http.MethodGet, "/api/v1/admin/invoices", ""},
		{http.MethodPost, "/api/v1/admin/invoices", `{"subject":"u1","amount_minor":100}`},
		{http.MethodGet, "/api/v1/admin/invoices/i1", ""},
		{http.MethodPost, "/api/v1/admin/invoices/i1/pay", `{}`},
		{http.MethodPost, "/api/v1/admin/invoices/i1/void", `{}`},
		{http.MethodGet, "/api/v1/admin/audit", ""},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			h := newTestAPIWith(&fakeStore{}, &fakeDirectory{})

			if rec := do(t, h, rt.method, rt.path, "", rt.body); rec.Code != http.StatusUnauthorized {
				t.Errorf("token nélkül = %d, várt 401", rec.Code)
			}
			// Érvényes token, de nincs rajta admin szerep: 403, nem 401.
			rec := do(t, h, rt.method, rt.path, "good-token", rt.body)
			if rec.Code != http.StatusForbidden {
				t.Errorf("nem admin token = %d, várt 403 (törzs: %s)", rec.Code, rec.Body.String())
			}
			// Az admin tokennek legalább az auth rétegen át kell jutnia.
			rec = do(t, h, rt.method, rt.path, "admin-token", rt.body)
			if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
				t.Errorf("admin token = %d, nem lett volna szabad elakadnia", rec.Code)
			}
		})
	}
}

func TestListUsersMergesKeycloakAndProfile(t *testing.T) {
	dir := &fakeDirectory{users: []keycloak.User{
		{ID: "u1", Username: "anna", Email: "anna@example.com", FirstName: "Anna", LastName: "Kis", Enabled: true, EmailVerified: true, CreatedTimestamp: 1700000000000},
		{ID: "u2", Username: "bela", Email: "bela@example.com", Enabled: false},
	}}
	st := &fakeStore{admin: adminFake{
		profiles: map[string]store.Profile{
			"u1": {Subject: "u1", Email: "anna@example.com", DisplayName: "Anna K.", Locale: "hu-HU", CreatedAt: time.Now()},
		},
		subscriptions: map[string]store.Subscription{
			"s1": {ID: "s1", Subject: "u1", Plan: "pro", Status: store.SubActive},
		},
	}}
	h := newTestAPIWith(st, dir)

	rec := do(t, h, http.MethodGet, "/api/v1/admin/users?q=an&page=2&page_size=10", "admin-token", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("státusz = %d (törzs: %s)", rec.Code, rec.Body.String())
	}
	var got usersResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("dekódolás: %v", err)
	}

	if dir.lastSearch != "an" || dir.lastFirst != 10 || dir.lastMax != 10 {
		t.Errorf("keresés/lapozás nem ment át: search=%q first=%d max=%d", dir.lastSearch, dir.lastFirst, dir.lastMax)
	}
	if len(got.Users) != 2 {
		t.Fatalf("felhasználók = %d, várt 2", len(got.Users))
	}
	if !got.Users[0].HasProfile || got.Users[0].DisplayName != "Anna K." {
		t.Errorf("u1 profilja nem fésülődött be: %+v", got.Users[0])
	}
	if got.Users[0].FullName != "Anna Kis" {
		t.Errorf("teljes név = %q, várt \"Anna Kis\"", got.Users[0].FullName)
	}
	if got.Users[0].Subscription == nil || got.Users[0].Subscription.Plan != "pro" {
		t.Errorf("u1 előfizetése hiányzik: %+v", got.Users[0].Subscription)
	}
	if got.Users[1].HasProfile {
		t.Error("u2-nek nincs profilja, mégis has_profile=true")
	}
	if got.Users[1].Subscription != nil {
		t.Error("u2-nek nincs előfizetése, mégis jött")
	}
	if got.Users[0].RegisteredAt == nil {
		t.Error("a regisztráció ideje hiányzik")
	}
}

func TestListUsersKeycloakDown(t *testing.T) {
	h := newTestAPIWith(&fakeStore{}, &fakeDirectory{listErr: errors.New("connection refused")})
	rec := do(t, h, http.MethodGet, "/api/v1/admin/users", "admin-token", "")
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("státusz = %d, várt 502 (törzs: %s)", rec.Code, rec.Body.String())
	}
}

func TestGetUserDetail(t *testing.T) {
	dir := &fakeDirectory{users: []keycloak.User{{ID: "u1", Username: "anna", Email: "anna@example.com", Enabled: true}}}

	tests := []struct {
		name       string
		subject    string
		store      *fakeStore
		wantStatus int
		wantSubs   int
	}{
		{
			name:    "profil nélküli, de létező Keycloak felhasználó",
			subject: "u1",
			store:   &fakeStore{},
			// A felhőbe még sosem lépett be: ez nem hiba.
			wantStatus: http.StatusOK,
		},
		{
			name:    "profillal és előfizetéssel",
			subject: "u1",
			store: &fakeStore{admin: adminFake{
				profiles:      map[string]store.Profile{"u1": {Subject: "u1", DisplayName: "Anna"}},
				subscriptions: map[string]store.Subscription{"s1": {ID: "s1", Subject: "u1", Plan: "pro", Status: store.SubActive}},
				invoices:      map[string]store.Invoice{"i1": {ID: "i1", Subject: "u1", AmountMinor: 299000, Status: store.InvoiceOpen}},
			}},
			wantStatus: http.StatusOK,
			wantSubs:   1,
		},
		{
			name:       "ismeretlen felhasználó",
			subject:    "nincs-ilyen",
			store:      &fakeStore{},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestAPIWith(tt.store, dir)
			rec := do(t, h, http.MethodGet, "/api/v1/admin/users/"+tt.subject, "admin-token", "")
			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			var got userDetailResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("dekódolás: %v", err)
			}
			if len(got.Subscriptions) != tt.wantSubs {
				t.Errorf("előfizetések = %d, várt %d", len(got.Subscriptions), tt.wantSubs)
			}
			if got.Entitlements == nil || got.Invoices == nil {
				t.Error("az üres listák null-ként jöttek vissza")
			}
		})
	}
}

func TestGrantSubscription(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		existing    map[string]store.Subscription
		wantStatus  int
		wantPlan    string
		wantPrice   int64
		wantGrants  []string
		wantInvoice int
	}{
		{
			name:        "pro csomag alapértelmezett árral és számlával",
			body:        `{"subject":"u1","plan":"pro","create_invoice":true}`,
			wantStatus:  http.StatusCreated,
			wantPlan:    "pro",
			wantPrice:   299000,
			wantGrants:  []string{"ai-semantic-search"},
			wantInvoice: 1,
		},
		{
			name:       "studio csomag két jogosultsággal, számla nélkül",
			body:       `{"subject":"u1","plan":"studio"}`,
			wantStatus: http.StatusCreated,
			wantPlan:   "studio",
			wantPrice:  799000,
			wantGrants: []string{"ai-semantic-search", "ai-video-understanding"},
		},
		{
			name:       "egyedi ár és időszak",
			body:       `{"subject":"u1","plan":"pro","price_minor":0,"period_days":365}`,
			wantStatus: http.StatusCreated,
			wantPlan:   "pro",
			wantPrice:  0,
			wantGrants: []string{"ai-semantic-search"},
		},
		{"ismeretlen csomag", `{"subject":"u1","plan":"arany"}`, nil, http.StatusBadRequest, "", 0, nil, 0},
		{"hiányzó subject", `{"plan":"pro"}`, nil, http.StatusBadRequest, "", 0, nil, 0},
		{"negatív ár", `{"subject":"u1","plan":"pro","price_minor":-1}`, nil, http.StatusBadRequest, "", 0, nil, 0},
		{"túl hosszú időszak", `{"subject":"u1","plan":"pro","period_days":99999}`, nil, http.StatusBadRequest, "", 0, nil, 0},
		{"ismeretlen mező", `{"subject":"u1","plan":"pro","kedvezmeny":10}`, nil, http.StatusBadRequest, "", 0, nil, 0},
		{"hibás json", `{`, nil, http.StatusBadRequest, "", 0, nil, 0},
		{
			name:       "már van élő előfizetése",
			body:       `{"subject":"u1","plan":"pro"}`,
			existing:   map[string]store.Subscription{"s1": {ID: "s1", Subject: "u1", Status: store.SubActive}},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &fakeStore{admin: adminFake{subscriptions: tt.existing}}
			h := newTestAPIWith(st, &fakeDirectory{})
			rec := do(t, h, http.MethodPost, "/api/v1/admin/subscriptions", "admin-token", tt.body)

			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusCreated {
				return
			}
			if st.admin.lastNewSub.Plan != tt.wantPlan {
				t.Errorf("csomag = %q, várt %q", st.admin.lastNewSub.Plan, tt.wantPlan)
			}
			if st.admin.lastNewSub.PriceMinor != tt.wantPrice {
				t.Errorf("ár = %d, várt %d", st.admin.lastNewSub.PriceMinor, tt.wantPrice)
			}
			if st.admin.lastNewSub.Currency != "HUF" {
				t.Errorf("pénznem = %q, várt HUF", st.admin.lastNewSub.Currency)
			}
			if st.admin.lastNewSub.Provider != "manual" {
				t.Errorf("provider = %q, várt manual", st.admin.lastNewSub.Provider)
			}
			assertGrants(t, st.admin.lastSync, tt.wantGrants, &st.admin.lastNewSub.PeriodEnd)
			if st.admin.createdInvoices != tt.wantInvoice {
				t.Errorf("kiállított számlák = %d, várt %d", st.admin.createdInvoices, tt.wantInvoice)
			}
		})
	}
}

// assertGrants ellenőrzi, hogy a jogosultság-szinkron a csomagnak megfelelő
// kulcsokat adja ki, és minden csomag-kulcsot felülír.
func assertGrants(t *testing.T, sync store.EntitlementSync, want []string, wantExpiry *time.Time) {
	t.Helper()
	got := append([]string(nil), sync.Grant...)
	sort.Strings(got)
	sort.Strings(want)
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("kiadott jogosultságok = %v, várt %v", got, want)
	}
	if len(sync.Managed) < 2 {
		t.Errorf("a vezérelt kulcsok listája túl rövid: %v", sync.Managed)
	}
	if wantExpiry != nil {
		if sync.ExpiresAt == nil {
			t.Fatal("a lejárat nil, pedig a periódus végének kellene lennie")
		}
		if !sync.ExpiresAt.Equal(*wantExpiry) {
			t.Errorf("lejárat = %v, várt %v", sync.ExpiresAt, *wantExpiry)
		}
	}
}

func TestChangeSubscriptionPlan(t *testing.T) {
	base := func() *fakeStore {
		return &fakeStore{admin: adminFake{subscriptions: map[string]store.Subscription{
			"s1": {ID: "s1", Subject: "u1", Plan: "pro", Status: store.SubActive,
				CurrentPeriodEnd: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)},
			"s2": {ID: "s2", Subject: "u2", Plan: "pro", Status: store.SubCanceled,
				CurrentPeriodEnd: time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)},
		}}}
	}

	tests := []struct {
		name       string
		id         string
		body       string
		wantStatus int
		wantGrants []string
	}{
		{"váltás studio-ra", "s1", `{"plan":"studio"}`, http.StatusOK,
			[]string{"ai-semantic-search", "ai-video-understanding"}},
		{"váltás új időszakkal", "s1", `{"plan":"pro","period_days":31}`, http.StatusOK,
			[]string{"ai-semantic-search"}},
		{"ismeretlen csomag", "s1", `{"plan":"arany"}`, http.StatusBadRequest, nil},
		{"nem létező előfizetés", "nincs", `{"plan":"pro"}`, http.StatusNotFound, nil},
		{"lemondott előfizetés", "s2", `{"plan":"studio"}`, http.StatusConflict, nil},
		{"rossz pénznem", "s1", `{"plan":"pro","currency":"forint"}`, http.StatusBadRequest, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := base()
			h := newTestAPIWith(st, &fakeDirectory{})
			rec := do(t, h, http.MethodPatch, "/api/v1/admin/subscriptions/"+tt.id, "admin-token", tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			assertGrants(t, st.admin.lastSync, tt.wantGrants, nil)
			if st.admin.lastSync.ExpiresAt == nil {
				t.Error("a lejáratnak a periódus végének kell lennie, nem nil-nek")
			}
		})
	}
}

func TestCancelSubscription(t *testing.T) {
	periodEnd := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		body       string
		wantAtEnd  bool
		wantGrants []string
		wantExpiry bool
		wantStatus int
	}{
		{
			name:       "azonnali lemondás elveszi a jogosultságokat",
			body:       `{"at_period_end":false}`,
			wantAtEnd:  false,
			wantGrants: nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "periódus végi lemondás meghagyja a jogosultságokat",
			body:       `{"at_period_end":true}`,
			wantAtEnd:  true,
			wantGrants: []string{"ai-semantic-search"},
			wantExpiry: true,
			wantStatus: http.StatusOK,
		},
		{
			// Üres törzs = azonnali lemondás, mert a bool alapértéke false.
			name:       "üres törzs",
			body:       `{}`,
			wantAtEnd:  false,
			wantGrants: nil,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &fakeStore{admin: adminFake{subscriptions: map[string]store.Subscription{
				"s1": {ID: "s1", Subject: "u1", Plan: "pro", Status: store.SubActive, CurrentPeriodEnd: periodEnd},
			}}}
			h := newTestAPIWith(st, &fakeDirectory{})
			rec := do(t, h, http.MethodPost, "/api/v1/admin/subscriptions/s1/cancel", "admin-token", tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if st.admin.lastCancelAtEnd != tt.wantAtEnd {
				t.Errorf("at_period_end = %v, várt %v", st.admin.lastCancelAtEnd, tt.wantAtEnd)
			}
			assertGrants(t, st.admin.lastSync, tt.wantGrants, nil)
			if tt.wantExpiry && (st.admin.lastSync.ExpiresAt == nil || !st.admin.lastSync.ExpiresAt.Equal(periodEnd)) {
				t.Errorf("lejárat = %v, várt %v", st.admin.lastSync.ExpiresAt, periodEnd)
			}
			if !tt.wantExpiry && st.admin.lastSync.ExpiresAt != nil {
				t.Errorf("azonnali lemondásnál nem járhat lejárat: %v", st.admin.lastSync.ExpiresAt)
			}
		})
	}
}

func TestReactivateSubscription(t *testing.T) {
	past := time.Now().UTC().Add(-48 * time.Hour)
	future := time.Now().UTC().Add(48 * time.Hour)

	tests := []struct {
		name          string
		sub           store.Subscription
		body          string
		wantStatus    int
		wantNewPeriod bool
	}{
		{
			name:       "élő periódus megmarad",
			sub:        store.Subscription{ID: "s1", Subject: "u1", Plan: "pro", Status: store.SubActive, CancelAtPeriodEnd: true, CurrentPeriodEnd: future},
			body:       `{}`,
			wantStatus: http.StatusOK,
		},
		{
			// Lejárt periódusnál új időszak kell, különben azonnal lejárt
			// jogosultságot adnánk vissza.
			name:          "lejárt periódusnál új nyílik",
			sub:           store.Subscription{ID: "s1", Subject: "u1", Plan: "pro", Status: store.SubCanceled, CurrentPeriodEnd: past},
			body:          `{}`,
			wantStatus:    http.StatusOK,
			wantNewPeriod: true,
		},
		{
			name:          "kért időszakhossz",
			sub:           store.Subscription{ID: "s1", Subject: "u1", Plan: "pro", Status: store.SubCanceled, CurrentPeriodEnd: future},
			body:          `{"period_days":90}`,
			wantStatus:    http.StatusOK,
			wantNewPeriod: true,
		},
		{
			name:       "eltűnt csomag",
			sub:        store.Subscription{ID: "s1", Subject: "u1", Plan: "regi-csomag", Status: store.SubCanceled, CurrentPeriodEnd: future},
			body:       `{}`,
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &fakeStore{admin: adminFake{subscriptions: map[string]store.Subscription{"s1": tt.sub}}}
			h := newTestAPIWith(st, &fakeDirectory{})
			rec := do(t, h, http.MethodPost, "/api/v1/admin/subscriptions/s1/reactivate", "admin-token", tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			if st.admin.lastSync.ExpiresAt == nil {
				t.Fatal("a jogosultság lejárata nil")
			}
			isNew := st.admin.lastSync.ExpiresAt.After(future.Add(time.Hour))
			if isNew != tt.wantNewPeriod {
				t.Errorf("új periódus = %v, várt %v (lejárat: %v)", isNew, tt.wantNewPeriod, st.admin.lastSync.ExpiresAt)
			}
		})
	}
}

func TestInvoiceListFilters(t *testing.T) {
	st := &fakeStore{admin: adminFake{
		invoices: map[string]store.Invoice{
			"i1": {ID: "i1", Subject: "u1", AmountMinor: 299000, Currency: "HUF", Status: store.InvoiceOpen},
			"i2": {ID: "i2", Subject: "u1", AmountMinor: 100000, Currency: "HUF", Status: store.InvoicePaid},
			"i3": {ID: "i3", Subject: "u2", AmountMinor: 500000, Currency: "HUF", Status: store.InvoiceOpen},
		},
		profiles: map[string]store.Profile{"u1": {Subject: "u1", Email: "anna@example.com", DisplayName: "Anna"}},
	}}
	h := newTestAPIWith(st, &fakeDirectory{})

	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantCount  int
		wantOpen   int64
	}{
		{"mind", "", http.StatusOK, 3, 799000},
		{"csak nyitott", "?status=open", http.StatusOK, 2, 799000},
		{"felhasználóra szűrve", "?subject=u1", http.StatusOK, 2, 299000},
		{"ismeretlen státusz", "?status=fizetve", http.StatusBadRequest, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, h, http.MethodGet, "/api/v1/admin/invoices"+tt.query, "admin-token", "")
			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			var got invoicesResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("dekódolás: %v", err)
			}
			if len(got.Invoices) != tt.wantCount {
				t.Errorf("számlák = %d, várt %d", len(got.Invoices), tt.wantCount)
			}
			if got.Totals.OpenMinor != tt.wantOpen {
				t.Errorf("nyitott összeg = %d, várt %d", got.Totals.OpenMinor, tt.wantOpen)
			}
			if got.Currency != "HUF" {
				t.Errorf("pénznem = %q, várt HUF", got.Currency)
			}
			// A felhasználó neve a helyi profilból jön, hogy a tábla
			// kirajzolása ne igényeljen Keycloak-hívást.
			if tt.name == "felhasználóra szűrve" && got.Invoices[0].User.DisplayName != "Anna" {
				t.Errorf("felhasználó neve = %q, várt Anna", got.Invoices[0].User.DisplayName)
			}
		})
	}
}

func TestInvoiceActions(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantState  string
	}{
		{"kifizetettre jelölés", http.MethodPost, "/api/v1/admin/invoices/i1/pay", `{}`, http.StatusOK, store.InvoicePaid},
		{"sztornó", http.MethodPost, "/api/v1/admin/invoices/i1/void", `{"note":"duplikátum"}`, http.StatusOK, store.InvoiceVoid},
		{"már kifizetett számla sztornója", http.MethodPost, "/api/v1/admin/invoices/i2/void", `{}`, http.StatusConflict, ""},
		{"már kifizetett számla újrafizetése", http.MethodPost, "/api/v1/admin/invoices/i2/pay", `{}`, http.StatusConflict, ""},
		{"nem létező számla", http.MethodPost, "/api/v1/admin/invoices/nincs/pay", `{}`, http.StatusNotFound, ""},
		{"jövőbeli fizetési dátum", http.MethodPost, "/api/v1/admin/invoices/i1/pay", `{"paid_at":"2099-01-01T00:00:00Z"}`, http.StatusBadRequest, ""},
		{"számla kiállítása", http.MethodPost, "/api/v1/admin/invoices", `{"subject":"u1","amount_minor":299000}`, http.StatusCreated, store.InvoiceOpen},
		{"nulla összegű számla", http.MethodPost, "/api/v1/admin/invoices", `{"subject":"u1","amount_minor":0}`, http.StatusBadRequest, ""},
		{"subject nélküli számla", http.MethodPost, "/api/v1/admin/invoices", `{"amount_minor":100}`, http.StatusBadRequest, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paid := time.Now().UTC()
			st := &fakeStore{admin: adminFake{invoices: map[string]store.Invoice{
				"i1": {ID: "i1", Subject: "u1", AmountMinor: 299000, Currency: "HUF", Status: store.InvoiceOpen},
				"i2": {ID: "i2", Subject: "u1", AmountMinor: 299000, Currency: "HUF", Status: store.InvoicePaid, PaidAt: &paid},
			}}}
			h := newTestAPIWith(st, &fakeDirectory{})
			rec := do(t, h, tt.method, tt.path, "admin-token", tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantState == "" {
				return
			}
			var wrapper struct {
				Invoice store.Invoice `json:"invoice"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &wrapper); err != nil {
				t.Fatalf("dekódolás: %v", err)
			}
			if wrapper.Invoice.Status != tt.wantState {
				t.Errorf("státusz = %q, várt %q", wrapper.Invoice.Status, tt.wantState)
			}
			if tt.wantState == store.InvoicePaid && wrapper.Invoice.PaidAt == nil {
				t.Error("a kifizetés dátuma hiányzik")
			}
		})
	}
}

func TestOverview(t *testing.T) {
	ov := store.Overview{
		ProfileCount: 8, ActiveSubscriptionCount: 3, MRRMinor: 1397000,
		ActiveByPlan:   []store.PlanCount{{Plan: "pro", Count: 2}, {Plan: "studio", Count: 1}},
		NewProfiles30d: 4, NewSubscriptions30d: 2,
		UnpaidInvoiceCount: 1, UnpaidInvoiceTotalMinor: 299000,
	}

	tests := []struct {
		name         string
		dir          *fakeDirectory
		storeErr     error
		wantStatus   int
		wantUserCnt  *int
		wantWarnings int
	}{
		{
			name:        "teljes válasz",
			dir:         &fakeDirectory{users: []keycloak.User{{ID: "u1"}, {ID: "u2"}}},
			wantStatus:  http.StatusOK,
			wantUserCnt: intp(2),
		},
		{
			// A Keycloak kiesése nem tünteti el a többi számot, csak a
			// felhasználószámot – az admin lássa, hogy miért hiányzik.
			name:         "keycloak nem elérhető",
			dir:          &fakeDirectory{countErr: errors.New("connection refused")},
			wantStatus:   http.StatusOK,
			wantWarnings: 1,
		},
		{
			name:       "adatbázis hiba",
			dir:        &fakeDirectory{},
			storeErr:   errors.New("boom"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &fakeStore{admin: adminFake{overview: ov, err: tt.storeErr}}
			h := newTestAPIWith(st, tt.dir)
			rec := do(t, h, http.MethodGet, "/api/v1/admin/overview", "admin-token", "")
			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			var got overviewResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("dekódolás: %v", err)
			}
			if got.MRRMinor != ov.MRRMinor || got.ProfileCount != ov.ProfileCount {
				t.Errorf("összesítők nem mentek át: %+v", got.Overview)
			}
			if len(got.Plans) < 2 {
				t.Errorf("csomagok = %d, várt legalább 2", len(got.Plans))
			}
			if len(got.Warnings) != tt.wantWarnings {
				t.Errorf("figyelmeztetések = %d, várt %d", len(got.Warnings), tt.wantWarnings)
			}
			switch {
			case tt.wantUserCnt == nil && got.UserCount != nil:
				t.Errorf("felhasználószám = %d, várt null", *got.UserCount)
			case tt.wantUserCnt != nil && (got.UserCount == nil || *got.UserCount != *tt.wantUserCnt):
				t.Errorf("felhasználószám = %v, várt %d", got.UserCount, *tt.wantUserCnt)
			}
		})
	}
}

func intp(v int) *int { return &v }

func TestSubscriptionListFilterValidation(t *testing.T) {
	st := &fakeStore{admin: adminFake{subscriptions: map[string]store.Subscription{
		"s1": {ID: "s1", Subject: "u1", Plan: "pro", Status: store.SubActive},
		"s2": {ID: "s2", Subject: "u2", Plan: "studio", Status: store.SubCanceled},
	}}}
	h := newTestAPIWith(st, &fakeDirectory{})

	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantCount  int
	}{
		{"mind", "", http.StatusOK, 2},
		{"csomagra szűrve", "?plan=pro", http.StatusOK, 1},
		{"státuszra szűrve", "?status=canceled", http.StatusOK, 1},
		{"ismeretlen csomag", "?plan=arany", http.StatusBadRequest, 0},
		{"ismeretlen státusz", "?status=felfuggesztve", http.StatusBadRequest, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, h, http.MethodGet, "/api/v1/admin/subscriptions"+tt.query, "admin-token", "")
			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				return
			}
			var got subscriptionsResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("dekódolás: %v", err)
			}
			if len(got.Subscriptions) != tt.wantCount {
				t.Errorf("előfizetések = %d, várt %d", len(got.Subscriptions), tt.wantCount)
			}
		})
	}
}

func TestPaginationClamp(t *testing.T) {
	tests := []struct {
		query    string
		wantPage int
		wantSize int
	}{
		{"", 1, defaultPageSize},
		{"?page=3&page_size=10", 3, 10},
		{"?page=0", 1, defaultPageSize},
		{"?page=-2&page_size=-5", 1, defaultPageSize},
		{"?page_size=5000", 1, maxPageSize},
		{"?page=abc&page_size=xyz", 1, defaultPageSize},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users"+tt.query, nil)
			page, size := pagination(req)
			if page != tt.wantPage || size != tt.wantSize {
				t.Errorf("page=%d size=%d, várt page=%d size=%d", page, size, tt.wantPage, tt.wantSize)
			}
		})
	}
}
