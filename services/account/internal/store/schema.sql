-- A séma szándékosan pici: a Keycloak birtokolja az identitást (sub, email),
-- itt csak az alkalmazás-oldali kiegészítő adatok élnek.

CREATE TABLE IF NOT EXISTS profiles (
    subject      TEXT PRIMARY KEY,
    email        TEXT NOT NULL DEFAULT '',
    username     TEXT NOT NULL DEFAULT '',
    display_name TEXT NOT NULL DEFAULT '',
    locale       TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS entitlements (
    subject     TEXT NOT NULL,
    feature_key TEXT NOT NULL,
    granted_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ,
    PRIMARY KEY (subject, feature_key)
);

-- A desktop app indulásonként egyszer kérdezi le a jogosultságokat,
-- a subject szerinti szűrés a jellemző hozzáférési minta.
CREATE INDEX IF NOT EXISTS entitlements_subject_idx ON entitlements (subject);

-- Előfizetések. A pénzösszegek egész számban, a pénznem legkisebb egységében
-- tárolódnak (HUF-nál fillér, tehát 2 990 Ft = 299000), hogy sose kelljen
-- lebegőpontos számmal pénzt számolni.
--
-- A provider és az external_* oszlopok ma "manual"/üresek: nincs fizetési
-- szolgáltató bekötve. Azért léteznek már most, hogy egy későbbi
-- Stripe/Paddle integráció ne igényeljen séma-átalakítást.
CREATE TABLE IF NOT EXISTS subscriptions (
    id                       TEXT PRIMARY KEY,
    subject                  TEXT NOT NULL,
    plan                     TEXT NOT NULL,
    status                   TEXT NOT NULL,
    current_period_start     TIMESTAMPTZ NOT NULL DEFAULT now(),
    current_period_end       TIMESTAMPTZ NOT NULL,
    price_minor              BIGINT NOT NULL DEFAULT 0,
    currency                 TEXT NOT NULL DEFAULT 'HUF',
    cancel_at_period_end     BOOLEAN NOT NULL DEFAULT false,
    canceled_at              TIMESTAMPTZ,
    provider                 TEXT NOT NULL DEFAULT 'manual',
    external_customer_id     TEXT NOT NULL DEFAULT '',
    external_subscription_id TEXT NOT NULL DEFAULT '',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT subscriptions_status_chk
        CHECK (status IN ('active', 'past_due', 'canceled', 'expired')),
    CONSTRAINT subscriptions_period_chk
        CHECK (current_period_end > current_period_start)
);

-- Egy felhasználónak egyszerre legfeljebb egy élő előfizetése lehet: ez a
-- szabály az adatbázisban él, nem csak a handlerben, mert az entitlementek
-- helyessége múlik rajta.
CREATE UNIQUE INDEX IF NOT EXISTS subscriptions_one_live_per_subject_idx
    ON subscriptions (subject)
    WHERE status IN ('active', 'past_due');

CREATE INDEX IF NOT EXISTS subscriptions_subject_idx ON subscriptions (subject);
CREATE INDEX IF NOT EXISTS subscriptions_status_idx ON subscriptions (status);

-- Emberi olvasásra szánt számlaszám (SNITT-2026-000042) sorszámozása.
CREATE SEQUENCE IF NOT EXISTS invoice_number_seq;

-- Számlák. Ez a saját könyvelési nyilvántartás, amivel az admin ma dolgozik.
-- Valódi terhelés nem történik: a "paid" státuszt ember állítja be, miután a
-- pénz máshol megérkezett. A külső szolgáltató később ugyanezeket a sorokat
-- fogja írni webhookból.
CREATE TABLE IF NOT EXISTS invoices (
    id                  TEXT PRIMARY KEY,
    number              TEXT NOT NULL UNIQUE,
    subject             TEXT NOT NULL,
    subscription_id     TEXT REFERENCES subscriptions (id) ON DELETE SET NULL,
    amount_minor        BIGINT NOT NULL,
    currency            TEXT NOT NULL DEFAULT 'HUF',
    status              TEXT NOT NULL,
    issued_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    due_at              TIMESTAMPTZ,
    paid_at             TIMESTAMPTZ,
    provider            TEXT NOT NULL DEFAULT 'manual',
    external_invoice_id TEXT NOT NULL DEFAULT '',
    note                TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT invoices_status_chk CHECK (status IN ('open', 'paid', 'void')),
    CONSTRAINT invoices_amount_chk CHECK (amount_minor >= 0),
    -- A kifizetés dátuma és a státusz nem térhet el egymástól.
    CONSTRAINT invoices_paid_at_chk
        CHECK ((status = 'paid') = (paid_at IS NOT NULL))
);

CREATE INDEX IF NOT EXISTS invoices_subject_idx ON invoices (subject);
CREATE INDEX IF NOT EXISTS invoices_status_idx ON invoices (status);
CREATE INDEX IF NOT EXISTS invoices_issued_at_idx ON invoices (issued_at DESC);

-- Admin napló. Ez pénzt érintő belső eszköz: minden módosító műveletről
-- marad nyom arról, hogy KI, MIKOR, MIT csinált. A sorok soha nem
-- módosulnak és nem törlődnek - a napló csak nő.
--
-- A summary emberi mondat (magyarul), a detail a művelet nyers paraméterei.
-- A kettő szándékos redundancia: a summary akkor is olvasható marad, ha a
-- kód később átalakul, a detail pedig akkor is elég, ha a summary kevés.
CREATE TABLE IF NOT EXISTS admin_audit (
    id            TEXT PRIMARY KEY,
    at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    actor_subject TEXT NOT NULL,
    -- Az adminról a tokenben látott címke (e-mail vagy felhasználónév).
    -- Azért másoljuk ide, mert a Keycloak fiók később törölhető, a napló
    -- viszont attól még legyen olvasható.
    actor_label   TEXT NOT NULL DEFAULT '',
    action        TEXT NOT NULL,
    target_type   TEXT NOT NULL DEFAULT '',
    target_id     TEXT NOT NULL DEFAULT '',
    -- Az érintett felhasználó, ha a művelethez tartozik ilyen.
    subject       TEXT NOT NULL DEFAULT '',
    summary       TEXT NOT NULL DEFAULT '',
    detail        JSONB NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT admin_audit_action_chk CHECK (action <> '')
);

CREATE INDEX IF NOT EXISTS admin_audit_at_idx ON admin_audit (at DESC);
CREATE INDEX IF NOT EXISTS admin_audit_subject_idx ON admin_audit (subject);
CREATE INDEX IF NOT EXISTS admin_audit_actor_idx ON admin_audit (actor_subject);
CREATE INDEX IF NOT EXISTS admin_audit_action_idx ON admin_audit (action);
