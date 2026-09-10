import { useAuth } from '../AuthProvider';
import '../styles.pricing.css';

/**
 * Az árazás azért ilyen szűkszavú, mert a lényeg egy mondat: az alapfunkció
 * ingyenes marad, a fizetős rész az, ami valódi AI-költséget jelent nekünk.
 */
export default function Pricing() {
  const { authenticated, register, ready } = useAuth();

  return (
    <section className="pricing" id="arak">
      <div className="pricing-inner">
        <h2>Mi ingyenes, és miért van fizetős rész?</h2>
        <p className="pricing-lead">
          A keresés és a kivágás ingyenes marad - ez a gépeden fut, nekünk nem kerül pénzbe. Az
          AI-alapú funkciók viszont valódi tokenköltséget jelentenek, ezért csak előfizetéssel
          érhetők el. Nincs benne trükk.
        </p>

        <div className="pricing-tiers">
          <article className="tier">
            <header>
              <h3>Ingyenes</h3>
              <p className="tier-price">
                0 Ft <span>örökre</span>
              </p>
            </header>
            <ul>
              <li>Korlátlan videó és leirat a saját gépeden</li>
              <li>Idézet-alapú keresés az összes feliratban</li>
              <li>Whisper átirat a tényleges hangsávból</li>
              <li>Több nyelvű leiratok filmenként</li>
              <li>Klipvágás fájlba, kézzel állítható tartománnyal</li>
            </ul>
            <p className="tier-note">
              Egy kivágás egy rövid hirdetés megtekintésével jár - ebből tartjuk fenn a fejlesztést.
            </p>
          </article>

          <article className="tier tier-pro">
            <span className="tier-flag">Előfizetés</span>
            <header>
              <h3>Pro</h3>
              <p className="tier-price">
                hamarosan <span>havidíjas</span>
              </p>
            </header>
            <ul>
              <li>
                <strong>AI-keresés:</strong> nem csak a szavakat találja meg, hanem a jelenetet is,
                amire gondolsz - körülírásból is
              </li>
              <li>
                <strong>Videóértelmezés:</strong> mi történik a jelenetben, kik szerepelnek benne,
                miről szól a párbeszéd
              </li>
              <li>Kérdezhetsz a videótáradtól, nem csak kereshetsz benne</li>
              <li>Nincsenek hirdetések</li>
            </ul>
            <p className="tier-note">
              Ezek a funkciók külső AI-modelleket használnak, ami tokenenként fizetendő - ezért nem
              fér bele az ingyenes csomagba. Az árazás a bevezetéskor lesz végleges.
            </p>
            {ready && !authenticated && (
              <button type="button" className="tier-cta" onClick={register}>
                Regisztrálok, szóljatok, ha indul
              </button>
            )}
          </article>
        </div>

        <p className="pricing-footnote">
          Az alkalmazás fiók nélkül is teljes értékű. A regisztráció csak a Pro funkciókhoz kell -
          a videóid akkor sem kerülnek fel sehova.
        </p>
      </div>
    </section>
  );
}
