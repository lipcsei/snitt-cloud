package store

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

// A naplónál a kockázat nem a Go kódban van, hanem az SQL-ben és a JSONB
// oda-vissza alakításában - ezt csak valódi Postgresen lehet ellenőrizni.
// TEST_DATABASE_URL nélkül a teszt kimarad, hogy a szokásos `go test ./...`
// adatbázis nélkül is fusson.
func TestAuditRoundTrip(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL nincs beállítva")
	}
	ctx := context.Background()

	st, err := New(ctx, url)
	if err != nil {
		t.Fatalf("kapcsolódás: %v", err)
	}
	// Cleanup és nem defer: a takarító DELETE-nek a pool bezárása ELŐTT kell
	// lefutnia, a t.Cleanup pedig fordított sorrendben hívódik.
	t.Cleanup(st.Close)
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("séma: %v", err)
	}

	// Saját, teszt-specifikus subject, hogy a valódi sorok ne zavarjanak be.
	subject := "test-subject-" + newID()
	actor := "test-actor-" + newID()

	in := NewAuditEntry{
		ActorSubject: actor,
		ActorLabel:   "teszt@snitt.local",
		Action:       AuditInvoiceVoid,
		TargetType:   "invoice",
		TargetID:     "inv-1",
		Subject:      subject,
		Summary:      "Számla sztornózva: SNITT-2026-000001, 2 990,00 HUF. Indok: téves kiállítás",
		Detail: map[string]any{
			"number": "SNITT-2026-000001", "amount_minor": int64(299000),
			"currency": "HUF", "note": "téves kiállítás",
		},
	}
	written, err := st.RecordAudit(ctx, in)
	if err != nil {
		t.Fatalf("napló írása: %v", err)
	}
	if written.ID == "" || written.At.IsZero() {
		t.Fatalf("hiányos visszakapott sor: %+v", written)
	}
	if written.Detail["number"] != "SNITT-2026-000001" {
		t.Errorf("detail visszaolvasva = %+v", written.Detail)
	}

	got, total, err := st.ListAudit(ctx, AuditFilter{Subject: subject, Limit: 10})
	if err != nil {
		t.Fatalf("napló lekérdezése: %v", err)
	}
	if total != 1 || len(got) != 1 {
		t.Fatalf("találatok = %d (total %d)", len(got), total)
	}
	if got[0].Summary != in.Summary || got[0].ActorLabel != in.ActorLabel {
		t.Errorf("visszaolvasott sor = %+v", got[0])
	}
	// A számok JSONB-ből float64-ként jönnek vissza: ez a napló olvasásánál
	// rendben van, de tudni kell róla.
	if v, ok := got[0].Detail["amount_minor"].(float64); !ok || v != 299000 {
		t.Errorf("amount_minor = %#v", got[0].Detail["amount_minor"])
	}

	// Szűrés: más művelet, más végrehajtó, jövőbeli kezdet - egyik sem talál.
	for name, f := range map[string]AuditFilter{
		"másik művelet":      {Subject: subject, Action: AuditInvoicePay, Limit: 10},
		"másik végrehajtó":   {Subject: subject, Actor: "valaki-mas", Limit: 10},
		"jövőbeli kezdettől": {Subject: subject, Since: ptr(time.Now().UTC().Add(time.Hour)), Limit: 10},
	} {
		rows, n, err := st.ListAudit(ctx, f)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if n != 0 || len(rows) != 0 {
			t.Errorf("%s: %d találat, várt 0", name, len(rows))
		}
	}

	// A teszt nem hagy szemetet a naplóban.
	t.Cleanup(func() {
		if _, err := st.pool.Exec(ctx, `DELETE FROM admin_audit WHERE subject = $1`, subject); err != nil {
			t.Errorf("takarítás: %v", err)
		}
	})
}

func ptr[T any](v T) *T { return &v }

// A kézi jogosultság kiadása felülír (upsert) és visszavon. Ugyanaz a
// kockázat, mint a naplónál: az SQL, nem a Go kód.
func TestEntitlementGrantRevoke(t *testing.T) {
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
	t.Cleanup(func() {
		if _, err := st.pool.Exec(ctx, `DELETE FROM entitlements WHERE subject = $1`, subject); err != nil {
			t.Errorf("takarítás: %v", err)
		}
	})

	// Először lejárat nélkül.
	e, err := st.GrantEntitlement(ctx, subject, "beta-access", nil)
	if err != nil {
		t.Fatalf("kiadás: %v", err)
	}
	if e.ExpiresAt != nil {
		t.Errorf("lejárat = %v, várt nil", e.ExpiresAt)
	}

	// Ugyanaz a kulcs újra: nem ütközés, hanem hosszabbítás.
	until := time.Now().UTC().Add(30 * 24 * time.Hour).Truncate(time.Second)
	e, err = st.GrantEntitlement(ctx, subject, "beta-access", &until)
	if err != nil {
		t.Fatalf("ismételt kiadás: %v", err)
	}
	if e.ExpiresAt == nil || !e.ExpiresAt.Equal(until) {
		t.Errorf("lejárat = %v, várt %v", e.ExpiresAt, until)
	}

	all, err := st.AllEntitlements(ctx, subject)
	if err != nil {
		t.Fatalf("lekérdezés: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("jogosultságok = %+v, várt 1 sor", all)
	}

	if _, err := st.RevokeEntitlement(ctx, subject, "beta-access"); err != nil {
		t.Fatalf("visszavonás: %v", err)
	}
	// Ami nincs kiadva, arra ErrNotFound jön - nem néma siker.
	if _, err := st.RevokeEntitlement(ctx, subject, "beta-access"); !errors.Is(err, ErrNotFound) {
		t.Errorf("ismételt visszavonás hibája = %v, várt ErrNotFound", err)
	}
}
