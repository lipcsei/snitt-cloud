import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { DEFAULT_LANG, LANGS, PATHS, TABLES, absoluteUrl, routeKeyFromPathname, useI18n } from '../i18n';
import type { Lang } from '../i18n';

/** Azok az útvonalak, amikhez a build előrendelt, indexelhető HTML-t készít. */
const PRERENDERED = new Set(['home', 'install']);

/** Az og:locale értékek nyelvkódonként - a Facebook ezt a formátumot várja. */
const OG_LOCALE: Record<Lang, string> = { hu: 'hu_HU', en: 'en_US', de: 'de_DE' };

/**
 * A dokumentumfej nyelvfüggő része. Nem renderel semmit, csak a <html lang>-ot,
 * a címet, a leírást, a megosztási metaadatokat és a nyelvi változatokra mutató
 * linkeket tartja szinkronban az útvonallal.
 *
 * A közösségi botok (Facebook, LinkedIn, Slack) nem futtatnak JavaScriptet, ezért
 * az igazi forrás a build-időben előrendelt HTML (scripts/prerender-meta.mjs).
 * Ez itt azért kell, hogy oldalon belüli navigáció után se maradjon ott az előző
 * oldal címe - se a fülön, se a böngésző megosztás funkciójában.
 */
export default function HeadMeta() {
  const { lang, t } = useI18n();
  const { pathname } = useLocation();
  const routeKey = routeKeyFromPathname(pathname);
  const { title, description } = t.meta[routeKey];
  const htmlLang = t.htmlLang;
  const imageAlt = t.og.imageAlt;

  useEffect(() => {
    const canonical = absoluteUrl(PATHS[lang][routeKey]);
    const image = absoluteUrl(`/og/snitt-${lang}.png`);

    document.documentElement.lang = htmlLang;
    document.title = title;

    setMeta('name', 'description', description);
    // A profil oldalnak nincs előrendelt HTML-je, és nincs is mit indexelni
    // rajta: bejelentkezés nélkül üres, bejelentkezve személyes.
    setMeta(
      'name',
      'robots',
      PRERENDERED.has(routeKey)
        ? 'index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1'
        : 'noindex, follow',
    );

    setMeta('property', 'og:title', title);
    setMeta('property', 'og:description', description);
    setMeta('property', 'og:url', canonical);
    setMeta('property', 'og:locale', OG_LOCALE[lang]);
    setMeta('property', 'og:image', image);
    setMeta('property', 'og:image:secure_url', image);
    setMeta('property', 'og:image:alt', imageAlt);

    setMeta('name', 'twitter:title', title);
    setMeta('name', 'twitter:description', description);
    setMeta('name', 'twitter:image', image);
    setMeta('name', 'twitter:image:alt', imageAlt);

    // Az og:locale:alternate többször szerepel, ezért nem frissítjük, hanem
    // újraírjuk a teljes halmazt.
    document.head
      .querySelectorAll('meta[property="og:locale:alternate"]')
      .forEach((node) => node.remove());
    for (const code of LANGS) {
      if (code !== lang) addMeta('property', 'og:locale:alternate', OG_LOCALE[code]);
    }

    // Mindent kitakarítunk, mielőtt kiírnánk az újakat: a korábbi oldal futásidejű
    // linkjeit és a build által beégetett statikus párjukat is - különben duplán
    // szerepelne a canonical és a hreflang.
    document.head
      .querySelectorAll('link[data-i18n], link[rel="canonical"], link[rel="alternate"][hreflang]')
      .forEach((node) => node.remove());
    for (const code of LANGS) {
      addLink('alternate', absoluteUrl(PATHS[code][routeKey]), TABLES[code].htmlLang);
    }
    // A nyelvsemleges belépési pont a magyar URL: onnan irányítunk tovább.
    addLink('alternate', absoluteUrl(PATHS[DEFAULT_LANG][routeKey]), 'x-default');
    addLink('canonical', canonical);
  }, [lang, htmlLang, title, description, imageAlt, routeKey]);

  return null;
}

/**
 * Egy meta tag beállítása. A kulcs attribútuma számít: a szabványos meta tagek
 * `name`-mel, az Open Graph tagek `property`-vel azonosítják magukat, és a
 * Facebook csak az utóbbit fogadja el.
 */
function setMeta(attr: 'name' | 'property', key: string, content: string) {
  const tag =
    document.head.querySelector<HTMLMetaElement>(`meta[${attr}="${key}"]`) ??
    addMeta(attr, key, content);
  tag.content = content;
}

function addMeta(attr: 'name' | 'property', key: string, content: string): HTMLMetaElement {
  const tag = document.createElement('meta');
  tag.setAttribute(attr, key);
  tag.content = content;
  document.head.appendChild(tag);
  return tag;
}

function addLink(rel: string, href: string, hreflang?: string) {
  const link = document.createElement('link');
  link.rel = rel;
  link.href = href;
  if (hreflang) link.hreflang = hreflang;
  link.dataset.i18n = 'true';
  document.head.appendChild(link);
}
