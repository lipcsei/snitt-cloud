/**
 * Google Analytics 4 (gtag.js) — CSAK a felhasználó előzetes hozzájárulásával (GDPR + ePrivacy).
 *
 *  - A címke NINCS az index.html-ben: a gtag.js-t az `enableAnalytics` tölti be, és csak akkor,
 *    ha a felhasználó elfogadta a mérést (a cookie-sávon). Előtte és elutasítás után semmilyen
 *    kérés nem megy a Google felé.
 *  - Google Consent Mode v2: az alapértelmezés minden tárolásra 'denied', a hozzájárulás
 *    megadásakor csak az `analytics_storage` kap 'granted'-et. A hirdetési jelzések (ad_storage,
 *    ad_user_data, ad_personalization) mindig tiltva maradnak.
 *  - Visszavonáskor: `window['ga-disable-<azonosító>'] = true` (a gtag.js új eseményt nem küld) és
 *    `analytics_storage` → 'denied'.
 *  - Nincs beépített (hardkódolt) mérési azonosító: amíg a `VITE_GA_MEASUREMENT_ID` build-időben
 *    üres, hiányzik vagy nem `G-XXXX` alakú, a modul no-op — nem tölt be semmit, nem dob hibát.
 *
 * A gombamester/frontend `lib/analytics.ts` mintáját követi, egyszerűsítve (nincs saját
 * cookie-domain / cookie-prefix kezelés, mert a snitt.video landing a valódi domaint szolgálja
 * ki, nem egy testvér-aldomaint).
 */

/** A felhasználó döntése. Ha nincs (`null`), még nem döntött: ilyenkor nem mérünk. */
export type AnalyticsConsent = 'granted' | 'denied';

/** A döntés localStorage kulcsa. */
export const CONSENT_STORAGE_KEY = 'snitt.analyticsConsent';

/** A betöltött gtag.js script-tag azonosítója: ebből tudjuk, hogy már bent van. */
export const GTAG_SCRIPT_ID = 'snitt-gtag';

export const GTAG_BASE_URL = 'https://www.googletagmanager.com/gtag/js';

/* ---------------------------------------------------------------- hozzájárulás */

/** A tárolt érték értelmezése: csak a pontos 'granted' / 'denied' számít, minden más = nincs döntés. */
export function parseConsent(raw: string | null | undefined): AnalyticsConsent | null {
  return raw === 'granted' || raw === 'denied' ? raw : null;
}

/** A döntés eszközönként, a localStorage-ban. Tiltott / hibás tárolónál: nincs döntés. */
export function readConsent(): AnalyticsConsent | null {
  try {
    return parseConsent(window.localStorage.getItem(CONSENT_STORAGE_KEY));
  } catch {
    return null;
  }
}

/** A tároló hibáját elnyeli: a döntés ilyenkor erre a munkamenetre érvényes. */
export function writeConsent(consent: AnalyticsConsent): void {
  try {
    window.localStorage.setItem(CONSENT_STORAGE_KEY, consent);
  } catch {
    // Privát mód vagy letiltott tárolás: a hozzájárulás ettől még működik,
    // csak nem marad meg a következő látogatásra.
  }
}

/** Egy `storage` esemény érinti-e a döntést (`key === null`: a lap teljes tárolóját törölték). */
export const isConsentStorageKey = (key: string | null): boolean =>
  key === null || key === CONSENT_STORAGE_KEY;

/* ---------------------------------------------------------------- hol aktív */

export interface AnalyticsEnv {
  VITE_GA_MEASUREMENT_ID?: string;
}

export interface AnalyticsConfig {
  measurementId: string;
}

const MEASUREMENT_ID_PATTERN = /^G-[A-Z0-9]{4,16}$/;

/** Üres, hiányzó vagy nem `G-XXXX` alakú azonosítónál: `null` (nincs beépített alapértelmezés). */
export function resolveMeasurementId(raw: string | undefined): string | null {
  const id = (raw ?? '').trim();
  return MEASUREMENT_ID_PATTERN.test(id) ? id : null;
}

/** `null` = ebben a buildben nincs mérés: ilyenkor sáv sincs, és semmi nem töltődik be. */
export function resolveAnalyticsConfig(env: AnalyticsEnv): AnalyticsConfig | null {
  const measurementId = resolveMeasurementId(env.VITE_GA_MEASUREMENT_ID);
  return measurementId ? { measurementId } : null;
}

/* ---------------------------------------------------------------- gtag betöltése */

export type Gtag = (...args: unknown[]) => void;

/** A window szükséges szelete (a tesztben kamu objektum). */
export interface AnalyticsWindow {
  dataLayer?: unknown[];
  gtag?: Gtag;
  /** a GA hivatalos kikapcsolója: ha `true`, a gtag.js ennek az azonosítónak új eseményt nem küld */
  [disableFlag: `ga-disable-${string}`]: boolean | undefined;
}

export interface AnalyticsScriptElement {
  id: string;
  async: boolean;
  src: string;
}

/** A document szükséges szelete (a tesztben kamu objektum). */
export interface AnalyticsDocument {
  getElementById(id: string): unknown;
  createElement(tagName: 'script'): AnalyticsScriptElement;
  head: { appendChild(node: AnalyticsScriptElement): unknown };
}

/** Hozzájárulás előtti alapállapot: minden tiltva (Consent Mode v2). */
export const CONSENT_DEFAULT = {
  ad_storage: 'denied',
  ad_user_data: 'denied',
  ad_personalization: 'denied',
  analytics_storage: 'denied',
} as const;

export const CONSENT_GRANTED = { analytics_storage: 'granted' } as const;
export const CONSENT_REVOKED = { analytics_storage: 'denied' } as const;

export const disableFlag = (measurementId: string): `ga-disable-${string}` => `ga-disable-${measurementId}`;

export const gtagScriptSrc = (measurementId: string): string =>
  `${GTAG_BASE_URL}?id=${encodeURIComponent(measurementId)}`;

/**
 * A szabványos gtag-sor: `function gtag(){dataLayer.push(arguments);}`. A gtag.js CSAK az
 * `arguments` objektumot ismeri fel parancsként (tömböt nem), ezért nem nyílfüggvény és nem rest
 * paraméter.
 */
function ensureGtag(win: AnalyticsWindow): Gtag {
  win.dataLayer = win.dataLayer ?? [];
  if (!win.gtag) {
    win.gtag = function gtag() {
      // eslint-disable-next-line prefer-rest-params
      (win.dataLayer ??= []).push(arguments);
    };
  }
  return win.gtag;
}

/** A gtag.js már bent van a lapon (ebben a betöltésben egyszer már engedélyezték). */
export function isGtagLoaded(doc: AnalyticsDocument): boolean {
  return doc.getElementById(GTAG_SCRIPT_ID) != null;
}

const browserWindow = (): AnalyticsWindow => window as unknown as AnalyticsWindow;
const browserDocument = (): AnalyticsDocument => document as unknown as AnalyticsDocument;

export type EnableResult = 'loaded' | 'resumed' | 'already-active';

/**
 * A mérés bekapcsolása — CSAK a felhasználó elfogadása után hívható.
 *  - először: dataLayer + gtag, consent default (minden tiltva), consent update (analytics_storage
 *    granted), `js`, `config` (→ első page_view), végül a gtag.js aszinkron script-tagje —
 *    egyetlenegyszer, ebben a sorrendben, tehát a Consent Mode v2 defaultja mindig a script
 *    betöltése ELŐTT áll be;
 *  - visszavonás után ugyanazon a lapon: a kikapcsoló le, `analytics_storage` újra granted (a
 *    következő oldalváltástól mér, újratöltés nem kell);
 *  - ha már mér: nem csinál semmit (React StrictMode kétszer futtatja az effektet).
 */
export function enableAnalytics(
  measurementId: string,
  win: AnalyticsWindow = browserWindow(),
  doc: AnalyticsDocument = browserDocument(),
): EnableResult {
  const flag = disableFlag(measurementId);
  if (isGtagLoaded(doc)) {
    if (win[flag] !== true) return 'already-active';
    win[flag] = false;
    ensureGtag(win)('consent', 'update', CONSENT_GRANTED);
    return 'resumed';
  }

  win[flag] = false;
  const gtag = ensureGtag(win);
  gtag('consent', 'default', CONSENT_DEFAULT);
  gtag('consent', 'update', CONSENT_GRANTED);
  gtag('js', new Date());
  gtag('config', measurementId);

  const script = doc.createElement('script');
  script.id = GTAG_SCRIPT_ID;
  script.async = true;
  script.src = gtagScriptSrc(measurementId);
  doc.head.appendChild(script);
  return 'loaded';
}

/**
 * A mérés leállítása (elutasítás, visszavonás, vagy nincs döntés): a GA kikapcsolója fel, és ha a
 * lap eddig mért, `analytics_storage` → denied.
 */
export function disableAnalytics(
  measurementId: string,
  win: AnalyticsWindow = browserWindow(),
  doc: AnalyticsDocument = browserDocument(),
): void {
  const wasLoaded = isGtagLoaded(doc);
  win[disableFlag(measurementId)] = true;
  if (wasLoaded && win.gtag) win.gtag('consent', 'update', CONSENT_REVOKED);
}

export type ApplyResult = EnableResult | 'stopped' | 'off';

/**
 * A döntés érvényesítése (az `AnalyticsConsentProvider` effektje): ha a buildben nincs mérés
 * (`config` null, azaz nincs `VITE_GA_MEASUREMENT_ID`), semmi; CSAK a 'granted' tölt be (vagy
 * folytat); minden más — elutasítás, még nincs döntés, érvénytelen tárolt érték — leállít, és a
 * Google felé semmit nem tölt be.
 */
export function applyConsent(
  config: AnalyticsConfig | null,
  consent: AnalyticsConsent | null,
  win: AnalyticsWindow = browserWindow(),
  doc: AnalyticsDocument = browserDocument(),
): ApplyResult {
  if (!config) return 'off';
  if (consent === 'granted') return enableAnalytics(config.measurementId, win, doc);
  disableAnalytics(config.measurementId, win, doc);
  return 'stopped';
}
