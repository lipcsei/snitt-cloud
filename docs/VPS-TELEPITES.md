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
fut, de saját compose projektként, saját ütemben telepítve (nincs hozzá
GitHub Actions workflow, kézzel települ – lásd annak README-jét).

**A TLS-terminátor (Caddy) SEM ide megy.** A 80/443-as portot egy hoston
egyszerre csak egy folyamat foghatja le, ezért egyetlen megosztott proxy
kezeli minden app helyett: [`lipcsei/edge`](https://github.com/lipcsei/edge).
Az köti az `auth.snitt.video`, `api.snitt.video`, `admin.snitt.video` (és a
breath) címeit a megfelelő konténerekhez két megosztott Docker hálózaton
keresztül (`sso-net`, `edge-net`). Ez a stack csak az `account` és `admin`
konténert csatlakoztatja az `edge-net`-re, egyedi aliassal.

**Indítási sorrend a VPS-en: `sso` (4b) → `edge` (4c) → ez (8).** Az első
kettő hozza létre a hálózatokat, amiket a mi konténereink használnak.

**A titkok (jelszavak, kliens secret) egyetlen forrása a GitHub repo
secretjei, nem a szerver.** A `deploy/.env.prod` fájlt maga a deploy workflow
írja fel a szerverre minden kitelepítéskor a `VPS_ENV` secretből – a
szerveren SOSEM szerkeszted kézzel; ha egy jelszót cserélnél, a GitHub
secretet frissíted és újrafuttatod a workflow-t.

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
(lásd 7. pont) – ezután a saját gépedről törölheted, csak a GitHub Actions
fogja használni.

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
és csak-olvasásra add hozzá a GitHubon (Settings → Deploy keys), lásd a
korábbi üzeneteinket erről. Utána:

```bash
sudo mkdir -p /opt/snitt-cloud
sudo chown deploy:deploy /opt/snitt-cloud
git clone git@github.com-snitt-cloud:lipcsei/snitt-cloud.git /opt/snitt-cloud
cd /opt/snitt-cloud
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
# KEYCLOAK_PUBLIC_URL=https://auth.snitt.video
chmod 600 .env.prod

# Az éles realm-fájlokat a fejlesztői realm-fájlokból GENERÁLJA (a dev
# fájlokban ismert jelszavú felhasználók vannak, azok élesben nem mehetnek
# ki) – a compose enélkül el sem indul. A címek a breath appéi (lásd annak
# docs/DEPLOY.md-jét); akkor is meg kell adni, ha a breath még nincs kint.
BREATH_APP_URL=https://breath.snitt.video \
BREATH_ADMIN_URL=https://breath-admin.snitt.video \
  ./scripts/make-prod-realms.sh

docker compose --env-file .env.prod \
  -f docker-compose.yml -f docker-compose.prod.yml \
  up -d --build
```

Ez hozza létre az `sso-net` Docker hálózatot is, amihez az `edge` proxy és a
snitt-cloud `account` szolgáltatása csatlakozik – ezért kell ennek a
lépésnek megelőznie a 4c. és a 8. pontot.

A generátor **figyelmeztet**, ha egy változatlanul másolt realm (pl. a `snitt`)
még ismert jelszavú fejlesztői felhasználókat tartalmaz – friss adatbázison
ezek importálódnak. Lásd `docs/FIOKOK-ELESITES.md` 5. pont: az első
indítás után töröld őket.

Ha a Keycloakot később, más appok bevonása miatt (pl. új realm hozzáadása)
frissíteni kell, ugyanígy, kézzel: `cd /opt/sso && git pull && docker
compose --env-file .env.prod -f docker-compose.yml -f
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
# szerkeszd: ACME_EMAIL, és a PUBLIC_*_URL címek (a DNS-t ELŐBB állítsd be,
# lásd 5. pont – enélkül a Let's Encrypt tanúsítványkérés elbukik)
chmod 600 .env.prod

docker compose --env-file .env.prod up -d
```

Ez hozza létre az `edge-net` hálózatot. A proxy akkor is elindul, ha az appok
még nem futnak – a címük addig `502`-t ad.

**Ha a szerveren korábban a snitt-cloud saját Caddyje futott**, az még foglalja
a 80/443-at: az átállás lépéseit (rövid kiesés) az
[`edge` README-je](https://github.com/lipcsei/edge#migrating-from-a-per-app-caddy-snitt-cloud-used-to-run-its-own)
írja le.

## 5. DNS

Állítsd be a DNS A-rekordokat (`auth`, `api`, `admin` a szerver IP-jére), és
várd meg, míg terjednek (`dig auth.snitt.video` mutassa a helyes IP-t) –
enélkül a 8. pontban a Let's Encrypt tanúsítványkérés elbukik.

## 6. `.env.prod` tartalmának előkészítése (a SAJÁT gépeden, nem a szerveren)

```bash
cp deploy/.env.prod.example /tmp/env.prod.draft
```

Töltsd ki `/tmp/env.prod.draft`-ot:
- `POSTGRES_PASSWORD` – generáld: `openssl rand -base64 24` (ez a snitt-cloud
  SAJÁT Postgresének jelszava – a 4b. pontban az `sso` repóhoz külön,
  MÁSIK generált jelszót adtál a Keycloak adatbázisának)
- a domainek (`PUBLIC_KEYCLOAK_URL` stb.) maradhatnak az alapértelmezésen, ha a `snitt.video`-t használod
- `KEYCLOAK_ADMIN_CLIENT_SECRET` egyelőre maradjon a fejlesztői értéken – ezt a 9. pontban, az ELSŐ felállás UTÁN cseréled, mert a kliens csak akkor jön létre (a realm importból, ami a 4b. pontban lefutott)

Ez a fájl a te géped `/tmp`-jében marad, git alá SOHA nem kerül – a teljes
tartalmát a következő pontban egy GitHub secretbe másolod, utána törölheted:
`rm /tmp/env.prod.draft`.

## 7. GitHub secretek és változók

A repo Settings → Environments → **`snitt-cloud-prod`** environment alatt hozd
létre (a workflow ezt az environmentet nevezi meg; repository szintű secretet
nem használ):

| Hova | Név | Érték |
|---|---|---|
| változó vagy secret | `VPS_HOST` | a szerver IP-je vagy hostneve |
| változó vagy secret | `VPS_USER` | `deploy` |
| változó vagy secret | `VPS_SSH_PORT` | csak ha nem a 22-es |
| **csak secret** | `VPS_SSH_KEY` | a 2. pontban generált **privát** kulcs teljes tartalma |
| **csak secret** | `VPS_ENV` | a 6. pontban kitöltött `/tmp/env.prod.draft` fájl **teljes** tartalma, változtatás nélkül bemásolva |

A `VPS_SSH_KEY` és a `VPS_ENV` kerüljön az *Environment secrets* alá: a változót (*Environment variables*) a GitHub sima szövegként tárolja, és a futási naplóban sem takarja ki. Ha a workflow ezt a kettőt változóként találja, hibával leáll.

## 8. Első kitelepítés

Ellenőrizd, hogy a 4b. pont `sso` stackje már fut (`docker compose -f
/opt/sso/docker-compose.yml -f /opt/sso/docker-compose.prod.yml ps` –
`keycloak` legyen `Up`), utána:

GitHub → repo → **Actions → Deploy VPS → Run workflow** (a `workflow_dispatch`
kézi indítás pont erre való – build nélkül, csak a meglévő kódot viszi ki).
Ez felírja a szerverre a `deploy/.env.prod`-ot a secretből, és elindítja a
snitt-cloud stacket (`postgres`, `account`, `admin`).

Kövesd a futást az Actions fülön; ha piros lesz, a log megmondja, melyik
lépésnél (SCP feltöltés vagy SSH-s `docker compose`) akadt el.

Ellenőrzés a szerveren:

```bash
cd /opt/snitt-cloud
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml ps
# a proxy naplója (tanúsítványkérés, 502-k) az edge stackben van:
docker compose -f /opt/edge/docker-compose.yml logs -f caddy
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

Emellett most, hogy a realm létrejött, a szerveren (SSH-val) – ez most az
`sso` repóból fut, a saját (4b. pontban beállított) `.env.prod`-jában lévő
jelszóval:

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
0 3 * * * /opt/snitt-cloud/deploy/scripts/backup.sh /var/backups/snitt >> /var/log/snitt-backup.log 2>&1
0 3 * * * /opt/sso/scripts/backup.sh /var/backups/sso >> /var/log/sso-backup.log 2>&1
```

Próbáld ki mindkettőt egyszer kézzel is – lásd
[`deploy/scripts/backup.sh`](../deploy/scripts/backup.sh) és az `sso` repó
`scripts/backup.sh` fejlécét: egy soha ki nem próbált mentés nem mentés,
csak remény.

## Ismert korlátok

- Ha a `VPS_ENV` secret és a szerveren futó `deploy/.env.prod` valaha
  szétcsúszna (pl. valaki mégis kézzel piszkálta a szerverit), a következő
  "Deploy VPS" futás visszaállítja a secretben tárolt állapotot – a secret a
  forrás igazság, nem a szerver.
- A `VPS_ENV` tartalma sehol máshol nincs meg mentve (a GitHub secretek
  utólag nem olvashatók vissza a felületen) – érdemes egy jelszókezelőbe is
  bemásolni, mielőtt letörlöd a saját géped `/tmp/env.prod.draft` fájlját.
- Ha a `deploy/docker-compose.prod.yml` service-neveit vagy a domainek
  számát bővíted, az `edge` repó `Caddyfile`-ját (és a `docker-compose.yml`-jét)
  és a `deploy.yml` `up -d --build` sorát is bővíteni kell – ezek szándékosan
  nem generálódnak automatikusan.
- Az `sso` és az `edge` repónak is saját `Deploy VPS` workflow-ja van
  (`sso-prod`, illetve `edge-prod` environment, mindegyikben a saját
  `VPS_*` secretek – a GitHub nem oszt meg secretet repók között). Amíg a
  secretek nincsenek beállítva, az automatikus futás zölden kimarad.
