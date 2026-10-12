package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/lipcsei/snitt-cloud/services/account/internal/auth"
	"github.com/lipcsei/snitt-cloud/services/account/internal/billing"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

// ---------- a webhookhoz kellő kamu store műveletek ----------

func (f *fakeStore) SubscriptionByExternalID(_ context.Context, provider, externalID string) (store.Subscription, error) {
	for _, s := range f.admin.subscriptions {
		if s.Provider == provider && externalID != "" && s.ExternalSubscriptionID == externalID {
			return s, nil
		}
	}
	return store.Subscription{}, store.ErrNotFound
}

func (f *fakeStore) ApplyProviderUpdate(_ context.Context, id string, upd store.ProviderUpdate, sync store.EntitlementSync) (store.Subscription, error) {
	f.admin.lastProviderUpdate, f.admin.lastSync = upd, sync
	s, ok := f.admin.subscriptions[id]
	if !ok {
		return store.Subscription{}, store.ErrNotFound
	}
	s.Status, s.CurrentPeriodStart, s.CurrentPeriodEnd, s.CancelAtPeriodEnd = upd.Status, upd.PeriodStart, upd.PeriodEnd, upd.CancelAtPeriodEnd
	f.admin.subscriptions[id] = s
	return s, nil
}

func (f *fakeStore) ProviderEventSeen(_ context.Context, provider, eventID string) (bool, error) {
	return f.admin.providerEvents[provider+"/"+eventID], nil
}

func (f *fakeStore) RecordProviderEvent(_ context.Context, provider, eventID, _ string) error {
	if f.admin.providerEvents == nil {
		f.admin.providerEvents = map[string]bool{}
	}
	f.admin.providerEvents[provider+"/"+eventID] = true
	return nil
}

// ---------- segédek ----------

const testWebhookSecret = "whsec_test"

// fakeStripe a Stripe API Checkout végpontját utánozza, és megjegyzi a kérést.
func fakeStripe(t *testing.T) (*httptest.Server, *url.Values) {
	t.Helper()
	var last url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, _, _ := r.BasicAuth(); user != "sk_test" {
			http.Error(w, `{"error":{"message":"bad key"}}`, http.StatusUnauthorized)
			return
		}
		_ = r.ParseForm()
		last = r.PostForm
		if r.URL.Path == "/v1/checkout/sessions" {
			_, _ = w.Write([]byte(`{"id":"cs_1","url":"https://checkout.stripe.test/cs_1"}`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &last
}

func newStripeAPI(t *testing.T, st *fakeStore, apiBase string) http.Handler {
	t.Helper()
	return New(Options{
		Store:    st,
		Verifier: fakeVerifier{valid: map[string]auth.Identity{"good-token": testIdentity}},
		Billing: billing.NewStripe(billing.StripeConfig{
			SecretKey:     "sk_test",
			WebhookSecret: testWebhookSecret,
			PriceIDs:      map[string]string{"monthly": "price_m", "lifetime": "price_l"},
			APIBase:       apiBase,
		}),
		DefaultCurrency:    "HUF",
		Log:                slog.New(slog.NewJSONHandler(io.Discard, nil)),
		CheckoutSuccessURL: "https://snitt.test/ok",
		CheckoutCancelURL:  "https://snitt.test/cancel",
	}).Handler()
}

func postWebhook(t *testing.T, h http.Handler, payload string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/webhook", strings.NewReader(payload))
	req.Header.Set("Stripe-Signature", billing.SignStripePayload([]byte(payload), testWebhookSecret, time.Now()))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// ---------- tesztek ----------

func TestPublicPlansListsSelfServicePlans(t *testing.T) {
	h := newTestAPIWith(&fakeStore{}, &fakeDirectory{})
	rec := do(t, h, http.MethodGet, "/api/v1/plans", "", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("státusz = %d", rec.Code)
	}
	var body struct {
		Plans           []publicPlan `json:"plans"`
		CheckoutEnabled bool         `json:"checkout_enabled"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.CheckoutEnabled {
		t.Error("kézi módban a checkout nem lehet bekapcsolva")
	}
	keys := []string{}
	for _, p := range body.Plans {
		keys = append(keys, p.Key)
	}
	if strings.Join(keys, ",") != "monthly,lifetime" {
		t.Errorf("csomagok = %v, várt monthly,lifetime (ár szerint)", keys)
	}
}

func TestCheckoutWithoutProviderIs501(t *testing.T) {
	h := newTestAPIWith(&fakeStore{}, &fakeDirectory{})
	rec := do(t, h, http.MethodPost, "/api/v1/checkout", "good-token", `{"plan":"monthly"}`)
	if rec.Code != http.StatusNotImplemented {
		t.Fatalf("státusz = %d, várt 501", rec.Code)
	}
}

func TestCheckoutCreatesStripeSession(t *testing.T) {
	srv, last := fakeStripe(t)
	h := newStripeAPI(t, &fakeStore{}, srv.URL)

	rec := do(t, h, http.MethodPost, "/api/v1/checkout", "good-token", `{"plan":"lifetime"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("státusz = %d (%s)", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "https://checkout.stripe.test/cs_1") {
		t.Errorf("válasz = %s", rec.Body.String())
	}
	got := *last
	if got.Get("mode") != "payment" || got.Get("line_items[0][price]") != "price_l" {
		t.Errorf("lifetime checkout paraméterek: mode=%q price=%q", got.Get("mode"), got.Get("line_items[0][price]"))
	}
	if got.Get("metadata[subject]") != testIdentity.Subject || got.Get("metadata[plan]") != "lifetime" {
		t.Errorf("metadata = %v", got)
	}

	rec = do(t, h, http.MethodPost, "/api/v1/checkout", "good-token", `{"plan":"monthly"}`)
	if rec.Code != http.StatusOK || (*last).Get("mode") != "subscription" {
		t.Errorf("havi checkout: státusz %d, mode %q", rec.Code, (*last).Get("mode"))
	}
}

func TestCheckoutRejects(t *testing.T) {
	srv, _ := fakeStripe(t)
	tests := []struct {
		name   string
		body   string
		token  string
		subs   map[string]store.Subscription
		status int
	}{
		{"token nélkül", `{"plan":"monthly"}`, "", nil, http.StatusUnauthorized},
		{"admin csomag nem vehető meg", `{"plan":"pro"}`, "good-token", nil, http.StatusBadRequest},
		{"már van élő előfizetés", `{"plan":"lifetime"}`, "good-token",
			map[string]store.Subscription{"s1": {ID: "s1", Subject: testIdentity.Subject, Plan: "monthly", Status: store.SubActive}},
			http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStripeAPI(t, &fakeStore{admin: adminFake{subscriptions: tt.subs}}, srv.URL)
			rec := do(t, h, http.MethodPost, "/api/v1/checkout", tt.token, tt.body)
			if rec.Code != tt.status {
				t.Errorf("státusz = %d, várt %d (%s)", rec.Code, tt.status, rec.Body.String())
			}
		})
	}
}

func TestWebhookRejectsBadSignature(t *testing.T) {
	h := newStripeAPI(t, &fakeStore{}, "http://unused")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/webhook", strings.NewReader(`{"id":"evt"}`))
	req.Header.Set("Stripe-Signature", "t=1,v1=00")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("státusz = %d, várt 400", rec.Code)
	}
}

func TestWebhookLifetimePurchaseGrantsPermanentAccess(t *testing.T) {
	st := &fakeStore{}
	h := newStripeAPI(t, st, "http://unused")
	payload := `{"id":"evt_1","type":"checkout.session.completed","data":{"object":{
		"mode":"payment","payment_status":"paid","client_reference_id":"user-123",
		"metadata":{"subject":"user-123","plan":"lifetime"},"customer":"cus_1",
		"payment_intent":"pi_1","amount_total":2499000,"currency":"huf"}}}`

	rec := postWebhook(t, h, payload)
	if rec.Code != http.StatusOK {
		t.Fatalf("státusz = %d (%s)", rec.Code, rec.Body.String())
	}
	sub := st.admin.lastNewSub
	if sub.Plan != "lifetime" || sub.Provider != "stripe" || sub.ExternalCustomerID != "cus_1" {
		t.Errorf("előfizetés = %+v", sub)
	}
	if sub.PeriodEnd.Before(time.Now().AddDate(50, 0, 0)) {
		t.Errorf("lifetime periódus vége túl közeli: %v", sub.PeriodEnd)
	}
	if st.admin.lastSync.ExpiresAt != nil {
		t.Error("lifetime jogosultságnak nem lehet lejárata")
	}
	if len(st.admin.lastSync.Grant) != 1 || st.admin.lastSync.Grant[0] != billing.FeatureAppAccess {
		t.Errorf("kiadott jogosultságok = %v", st.admin.lastSync.Grant)
	}
	inv := st.admin.invoices["inv-new"]
	if inv.Status != store.InvoicePaid || inv.ExternalInvoiceID != "pi_1" || inv.Currency != "HUF" {
		t.Errorf("számla = %+v", inv)
	}

	// Ugyanaz az esemény még egyszer: nem ad ki második előfizetést.
	st.admin.createdInvoices = 0
	rec = postWebhook(t, h, payload)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "duplicate") || st.admin.createdInvoices != 0 {
		t.Errorf("ismételt esemény: %d %s, új számlák: %d", rec.Code, rec.Body.String(), st.admin.createdInvoices)
	}
}

func TestWebhookSubscriptionLifecycle(t *testing.T) {
	st := &fakeStore{}
	h := newStripeAPI(t, st, "http://unused")

	rec := postWebhook(t, h, `{"id":"evt_c","type":"checkout.session.completed","data":{"object":{
		"mode":"subscription","payment_status":"paid","metadata":{"subject":"user-9","plan":"monthly"},
		"customer":"cus_9","subscription":"sub_9","invoice":"in_9","amount_total":149000,"currency":"huf"}}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("checkout: %d %s", rec.Code, rec.Body.String())
	}
	if st.admin.lastSync.ExpiresAt == nil {
		t.Fatal("havi előfizetés jogosultságának lejárata kell legyen")
	}

	start, end := time.Now().Unix(), time.Now().AddDate(0, 1, 0).Unix()
	rec = postWebhook(t, h, `{"id":"evt_u","type":"customer.subscription.updated","data":{"object":{
		"id":"sub_9","status":"active","cancel_at_period_end":true,
		"items":{"data":[{"current_period_start":`+itoa(start)+`,"current_period_end":`+itoa(end)+`}]}}}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: %d %s", rec.Code, rec.Body.String())
	}
	upd := st.admin.lastProviderUpdate
	if upd.Status != store.SubActive || !upd.CancelAtPeriodEnd || upd.PeriodEnd.Unix() != end {
		t.Errorf("frissítés = %+v", upd)
	}
	if len(st.admin.lastSync.Grant) == 0 || st.admin.lastSync.ExpiresAt.Unix() != end {
		t.Errorf("a jogosultság a periódus végéig kell járjon: %+v", st.admin.lastSync)
	}

	rec = postWebhook(t, h, `{"id":"evt_d","type":"customer.subscription.deleted","data":{"object":{"id":"sub_9","status":"canceled"}}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: %d %s", rec.Code, rec.Body.String())
	}
	if st.admin.lastProviderUpdate.Status != store.SubCanceled || len(st.admin.lastSync.Grant) != 0 {
		t.Errorf("megszűnéskor a jogosultságot vissza kell vonni: %+v / %+v", st.admin.lastProviderUpdate, st.admin.lastSync)
	}
}

func TestWebhookUnknownSubscriptionAsksForRetry(t *testing.T) {
	h := newStripeAPI(t, &fakeStore{}, "http://unused")
	rec := postWebhook(t, h, `{"id":"evt_x","type":"customer.subscription.updated","data":{"object":{"id":"sub_unknown","status":"active"}}}`)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("státusz = %d, várt 500 (a Stripe később újraküldi)", rec.Code)
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }
