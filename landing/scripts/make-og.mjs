// Az Open Graph előnézetképeket (1200x630) generálja nyelvenként.
//
// Miért külön szkript és nem build-lépés: a rendereléshez rsvg-convert kell,
// ami a CI futtatóján nincs fent. A kész PNG-k ezért be vannak commitolva a
// public/og alá, és csak akkor kell újragenerálni, ha a hero szövege változik:
//
//   npm run og
//
// A szöveg forrása ugyanaz a nyelvi tábla, amit az oldal használ - így a
// megosztási kép és a címsor nem tud elcsúszni egymástól.

import { execFile } from 'node:child_process';
import { mkdir, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { promisify } from 'node:util';
import { loadStrings } from './load-strings.mjs';

const run = promisify(execFile);

const LANGS = ['hu', 'en', 'de'];
const OUT_DIR = join('public', 'og');
const SVG_DIR = join('assets', 'og');

const W = 1200;
const H = 630;

/** A `*kiemelt*` jelölést bontja szét sima és kiemelt darabokra. */
function parts(line) {
  return line
    .split(/(\*[^*]+\*)/g)
    .filter(Boolean)
    .map((piece) =>
      piece.startsWith('*') && piece.endsWith('*')
        ? { text: piece.slice(1, -1), accent: true }
        : { text: piece, accent: false },
    );
}

function esc(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
}


function tspans(line) {
  return parts(line)
    .map((p) => `<tspan${p.accent ? ' fill="url(#accent)"' : ''}>${esc(p.text)}</tspan>`)
    .join('');
}

function svg({ title, sub }) {
  const [first, second = ''] = title.split('\n');
  return `<svg xmlns="http://www.w3.org/2000/svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}">
  <defs>
    <linearGradient id="mark" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%" stop-color="#5b8dff"/><stop offset="100%" stop-color="#2a5fe0"/>
    </linearGradient>
    <linearGradient id="accent" x1="0" y1="0" x2="1" y2="1">
      <stop offset="0%" stop-color="#7aa5ff"/><stop offset="100%" stop-color="#ff5c8a"/>
    </linearGradient>
    <radialGradient id="glow" cx="0.18" cy="0.06" r="0.75">
      <stop offset="0%" stop-color="#2f6df6" stop-opacity="0.30"/>
      <stop offset="100%" stop-color="#2f6df6" stop-opacity="0"/>
    </radialGradient>
    <clipPath id="t"><path d="M0 0h64v26L0 38z"/></clipPath>
    <clipPath id="b"><path d="M0 42l64-12v34H0z"/></clipPath>
  </defs>

  <rect width="${W}" height="${H}" fill="#0a0b0f"/>
  <rect width="${W}" height="${H}" fill="url(#glow)"/>
  <rect x="0" y="${H - 6}" width="${W}" height="6" fill="url(#accent)"/>

  <g transform="translate(80,64) scale(0.86)">
    <rect width="64" height="64" rx="15" fill="url(#mark)"/>
    <g clip-path="url(#t)"><g transform="translate(5,0)">
      <rect x="19" y="5" width="26" height="54" rx="3.5" fill="#fff"/>
      <rect x="22" y="9" width="3.6" height="5" rx="1.1" fill="#2a5fe0"/>
      <rect x="38.4" y="9" width="3.6" height="5" rx="1.1" fill="#2a5fe0"/>
      <rect x="22" y="19" width="3.6" height="5" rx="1.1" fill="#2a5fe0"/>
      <rect x="38.4" y="19" width="3.6" height="5" rx="1.1" fill="#2a5fe0"/>
    </g></g>
    <g clip-path="url(#b)"><g transform="translate(-5,0)">
      <rect x="19" y="5" width="26" height="54" rx="3.5" fill="#fff"/>
      <rect x="22" y="41" width="3.6" height="5" rx="1.1" fill="#2a5fe0"/>
      <rect x="38.4" y="41" width="3.6" height="5" rx="1.1" fill="#2a5fe0"/>
      <rect x="22" y="51" width="3.6" height="5" rx="1.1" fill="#2a5fe0"/>
      <rect x="38.4" y="51" width="3.6" height="5" rx="1.1" fill="#2a5fe0"/>
    </g></g>
  </g>
  <text x="158" y="112" font-family="Helvetica Neue, Helvetica, Arial, sans-serif" font-size="40" font-weight="700" fill="#e9ebf2" letter-spacing="-0.5">Snitt</text>

  <!-- xml:space="preserve": a kiemelt és a sima rész határán lévő szóközt az
       SVG különben összevonná ("Lieblingsszeneaus."). Emiatt a tspanek egy
       sorban vannak, hogy a behúzás se kerüljön bele a szövegbe. -->
  <text font-family="Helvetica Neue, Helvetica, Arial, sans-serif" font-size="76" font-weight="700" fill="#e9ebf2" letter-spacing="-2" xml:space="preserve"><tspan x="80" y="286">${tspans(first)}</tspan><tspan x="80" y="374">${tspans(second)}</tspan></text>

  <text x="80" y="448" font-family="Helvetica Neue, Helvetica, Arial, sans-serif" font-size="29" fill="#9fa4b4">${esc(sub)}</text>

  <text x="80" y="556" font-family="Helvetica Neue, Helvetica, Arial, sans-serif" font-size="30" font-weight="700" fill="#e9ebf2">snitt.video</text>
  <text x="${W - 80}" y="556" text-anchor="end" font-family="Helvetica Neue, Helvetica, Arial, sans-serif" font-size="26" fill="#6f7484">Windows · macOS · Linux</text>
</svg>
`;
}

const tables = await loadStrings(LANGS);
await mkdir(OUT_DIR, { recursive: true });
await mkdir(SVG_DIR, { recursive: true });

for (const lang of LANGS) {
  const t = tables[lang];
  const markup = svg({ title: t.hero.title, sub: t.og.imageSub });
  const svgPath = join(SVG_DIR, `${lang}.svg`);
  const pngPath = join(OUT_DIR, `snitt-${lang}.png`);
  await writeFile(svgPath, markup);
  await run('rsvg-convert', ['-w', String(W), '-h', String(H), svgPath, '-o', pngPath]);
  console.log('og:', pngPath);
}
