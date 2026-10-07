# snitt-cloud

A [Snitt](https://github.com/lipcsei/snitt) asztali alkalmazás **opcionális felhő oldala**: a
nyilvános landing (`snitt.video`), a fiókokat és előfizetéseket nyilvántartó account szolgáltatás,
és a hozzá tartozó belső admin felület.

A desktop app ettől függetlenül, fiók és hálózat nélkül is teljes értékűen működik; videó vagy
átirat soha nem kerül ide. A miértekről a [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) szól.

## Mi van benne

| Mappa | Mi | Részletek |
|---|---|---|
| [`landing/`](landing) | A marketing oldal és a `/profil` (React + Vite) | [`landing/README.md`](landing/README.md) |
| [`admin/`](admin) | Belső admin: felhasználók, előfizetések, számlázás (React + Vite) | [`admin/README.md`](admin/README.md) |
| [`services/account/`](services/account) | Az egyetlen saját backend (Go, Postgres): profil, entitlementek, előfizetések, számlák, admin napló | [`services/account/README.md`](services/account/README.md) |
| [`deploy/`](deploy) | A fejlesztői és az éles Docker Compose | |
| [`docs/`](docs) | Architektúra, VPS-telepítés, élesítési ellenőrzőlista | lásd lent |
| `.commons/` | A közös [`commons`](https://github.com/lipcsei/commons) repó submodule-ként (mentő szkriptek) | |

A bejelentkezést **nem** ez a repó adja: a közös Keycloak a [`sso`](https://github.com/lipcsei/sso)
repóban fut, ebben a `snitt` realm tartozik ide.

## Indítás

Kell hozzá Docker, és a közös Keycloaknak **előbb** futnia kell (ő hozza létre az `sso-net`
hálózatot):

```bash
make -C ../sso up     # a közös Keycloak (http://localhost:8081)
git submodule update --init   # a .commons, első alkalommal
make up               # Postgres + account + admin + landing
```

A build a privát `commons` Go modult SSH-n éri el, ezért olyan SSH-ügynök kell, ami hozzáfér a
GitHubhoz (`ssh-add -l`).

| Mi | Cím |
|---|---|
| Account API | <http://localhost:8090> |
| Admin | <http://localhost:5175> |
| Landing (a valódi production build, nginx mögül) | <http://localhost:8088> |
| Keycloak (az `sso` repóból) | <http://localhost:8081> |
| Postgres | `localhost:5432` |

A fejlesztői fiókok a `snitt` realm importjából jönnek, lásd az `sso` repó `realms/snitt-realm.json`
fájlját.

Konténer nélkül, gyors újratöltéssel:

```bash
make install        # a landing és az admin függőségei
make dev-account    # az account helyben (:8090), a compose konténere helyett
make dev-landing    # Vite: http://localhost:5174
make dev-admin      # Vite: http://localhost:5175, a compose konténere helyett
```

A többi parancsot a `make help` sorolja fel.

## Ellenőrzés

```bash
make check          # lint + teszt + build + statikus elemzés: nagyjából amit a CI futtat
make test           # csak az account tesztjei (adatbázis és Keycloak nélkül futnak)
make compose-check  # a fejlesztői és az éles compose érvényes-e
make static-vuln    # sebezhetőség-keresés (külön, mert lassú)
```

A CI (`.github/workflows/ci.yml`) négy külön jobban futtatja ugyanezt: az account szolgáltatás
(build, vet, teszt, gofmt), a landing és az admin (eslint, `tsc`, build), valamint a `make static`.
A sebezhetőség-keresés nem a CI része, hanem naponta fut (`vuln.yml`), hogy egy frissen bejelentett
CVE ne állítsa meg a kitelepítést. A privát `commons` repót a CI a `COMMONS_DEPLOY_KEY` repository
secrettel éri el; a GitHub-beállítások listája a [`docs/VPS-TELEPITES.md`](docs/VPS-TELEPITES.md)
7. pontjában van.

## Kitelepítés

Két helyre megy, mindkettő a `main` ágról, a zöld CI után magától:

- **A landing a GitHub Pages-re** (`snitt.video`): a `ci.yml` utolsó lépése.
- **Az account, az admin és a Postgres a közös VPS-re**: a `Deploy VPS` workflow
  (`.github/workflows/deploy.yml`). Élesben az [`edge`](https://github.com/lipcsei/edge) proxy éri el
  őket az `edge-net` hálózaton, `snitt-account` és `snitt-admin` néven (`api.snitt.video`,
  `admin.snitt.video`).

Az éles beállítások a `snitt-cloud-prod` GitHub Environmentben vannak; a `VPS_ENV` secretből lesz a
szerveren a `deploy/.env.prod` (sablon: `deploy/.env.prod.example`). Feltölteni a
[`vps`](https://github.com/lipcsei/vps) repóból kell, nem kézzel.

Egy üres szerver előkészítése lépésről lépésre: [`docs/VPS-TELEPITES.md`](docs/VPS-TELEPITES.md).

## Mentés és visszaállítás

```bash
make backup BACKUP_DIR=/mentesek/helye
make restore FILE=/mentesek/snitt-….sql.gz     # FELÜLÍRJA az adatbázist
```

Élesben az env-fájlt is meg kell adni: `make backup DOTENV=deploy/.env.prod BACKUP_DIR=…`. A
szkriptek a `.commons` submodule-ból jönnek, ezért a szerveren is inicializálni kell
(`git submodule update --init`), különben nincs mit futtatni.

## Dokumentáció

- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md): a desktop app és a felhő oldal viszonya, az adatok
  útja, és hogy miért ilyen kicsi a backend.
- [`docs/VPS-TELEPITES.md`](docs/VPS-TELEPITES.md): az első felállítás egy üres szerveren.
- [`docs/FIOKOK-ELESITES.md`](docs/FIOKOK-ELESITES.md): mit kell átállítani a Keycloakban az első
  valódi felhasználó előtt.
