import { Link } from 'react-router-dom';
import { useT } from '../../i18n';
import { useAnalyticsConsent } from './AnalyticsConsentProvider';
import '../../styles/cookie-banner.css';

/**
 * Nem-modális, alul rögzített cookie-sáv. Csak akkor jelenik meg, ha ebben a buildben van GA
 * mérési azonosító (`VITE_GA_MEASUREMENT_ID`) ÉS a felhasználó még nem döntött. A két gomb
 * egyforma hangsúlyú: az elutasítás nem kevésbé elérhető, mint az elfogadás.
 */
export default function ConsentBanner() {
  const t = useT();
  const { available, consent, setConsent } = useAnalyticsConsent();

  if (!available || consent !== null) return null;

  return (
    <section className="cookie-bar" aria-labelledby="cookie-bar-text">
      <div className="container cookie-bar-inner">
        <p id="cookie-bar-text" className="cookie-bar-text">
          {t.cookieBanner.message}{' '}
          <Link to="/adatvedelem" className="cookie-bar-link">
            {t.cookieBanner.link}
          </Link>
        </p>
        <div className="cookie-bar-actions">
          <button type="button" className="btn btn-primary" onClick={() => setConsent('granted')}>
            {t.cookieBanner.accept}
          </button>
          <button type="button" className="btn btn-outline" onClick={() => setConsent('denied')}>
            {t.cookieBanner.reject}
          </button>
        </div>
      </div>
    </section>
  );
}
