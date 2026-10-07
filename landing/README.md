# Snitt — landing

A Snitt asztali alkalmazás nyilvános bemutatkozó oldala. Vite + React 18 + TypeScript,
`react-router-dom` az útvonalakhoz, sima CSS (`src/styles.css`) CSS-változókkal — se Tailwind, se
UI könyvtár, se CSS-in-JS.

Három nyelven fut (magyar, angol, német); a nyelv kizárólag az URL-ből derül ki
(`src/i18n/index.tsx`, `PATHS`). Az útvonalak:

| Oldal | hu | en | de |
| --- | --- | --- | --- |
| A marketing oldal (hero, hogyan működik, funkciók, letöltés, fiók, GYIK) | `/` | `/en` | `/de` |
| Telepítési útmutató | `/telepites` | `/en/install` | `/de/installation` |
| Profil (védett) | `/profil` | `/en/profile` | `/de/profil` |

- A profil a bejelentkezett felhasználó nevét és e-mail címét mutatja a tokenből, linkkel a
  Keycloak fiókkonzolra és kijelentkezés gombbal. Bejelentkezés nélkül átirányít a Keycloakra.
- `/adatvedelem` és `/aszf` — a jogi oldalak, nyelv-előtag nélkül (egyelőre csak magyarul).
- Ismeretlen útvonal az adott nyelv főoldalára irányít.

## Fejlesztés

```bash
npm install
npm run dev        # http://localhost:5174
```

A repo gyökeréből ugyanez: `make install`, majd `make dev-landing`. A `make up` a valódi production
buildet is felhozza egy nginx konténerben a <http://localhost:8088> címen.

További parancsok:

```bash
npm run lint       # eslint
npm run typecheck  # tsc --noEmit
npm run build      # typecheck + produkciós build a dist/ mappába + útvonalankénti meta-fájlok
npm run preview    # a legyártott build kiszolgálása
npm run og         # a megosztási képek (og:image) újragyártása (a gyökérből: make og)
```

A build utolsó lépése (`scripts/prerender-meta.mjs`) útvonalanként külön HTML-t ír a helyes nyelvi
metaadatokkal, mert a közösségi oldalak előnézet-botjai nem futtatnak JavaScriptet.

Node 20 szükséges.

## Letöltés-gombok

A telepítők a Snitt saját szerveréről töltődnek le (`https://api.snitt.video/downloads`, az edge
proxy szolgálja ki). A gombok a `latest.json`-ból (`src/release.ts`) tudják a verziót, a méretet és a
fájlnevet; amíg nincs kiadás vagy a szerver nem érhető el, a „hamarosan” üzenetet mutatják.
Felülírható: `VITE_DOWNLOAD_BASE_URL` (a CI-ban a `DOWNLOAD_BASE_URL` repository variable). A
kiadás menete: a `snitt` repo `docs/KIADAS.md`-je.

## Környezeti változók

Másold le a `.env.example` fájlt `.env` néven, ha az alapértelmezéseken változtatni kell.

| Változó                   | A `.env.example` értéke | Mire való                                    |
| ------------------------- | ----------------------- | -------------------------------------------- |
| `VITE_KEYCLOAK_URL`       | `http://localhost:8081` | A Keycloak szerver gyökér URL-je             |
| `VITE_KEYCLOAK_REALM`     | `snitt`              | A realm neve                                 |
| `VITE_KEYCLOAK_CLIENT_ID` | `snitt-landing`      | A landing oldalhoz tartozó public client neve |
| `VITE_DOWNLOAD_BASE_URL`  | *(üres)*                | A telepítők letöltési helye; üresen `https://api.snitt.video/downloads` |
| `VITE_SENTRY_DSN`         | *(üres)*                | Opcionális hibajelentés a GlitchTipbe; üresen teljesen kikapcsolva |
| `VITE_SENTRY_ENVIRONMENT` | a Vite módja            | A környezet neve az eseményeken              |
| `VITE_GA_MEASUREMENT_ID`  | *(üres)*                | Google Analytics 4 mérési azonosító; üresen az analitika ki van kapcsolva, és a cookie-sáv sem jelenik meg |

A három Keycloak-változónak **a kódban nincs alapértéke** (`src/auth.ts`): ha bármelyik üres vagy
hiányzik, a bejelentkezés ki van kapcsolva, és a `keycloak-js` el sem indul. A fenti értékeket a
`.env.example` és a Docker-kép build-argumentumai adják; `.env` nélkül a `npm run dev` fiókok nélkül
fut.

A build még ezeket olvassa (a `.env.example`-ben nem szerepelnek):
`VITE_BASE` (alapútvonal, ha a GitHub Pages alkönyvtárban szolgál ki; alapból `/`), `SITE_URL` (a
kanonikus és megosztási URL-ek origója; alapból `https://snitt.video`), valamint az opcionális
`FB_APP_ID` és `TWITTER_SITE` a `scripts/prerender-meta.mjs`-ben.

A Vite csak build- és dev-időben olvassa ezeket, tehát környezetváltáskor újra kell buildelni.

Az éles (GitHub Pages) build a `ci.yml` `deploy-landing` jobjában a repó *repository variable*-jaiból
kapja őket: `KEYCLOAK_URL`, `KEYCLOAK_REALM`, `KEYCLOAK_CLIENT_ID`, `DOWNLOAD_BASE_URL`,
`SENTRY_DSN_LANDING`, és a `PAGES_CNAME` (a saját domain; ebből lesz a `SITE_URL`, a `VITE_BASE=/` és
a `dist/CNAME`). A `VITE_GA_MEASUREMENT_ID`-t, az `FB_APP_ID`-t és a `TWITTER_SITE`-ot a `ci.yml`
jelenleg nem adja át.

A hibajelentés (`src/sentry.tsx`) nyilvános oldalon fut, ezért az SDK nem része a fő csomagnak:
csak beállított `VITE_SENTRY_DSN` mellett, külön csomagként töltődik be (üresen a csomag el sem
készül). Élesben a GitHub Pages build a `SENTRY_DSN_LANDING` repository variable-ből kapja
(`ci.yml`); részletek: [`docs/VPS-TELEPITES.md`](../docs/VPS-TELEPITES.md) 11. pont.

## Auth és Keycloak

A bejelentkezés `keycloak-js`-sel megy (`src/auth.ts` a konfiguráció, `src/AuthProvider.tsx` a
`useAuth()` context).

Fontos, hogy **az auth itt opcionális ráadás**: az oldal betöltéskor nem irányít át a Keycloakra.
Az init `check-sso` módban, PKCE-vel (`S256`) fut, és rejtett iframe-ben
(`public/silent-check-sso.html`) nézi meg, van-e élő session. Ha a Keycloak nem érhető el, az init
elhasal vagy timeoutol — a marketing oldal ettől ugyanúgy megjelenik, csak a bejelentkezés gomb
nem visz sehová.

A `useAuth()` a következőket adja: `authenticated`, `profile` (`name`, `email`), `login()`,
`register()` (a Keycloak regisztrációs belépőpontja), `logout()`, `accountUrl` (fiókkonzol), plusz
egy `ready` jelzés arra, hogy a session-ellenőrzés lefutott-e.

A bejelentkezett, de még meg nem erősített e-mail-című felhasználónak egy elrejthető sáv szól
(`src/features/emailVerify`); a belépést ez nem akadályozza.

A Keycloak oldalán a `snitt-landing` clientnél be kell állítani (a fejlesztői `snitt` realm, az `sso`
repó `realms/snitt-realm.json` fájlja ezt már tartalmazza):

- standard flow engedélyezve, public client (nincs secret),
- valid redirect URI: `http://localhost:5174/*`,
- valid post logout redirect URI: `http://localhost:5174/*`,
- web origin: `http://localhost:5174`.

A regisztráció csak akkor jelenik meg, ha a realmben be van kapcsolva a *User registration*.

> A weboldali fiók a készülő extra funkciókhoz kell. Az asztali alkalmazás fiók nélkül,
> önállóan is teljes értékű — az oldal szövege is ezt mondja, ne írjuk át másra.

## Kiszolgálás

Az alkalmazás SPA, kliensoldali útvonalakkal. Éles kiszolgálásnál minden ismeretlen útvonalat az
`index.html`-re kell irányítani, különben a `/profil` közvetlen megnyitása (és a Keycloak
visszairányítása oda) 404-et ad.

- **Élesben a GitHub Pages** szolgálja ki (`snitt.video`): a `ci.yml` `deploy-landing` jobja a `main`
  ágra kerülő, zöld build után telepíti. Az ismeretlen útvonalakhoz az `index.html`-t `404.html`
  néven is kiteszi.
- **Helyben, konténerben** (`make up`, <http://localhost:8088>) nginx szolgálja ki ugyanezt a buildet
  (`nginx.conf`, a konténerben a 8080-as porton). A VPS-re a landing nem települ.
