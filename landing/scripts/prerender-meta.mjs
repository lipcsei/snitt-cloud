// Build utáni lépés: útvonalanként külön HTML fájl, a helyes nyelvi
// metaadatokkal beégetve.
//
// Miért kell: az oldal egyoldalas alkalmazás, a fejlécet a HeadMeta futásidőben
// állítja be - a közösségi média előnézet-botjai (Facebook, LinkedIn, Slack)
// viszont NEM futtatnak JavaScriptet. Nélküle egy megosztott angol link magyar
// címmel és leírással jelenne meg.
//
// A szövegek forrása ugyanaz a szótár, amit az alkalmazás használ (esbuilddel
// töltjük be a TypeScriptet), így nem tud eltérni a kettő.

import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { dirname, join } from 'node:path';
import { loadStrings } from './load-strings.mjs';

const DIST = 'dist';

/** Az oldal nyilvános origója - ebből lesz minden abszolút URL (og:url, og:image). */
const SITE_URL = (process.env.SITE_URL || 'https://snitt.video').replace(/\/$/, '');

/**
 * Az alapútvonal, ahogy a Vite is kapja: saját domainnél "/", GitHub Pages
 * alkönyvtárnál "/<repo>/". Enélkül a megosztási kép URL-je 404 lenne.
 */
const BASE = `/${(process.env.VITE_BASE || '/').replace(/^\/|\/$/g, '')}/`.replace('//', '/');

/** Ha egyszer lesz Facebook-alkalmazás, elég a repository variable-t kitölteni. */
const FB_APP_ID = process.env.FB_APP_ID || '';
/** Az X/Twitter fiók, ha lesz - pl. "@snittvideo". */
const TWITTER_SITE = process.env.TWITTER_SITE || '';

const THEME_COLOR = '#0a0b0f';

/** Egy oldalon belüli útvonalból teljes, abszolút URL. */
function absUrl(path) {
  const prefix = SITE_URL + BASE.replace(/\/$/, '');
  return path === '/' ? `${prefix}/` : prefix + path;
}

/** A public/ alatti fájlokból teljes, abszolút URL (a közösségi botok relatívat nem fogadnak el). */
function assetUrl(name) {
  return SITE_URL + BASE + name;
}

/** A nyelvek, ahogy az alkalmazásban is: a magyar a nyelvsemleges alap. */
const LANGS = ['hu', 'en', 'de'];
const DEFAULT_LANG = 'hu';

/** A hreflang mellé az og:locale is nyelvenkénti. */
const OG_LOCALE = { hu: 'hu_HU', en: 'en_US', de: 'de_DE' };

/** Az útvonalak, amiknek van értelme önálló, indexelhető HTML-t adni. */
const ROUTES = [
  { key: 'home', hu: '/', en: '/en', de: '/de' },
  { key: 'install', hu: '/telepites', en: '/en/install', de: '/de/installation' },
];

function head({ lang, code, ogLocale, title, description, canonical, urls, imageAlt }) {
  const image = assetUrl(`og/snitt-${code}.png`);
  return `<html lang="${lang}">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta name="color-scheme" content="dark" />
    <meta name="theme-color" content="${THEME_COLOR}" />
    <title>${esc(title)}</title>
    <meta name="description" content="${esc(description)}" />
    <meta name="author" content="Snitt" />
    <meta name="application-name" content="Snitt" />
    <meta name="apple-mobile-web-app-title" content="Snitt" />
    <meta name="robots" content="index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1" />

    <!-- Open Graph: ezt olvassa a Facebook, a Messenger, a LinkedIn és a Slack is. -->
    <meta property="og:type" content="website" />
    <meta property="og:site_name" content="Snitt" />
    <meta property="og:locale" content="${ogLocale}" />
${LANGS.filter((other) => other !== code)
  .map((other) => `    <meta property="og:locale:alternate" content="${OG_LOCALE[other]}" />`)
  .join('\n')}
    <meta property="og:title" content="${esc(title)}" />
    <meta property="og:description" content="${esc(description)}" />
    <meta property="og:url" content="${canonical}" />
    <meta property="og:image" content="${image}" />
    <meta property="og:image:secure_url" content="${image}" />
    <meta property="og:image:type" content="image/png" />
    <meta property="og:image:width" content="1200" />
    <meta property="og:image:height" content="630" />
    <meta property="og:image:alt" content="${esc(imageAlt)}" />
${FB_APP_ID ? `    <meta property="fb:app_id" content="${FB_APP_ID}" />\n` : ''}
    <!-- X/Twitter kártya: külön névtér, de ugyanaz a kép és szöveg. -->
    <meta name="twitter:card" content="summary_large_image" />
${TWITTER_SITE ? `    <meta name="twitter:site" content="${TWITTER_SITE}" />\n    <meta name="twitter:creator" content="${TWITTER_SITE}" />\n` : ''}    <meta name="twitter:title" content="${esc(title)}" />
    <meta name="twitter:description" content="${esc(description)}" />
    <meta name="twitter:image" content="${image}" />
    <meta name="twitter:image:alt" content="${esc(imageAlt)}" />

    <link rel="canonical" href="${canonical}" />
${LANGS.map((other) => `    <link rel="alternate" hreflang="${TABLE_LANG[other]}" href="${urls[other]}" />`).join('\n')}
    <link rel="alternate" hreflang="x-default" href="${urls[DEFAULT_LANG]}" />
    <link rel="icon" type="image/svg+xml" href="${BASE}favicon.svg" />
    <link rel="icon" type="image/png" sizes="32x32" href="${BASE}favicon-32.png" />
    <link rel="apple-touch-icon" sizes="180x180" href="${BASE}apple-touch-icon.png" />
    <link rel="manifest" href="${BASE}site.webmanifest" />
    <script type="application/ld+json">${jsonLd({ title, description, canonical, image })}</script>`;
}

/**
 * Strukturált adat a keresőknek: maga az oldal és a letölthető alkalmazás.
 * A Facebooknak nem kell, a Google gazdag találatához viszont igen.
 */
function jsonLd({ title, description, canonical, image }) {
  return JSON.stringify([
    {
      '@context': 'https://schema.org',
      '@type': 'WebSite',
      name: 'Snitt',
      url: absUrl('/'),
      inLanguage: LANGS.map((code) => TABLE_LANG[code]),
    },
    {
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      name: 'Snitt',
      applicationCategory: 'MultimediaApplication',
      operatingSystem: 'Windows, macOS, Linux',
      url: canonical,
      image,
      description,
      headline: title,
      offers: { '@type': 'Offer', price: '0', priceCurrency: 'EUR' },
    },
  ]);
}

function esc(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

const tables = await loadStrings(LANGS);

/** A `htmlLang` értékek nyelvkódonként - a hreflang és a JSON-LD is ezt használja. */
const TABLE_LANG = Object.fromEntries(LANGS.map((code) => [code, tables[code].htmlLang]));
const template = await readFile(join(DIST, 'index.html'), 'utf8');

// A Vite által generált fejlécből csak a <script>/<link rel=stylesheet>
// sorokat vesszük át - a metaadatokat mi írjuk.
const assetTags = [...template.matchAll(/<(script|link)[^>]*>(<\/script>)?/g)]
  .map((m) => m[0])
  .filter((tag) => tag.includes('/assets/') || tag.includes('crossorigin'))
  .join('\n    ');

for (const route of ROUTES) {
  const urls = Object.fromEntries(LANGS.map((code) => [code, absUrl(route[code])]));

  for (const lang of LANGS) {
    const strings = tables[lang];
    const { title, description } = strings.meta[route.key];
    const canonical = urls[lang];
    const html = `<!doctype html>
${head({
      lang: strings.htmlLang,
      code: lang,
      ogLocale: OG_LOCALE[lang],
      title,
      description,
      canonical,
      urls,
      imageAlt: strings.og.imageAlt,
    })}
    ${assetTags}
  </head>
  <body>
    <div id="root"></div>
  </body>
</html>
`;
    const outPath = route[lang] === '/' ? join(DIST, 'index.html') : join(DIST, route[lang], 'index.html');
    await mkdir(dirname(outPath), { recursive: true });
    await writeFile(outPath, html);
    console.log('prerender:', route[lang], `(${strings.htmlLang})`);
  }
}

// A 404-re a magyar kezdőlap megy: ismeretlen útvonalnál az alkalmazás
// kliensoldalon dönti el, hova tartozik.
await writeFile(join(DIST, '404.html'), await readFile(join(DIST, 'index.html'), 'utf8'));
console.log('prerender: 404.html');

// Sitemap az indexelhető útvonalakról, a nyelvi párokkal együtt - a kereső így
// nem külön oldalnak látja a három fordítást, hanem egymás változatainak.
const sitemap = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">
${ROUTES.flatMap((route) => {
  const urls = Object.fromEntries(LANGS.map((code) => [code, absUrl(route[code])]));
  const alternates = [
    ...LANGS.map((code) => `    <xhtml:link rel="alternate" hreflang="${TABLE_LANG[code]}" href="${urls[code]}"/>`),
    `    <xhtml:link rel="alternate" hreflang="x-default" href="${urls[DEFAULT_LANG]}"/>`,
  ].join('\n');
  return LANGS.map(
    (code) => `  <url>\n    <loc>${urls[code]}</loc>\n${alternates}\n  </url>`,
  );
}).join('\n')}
</urlset>
`;
await writeFile(join(DIST, 'sitemap.xml'), sitemap);
console.log('prerender: sitemap.xml');

await writeFile(
  join(DIST, 'robots.txt'),
  `User-agent: *\nAllow: /\n\nSitemap: ${assetUrl('sitemap.xml')}\n`,
);
console.log('prerender: robots.txt');
