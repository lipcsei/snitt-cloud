import type { ReactNode } from 'react';
import { ApiError } from '../api';

export function Loading({ label = 'Betöltés…' }: { label?: string }) {
  return (
    <div className="state" role="status" aria-live="polite">
      <span className="spinner" aria-hidden="true" />
      <span>{label}</span>
    </div>
  );
}

export function Empty({ title, hint }: { title: string; hint?: string }) {
  return (
    <div className="state state--empty">
      <strong>{title}</strong>
      {hint ? <span className="muted">{hint}</span> : null}
    </div>
  );
}

/** A szerver hibaüzenetét mutatja, nem egy általános "hiba történt" szöveget. */
export function ErrorBox({
  error,
  onRetry,
  children,
}: {
  error: unknown;
  onRetry?: () => void;
  children?: ReactNode;
}) {
  const status = error instanceof ApiError ? error.status : 0;
  const message =
    error instanceof Error ? error.message : 'Ismeretlen hiba történt a kérés közben.';

  return (
    <div className="state state--error" role="alert">
      <div className="state__head">
        <strong>Nem sikerült a művelet</strong>
        {status > 0 ? <span className="pill pill--danger">HTTP {status}</span> : null}
      </div>
      <span>{message}</span>
      {children}
      {onRetry ? (
        <button type="button" className="btn btn--ghost" onClick={onRetry}>
          Újrapróbálom
        </button>
      ) : null}
    </div>
  );
}

/** Egy művelet után megjelenő, elhalványuló visszajelzés sáv. */
export function Banner({
  kind,
  message,
  onClose,
}: {
  kind: 'ok' | 'warn' | 'error';
  message: string;
  onClose: () => void;
}) {
  return (
    <div className={`banner banner--${kind}`} role="status">
      <span>{message}</span>
      <button type="button" className="banner__close" onClick={onClose} aria-label="Bezárás">
        ×
      </button>
    </div>
  );
}
