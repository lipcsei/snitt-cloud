import { createContext, useContext, useMemo, type ReactNode } from 'react';
import { useLocation } from 'react-router-dom';
import { hu } from './hu';
import { en } from './en';
import type { Lang, RouteKey, Strings } from './types';

export type { Lang, RouteKey, Strings } from './types';

const TABLES: Record<Lang, Strings> = { hu, en };

export const LANGS = ['hu', 'en'] as const;

/**
 * A nyelvenkénti útvonalak. A nyelv KIZÁRÓLAG az URL-ből derül ki
 * (langFromPathname), külön állapotban nem tároljuk - így nem tud eltérni a
 * kettő. A böngészőben eltárolt preferencia csak az átirányítást befolyásolja.
 */
export const PATHS: Record<Lang, Record<RouteKey, string>> = {
  hu: { home: '/', install: '/telepites', profile: '/profil' },
  en: { home: '/en', install: '/en/install', profile: '/en/profile' },
};

const STORAGE_KEY = 'snitt.lang';

function normalize(pathname: string): string {
  const trimmed = pathname.replace(/\/+$/, '');
  return trimmed === '' ? '/' : trimmed;
}

export function langFromPathname(pathname: string): Lang {
  const path = normalize(pathname);
  return path === '/en' || path.startsWith('/en/') ? 'en' : 'hu';
}

/** Melyik oldalon vagyunk - a nyelvváltó ehhez keresi meg a másik nyelvű párját. */
export function routeKeyFromPathname(pathname: string): RouteKey {
  const lang = langFromPathname(pathname);
  const path = normalize(pathname);
  const entries = Object.entries(PATHS[lang]) as [RouteKey, string][];
  return entries.find(([, value]) => normalize(value) === path)?.[0] ?? 'home';
}

/** A Vite base-t is figyelembe vevő abszolút URL (hreflang, kijelentkezés). */
export function absoluteUrl(path: string): string {
  const base = import.meta.env.BASE_URL.replace(/\/$/, '');
  const origin = typeof window === 'undefined' ? '' : window.location.origin;
  return `${origin}${base}${path === '/' ? '/' : path}`;
}

export function rememberLang(lang: Lang): void {
  try {
    window.localStorage.setItem(STORAGE_KEY, lang);
  } catch {
    // Privát mód vagy letiltott tárolás: a nyelvválasztás ettől még működik,
    // csak nem marad meg a következő látogatásra.
  }
}

function storedLang(): Lang | null {
  try {
    const value = window.localStorage.getItem(STORAGE_KEY);
    return value === 'hu' || value === 'en' ? value : null;
  } catch {
    return null;
  }
}

/**
 * Csak a nyelvsemleges "/" belépési pontra vonatkozik. Aki egyszer kézzel
 * választott nyelvet, azt onnantól a döntése viszi - így nem lehet beragadni
 * egy nyelvbe, amit a böngésző beállítása erőltet rá.
 */
export function prefersEnglishOnEntry(): boolean {
  const stored = storedLang();
  if (stored) return stored === 'en';
  if (typeof navigator === 'undefined') return false;

  // A kettő közül az dönt, amelyik előrébb van a böngésző listájában: egy
  // en-GB > hu-HU beállítású látogató angolul olvasna, egy hu-HU > en-US pedig
  // magyarul. Ha egyik sem szerepel benne, az angol a jobb tipp.
  const list = navigator.languages?.length ? navigator.languages : [navigator.language];
  for (const item of list) {
    const code = item?.toLowerCase() ?? '';
    if (code.startsWith('hu')) return false;
    if (code.startsWith('en')) return true;
  }
  return true;
}

/** "Letöltés — {os}" + { os: 'Windows' } → "Letöltés — Windows" */
export function fill(template: string, vars: Record<string, string>): string {
  return template.replace(/\{(\w+)\}/g, (match, key: string) => vars[key] ?? match);
}

type I18n = {
  lang: Lang;
  t: Strings;
  /** Az aktuális nyelv útvonala az adott oldalhoz. */
  path: (key: RouteKey) => string;
};

const I18nContext = createContext<I18n | null>(null);

export function LanguageProvider({ children }: { children: ReactNode }) {
  const { pathname } = useLocation();
  const lang = langFromPathname(pathname);

  const value = useMemo<I18n>(
    () => ({ lang, t: TABLES[lang], path: (key: RouteKey) => PATHS[lang][key] }),
    [lang],
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n(): I18n {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error('useI18n() csak LanguageProvider-en belül használható');
  return ctx;
}

/** A leggyakoribb eset: csak a szövegtábla kell. */
export function useT(): Strings {
  return useI18n().t;
}

export function useLang(): Lang {
  return useI18n().lang;
}
