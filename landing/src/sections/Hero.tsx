import { useAuth } from '../AuthProvider';
import AppPreview from '../components/AppPreview';
import { useNotYet } from '../NotYetProvider';
import { useT } from '../i18n';
import Rich from '../i18n/Rich';

export default function Hero() {
  const { authenticated } = useAuth();
  const showNotYet = useNotYet();
  const t = useT();

  return (
    <section className="hero">
      <div className="container hero-inner">
        <div className="hero-copy">
          <span className="pill">
            <span className="pill-dot" />
            {t.hero.pill}
          </span>

          <h1>
            <Rich text={t.hero.title} />
          </h1>

          <p className="lead">{t.hero.lead}</p>

          <div className="hero-cta">
            <button
              type="button"
              className="btn btn-primary btn-lg"
              onClick={() => showNotYet('download')}
            >
              {t.hero.download}
            </button>
            {!authenticated && (
              <button
                type="button"
                className="btn btn-outline btn-lg"
                onClick={() => showNotYet('register')}
              >
                {t.hero.register}
              </button>
            )}
          </div>

          <ul className="hero-facts">
            {t.hero.facts.map((fact) => (
              <li key={fact}>{fact}</li>
            ))}
          </ul>
        </div>

        <div className="hero-visual">
          <AppPreview />
        </div>
      </div>
    </section>
  );
}
