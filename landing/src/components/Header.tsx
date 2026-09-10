import { Link } from 'react-router-dom';
import { useAuth } from '../AuthProvider';
import Logo from './Logo';
import SectionLink from './SectionLink';

export default function Header() {
  const { ready, authenticated, profile, login, register } = useAuth();

  return (
    <header className="site-header">
      <div className="container header-inner">
        <Link to="/" className="brand" aria-label="Snitt főoldal">
          <Logo size={32} />
          <span className="brand-name">Snitt</span>
        </Link>

        <nav className="site-nav" aria-label="Fő navigáció">
          <SectionLink to="hogyan">Hogyan működik</SectionLink>
          <SectionLink to="funkciok">Funkciók</SectionLink>
          <SectionLink to="letoltes">Letöltés</SectionLink>
          <SectionLink to="gyik">GYIK</SectionLink>
        </nav>

        <div className="header-actions">
          {authenticated && profile ? (
            <Link to="/profil" className="btn btn-ghost user-chip">
              <span className="avatar" aria-hidden="true">
                {profile.name.slice(0, 1).toUpperCase()}
              </span>
              <span className="user-name">{profile.name}</span>
            </Link>
          ) : (
            <>
              <button type="button" className="btn btn-ghost" onClick={login} disabled={!ready}>
                Bejelentkezés
              </button>
              <button
                type="button"
                className="btn btn-primary btn-sm-hide"
                onClick={register}
                disabled={!ready}
              >
                Regisztráció
              </button>
            </>
          )}
        </div>
      </div>
    </header>
  );
}
