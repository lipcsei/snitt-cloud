# Snitt – architektúra

## Két, egymástól független fél

**A desktop app (Windows/macOS/Linux)** a termék. Indexeli a felhasználó saját videótárát,
Whisperrel és a meglévő feliratfájlokból átiratot készít (filmenként több nyelven), a kimondott
szövegre lehet benne keresni, és a találatot ki lehet vágni egy helyi videófájlba.

Ez az app **önmagában, fiók nélkül, teljes értékűen működik**. Minden adata – az index, az
átiratok, a beállítások – egy helyi SQLite fájlban él a felhasználó gépén. Videó vagy átirat
soha nem hagyja el a gépet. A felhő oldal **teljes hiánya mellett is** működnie kell: nincs
kötelező bejelentkezés, nincs hálózati hívás a működéshez.

**A felhő oldal (ez a repo)** ehhez képest opcionális kiegészítés:

- **landing** (React, `landing/`, `:5174`) – a marketing oldal, ahol a látogató regisztrálhat és
  kezelheti a profilját.
- **admin** (React, `admin/`, `:5175`) – belső admin felület: felhasználók, előfizetések,
  számlázás. Nem publikus; a Keycloak `admin` realm szerepéhez kötött.
- **Keycloak** – az identitásszolgáltató. Ő birtokolja a felhasználót: regisztráció, jelszó,
  e-mail-cím, a stabil `sub` azonosító. Saját kódot erre nem írunk.
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

## Hol csatlakozik majd a fizetési szolgáltató

Valódi fizetési szolgáltató (Stripe, Paddle) **nincs bekötve**, és a rendszer senkit nem
terhel. Az admin kézzel ad előfizetést és kézzel jelöl kifizetettnek egy számlát – ez a
könyvelés vezetése, nem fizetés.

A varrat egy helyen van: **`services/account/internal/billing`**, a `Provider` interfész. Ma
egyetlen implementációja a `Manual`, ami semmilyen külső hívást nem végez. A HTTP réteg soha nem
hív szolgáltató-specifikus kódot, csak ezt az interfészt (`EnsureCustomer`, `StartSubscription`,
`CancelSubscription`), ezért egy valódi szolgáltató bekötése:

1. új `Provider` implementáció ugyanebben a csomagban,
2. a `cmd/server/main.go` a `billing.Manual{}` helyett azt adja a `httpapi.Options`-nek,
3. a szolgáltató webhookja egy új végponton a már meglévő `store.MarkInvoicePaid` /
   előfizetés-frissítő függvényeket hívja.

A séma ehhez már készen áll: a `subscriptions` és az `invoices` tábla is hordoz `provider` és
`external_*` oszlopot, ma `manual` / üres értékkel. Ezért a bekötéshez nem kell séma-átalakítás,
és a meglévő, kézzel felvitt sorok is megkülönböztethetők maradnak.

## A desktop app jövőbeli belépése

A realmben már benne van a `snitt-desktop` publikus kliens (PKCE, `http://127.0.0.1:53312/callback`
loopback redirect), de még nincs használatban. Amikor a desktop app bejelentkezik, ugyanazt a
`GET /api/v1/entitlements` végpontot fogja hívni, amit a landing – nem kell külön API. A desktop
kliens tokenjébe egy audience mapper teszi bele a `snitt-landing` audience-t, hogy az account
szolgáltatásnak egyetlen elvárt audience-t kelljen ellenőriznie.

Fontos, hogy ez **opcionális marad**: a bejelentkezés csak extra funkciókat kapcsol be, a
meglévő működés bejelentkezés nélkül változatlan.
