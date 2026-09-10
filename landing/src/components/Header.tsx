import { Link } from 'react-router-dom';
import { useAuth } from '../AuthProvider';
import Logo from './Logo';
import SectionLink from './SectionLink';
import LanguageSwitcher from './LanguageSwitcher';
import { useNotYet } from '../NotYetProvider';
import { useI18n } from '../i18n';

export default function Header() {
  const { ready, authenticated, profile, login } = useAuth();
  const showNotYet = useNotYet();
  const { t, path } = useI18n();

  return (
    <header className="site-header">
      <div className="container header-inner">
        <Link to={path('home')} className="brand" aria-label={t.header.brandAria}>
          <Logo size={32} />
          <span className="brand-name">Snitt</span>
        </Link>

        <nav className="site-nav" aria-label={t.header.navAria}>
          <SectionLink to={t.sections.how}>{t.header.navHow}</SectionLink>
          <SectionLink to={t.sections.features}>{t.header.navFeatures}</SectionLink>
          <SectionLink to={t.sections.downloads}>{t.header.navDownloads}</SectionLink>
          <SectionLink to={t.sections.faq}>{t.header.navFaq}</SectionLink>
        </nav>

        <div className="header-actions">
          <LanguageSwitcher />
          {authenticated && profile ? (
            <Link to={path('profile')} className="btn btn-ghost user-chip">
              <span className="avatar" aria-hidden="true">
                {profile.name.slice(0, 1).toUpperCase()}
              </span>
              <span className="user-name">{profile.name}</span>
            </Link>
          ) : (
            <>
              <button type="button" className="btn btn-ghost" onClick={login} disabled={!ready}>
                {t.header.login}
              </button>
              <button
                type="button"
                className="btn btn-primary btn-sm-hide"
                onClick={() => showNotYet('register')}
              >
                {t.header.register}
              </button>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
