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
	// Kind: "access" (hozzáférést adó előfizetés) vagy "ai" (AI tokencsomag).
	Kind string `json:"kind"`
	// AITokens AI csomagnál a csomag tokenkerete (havi csomagnál
	// periódusonként), egyébként 0.
	AITokens int64 `json:"ai_tokens"`
}

// handlePublicPlans az önkiszolgálóan megvehető csomagok - a desktop app
// fizetési képernyője ebből mutat árat. Nyilvános: árakat bárki láthat. Az
// AI csomagok csak akkor szerepelnek, ha van hozzájuk beállított ár.
func (a *API) handlePublicPlans(w http.ResponseWriter, _ *http.Request) {
	out := []publicPlan{}
	for _, p := range billing.AllPlans() {
		if !p.SelfService || (p.IsAI() && !a.offers(p)) {
			continue
		}
		kind := "access"
		if p.IsAI() {
			kind = billing.KindAI
		}
		out = append(out, publicPlan{
			Key: p.Key, Name: p.Name, Description: p.Description, PriceMinor: p.PriceMinor,
			Lifetime: p.Lifetime, Kind: kind, AITokens: p.AITokens,
		})
	}
	_, checkout := a.billing.(billing.CheckoutProvider)
	writeJSON(w, http.StatusOK, map[string]any{
		"plans":            out,
		"currency":         a.currency,
		"checkout_enabled": checkout,
	})
}

// offers igaz, ha a fizetési szolgáltató a csomagot ténylegesen árulja. A
// csomagonkénti elérhetőséget nem ismerő szolgáltatónál (Manual) hamis.
func (a *API) offers(p billing.Plan) bool {
	o, ok := a.billing.(billing.PlanOfferer)
	return ok && o.Offers(p.Key)
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
	plan, ok := billing.LookupAnyPlan(req.Plan)
	if !ok || !plan.SelfService {
		writeError(w, http.StatusBadRequest, "ez a csomag nem vásárolható meg: "+req.Plan)
		return
	}
	if plan.IsAI() {
		if !a.checkAICheckout(w, r, id.Subject, plan) {
			return
		}
	} else {
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

// checkAICheckout az AI csomag vásárlásának feltételei. Az AI csomag nem
// számít bele az "egy élő előfizetés" szabályba (a hozzáférési csomag mellé
// is megvehető), a havi AI csomagból viszont egyszerre csak egy futhat. A
// tokencsomag bármikor, akárhányszor megvehető.
func (a *API) checkAICheckout(w http.ResponseWriter, r *http.Request, subject string, plan billing.Plan) bool {
	if !a.offers(plan) {
		writeError(w, http.StatusNotImplemented, "ez az AI csomag még nem vásárolható meg")
		return false
	}
	if plan.Lifetime {
		return true
	}
	wallet, err := a.store.AIWallet(r.Context(), subject)
	if err != nil {
		serverError(w, r, err, http.StatusInternalServerError, "az AI egyenleg nem elérhető")
		return false
	}
	if aiSubscriptionLive(wallet, time.Now()) {
		writeError(w, http.StatusConflict, "már van aktív AI csomagod")
		return false
	}
	return true
}

// aiSubscriptionLive igaz, ha a tárca havi AI előfizetése még él a
// szolgáltatónál (aktív vagy késedelmes, és a periódus nem járt le).
func aiSubscriptionLive(w store.AIWallet, now time.Time) bool {
	if w.Status != store.SubActive && w.Status != store.SubPastDue {
		return false
	}
	return w.PeriodEnd != nil && w.PeriodEnd.After(now)
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
	if p, ok := billing.LookupAnyPlan(ev.PlanKey); ok && p.IsAI() {
		return a.applyAICheckout(ctx, provider, p, ev)
	}
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
	if handled, err := a.applyAISubscriptionUpdate(ctx, ev); handled {
		return err
	}
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
	if handled, err := a.applyAIInvoicePaid(ctx, provider, ev); handled {
		return err
	}
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
	return a.recordPaidInvoiceFor(ctx, provider, sub.Subject, &sub.ID, amount, currency, paymentID, note)
}

// recordPaidInvoiceFor kifizetett számlát rögzít; a subscriptionID nil az
// AI csomagoknál, amiknek nincs subscriptions soruk.
func (a *API) recordPaidInvoiceFor(ctx context.Context, provider, subject string, subscriptionID *string, amount int64, currency, paymentID, note string) error {
	if amount <= 0 {
		return nil
	}
	inv, err := a.store.CreateInvoice(ctx, store.NewInvoice{
		Subject:        subject,
		SubscriptionID: subscriptionID,
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

// ---------- AI csomagok a webhookban ----------

// aiProvisionalPeriodEnd a havi AI csomag ideiglenes periódusának vége a checkout után:
// a pontos időszakot a customer.subscription.updated esemény hozza, addig a
// keret azonnal éljen. A két nap ráhagyás a megújulási terhelés csúszására.
func aiProvisionalPeriodEnd(now time.Time) time.Time {
	return now.AddDate(0, 1, 2)
}

// applyAICheckout a sikeres AI vásárlás: tokencsomagnál jóváírja a
// tokeneket, havi AI csomagnál elindítja a tárca előfizetését. A
// subscriptions táblát és a jogosultságokat nem érinti.
func (a *API) applyAICheckout(ctx context.Context, provider string, plan billing.Plan, ev billing.WebhookEvent) error {
	if ev.Subject == "" {
		return fmt.Errorf("az AI checkout eseményből hiányzik a felhasználó (plan=%q)", plan.Key)
	}
	currency := ev.Currency
	if currency == "" {
		currency = a.currency
	}
	if plan.Lifetime {
		w, err := a.store.CreditAITokens(ctx, ev.Subject, plan.AITokens)
		if err != nil {
			return fmt.Errorf("AI tokenek jóváírása fizetés után: %w", err)
		}
		a.systemAudit(ctx, store.NewAuditEntry{
			Action:     store.AuditAITokensCredit,
			TargetType: "ai_wallet",
			TargetID:   ev.Subject,
			Subject:    ev.Subject,
			Summary:    fmt.Sprintf("Online vásárlás: %s, %d AI token jóváírva, %s.", plan.Name, plan.AITokens, moneyText(ev.AmountMinor, currency)),
			Detail:     map[string]any{"plan": plan.Key, "event": ev.ID, "provider": provider, "tokens": plan.AITokens, "purchased_tokens": w.PurchasedTokens},
		})
		return a.recordPaidInvoiceFor(ctx, provider, ev.Subject, nil, ev.AmountMinor, currency, ev.PaymentID, "Online fizetés: "+plan.Name+".")
	}

	if ev.SubscriptionID == "" {
		return fmt.Errorf("a havi AI csomag checkout eseményéből hiányzik az előfizetés")
	}
	// Újraküldött esemény, aminek a rögzítése korábban elveszett.
	if w, err := a.store.AIWalletByExternalSubscription(ctx, ev.SubscriptionID); err == nil && w.Subject == ev.Subject {
		return nil
	}
	end := aiProvisionalPeriodEnd(time.Now().UTC())
	if _, err := a.store.SetAISubscription(ctx, ev.Subject, store.AISubscription{
		ExternalSubscriptionID: ev.SubscriptionID,
		Status:                 store.SubActive,
		PeriodTokens:           plan.AITokens,
		PeriodEnd:              end,
	}); err != nil {
		return fmt.Errorf("AI előfizetés indítása fizetés után: %w", err)
	}
	a.systemAudit(ctx, store.NewAuditEntry{
		Action:     store.AuditAISubscription,
		TargetType: "ai_wallet",
		TargetID:   ev.Subject,
		Subject:    ev.Subject,
		Summary:    fmt.Sprintf("Online vásárlás: %s, havi %d AI token, %s.", plan.Name, plan.AITokens, moneyText(ev.AmountMinor, currency)),
		Detail:     map[string]any{"plan": plan.Key, "event": ev.ID, "provider": provider, "external_subscription_id": ev.SubscriptionID, "period_end": end},
	})
	return a.recordPaidInvoiceFor(ctx, provider, ev.Subject, nil, ev.AmountMinor, currency, ev.PaymentID, "Online fizetés: "+plan.Name+".")
}

// applyAISubscriptionUpdate a havi AI csomag Stripe előfizetésének
// változását vezeti át a tárcán. A handled hamis, ha az azonosító nem AI
// tárcáé - ilyenkor a hívó a subscriptions táblában keres tovább.
func (a *API) applyAISubscriptionUpdate(ctx context.Context, ev billing.WebhookEvent) (handled bool, err error) {
	w, err := a.store.AIWalletByExternalSubscription(ctx, ev.SubscriptionID)
	if errors.Is(err, store.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	upd := store.AISubscriptionUpdate{Status: ev.Status}
	if !ev.PeriodEnd.IsZero() {
		end := ev.PeriodEnd
		upd.PeriodEnd = &end
	}
	if ev.Kind == billing.WebhookSubscriptionDeleted {
		// Megszűnéskor a maradék havi keret elvész; a vásárolt tokenek maradnak.
		zero := int64(0)
		upd.Status, upd.PeriodTokens = store.SubCanceled, &zero
	}
	if _, err := a.store.UpdateAISubscription(ctx, ev.SubscriptionID, upd); err != nil {
		return true, fmt.Errorf("AI előfizetés frissítése: %w", err)
	}
	if ev.Kind == billing.WebhookSubscriptionDeleted {
		a.systemAudit(ctx, store.NewAuditEntry{
			Action:     store.AuditAISubscription,
			TargetType: "ai_wallet",
			TargetID:   w.Subject,
			Subject:    w.Subject,
			Summary:    "Az AI csomag megszűnt a fizetési szolgáltatónál.",
			Detail:     map[string]any{"event": ev.ID, "external_subscription_id": ev.SubscriptionID},
		})
	}
	return true, nil
}

// applyAIInvoicePaid a havi AI csomag megújulása: a keret visszaáll a
// csomag teljes keretére (a maradék nem görgetődik át), és a terhelés
// bekerül a számlakönyvbe.
func (a *API) applyAIInvoicePaid(ctx context.Context, provider string, ev billing.WebhookEvent) (handled bool, err error) {
	w, err := a.store.AIWalletByExternalSubscription(ctx, ev.SubscriptionID)
	if errors.Is(err, store.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return true, err
	}
	plan, ok := billing.LookupAnyPlan(ev.PlanKey)
	if !ok || !plan.IsAI() || plan.Lifetime {
		// Régebbi előfizetés metadata nélkül: csak egy havi AI csomag van.
		plan, _ = billing.LookupAnyPlan("ai_monthly")
	}
	tokens := plan.AITokens
	upd := store.AISubscriptionUpdate{Status: store.SubActive, PeriodTokens: &tokens}
	if !ev.PeriodEnd.IsZero() {
		end := ev.PeriodEnd
		upd.PeriodEnd = &end
	}
	if _, err := a.store.UpdateAISubscription(ctx, ev.SubscriptionID, upd); err != nil {
		return true, fmt.Errorf("AI keret megújítása: %w", err)
	}
	currency := ev.Currency
	if currency == "" {
		currency = a.currency
	}
	return true, a.recordPaidInvoiceFor(ctx, provider, w.Subject, nil, ev.AmountMinor, currency, ev.PaymentID, "Megújulás: "+plan.Name+".")
}
