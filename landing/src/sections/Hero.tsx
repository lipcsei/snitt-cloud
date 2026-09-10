import { useAuth } from '../AuthProvider';
import AppPreview from '../components/AppPreview';
import { useNotYet } from '../NotYetProvider';

export default function Hero() {
  const { authenticated } = useAuth();
  const showNotYet = useNotYet();

  return (
    <section className="hero">
      <div className="container hero-inner">
        <div className="hero-copy">
          <span className="pill">
            <span className="pill-dot" />
            Asztali alkalmazás · Windows · macOS · Linux
          </span>

          <h1>
            Megvan a mondat.
            <br />
            Csak azt nem tudod, <em>melyik filmben</em>.
          </h1>

          <p className="lead">
            A Snitt átiratot készít a videótáradról, és megkeresi benne a beírt idézetet — egy
            kétórás filmben másodpercek alatt. Aztán ki is vágja neked azt a jelenetet.
          </p>

          <div className="hero-cta">
            <button
              type="button"
              className="btn btn-primary btn-lg"
              onClick={() => showNotYet('download')}
            >
              Letöltés
            </button>
            {!authenticated && (
              <button
                type="button"
                className="btn btn-outline btn-lg"
                onClick={() => showNotYet('register')}
              >
                Regisztráció
              </button>
            )}
          </div>

          <ul className="hero-facts">
            <li>Nincs feltöltés, nincs felhő</li>
            <li>Egy SQLite fájl a saját gépeden</li>
            <li>Fiók nélkül is teljes értékű</li>
          </ul>
        </div>

        <div className="hero-visual">
          <AppPreview />
        </div>
      </div>
    </section>
  );
}
