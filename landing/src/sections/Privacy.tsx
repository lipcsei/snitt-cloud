import { useT } from '../i18n';
import Rich from '../i18n/Rich';
import '../styles.privacy.css';

/**
 * A termék legfontosabb ígérete, ezért kap saját, hangsúlyos szakaszt:
 * nincs távoli szerver, mindenki a SAJÁT adatbázisát építi a saját gépén.
 */
export default function Privacy() {
  const t = useT();

  return (
    <section className="privacy" id={t.sections.privacy}>
      <div className="privacy-inner">
        <p className="privacy-eyebrow">{t.privacy.eyebrow}</p>
        <h2>
          {t.privacy.title}
          <br />
          <span className="privacy-accent">{t.privacy.titleAccent}</span>
        </h2>
        <p className="privacy-lead">{t.privacy.lead}</p>

        <ul className="privacy-points">
          {t.privacy.points.map((point) => (
            <li key={point.title}>
              <span className="privacy-point-title">{point.title}</span>
              {point.body}
            </li>
          ))}
        </ul>

        <p className="privacy-note">
          <Rich text={t.privacy.note} />
        </p>
      </div>
    </section>
  );
}
