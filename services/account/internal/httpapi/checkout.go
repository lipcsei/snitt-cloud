package httpapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lipcsei/snitt-cloud/services/account/internal/auth"
	"github.com/lipcsei/snitt-cloud/services/account/internal/billing"
	"github.com/lipcsei/snitt-cloud/services/account/internal/errtrack"
	"github.com/lipcsei/snitt-cloud/services/account/internal/store"
)

// Önkiszolgáló vásárlás: a desktop app (vagy a landing) a bejelentkezett
// felhasználó tokenjével kér egy fizetési oldalt, a böngésző a szolgáltató
// (Stripe) hosztolt oldalára megy, a fizetés eredménye pedig webhookon
// érkezik vissza. Kártyaadat soha nem jár nálunk.

// webhookActor a webhookból végzett műveletek szereplője az admin naplóban.
const webhookActor = "system:billing-webhook"

// maxWebhookBody: a Stripe események jóval kisebbek ennél.
const maxWebhookBody = 1 << 20

type publicPlan struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PriceMinor  int64  `json:"price_minor"`
	Lifetime    bool   `json:"lifetime"`
}

// handlePublicPlans az önkiszolgálóan megvehető csomagok - a desktop app
// fizetési képernyője ebből mutat árat. Nyilvános: árakat bárki láthat.
func (a *API) handlePublicPlans(w http.ResponseWriter, _ *http.Request) {
	out := []publicPlan{}
	for _, p := range billing.Plans() {
		if !p.SelfService {
			continue
		}
		out = append(out, publicPlan{Key: p.Key, Name: p.Name, Description: p.Description, PriceMinor: p.PriceMinor, Lifetime: p.Lifetime})
	}
	_, checkout := a.billing.(billing.CheckoutProvider)
	writeJSON(w, http.StatusOK, map[string]any{
		"plans":            out,
		"currency":         a.currency,
		"checkout_enabled": checkout,
	})
}

type checkoutRequest struct {
	Plan string `json:"plan"`
}

func (a *API) handleCheckout(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "hiányzó identitás")
		return
	}
	cp, ok := a.billing.(billing.CheckoutProvider)
	if !ok {
		writeError(w, http.StatusNotImplemented, "az online fizetés még nincs bekapcsolva")
		return
	}
	var req checkoutRequest
	if !decodeBody(w, r, &req) {
		return
	}
	plan, ok := billing.LookupPlan(req.Plan)
	if !ok || !plan.SelfService {
		writeError(w, http.StatusBadRequest, "ez a csomag nem vásárolható meg: "+req.Plan)
		return
	}
	// Egy felhasználónak egyszerre egy élő előfizetése lehet: aki már
	// fizet, ne fizessen kétszer (a csomagváltás egyelőre az ügyfélszolgálaton megy).
	live, err := a.store.LiveSubscriptionsBySubjects(r.Context(), []string{id.Subject})
	if err != nil {
		serverError(w, r, err, http.StatusInternalServerError, "az előfizetés állapota nem elérhető")
		return
	}
	if cur, has := live[id.Subject]; has {
		writeError(w, http.StatusConflict, "már van élő előfizetésed ("+cur.Plan+")")
		return
	}
	// Az első hívás itt is létrehozza a profilt, ha a felhasználó még
	// sosem nyitotta meg a landinget - a webhook már létező profilt talál.
	if _, err := a.store.UpsertProfile(r.Context(), id.Subject, id.Email, id.PreferredUsername, id.Name); err != nil {
		serverError(w, r, err, http.StatusInternalServerError, "profil nem elérhető")
		return
	}
	url, err := cp.CreateCheckout(r.Context(), billing.CheckoutRequest{
		Subject:    id.Subject,
		Email:      id.Email,
		Plan:       plan,
		SuccessURL: a.checkoutSuccessURL,
		CancelURL:  a.checkoutCancelURL,
	})
	if err != nil {
		a.log.ErrorContext(r.Context(), "checkout indítása sikertelen", "error", err, "plan", plan.Key)
		serverError(w, r, err, http.StatusBadGateway, "a fizetési szolgáltató nem elérhető")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// handleBillingWebhook a fizetési szolgáltató eseményeit dolgozza fel. Nincs
// rajta token-ellenőrzés: a hitelességet az aláírás adja. Hibánál 5xx-et
// adunk, hogy a szolgáltató újrapróbálja; a már feldolgozott eseményt
// (újraküldés) csendben nyugtázzuk.
func (a *API) handleBillingWebhook(w http.ResponseWriter, r *http.Request) {
	wp, ok := a.billing.(billing.WebhookProvider)
	if !ok {
		writeError(w, http.StatusNotFound, "nincs webhookot fogadó fizetési szolgáltató")
		return
	}
	payload, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBody))
	if err != nil {
		writeError(w, http.StatusBadRequest, "a törzs nem olvasható")
		return
	}
	ev, err := wp.ParseWebhook(payload, r.Header.Get(wp.SignatureHeader()), time.Now())
	if err != nil {
		if errors.Is(err, billing.ErrBadSignature) {
			a.log.WarnContext(r.Context(), "webhook: érvénytelen aláírás")
			writeError(w, http.StatusBadRequest, "érvénytelen aláírás")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if ev.Kind == billing.WebhookIgnored {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}
	provider := a.billing.Name()
	seen, err := a.store.ProviderEventSeen(r.Context(), provider, ev.ID)
	if err != nil {
		serverError(w, r, err, http.StatusInternalServerError, "webhook állapot nem elérhető")
		return
	}
	if seen {
		writeJSON(w, http.StatusOK, map[string]string{"status": "duplicate"})
		return
	}

	switch ev.Kind {
	case billing.WebhookCheckoutCompleted:
		err = a.applyCheckoutCompleted(r.Context(), provider, ev)
	case billing.WebhookSubscriptionUpdated, billing.WebhookSubscriptionDeleted:
		err = a.applySubscriptionUpdate(r.Context(), provider, ev)
	case billing.WebhookInvoicePaid:
		err = a.applyInvoicePaid(r.Context(), provider, ev)
	}
	if err != nil {
		a.log.ErrorContext(r.Context(), "webhook feldolgozása sikertelen", "error", err, "event", ev.ID, "type", ev.Type)
		errtrack.Report(r, err, http.StatusInternalServerError)
		writeError(w, http.StatusInternalServerError, "a webhook feldolgozása nem sikerült")
		return
	}
	if err := a.store.RecordProviderEvent(r.Context(), provider, ev.ID, ev.Type); err != nil {
		// A feldolgozás megtörtént; egy újraküldést az ütközés-ellenőrzés kezel.
		a.log.WarnContext(r.Context(), "webhook esemény rögzítése sikertelen", "error", err, "event", ev.ID)
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// applyCheckoutCompleted a sikeres fizetés után kiadja az előfizetést (vagy
// a lifetime hozzáférést) és rögzíti a kifizetett számlát.
func (a *API) applyCheckoutCompleted(ctx context.Context, provider string, ev billing.WebhookEvent) error {
	plan, ok := billing.LookupPlan(ev.PlanKey)
	if !ok || ev.Subject == "" {
		return fmt.Errorf("a checkout eseményből hiányzik a csomag vagy a felhasználó (plan=%q)", ev.PlanKey)
	}
	if ev.SubscriptionID != "" {
		// Újraküldött esemény, aminek a rögzítése korábban elveszett.
		if _, err := a.store.SubscriptionByExternalID(ctx, provider, ev.SubscriptionID); err == nil {
			return nil
		}
	}
	now := time.Now().UTC()
	// Havi előfizetésnél a pontos időszakot a customer.subscription.updated
	// esemény hozza; addig egy hónapot adunk, hogy a hozzáférés azonnal éljen.
	end := plan.PeriodEnd(now, 31)
	currency := ev.Currency
	if currency == "" {
		currency = a.currency
	}
	amount := ev.AmountMinor
	sub, err := a.store.GrantSubscription(ctx, store.NewSubscription{
		Subject:                ev.Subject,
		Plan:                   plan.Key,
		PriceMinor:             amount,
		Currency:               currency,
		PeriodStart:            now,
		PeriodEnd:              end,
		Provider:               provider,
		ExternalCustomerID:     ev.CustomerID,
		ExternalSubscriptionID: ev.SubscriptionID,
	}, store.EntitlementSync{
		Managed:   billing.ManagedFeatureKeys(),
		Grant:     plan.FeatureKeys,
		ExpiresAt: plan.EntitlementExpiry(end),
	})
	if err != nil {
		return fmt.Errorf("előfizetés kiadása fizetés után: %w", err)
	}
	a.systemAudit(ctx, store.NewAuditEntry{
		Action:     store.AuditSubscriptionGrant,
		TargetType: "subscription",
		TargetID:   sub.ID,
		Subject:    sub.Subject,
		Summary:    fmt.Sprintf("Online vásárlás: %s csomag, %s.", plan.Name, moneyText(amount, currency)),
		Detail:     map[string]any{"plan": plan.Key, "event": ev.ID, "provider": provider, "external_subscription_id": ev.SubscriptionID},
	})
	return a.recordPaidInvoice(ctx, provider, sub, amount, currency, ev.PaymentID, "Online fizetés: "+plan.Name+".")
}

// applySubscriptionUpdate a szolgáltató szerinti állapotra hozza az
// előfizetést (megújulás, lemondás, sikertelen terhelés).
func (a *API) applySubscriptionUpdate(ctx context.Context, provider string, ev billing.WebhookEvent) error {
	sub, err := a.store.SubscriptionByExternalID(ctx, provider, ev.SubscriptionID)
	if errors.Is(err, store.ErrNotFound) {
		// A Stripe nem garantál sorrendet: ez megelőzhette a checkout
		// eseményt. Hibát adunk, hogy később újraküldje.
		return fmt.Errorf("ismeretlen előfizetés: %s", ev.SubscriptionID)
	}
	if err != nil {
		return err
	}
	plan, ok := billing.LookupPlan(sub.Plan)
	if !ok {
		return fmt.Errorf("a rögzített csomag (%s) már nem létezik", sub.Plan)
	}
	upd := store.ProviderUpdate{
		Status:            ev.Status,
		PeriodStart:       ev.PeriodStart,
		PeriodEnd:         ev.PeriodEnd,
		CancelAtPeriodEnd: ev.CancelAtPeriodEnd,
	}
	if ev.Kind == billing.WebhookSubscriptionDeleted {
		upd.Status = store.SubCanceled
	}
	if upd.PeriodStart.IsZero() || !upd.PeriodEnd.After(upd.PeriodStart) {
		upd.PeriodStart, upd.PeriodEnd = sub.CurrentPeriodStart, sub.CurrentPeriodEnd
	}
	sync := store.EntitlementSync{Managed: billing.ManagedFeatureKeys()}
	if upd.Status == store.SubActive || upd.Status == store.SubPastDue {
		sync.Grant, sync.ExpiresAt = plan.FeatureKeys, plan.EntitlementExpiry(upd.PeriodEnd)
	}
	sub, err = a.store.ApplyProviderUpdate(ctx, sub.ID, upd, sync)
	if err != nil {
		return err
	}
	if ev.Kind == billing.WebhookSubscriptionDeleted {
		a.systemAudit(ctx, store.NewAuditEntry{
			Action:     store.AuditSubscriptionCancel,
			TargetType: "subscription",
			TargetID:   sub.ID,
			Subject:    sub.Subject,
			Summary:    fmt.Sprintf("Az előfizetés megszűnt a fizetési szolgáltatónál: %s csomag.", sub.Plan),
			Detail:     map[string]any{"plan": sub.Plan, "event": ev.ID, "provider": provider},
		})
	}
	return nil
}

// applyInvoicePaid a megújulási terhelést veszi fel a saját számlakönyvbe.
func (a *API) applyInvoicePaid(ctx context.Context, provider string, ev billing.WebhookEvent) error {
	sub, err := a.store.SubscriptionByExternalID(ctx, provider, ev.SubscriptionID)
	if err != nil {
		return fmt.Errorf("megújulási számla előfizetése (%s): %w", ev.SubscriptionID, err)
	}
	currency := ev.Currency
	if currency == "" {
		currency = sub.Currency
	}
	return a.recordPaidInvoice(ctx, provider, sub, ev.AmountMinor, currency, ev.PaymentID, "Megújulás: "+sub.Plan+".")
}

func (a *API) recordPaidInvoice(ctx context.Context, provider string, sub store.Subscription, amount int64, currency, paymentID, note string) error {
	if amount <= 0 {
		return nil
	}
	inv, err := a.store.CreateInvoice(ctx, store.NewInvoice{
		Subject:        sub.Subject,
		SubscriptionID: &sub.ID,
		AmountMinor:    amount,
		Currency:       strings.ToUpper(currency),
		Provider:       provider,
		Note:           note,
	})
	if err != nil {
		return err
	}
	if _, err := a.store.MarkInvoicePaid(ctx, inv.ID, time.Now().UTC(), paymentID); err != nil {
		return err
	}
	return nil
}

// systemAudit a nem admin által, hanem a fizetési szolgáltató eseményéből
// végzett műveletet naplózza - a pénzt érintő változásnak itt is nyoma marad.
func (a *API) systemAudit(ctx context.Context, e store.NewAuditEntry) {
	e.ActorSubject = webhookActor
	e.ActorLabel = "Fizetési szolgáltató (" + a.billing.Name() + ")"
	if _, err := a.store.RecordAudit(ctx, e); err != nil {
		a.log.ErrorContext(ctx, "napló írása sikertelen", "error", err, "action", e.Action, "target_id", e.TargetID)
	}
}
