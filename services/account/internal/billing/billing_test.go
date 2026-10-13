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

func TestAsyncPaymentSucceededCompletesCheckout(t *testing.T) {
	now := time.Now()
	payload := []byte(`{"id":"evt","type":"checkout.session.async_payment_succeeded","data":{"object":{"payment_status":"paid","metadata":{"subject":"u","plan":"lifetime"},"payment_intent":"pi_1","amount_total":2499000,"currency":"huf"}}}`)
	ev, err := NewStripe(StripeConfig{WebhookSecret: "w"}).ParseWebhook(payload, SignStripePayload(payload, "w", now), now)
	if err != nil || ev.Kind != WebhookCheckoutCompleted || ev.Subject != "u" || ev.PaymentID != "pi_1" {
		t.Errorf("késleltetett fizetés: %+v err=%v", ev, err)
	}
}

func TestAIPlansAreNotSubscriptionPlans(t *testing.T) {
	for _, key := range []string{"ai_monthly", "ai_tokens"} {
		p, ok := LookupAnyPlan(key)
		if !ok || !p.IsAI() || !p.SelfService || len(p.FeatureKeys) != 0 || p.AITokens <= 0 {
			t.Errorf("%s = %+v", key, p)
		}
		if _, ok := LookupPlan(key); ok {
			t.Errorf("a(z) %s nem lehet előfizetésként kiadható csomag", key)
		}
		for _, sp := range Plans() {
			if sp.Key == key {
				t.Errorf("a(z) %s nem szerepelhet a Plans() listában", key)
			}
		}
	}
	if m, _ := LookupAnyPlan("ai_monthly"); m.Lifetime || m.AITokens != 1_000_000 || m.PriceMinor != 299000 {
		t.Errorf("ai_monthly = %+v", m)
	}
	if tk, _ := LookupAnyPlan("ai_tokens"); !tk.Lifetime || tk.AITokens != 500_000 || tk.PriceMinor != 199000 {
		t.Errorf("ai_tokens = %+v", tk)
	}
	if got := ManagedFeatureKeys(); len(got) != 3 {
		t.Errorf("a vezérelt jogosultságok nem változhatnak: %v", got)
	}
	if len(AllPlans()) != len(Plans())+2 {
		t.Errorf("AllPlans = %d, Plans = %d", len(AllPlans()), len(Plans()))
	}
}

func TestStripeOffersOnlyPricedPlans(t *testing.T) {
	s := NewStripe(StripeConfig{PriceIDs: map[string]string{"monthly": "price_m", "ai_tokens": ""}})
	if !s.Offers("monthly") || s.Offers("ai_tokens") || s.Offers("ai_monthly") {
		t.Error("csak a beárazott csomag kínálható")
	}
}

func TestInvoicePaidExposesSubscriptionMetadataAndPeriod(t *testing.T) {
	now := time.Now()
	payload := []byte(`{"id":"evt","type":"invoice.paid","data":{"object":{"id":"in_1","billing_reason":"subscription_cycle",
		"amount_paid":299000,"currency":"huf",
		"parent":{"subscription_details":{"subscription":"sub_ai","metadata":{"subject":"u","plan":"ai_monthly"}}},
		"lines":{"data":[{"period":{"start":1700000000,"end":1702592000}}]}}}}`)
	ev, err := NewStripe(StripeConfig{WebhookSecret: "w"}).ParseWebhook(payload, SignStripePayload(payload, "w", now), now)
	if err != nil || ev.Kind != WebhookInvoicePaid || ev.SubscriptionID != "sub_ai" || ev.PlanKey != "ai_monthly" || ev.Subject != "u" {
		t.Fatalf("esemény = %+v err=%v", ev, err)
	}
	if ev.PeriodStart.Unix() != 1700000000 || ev.PeriodEnd.Unix() != 1702592000 {
		t.Errorf("időszak = %v - %v", ev.PeriodStart, ev.PeriodEnd)
	}
}
