# Snitt account szolgáltatás

Kicsi Go HTTP szolgáltatás a Snitt **opcionális felhő oldalához**. Négy dolgot csinál:

1. **Profil** – a landing oldalon regisztrált felhasználó alkalmazás-oldali profilját tárolja
   (megjelenítendő név, nyelv). Az identitást (jelszó, e-mail, `sub`) a Keycloak birtokolja,
   ez a szolgáltatás csak kiegészítő adatokat tart.
2. **Entitlements** – melyik prémium funkcióra jogosult a bejelentkezett felhasználó.
   Ezt később a desktop app is le fogja kérdezni; a végpont már most létezik.
3. **Admin API** – a belső admin felület ([`admin/`](../../admin)) mögötti végpontok:
   felhasználók, előfizetések, számlázás és összesítők. Csak `admin` realm szereppel.
4. **Admin napló** – minden előfizetést, számlát, jogosultságot és fiókállapotot módosító admin
   művelet nyomot hagy az `admin_audit` táblában: ki, mikor, mit csinált.

A desktop app ettől függetlenül, fiók nélkül is teljesen működik – lásd
[`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md).

## Futtatás

### A teljes fejlesztői stack (ajánlott)

A Keycloak a közös [`sso`](https://github.com/lipcsei/sso) repóban fut (`http://localhost:8081`, benne a
`snitt` realm), azt kell **előbb** elindítani. Utána a repo gyökeréből:

```bash
make -C ../sso up     # a közös Keycloak; ő hozza létre az sso-net hálózatot
make up
```

Ez elindítja a Postgrest, az admint, a landinget és ezt a szolgáltatást a `http://localhost:8090` címen.
A kép építése a privát `github.com/lipcsei/commons` Go modult SSH-n éri el (`--ssh default`), ezért
olyan SSH-ügynök kell, ami hozzáfér a GitHubhoz.

### Csak a szolgáltatás, helyben

Feltételezi, hogy a Postgres és a Keycloak már fut. A repo gyökeréből a `make dev-account` ugyanezt
csinálja (létrehozza a `.env`-et, és előbb leállítja a compose `account` konténerét, ami ugyanazt a
8090-es portot fogja):

```bash
cp .env.example .env      # majd szerkeszd, ha kell
set -a && source .env && set +a
go run ./cmd/server
```

A privát `commons` modul miatt a `go` parancsnak kell a `GOPRIVATE=github.com/lipcsei/*` beállítás és
git-hozzáférés a repóhoz.

Az adatbázis-séma indulásnál automatikusan létrejön (beágyazott `schema.sql`,
csupa `CREATE TABLE IF NOT EXISTS`), külön migrációs lépés nincs.

### Tesztek

```bash
go build ./... && go vet ./... && go test ./...
```

A repo gyökeréből: `make test` (csak a tesztek), `make lint` (gofmt-ellenőrzés, `go vet`,
golangci-lint), `make check` (minden, amit a CI is futtat).

A handler-tesztekhez nem kell sem élő Keycloak, sem adatbázis: a token-ellenőrzés az
`auth.TokenVerifier` interfész mögött, a tárolás a `httpapi.Store` interfész mögött van kicserélve.

## Környezeti változók

| Változó | Alapértelmezés | Leírás |
| --- | --- | --- |
| `HTTP_ADDR` | `:8090` | HTTP figyelési cím. |
| `DATABASE_URL` | `postgres://snitt:snitt@localhost:5432/snitt?sslmode=disable` | PostgreSQL kapcsolat. |
| `KEYCLOAK_ISSUER` | `http://localhost:8081/realms/snitt` | A realm issuer URL-je; ennek egyeznie kell a tokenek `iss` claimjével. |
| `KEYCLOAK_DISCOVERY_URL` | = `KEYCLOAK_ISSUER` | Opcionális. Akkor kell, ha a discovery más címen érhető el, mint az issuer (Dockerben: `http://keycloak:8080/realms/snitt`, az `sso-net` hálózaton). |
| `KEYCLOAK_AUDIENCE` | `snitt-landing,snitt-admin` | A tokenben elfogadott `aud` értékek, vesszővel elválasztva (minden kliens a saját client id-jével kap tokent). |
| `CORS_ORIGINS` | `http://localhost:5174,http://localhost:5175` | Vesszővel elválasztott lista a böngészőből hívó originokról (landing, admin). |
| `KEYCLOAK_BASE_URL` | `http://localhost:8081` | A Keycloak gyökere az Admin REST API-hoz. Dockerben: `http://keycloak:8080` (az `sso-net` hálózaton). |
| `KEYCLOAK_REALM` | `snitt` | A realm neve az Admin API útvonalakhoz. |
| `KEYCLOAK_ADMIN_CLIENT_ID` | `snitt-admin-api` | Bizalmas, service accountos kliens a felhasználók olvasásához és a fiók letiltásához. |
| `KEYCLOAK_ADMIN_CLIENT_SECRET` | `snitt-admin-api-dev-secret` | Fejlesztői titok; élesben kötelezően felülírandó. |
| `ADMIN_ROLE` | `admin` | Az a realm szerep, ami az `/api/v1/admin/*` végpontokat nyitja. |
| `DEFAULT_CURRENCY` | `HUF` | Alapértelmezett pénznem (ISO 4217). |
| `SENTRY_DSN` | *(üres)* | Opcionális hibajelentés a GlitchTipbe (a projekt *Client Keys (DSN)* oldaláról). **Üresen teljesen kikapcsolva**: az SDK el sem indul. Composeban a `SENTRY_DSN_ACCOUNT`-ból jön. |
| `SENTRY_ENVIRONMENT` | `development` | A környezet neve az eseményeken (az éles compose `production`-t állít). |
| `SENTRY_RELEASE` | *(üres)* | A kiadás azonosítója az eseményeken (opcionális). |

Hibajelentéskor a handlerben keletkező pánik (stacktrace-szel) és minden visszaadott `5xx` válasz
egy-egy eseményt ad; a `4xx`-et nem jelenti. Személyes adat, kérés törzse és lekérdezés-szöveg nem
kerül bele. Részletek: [`docs/VPS-TELEPITES.md`](../../docs/VPS-TELEPITES.md) 11. pont.

A service accountnak a `realm-management` kliensen `view-users` (a lapozáshoz `query-users`, a
fiók letiltásához/engedélyezéséhez pedig `manage-users`) szerep kell; a Keycloak (megosztott
szolgáltatás, lásd [`sso`](https://github.com/lipcsei/sso)) `snitt-realm.json`-ja ezt már
tartalmazza.

> A realm import **csak akkor fut le, ha a realm még nem létezik.** Egy már működő Keycloakon
> tehát az export módosítása önmagában nem hat: a `manage-users` szerepet pótolni kell. Erre
> való az sso repó [`scripts/grant-manage-users.sh`](https://github.com/lipcsei/sso/blob/main/scripts/grant-manage-users.sh)
> szkriptje, vagy kézzel: Clients → `snitt-admin-api` → Service accounts roles → Assign role →
> `realm-management manage-users`. Enélkül a fiók letiltása `502`-t ad, a hibaüzenetben a
> Keycloak `403`-mal.

## Végpontok

Minden `/api/v1/...` végpont `Authorization: Bearer <access_token>` fejlécet vár; hiányzó vagy
érvénytelen token esetén a válasz `401`.

### `GET /healthz`

Hitelesítés nélkül hívható. Rövid időkorláttal megpingeli az adatbázist.
`200` `{"status":"ok","database":"up"}`, vagy `503` `{"status":"degraded","database":"down"}`, ha az
adatbázis nem elérhető.

### `GET /api/v1/me`

A hívó profilja. Ha még nincs sor, az első híváskor létrejön a token claimjeiből
(nincs külön regisztrációs végpont – azt már a Keycloak elvégezte).

```json
{
  "sub": "3f0c...",
  "email": "user@example.com",
  "username": "user",
  "display_name": "Teszt Elek",
  "locale": "hu-HU",
  "created_at": "2026-09-10T10:00:00Z",
  "updated_at": "2026-09-10T10:00:00Z"
}
```

### `PATCH /api/v1/me`

Csak a `display_name` és a `locale` módosítható. A nem küldött mező változatlan marad.
Az identitást birtokló mezők (`sub`, `email`) küldése `400`-at eredményez, nem csendes eldobást.

```json
{ "display_name": "Új Név", "locale": "hu-HU" }
```

Válasz: a frissített profil (`200`).

### `GET /api/v1/entitlements`

A hívó le nem járt jogosultságai. **Az üres lista normál válasz, nem hiba.**

```json
{
  "entitlements": [
    { "feature_key": "cloud-sync", "granted_at": "2026-09-01T00:00:00Z", "expires_at": null }
  ]
}
```

A válasz szándékosan objektum, nem csupasz tömb: így később bővíthető a kliensek törése nélkül.
Jogosultságot az előfizetés kiadása ad ki automatikusan (a csomag `feature_keys` mezője
szerint), a csomagon kívülieket pedig az admin, kézzel – lásd a következő végpontot.

### Az admin API áttekintése

Minden `/api/v1/admin/*` végpont az `ADMIN_ROLE` realm szerepet kéri; enélkül `403`. A teljes lista
(`internal/httpapi/admin.go`), az alábbi szakaszok csak a nem magától értetődőket részletezik:

| Végpont | Mit csinál |
| --- | --- |
| `GET /api/v1/admin/overview` | A vezérlőpult összesítői. |
| `GET /api/v1/admin/plans` | A választható csomagok (`internal/billing`). |
| `GET /api/v1/admin/users` | Felhasználólista a Keycloakból, a helyi profillal összefésülve. |
| `GET /api/v1/admin/users/{subject}` | Egy felhasználó adatlapja. |
| `PATCH /api/v1/admin/users/{subject}` | A Keycloak fiók engedélyezése / letiltása. |
| `POST /api/v1/admin/users/{subject}/entitlements` | Kézi jogosultság-kiadás. |
| `DELETE /api/v1/admin/users/{subject}/entitlements/{key}` | Jogosultság visszavonása. |
| `GET /api/v1/admin/subscriptions` | Előfizetések listája. |
| `POST /api/v1/admin/subscriptions` | Előfizetés kiadása. |
| `GET /api/v1/admin/subscriptions/{id}` | Egy előfizetés. |
| `PATCH /api/v1/admin/subscriptions/{id}` | Csomagváltás. |
| `POST /api/v1/admin/subscriptions/{id}/cancel` | Lemondás. |
| `POST /api/v1/admin/subscriptions/{id}/reactivate` | Visszakapcsolás. |
| `GET /api/v1/admin/invoices` | Számlák listája. |
| `POST /api/v1/admin/invoices` | Számla kiállítása. |
| `GET /api/v1/admin/invoices/{id}` | Egy számla. |
| `POST /api/v1/admin/invoices/{id}/pay` | Kifizetettre jelölés. |
| `POST /api/v1/admin/invoices/{id}/void` | Sztornó. |
| `GET /api/v1/admin/audit` | Az admin napló. |

### `POST /api/v1/admin/users/{subject}/entitlements`

Kézi jogosultság-kiadás. A kulcs szándékosan szabad szöveg (`^[a-z0-9]([a-z0-9._-]{0,62}[a-z0-9])?$`):
a csomagokon kívüli kulcsok – béta-hozzáférés, egyedi megállapodás – éppen ettől lehetségesek.

```json
{ "feature_key": "beta-access", "days": 30, "note": "egyedi megállapodás szerint" }
```

A `days` elhagyása **lejárat nélküli** jogosultságot ad; ugyanarra a kulcsra ismételve nem
ütközik, hanem hosszabbít. Válasz `201`, a kiadott jogosultsággal.

Ha a kulcs csomaghoz tartozik (`ai-semantic-search`, `ai-video-understanding`), a válasz
`warning` mezőt is hoz: az ilyen kulcsokat a felhasználó következő előfizetés-művelete
felülírja, mert azokat a csomag vezérli.

### `DELETE /api/v1/admin/users/{subject}/entitlements/{key}`

Jogosultság visszavonása. Ha nem volt kiadva: `404` – nem néma siker.

### `PATCH /api/v1/admin/users/{subject}`

A Keycloak fiók engedélyezése vagy letiltása. Ez az egyetlen admin művelet, ami az
identitást birtokló rendszerbe ír.

```json
{ "enabled": false, "note": "visszaélés gyanúja" }
```

Válasz `200` `{"enabled": false, "changed": true}`. Ha a fiók már ebben az állapotban volt,
`changed: false`, és nem történik sem Keycloak-írás, sem naplózás. A letiltott felhasználó nem
tud belépni és a tokenjét sem tudja frissíteni; a jogosultságai és a számlái érintetlenek
maradnak.

Ehhez a service accountnak `manage-users` szerep kell (lásd fent); enélkül a válasz `502`.

### `GET /api/v1/admin/audit`

Az admin napló, legfrissebb elöl. Csak `admin` realm szereppel.

Szűrők (mind opcionális): `action`, `subject` (érintett felhasználó), `actor` (a művelet
végrehajtója), `from`, `to` (`2026-09-10` vagy teljes RFC3339; a `to` napra pontos alakja
**bezárólag** értendő), `page`, `page_size`.

```json
{
  "entries": [
    {
      "id": "9f1c…",
      "at": "2026-09-10T19:12:00Z",
      "actor_subject": "3f0c…",
      "actor_label": "admin@snitt.video",
      "action": "invoice.void",
      "target_type": "invoice",
      "target_id": "b21a…",
      "subject": "7cd2…",
      "summary": "Számla sztornózva: SNITT-2026-000042, 2 990,00 HUF. Indok: téves kiállítás",
      "detail": { "number": "SNITT-2026-000042", "amount_minor": 299000, "currency": "HUF" }
    }
  ],
  "page": 1,
  "page_size": 50,
  "total": 1,
  "actions": ["subscription.grant", "…"]
}
```

Az `actions` a szűrő lehetséges értékeit adja, hogy a felület ne másolja le a szerver listáját.

**A napló csak nő**: bejegyzést módosítani vagy törölni egyetlen végpont sem tud. A naplózás a
művelet UTÁN, külön írásban történik, és a bukása **nem** bukatja el a műveletet – a pénzt érintő
lépés ilyenkor már megtörtént, egy hibás válasz csak félrevezetné az admint. A sikertelen
naplóírás `ERROR` szinten kimegy a logba.

Az `admin_audit` az account szolgáltatás adatbázisában van, tehát
a repó `make backup` célja menti (lásd a gyökér [README](../../README.md)-jét).

## Felépítés

```
cmd/server/main.go        indítás, graceful shutdown
internal/config           env-alapú konfiguráció
internal/auth             OIDC token-ellenőrzés + middleware (TokenVerifier interfész)
internal/store            pgxpool, beágyazott schema.sql, lekérdezések
internal/billing          csomagok, csomag → funkciókulcs leképezés, a fizetési Provider interfész
internal/httpapi          routing, handlerek, CORS, admin napló, tesztek
```

A hibajelentés (pánik és 5xx, Sentry SDK) és a Keycloak Admin API kliense nem itt él, hanem a közös,
privát [`commons`](https://github.com/lipcsei/commons) Go modulban (`errtrack`, `keycloak`).
