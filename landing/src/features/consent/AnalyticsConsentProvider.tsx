import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import {
  applyConsent,
  isConsentStorageKey,
  readConsent,
  resolveAnalyticsConfig,
  writeConsent,
} from '../../lib/analytics';
import type { AnalyticsConfig, AnalyticsConsent } from '../../lib/analytics';

export interface AnalyticsConsentState {
  /** ebben a buildben van-e mérés (van érvényes `VITE_GA_MEASUREMENT_ID`); ha nincs, sáv sincs */
  available: boolean;
  /** a felhasználó döntése ezen az eszközön; `null`: még nem döntött (nem mérünk) */
  consent: AnalyticsConsent | null;
  setConsent: (next: AnalyticsConsent) => void;
}

const AnalyticsConsentContext = createContext<AnalyticsConsentState | null>(null);

/** A mérési hozzájárulás állapota. Csak az `<AnalyticsConsentProvider>` alatt használható. */
export function useAnalyticsConsent(): AnalyticsConsentState {
  const state = useContext(AnalyticsConsentContext);
  if (!state) throw new Error('useAnalyticsConsent() csak az <AnalyticsConsentProvider> alatt használható');
  return state;
}

const buildConfig = (): AnalyticsConfig | null => resolveAnalyticsConfig(import.meta.env);

interface ProviderProps {
  children: ReactNode;
  /** csak teszthez: alapból a build env-jéből jön (`resolveAnalyticsConfig`) */
  config?: AnalyticsConfig | null;
}

/**
 * A Google Analytics hozzájárulás és annak érvényesítése. A döntés: 'granted' → betöltés /
 * folytatás; 'denied' vagy nincs döntés → leállítás (`applyConsent`, a lib/analytics.ts
 * részletezi). Amíg nincs `VITE_GA_MEASUREMENT_ID`, `available` mindig false: se sáv, se mérés.
 */
export function AnalyticsConsentProvider({ children, config: given }: ProviderProps) {
  const [config] = useState(() => (given === undefined ? buildConfig() : given));
  const [consent, setConsentState] = useState(readConsent);

  useEffect(() => {
    applyConsent(config, consent);
  }, [config, consent]);

  // egy másik lapon döntöttek (vagy vonták vissza): ez a lap is kövesse
  useEffect(() => {
    const onStorage = (e: StorageEvent) => {
      if (isConsentStorageKey(e.key)) setConsentState(readConsent());
    };
    window.addEventListener('storage', onStorage);
    return () => window.removeEventListener('storage', onStorage);
  }, []);

  const setConsent = useCallback((next: AnalyticsConsent) => {
    // tiltott tárolónál is érvényes erre a munkamenetre
    writeConsent(next);
    setConsentState(next);
  }, []);

  const value = useMemo(() => ({ available: config !== null, consent, setConsent }), [config, consent, setConsent]);
  return <AnalyticsConsentContext.Provider value={value}>{children}</AnalyticsConsentContext.Provider>;
}
