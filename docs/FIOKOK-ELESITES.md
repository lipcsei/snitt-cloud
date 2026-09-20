# Fiókok élesítése – ellenőrzőlista

A Keycloak beállításai fejlesztéshez vannak hangolva. Ez a lista azt sorolja
fel, amit **az első valódi felhasználó előtt** meg kell változtatni – a
sorrend a fontosság sorrendje.

## 1. SMTP – enélkül a jelszó-emlékeztető néma

A realmben `resetPasswordAllowed: true`, tehát a bejelentkező oldalon ott az
„Elfelejtetted a jelszavad?" link. SMTP nélkül a felhasználó beírja az
e-mail-címét, és **soha nem érkezik semmi** – ez rosszabb, mint ha nem is
kínálnánk.

Ugyanez kell az e-mail-cím igazolásához is (lásd 2. pont).

**A Keycloak most a megosztott [sso](https://github.com/lipcsei/sso) repóban
fut** (`realms/snitt-realm.json`-jának `smtpServer` blokkja a `mailpit`
fejlesztői mockra mutat – az is abban a repóban van, `http://localhost:8025`
a webes felülete – minden "kiküldött" levelet ott lehet megnézni, valódi
SMTP vagy külső szolgáltatás nélkül). Ez **csak fejlesztéshez** való: az sso
repó `docker-compose.prod.yml`-je nem hozza fel a `mailpit`-et, tehát
élesben ez a beállítás hatástalan levélküldésre – a valódi SMTP-t a
Keycloak admin konzolján (Realm settings → Email) kell beállítani, mielőtt
igazi felhasználó regisztrál.

Keycloak → Realm settings → Email.

## 2. `verifyEmail: true`

Ma bárki regisztrálhat más e-mail-címével. Amíg csak keresés van, ez
kellemetlen; **előfizetés és számlázás mellett viszont már adatvédelmi
kérdés**, mert a számlaértesítő idegenhez menne.

Csak az SMTP beállítása UTÁN kapcsold be, különben senki nem tud belépni.

## 3. `sslRequired: external`

Most `none`, mert a helyi fejlesztés HTTP-n megy. Élesben ez azt jelentené,
hogy a jelszó titkosítatlanul is elmehet. Az `external` a helyes érték: a
loopback maradhat HTTP, minden más kötelezően HTTPS.

## 4. A kliensek visszairányítási címei

A `snitt-landing` és a `snitt-admin` a fejlesztői `localhost` címeket
tartalmazza. Éles domainre kell állítani – a `snitt-desktop` kivétel: annak a
loopback címe a helyes, azt ne írd át (RFC 8252).

## 5. A fejlesztői fiókok törlése

A realm exportban két beépített fiók van (`snittadmin`, `demo`), ismert
jelszóval. Éles realmben ezeknek nincs helyük: töröld őket, és hozz létre
saját admint.

## 6. Keycloak bootstrap admin

A `admin`/`admin` párost (`KEYCLOAK_ADMIN`, `KEYCLOAK_ADMIN_PASSWORD`) most
a megosztott [sso](https://github.com/lipcsei/sso) repó saját `.env.prod`-ja
adja, nem ez a repó. Élesben ott cseréld, és a Keycloak admin felületét ne
tedd ki a nyilvános internetre.

## 7. Mentés – és próbáld is ki

A `deploy/scripts/backup.sh` itt csak az account szolgáltatás adatbázisát
menti (profilok, előfizetések, számlák). A Keycloak adatbázisa (felhasználók,
jelszó-lenyomatok, kliensbeállítások, mindkét app realmje) külön van, a
megosztott [sso](https://github.com/lipcsei/sso) repóban – annak saját
`scripts/backup.sh`-ja menti, külön cronnal. Mindkettő **nem
újratermelhető adat** – a landing statikus, a képek újraépíthetők, a
felhasználók videói pedig a saját gépükön vannak – tehát mindkét cront be
kell állítani.

```bash
# naponta hajnali 3-kor
0 3 * * * /opt/snitt/deploy/scripts/backup.sh /var/backups/snitt >> /var/log/snitt-backup.log 2>&1
0 3 * * * /opt/sso/scripts/backup.sh /var/backups/sso >> /var/log/sso-backup.log 2>&1
```

Két dolog, ami nélkül a mentés csak illúzió:

- **Vidd le a VPS-ről.** Egy mentés ugyanazon a lemezen, ami elveszhet, nem
  mentés. `rsync`, `rclone` vagy a szolgáltató objektumtára – mindegy, csak
  máshol legyen.
- **Állítsd is vissza egyszer.** A `restore.sh` megvan hozzá. Egy soha ki nem
  próbált mentés nem mentés, csak remény.

## Ami már jó, és nem kell hozzányúlni

- **Jelszóházirend**: `length(10) and notUsername and notEmail`. NIST-igazodó:
  a hossz számít, nem a kötelező írásjelek. A felhasználónév és az e-mail
  tiltása azért kell, mert azt a támadó eleve ismeri.
- **Jelszókitalálás elleni védelem**: bekapcsolva, növekvő várakozással. Nem
  végleges zárolás – azzal bárki kizárhatna egy másik felhasználót a
  fiókjából.
- **Munkamenet hossza**: 30 nap tétlenség, 90 nap maximum. Az asztali
  alkalmazásnak hosszú munkamenet kell; az alapértelmezett 30 perc után újra
  be kellene jelentkezni.
- **Nyelvek**: hu, en, de – a bejelentkezés ugyanazon a nyelven jön, mint a
  landing és az alkalmazás.
- **A desktop kliens** titok nélküli, PKCE-vel – asztali alkalmazásba nem
  lehet titkot csomagolni.
- **A Keycloak a Postgresbe ír**, saját adatbázisba (a megosztott
  [sso](https://github.com/lipcsei/sso) repóban, saját Postgresébe – nem
  ennek a repónak a Postgresébe). Ez nem apróság: a `start-dev` alapból a
  konténerbe ágyazott H2-t használná, kötet nélkül – minden
  újralétrehozáskor elveszne minden felhasználó és jelszó, és menteni sem
  lenne mit.
