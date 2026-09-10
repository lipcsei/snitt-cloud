import { Link } from 'react-router-dom';
import Logo from './Logo';
import SectionLink from './SectionLink';
import { RELEASES_URL, REPO_URL } from '../constants';

export default function Footer() {
  return (
    <footer className="site-footer">
      <div className="container footer-inner">
        <div className="footer-brand">
          <Link to="/" className="brand" aria-label="Snitt főoldal">
            <Logo size={28} />
            <span className="brand-name">Snitt</span>
          </Link>
          <p className="footer-tagline">
            Keresés a saját videótáradban — idézet alapján, a saját gépeden.
          </p>
        </div>

        <nav className="footer-links" aria-label="Lábléc navigáció">
          <SectionLink to="hogyan">Hogyan működik</SectionLink>
          <SectionLink to="funkciok">Funkciók</SectionLink>
          <SectionLink to="letoltes">Letöltés</SectionLink>
          <SectionLink to="fiok">Fiók</SectionLink>
          <SectionLink to="gyik">GYIK</SectionLink>
          <a href={REPO_URL} target="_blank" rel="noreferrer noopener">
            Forráskód
          </a>
          <a href={RELEASES_URL} target="_blank" rel="noreferrer noopener">
            Kiadások
          </a>
        </nav>
      </div>

      <div className="container footer-bottom">
        <span>© {new Date().getFullYear()} Snitt</span>
        <span>Minden feldolgozás helyben fut. A videótárad nem hagyja el a gépedet.</span>
      </div>
    </footer>
  );
}
