import { useEffect, useState } from 'react';
import type { ReactNode } from 'react';
import { Link } from 'react-router-dom';
import { fetchLegalDocument, formatEffectiveDate, jogiEnabled } from '../lib/jogi';
import type { LegalDocument, LegalKind } from '../lib/jogi';
import '../styles/legal.css';

interface Props {
  kind: LegalKind;
  /** A cím fölötti rövid felirat (ugyanaz, mint a beégetett oldalon). */
  eyebrow: string;
  /** A másik jogi oldal, a lap alján. */
  other: { to: string; label: string };
  /**
   * A csomagba égetett oldal (pages/Terms.tsx, pages/PrivacyPolicy.tsx). Akkor látszik, ha a jogi
   * szolgáltatás nincs bekötve, vagy épp nem érhető el.
   */
  fallback: ReactNode;
}

type State = { status: 'loading' } | { status: 'failed' } | { status: 'ready'; doc: LegalDocument };

/**
 * Egy jogi oldal, amelynek a szövege a közös jogi szolgáltatásból jön (lib/jogi.ts): ott
 * szerkesztik és verziózzák, ez az oldal mindig a HATÁLYOS verziót mutatja. A keret (a nyelvi
 * megjegyzés, a cím, a lábléc) ugyanaz, mint a beégetett oldalaké.
 */
export default function RemoteLegal({ kind, eyebrow, other, fallback }: Props) {
  const [state, setState] = useState<State>({ status: 'loading' });

  useEffect(() => {
    if (!jogiEnabled) return undefined;
    let cancelled = false;
    fetchLegalDocument(kind).then(
      (doc) => {
        if (!cancelled) setState({ status: 'ready', doc });
      },
      () => {
        if (!cancelled) setState({ status: 'failed' });
      },
    );
    return () => {
      cancelled = true;
    };
  }, [kind]);

  if (!jogiEnabled || state.status === 'failed') return <>{fallback}</>;
  if (state.status === 'loading') {
    return (
      <main className="page legal-page" aria-busy="true">
        <div className="container container-narrow legal-content">
          <p className="muted">A dokumentum betöltése…</p>
        </div>
      </main>
    );
  }

  const { doc } = state;
  return (
    <main className="page legal-page">
      <div className="container container-narrow legal-content">
        <div className="legal-notice">
          <p>
            This document is currently available in Hungarian only, as the service operates under
            Hungarian law. An English version will follow.
          </p>
          <p>
            Dieses Dokument ist derzeit nur auf Ungarisch verfügbar, da der Dienst dem ungarischen
            Recht unterliegt. Eine deutsche Version folgt.
          </p>
        </div>

        <span className="eyebrow">{eyebrow}</span>
        <h1>{doc.title} — Snitt</h1>
        <p className="muted legal-effective">
          Hatályos: {formatEffectiveDate(doc.effectiveAt)} napjától · {doc.version}. verzió
        </p>

        {doc.upcoming && (
          <div className="legal-notice" role="note">
            <p>
              <strong>Változás lesz:</strong> {formatEffectiveDate(doc.upcoming.effectiveAt)} napjától
              új verzió lép hatályba.{doc.upcoming.changeSummary && ` ${doc.upcoming.changeSummary}`}
            </p>
          </div>
        )}

        {/* A HTML a SAJÁT jogi szolgáltatásunkból jön, amely Markdownból állítja elő, és a
            forrásban lévő nyers HTML-t nem engedi át (l. lib/jogi.ts) - ezért illeszthető be. */}
        <div className="legal-remote" dangerouslySetInnerHTML={{ __html: doc.html }} />

        <div className="legal-links">
          <Link to={other.to} className="btn btn-outline">
            {other.label}
          </Link>
          <Link to="/" className="btn btn-outline">
            Vissza a főoldalra
          </Link>
        </div>
      </div>
    </main>
  );
}
