import { useState } from 'react';

export type Done = (message: string) => void;

/**
 * Egy admin művelet lefuttatása dialógusból: zárolt gombok futás közben,
 * siker esetén bezárás + visszajelzés, hiba esetén a dialógus NYITVA marad a
 * szerver üzenetével, hogy az admin lássa, mi nem sikerült, és a beírt adatok
 * se vesszenek el.
 */
export function useAction(onDone: Done, onClose: () => void) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<unknown>(null);

  const run = async (fn: () => Promise<unknown>, message: string) => {
    setBusy(true);
    setError(null);
    try {
      await fn();
      onClose();
      onDone(message);
    } catch (err) {
      setError(err);
    } finally {
      setBusy(false);
    }
  };

  return { busy, error, run };
}
