import { Link } from 'react-router-dom';
import '../styles/legal.css';

/**
 * Jogi oldal (nem a marketing snippet - lásd sections/Privacy.tsx). A jogi dokumentumok
 * egyelőre kizárólag magyarul készülnek el (a Szolgáltató magyar joghatóság alatt működik),
 * ezért ez az oldal a nyelvi útvonal-rendszeren (PATHS/RouteKey) kívül, egyetlen, nyelv-előtag
 * nélküli útvonalon (/adatvedelem) érhető el - lásd App.tsx.
 */
export default function PrivacyPolicy() {
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

        <span className="eyebrow">Adatvédelem</span>
        <h1>Adatvédelmi tájékoztató — Snitt</h1>
        <p className="muted legal-effective">Hatályos: 2026. szeptember 22-től</p>

        <section>
          <h2>1. Bevezetés</h2>
          <p>
            Jelen tájékoztató a Snitt alkalmazás és a snitt.video weboldal („Szolgáltatás")
            használata során kezelt személyes adatokról ad tájékoztatást, a GDPR és az információs
            önrendelkezési jogról szóló 2011. évi CXII. törvény alapján.
          </p>
        </section>

        <section>
          <h2>2. Az adatkezelő adatai</h2>
          <p>
            Adatkezelő: <strong>Lipcsei Sándor</strong>, aki jelenleg magánszemélyként üzemelteti a
            Szolgáltatást (egyéni vállalkozói nyilvántartásba vétele folyamatban; az adatok a
            bejegyzést követően frissülnek).
          </p>
          <p>
            Kapcsolat: <strong>sandor.lipcsei@gmail.com</strong>
          </p>
        </section>

        <section>
          <h2>3. A kezelt adatok köre és célja</h2>
          <p>
            <strong>a) A Snitt asztali alkalmazás videófeldolgozása:</strong> a felhasználó saját
            videóit a Snitt kizárólag a felhasználó saját gépén, helyben dolgozza fel és indexeli
            — ezek az adatok (videók, feliratok, keresési előzmények) nem kerülnek fel a
            Szolgáltató szerverére, azokhoz a Szolgáltató nem fér hozzá.
          </p>
          <p>
            <strong>b) Fiók/profil</strong> (amennyiben regisztrálsz a weboldalon): e-mail cím,
            felhasználónév, és a fiók működéséhez szükséges alapadatok. Cél: a fiók létrehozása,
            azonosítás, a letöltéshez/frissítésekhez kapcsolódó szolgáltatások nyújtása.
          </p>
          <p>
            <strong>c) Hibabejelentés</strong> (opcionális, kizárólag hozzájárulással vagy a
            felhasználó kifejezett beküldésével): amennyiben a Snitt alkalmazásban hiba történik
            és a felhasználó ehhez hozzájárul, technikai adatok (pl. alkalmazás-verzió,
            hibaüzenet, operációs rendszer típusa) kerülnek elküldésre a Szolgáltató saját
            üzemeltetésű hibakövető rendszerébe (GlitchTip, errors.snitt.video), kizárólag a hiba
            kijavítása céljából. Videótartalom vagy személyes fájl soha nem része a
            hibajelentésnek.
          </p>
          <p>
            <strong>d) Analitikai adatok</strong> (Google Analytics a snitt.video weboldalon),
            kizárólag hozzájárulás esetén: a weboldal látogatottsági statisztikái. A hozzájárulás a
            belépéskor megjelenő sávon adható meg vagy tagadható meg.
          </p>
        </section>

        <section>
          <h2>4. Az adatkezelés jogalapja</h2>
          <ul>
            <li>Fiók/profil adatok: GDPR 6. cikk (1) b) — szerződés teljesítése.</li>
            <li>
              Hibabejelentés: GDPR 6. cikk (1) a) — hozzájárulás (ahol a beküldés automatikus és
              nem kifejezetten felhasználói kattintáshoz kötött, ott GDPR 6. cikk (1) f) — jogos
              érdek a szoftverhibák elhárításához).
            </li>
            <li>Analitikai sütik: GDPR 6. cikk (1) a) — önkéntes hozzájárulás.</li>
          </ul>
        </section>

        <section>
          <h2>5. Adatfeldolgozók és címzettek</h2>
          <ul>
            <li>
              Tárhely (weboldal): saját üzemeltetésű szerver / statikus tárhely (RackForest Kft.,
              Magyarország, illetve a telepítő fájlok terjesztésére esetlegesen GitHub Releases)
            </li>
            <li>Hibakövetés: saját üzemeltetésű GlitchTip (errors.snitt.video)</li>
            <li>Analitika (csak hozzájárulással): Google Ireland Limited (Google Analytics)</li>
          </ul>
          <p>
            A felhasználó videói, feliratai és keresései nem kerülnek átadásra semmilyen harmadik
            félnek, mivel azok soha nem hagyják el a felhasználó gépét.
          </p>
        </section>

        <section>
          <h2>6. Az adatok megőrzésének ideje</h2>
          <p>
            A fiókadatok a fiók törléséig kerülnek tárolásra. Hibabejelentések legfeljebb 30 napig.
            Analitikai sütik legfeljebb 14 hónapig, vagy a hozzájárulás visszavonásáig.
          </p>
        </section>

        <section>
          <h2>7. Cookie-k és hasonló technológiák</h2>
          <ul>
            <li>
              <strong>Feltétlenül szükséges cookie-k</strong>: a weboldal alapvető működéséhez
              (pl. nyelvi beállítás megjegyzése, bejelentkezési munkamenet), ezek elfogadása nem
              opcionális.
            </li>
            <li>
              <strong>Analitikai cookie-k</strong>: kizárólag hozzájárulás esetén töltődnek be, a
              belépéskor megjelenő sávon bármikor elutasíthatók.
            </li>
          </ul>
        </section>

        <section>
          <h2>8. Az érintett jogai</h2>
          <p>A felhasználó bármikor jogosult:</p>
          <ul>
            <li>tájékoztatást kérni a róla kezelt adatokról (hozzáférési jog),</li>
            <li>kérni a pontatlan adatok helyesbítését,</li>
            <li>kérni az adatok törlését („elfeledtetéshez való jog"), pl. a fiók törlésével,</li>
            <li>kérni az adatkezelés korlátozását,</li>
            <li>tiltakozni az adatkezelés ellen (jogos érdeken alapuló adatkezelés esetén),</li>
            <li>kérni az adatok hordozhatóságát (gépileg olvasható formátumban),</li>
            <li>
              a hozzájárulását bármikor, indokolás nélkül visszavonni (ez nem érinti a
              visszavonás előtti adatkezelés jogszerűségét),
            </li>
            <li>
              panaszt tenni a Nemzeti Adatvédelmi és Információszabadság Hatóságnál (NAIH, 1055
              Budapest, Falk Miksa utca 9-11., ugyfelszolgalat@naih.hu, www.naih.hu), vagy
              bírósághoz fordulni.
            </li>
          </ul>
        </section>

        <section>
          <h2>9. Adatbiztonság</h2>
          <p>
            A weboldal kizárólag titkosított (HTTPS) kapcsolaton érhető el. A Snitt asztali
            alkalmazás a felhasználó videóit soha nem továbbítja a Szolgáltató felé.
          </p>
        </section>

        <section>
          <h2>10. A tájékoztató módosítása</h2>
          <p>
            Jelen tájékoztatót időről időre frissíthetjük, a mindenkor hatályos verzió ezen az
            oldalon érhető el.
          </p>
        </section>

        <section>
          <h2>11. Kapcsolat</h2>
          <p>
            Kérdés vagy kérelem esetén: <strong>sandor.lipcsei@gmail.com</strong>
          </p>
        </section>

        <div className="legal-links">
          <Link to="/aszf" className="btn btn-outline">
            Általános Szerződési Feltételek
          </Link>
          <Link to="/" className="btn btn-outline">
            Vissza a főoldalra
          </Link>
        </div>
      </div>
    </main>
  );
}
