# Snitt — landing

A Snitt asztali alkalmazás nyilvános bemutatkozó oldala. Vite + React 18 + TypeScript,
`react-router-dom` az útvonalakhoz, sima CSS (`src/styles.css`) CSS-változókkal — se Tailwind, se
UI könyvtár, se CSS-in-JS.

Két útvonal van:

- `/` — a marketing oldal (hero, hogyan működik, funkciók, letöltés, fiók, GYIK).
- `/profil` — védett oldal: a bejelentkezett felhasználó neve és e-mail címe a tokenből, link a
  Keycloak fiókkonzolra, és kijelentkezés gomb. Bejelentkezés nélkül átirányít a Keycloakra.

## Fejlesztés

```bash
npm install
npm run dev        # http://localhost:5174
```

További parancsok:

```bash
npm run typecheck  # tsc --noEmit
npm run build      # typecheck + produkciós build a dist/ mappába
npm run preview    # a legyártott build kiszolgálása
```

Node 20 szükséges.

## Környezeti változók

Másold le a `.env.example` fájlt `.env` néven, ha az alapértelmezéseken változtatni kell.

| Változó                   | Alapértelmezés          | Mire való                                    |
| ------------------------- | ----------------------- | -------------------------------------------- |
| `VITE_KEYCLOAK_URL`       | `http://localhost:8081` | A Keycloak szerver gyökér URL-je             |
| `VITE_KEYCLOAK_REALM`     | `snitt`              | A realm neve                                 |
| `VITE_KEYCLOAK_CLIENT_ID` | `snitt-landing`      | A landing oldalhoz tartozó public client neve |

A Vite csak build- és dev-időben olvassa ezeket, tehát környezetváltáskor újra kell buildelni.

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

A Keycloak oldalán a `snitt-landing` clientnél be kell állítani:

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
