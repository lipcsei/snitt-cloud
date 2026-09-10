import { Link, useLocation } from 'react-router-dom';
import { LANGS, PATHS, rememberLang, routeKeyFromPathname, useI18n } from '../i18n';
import type { Lang } from '../i18n';

/**
 * HU / EN / DE váltó. Az aktuális oldal másik nyelvű párjára visz (nem a
 * főoldalra), és megjegyzi a választást, hogy a "/" ne tippelgessen legközelebb.
 */
export default function LanguageSwitcher() {
  const { lang, t } = useI18n();
  const { pathname } = useLocation();
  const routeKey = routeKeyFromPathname(pathname);

  const labels: Record<Lang, string> = {
    hu: t.header.langHu,
    en: t.header.langEn,
    de: t.header.langDe,
  };
  const titles: Record<Lang, string> = {
    hu: t.header.langHuTitle,
    en: t.header.langEnTitle,
    de: t.header.langDeTitle,
  };

  return (
    <nav className="lang-switch" aria-label={t.header.langAria}>
      {LANGS.map((item) => (
        <Link
          key={item}
          to={PATHS[item][routeKey]}
          className={item === lang ? 'lang-opt lang-opt-on' : 'lang-opt'}
          lang={item}
          hrefLang={item}
          title={titles[item]}
          aria-current={item === lang ? 'true' : undefined}
          onClick={() => rememberLang(item)}
        >
          {labels[item]}
        </Link>
      ))}
    </nav>
  );
}
