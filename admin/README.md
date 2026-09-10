# Snitt admin

Belső admin felület a Snitt **felhő oldalához**: felhasználók, előfizetések és számlázás egy
helyen. Nem publikus oldal – a hozzáférést a Keycloak `admin` realm szerepe dönti el, és az
account szolgáltatás minden kérésnél ellenőrzi.

Ez a felület **nem** kezeli a desktop appot és nem lát bele a felhasználók videóiba. A Snitt
desktop app fiók nélkül, teljesen önállóan működik – lásd
[`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md).

## Futtatás

Kell hozzá a futó fejlesztői stack (Postgres + Keycloak + account szolgáltatás). A repo
gyökeréből:

```bash
docker compose -f deploy/docker-compose.yml up --build
```

Majd külön terminálban:

```bash
cd admin
cp .env.example .env      # majd szerkeszd, ha kell
npm install
npm run dev               # http://localhost:5175
```

A stackben előre be van importálva egy admin fiók:

| Felhasználó | Jelszó | Szerep |
| --- | --- | --- |
| `snittadmin` | `snittadmin1234` | `admin` realm szerep |

A `demo` / `demo12345678` fiókkal is be lehet lépni, de az nem admin: ilyenkor a felület a
„Nincs jogosultságod” képernyőt mutatja, nem egy félig működő táblázatot.

> A jelszavak azért ilyen hosszúak, mert a realm jelszóházirendje legalább 10 karaktert kér –
> a rövidebbekkel az importálás el sem indulna. Éles környezetben ezeknek a fiókoknak nincs
> helyük: lásd [`docs/FIOKOK-ELESITES.md`](../docs/FIOKOK-ELESITES.md).

### Ellenőrzések

```bash
npx tsc --noEmit
npm run build
```

## Környezeti változók

| Változó | Alapértelmezés | Leírás |
| --- | --- | --- |
| `VITE_KEYCLOAK_URL` | `http://localhost:8081` | A Keycloak gyökér URL-je. |
| `VITE_KEYCLOAK_REALM` | `snitt` | A realm neve. |
| `VITE_KEYCLOAK_CLIENT_ID` | `snitt-admin` | Az admin felület publikus kliense (PKCE). |
| `VITE_API_BASE_URL` | `http://localhost:8090` | Az account szolgáltatás gyökér URL-je. |

## Jogosultság

A böngésző oldali ellenőrzés csak kényelmi: a felület megnézi a token `realm_access.roles`
mezőjét, és `admin` szerep nélkül a „Nincs jogosultságod” képernyőt mutatja. Az **érdemi**
ellenőrzés a szerveren van: az `/api/v1/admin/*` végpontok szerep nélkül `403`-at adnak, akkor
is, ha valaki közvetlenül hívja őket.

Külön admin jelszó vagy API kulcs nincs: ugyanaz a Keycloak token azonosít, mint a landingen,
csak a szerep más.

## Képernyők

- **Vezérlőpult** – regisztrált felhasználók, felhő profilok, aktív előfizetések csomagonként,
  MRR, 30 napos növekedés, nyitott számlák darabszáma és összege. Ha a Keycloak nem elérhető,
  a felhasználószám helyén `–` áll, és egy sáv megmondja, miért – a többi szám marad.
- **Felhasználók** – kereshető (e-mail, felhasználónév, név), lapozott lista. Az adatlapon a
  Keycloak identitás, a helyi profil, a jogosultságok, az előfizetések és a számlák együtt
  látszanak, a műveletekkel egy helyen.
- **Előfizetések** – csomag és állapot szerint szűrhető lista, kiadás / csomagváltás /
  lemondás / visszakapcsolás.
- **Számlázás** – állapot és felhasználó szerinti szűrés, szűrésre vetített összesítők
  (nyitott, kifizetve, összesen), kifizetettre jelölés és sztornó.

## Pénzt érintő műveletek

Ez pénzügyi belső eszköz, ezért:

- minden visszavonhatatlan művelet (lemondás, kifizetettre jelölés, sztornó) megerősítő
  párbeszédablakon megy át, ami előre megmutatja a következményt;
- a sztornóhoz **kötelező** indoklás;
- a szerver hibáit a felület szó szerint megmutatja (pl. „csak nyitott számla sztornózható”),
  nem nyeli le őket;
- **valódi terhelés sehol nem történik.** A számla „kifizetve” állapota könyvelési jelölés:
  a pénznek máshol kell beérkeznie.

## A fizetési szolgáltató illesztési pontja

Ma nincs Stripe/Paddle integráció. A varrat a szerveren van:
`services/account/internal/billing` – a `Provider` interfész és a jelenlegi `Manual`
implementáció. A felület soha nem hív szolgáltató-specifikus végpontot, csak az admin API-t,
ezért egy valódi szolgáltató bekötése a felületet nem érinti (legfeljebb a „terhelés nem
történik” magyarázó szövegek tűnnek el).

## Üzemeltetés (Docker, VPS)

Az admin felület statikus SPA, amit nginx szolgál ki - fejlesztés közben a
compose stackben fut (`http://localhost:5175`), élesben ugyanez a kép megy a
VPS-re:

```bash
docker compose -f deploy/docker-compose.yml up -d --build admin
```

**Fontos**: a Vite a `VITE_*` értékeket **fordításkor** helyettesíti be, tehát
ezek nem állíthatók a konténer indításakor - más környezethez újra kell
építeni a képet:

```bash
docker build -t snitt-admin:latest \
  --build-arg VITE_KEYCLOAK_URL=https://auth.snitt.video \
  --build-arg VITE_KEYCLOAK_REALM=snitt \
  --build-arg VITE_KEYCLOAK_CLIENT_ID=snitt-admin \
  --build-arg VITE_API_BASE_URL=https://api.snitt.video \
  ./admin
```

A VPS-en érdemes még:

- **HTTPS-t tenni elé** (Caddy/Traefik/nginx reverse proxy) - a Keycloak
  bejelentkezés éles környezetben HTTPS-t vár,
- a Keycloak `snitt-admin` kliensénél a valódi domainre állítani a
  `redirectUris` és `webOrigins` mezőket,
- **hozzáférést korlátozni** (IP-szűrés vagy VPN): ez belső eszköz, amiben
  felhasználói adatok és számlázás van. A Keycloak `admin` szerep önmagában is
  véd, de a felület nyilvános kitettségét nincs miért vállalni.

## Felépítés

```
src/auth.ts             Keycloak konfiguráció, szerep-olvasás a tokenből
src/AuthProvider.tsx    bejelentkezés, token-frissítés, jogosultsági állapot
src/api.ts              tipizált admin API kliens, ApiError a szerver üzenetével
src/hooks.ts            useAsync / useDebounced / usePlans
src/format.ts           pénz-, dátum- és státuszformázás (hu-HU)
src/components/         elrendezés, tábla-kiegészítők, modálisok, műveletek
src/pages/              Vezérlőpult, Felhasználók, Adatlap, Előfizetések, Számlázás
src/styles.css          a teljes stíluslap, CSS változókkal
```
