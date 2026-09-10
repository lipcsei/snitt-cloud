import { useAuth } from '../AuthProvider';
import { useNotYet } from '../NotYetProvider';
import { useT } from '../i18n';
import Rich from '../i18n/Rich';
import '../styles.pricing.css';

/**
 * Az árazás azért ilyen szűkszavú, mert a lényeg egy mondat: az alapfunkció
 * ingyenes marad, a fizetős rész az, ami valódi AI-költséget jelent nekünk.
 */
export default function Pricing() {
  const { authenticated, ready } = useAuth();
  const showNotYet = useNotYet();
  const t = useT();

  return (
    <section className="pricing" id={t.sections.pricing}>
      <div className="pricing-inner">
        <h2>{t.pricing.title}</h2>
        <p className="pricing-lead">{t.pricing.lead}</p>

        <div className="pricing-tiers">
          <article className="tier">
            <header>
              <h3>{t.pricing.free.name}</h3>
              <p className="tier-price">
                {t.pricing.free.price} <span>{t.pricing.free.priceNote}</span>
              </p>
            </header>
            <ul>
              {t.pricing.free.items.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
            <p className="tier-note">{t.pricing.free.note}</p>
          </article>

          <article className="tier tier-pro">
            <span className="tier-flag">{t.pricing.pro.flag}</span>
            <header>
              <h3>{t.pricing.pro.name}</h3>
              <p className="tier-price">
                {t.pricing.pro.price} <span>{t.pricing.pro.priceNote}</span>
              </p>
            </header>
            <ul>
              {t.pricing.pro.items.map((item) => (
                <li key={item}>
                  <Rich text={item} />
                </li>
              ))}
            </ul>
            <p className="tier-note">{t.pricing.pro.note}</p>
            {ready && !authenticated && (
              <button type="button" className="tier-cta" onClick={() => showNotYet('register')}>
                {t.pricing.pro.cta}
              </button>
            )}
          </article>
        </div>

        <p className="pricing-footnote">{t.pricing.footnote}</p>
      </div>
    </section>
  );
}
