# VPS-telepítés – egy üres szerver előkészítése

Ez a lista egy vadonatúj, semmivel fel nem installált szerveren (Ubuntu/Debian
feltételezve) viszi végig a Snitt felhő oldalának (`account` + `admin` +
Keycloak + Postgres, Caddy TLS-terminátorral) első felállítását. A cél
állapot: `git push origin main` után a GitHub Actions automatikusan
kitelepít – lásd [`.github/workflows/deploy.yml`](../.github/workflows/deploy.yml).

A `landing` NEM ide megy – az GitHub Pages-en fut, lásd
[`ci.yml`](../.github/workflows/ci.yml). Ez a gép csak a másik négy
komponenst szolgálja ki.

**A titkok (jelszavak, kliens secret) egyetlen forrása a GitHub repo
secretjei, nem a szerver.** A `deploy/.env.prod` fájlt maga a deploy workflow
írja fel a szerverre minden kitelepítéskor a `VPS_ENV_PROD` secretből – a
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

# Csak SSH, HTTP, HTTPS legyen nyitva kifelé - minden más (Postgres, Keycloak,
# account, admin) a docker-compose.prod.yml miatt úgyis csak 127.0.0.1-re
# lesz kötve, a tűzfal ez ellen egy második védelmi vonal.
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

## 5. DNS

Állítsd be a DNS A-rekordokat (`auth`, `api`, `admin` a szerver IP-jére), és
várd meg, míg terjednek (`dig auth.snitt.video` mutassa a helyes IP-t) –
enélkül a 8. pontban a Let's Encrypt tanúsítványkérés elbukik.

## 6. `.env.prod` tartalmának előkészítése (a SAJÁT gépeden, nem a szerveren)

```bash
cp deploy/.env.prod.example /tmp/env.prod.draft
```

Töltsd ki `/tmp/env.prod.draft`-ot:
- `POSTGRES_PASSWORD` és `KEYCLOAK_ADMIN_PASSWORD` – generáld: `openssl rand -base64 24` (futtasd kétszer, külön-külön érték kell)
- a domainek (`PUBLIC_KEYCLOAK_URL` stb.) maradhatnak az alapértelmezésen, ha a `snitt.video`-t használod
- `KEYCLOAK_ADMIN_CLIENT_SECRET` egyelőre maradjon a fejlesztői értéken – ezt a 9. pontban, az ELSŐ felállás UTÁN cseréled, mert a kliens csak akkor jön létre (a realm importból)

Ez a fájl a te géped `/tmp`-jében marad, git alá SOHA nem kerül – a teljes
tartalmát a következő pontban egy GitHub secretbe másolod, utána törölheted:
`rm /tmp/env.prod.draft`.

## 7. GitHub secretek

A repo Settings → Secrets and variables → Actions alatt hozd létre:

| Secret | Érték |
| --- | --- |
| `VPS_HOST` | a szerver IP-je vagy hostneve |
| `VPS_USER` | `deploy` |
| `VPS_SSH_KEY` | a 2. pontban generált **privát** kulcs teljes tartalma |
| `VPS_SSH_PORT` | csak akkor kell, ha nem a 22-es portot használod |
| `VPS_ENV_PROD` | a 6. pontban kitöltött `/tmp/env.prod.draft` fájl **teljes** tartalma, változtatás nélkül bemásolva |

## 8. Első kitelepítés

GitHub → repo → **Actions → Deploy VPS → Run workflow** (a `workflow_dispatch`
kézi indítás pont erre való – build nélkül, csak a meglévő kódot viszi ki).
Ez felírja a szerverre a `deploy/.env.prod`-ot a secretből, és elindítja a
teljes stacket (`postgres`, `keycloak`, `account`, `admin`, `caddy`).

Kövesd a futást az Actions fülön; ha piros lesz, a log megmondja, melyik
lépésnél (SCP feltöltés vagy SSH-s `docker compose`) akadt el.

Ellenőrzés a szerveren:

```bash
cd /opt/snitt-cloud
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml ps
docker compose -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml logs -f caddy
```

Az első indításkor a Caddy kikéri a Let's Encrypt tanúsítványokat mindhárom
domainhez – ez néhány másodperc, ha a DNS már terjedt. Ha "no such host"
vagy "connection refused" hibát látsz a logban, a DNS még nem ért oda, vagy
a 80/443 port nincs nyitva kifelé.

Nyisd meg böngészőben:
- `https://auth.snitt.video` – Keycloak bejelentkező oldal
- `https://admin.snitt.video` – admin felület

## 9. Élesítési lépések (kötelező, mielőtt igazi felhasználó regisztrál)

Kövesd végig [`docs/FIOKOK-ELESITES.md`](FIOKOK-ELESITES.md)-t: SMTP, e-mail
igazolás, a Keycloak kliensek valódi redirect URI-jai, a fejlesztői
`snittadmin`/`demo` fiókok törlése, a bootstrap admin jelszó cseréje.

Emellett most, hogy a realm létrejött, a szerveren (SSH-val):

```bash
cd /opt/snitt-cloud
KEYCLOAK_ADMIN_PASSWORD="<amit a VPS_ENV_PROD-ba írtál>" \
  COMPOSE_FILE="deploy/docker-compose.yml" \
  ./deploy/scripts/grant-manage-users.sh
```

Majd a Keycloak admin konzolján (Clients → `snitt-admin-api` → Credentials)
generálj egy új client secretet. Ezt **nem** a szerveren írod be: frissítsd a
`VPS_ENV_PROD` GitHub secretet az új `KEYCLOAK_ADMIN_CLIENT_SECRET` értékkel,
majd futtasd újra a "Deploy VPS" workflow-t (Actions → Run workflow) – ez
felülírja a szerveren a `.env.prod`-ot és újraindítja az `account`
szolgáltatást az új secrettel.

## 10. Napi mentés (cron)

```bash
sudo crontab -u deploy -e
```

```
0 3 * * * /opt/snitt-cloud/deploy/scripts/backup.sh /var/backups/snitt >> /var/log/snitt-backup.log 2>&1
```

Próbáld ki egyszer kézzel is – lásd
[`deploy/scripts/backup.sh`](../deploy/scripts/backup.sh) fejlécét: egy
soha ki nem próbált mentés nem mentés, csak remény.

## Ismert korlátok

- Ha a `VPS_ENV_PROD` secret és a szerveren futó `deploy/.env.prod` valaha
  szétcsúszna (pl. valaki mégis kézzel piszkálta a szerverit), a következő
  "Deploy VPS" futás visszaállítja a secretben tárolt állapotot – a secret a
  forrás igazság, nem a szerver.
- A `VPS_ENV_PROD` tartalma sehol máshol nincs meg mentve (a GitHub secretek
  utólag nem olvashatók vissza a felületen) – érdemes egy jelszókezelőbe is
  bemásolni, mielőtt letörlöd a saját géped `/tmp/env.prod.draft` fájlját.
- Ha a `deploy/docker-compose.prod.yml` service-neveit vagy a domainek
  számát bővíted, a `Caddyfile`-t és a `deploy.yml` `up -d --build` sorát is
  bővíteni kell – ezek szándékosan nem generálódnak automatikusan.
