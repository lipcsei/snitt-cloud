import { useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { api } from '../api';
import { Page } from '../components/Layout';
import { Pagination } from '../components/Pagination';
import { Empty, ErrorBox, Loading } from '../components/States';
import { auditActionLabel, formatDateTime } from '../format';
import { useAsync } from '../hooks';
import type { AuditEntry } from '../types';

const PAGE_SIZE = 50;

/** Az érintett rekord oldala, ha a felület tud rá hivatkozni. */
function targetLink(entry: AuditEntry) {
  if (entry.target_type === 'subscription') {
    return { to: `/elofizetesek?subject=${encodeURIComponent(entry.subject)}`, label: 'előfizetés' };
  }
  if (entry.target_type === 'invoice') {
    return { to: `/szamlazas?subject=${encodeURIComponent(entry.subject)}`, label: 'számla' };
  }
  return null;
}

/** A művelet nyers paraméterei, csak kérésre kinyitva. */
function Detail({ detail }: { detail: Record<string, unknown> }) {
  const [open, setOpen] = useState(false);
  const keys = Object.keys(detail ?? {});
  if (keys.length === 0) return null;

  return (
    <div className="audit__detail">
      <button type="button" className="link link--quiet" onClick={() => setOpen((v) => !v)}>
        {open ? 'Részletek elrejtése' : `Részletek (${keys.length})`}
      </button>
      {open ? <pre className="audit__json">{JSON.stringify(detail, null, 2)}</pre> : null}
    </div>
  );
}

export function Audit() {
  const [params, setParams] = useSearchParams();
  const page = Number(params.get('page') ?? '1') || 1;
  const action = params.get('action') ?? '';
  const subject = params.get('subject') ?? '';
  const actor = params.get('actor') ?? '';
  const from = params.get('from') ?? '';
  const to = params.get('to') ?? '';

  const setParam = (key: string, value: string) => {
    const p = new URLSearchParams(params);
    if (value) p.set(key, value);
    else p.delete(key);
    if (key !== 'page') p.set('page', '1');
    setParams(p);
  };

  const { data, loading, error, reload } = useAsync(
    () =>
      api.audit({
        action: action || undefined,
        subject: subject || undefined,
        actor: actor || undefined,
        from: from || undefined,
        to: to || undefined,
        page,
        page_size: PAGE_SIZE,
      }),
    [action, subject, actor, from, to, page],
  );

  // A műveletlistát a szerver adja; amíg nincs válasz, a szűrő üres.
  const actions = data?.actions ?? [];
  const filtered = Boolean(action || subject || actor || from || to);

  return (
    <Page
      title="Napló"
      description="Ki, mikor, mit módosított. A napló csak nő: bejegyzést szerkeszteni és törölni nem lehet."
      actions={
        <button type="button" className="btn btn--ghost" onClick={reload}>
          Frissítés
        </button>
      }
    >
      <div className="toolbar">
        <label className="field">
          <span>Művelet</span>
          <select value={action} onChange={(e) => setParam('action', e.target.value)}>
            <option value="">Mind</option>
            {actions.map((a) => (
              <option key={a} value={a}>
                {auditActionLabel(a)}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          <span>Ettől</span>
          <input type="date" value={from} onChange={(e) => setParam('from', e.target.value)} />
        </label>
        <label className="field">
          <span>Eddig</span>
          <input type="date" value={to} onChange={(e) => setParam('to', e.target.value)} />
        </label>
        {subject ? (
          <div className="filter-chip">
            <span className="muted">Egy felhasználóra szűrve</span>
            <Link className="link mono" to={`/felhasznalok/${encodeURIComponent(subject)}`}>
              {subject}
            </Link>
            <button type="button" className="btn btn--ghost" onClick={() => setParam('subject', '')}>
              Törlés
            </button>
          </div>
        ) : null}
        {actor ? (
          <div className="filter-chip">
            <span className="muted">Egy adminra szűrve</span>
            <span className="mono">{actor}</span>
            <button type="button" className="btn btn--ghost" onClick={() => setParam('actor', '')}>
              Törlés
            </button>
          </div>
        ) : null}
      </div>

      {error ? <ErrorBox error={error} onRetry={reload} /> : null}
      {loading && !data ? <Loading label="Napló betöltése…" /> : null}

      {data && !error ? (
        data.entries.length === 0 ? (
          <Empty
            title={filtered ? 'Nincs találat' : 'A napló üres'}
            hint={
              filtered
                ? 'A szűrőkre nem illeszkedik bejegyzés.'
                : 'Bejegyzés akkor keletkezik, amikor valaki előfizetést vagy számlát módosít.'
            }
          />
        ) : (
          <div className="panel">
            <table className="table">
              <thead>
                <tr>
                  <th>Mikor</th>
                  <th>Ki</th>
                  <th>Mit</th>
                  <th>Kivel</th>
                </tr>
              </thead>
              <tbody>
                {data.entries.map((e) => {
                  const target = targetLink(e);
                  return (
                    <tr key={e.id}>
                      <td className="muted nowrap">{formatDateTime(e.at)}</td>
                      <td>
                        <button
                          type="button"
                          className="link link--quiet"
                          title="Szűrés erre az adminra"
                          onClick={() => setParam('actor', e.actor_subject)}
                        >
                          {e.actor_label || e.actor_subject}
                        </button>
                      </td>
                      <td>
                        <span className="pill">{auditActionLabel(e.action)}</span>
                        <div className="audit__summary">{e.summary}</div>
                        <Detail detail={e.detail} />
                      </td>
                      <td>
                        {e.subject ? (
                          <Link
                            className="link"
                            to={`/felhasznalok/${encodeURIComponent(e.subject)}`}
                          >
                            {e.subject}
                          </Link>
                        ) : (
                          <span className="muted">–</span>
                        )}
                        {target ? (
                          <div>
                            <Link className="link link--quiet" to={target.to}>
                              {target.label}
                            </Link>
                          </div>
                        ) : null}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            <Pagination
              page={data.page}
              pageSize={data.page_size}
              total={data.total}
              onPage={(next) => setParam('page', String(next))}
            />
          </div>
        )
      ) : null}
    </Page>
  );
}
