import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from 'react';
import NotYetModal from './components/NotYetModal';

type Kind = 'download' | 'register';

const NotYetContext = createContext<((kind: Kind) => void) | null>(null);

/**
 * A letöltés és a regisztráció még nem elérhető (a telepítők nincsenek kiadva,
 * a felhő oldal csak helyben fut), ezért ezek a gombok egyelőre egy türelemre
 * intő üzenetet nyitnak. Ha majd élesedik, elég ezt a providert kivenni.
 */
export function NotYetProvider({ children }: { children: ReactNode }) {
  const [kind, setKind] = useState<Kind | null>(null);
  const show = useCallback((k: Kind) => setKind(k), []);
  const value = useMemo(() => show, [show]);

  return (
    <NotYetContext.Provider value={value}>
      {children}
      {kind && <NotYetModal kind={kind} onClose={() => setKind(null)} />}
    </NotYetContext.Provider>
  );
}

/** showNotYet(kind) - a gombokra kötve. */
export function useNotYet(): (kind: Kind) => void {
  const ctx = useContext(NotYetContext);
  if (!ctx) throw new Error('useNotYet csak NotYetProvider-en belül használható');
  return ctx;
}
