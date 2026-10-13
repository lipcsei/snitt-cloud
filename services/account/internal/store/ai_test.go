package store

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestAIWalletDebitOrder(t *testing.T) {
	now := time.Now()
	future, past := now.Add(24*time.Hour), now.Add(-time.Hour)

	tests := []struct {
		name                string
		w                   AIWallet
		amount              int64
		wantPeriod, wantBuy int64
		wantTotalBefore     int64
	}{
		{"előbb a havi keret", AIWallet{Status: SubActive, PeriodEnd: &future, PeriodTokens: 100, PurchasedTokens: 50}, 30, 70, 50, 150},
		{"a maradék a vásároltból", AIWallet{Status: SubActive, PeriodEnd: &future, PeriodTokens: 100, PurchasedTokens: 50}, 120, 0, 30, 150},
		{"nulla alá nem megy", AIWallet{Status: SubActive, PeriodEnd: &future, PeriodTokens: 10, PurchasedTokens: 5}, 100, 0, 0, 15},
		{"lejárt keret nem költhető", AIWallet{Status: SubActive, PeriodEnd: &past, PeriodTokens: 100, PurchasedTokens: 50}, 30, 100, 20, 50},
		{"lemondott keret nem költhető", AIWallet{Status: SubCanceled, PeriodEnd: &future, PeriodTokens: 100, PurchasedTokens: 50}, 30, 100, 20, 50},
		{"késedelmes keret nem költhető", AIWallet{Status: SubPastDue, PeriodEnd: &future, PeriodTokens: 100}, 30, 100, 0, 0},
		{"üres tárca", AIWallet{}, 10, 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.w.Total(now); got != tt.wantTotalBefore {
				t.Errorf("egyenleg = %d, várt %d", got, tt.wantTotalBefore)
			}
			got := tt.w.Debit(now, tt.amount)
			if got.PeriodTokens != tt.wantPeriod || got.PurchasedTokens != tt.wantBuy {
				t.Errorf("levonás után keret=%d vásárolt=%d, várt %d / %d",
					got.PeriodTokens, got.PurchasedTokens, tt.wantPeriod, tt.wantBuy)
			}
		})
	}
}

// Az AI tárca SQL-je (sorzár, upsert, COALESCE-es frissítés) csak valódi
// Postgresen ellenőrizhető; TEST_DATABASE_URL nélkül kimarad.
func TestAIWalletRoundTrip(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL nincs beállítva")
	}
	ctx := context.Background()
	st, err := New(ctx, url)
	if err != nil {
		t.Fatalf("kapcsolódás: %v", err)
	}
	t.Cleanup(st.Close)
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("séma: %v", err)
	}
	subject := "test-subject-" + newID()
	ext := "sub_test_" + newID()
	t.Cleanup(func() {
		_, _ = st.pool.Exec(ctx, `DELETE FROM ai_usage WHERE subject = $1`, subject)
		_, _ = st.pool.Exec(ctx, `DELETE FROM ai_wallets WHERE subject = $1`, subject)
	})

	w, err := st.AIWallet(ctx, subject)
	if err != nil || w.Total(time.Now()) != 0 {
		t.Fatalf("üres tárca = %+v, %v", w, err)
	}
	if _, err := st.CreditAITokens(ctx, subject, 500); err != nil {
		t.Fatal(err)
	}
	end := time.Now().Add(30 * 24 * time.Hour)
	if _, err := st.SetAISubscription(ctx, subject, AISubscription{
		ExternalSubscriptionID: ext, Status: SubActive, PeriodTokens: 1000, PeriodEnd: end,
	}); err != nil {
		t.Fatal(err)
	}
	w, err = st.DebitAITokens(ctx, subject, NewAIUsage{Kind: AIUsageAnswer, Model: "m", InputTokens: 1, OutputTokens: 1, ChargedTokens: 1200})
	if err != nil {
		t.Fatal(err)
	}
	if w.PeriodTokens != 0 || w.PurchasedTokens != 300 {
		t.Errorf("levonás után = %+v", w)
	}
	reset := int64(1000)
	w, err = st.UpdateAISubscription(ctx, ext, AISubscriptionUpdate{PeriodTokens: &reset})
	if err != nil || w.PeriodTokens != 1000 || w.Status != SubActive || w.ExternalSubscriptionID != ext {
		t.Errorf("megújítás = %+v, %v", w, err)
	}
	if _, err := st.UpdateAISubscription(ctx, "sub_nincs_"+newID(), AISubscriptionUpdate{Status: SubCanceled}); !errors.Is(err, ErrNotFound) {
		t.Errorf("ismeretlen azonosító: %v", err)
	}
	if got, err := st.AIWalletByExternalSubscription(ctx, ext); err != nil || got.Subject != subject {
		t.Errorf("keresés külső azonosítóval = %+v, %v", got, err)
	}
}
