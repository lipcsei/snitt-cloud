# Snitt account szolgáltatás

Kicsi Go HTTP szolgáltatás a Snitt **opcionális felhő oldalához**. Két dolgot csinál:

1. **Profil** – a landing oldalon regisztrált felhasználó alkalmazás-oldali profilját tárolja
   (megjelenítendő név, nyelv). Az identitást (jelszó, e-mail, `sub`) a Keycloak birtokolja,
   ez a szolgáltatás csak kiegészítő adatokat tart.
2. **Entitlements** – melyik prémium funkcióra jogosult a bejelentkezett felhasználó.
   Ezt később a desktop app is le fogja kérdezni; a végpont már most létezik.
3. **Admin API** – a belső admin felület ([`admin/`](../../admin)) mögötti végpontok:
   felhasználók, előfizetések, számlázás és összesítők. Csak `admin` realm szereppel.
4. **Admin napló** – minden előfizetést és számlát módosító admin művelet nyomot hagy az
   `admin_audit` táblában: ki, mikor, mit csinált.

A desktop app ettől függetlenül, fiók nélkül is teljesen működik – lásd
[`docs/ARCHITECTURE.md`](../../docs/ARCHITECTURE.md).

## Futtatás

### A teljes fejlesztői stack (ajánlott)

A repo gyökeréből:

```bash
docker compose -f deploy/docker-compose.yml up --build
```

Ez elindítja a Postgrest, a Keycloakot (`http://localhost:8081`, admin: `admin` / `admin`,
`snitt` realm importtal) és ezt a szolgáltatást a `http://localhost:8090` címen.

### Csak a szolgáltatás, helyben

Feltételezi, hogy a Postgres és a Keycloak már fut.

```bash
cp .env.example .env      # majd szerkeszd, ha kell
set -a && source .env && set +a
go run ./cmd/server
```

Az adatbázis-séma indulásnál automatikusan létrejön (beágyazott `schema.sql`,
csupa `CREATE TABLE IF NOT EXISTS`), külön migrációs lépés nincs.

### Tesztek

```bash
go build ./... && go vet ./... && go test ./...
```

A handler-tesztekhez nem kell sem élő Keycloak, sem adatbázis: a token-ellenőrzés az
`auth.TokenVerifier` interfész mögött, a tárolás a `httpapi.Store` interfész mögött van kicserélve.

## Környezeti változók

| Változó | Alapértelmezés | Leírás |
| --- | --- | --- |
| `HTTP_ADDR` | `:8090` | HTTP figyelési cím. |
| `DATABASE_URL` | `postgres://snitt:snitt@localhost:5432/snitt?sslmode=disable` | PostgreSQL kapcsolat. |
| `KEYCLOAK_ISSUER` | `http://localhost:8081/realms/snitt` | A realm issuer URL-je; ennek egyeznie kell a tokenek `iss` claimjével. |
| `KEYCLOAK_DISCOVERY_URL` | = `KEYCLOAK_ISSUER` | Opcionális. Akkor kell, ha a discovery más címen érhető el, mint az issuer (Dockerben: `http://keycloak:8081/realms/snitt`). |
| `KEYCLOAK_AUDIENCE` | `snitt-landing` | A tokenben elvárt `aud`. |
| `CORS_ORIGINS` | `http://localhost:5174,http://localhost:5175` | Vesszővel elválasztott lista a böngészőből hívó originokról (landing, admin). |
| `KEYCLOAK_BASE_URL` | `http://localhost:8081` | A Keycloak gyökere az Admin REST API-hoz. Dockerben: `http://keycloak:8081`. |
| `KEYCLOAK_REALM` | `snitt` | A realm neve az Admin API útvonalakhoz. |
| `KEYCLOAK_ADMIN_CLIENT_ID` | `snitt-admin-api` | Bizalmas, service accountos kliens a felhasználók olvasásához. |
| `KEYCLOAK_ADMIN_CLIENT_SECRET` | `snitt-admin-api-dev-secret` | Fejlesztői titok; élesben kötelezően felülírandó. |
| `ADMIN_ROLE` | `admin` | Az a realm szerep, ami az `/api/v1/admin/*` végpontokat nyitja. |
| `DEFAULT_CURRENCY` | `HUF` | Alapértelmezett pénznem (ISO 4217). |

A service accountnak a `realm-management` kliensen `view-users` (és a lapozáshoz `query-users`)
szerep kell; a [`keycloak/realm-export.json`](../../keycloak/realm-export.json) ezt már
tartalmazza.

## Végpontok

Minden `/api/v1/...` végpont `Authorization: Bearer <access_token>` fejlécet vár; hiányzó vagy
érvénytelen token esetén a válasz `401`.

### `GET /healthz`

Hitelesítés nélkül hívható. Rövid időkorláttal megpingeli az adatbázist.
`200` `{"status":"ok","database":"up"}`, vagy `503`, ha az adatbázis nem elérhető.

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
Jogosultságot egyelőre nem ad ki API – kézzel, SQL-lel kerül a táblába.

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

Az `admin_audit` az account szolgáltatás adatbázisában van, tehát a
[`deploy/scripts/backup.sh`](../../deploy/scripts/backup.sh) menti.

## Felépítés

```
cmd/server/main.go        indítás, graceful shutdown
internal/config           env-alapú konfiguráció
internal/auth             OIDC token-ellenőrzés + middleware (TokenVerifier interfész)
internal/store            pgxpool, beágyazott schema.sql, lekérdezések
internal/httpapi          routing, handlerek, CORS, admin napló, tesztek
```
