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

import { build } from 'esbuild';
import { mkdir, readFile, rm, writeFile } from 'node:fs/promises';
import { dirname, join } from 'node:path';
import { pathToFileURL } from 'node:url';

const DIST = 'dist';
const SITE_URL = process.env.SITE_URL || 'https://snitt.video';

/** Az útvonalak, amiknek van értelme önálló, indexelhető HTML-t adni. */
const ROUTES = [
  { key: 'home', hu: '/', en: '/en' },
  { key: 'install', hu: '/telepites', en: '/en/install' },
];

async function loadStrings() {
  const outfile = join('node_modules', '.cache', 'snitt-i18n.mjs');
  await mkdir(dirname(outfile), { recursive: true });
  await build({
    entryPoints: ['src/i18n/hu.ts', 'src/i18n/en.ts'],
    bundle: true,
    format: 'esm',
    platform: 'node',
    outfile,
    logLevel: 'silent',
    // Egy fájlba fűzzük a kettőt, hogy egyetlen importtal elérhető legyen.
    stdin: undefined,
  }).catch(async () => {
    // Több belépési pont egy outfile-ba nem megy: ilyenkor egy köztes modult
    // fordítunk, ami mindkettőt újraexportálja.
    const shim = join('node_modules', '.cache', 'snitt-i18n-entry.ts');
    await writeFile(shim, `export { hu } from '${pathToFileURL(join(process.cwd(), 'src/i18n/hu.ts')).pathname}';\nexport { en } from '${pathToFileURL(join(process.cwd(), 'src/i18n/en.ts')).pathname}';\n`);
    await build({ entryPoints: [shim], bundle: true, format: 'esm', platform: 'node', outfile, logLevel: 'silent' });
    await rm(shim, { force: true });
  });
  return import(pathToFileURL(join(process.cwd(), outfile)).href);
}

function head({ lang, title, description, canonical, huUrl, enUrl }) {
  return `<html lang="${lang}">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta name="color-scheme" content="dark" />
    <title>${esc(title)}</title>
    <meta name="description" content="${esc(description)}" />
    <meta property="og:type" content="website" />
    <meta property="og:site_name" content="Snitt" />
    <meta property="og:locale" content="${lang === 'hu' ? 'hu_HU' : 'en_US'}" />
    <meta property="og:title" content="${esc(title)}" />
    <meta property="og:description" content="${esc(description)}" />
    <meta property="og:url" content="${canonical}" />
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content="${esc(title)}" />
    <meta name="twitter:description" content="${esc(description)}" />
    <link rel="canonical" href="${canonical}" />
    <link rel="alternate" hreflang="hu" href="${huUrl}" />
    <link rel="alternate" hreflang="en" href="${enUrl}" />
    <link rel="alternate" hreflang="x-default" href="${huUrl}" />
    <link rel="icon" type="image/svg+xml" href="/favicon.svg" />`;
}

function esc(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

const { hu, en } = await loadStrings();
const template = await readFile(join(DIST, 'index.html'), 'utf8');

// A Vite által generált fejlécből csak a <script>/<link rel=stylesheet>
// sorokat vesszük át - a metaadatokat mi írjuk.
const assetTags = [...template.matchAll(/<(script|link)[^>]*>(<\/script>)?/g)]
  .map((m) => m[0])
  .filter((tag) => tag.includes('/assets/') || tag.includes('crossorigin'))
  .join('\n    ');

for (const route of ROUTES) {
  for (const lang of ['hu', 'en']) {
    const strings = lang === 'hu' ? hu : en;
    const { title, description } = strings.meta[route.key];
    const canonical = SITE_URL + route[lang];
    const html = `<!doctype html>
${head({
      lang: strings.htmlLang,
      title,
      description,
      canonical,
      huUrl: SITE_URL + route.hu,
      enUrl: SITE_URL + route.en,
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
