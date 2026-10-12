package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// StripeConfig a Stripe illesztő beállításai. A PriceIDs a csomagkulcsot a
// Stripe-ban létrehozott árhoz (price_...) köti: az összeget és az
// ismétlődést (havi / egyszeri) a Stripe oldali ár határozza meg, a
// Plan.PriceMinor csak a saját nyilvántartásunk és a felület számára van.
type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
	PriceIDs      map[string]string
	// ManagedPayments: a Stripe a kereskedő (merchant of record) - ő számolja
	// és fizeti be az áfát országonként, ő állítja ki a számlát/nyugtát, és ő
	// kezeli a visszatérítési és vitás ügyeket. Ilyenkor saját számlát nem
	// kérünk a Checkouttól (a Stripe nem is engedi).
	ManagedPayments bool
	// APIBase a tesztekhez cserélhető; üresen a valódi Stripe API.
	APIBase string
}

// StripeAPIVersion a hívásainkban küldött Stripe-Version. A webhook végpontot
// a Stripe Dashboardon ugyanerre a verzióra kell állítani.
const StripeAPIVersion = "2025-03-31.basil"

// Stripe a Stripe Checkout + webhook alapú fizetési szolgáltató. Szándékosan
// nem a hivatalos SDK-t használja: négy végpontot hívunk, ehhez egy nagy
// függőség több kockázat, mint amennyi munkát megspórol.
type Stripe struct {
	cfg    StripeConfig
	client *http.Client
}

// NewStripe összeállítja az illesztőt.
func NewStripe(cfg StripeConfig) *Stripe {
	if cfg.APIBase == "" {
		cfg.APIBase = "https://api.stripe.com"
	}
	return &Stripe{cfg: cfg, client: &http.Client{Timeout: 20 * time.Second}}
}

// Name a "stripe" azonosító.
func (*Stripe) Name() string { return "stripe" }

// EnsureCustomer üres azonosítót ad: az ügyfelet a Checkout hozza létre a
// fizetéskor, az azonosítója a webhookkal érkezik.
func (*Stripe) EnsureCustomer(context.Context, string, string) (string, error) { return "", nil }

// StartSubscription üres azonosítót ad: Stripe-on az előfizetés a Checkouton
// indul (a felhasználó adja meg a kártyáját), nem szerver oldalról. Az admin
// által kézzel kiadott előfizetés így továbbra is kézi sor marad.
func (*Stripe) StartSubscription(context.Context, Charge) (string, error) { return "", nil }

// CancelSubscription lemondja a Stripe oldali előfizetést. Kézi sornál (üres
// külső azonosító) nincs mit lemondani.
func (s *Stripe) CancelSubscription(ctx context.Context, externalID string, atPeriodEnd bool) error {
	if externalID == "" {
		return nil
	}
	path := "/v1/subscriptions/" + url.PathEscape(externalID)
	if atPeriodEnd {
		_, err := s.call(ctx, http.MethodPost, path, url.Values{"cancel_at_period_end": {"true"}})
		return err
	}
	_, err := s.call(ctx, http.MethodDelete, path, nil)
	return err
}

// CreateCheckout Stripe Checkout munkamenetet nyit, és a fizetési oldal
// címét adja vissza. A subject a metadata-ban és a client_reference_id-ben
// utazik: a webhook ebből tudja, kinek kell a jogosultságot kiadni.
func (s *Stripe) CreateCheckout(ctx context.Context, req CheckoutRequest) (string, error) {
	price := s.cfg.PriceIDs[req.Plan.Key]
	if price == "" {
		return "", fmt.Errorf("a(z) %s csomaghoz nincs Stripe ár beállítva", req.Plan.Key)
	}
	form := url.Values{
		"line_items[0][price]":    {price},
		"line_items[0][quantity]": {"1"},
		"success_url":             {req.SuccessURL},
		"cancel_url":              {req.CancelURL},
		"client_reference_id":     {req.Subject},
		"metadata[subject]":       {req.Subject},
		"metadata[plan]":          {req.Plan.Key},
		"locale":                  {"auto"},
	}
	if req.Email != "" {
		form.Set("customer_email", req.Email)
	}
	if req.Plan.Lifetime {
		form.Set("mode", "payment")
		// Egyszeri fizetésnél is legyen Stripe ügyfél: a számla és a
		// későbbi ügyfélszolgálati keresés ehhez kötődik.
		form.Set("customer_creation", "always")
		if !s.cfg.ManagedPayments {
			form.Set("invoice_creation[enabled]", "true")
		}
	} else {
		form.Set("mode", "subscription")
		// Az előfizetés eseményei (megújulás, lemondás) ebből tudják, kié.
		form.Set("subscription_data[metadata][subject]", req.Subject)
		form.Set("subscription_data[metadata][plan]", req.Plan.Key)
	}
	if s.cfg.ManagedPayments {
		form.Set("managed_payments[enabled]", "true")
	}
	body, err := s.call(ctx, http.MethodPost, "/v1/checkout/sessions", form)
	if err != nil {
		return "", err
	}
	var out struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(body, &out); err != nil || out.URL == "" {
		return "", fmt.Errorf("a Stripe nem adott fizetési címet")
	}
	return out.URL, nil
}

func (s *Stripe) call(ctx context.Context, method, path string, form url.Values) ([]byte, error) {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, s.cfg.APIBase+path, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.cfg.SecretKey, "")
	// Rögzített API-verzió: a fiók alapértelmezése ne változtassa meg alattunk
	// a válaszok alakját. A Managed Payments ennél régebbin nem érhető el.
	req.Header.Set("Stripe-Version", StripeAPIVersion)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("stripe hívás: %w", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("stripe válasz olvasása: %w", err)
	}
	if resp.StatusCode >= 300 {
		var e struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(data, &e)
		return nil, fmt.Errorf("stripe %d: %s", resp.StatusCode, e.Error.Message)
	}
	return data, nil
}

// ---------- webhook ----------

// WebhookKind a szolgáltató eseményének szolgáltató-független fajtája.
type WebhookKind string

// A feldolgozott eseményfajták. Minden más WebhookIgnored.
const (
	WebhookIgnored             WebhookKind = ""
	WebhookCheckoutCompleted   WebhookKind = "checkout_completed"
	WebhookSubscriptionUpdated WebhookKind = "subscription_updated"
	WebhookSubscriptionDeleted WebhookKind = "subscription_deleted"
	WebhookInvoicePaid         WebhookKind = "invoice_paid"
)

// WebhookEvent egy ellenőrzött, szolgáltató-független formára hozott
// webhook esemény. A HTTP réteg csak ezt látja, a Stripe JSON-t nem.
type WebhookEvent struct {
	ID   string
	Type string
	Kind WebhookKind

	Subject        string
	PlanKey        string
	CustomerID     string
	SubscriptionID string
	// PaymentID a kifizetés azonosítója (invoice vagy payment_intent) - a
	// saját számla external_invoice_id mezőjébe kerül.
	PaymentID   string
	AmountMinor int64
	Currency    string

	// Előfizetés-eseményeknél: a store státuszkészletére leképezve.
	Status            string
	PeriodStart       time.Time
	PeriodEnd         time.Time
	CancelAtPeriodEnd bool
}

// WebhookProvider a webhookot fogadni tudó szolgáltató.
type WebhookProvider interface {
	CheckoutProvider
	// ParseWebhook ellenőrzi az aláírást, és értelmezi az eseményt.
	ParseWebhook(payload []byte, signatureHeader string, now time.Time) (WebhookEvent, error)
	// SignatureHeader az aláírást hordozó HTTP fejléc neve.
	SignatureHeader() string
}

// ErrBadSignature: a webhook aláírása hiányzik, hibás vagy túl régi.
var ErrBadSignature = errors.New("érvénytelen webhook aláírás")

// webhookTolerance: ennél régebbi időbélyegű eseményt nem fogadunk el
// (visszajátszás elleni védelem, a Stripe ajánlása szerint 5 perc).
const webhookTolerance = 5 * time.Minute

// SignatureHeader a "Stripe-Signature" fejléc.
func (*Stripe) SignatureHeader() string { return "Stripe-Signature" }

// ParseWebhook a Stripe aláírást ellenőrzi (HMAC-SHA256 a "t.payload"
// szövegen), majd a minket érdeklő eseményeket értelmezi.
func (s *Stripe) ParseWebhook(payload []byte, header string, now time.Time) (WebhookEvent, error) {
	if err := verifyStripeSignature(payload, header, s.cfg.WebhookSecret, now); err != nil {
		return WebhookEvent{}, err
	}
	var env struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			Object json.RawMessage `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &env); err != nil {
		return WebhookEvent{}, fmt.Errorf("webhook értelmezése: %w", err)
	}
	ev := WebhookEvent{ID: env.ID, Type: env.Type}

	switch env.Type {
	case "checkout.session.completed":
		var o struct {
			Mode              string            `json:"mode"`
			PaymentStatus     string            `json:"payment_status"`
			ClientReferenceID string            `json:"client_reference_id"`
			Metadata          map[string]string `json:"metadata"`
			Customer          string            `json:"customer"`
			Subscription      string            `json:"subscription"`
			Invoice           string            `json:"invoice"`
			PaymentIntent     string            `json:"payment_intent"`
			AmountTotal       int64             `json:"amount_total"`
			Currency          string            `json:"currency"`
		}
		if err := json.Unmarshal(env.Data.Object, &o); err != nil {
			return WebhookEvent{}, fmt.Errorf("checkout esemény: %w", err)
		}
		// Késleltetett fizetési módoknál (pl. átutalás) a munkamenet
		// "kifizetetlenül" is lezárulhat - ilyenkor még nem jár semmi.
		if o.PaymentStatus != "paid" && o.PaymentStatus != "no_payment_required" {
			return ev, nil
		}
		ev.Kind = WebhookCheckoutCompleted
		ev.Subject = firstNonEmpty(o.Metadata["subject"], o.ClientReferenceID)
		ev.PlanKey = o.Metadata["plan"]
		ev.CustomerID = o.Customer
		ev.SubscriptionID = o.Subscription
		ev.PaymentID = firstNonEmpty(o.Invoice, o.PaymentIntent)
		ev.AmountMinor = o.AmountTotal
		ev.Currency = strings.ToUpper(o.Currency)

	case "customer.subscription.updated", "customer.subscription.deleted":
		var o struct {
			ID                 string            `json:"id"`
			Customer           string            `json:"customer"`
			Status             string            `json:"status"`
			CancelAtPeriodEnd  bool              `json:"cancel_at_period_end"`
			CurrentPeriodStart int64             `json:"current_period_start"`
			CurrentPeriodEnd   int64             `json:"current_period_end"`
			Metadata           map[string]string `json:"metadata"`
			Items              struct {
				Data []struct {
					CurrentPeriodStart int64 `json:"current_period_start"`
					CurrentPeriodEnd   int64 `json:"current_period_end"`
				} `json:"data"`
			} `json:"items"`
		}
		if err := json.Unmarshal(env.Data.Object, &o); err != nil {
			return WebhookEvent{}, fmt.Errorf("előfizetés esemény: %w", err)
		}
		ev.Kind = WebhookSubscriptionUpdated
		if env.Type == "customer.subscription.deleted" {
			ev.Kind = WebhookSubscriptionDeleted
		}
		ev.SubscriptionID = o.ID
		ev.CustomerID = o.Customer
		ev.Subject = o.Metadata["subject"]
		ev.PlanKey = o.Metadata["plan"]
		ev.Status = stripeStatus(o.Status)
		ev.CancelAtPeriodEnd = o.CancelAtPeriodEnd
		// Az újabb API-verziókban az időszak a tételeken van, nem az
		// előfizetésen - mindkettőt elfogadjuk.
		start, end := o.CurrentPeriodStart, o.CurrentPeriodEnd
		if (start == 0 || end == 0) && len(o.Items.Data) > 0 {
			start, end = o.Items.Data[0].CurrentPeriodStart, o.Items.Data[0].CurrentPeriodEnd
		}
		if start > 0 && end > 0 {
			ev.PeriodStart, ev.PeriodEnd = time.Unix(start, 0).UTC(), time.Unix(end, 0).UTC()
		}

	case "invoice.paid":
		var o struct {
			ID            string `json:"id"`
			Customer      string `json:"customer"`
			Subscription  string `json:"subscription"`
			BillingReason string `json:"billing_reason"`
			AmountPaid    int64  `json:"amount_paid"`
			Currency      string `json:"currency"`
			Parent        struct {
				SubscriptionDetails struct {
					Subscription string `json:"subscription"`
				} `json:"subscription_details"`
			} `json:"parent"`
		}
		if err := json.Unmarshal(env.Data.Object, &o); err != nil {
			return WebhookEvent{}, fmt.Errorf("számla esemény: %w", err)
		}
		// Az első számlát a checkout esemény már rögzíti; itt csak a
		// megújulásokat vesszük fel a saját számlakönyvbe.
		if o.BillingReason != "subscription_cycle" {
			return ev, nil
		}
		ev.Kind = WebhookInvoicePaid
		ev.SubscriptionID = firstNonEmpty(o.Subscription, o.Parent.SubscriptionDetails.Subscription)
		ev.CustomerID = o.Customer
		ev.PaymentID = o.ID
		ev.AmountMinor = o.AmountPaid
		ev.Currency = strings.ToUpper(o.Currency)
	}
	return ev, nil
}

// stripeStatus a Stripe előfizetés-státuszát a saját négy állapotunkra képezi.
func stripeStatus(s string) string {
	switch s {
	case "active", "trialing":
		return "active"
	case "past_due", "unpaid", "incomplete":
		return "past_due"
	case "canceled", "incomplete_expired":
		return "canceled"
	default:
		return "expired"
	}
}

func verifyStripeSignature(payload []byte, header, secret string, now time.Time) error {
	if secret == "" || header == "" {
		return ErrBadSignature
	}
	var ts int64
	var sigs []string
	for _, part := range strings.Split(header, ",") {
		k, v, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch k {
		case "t":
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return ErrBadSignature
			}
			ts = n
		case "v1":
			sigs = append(sigs, v)
		}
	}
	if ts == 0 || len(sigs) == 0 {
		return ErrBadSignature
	}
	if d := now.Sub(time.Unix(ts, 0)); d > webhookTolerance || d < -webhookTolerance {
		return ErrBadSignature
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10)))
	mac.Write([]byte("."))
	mac.Write(payload)
	want := mac.Sum(nil)
	for _, s := range sigs {
		got, err := hex.DecodeString(s)
		if err == nil && hmac.Equal(got, want) {
			return nil
		}
	}
	return ErrBadSignature
}

// SignStripePayload a tesztekhez: a Stripe-pal azonos módon aláírt fejlécet ad.
func SignStripePayload(payload []byte, secret string, ts time.Time) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts.Unix(), 10)))
	mac.Write([]byte("."))
	mac.Write(payload)
	return fmt.Sprintf("t=%d,v1=%s", ts.Unix(), hex.EncodeToString(mac.Sum(nil)))
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
