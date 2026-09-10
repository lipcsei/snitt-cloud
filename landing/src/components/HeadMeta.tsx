import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { PATHS, absoluteUrl, routeKeyFromPathname, useI18n } from '../i18n';
import { hu } from '../i18n/hu';
import { en } from '../i18n/en';

/**
 * A dokumentumfej nyelvfüggő része. Nem renderel semmit, csak a <html lang>-ot,
 * a címet, a leírást és a nyelvi változatokra mutató linkeket tartja szinkronban
 * az útvonallal - így az angol oldal saját, indexelhető URL-en él.
 */
export default function HeadMeta() {
  const { lang, t } = useI18n();
  const { pathname } = useLocation();
  const routeKey = routeKeyFromPathname(pathname);
  const { title, description } = t.meta[routeKey];
  const htmlLang = t.htmlLang;

  useEffect(() => {
    document.documentElement.lang = htmlLang;
    document.title = title;
    setMeta('description', description);

    // Mindent kitakarítunk, mielőtt kiírnánk az újakat: a korábbi oldal futásidejű
    // linkjeit és a build által beégetett statikus párjukat is - különben duplán
    // szerepelne a canonical és a hreflang.
    document.head
      .querySelectorAll('link[data-i18n], link[rel="canonical"], link[rel="alternate"][hreflang]')
      .forEach((node) => node.remove());
    addLink('alternate', absoluteUrl(PATHS.hu[routeKey]), hu.htmlLang);
    addLink('alternate', absoluteUrl(PATHS.en[routeKey]), en.htmlLang);
    // A nyelvsemleges belépési pont a magyar URL: onnan irányítunk tovább.
    addLink('alternate', absoluteUrl(PATHS.hu[routeKey]), 'x-default');
    addLink('canonical', absoluteUrl(PATHS[lang][routeKey]));
  }, [lang, htmlLang, title, description, routeKey]);

  return null;
}

function setMeta(name: string, content: string) {
  let tag = document.head.querySelector<HTMLMetaElement>(`meta[name="${name}"]`);
  if (!tag) {
    tag = document.createElement('meta');
    tag.name = name;
    document.head.appendChild(tag);
  }
  tag.content = content;
}

function addLink(rel: string, href: string, hreflang?: string) {
  const link = document.createElement('link');
  link.rel = rel;
  link.href = href;
  if (hreflang) link.hreflang = hreflang;
  link.dataset.i18n = 'true';
  document.head.appendChild(link);
}
