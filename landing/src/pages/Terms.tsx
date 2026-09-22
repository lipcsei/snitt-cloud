import { Link } from 'react-router-dom';
import '../styles/legal.css';

/**
 * Jogi oldal, lásd PrivacyPolicy.tsx megjegyzését: kizárólag magyarul, nyelvi útvonal-előtag
 * nélkül (/aszf), a hu/en/de PATHS-rendszeren kívül - App.tsx-ben egy külön <Route>.
 */
export default function Terms() {
  return (
    <main className="page legal-page">
      <div className="container container-narrow legal-content">
        <div className="legal-notice">
          <p>
            This document is currently available in Hungarian only, as the service operates under
            Hungarian law. An English version will follow.
          </p>
          <p>
            Dieses Dokument ist derzeit nur auf Ungarisch verfügbar, da der Dienst dem ungarischen
            Recht unterliegt. Eine deutsche Version folgt.
          </p>
        </div>

        <span className="eyebrow">ÁSZF</span>
        <h1>Általános Szerződési Feltételek — Snitt</h1>
        <p className="muted legal-effective">Hatályos: 2026. szeptember 22-től</p>

        <section>
          <h2>1. Általános rendelkezések</h2>
          <p>
            Jelen Általános Szerződési Feltételek (ÁSZF) a Snitt asztali alkalmazás és a
            snitt.video weboldal („Szolgáltatás") használatának feltételeit szabályozzák. A
            Szolgáltatás letöltésével, telepítésével vagy használatával a felhasználó elfogadja a
            jelen ÁSZF rendelkezéseit.
          </p>
        </section>

        <section>
          <h2>2. A Szolgáltató adatai</h2>
          <p>
            Szolgáltató: <strong>Lipcsei Sándor</strong> (magánszemélyként üzemeltetett
            szolgáltatás, egyéni vállalkozói nyilvántartásba vétele folyamatban).
          </p>
          <p>Kapcsolat: sandor.lipcsei@gmail.com</p>
        </section>

        <section>
          <h2>3. A Szolgáltatás leírása</h2>
          <p>
            A Snitt egy asztali alkalmazás, amely lehetővé teszi, hogy a felhasználó egy beírt
            idézet vagy szövegrészlet alapján megtalálja és kivágja a hozzá tartozó jelenetet a
            saját, helyben tárolt videóiból. A feldolgozás teljes egészében a felhasználó saját
            gépén történik.
          </p>
          <p>
            <strong>Fontos figyelmeztetés — szerzői jog:</strong> A Szolgáltatás egy eszköz, amely
            a felhasználó saját tulajdonában lévő vagy általa jogszerűen birtokolt videófájlok
            feldolgozására szolgál. A felhasználó kizárólagos felelőssége, hogy a Szolgáltatással
            létrehozott videórészleteket (klipeket) a szerzői jogi és egyéb vonatkozó
            jogszabályoknak megfelelően használja fel, ossza meg vagy tegye közzé. A Szolgáltató
            nem vállal felelősséget a felhasználó által létrehozott vagy megosztott tartalmak
            jogszerűségéért.
          </p>
        </section>

        <section>
          <h2>4. Regisztráció és felhasználói fiók</h2>
          <p>
            A Szolgáltatás alapfunkciói regisztráció nélkül, helyben használhatók. Amennyiben a
            weboldalon fiókot hoz létre (pl. frissítésekhez vagy bővített funkciókhoz), köteles
            valós e-mail címet megadni és a hozzáférési adatait bizalmasan kezelni.
          </p>
        </section>

        <section>
          <h2>5. Díjazás</h2>
          <p>
            A Szolgáltatás jelenleg díjmentesen vehető igénybe. A Szolgáltató fenntartja a jogot
            fizetős funkciók jövőbeli bevezetésére, előzetes tájékoztatás és a jelen ÁSZF
            kiegészítése mellett.
          </p>
        </section>

        <section>
          <h2>6. Felhasználói magatartás</h2>
          <p>
            Tilos: a Szolgáltatás rendeltetésétől eltérő, jogszabályba ütköző célra történő
            felhasználása (pl. szerzői joggal védett tartalom jogosulatlan másolása, terjesztése),
            a Szolgáltatás visszafejtése kereskedelmi célból, a Szolgáltatás működésének szándékos
            akadályozása.
          </p>
        </section>

        <section>
          <h2>7. Szellemi tulajdon</h2>
          <p>
            A Szolgáltatás szoftvere, megjelenése és márkaneve a Szolgáltató szellemi tulajdona (a
            nyílt forráskódú komponensek kivételével, amelyek saját licencfeltételeik szerint
            használhatók, lásd a projekt nyilvános tárhelyét/repository-ját).
          </p>
        </section>

        <section>
          <h2>8. Felelősség korlátozása</h2>
          <p>
            A Szolgáltató a Szolgáltatást „ahogy van" alapon nyújtja. A 3. pontban foglalt szerzői
            jogi figyelmeztetésen túl, a Szolgáltató a jogszabály által megengedett legnagyobb
            mértékben kizárja felelősségét a Szolgáltatás elérhetetlenségéből, hibájából, vagy a
            felhasználó saját videóiban bekövetkező adatvesztésből eredő közvetett vagy
            következményi károkért. A felhasználó felelőssége saját fájljairól rendszeres
            biztonsági mentést készíteni.
          </p>
        </section>

        <section>
          <h2>9. A Szolgáltatás elérhetősége, módosítása, megszüntetése</h2>
          <p>
            A Szolgáltató törekszik a folyamatos elérhetőségre, de nem garantálja a megszakítás
            nélküli működést, és jogosult a Szolgáltatást módosítani vagy előzetes tájékoztatás
            mellett megszüntetni.
          </p>
        </section>

        <section>
          <h2>10. Fiók felfüggesztése, törlése</h2>
          <p>
            A Szolgáltató jogosult a jelen ÁSZF-et súlyosan vagy ismételten megszegő felhasználó
            fiókját felfüggeszteni vagy törölni. A felhasználó saját fiókját bármikor törölheti a
            Kapcsolat e-mail címen küldött kérelemmel.
          </p>
        </section>

        <section>
          <h2>11. Panaszkezelés és jogérvényesítés</h2>
          <p>
            Panasz esetén a felhasználó a sandor.lipcsei@gmail.com címen jelezheti kifogását,
            ennek eredménytelensége esetén fogyasztóvédelmi hatósághoz, békéltető testülethez vagy
            bírósághoz fordulhat.
          </p>
        </section>

        <section>
          <h2>12. Vegyes rendelkezések</h2>
          <p>A jelen ÁSZF-ben nem szabályozott kérdésekben a magyar jog az irányadó.</p>
        </section>

        <section>
          <h2>13. Hatálybalépés</h2>
          <p>Jelen ÁSZF 2026. szeptember 22-től hatályos.</p>
        </section>

        <div className="legal-links">
          <Link to="/adatvedelem" className="btn btn-outline">
            Adatvédelmi tájékoztató
          </Link>
          <Link to="/" className="btn btn-outline">
            Vissza a főoldalra
          </Link>
        </div>
      </div>
    </main>
  );
}
