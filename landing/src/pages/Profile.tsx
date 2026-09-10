import { useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../AuthProvider';

export default function Profile() {
  const { ready, authenticated, profile, login, logout, accountUrl } = useAuth();

  // Védett oldal: amint kiderül, hogy nincs session, megyünk a bejelentkezésre.
  useEffect(() => {
    if (ready && !authenticated) login();
  }, [ready, authenticated, login]);

  if (!ready) {
    return (
      <main className="page">
        <div className="container container-narrow">
          <p className="muted">Bejelentkezés ellenőrzése…</p>
        </div>
      </main>
    );
  }

  if (!authenticated || !profile) {
    return (
      <main className="page">
        <div className="container container-narrow">
          <h1>Bejelentkezés szükséges</h1>
          <p className="muted">
            Ez az oldal csak bejelentkezve érhető el. Ha az átirányítás nem indul el magától:
          </p>
          <div className="account-cta">
            <button type="button" className="btn btn-primary" onClick={login}>
              Bejelentkezés
            </button>
            <Link to="/" className="btn btn-outline">
              Vissza a főoldalra
            </Link>
          </div>
        </div>
      </main>
    );
  }

  return (
    <main className="page">
      <div className="container container-narrow">
        <span className="eyebrow">Profil</span>
        <h1>Szia, {profile.name}!</h1>
        <p className="muted">
          Ez a fiók a weboldalhoz tartozik. Az asztali alkalmazás használatához nincs rá szükség —
          a készülő extra funkciókhoz lesz.
        </p>

        <div className="panel profile-panel">
          <dl className="profile-data">
            <div>
              <dt>Név</dt>
              <dd>{profile.name}</dd>
            </div>
            <div>
              <dt>E-mail</dt>
              <dd>{profile.email || <span className="muted">nincs megadva</span>}</dd>
            </div>
          </dl>

          <div className="profile-actions">
            <a
              className="btn btn-primary"
              href={accountUrl}
              target="_blank"
              rel="noreferrer noopener"
            >
              Fiók kezelése
            </a>
            <button type="button" className="btn btn-outline" onClick={logout}>
              Kijelentkezés
            </button>
          </div>

          <p className="profile-hint">
            A „Fiók kezelése” a Keycloak fiókkonzolját nyitja meg: ott tudsz jelszót változtatni,
            kétlépcsős azonosítást beállítani vagy a fiókodat törölni.
          </p>
        </div>

        <Link to="/" className="back-link">
          ← Vissza a főoldalra
        </Link>
      </div>
    </main>
  );
}
