import { useCallback, useEffect, useRef, useState } from 'react';
import { api } from './api';
import type { Plan } from './types';

export type Async<T> = {
  data: T | null;
  loading: boolean;
  error: unknown;
  reload: () => void;
};

/**
 * Egyszerű lekérdezés-hook betöltés/hiba állapottal. A `deps` változásakor
 * újratölt, és eldobja az időközben megérkező, elavult válaszokat.
 */
export function useAsync<T>(fn: () => Promise<T>, deps: unknown[]): Async<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<unknown>(null);
  const [nonce, setNonce] = useState(0);
  const latest = useRef(0);

  // A callback minden rendernél új referencia lenne, ezért a deps dönt.
  const run = useRef(fn);
  run.current = fn;

  useEffect(() => {
    const ticket = ++latest.current;
    setLoading(true);
    setError(null);

    run
      .current()
      .then((result) => {
        if (ticket !== latest.current) return;
        setData(result);
        setLoading(false);
      })
      .catch((err) => {
        if (ticket !== latest.current) return;
        setError(err);
        setLoading(false);
      });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, nonce]);

  const reload = useCallback(() => setNonce((n) => n + 1), []);
  return { data, loading, error, reload };
}

/** Késleltetett érték, hogy a keresőmező ne indítson kérést minden leütésre. */
export function useDebounced<T>(value: T, delay = 300): T {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const t = window.setTimeout(() => setDebounced(value), delay);
    return () => window.clearTimeout(t);
  }, [value, delay]);
  return debounced;
}

// A csomaglista ritkán változik, és több képernyő is kéri: egyszer töltjük le.
let plansPromise: Promise<{ plans: Plan[]; currency: string }> | null = null;

export function usePlans(): Async<{ plans: Plan[]; currency: string }> {
  return useAsync(() => {
    if (!plansPromise) {
      plansPromise = api.plans().catch((err) => {
        plansPromise = null;
        throw err;
      });
    }
    return plansPromise;
  }, []);
}
