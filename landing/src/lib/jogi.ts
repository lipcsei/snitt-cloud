/**
 * A KÖZÖS JOGI SZOLGÁLTATÁS (lipcsei/jogi) kliense.
 *
 * Az ÁSZF és az adatvédelmi tájékoztató szövegét nem ez a repó gondozza: a jogi szolgáltatásban
 * élnek, verziózva. A jogi oldalak (/aszf, /adatvedelem) onnan töltik a HATÁLYOS szöveget,
 * hitelesítés nélkül. (Elfogadást a landing nem kér: itt nincs regisztrációhoz kötött szerződés.)
 *
 * KÉT SZABÁLY, amire a hívók építenek:
 *
 * 1. Ha a `VITE_JOGI_URL` nincs megadva, minden ki van kapcsolva (`jogiEnabled === false`): az
 *    oldalak a csomagba égetett szöveget mutatják, mint a bekötés előtt.
 * 2. A szöveg lekérése hibát dob, ha a szolgáltatás nem elérhető - a hívó ilyenkor a beégetett
 *    szövegre esik vissza. Egy jogi oldal nem lehet üres attól, hogy egy másik szolgáltatás áll.
 */

/** Az app rövid neve a jogi szolgáltatásban (az útvonalak része). */
export const JOGI_APP = 'snitt';

/** Ennyit várunk a jogi szolgáltatásra, mielőtt a tartalékra esünk vissza. */
const TIMEOUT_MS = 5000;

/**
 * A beállított cím eredetre + útvonalra normalizálva, záró `/` nélkül; üres sztring, ha nincs
 * megadva vagy nem http(s) cím. Tiszta függvény, hogy tesztelhető legyen.
 */
export function normalizeJogiUrl(raw: string | undefined): string {
  const value = raw?.trim();
  if (!value) return '';
  try {
    const url = new URL(value);
    if (url.protocol !== 'https:' && url.protocol !== 'http:') return '';
    return `${url.origin}${url.pathname}`.replace(/\/+$/, '');
  } catch {
    return '';
  }
}

export const JOGI_URL = normalizeJogiUrl(import.meta.env.VITE_JOGI_URL);

/** Be van-e kötve a jogi szolgáltatás ebben a buildben. */
export const jogiEnabled = JOGI_URL !== '';

export type LegalKind = 'aszf' | 'adatvedelem';

export interface LegalVersionInfo {
  version: number;
  effectiveAt: string;
  changeSummary: string;
}

export interface LegalDocument extends LegalVersionInfo {
  title: string;
  /**
   * A szöveg kész HTML-ként. MEGBÍZHATÓ FORRÁS: a saját szolgáltatásunk állítja elő Markdownból,
   * és a forrásban lévő nyers HTML-t nem engedi át - ezért (és csak ezért) illeszthető be
   * közvetlenül az oldalba.
   */
  html: string;
  /** A már közzétett, de még nem hatályos következő verzió, ha van. */
  upcoming?: LegalVersionInfo;
}

type Fetch = typeof fetch;

function publicBase(base: string): string {
  return `${base}/api/v1/public/apps/${JOGI_APP}`;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function versionInfo(value: unknown): LegalVersionInfo | null {
  if (!isRecord(value)) return null;
  const { version, effectiveAt, changeSummary } = value;
  if (typeof version !== 'number' || typeof effectiveAt !== 'string') return null;
  return { version, effectiveAt, changeSummary: typeof changeSummary === 'string' ? changeSummary : '' };
}

/**
 * Egy dokumentum hatályos szövege. HIBÁT DOB, ha a szolgáltatás nem elérhető, nem válaszol időben,
 * vagy a válasz nem a várt alakú - a hívó ilyenkor a beégetett szöveget mutatja.
 */
export async function fetchLegalDocument(kind: LegalKind, base = JOGI_URL, fetchImpl: Fetch = fetch): Promise<LegalDocument> {
  const res = await fetchImpl(`${publicBase(base)}/documents/${kind}`, { signal: AbortSignal.timeout(TIMEOUT_MS) });
  if (!res.ok) throw new Error(`jogi: ${res.status}`);
  const body: unknown = await res.json();
  const info = versionInfo(body);
  if (!isRecord(body) || !info || typeof body.title !== 'string' || typeof body.html !== 'string' || !body.html.trim()) {
    throw new Error('jogi: váratlan válasz');
  }
  return { ...info, title: body.title, html: body.html, upcoming: versionInfo(body.upcoming) ?? undefined };
}

const dateFormat = new Intl.DateTimeFormat('hu-HU', { year: 'numeric', month: 'long', day: 'numeric', timeZone: 'Europe/Budapest' });

/** A hatálybalépés napja magyarul, magyar idő szerint: „2026. szeptember 22.” */
export function formatEffectiveDate(iso: string): string {
  const date = new Date(iso);
  return Number.isNaN(date.getTime()) ? '' : dateFormat.format(date);
}
