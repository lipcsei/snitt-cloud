import { Link } from 'react-router-dom';
import { useAuth } from '../AuthProvider';
import { useNotYet } from '../NotYetProvider';
import { fill, useI18n } from '../i18n';
import Rich from '../i18n/Rich';

export default function Account() {
  const { authenticated, profile } = useAuth();
  const showNotYet = useNotYet();
  const { t, path } = useI18n();

  return (
    <section id={t.sections.account} className="section section-alt">
      <div className="container">
        <div className="account">
          <div className="account-copy">
            <span className="eyebrow eyebrow-pink">{t.account.eyebrow}</span>
            <h2>{t.account.title}</h2>
            <p>{t.account.body1}</p>
            <p>
              <Rich text={t.account.body2} />
            </p>

            {authenticated && profile ? (
              <div className="account-cta">
                <Link to={path('profile')} className="btn btn-primary">
                  {t.account.openProfile}
                </Link>
                <span className="account-note">
                  <Rich text={fill(t.account.signedInAs, { name: profile.name })} />
                </span>
              </div>
            ) : (
              <div className="account-cta">
                <button
                  type="button"
                  className="btn btn-primary"
                  onClick={() => showNotYet('register')}
                >
                  {t.account.register}
                </button>
                <button type="button" className="btn btn-outline" onClick={() => showNotYet('login')}>
                  {t.account.login}
                </button>
              </div>
            )}
          </div>

          <ul className="account-list">
            {t.account.list.map((item) => (
              <li key={item.title}>
                <strong>{item.title}</strong>
                <span>{item.body}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>
    </section>
  );
}
