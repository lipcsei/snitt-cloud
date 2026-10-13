# Snitt – architektúra

## Két, egymástól független fél

**A desktop app (Windows/macOS/Linux)** a termék. Indexeli a felhasználó saját videótárát,
Whisperrel és a meglévő feliratfájlokból átiratot készít (filmenként több nyelven), a kimondott
szövegre lehet benne keresni, és a találatot ki lehet vágni egy helyi videófájlba.

Ez az app **önmagában, fiók nélkül működik**, amíg a kezdő kreditkerete tart (lásd
[Üzleti modell](#üzleti-modell-kreditkeret-előfizetés-lifetime)). Minden adata – az index, az
átiratok, a beállítások – egy helyi SQLite fájlban él a felhasználó gépén. Videó vagy átirat
soha nem hagyja el a gépet. A kreditkereten belül a felhő oldal **teljes hiánya mellett is**
működnie kell: addig nincs kötelező bejelentkezés, nincs hálózati hívás a működéshez.

**A felhő oldal (ez a repo)** ehhez képest opcionális kiegészítés:

- **landing** (React, `landing/`, `:5174`) – a marketing oldal, ahol a látogató regisztrálhat és
  kezelheti a profilját.
- **admin** (React, `admin/`, `:5175`) – belső admin felület: felhasználók, előfizetések,
  számlázás. Nem publikus; a Keycloak `admin` realm szerepéhez kötött.
- **Keycloak** – az identitásszolgáltató. Ő birtokolja a felhasználót: regisztráció, jelszó,
  e-mail-cím, a stabil `sub` azonosító. Saját kódot erre nem írunk. **Nem ebben a repóban fut**:
  megosztott szolgáltatás ([`sso`](https://github.com/lipcsei/sso)), amit más appok (jelenleg a
  `breath`) is használnak – ez a repo egy `snitt` nevű saját realmet kap benne, elkülönítve a
  többi app felhasználóitól.
- **account szolgáltatás** (Go, `services/account/`) – az egyetlen saját backend. A Keycloak
  által kiállított JWT-t ellenőrzi, és Postgresben tárolja az alkalmazás-oldali profilt
  (megjelenítendő név, nyelv), az **entitlementeket** (mely extra funkciók járnak), valamint az
  **előfizetéseket** és a **számlákat**, amikből az entitlementek következnek.

## Miért van szétválasztva

A desktop app és a felhő oldal élettartama, kockázata és kiadási üteme különbözik. A desktop app
adatai privátak és nagy méretűek (videó), ezeknek nincs helyük szerveren. A felhő oldal viszont
csak apró, megosztható adatokat kezel (ki vagy, mire vagy jogosult). Ezért a felhő nem egy
"szinkron backend", hanem egy külön termék-fél, amit ki lehet kapcsolni.

A backend oldalon a szétválasztás szándékosan sekély: **egy** kicsi szolgáltatás,
mellette Keycloak és Postgres. Nincs API gateway, nincs üzenetsor, nincs service discovery.
Az identitás azért külön komponens (Keycloak), mert a hitelesítést nem érdemes saját kézzel
megírni; minden más egy binárisban elfér. Ha később kell egy második szolgáltatás (pl. fizetés),
az önállóan, ugyanezen a mintán jön be – az account szolgáltatás nem nő tovább.

## Az adatok folyása

```
[ desktop app ] --- helyi SQLite --- (nincs hálózat)
       :
       : később, opcionálisan: bejelentkezés (snitt-desktop kliens, PKCE, loopback redirect)
       v
[ Keycloak ] <-- OIDC + PKCE -- [ landing (React, :5174) ]
     |     ^                              |
     |     +---- OIDC + PKCE ------- [ admin (React, :5175) ]
     |                                    |
     | JWT (aud: snitt-landing)           | Bearer token
     |                                    v
     +-- JWKS / discovery --------> [ account (:8090) ] --> [ Postgres ]
     ^                                    |
     +-- Admin REST API (service account) +
        (snitt-admin-api kliens, csak olvasás)
```

A landing böngészőből, PKCE-vel jelentkezik be a `snitt-landing` publikus kliensen, majd a
kapott access tokent `Authorization: Bearer` fejlécben küldi az account szolgáltatásnak. Az
account nem hívja vissza a Keycloakot kérésenként: OIDC discoveryvel egyszer lehúzza a
kulcsokat, és helyben ellenőrzi az aláírást, az issuert, a lejáratot és az audience-t. Emiatt a
Keycloak kiesése a már kiadott tokeneket nem érvényteleníti, csak új bejelentkezést akadályoz.

A profil sort nem külön regisztrációs hívás hozza létre, hanem az első `GET /api/v1/me`
upsertje a token claimjeiből. Így a landingnek nincs "regisztráció után hívd meg ezt is"
lépése, és a Keycloakban létező, de itt még ismeretlen felhasználó sem tud inkonzisztens
állapotba kerülni.

## Az admin felület

Az `admin/` egy React SPA a `:5175` porton, ugyanazon a mintán, mint a landing: PKCE-vel
jelentkezik be a Keycloakon (`snitt-admin` publikus kliens), és `Authorization: Bearer` fejléccel
hívja az account szolgáltatás `/api/v1/admin/*` végpontjait.

**Jogosultság.** Nincs második hitelesítési mechanizmus: ugyanaz a Keycloak token azonosít, mint
mindenhol máshol, csak egy `admin` realm szerepet is hordoznia kell. A szerveren ez két réteg:
a meglévő token-ellenőrző middleware fölé kerül egy szerep-ellenőrző, ami a szerep hiányában
`403`-at ad. A böngésző oldali „Nincs jogosultságod” képernyő ehhez képest csak kényelmi
funkció, nem védelmi vonal.

**Felhasználói adat.** Az identitást a Keycloak birtokolja, ezért az admin felhasználólistája
onnan jön, a **Keycloak Admin REST API**-ról – egy bizalmas, service accountos kliensen
(`snitt-admin-api`, `realm-management: view-users`, `query-users`). Ezt az account szolgáltatás
kéri le, és a helyi `profiles` sorral fésüli össze; a Keycloak adatait szándékosan **nem**
másoljuk Postgresbe, hogy ne legyen két, egymástól elcsúszó igazság. Ennek az ára, hogy a
Keycloak kiesésekor a felhasználólista nem elérhető (`502`), a többi képernyő viszont működik.

**Táblák.** A `subscriptions` a fizetős csomagok nyilvántartása (subject, csomag, státusz,
elszámolási időszak, ár, `cancel_at_period_end`, külső azonosítók), az `invoices` pedig a saját
számlakönyv. Pénzt mindkettő a pénznem legkisebb egységében, egész számként tárol. Egy
felhasználónak egyszerre legfeljebb egy élő előfizetése lehet – ez egy részleges unique index az
adatbázisban, nem csak handler-szintű szabály, mert az entitlementek helyessége múlik rajta.

**Az entitlementek forrása.** Az előfizetés kiadása, a csomagváltás, a lemondás és a
visszakapcsolás ugyanabban a tranzakcióban írja az `entitlements` táblát, mint a
`subscriptions`-t. A csomag → funkciókulcs leképezés kódban van (`internal/billing`), és a
szinkronizáció csak a csomagok által vezérelt kulcsokat írja felül, a kézzel adott, csomagtól
független jogosultságokat érintetlenül hagyja. A desktop app így továbbra is egyetlen
végpontot, a `GET /api/v1/entitlements`-t kérdezi.

## Üzleti modell: kreditkeret, előfizetés, lifetime

A desktop app ingyen letölthető, és egy **kezdő kreditkerettel** indul: minden importált videó
(mappából vagy linkről) egy kreditet visz el. Amíg van kredit, fiók nélkül is teljes értékű.
Ha elfogyott, az alkalmazás zárolódik, és csak bejelentkezve, a **`snitt-app-access`**
jogosultsággal használható tovább. Ezt minden fizetős csomag megadja:

| Csomag | Ár | Fajta | Ki adhatja ki |
|---|---|---|---|
| `monthly` (Előfizetés) | 1 490 Ft / hó | ismétlődő | önkiszolgáló (Stripe) vagy admin |
| `lifetime` (Lifetime) | 24 990 Ft | egyszeri, soha nem jár le | önkiszolgáló (Stripe) vagy admin |
| `pro`, `studio` | 2 990 / 7 990 Ft | ismétlődő, AI-funkciókkal | admin |

A **lifetime** is egy `subscriptions` sor (így a „legfeljebb egy élő előfizetés” szabály és az
entitlement-szinkron változatlan), csak a periódusa 100 év, a jogosultságnak pedig nincs
lejárata. Az MRR nem számolja bele: az egyszeri díj nem havi bevétel.

## A fizetési szolgáltató: Stripe Checkout

A varrat továbbra is **`services/account/internal/billing`**, a `Provider` interfész. Két
implementáció van:

- **`Manual`** – nincs külső hívás; az admin kézzel ad előfizetést és jelöl kifizetettnek egy
  számlát. Ez az alapértelmezés, ha a `STRIPE_SECRET_KEY` üres.
- **`Stripe`** – a felhasználó maga vásárol. Az SDK helyett közvetlen HTTP-hívások: négy
  végpontról van szó, és így nincs nagy, gyakran változó függőség.

A folyamat:

1. A desktop app a bejelentkezett felhasználó tokenjével hívja a `POST /api/v1/checkout`-ot
   (`{"plan": "monthly" | "lifetime"}`). Az account egy Stripe Checkout munkamenetet nyit
   (havinál `mode=subscription`, lifetime-nál `mode=payment`), a `subject`-et a metadata-ba
   téve, és visszaadja a fizetési oldal címét. Az app ezt a **rendszer böngészőjében** nyitja
   meg. Kártyaadat soha nem jár nálunk.
2. A fizetés eredménye webhookon jön: `POST /api/v1/billing/webhook`. Nincs rajta
   token-ellenőrzés; a hitelességet a `Stripe-Signature` HMAC aláírás adja (5 perces
   időablakkal a visszajátszás ellen).
   - `checkout.session.completed` (vagy késleltetett fizetési módnál
     `checkout.session.async_payment_succeeded`) → előfizetés (vagy lifetime) kiadása + kifizetett számla,
   - `customer.subscription.updated` → időszak, státusz, periódus végi lemondás,
   - `customer.subscription.deleted` → lemondva, a jogosultság visszavonva,
   - `invoice.paid` (megújulás) → kifizetett számla a saját könyvbe.
3. A `billing_events` tábla a már feldolgozott eseményeket tartja: a Stripe ugyanazt többször
   is elküldheti. A feldolgozás hibája 5xx, hogy a Stripe újrapróbálja; a sorrend sem
   garantált, ezért az ismeretlen előfizetésre vonatkozó frissítés is 5xx (később újra jön).
4. A webhookból végzett műveletek is az admin naplóba kerülnek, `system:billing-webhook`
   szereplővel.

**Adó, számla, viták: Stripe Managed Payments.** Alapból (`STRIPE_MANAGED_PAYMENTS=true`) a
Stripe a kereskedő: ő számolja és fizeti be az áfát a vevő országa szerint, ő küldi a
számlát/nyugtát, és ő kezeli a visszatérítéseket és a vitás terheléseket, tranzakciónként +3,5%
díjért. Ehhez a Checkout `managed_payments[enabled]=true` paramétert kap, saját számlát
(`invoice_creation`) nem kér, és minden hívás `Stripe-Version: 2025-03-31.basil` fejlécet
küld (a webhook végpontot is erre a verzióra kell állítani). A Stripe termékeken az adókód
`txcd_10202000`, az árak adózási módja `inclusive`, így a kiírt ár a vevő végösszege. A saját
`invoices` táblába írt összeg ennek megfelelően bruttó, a könyvelési bizonylat a Stripe-é.

A `GET /api/v1/plans` nyilvános: a desktop app fizetési képernyője ebből mutat árat, és azt is
megmondja, be van-e kapcsolva az online fizetés (`checkout_enabled`). Csomagonként a `kind` mező
`access` (hozzáférést adó csomag) vagy `ai` (AI tokencsomag, lásd lent), az `ai_tokens` pedig az
AI csomag tokenkerete (egyébként `0`).

## AI tárca: Snitt AI token

Az okos keresés és a válaszadás a saját videótárból felhős AI-t használ, aminek valódi
önköltsége van. Ezt nem a hozzáférési csomagba építjük be, hanem **Snitt AI tokenben** adjuk
el, egy külön tárcából (`ai_wallets`). A videók és az átiratok továbbra sem kerülnek
szerverre: a desktop app helyben indexel, és hívásonként csak a beágyazandó szövegeket, illetve
a kérdést és a helyben megtalált részleteket küldi fel. Ezeket nem tároljuk, csak a
tokenszámokat (`ai_usage`).

| Csomag | Ár | Fajta | Tokenek |
|---|---|---|---|
| `ai_monthly` (AI csomag) | 2 990 Ft / hó | ismétlődő Stripe előfizetés | 1 000 000 / periódus, nem görgetődik át |
| `ai_tokens` (AI tokencsomag) | 1 990 Ft | egyszeri (`mode=payment`) | 500 000, soha nem jár le |

- Az AI csomag **nem** `subscriptions` sor, és nem ad entitlementet: a „legfeljebb egy élő
  előfizetés” szabály a hozzáférési csomagé, az AI csomag mellette, attól függetlenül vehető.
  Az admin előfizetés-kezelése (kiadás, csomagváltás) az AI csomagokat nem ismeri
  (`billing.LookupPlan` nem találja őket; a teljes lista a `billing.AllPlans`).
- A tárcának két egyenlege van: a **havi keret** (`period_tokens`, csak aktív AI előfizetésnél
  és a `period_end` előtt költhető; megújuláskor visszaáll a teljes keretre) és a **vásárolt**
  egyenleg (`purchased_tokens`, nem jár le). Levonáskor előbb a havi keret fogy (az úgyis
  elveszne), aztán a vásárolt. A levonás `SELECT … FOR UPDATE` alatt, egy tranzakcióban fut a
  használati napló írásával.
- Egyenleg-ellenőrzés a hívás előtt (0 → `402`), levonás a hívás után a ténylegesen
  elszámolt tokenek szerint. Az utolsó kérés így kicsit túlköltheti a maradékot: az egyenleg
  ilyenkor 0-n áll meg, a különbözetet elnyeljük.
- Az árak (`STRIPE_PRICE_AI_MONTHLY`, `STRIPE_PRICE_AI_TOKENS`) opcionálisak: ár nélkül a
  csomag nem jelenik meg a `GET /api/v1/plans`-ban, és a checkout `501`-et ad rá.

**Webhook.** A checkout esemény metadata-jában utazó csomagkulcs dönt: `ai_tokens` → jóváírás a
vásárolt egyenlegre; `ai_monthly` → a tárca AI előfizetése aktív, a keret teljes, a periódus
vége ideiglenesen most + 1 hónap + 2 nap (a pontosat a `customer.subscription.updated` hozza).
Az előfizetés-eseményeknél (`customer.subscription.updated/deleted`, `invoice.paid`) előbb az AI
tárcák közt keressük a külső azonosítót, és csak ha ott nincs, akkor a `subscriptions` táblában;
ha egyik sem ismeri, marad az `5xx` és a Stripe újrapróbálkozása. Megszűnéskor a havi keret
nullázódik, a vásárolt tokenek megmaradnak. A kifizetett számla minden esetben bekerül az
`invoices` táblába (előfizetés-hivatkozás nélkül), a művelet pedig a naplóba
(`ai.tokens_credit`, `ai.subscription`).

**Mérés.** 1 Snitt AI token ≈ 1 Claude Opus 5.5 bemeneti token. A levonás
`ceil(bemenet * w_be + kimenet * w_ki)`, legalább 1, ha bármennyi token fogyott:

| Modell | `w_be` | `w_ki` |
|---|---|---|
| `claude-opus-5-5` (alapértelmezett válaszadó) | 1 | 5 |
| `claude-sonnet-5-5` | 0,5 | 2,5 |
| `claude-haiku-5-5` | 0,025 | 0,125 |
| `voyage-4` (alapértelmezett beágyazó) | 0,02 | 0 |
| bármi más | 1 | 5 (az Opus súlya) |

**Végpontok** (mind bejelentkezett felhasználónak, `internal/httpapi/ai.go`):

- `GET /api/v1/ai/balance` – egyenleg és hogy be van-e kapcsolva az AI.
- `POST /api/v1/ai/embed` – beágyazás a Voyage AI-jal (`voyage-4`, 1024 dimenzió); 1–128
  szöveg, egyenként legfeljebb 8000 karakter, `input_type`: `query` vagy `document`.
- `POST /api/v1/ai/answer` – rövid, forrásolt válasz Claude-dal (hivatalos Anthropic Go SDK,
  alacsony erőfeszítés). A modell csak a számozott részletekből válaszolhat, `[n]` jelöléssel
  hivatkozik; ezekből lesz a `cited_ids`. Ha a modell visszautasítja a kérdést, `502`, de a
  felhasznált tokeneket levonjuk.

Bármelyik szolgáltató kulcsa hiányozhat (`VOYAGE_API_KEY`, `ANTHROPIC_API_KEY`): ilyenkor az
adott végpont `503`. A válaszok pontos alakja a `services/account/README.md`-ben van.

## A desktop app jövőbeli belépése

A realmben már benne van a `snitt-desktop` publikus kliens (PKCE, `http://127.0.0.1:53312/callback`
loopback redirect), de még nincs használatban. Amikor a desktop app bejelentkezik, ugyanazt a
`GET /api/v1/entitlements` végpontot fogja hívni, amit a landing – nem kell külön API. A desktop
kliens tokenjébe egy audience mapper teszi bele a `snitt-landing` audience-t, hogy az account
szolgáltatásnak egyetlen elvárt audience-t kelljen ellenőriznie.

A bejelentkezés addig opcionális, amíg a kezdő kreditkeret tart; utána a használathoz a
`snitt-app-access` jogosultság kell (lásd fent). A videók és az átiratok ekkor sem kerülnek
szerverre: a felhő csak azt tudja, ki vagy és mire vagy jogosult.
