import { Link } from 'react-router-dom';
import Logo from './Logo';
import SectionLink from './SectionLink';
import { RELEASES_URL, REPO_URL } from '../constants';
import { useI18n } from '../i18n';

export default function Footer() {
  const { t, path } = useI18n();

  return (
    <footer className="site-footer">
      <div className="container footer-inner">
        <div className="footer-brand">
          <Link to={path('home')} className="brand" aria-label={t.header.brandAria}>
            <Logo size={28} />
            <span className="brand-name">Snitt</span>
          </Link>
          <p className="footer-tagline">{t.footer.tagline}</p>
        </div>

        <nav className="footer-links" aria-label={t.footer.navAria}>
          <SectionLink to={t.sections.how}>{t.header.navHow}</SectionLink>
          <SectionLink to={t.sections.features}>{t.header.navFeatures}</SectionLink>
          <SectionLink to={t.sections.downloads}>{t.header.navDownloads}</SectionLink>
          <SectionLink to={t.sections.account}>{t.account.eyebrow}</SectionLink>
          <SectionLink to={t.sections.faq}>{t.header.navFaq}</SectionLink>
          <a href={REPO_URL} target="_blank" rel="noreferrer noopener">
            {t.footer.source}
          </a>
          <a href={RELEASES_URL} target="_blank" rel="noreferrer noopener">
            {t.footer.releases}
          </a>
        </nav>
      </div>

      <div className="container footer-bottom">
        <span>© {new Date().getFullYear()} Snitt</span>
        <span>{t.footer.bottom}</span>
      </div>
    </footer>
  );
}
