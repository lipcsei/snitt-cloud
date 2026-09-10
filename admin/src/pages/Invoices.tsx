import { useCallback, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { api } from '../api';
import { InvoiceBadge } from '../components/Badges';
import { InvoiceActions } from '../components/InvoiceActions';
import { Page } from '../components/Layout';
import { Pagination } from '../components/Pagination';
import { Banner, Empty, ErrorBox, Loading } from '../components/States';
import { INVOICE_STATUS_LABELS, formatDate, formatMoney } from '../format';
import { useAsync } from '../hooks';
import type { InvoiceStatus } from '../types';

const PAGE_SIZE = 25;
const STATUSES: InvoiceStatus[] = ['open', 'paid', 'void'];

export function Invoices() {
  const [params, setParams] = useSearchParams();
  const page = Number(params.get('page') ?? '1') || 1;
  const status = params.get('status') ?? '';
  const subject = params.get('subject') ?? '';
  const [notice, setNotice] = useState<string | null>(null);

  const setParam = (key: string, value: string) => {
    const p = new URLSearchParams(params);
    if (value) p.set(key, value);
    else p.delete(key);
    if (key !== 'page') p.set('page', '1');
    setParams(p);
  };

  const { data, loading, error, reload } = useAsync(
    () =>
      api.invoices({
        status: status || undefined,
        subject: subject || undefined,
        page,
        page_size: PAGE_SIZE,
      }),
    [status, subject, page],
  );

  const done = useCallback(
    (message: string) => {
      setNotice(message);
      reload();
    },
    [reload],
  );

  const currency = data?.currency ?? 'HUF';

  return (
    <Page
      title="Számlázás"
      description="A saját könyvelési nyilvántartás. Fizetési szolgáltató nincs bekötve: a beérkezést kézzel jelöljük."
      actions={
        <button type="button" className="btn btn--ghost" onClick={reload}>
          Frissítés
        </button>
      }
    >
      {notice ? <Banner kind="ok" message={notice} onClose={() => setNotice(null)} /> : null}

      <div className="toolbar">
        <label className="field">
          <span>Állapot</span>
          <select value={status} onChange={(e) => setParam('status', e.target.value)}>
            <option value="">Mind</option>
            {STATUSES.map((s) => (
              <option key={s} value={s}>
                {INVOICE_STATUS_LABELS[s]}
              </option>
            ))}
          </select>
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
      </div>

      {data ? (
        <section className="stats stats--inline">
          <div className="stat">
            <span className="stat__label">Szűrt tételek</span>
            <strong className="stat__value">{data.totals.count}</strong>
          </div>
          <div className="stat stat--warn">
            <span className="stat__label">Nyitott</span>
            <strong className="stat__value">
              {formatMoney(data.totals.open_minor, currency)}
            </strong>
          </div>
          <div className="stat stat--accent">
            <span className="stat__label">Kifizetve</span>
            <strong className="stat__value">
              {formatMoney(data.totals.paid_minor, currency)}
            </strong>
          </div>
          <div className="stat">
            <span className="stat__label">Összesen</span>
            <strong className="stat__value">
              {formatMoney(data.totals.total_minor, currency)}
            </strong>
          </div>
        </section>
      ) : null}

      {error ? <ErrorBox error={error} onRetry={reload} /> : null}
      {loading && !data ? <Loading label="Számlák betöltése…" /> : null}

      {data && !error ? (
        data.invoices.length === 0 ? (
          <Empty
            title="Nincs számla"
            hint="A szűrőkre nem illeszkedik tétel. Számlát egy felhasználó adatlapjáról lehet kiállítani."
          />
        ) : (
          <div className="panel">
            <table className="table">
              <thead>
                <tr>
                  <th>Számlaszám</th>
                  <th>Felhasználó</th>
                  <th>Állapot</th>
                  <th>Kiállítva</th>
                  <th>Határidő</th>
                  <th>Fizetve</th>
                  <th className="num">Összeg</th>
                  <th className="right">Műveletek</th>
                </tr>
              </thead>
              <tbody>
                {data.invoices.map((inv) => {
                  const overdue =
                    inv.status === 'open' &&
                    inv.due_at !== null &&
                    new Date(inv.due_at) < new Date();
                  return (
                    <tr key={inv.id}>
                      <td className="mono">{inv.number}</td>
                      <td>
                        <Link
                          className="link"
                          to={`/felhasznalok/${encodeURIComponent(inv.subject)}`}
                        >
                          {inv.user.display_name || inv.user.email || inv.subject}
                        </Link>
                      </td>
                      <td>
                        <span className="badges">
                          <InvoiceBadge status={inv.status} />
                          {overdue ? <span className="pill pill--danger">lejárt</span> : null}
                        </span>
                      </td>
                      <td className="muted">{formatDate(inv.issued_at)}</td>
                      <td className="muted">{formatDate(inv.due_at)}</td>
                      <td className="muted">{formatDate(inv.paid_at)}</td>
                      <td className="num">{formatMoney(inv.amount_minor, inv.currency)}</td>
                      <td className="right">
                        <InvoiceActions invoice={inv} onDone={done} />
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
            <Pagination
              page={data.page}
              pageSize={data.page_size}
              total={data.totals.count}
              onPage={(next) => setParam('page', String(next))}
            />
          </div>
        )
      ) : null}
    </Page>
  );
}
