# VPS-telepítés – egy üres szerver előkészítése

Ez a lista egy vadonatúj, semmivel fel nem installált szerveren (Ubuntu/Debian
feltételezve) viszi végig a Snitt felhő oldalának (`account` + `admin` +
Postgres) első felállítását; a TLS-t a megosztott `edge` proxy adja (lásd
4c. pont). A cél állapot:
`git push origin main` után a GitHub Actions automatikusan kitelepít –
lásd [`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml).

A `landing` NEM ide megy – az GitHub Pages-en fut, lásd
[`ci.yml`](../.github/workflows/ci.yml). Ez a gép csak a másik három
komponenst szolgálja ki.

**Keycloak SEM ide megy.** Megosztott szolgáltatás, külön repóban:
[`lipcsei/sso`](https://github.com/lipcsei/sso) – *ugyanezen* a szerveren
fut, de saját compose projektként, saját ütemben telepítve (a saját
`Deploy VPS` workflow-jával – lásd annak README-jét).

**A TLS-terminátor (Caddy) SEM ide megy.** A 80/443-as portot egy hoston
egyszerre csak egy folyamat foghatja le, ezért egyetlen megosztott proxy
kezeli minden app helyett: [`lipcsei/edge`](https://github.com/lipcsei/edge).
Az köti az `auth.snitt.video`, `api.snitt.video`, `admin.snitt.video` (és a
breath, a gombamester meg a glitchtip) címeit a megfelelő konténerekhez két
megosztott Docker hálózaton keresztül (`sso-net`, `edge-net`). Ez a stack csak az `account` és `admin`
konténert csatlakoztatja az `edge-net`-re, egyedi aliassal.

**Indítási sorrend a VPS-en: `sso` (4b) → `edge` (4c) → ez (8).** Az első
kettő hozza létre a hálózatokat, amiket a mi konténereink használnak.

**A titkok (jelszavak, kliens secret) egyetlen forrása a GitHub repo
secretjei, nem a szerver.** A `deploy/.env.prod` fájlt maga a deploy workflow
írja fel a szerverre minden kitelepítéskor a `VPS_ENV` secretből – a
szerveren SOSEM szerkeszted kézzel; ha egy jelszót cserélnél, a GitHub
secretet frissíted és újrafuttatod a workflow-t.

**A secreteket és a változókat nem kell kézzel kattintgatni.** A
[`lipcsei/vps`](https://github.com/lipcsei/vps) repó `github-env.sh` szkriptje
tölti fel őket mind a hat stack (`sso`, `edge`, `snitt-cloud`, `breath`,
`gombamester`, `glitchtip`) `<repó>-prod` environmentjébe a saját géped
`env/<repó>.env` fájljaiból, és feltöltés előtt ellenőrzi is őket (lásd 6–7.
pont). A lenti kézi út ugyanazt az eredményt adja.

## 0. Mire lesz szükséged előre

- A szerver IP-je, és root (vagy sudo-képes) SSH-hozzáférés hozzá.
- Három domain/aldomain, amit a DNS-szolgáltatódnál A-rekordként a szerver
  IP-jére tudsz állítani – pl. `auth.snitt.video`, `api.snitt.video`,
  `admin.snitt.video`. (A `snitt.video` gyökér marad a GitHub Pages-nél.)

## 1. Alap frissítés és tűzfal

```bash
ssh root@<szerver-ip>

apt update && apt upgrade -y

# Csak SSH, HTTP, HTTPS legyen nyitva kifelé - minden más (mindkét compose
# projekt Postgrese, a Keycloak, account, admin) a docker-compose.prod.yml
# fájlok miatt úgyis csak 127.0.0.1-re lesz kötve, a tűzfal ez ellen egy
# második védelmi vonal.
apt install -y ufw
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw --force enable
```

## 2. Külön deploy felhasználó (ne root-tal fusson a CI)

```bash
adduser --disabled-password --gecos "" deploy
usermod -aG sudo deploy

mkdir -p /home/deploy/.ssh
chmod 700 /home/deploy/.ssh
```

Generálj egy **külön, csak erre a célra szánt** SSH-kulcsot a saját gépeden
(ne a személyes kulcsodat használd a GitHub Actionshöz):

```bash
ssh-keygen -t ed25519 -C "snitt-cloud-deploy" -f ~/.ssh/snitt_deploy_key -N ""
```

A publikus kulcsot másold fel a szerverre a `deploy` felhasználóhoz:

```bash
cat ~/.ssh/snitt_deploy_key.pub | ssh root@<szerver-ip> \
  "cat >> /home/deploy/.ssh/authorized_keys && chmod 600 /home/deploy/.ssh/authorized_keys && chown -R deploy:deploy /home/deploy/.ssh"
```

Ellenőrizd, hogy ezzel be tudsz-e lépni:

```bash
ssh -i ~/.ssh/snitt_deploy_key deploy@<szerver-ip>
```

A **privát** kulcs (`~/.ssh/snitt_deploy_key`, kulcs eleje `-----BEGIN
OPENSSH PRIVATE KEY-----`) tartalma kerül a `VPS_SSH_KEY` GitHub secretbe
(lásd 7. pont). Ha a secreteket a `vps` repóból töltöd fel, a kulcsnak a
saját gépeden kell maradnia: a `vps.conf` `SSH_KEY_FILE`-ja erre a fájlra
mutat (`~/.ssh/snitt_deploy_key`), és minden `push` innen olvassa. Ha kézzel
vetted fel a secretet, a saját gépedről törölheted.

## 3. Docker telepítése

```bash
# root-ként vagy sudo-val:
curl -fsSL https://get.docker.com | sh
usermod -aG docker deploy
```

Lépj ki és jelentkezz be újra `deploy`-ként, hogy a docker csoporttagság
érvénybe lépjen, majd ellenőrizd:

```bash
ssh -i ~/.ssh/snitt_deploy_key deploy@<szerver-ip>
docker compose version
```

## 4. Repo klónozása

A GitHub Actions workflow `/opt/snitt-cloud`-ba vár – vagy ide klónozd, vagy
írd át az utat a `deploy.yml`-ben és ebben a leírásban egyszerre.

Mivel a repo privát, deploy kulcs kell hozzá – ezt **a szerveren** generáld,
és csak-olvasásra add hozzá a GitHubon (a repó Settings → Deploy keys
oldalán). A GitHub egy deploy kulcsot csak egy repóhoz enged felvenni, ezért
repónként saját kulcs és host-alias kell; a lenti `github.com-snitt-cloud` egy
ilyen alias a `deploy` felhasználó `~/.ssh/config` fájljában (`HostName
github.com`, `IdentityFile` a repó kulcsa, `IdentitiesOnly yes`). Utána:

```bash
sudo mkdir -p /opt/snitt-cloud
sudo chown deploy:deploy /opt/snitt-cloud
git clone git@github.com-snitt-cloud:lipcsei/snitt-cloud.git /opt/snitt-cloud
cd /opt/snitt-cloud
```

**Kell egy második kulcs is, a privát `commons` repóhoz.** A deploy workflow a
szerveren a `~/.ssh/commons_deploy_key` fájlból indít egy eldobható
SSH-ügynököt: ezzel húzza le a `.commons` submodule-t, és ezzel éri el a kép
építése a `github.com/lipcsei/commons` Go modult (`--ssh default`). Enélkül a
kitelepítés az `ssh-add` lépésnél megáll. Generáld a szerveren, pontosan ezen
a néven, jelszó nélkül, és a publikus felét vedd fel csak-olvasható deploy
kulcsként a `lipcsei/commons` repóban:

```bash
ssh-keygen -t ed25519 -f ~/.ssh/commons_deploy_key -N "" -C "vps commons"
cat ~/.ssh/commons_deploy_key.pub
```

## 4b. A megosztott Keycloak (`sso` repó) felállítása

Az **első** felállítást kézzel kell elvégezni; utána az `sso` repó saját
`Deploy VPS` GitHub Actions workflow-ja (environment: `sso-prod`, ugyanazok a
secretek, mint itt) frissíti minden zöld `main` után – lásd annak README-jét.
Az első indítást **mielőtt** a 8. pontban először kitelepíted a snitt-cloud
stacket (az `sso` stack hozza létre az `sso-net` hálózatot, amihez a mi
konténereink és az `edge` proxy csatlakozik).

```bash
sudo mkdir -p /opt/sso
sudo chown deploy:deploy /opt/sso
git clone git@github.com:lipcsei/sso.git /opt/sso
cd /opt/sso

cp .env.prod.example .env.prod
# szerkeszd: POSTGRES_PASSWORD, KEYCLOAK_ADMIN_PASSWORD (mindkettő saját,
# a snitt-cloud .env.prod-jában lévőktől ELTÉRŐ generált jelszó legyen),
# KEYCLOAK_PUBLIC_URL=https://auth.snitt.video, és a többi app realmjének
# címei és kliens-titkai (BREATH_*, GOMBAMESTER_*, GLITCHTIP_*) – a sablon
# megjegyzései mindegyiket elmagyarázzák
chmod 600 .env.prod

# Az éles realm-fájlokat a fejlesztői realm-fájlokból GENERÁLJA (a dev
# fájlokban ismert jelszavú felhasználók vannak, azok élesben nem mehetnek
# ki) – a compose enélkül el sem indul. A generátor a beállításait a
# .env.prod-ból olvassa, és hibával megáll, ha egy megerősített realm
# (breath, gombamester, glitchtip) https:// címe vagy generált kliens-titka
# hiányzik – akkor is, ha az az app még nincs kint.
./scripts/make-prod-realms.sh

docker compose --env-file .env.prod \
  -f docker-compose.yml -f docker-compose.prod.yml \
  up -d --build --remove-orphans
```

Ugyanezt a `.env.prod`-ot kell az `sso-prod` environment `VPS_ENV` secretjébe
is feltölteni (a `vps` repóban: `env/sso.env`): az `sso` saját `Deploy VPS`
workflow-ja minden futáskor azzal **írja felül** a szerveren lévőt, majd
újragenerálja a realmeket.

Ez hozza létre az `sso-net` Docker hálózatot is, amihez az `edge` proxy és a
snitt-cloud `account` szolgáltatása csatlakozik – ezért kell ennek a
lépésnek megelőznie a 4c. és a 8. pontot.

A generátor **figyelmeztet**, ha egy változatlanul másolt realm (pl. a `snitt`)
még ismert jelszavú fejlesztői felhasználókat tartalmaz – friss adatbázison
ezek importálódnak. Lásd `docs/FIOKOK-ELESITES.md` 5. pont: az első
indítás után töröld őket.

Ha a Keycloakot később, más appok bevonása miatt (pl. új realm hozzáadása)
frissíteni kell, azt az `sso` repó `main` ágára kerülő commit után a
`Deploy VPS` workflow elvégzi. Kézzel csak végszükségben: `cd /opt/sso &&
git pull && docker compose --env-file .env.prod -f docker-compose.yml -f
docker-compose.prod.yml up -d --build`.

## 4c. A megosztott proxy (`edge` repó) felállítása

Ez kezeli a 80/443-at és a TLS-tanúsítványokat mindenki helyett. Az első
indítást kézzel végzed, a 4b. pont UTÁN, de a 8. pont ELŐTT; utána az `edge`
repó saját `Deploy VPS` workflow-ja (environment: `edge-prod`) frissíti – ami
a Caddyfile-t a futó proxy érintése ELŐTT érvényesíti.

```bash
sudo mkdir -p /opt/edge
sudo chown deploy:deploy /opt/edge
git clone git@github.com:lipcsei/edge.git /opt/edge
cd /opt/edge

cp .env.prod.example .env.prod
# szerkeszd: ACME_EMAIL, a PUBLIC_*_URL címek és a *_REDIRECT_FROM listák (a
# DNS-t ELŐBB állítsd be, lásd 5. pont – enélkül a Let's Encrypt
# tanúsítványkérés elbukik). MINDEN változó kötelező, a többi appé is: ha egy
# hiányzik, a compose el sem indul.
chmod 600 .env.prod

# A telepítők mappája (api.snitt.video/downloads): léteznie kell az első
# indítás ELŐTT, különben a Docker üres, root tulajdonú mappaként hozza létre.
# Az egyszeri beállítása a snitt repó docs/KIADAS.md-jében van.
ls -ld /srv/snitt-downloads

docker compose --env-file .env.prod up -d
```

Ez hozza létre az `edge-net` hálózatot. A proxy akkor is elindul, ha az appok
még nem futnak – a címük addig `502`-t ad. Ezt a `.env.prod`-ot is a `VPS_ENV`
secret írja felül minden kitelepítéskor (`edge-prod` environment, a `vps`
repóban `env/edge.env`).

**Ha a szerveren korábban a snitt-cloud saját Caddyje futott**, az még foglalja
a 80/443-at: az átállás lépéseit (rövid kiesés) az
[`edge` README-je](https://github.com/lipcsei/edge#migrating-from-a-per-app-caddy-snitt-cloud-used-to-run-its-own)
írja le. Ez már csak történeti: ebben a repóban nincs saját Caddy, új
szerveren ezzel nincs teendő.

## 5. DNS

Állítsd be a DNS A-rekordokat (`auth`, `api`, `admin` a szerver IP-jére), és
várd meg, míg terjednek (`dig auth.snitt.video` mutassa a helyes IP-t) –
enélkül a 8. pontban a Let's Encrypt tanúsítványkérés elbukik.

## 6. `.env.prod` tartalmának előkészítése (a SAJÁT gépeden, nem a szerveren)

**A `vps` repóval (ajánlott):** a `../vps` mappában a `./github-env.sh init
snitt-cloud` létrehozza az `env/snitt-cloud.env` fájlt a
`deploy/.env.prod.example`-ből. Azt töltsd ki az alábbiak szerint; az a fájl
marad a titkok egyetlen olvasható példánya (git-ignore-olt, legyen róla
mentés). A `./github-env.sh check snitt-cloud` hálózat nélkül ellenőrzi:
megvan-e és ki van-e töltve minden kulcs, nem maradt-e bent `CSERÉLD-LE`
helykitöltő, és a több repóban is szereplő értékek (pl. a
`PUBLIC_KEYCLOAK_URL`, `PUBLIC_API_URL`, `PUBLIC_ADMIN_URL`) egyeznek-e az
`sso` és az `edge` env-fájljával.

**Kézzel:**

```bash
cp deploy/.env.prod.example /tmp/env.prod.draft
```

Töltsd ki `/tmp/env.prod.draft`-ot (vagy az `env/snitt-cloud.env`-et):
- `POSTGRES_PASSWORD` – generáld: `openssl rand -base64 24` (ez a snitt-cloud
  SAJÁT Postgresének jelszava – a 4b. pontban az `sso` repóhoz külön,
  MÁSIK generált jelszót adtál a Keycloak adatbázisának)
- a domainek (`PUBLIC_KEYCLOAK_URL` stb.) maradhatnak az alapértelmezésen, ha a `snitt.video`-t használod
- `KEYCLOAK_ADMIN_CLIENT_SECRET` egyelőre maradjon a fejlesztői értéken – ezt a 9. pontban, az ELSŐ felállás UTÁN cseréled, mert a kliens csak akkor jön létre (a realm importból, ami a 4b. pontban lefutott)

Ez a fájl a te géped `/tmp`-jében marad, git alá SOHA nem kerül – a teljes
tartalmát a következő pontban egy GitHub secretbe másolod, utána törölheted:
`rm /tmp/env.prod.draft`.

## 7. GitHub secretek és változók

**A `vps` repóval:** `./github-env.sh push snitt-cloud` (előtte `push
--dry-run` megmutatja, mit tenne). Létrehozza az environmentet, beállítja a
`VPS_HOST` / `VPS_USER` változót a `vps.conf`-ból, a `VPS_SSH_KEY` secretet a
`vps.conf` `SSH_KEY_FILE`-jából és a `VPS_ENV` secretet az
`env/snitt-cloud.env` teljes tartalmából. A `push --deploy snitt-cloud` utána
a `Deploy VPS` workflow-t is elindítja (ez a 8. pont).

**Kézzel:** a repo Settings → Environments → **`snitt-cloud-prod`** environment
alatt hozd létre (a deploy workflow ezt az environmentet nevezi meg;
repository szintű secretet nem használ):

| Hova | Név | Érték |
|---|---|---|
| változó vagy secret | `VPS_HOST` | a szerver IP-je vagy hostneve |
| változó vagy secret | `VPS_USER` | `deploy` |
| változó vagy secret | `VPS_SSH_PORT` | csak ha nem a 22-es |
| **csak secret** | `VPS_SSH_KEY` | a 2. pontban generált **privát** kulcs teljes tartalma |
| **csak secret** | `VPS_ENV` | a 6. pontban kitöltött `/tmp/env.prod.draft` fájl **teljes** tartalma, változtatás nélkül bemásolva |

A `VPS_SSH_KEY` és a `VPS_ENV` kerüljön az *Environment secrets* alá: a változót (*Environment variables*) a GitHub sima szövegként tárolja, és a futási naplóban sem takarja ki. Ha a workflow ezt a kettőt változóként találja, hibával leáll.

A **CI**-nak (nem a kitelepítésnek) külön, repository szintű beállítások kellenek; ezeket a `vps`
repó nem kezeli:

| Hova | Név | Mire |
|---|---|---|
| Actions secret | `COMMONS_DEPLOY_KEY` | Csak-olvasható deploy kulcs (privát fele) a `lipcsei/commons` repóhoz: a CI ezzel húzza le a `.commons` submodule-t és a privát Go modult. Enélkül az `account-service` és a `static` job (és a napi sebezhetőség-keresés) elbukik, és zöld CI híján a kitelepítés sem indul. |
| Dependabot secret | `COMMONS_READ_TOKEN` | Csak a `commons` repót olvasó token, hogy a Dependabot Go-frissítései ne akadjanak el (`.github/dependabot.yml`). |
| Actions variable | `KEYCLOAK_URL`, `KEYCLOAK_REALM`, `KEYCLOAK_CLIENT_ID` | A landing GitHub Pages buildjének Keycloak-beállításai. Üresen a landing fiókok nélkül épül. |
| Actions variable | `PAGES_CNAME` | A landing saját domainje (`snitt.video`); üresen a `*.github.io` alkönyvtárba épül. |
| Actions variable | `DOWNLOAD_BASE_URL`, `SENTRY_DSN_LANDING` | Opcionális: a telepítők helye, illetve a landing hibajelentése. |

## 8. Első kitelepítés

Ellenőrizd, hogy a 4b. pont `sso` stackje már fut (`cd /opt/sso && docker
compose --env-file .env.prod -f docker-compose.yml -f docker-compose.prod.yml
ps` – `keycloak` legyen `Up`; az `--env-file` nélkül a compose a kötelező
jelszavak híján nem renderel), és a 4c. pont `edge` stackje is (az hozza létre
az `edge-net` hálózatot), utána:

GitHub → repo → **Actions → Deploy VPS → Run workflow** (a `workflow_dispatch`
kézi indítás pont erre való – nem vár CI-futásra, a `main` ág meglévő kódját
viszi ki). Ez felírja a szerverre a `deploy/.env.prod`-ot a secretből,
`git reset --hard origin/main`-nel frissíti a klónt, lehúzza a `.commons`
submodule-t, majd a szerveren megépíti és elindítja a snitt-cloud stacket
(`postgres`, `account`, `admin`).

Kövesd a futást az Actions fülön; ha piros lesz, a log megmondja, melyik
lépésnél (a beállítások ellenőrzése, SCP feltöltés vagy SSH-s `docker
compose`) akadt el. Ha egy secret hiányzik, az első lépés név szerint megmondja,
melyik. (Ez a workflow – a közös `lipcsei/workflows` változattal szemben – nem
várja meg, hogy a konténerek healthy állapotba kerüljenek: a zöld futás csak
annyit jelent, hogy a `docker compose up` lefutott.)

Ellenőrzés a szerveren:

```bash
cd /opt/snitt-cloud
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml ps
# a proxy naplója (tanúsítványkérés, 502-k) az edge stackben van; az
# --env-file kell, mert ott minden változó kötelező:
(cd /opt/edge && docker compose --env-file .env.prod logs -f caddy)
```

A Let's Encrypt tanúsítványokat az `edge` proxy kéri (4c. pont) – ez néhány
másodperc, ha a DNS már terjedt. Ha "no such host"
vagy "connection refused" hibát látsz a logban, a DNS még nem ért oda, vagy
a 80/443 port nincs nyitva kifelé.

Nyisd meg böngészőben:
- `https://auth.snitt.video` – Keycloak bejelentkező oldal
- `https://admin.snitt.video` – admin felület

## 9. Élesítési lépések (kötelező, mielőtt igazi felhasználó regisztrál)

Kövesd végig [`docs/FIOKOK-ELESITES.md`](FIOKOK-ELESITES.md)-t: SMTP, e-mail
igazolás, a Keycloak kliensek valódi redirect URI-jai, a fejlesztői
`snittadmin`/`demo` fiókok törlése, a bootstrap admin jelszó cseréje.

A `manage-users` szerep (a fiók letiltásához kell) a mostani
`snitt-realm.json`-ban már benne van, tehát **friss** telepítésnél az import
megadja. Ha a realm még a szerep bevezetése előtt jött létre, pótolni kell a
szerveren (SSH-val) – ez az `sso` repóból fut, a saját (4b. pontban
beállított) `.env.prod`-jában lévő jelszóval:

```bash
cd /opt/sso
KEYCLOAK_ADMIN_PASSWORD="<amit az sso .env.prod-jába írtál>" \
  ./scripts/grant-manage-users.sh
```

Majd a Keycloak admin konzolján (Clients → `snitt-admin-api` → Credentials)
generálj egy új client secretet. Ezt **nem** a szerveren írod be: frissítsd a
`VPS_ENV` GitHub secretet az új `KEYCLOAK_ADMIN_CLIENT_SECRET` értékkel,
majd futtasd újra a "Deploy VPS" workflow-t (Actions → Run workflow) – ez
felülírja a szerveren a `.env.prod`-ot és újraindítja az `account`
szolgáltatást az új secrettel.

## 10. Napi mentés (cron)

Két külön mentés, két külön repóból – az account szolgáltatás adatai és a
Keycloak (mindkét app felhasználói) most külön adatbázisban vannak:

```bash
sudo crontab -u deploy -e
```

```
0 3 * * * cd /opt/snitt-cloud && PROJECT_NAME=snitt DB_ENGINE=postgres DB_SERVICE=postgres DB_NAME=snitt DB_USER=snitt COMPOSE_FILE=deploy/docker-compose.yml COMPOSE_ENV_FILE=/opt/snitt-cloud/deploy/.env.prod ./.commons/ops/backup.sh /var/backups/snitt >> /var/log/snitt-backup.log 2>&1
0 3 * * * COMPOSE_ENV_FILE=/opt/sso/.env.prod /opt/sso/scripts/backup.sh /var/backups/sso >> /var/log/sso-backup.log 2>&1
```

A mentő szkript a `.commons` submodule-ban van, tehát a szerveren előbb `git submodule update --init`
kell. A hosszú sor ugyanaz, amit helyben a `make backup DOTENV=deploy/.env.prod BACKUP_DIR=…` futtat;
a szerveren nincs `make`. A `COMPOSE_ENV_FILE` nélkül a compose jelszavak híján nem renderel, és nem
készül mentés.

Próbáld ki mindkettőt egyszer kézzel is – lásd
a `commons` repó `ops/backup.sh` és az `sso` repó
`scripts/backup.sh` fejlécét: egy soha ki nem próbált mentés nem mentés,
csak remény. A visszaállítás párja a `.commons/ops/restore.sh` (helyben
`make restore FILE=…`), ugyanazokkal a változókkal és `STOP_SERVICES=account`
beállítással; **felülírja** az adatbázist.

## 11. Hibajelentés a GlitchTipbe (opcionális)

A szolgáltatások a saját hibakövetőnkre ([`lipcsei/glitchtip`](https://github.com/lipcsei/glitchtip))
küldhetik a hibáikat, a szabványos Sentry SDK-kkal. **Minden opcionális: üres DSN = teljesen
kikapcsolva** – az SDK el sem indul, nincs hálózati forgalom, és semmi nem változik.

| `.env` változó | Melyik projekt | Mit jelent |
|---|---|---|
| `SENTRY_DSN_ACCOUNT` | `snitt-account` | Az account szolgáltatás pánikjait és a visszaadott 5xx hibáit (4xx-et soha). Futásidőben olvassa (a konténerben `SENTRY_DSN`), újraindítás elég. |
| `SENTRY_DSN_ADMIN` | `snitt-admin` | Az admin felület böngészős és render hibáit. **Build-időben** épül a képbe: a változtatásához újra kell építeni. |
| `SENTRY_DSN_LANDING` | `snitt-landing` | A landing böngészős és render hibáit. Ugyanúgy build-idejű; élesben a landing GitHub Pages-en épül, ott a `SENTRY_DSN_LANDING` *repository variable* adja (lásd [`ci.yml`](../.github/workflows/ci.yml)). |

A DSN a GlitchTip projekt *Client Keys (DSN)* oldalán van (a projekteket a `glitchtip` repó
`make projects` parancsa hozza létre). A környezet neve az eseményeken fejlesztői stackben
`development`, az éles composeban `production`.

- **Helyben:** `cp deploy/.env.example deploy/.env`, és töltsd ki a DSN-eket (a Dockerben futó
  account-nak `host.docker.internal` kell a `localhost` helyett).
- **Élesben:** a `deploy/.env.prod.example`-ben a három kulcs **kikommentelve** áll, mert az éles
  GlitchTip még nincs felállítva, és a `vps/github-env.sh check` minden kommentjel nélküli kulcsot
  kitöltöttnek vár. Ha megvan a DSN: vedd le a kommentjelet, írd be, frissítsd a `VPS_ENV` secretet,
  és futtasd újra a *Deploy VPS* workflow-t (az `up -d --build` újraépíti az admin képet is).
- **Adatvédelem:** nincs személyes adat (IP, süti, `Authorization` fejléc), a kérés törzse nem megy
  ki, a lekérdezés-szöveget (pl. az admin keresője: `?q=`) és az URL-fragmentet levágjuk; nincs
  tracing és session replay.
- **CSP:** a repóban (nginx, Go, `index.html`) nincs Content-Security-Policy, ezért a böngésző
  szabadon küldhet a GlitchTipnek. Ha később bevezetsz egyet, a GlitchTip hostját fel kell venni a
  `connect-src`-be.

## Ismert korlátok

- Ha a `VPS_ENV` secret és a szerveren futó `deploy/.env.prod` valaha
  szétcsúszna (pl. valaki mégis kézzel piszkálta a szerverit), a következő
  "Deploy VPS" futás visszaállítja a secretben tárolt állapotot – a secret a
  forrás igazság, nem a szerver.
- A `VPS_ENV` tartalma sehol máshol nincs meg mentve (a GitHub secretek
  utólag nem olvashatók vissza a felületen) – érdemes egy jelszókezelőbe is
  bemásolni, mielőtt letörlöd a saját géped `/tmp/env.prod.draft` fájlját.
- Ha a `deploy/docker-compose.prod.yml` service-neveit vagy a domainek
  számát bővíted, az `edge` repó `Caddyfile`-ját (és a `docker-compose.yml`-jét,
  `.env.prod.example`-jét) és a `deploy.yml` `up -d --build` sorát is bővíteni kell – ezek szándékosan
  nem generálódnak automatikusan.
- Az `sso` és az `edge` repónak is saját `Deploy VPS` workflow-ja van
  (`sso-prod`, illetve `edge-prod` environment, mindegyikben a saját
  `VPS_*` secretek – a GitHub nem oszt meg secretet repók között). Amíg a
  secretek nincsenek beállítva, azoknál az automatikus futás zölden kimarad.
  Ennek a repónak a `Deploy VPS` workflow-ja szigorúbb: hiányzó secretnél az
  automatikus futás is **hibával** áll meg.
