package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/lipcsei/snitt-cloud/services/account/internal/ai"
	"github.com/lipcsei/snitt-cloud/services/account/internal/auth"
	"github.com/lipcsei/snitt-cloud/services/account/internal/billing"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

// ---------- kamu AI tárca ----------

// aiFake a memóriában tartott AI tárcák. A levonás sorrendjét és a nullánál
// megálló vágást ugyanaz a store.AIWallet.Debit végzi, mint élesben.
type aiFake struct {
	wallets map[string]store.AIWallet
	usage   []store.NewAIUsage
	err     error
}

func (a *aiFake) ensure() {
	if a.wallets == nil {
		a.wallets = map[string]store.AIWallet{}
	}
}

func (f *fakeStore) AIWallet(_ context.Context, subject string) (store.AIWallet, error) {
	if f.wallets.err != nil {
		return store.AIWallet{}, f.wallets.err
	}
	if w, ok := f.wallets.wallets[subject]; ok {
		return w, nil
	}
	return store.AIWallet{Subject: subject}, nil
}

func (f *fakeStore) AIWalletByExternalSubscription(_ context.Context, externalID string) (store.AIWallet, error) {
	for _, w := range f.wallets.wallets {
		if externalID != "" && w.ExternalSubscriptionID == externalID {
			return w, nil
		}
	}
	return store.AIWallet{}, store.ErrNotFound
}

func (f *fakeStore) DebitAITokens(_ context.Context, subject string, u store.NewAIUsage) (store.AIWallet, error) {
	if f.wallets.err != nil {
		return store.AIWallet{}, f.wallets.err
	}
	f.wallets.ensure()
	w := f.wallets.wallets[subject]
	w.Subject = subject
	w = w.Debit(time.Now(), u.ChargedTokens)
	f.wallets.wallets[subject] = w
	f.wallets.usage = append(f.wallets.usage, u)
	return w, nil
}

func (f *fakeStore) CreditAITokens(_ context.Context, subject string, tokens int64) (store.AIWallet, error) {
	f.wallets.ensure()
	w := f.wallets.wallets[subject]
	w.Subject = subject
	w.PurchasedTokens += tokens
	f.wallets.wallets[subject] = w
	return w, nil
}

func (f *fakeStore) SetAISubscription(_ context.Context, subject string, in store.AISubscription) (store.AIWallet, error) {
	f.wallets.ensure()
	w := f.wallets.wallets[subject]
	w.Subject = subject
	end := in.PeriodEnd
	w.ExternalSubscriptionID, w.Status, w.PeriodTokens, w.PeriodEnd = in.ExternalSubscriptionID, in.Status, in.PeriodTokens, &end
	f.wallets.wallets[subject] = w
	return w, nil
}

func (f *fakeStore) UpdateAISubscription(_ context.Context, externalID string, upd store.AISubscriptionUpdate) (store.AIWallet, error) {
	for subject, w := range f.wallets.wallets {
		if externalID == "" || w.ExternalSubscriptionID != externalID {
			continue
		}
		if upd.Status != "" {
			w.Status = upd.Status
		}
		if upd.PeriodEnd != nil {
			w.PeriodEnd = upd.PeriodEnd
		}
		if upd.PeriodTokens != nil {
			w.PeriodTokens = *upd.PeriodTokens
		}
		f.wallets.wallets[subject] = w
		return w, nil
	}
	return store.AIWallet{}, store.ErrNotFound
}

// ---------- kamu AI szolgáltatók ----------

type fakeEmbedder struct {
	res   ai.EmbedResult
	err   error
	calls int
}

func (e *fakeEmbedder) Model() string { return "voyage-4" }

func (e *fakeEmbedder) Embed(_ context.Context, texts []string, _ string) (ai.EmbedResult, error) {
	e.calls++
	if e.err != nil {
		return ai.EmbedResult{}, e.err
	}
	res := e.res
	res.Embeddings = make([][]float32, len(texts))
	for i := range texts {
		res.Embeddings[i] = []float32{0.1, 0.2}
	}
	return res, nil
}

type fakeAnswerer struct {
	res   ai.AnswerResult
	err   error
	calls int
	last  ai.AnswerRequest
}

func (a *fakeAnswerer) Model() string { return "claude-opus-5-5" }

func (a *fakeAnswerer) Answer(_ context.Context, req ai.AnswerRequest) (ai.AnswerResult, error) {
	a.calls++
	a.last = req
	return a.res, a.err
}

func newAIAPI(st *fakeStore, emb Embedder, ans Answerer) http.Handler {
	return New(Options{
		Store:           st,
		Verifier:        fakeVerifier{valid: map[string]auth.Identity{"good-token": testIdentity}},
		DefaultCurrency: "HUF",
		Log:             slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Embedder:        emb,
		Answerer:        ans,
	}).Handler()
}

func walletStore(w store.AIWallet) *fakeStore {
	w.Subject = testIdentity.Subject
	return &fakeStore{wallets: aiFake{wallets: map[string]store.AIWallet{testIdentity.Subject: w}}}
}

func ptrTime(t time.Time) *time.Time { return &t }

const answerBody = `{"question":"Hogyan főzzem a rizst?","language":"hu","passages":[
	{"id":"p1","video_title":"Főzés","start_ms":1000,"text":"főzd"},
	{"id":"p2","video_title":"Főzés","start_ms":61000,"text":"mosd"}]}`

// ---------- tesztek: AI végpontok ----------

func TestAIBalance(t *testing.T) {
	end := time.Now().Add(10 * 24 * time.Hour).UTC().Truncate(time.Second)
	st := walletStore(store.AIWallet{Status: store.SubActive, PeriodEnd: &end, PeriodTokens: 700, PurchasedTokens: 300})
	rec := do(t, newAIAPI(st, &fakeEmbedder{}, nil), http.MethodGet, "/api/v1/ai/balance", "good-token", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("státusz = %d", rec.Code)
	}
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got["enabled"] != true || got["period_tokens"] != float64(700) || got["purchased_tokens"] != float64(300) ||
		got["total_tokens"] != float64(1000) || got["subscription_active"] != true || got["period_end"] != end.Format(time.RFC3339) {
		t.Errorf("egyenleg = %v", got)
	}

	// Sosem vásárolt felhasználó, szolgáltató nélkül.
	rec = do(t, newAIAPI(&fakeStore{}, nil, nil), http.MethodGet, "/api/v1/ai/balance", "good-token", "")
	got = nil
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got["enabled"] != false || got["total_tokens"] != float64(0) || got["period_end"] != nil || got["subscription_active"] != false {
		t.Errorf("üres egyenleg = %v", got)
	}

	if rec := do(t, newAIAPI(&fakeStore{}, nil, nil), http.MethodGet, "/api/v1/ai/balance", "", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("token nélkül: %d", rec.Code)
	}
}

func TestAIEndpointsWithoutProviderAre503(t *testing.T) {
	h := newAIAPI(walletStore(store.AIWallet{PurchasedTokens: 1000}), nil, nil)
	if rec := do(t, h, http.MethodPost, "/api/v1/ai/embed", "good-token", `{"texts":["a"],"input_type":"query"}`); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("embed: %d", rec.Code)
	}
	if rec := do(t, h, http.MethodPost, "/api/v1/ai/answer", "good-token", answerBody); rec.Code != http.StatusServiceUnavailable {
		t.Errorf("answer: %d", rec.Code)
	}
}

func TestAIEndpointsWithoutTokensAre402(t *testing.T) {
	expired := time.Now().Add(-time.Hour)
	stores := map[string]*fakeStore{
		"üres tárca":   {},
		"lejárt keret": walletStore(store.AIWallet{Status: store.SubActive, PeriodEnd: &expired, PeriodTokens: 5000}),
	}
	for name, st := range stores {
		t.Run(name, func(t *testing.T) {
			emb, ans := &fakeEmbedder{}, &fakeAnswerer{}
			h := newAIAPI(st, emb, ans)
			rec := do(t, h, http.MethodPost, "/api/v1/ai/embed", "good-token", `{"texts":["a"],"input_type":"query"}`)
			if rec.Code != http.StatusPaymentRequired || !strings.Contains(rec.Body.String(), "Elfogytak az AI tokenjeid") {
				t.Errorf("embed: %d %s", rec.Code, rec.Body.String())
			}
			rec = do(t, h, http.MethodPost, "/api/v1/ai/answer", "good-token", answerBody)
			if rec.Code != http.StatusPaymentRequired {
				t.Errorf("answer: %d", rec.Code)
			}
			if emb.calls+ans.calls != 0 {
				t.Error("egyenleg nélkül nem hívhatjuk a szolgáltatót")
			}
		})
	}
}

func TestAIEmbedChargesPeriodTokensFirst(t *testing.T) {
	end := time.Now().Add(24 * time.Hour)
	st := walletStore(store.AIWallet{Status: store.SubActive, PeriodEnd: &end, PeriodTokens: 100, PurchasedTokens: 1000})
	emb := &fakeEmbedder{res: ai.EmbedResult{Model: "voyage-4", Dims: 2, InputTokens: 10_000}}
	rec := do(t, newAIAPI(st, emb, nil), http.MethodPost, "/api/v1/ai/embed", "good-token", `{"texts":["egy","kettő"],"input_type":"document"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("státusz = %d (%s)", rec.Code, rec.Body.String())
	}
	var got aiEmbedResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	// 10 000 Voyage token * 0,02 = 200 Snitt AI token: 100 a keretből, 100 a vásároltból.
	if got.ChargedTokens != 200 || got.TotalTokens != 900 || got.Dims != 2 || len(got.Embeddings) != 2 || got.Model != "voyage-4" {
		t.Errorf("válasz = %+v", got)
	}
	w := st.wallets.wallets[testIdentity.Subject]
	if w.PeriodTokens != 0 || w.PurchasedTokens != 900 {
		t.Errorf("tárca = %+v", w)
	}
	if len(st.wallets.usage) != 1 || st.wallets.usage[0].Kind != store.AIUsageEmbed || st.wallets.usage[0].InputTokens != 10_000 {
		t.Errorf("használati napló = %+v", st.wallets.usage)
	}
}

func TestAIAnswerChargesAndClampsAtZero(t *testing.T) {
	expired := time.Now().Add(-time.Hour)
	// A lejárt keret nem költhető, és nem is fogy: a vásárolt egyenleg viszi a terhet.
	st := walletStore(store.AIWallet{Status: store.SubActive, PeriodEnd: &expired, PeriodTokens: 5000, PurchasedTokens: 1500})
	ans := &fakeAnswerer{res: ai.AnswerResult{
		Answer: "Mosd meg [2], aztán főzd [1].", CitedIDs: []string{"p2", "p1"},
		Model: "claude-opus-5-5", InputTokens: 1000, OutputTokens: 200,
	}}
	rec := do(t, newAIAPI(st, nil, ans), http.MethodPost, "/api/v1/ai/answer", "good-token", answerBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("státusz = %d (%s)", rec.Code, rec.Body.String())
	}
	var got aiAnswerResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	// 1000 * 1 + 200 * 5 = 2000, de csak 1500 volt: a különbözetet elnyeljük.
	if got.ChargedTokens != 2000 || got.TotalTokens != 0 || strings.Join(got.CitedIDs, ",") != "p2,p1" || got.Model != "claude-opus-5-5" {
		t.Errorf("válasz = %+v", got)
	}
	w := st.wallets.wallets[testIdentity.Subject]
	if w.PurchasedTokens != 0 || w.PeriodTokens != 5000 {
		t.Errorf("tárca = %+v", w)
	}
	if ans.last.Language != "hu" || len(ans.last.Passages) != 2 || ans.last.Passages[1].StartMS != 61000 {
		t.Errorf("a szolgáltató ezt kapta: %+v", ans.last)
	}
}

func TestAIAnswerRefusalIs502AndCharged(t *testing.T) {
	st := walletStore(store.AIWallet{PurchasedTokens: 10_000})
	ans := &fakeAnswerer{res: ai.AnswerResult{Model: "claude-opus-5-5", InputTokens: 800, OutputTokens: 2}, err: ai.ErrRefusal}
	rec := do(t, newAIAPI(st, nil, ans), http.MethodPost, "/api/v1/ai/answer", "good-token", answerBody)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("státusz = %d", rec.Code)
	}
	if w := st.wallets.wallets[testIdentity.Subject]; w.PurchasedTokens != 10_000-810 {
		t.Errorf("tárca = %+v", w)
	}
}

func TestAIProviderErrorIsNotCharged(t *testing.T) {
	st := walletStore(store.AIWallet{PurchasedTokens: 10_000})
	rec := do(t, newAIAPI(st, &fakeEmbedder{err: errors.New("hálózat")}, nil), http.MethodPost, "/api/v1/ai/embed", "good-token", `{"texts":["a"],"input_type":"query"}`)
	if rec.Code != http.StatusBadGateway || len(st.wallets.usage) != 0 {
		t.Errorf("státusz = %d, napló = %v", rec.Code, st.wallets.usage)
	}
}

func TestAIRequestValidation(t *testing.T) {
	long := strings.Repeat("é", 8001)
	many := `"` + strings.Repeat(`a","`, 128) + `a"`
	passages41 := strings.TrimSuffix(strings.Repeat(`{"id":"x","text":"t"},`, 41), ",")
	bigText := strings.Repeat("a", 60001)
	tests := []struct {
		name, path, body string
	}{
		{"rossz input_type", "/api/v1/ai/embed", `{"texts":["a"],"input_type":"x"}`},
		{"nincs szöveg", "/api/v1/ai/embed", `{"texts":[],"input_type":"query"}`},
		{"túl sok szöveg", "/api/v1/ai/embed", `{"texts":[` + many + `],"input_type":"query"}`},
		{"túl hosszú szöveg", "/api/v1/ai/embed", `{"texts":["` + long + `"],"input_type":"query"}`},
		{"ismeretlen mező", "/api/v1/ai/embed", `{"texts":["a"],"input_type":"query","x":1}`},
		{"rövid kérdés", "/api/v1/ai/answer", `{"question":"a","passages":[{"id":"p","text":"t"}]}`},
		{"hosszú kérdés", "/api/v1/ai/answer", `{"question":"` + strings.Repeat("k", 1001) + `","passages":[{"id":"p","text":"t"}]}`},
		{"nincs részlet", "/api/v1/ai/answer", `{"question":"kérdés","passages":[]}`},
		{"túl sok részlet", "/api/v1/ai/answer", `{"question":"kérdés","passages":[` + passages41 + `]}`},
		{"túl sok szöveg összesen", "/api/v1/ai/answer", `{"question":"kérdés","passages":[{"id":"p","text":"` + bigText + `"}]}`},
		{"részlet id nélkül", "/api/v1/ai/answer", `{"question":"kérdés","passages":[{"text":"t"}]}`},
		{"rossz nyelv", "/api/v1/ai/answer", `{"question":"kérdés","language":"magyar!","passages":[{"id":"p","text":"t"}]}`},
	}
	h := newAIAPI(walletStore(store.AIWallet{PurchasedTokens: 1000}), &fakeEmbedder{}, &fakeAnswerer{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := do(t, h, http.MethodPost, tt.path, "good-token", tt.body); rec.Code != http.StatusBadRequest {
				t.Errorf("státusz = %d, várt 400 (%s)", rec.Code, rec.Body.String())
			}
		})
	}
}

// ---------- tesztek: AI csomagok vásárlása és a webhook ----------

func newStripeAIAPI(t *testing.T, st *fakeStore, apiBase string) http.Handler {
	t.Helper()
	return New(Options{
		Store:    st,
		Verifier: fakeVerifier{valid: map[string]auth.Identity{"good-token": testIdentity}},
		Billing: billing.NewStripe(billing.StripeConfig{
			SecretKey:     "sk_test",
			WebhookSecret: testWebhookSecret,
			PriceIDs: map[string]string{
				"monthly": "price_m", "lifetime": "price_l",
				"ai_monthly": "price_aim", "ai_tokens": "price_ait",
			},
			APIBase: apiBase,
		}),
		DefaultCurrency:    "HUF",
		Log:                slog.New(slog.NewJSONHandler(io.Discard, nil)),
		CheckoutSuccessURL: "https://snitt.test/ok",
		CheckoutCancelURL:  "https://snitt.test/cancel",
	}).Handler()
}

func TestPublicPlansAIEntries(t *testing.T) {
	type plan struct {
		Key      string `json:"key"`
		Kind     string `json:"kind"`
		AITokens int64  `json:"ai_tokens"`
		Lifetime bool   `json:"lifetime"`
	}
	list := func(h http.Handler) []plan {
		var body struct {
			Plans []plan `json:"plans"`
		}
		rec := do(t, h, http.MethodGet, "/api/v1/plans", "", "")
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		return body.Plans
	}

	got := list(newStripeAIAPI(t, &fakeStore{}, "http://unused"))
	keys := []string{}
	for _, p := range got {
		keys = append(keys, p.Key+":"+p.Kind)
	}
	if strings.Join(keys, ",") != "monthly:access,ai_tokens:ai,ai_monthly:ai,lifetime:access" {
		t.Errorf("csomagok = %v", keys)
	}
	for _, p := range got {
		if (p.Key == "ai_monthly" && (p.AITokens != 1_000_000 || p.Lifetime)) || (p.Key == "ai_tokens" && (p.AITokens != 500_000 || !p.Lifetime)) {
			t.Errorf("AI csomag = %+v", p)
		}
	}

	// Ár nélkül az AI csomag nem jelenik meg.
	got = list(newStripeAPI(t, &fakeStore{}, "http://unused"))
	for _, p := range got {
		if p.Kind == billing.KindAI {
			t.Errorf("ár nélküli AI csomag a listában: %+v", p)
		}
	}
}

func TestAICheckoutRules(t *testing.T) {
	srv, last := fakeStripe(t)
	future := time.Now().Add(10 * 24 * time.Hour)
	liveAccess := map[string]store.Subscription{"s1": {ID: "s1", Subject: testIdentity.Subject, Plan: "monthly", Status: store.SubActive}}

	// Élő hozzáférési előfizetés mellé is megvehető az AI csomag.
	st := &fakeStore{admin: adminFake{subscriptions: liveAccess}}
	h := newStripeAIAPI(t, st, srv.URL)
	rec := do(t, h, http.MethodPost, "/api/v1/checkout", "good-token", `{"plan":"ai_monthly"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("ai_monthly: %d %s", rec.Code, rec.Body.String())
	}
	got := *last
	if got.Get("mode") != "subscription" || got.Get("line_items[0][price]") != "price_aim" ||
		got.Get("metadata[plan]") != "ai_monthly" || got.Get("subscription_data[metadata][plan]") != "ai_monthly" ||
		got.Get("subscription_data[metadata][subject]") != testIdentity.Subject {
		t.Errorf("ai_monthly checkout paraméterek = %v", got)
	}

	rec = do(t, h, http.MethodPost, "/api/v1/checkout", "good-token", `{"plan":"ai_tokens"}`)
	if rec.Code != http.StatusOK || (*last).Get("mode") != "payment" || (*last).Get("line_items[0][price]") != "price_ait" {
		t.Errorf("ai_tokens: %d mode=%q", rec.Code, (*last).Get("mode"))
	}

	// Már futó AI csomag mellé második nem vehető, tokencsomag igen.
	st = walletStore(store.AIWallet{Status: store.SubActive, PeriodEnd: &future, ExternalSubscriptionID: "sub_ai"})
	h = newStripeAIAPI(t, st, srv.URL)
	if rec := do(t, h, http.MethodPost, "/api/v1/checkout", "good-token", `{"plan":"ai_monthly"}`); rec.Code != http.StatusConflict {
		t.Errorf("második AI csomag: %d, várt 409", rec.Code)
	}
	if rec := do(t, h, http.MethodPost, "/api/v1/checkout", "good-token", `{"plan":"ai_tokens"}`); rec.Code != http.StatusOK {
		t.Errorf("tokencsomag futó AI csomag mellé: %d", rec.Code)
	}

	// Lemondott AI csomag után újra megvehető.
	st = walletStore(store.AIWallet{Status: store.SubCanceled, PeriodEnd: &future, ExternalSubscriptionID: "sub_ai"})
	if rec := do(t, newStripeAIAPI(t, st, srv.URL), http.MethodPost, "/api/v1/checkout", "good-token", `{"plan":"ai_monthly"}`); rec.Code != http.StatusOK {
		t.Errorf("lemondott után: %d", rec.Code)
	}

	// Ár nélkül az AI csomag nem vásárolható.
	if rec := do(t, newStripeAPI(t, &fakeStore{}, srv.URL), http.MethodPost, "/api/v1/checkout", "good-token", `{"plan":"ai_tokens"}`); rec.Code != http.StatusNotImplemented {
		t.Errorf("ár nélküli AI csomag: %d, várt 501", rec.Code)
	}
}

func TestAdminCannotGrantAIPlan(t *testing.T) {
	h := newTestAPIWith(&fakeStore{}, &fakeDirectory{})
	rec := do(t, h, http.MethodPost, "/api/v1/admin/subscriptions", "admin-token", `{"subject":"u1","plan":"ai_monthly"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("státusz = %d, várt 400", rec.Code)
	}
}

func TestWebhookAITokensPurchaseCreditsWallet(t *testing.T) {
	st := &fakeStore{}
	h := newStripeAIAPI(t, st, "http://unused")
	rec := postWebhook(t, h, `{"id":"evt_t","type":"checkout.session.completed","data":{"object":{
		"mode":"payment","payment_status":"paid","metadata":{"subject":"user-1","plan":"ai_tokens"},
		"customer":"cus_1","payment_intent":"pi_t","amount_total":199000,"currency":"huf"}}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("státusz = %d (%s)", rec.Code, rec.Body.String())
	}
	if w := st.wallets.wallets["user-1"]; w.PurchasedTokens != 500_000 || w.Status != "" {
		t.Errorf("tárca = %+v", w)
	}
	if st.admin.lastNewSub.Subject != "" || len(st.admin.subscriptions) != 0 || len(st.admin.lastSync.Grant) != 0 {
		t.Errorf("AI vásárlás nem hozhat létre előfizetést: %+v", st.admin.lastNewSub)
	}
	inv := st.admin.invoices["inv-new"]
	if inv.Status != store.InvoicePaid || inv.SubscriptionID != nil || inv.ExternalInvoiceID != "pi_t" || inv.AmountMinor != 199000 {
		t.Errorf("számla = %+v", inv)
	}
	if len(st.admin.audit) != 1 || st.admin.audit[0].Action != store.AuditAITokensCredit || st.admin.audit[0].ActorSubject != webhookActor {
		t.Errorf("napló = %+v", st.admin.audit)
	}

	// Másodszor ugyanaz az esemény: nem írunk jóvá kétszer.
	postWebhook(t, h, `{"id":"evt_t","type":"checkout.session.completed","data":{"object":{
		"mode":"payment","payment_status":"paid","metadata":{"subject":"user-1","plan":"ai_tokens"},"amount_total":199000}}}`)
	if w := st.wallets.wallets["user-1"]; w.PurchasedTokens != 500_000 {
		t.Errorf("ismételt esemény után = %+v", w)
	}
}

func TestWebhookAIMonthlyLifecycle(t *testing.T) {
	// A felhasználónak élő hozzáférési előfizetése is van: azt az AI
	// csomag eseményei nem érinthetik.
	access := store.Subscription{ID: "s1", Subject: "user-2", Plan: "monthly", Status: store.SubActive, Provider: "stripe", ExternalSubscriptionID: "sub_access"}
	st := &fakeStore{admin: adminFake{subscriptions: map[string]store.Subscription{"s1": access}}}
	h := newStripeAIAPI(t, st, "http://unused")

	rec := postWebhook(t, h, `{"id":"evt_c","type":"checkout.session.completed","data":{"object":{
		"mode":"subscription","payment_status":"paid","metadata":{"subject":"user-2","plan":"ai_monthly"},
		"customer":"cus_2","subscription":"sub_ai","invoice":"in_1","amount_total":299000,"currency":"huf"}}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("checkout: %d %s", rec.Code, rec.Body.String())
	}
	w := st.wallets.wallets["user-2"]
	wantEnd := time.Now().AddDate(0, 1, 2)
	if w.Status != store.SubActive || w.PeriodTokens != 1_000_000 || w.ExternalSubscriptionID != "sub_ai" ||
		w.PeriodEnd == nil || w.PeriodEnd.Sub(wantEnd).Abs() > time.Minute {
		t.Errorf("tárca a vásárlás után = %+v", w)
	}
	if st.admin.lastNewSub.Subject != "" || st.admin.invoices["inv-new"].SubscriptionID != nil {
		t.Error("AI csomag nem hozhat létre subscriptions sort, és a számlája sem köthető hozzá")
	}

	// A pontos időszakot a subscription.updated hozza.
	start, end := time.Now().Unix(), time.Now().AddDate(0, 1, 0).Unix()
	rec = postWebhook(t, h, `{"id":"evt_u","type":"customer.subscription.updated","data":{"object":{
		"id":"sub_ai","status":"active","metadata":{"subject":"user-2","plan":"ai_monthly"},
		"items":{"data":[{"current_period_start":`+itoa(start)+`,"current_period_end":`+itoa(end)+`}]}}}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: %d %s", rec.Code, rec.Body.String())
	}
	if w := st.wallets.wallets["user-2"]; w.PeriodEnd.Unix() != end || w.Status != store.SubActive {
		t.Errorf("tárca a frissítés után = %+v", w)
	}
	if st.admin.lastProviderUpdate.Status != "" {
		t.Error("az AI előfizetés eseménye nem módosíthatja a hozzáférési előfizetést")
	}

	// Elköltünk a keretből, majd megújul: a keret visszaáll, nem görgetődik.
	ww := st.wallets.wallets["user-2"]
	ww.PeriodTokens = 12
	st.wallets.wallets["user-2"] = ww
	nextEnd := time.Now().AddDate(0, 2, 0).Unix()
	rec = postWebhook(t, h, `{"id":"evt_i","type":"invoice.paid","data":{"object":{
		"id":"in_2","customer":"cus_2","billing_reason":"subscription_cycle","amount_paid":299000,"currency":"huf",
		"parent":{"subscription_details":{"subscription":"sub_ai","metadata":{"subject":"user-2","plan":"ai_monthly"}}},
		"lines":{"data":[{"period":{"start":`+itoa(end)+`,"end":`+itoa(nextEnd)+`}}]}}}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("invoice: %d %s", rec.Code, rec.Body.String())
	}
	w = st.wallets.wallets["user-2"]
	if w.PeriodTokens != 1_000_000 || w.PeriodEnd.Unix() != nextEnd {
		t.Errorf("tárca a megújulás után = %+v", w)
	}
	if inv := st.admin.invoices["inv-new"]; inv.ExternalInvoiceID != "in_2" || inv.Status != store.InvoicePaid || inv.Subject != "user-2" {
		t.Errorf("megújulási számla = %+v", inv)
	}

	// Megszűnés: a havi keret elvész, a vásárolt tokenek maradnak.
	ww = st.wallets.wallets["user-2"]
	ww.PurchasedTokens = 77
	st.wallets.wallets["user-2"] = ww
	rec = postWebhook(t, h, `{"id":"evt_d","type":"customer.subscription.deleted","data":{"object":{"id":"sub_ai","status":"canceled"}}}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("delete: %d %s", rec.Code, rec.Body.String())
	}
	w = st.wallets.wallets["user-2"]
	if w.Status != store.SubCanceled || w.PeriodTokens != 0 || w.PurchasedTokens != 77 || w.Total(time.Now()) != 77 {
		t.Errorf("tárca a megszűnés után = %+v", w)
	}
	if s := st.admin.subscriptions["s1"]; s.Status != store.SubActive {
		t.Errorf("a hozzáférési előfizetés nem változhat: %+v", s)
	}
}

func TestWebhookUnknownAISubscriptionAsksForRetry(t *testing.T) {
	h := newStripeAIAPI(t, &fakeStore{}, "http://unused")
	rec := postWebhook(t, h, `{"id":"evt_x","type":"customer.subscription.updated","data":{"object":{
		"id":"sub_ai_unknown","status":"active","metadata":{"plan":"ai_monthly"}}}}`)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("státusz = %d, várt 500 (a checkout esemény előtt érkezett)", rec.Code)
	}
}
