import '../styles.privacy.css';

/**
 * A termék legfontosabb ígérete, ezért kap saját, hangsúlyos szakaszt:
 * nincs távoli szerver, mindenki a SAJÁT adatbázisát építi a saját gépén.
 */
export default function Privacy() {
  return (
    <section className="privacy" id="adataid">
      <div className="privacy-inner">
        <p className="privacy-eyebrow">Ahogy működnie kell</p>
        <h2>
          Semmi nem kerül távoli szerverre.
          <br />
          <span className="privacy-accent">A videótárad a te gépeden marad.</span>
        </h2>
        <p className="privacy-lead">
          A Snitt nem szolgáltatás, hanem egy program a gépeden. Mindenki a saját adatbázisát építi
          fel: a videók, a leiratok és a kivágott részletek is nálad maradnak. Nincs feltöltés,
          nincs fiókhoz kötött tárhely.
        </p>

        <ul className="privacy-points">
          <li>
            <span className="privacy-point-title">Saját adatbázis</span>
            Egyetlen SQLite fájl az alkalmazás adatkönyvtárában. Le tudod másolni, el tudod menteni,
            ki tudod törölni - a tiéd.
          </li>
          <li>
            <span className="privacy-point-title">A fájljaid a helyükön maradnak</span>
            A mappából importált filmeket nem másolja sehova, csak hivatkozik rájuk. Egy 40 GB-os
            gyűjteménytől sem hízik meg a program.
          </li>
          <li>
            <span className="privacy-point-title">A feldolgozás is helyben fut</span>
            A leirat a te processzorodon készül (Whisper), nem egy felhős API-n. A hanganyag nem
            hagyja el a gépet.
          </li>
          <li>
            <span className="privacy-point-title">A megosztás a te döntésed</span>
            A kivágás eredménye egy fájl a lemezeden. Nincs automatikus megosztó link, amit valaki
            más is elérhetne.
          </li>
        </ul>

        <p className="privacy-note">
          Amit az alkalmazás <em>küld</em>: a hirdetéseket letölti, és ha bekapcsolod, a hibákat
          jelenti nekünk - ezekben nincs benne se videó, se leirat, se az, amit kerestél. A fiók a
          weboldalon csak a későbbi extrákhoz kell; az alkalmazás bejelentkezés nélkül is teljes
          értékű.
        </p>
      </div>
    </section>
  );
}
