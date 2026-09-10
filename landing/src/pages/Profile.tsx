import { Link } from 'react-router-dom';
import { useAuth } from '../AuthProvider';
import { useNotYet } from '../NotYetProvider';
import { fill, useI18n } from '../i18n';

export default function Profile() {
  const { ready, authenticated, profile, logout, accountUrl } = useAuth();
  const { t, path } = useI18n();
  const showNotYet = useNotYet();

  // Amíg a fiókok nincsenek élesben, ez az oldal NEM irányít automatikusan a
  // Keycloakra: az csak helyben létezik, a látogató kapcsolódási hibát kapna.

  if (!ready) {
    return (
      <main className="page">
        <div className="container container-narrow">
          <p className="muted">{t.profile.checking}</p>
        </div>
      </main>
    );
  }

  if (!authenticated || !profile) {
    return (
      <main className="page">
        <div className="container container-narrow">
          <h1>{t.profile.needLoginTitle}</h1>
          <p className="muted">{t.profile.needLoginBody}</p>
          <div className="account-cta">
            <button type="button" className="btn btn-primary" onClick={() => showNotYet('login')}>
              {t.profile.login}
            </button>
            <Link to={path('home')} className="btn btn-outline">
              {t.profile.backHome}
            </Link>
          </div>
        </div>
      </main>
    );
  }

  return (
    <main className="page">
      <div className="container container-narrow">
        <span className="eyebrow">{t.profile.eyebrow}</span>
        <h1>{fill(t.profile.greeting, { name: profile.name })}</h1>
        <p className="muted">{t.profile.intro}</p>

        <div className="panel profile-panel">
          <dl className="profile-data">
            <div>
              <dt>{t.profile.nameLabel}</dt>
              <dd>{profile.name}</dd>
            </div>
            <div>
              <dt>{t.profile.emailLabel}</dt>
              <dd>{profile.email || <span className="muted">{t.profile.noEmail}</span>}</dd>
            </div>
          </dl>

          <div className="profile-actions">
            <a
              className="btn btn-primary"
              href={accountUrl}
              target="_blank"
              rel="noreferrer noopener"
            >
              {t.profile.manage}
            </a>
            <button type="button" className="btn btn-outline" onClick={logout}>
              {t.profile.logout}
            </button>
          </div>

          <p className="profile-hint">{t.profile.hint}</p>
        </div>

        <Link to={path('home')} className="back-link">
          {t.common.backToHome}
        </Link>
      </div>
    </main>
  );
}
