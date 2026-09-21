import { useEffect, useState } from 'react';
import { DOWNLOAD_BASE_URL } from './constants';

/**
 * A legújabb kiadás leírása: a szerveren lévő `latest.json` (lásd a snitt repo
 * scripts/publish-downloads.sh-ját). A fájlnevek ÁLLANDÓK (a verzió nincs bennük),
 * a letöltési cím ezért mindig `<alap>/latest/<név>`.
 */
export type ReleaseKey = 'windows' | 'macos' | 'linux-appimage' | 'linux-deb' | 'linux-rpm';

export type ReleaseFile = { name: string; file: string; size: number; sha256: string };

export type Release = {
  version: string;
  date: string;
  files: Partial<Record<ReleaseKey, ReleaseFile>>;
};

const KEYS: readonly ReleaseKey[] = ['windows', 'macos', 'linux-appimage', 'linux-deb', 'linux-rpm'];
const VERSION_RE = /^[0-9]+(\.[0-9]+){1,3}([-+][0-9A-Za-z.]+)?$/;
// A fájlnév a letöltési címbe kerül: csak sima fájlnév-karakterek, útvonal-elemek nélkül.
const NAME_RE = /^[A-Za-z0-9][A-Za-z0-9._-]*$/;

function parseFile(raw: unknown): ReleaseFile | null {
  if (!raw || typeof raw !== 'object') return null;
  const f = raw as Record<string, unknown>;
  if (typeof f.name !== 'string' || !NAME_RE.test(f.name)) return null;
  if (typeof f.file !== 'string' || !NAME_RE.test(f.file)) return null;
  if (typeof f.size !== 'number' || !Number.isFinite(f.size) || f.size < 0) return null;
  if (typeof f.sha256 !== 'string' || !/^[0-9a-f]{64}$/.test(f.sha256)) return null;
  return { name: f.name, file: f.file, size: f.size, sha256: f.sha256 };
}

/** A szerver válaszát érvényesíti; érvénytelen vagy üres kiadásra null. */
export function parseRelease(raw: unknown): Release | null {
  if (!raw || typeof raw !== 'object') return null;
  const r = raw as Record<string, unknown>;
  if (typeof r.version !== 'string' || !VERSION_RE.test(r.version)) return null;
  const date = typeof r.date === 'string' ? r.date : '';
  const files: Release['files'] = {};
  const rawFiles = r.files;
  if (rawFiles && typeof rawFiles === 'object') {
    for (const key of KEYS) {
      const f = parseFile((rawFiles as Record<string, unknown>)[key]);
      if (f) files[key] = f;
    }
  }
  return Object.keys(files).length > 0 ? { version: r.version, date, files } : null;
}

/** A fájl letöltési címe. */
export function fileUrl(f: ReleaseFile): string {
  return `${DOWNLOAD_BASE_URL}/latest/${f.name}`;
}

/** Az ellenőrzőösszegek fájlja (a latest/ mappa összes fájljára). */
export const CHECKSUMS_URL = `${DOWNLOAD_BASE_URL}/latest/SHA256SUMS`;

/** Ember-olvasható méret: 12,3 MB. */
export function fmtSize(bytes: number, lang: string): string {
  const mb = bytes / (1024 * 1024);
  const n = mb >= 100 ? Math.round(mb).toString() : mb.toFixed(1);
  return `${lang === 'hu' || lang === 'de' ? n.replace('.', ',') : n} MB`;
}

export type ReleaseState =
  | { status: 'loading'; release: null }
  | { status: 'ready'; release: Release }
  | { status: 'none'; release: null };

/**
 * Betölti a legújabb kiadás leírását. Ha a szerver nem elérhető, a fájl hiányzik
 * (még nem volt kiadás) vagy érvénytelen, az állapot 'none': a letöltés-gombok
 * ilyenkor a "hamarosan" üzenetet mutatják, nem egy halott hivatkozást.
 */
export function useLatestRelease(): ReleaseState {
  const [state, setState] = useState<ReleaseState>({ status: 'loading', release: null });

  useEffect(() => {
    const ctrl = new AbortController();
    fetch(`${DOWNLOAD_BASE_URL}/latest.json`, { signal: ctrl.signal, cache: 'no-cache' })
      .then((res) => (res.ok ? res.json() : null))
      .then((json: unknown) => {
        const release = parseRelease(json);
        setState(release ? { status: 'ready', release } : { status: 'none', release: null });
      })
      .catch((err: unknown) => {
        if ((err as { name?: string })?.name === 'AbortError') return;
        setState({ status: 'none', release: null });
      });
    return () => ctrl.abort();
  }, []);

  return state;
}
