package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

// liveSub egy élő előfizetés a naplózási tesztekhez.
func liveSub() map[string]store.Subscription {
	now := time.Now().UTC()
	return map[string]store.Subscription{"s1": {
		ID: "s1", Subject: "u1", Plan: "pro", Status: store.SubActive,
		CurrentPeriodStart: now.Add(-24 * time.Hour), CurrentPeriodEnd: now.Add(24 * time.Hour),
		PriceMinor: 299000, Currency: "HUF",
	}}
}

func openInvoice() map[string]store.Invoice {
	return map[string]store.Invoice{"i1": {
		ID: "i1", Number: "SNITT-2026-000042", Subject: "u1",
		AmountMinor: 299000, Currency: "HUF", Status: store.InvoiceOpen,
		IssuedAt: time.Now().UTC(),
	}}
}

// TestMutationsAreAudited: minden pénzt érintő admin művelet nyomot hagy a
// naplóban, a művelet gépi kulcsával és az érintett felhasználóval együtt.
func TestMutationsAreAudited(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		subs       map[string]store.Subscription
		invoices   map[string]store.Invoice
		wantStatus int
		wantAction string
		wantTarget string
	}{
		{
			name: "előfizetés kiadása", method: http.MethodPost, path: "/api/v1/admin/subscriptions",
			body: `{"subject":"u1","plan":"pro"}`, wantStatus: http.StatusCreated,
			wantAction: store.AuditSubscriptionGrant, wantTarget: "sub-new",
		},
		{
			name: "csomagváltás", method: http.MethodPatch, path: "/api/v1/admin/subscriptions/s1",
			body: `{"plan":"studio"}`, subs: liveSub(), wantStatus: http.StatusOK,
			wantAction: store.AuditSubscriptionChangePlan, wantTarget: "s1",
		},
		{
			name: "lemondás", method: http.MethodPost, path: "/api/v1/admin/subscriptions/s1/cancel",
			body: `{"at_period_end":true}`, subs: liveSub(), wantStatus: http.StatusOK,
			wantAction: store.AuditSubscriptionCancel, wantTarget: "s1",
		},
		{
			name: "visszakapcsolás", method: http.MethodPost, path: "/api/v1/admin/subscriptions/s1/reactivate",
			body: `{}`, subs: liveSub(), wantStatus: http.StatusOK,
			wantAction: store.AuditSubscriptionReactivate, wantTarget: "s1",
		},
		{
			name: "számla kiállítása", method: http.MethodPost, path: "/api/v1/admin/invoices",
			body: `{"subject":"u1","amount_minor":299000}`, wantStatus: http.StatusCreated,
			wantAction: store.AuditInvoiceCreate, wantTarget: "inv-new",
		},
		{
			name: "kifizetettre jelölés", method: http.MethodPost, path: "/api/v1/admin/invoices/i1/pay",
			body: `{}`, invoices: openInvoice(), wantStatus: http.StatusOK,
			wantAction: store.AuditInvoicePay, wantTarget: "i1",
		},
		{
			name: "sztornó", method: http.MethodPost, path: "/api/v1/admin/invoices/i1/void",
			body: `{"note":"téves kiállítás"}`, invoices: openInvoice(), wantStatus: http.StatusOK,
			wantAction: store.AuditInvoiceVoid, wantTarget: "i1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &fakeStore{admin: adminFake{subscriptions: tt.subs, invoices: tt.invoices}}
			h := newTestAPIWith(st, &fakeDirectory{})

			rec := do(t, h, tt.method, tt.path, "admin-token", tt.body)
			if rec.Code != tt.wantStatus {
				t.Fatalf("státusz = %d, várt %d (törzs: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if len(st.admin.audit) != 1 {
				t.Fatalf("naplósorok száma = %d, várt 1: %+v", len(st.admin.audit), st.admin.audit)
			}

			e := st.admin.audit[0]
			if e.Action != tt.wantAction {
				t.Errorf("művelet = %q, várt %q", e.Action, tt.wantAction)
			}
			if e.TargetID != tt.wantTarget {
				t.Errorf("cél = %q, várt %q", e.TargetID, tt.wantTarget)
			}
			if e.Subject != "u1" {
				t.Errorf("érintett felhasználó = %q, várt u1", e.Subject)
			}
			// A napló attól napló, hogy megmondja, ki csinálta.
			if e.ActorSubject != adminIdentity.Subject {
				t.Errorf("végrehajtó = %q, várt %q", e.ActorSubject, adminIdentity.Subject)
			}
			if e.ActorLabel != adminIdentity.Email {
				t.Errorf("végrehajtó címkéje = %q, várt %q", e.ActorLabel, adminIdentity.Email)
			}
			if e.Summary == "" {
				t.Error("az összefoglaló üres")
			}
			if len(e.Detail) == 0 {
				t.Error("a részletek üresek")
			}
		})
	}
}

// A számlával együtt kiadott előfizetés két külön eseményt hagy: a kiadást
// és a számlát. Az egyik nélkül a napló nem mondaná el, mi történt.
func TestGrantWithInvoiceAuditsBoth(t *testing.T) {
	st := &fakeStore{}
	h := newTestAPIWith(st, &fakeDirectory{})

	rec := do(t, h, http.MethodPost, "/api/v1/admin/subscriptions", "admin-token",
		`{"subject":"u1","plan":"pro","create_invoice":true}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("státusz = %d (törzs: %s)", rec.Code, rec.Body.String())
	}
	if len(st.admin.audit) != 2 {
		t.Fatalf("naplósorok száma = %d, várt 2: %+v", len(st.admin.audit), st.admin.audit)
	}
	if got := st.admin.audit[0].Action; got != store.AuditSubscriptionGrant {
		t.Errorf("első művelet = %q", got)
	}
	if got := st.admin.audit[1].Action; got != store.AuditInvoiceCreate {
		t.Errorf("második művelet = %q", got)
	}
}

// Ami nem történt meg, arról ne is legyen napló.
func TestFailedMutationIsNotAudited(t *testing.T) {
	st := &fakeStore{admin: adminFake{subscriptions: liveSub()}}
	h := newTestAPIWith(st, &fakeDirectory{})

	// Ugyanarra a felhasználóra már van élő előfizetés: ütközés.
	rec := do(t, h, http.MethodPost, "/api/v1/admin/subscriptions", "admin-token",
		`{"subject":"u1","plan":"pro"}`)
	if rec.Code != http.StatusConflict {
		t.Fatalf("státusz = %d, várt 409 (törzs: %s)", rec.Code, rec.Body.String())
	}
	if len(st.admin.audit) != 0 {
		t.Fatalf("sikertelen művelet naplósorai = %+v", st.admin.audit)
	}
}

// A naplóírás bukása nem bukatja el a már elvégzett műveletet: a pénzt
// érintő lépés megtörtént, hibás válasz csak félrevezetné az admint.
func TestAuditFailureDoesNotFailTheOperation(t *testing.T) {
	st := &fakeStore{admin: adminFake{
		invoices: openInvoice(),
		auditErr: errors.New("a napló nem elérhető"),
	}}
	h := newTestAPIWith(st, &fakeDirectory{})

	rec := do(t, h, http.MethodPost, "/api/v1/admin/invoices/i1/pay", "admin-token", `{}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("státusz = %d, várt 200 (törzs: %s)", rec.Code, rec.Body.String())
	}
	if st.admin.invoices["i1"].Status != store.InvoicePaid {
		t.Errorf("a számla státusza = %q, várt paid", st.admin.invoices["i1"].Status)
	}
}

func TestAuditListFilters(t *testing.T) {
	old := time.Now().UTC().Add(-72 * time.Hour)
	st := &fakeStore{admin: adminFake{audit: []store.AuditEntry{
		{ID: "a1", At: time.Now().UTC(), ActorSubject: "admin-1", Action: store.AuditInvoicePay, Subject: "u1"},
		{ID: "a2", At: time.Now().UTC(), ActorSubject: "admin-2", Action: store.AuditInvoiceVoid, Subject: "u2"},
		{ID: "a3", At: old, ActorSubject: "admin-1", Action: store.AuditSubscriptionGrant, Subject: "u1"},
	}}}
	h := newTestAPIWith(st, &fakeDirectory{})

	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{"szűrő nélkül", "", []string{"a1", "a2", "a3"}},
		{"felhasználó szerint", "?subject=u1", []string{"a1", "a3"}},
		{"végrehajtó szerint", "?actor=admin-2", []string{"a2"}},
		{"művelet szerint", "?action=" + store.AuditInvoicePay, []string{"a1"}},
		{"dátumtól", "?from=" + time.Now().UTC().Add(-24*time.Hour).Format(dateLayout), []string{"a1", "a2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, h, http.MethodGet, "/api/v1/admin/audit"+tt.query, "admin-token", "")
			if rec.Code != http.StatusOK {
				t.Fatalf("státusz = %d (törzs: %s)", rec.Code, rec.Body.String())
			}
			var resp struct {
				Entries []store.AuditEntry `json:"entries"`
				Total   int                `json:"total"`
				Actions []string           `json:"actions"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("válasz: %v", err)
			}
			got := make([]string, 0, len(resp.Entries))
			for _, e := range resp.Entries {
				got = append(got, e.ID)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("sorok = %v, várt %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("sorok = %v, várt %v", got, tt.want)
				}
			}
			if resp.Total != len(tt.want) {
				t.Errorf("total = %d, várt %d", resp.Total, len(tt.want))
			}
			// A szűrő legördülőjének a szerver adja a lehetséges értékeket.
			if len(resp.Actions) != len(store.AuditActions()) {
				t.Errorf("műveletlista = %v", resp.Actions)
			}
		})
	}
}

func TestAuditListRejectsUnknownFilters(t *testing.T) {
	h := newTestAPIWith(&fakeStore{}, &fakeDirectory{})

	for _, q := range []string{"?action=torles", "?from=tegnap", "?to=2026-13-45"} {
		rec := do(t, h, http.MethodGet, "/api/v1/admin/audit"+q, "admin-token", "")
		if rec.Code != http.StatusBadRequest {
			t.Errorf("%s: státusz = %d, várt 400 (törzs: %s)", q, rec.Code, rec.Body.String())
		}
	}
}

func TestMoneyText(t *testing.T) {
	tests := []struct {
		minor    int64
		currency string
		want     string
	}{
		{299000, "HUF", "2 990,00 HUF"},
		{0, "HUF", "0,00 HUF"},
		{5, "HUF", "0,05 HUF"},
		{123456789, "EUR", "1 234 567,89 EUR"},
		{-299000, "HUF", "-2 990,00 HUF"},
		{100000, "", "1 000,00"},
	}
	for _, tt := range tests {
		if got := moneyText(tt.minor, tt.currency); got != tt.want {
			t.Errorf("moneyText(%d, %q) = %q, várt %q", tt.minor, tt.currency, got, tt.want)
		}
	}
}
