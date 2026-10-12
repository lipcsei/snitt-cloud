package billing

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestEveryPaidPlanGrantsAppAccess(t *testing.T) {
	for _, p := range Plans() {
		found := false
		for _, k := range p.FeatureKeys {
			found = found || k == FeatureAppAccess
		}
		if !found {
			t.Errorf("a(z) %s csomag nem adja meg az alkalmazás használatát", p.Key)
		}
	}
}

func TestLifetimePlanNeverExpires(t *testing.T) {
	p, ok := LookupPlan("lifetime")
	if !ok || !p.Lifetime || !p.SelfService {
		t.Fatalf("lifetime csomag = %+v", p)
	}
	now := time.Now()
	if end := p.PeriodEnd(now, 30); end.Before(now.AddDate(99, 0, 0)) {
		t.Errorf("periódus vége = %v", end)
	}
	if p.EntitlementExpiry(now) != nil {
		t.Error("lifetime jogosultságnak nem lehet lejárata")
	}
	m, _ := LookupPlan("monthly")
	if exp := m.EntitlementExpiry(now); exp == nil || !exp.Equal(now) {
		t.Errorf("havi csomag lejárata = %v", exp)
	}
}

func TestStripeSignature(t *testing.T) {
	payload := []byte(`{"id":"evt_1","type":"ping"}`)
	now := time.Now()
	s := NewStripe(StripeConfig{WebhookSecret: "whsec"})

	if _, err := s.ParseWebhook(payload, SignStripePayload(payload, "whsec", now), now); err != nil {
		t.Errorf("érvényes aláírás elutasítva: %v", err)
	}
	cases := map[string]string{
		"rossz titok":        SignStripePayload(payload, "other", now),
		"régi időbélyeg":     SignStripePayload(payload, "whsec", now.Add(-10*time.Minute)),
		"hiányzó fejléc":     "",
		"értelmetlen fejléc": "garbage",
	}
	for name, header := range cases {
		if _, err := s.ParseWebhook(payload, header, now); !errors.Is(err, ErrBadSignature) {
			t.Errorf("%s: hiba = %v, várt ErrBadSignature", name, err)
		}
	}
	if _, err := s.ParseWebhook([]byte(`{"x":1}`), SignStripePayload(payload, "whsec", now), now); !errors.Is(err, ErrBadSignature) {
		t.Errorf("módosított törzs elfogadva: %v", err)
	}
}

func TestStripeIgnoresUnpaidCheckout(t *testing.T) {
	payload := []byte(`{"id":"evt","type":"checkout.session.completed","data":{"object":{"payment_status":"unpaid","metadata":{"subject":"u","plan":"lifetime"}}}}`)
	now := time.Now()
	ev, err := NewStripe(StripeConfig{WebhookSecret: "w"}).ParseWebhook(payload, SignStripePayload(payload, "w", now), now)
	if err != nil || ev.Kind != WebhookIgnored {
		t.Errorf("kifizetetlen checkout: kind=%q err=%v", ev.Kind, err)
	}
}

func TestCheckoutManagedPayments(t *testing.T) {
	lifetime, ok := LookupPlan("lifetime")
	if !ok {
		t.Fatal("nincs lifetime csomag")
	}
	for _, managed := range []bool{true, false} {
		var form url.Values
		var version string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = r.ParseForm()
			form, version = r.PostForm, r.Header.Get("Stripe-Version")
			_, _ = w.Write([]byte(`{"url":"https://checkout.stripe.test/cs_1"}`))
		}))
		s := NewStripe(StripeConfig{PriceIDs: map[string]string{"lifetime": "price_l"}, APIBase: srv.URL, ManagedPayments: managed})
		_, err := s.CreateCheckout(context.Background(), CheckoutRequest{Plan: lifetime, Subject: "u", SuccessURL: "https://x", CancelURL: "https://y"})
		srv.Close()
		if err != nil {
			t.Fatal(err)
		}
		if version != StripeAPIVersion {
			t.Errorf("Stripe-Version = %q", version)
		}
		if got := form.Get("managed_payments[enabled]") == "true"; got != managed {
			t.Errorf("managed=%v: managed_payments[enabled] = %q", managed, form.Get("managed_payments[enabled]"))
		}
		// Managed Paymentsnél a számlát a Stripe adja, saját nem kérhető.
		if got := form.Get("invoice_creation[enabled]") == "true"; got == managed {
			t.Errorf("managed=%v: invoice_creation = %q", managed, form.Get("invoice_creation[enabled]"))
		}
	}
}
