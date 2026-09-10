import { useCallback, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { api } from '../api';
import { PlanBadge, SubscriptionBadge } from '../components/Badges';
import { Page } from '../components/Layout';
import { Pagination } from '../components/Pagination';
import { SubscriptionActions } from '../components/SubscriptionActions';
import { Banner, Empty, ErrorBox, Loading } from '../components/States';
import { formatDate, formatMoney } from '../format';
import { useAsync, usePlans } from '../hooks';
import { SUBSCRIPTION_STATUS_LABELS } from '../format';
import type { SubscriptionStatus } from '../types';

const PAGE_SIZE = 25;
const STATUSES: SubscriptionStatus[] = ['active', 'past_due', 'canceled', 'expired'];

export function Subscriptions() {
  const [params, setParams] = useSearchParams();
  const page = Number(params.get('page') ?? '1') || 1;
  const plan = params.get('plan') ?? '';
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

  const plans = usePlans();
  const { data, loading, error, reload } = useAsync(
    () =>
      api.subscriptions({
        plan: plan || undefined,
        status: status || undefined,
        subject: subject || undefined,
        page,
        page_size: PAGE_SIZE,
      }),
    [plan, status, subject, page],
  );

  const done = useCallback(
    (message: string) => {
      setNotice(message);
      reload();
    },
    [reload],
  );

  return (
    <Page
      title="Előfizetések"
      description="A fizetős csomagok nyilvántartása. A kiadott jogosultságok ezekből következnek."
      actions={
        <button type="button" className="btn btn--ghost" onClick={reload}>
          Frissítés
        </button>
      }
    >
      {notice ? <Banner kind="ok" message={notice} onClose={() => setNotice(null)} /> : null}

      <div className="toolbar">
        <label className="field">
          <span>Csomag</span>
          <select value={plan} onChange={(e) => setParam('plan', e.target.value)}>
            <option value="">Mind</option>
            {(plans.data?.plans ?? []).map((p) => (
              <option key={p.key} value={p.key}>
                {p.name}
              </option>
            ))}
          </select>
        </label>
        <label className="field">
          <span>Állapot</span>
          <select value={status} onChange={(e) => setParam('status', e.target.value)}>
            <option value="">Mind</option>
            {STATUSES.map((s) => (
              <option key={s} value={s}>
                {SUBSCRIPTION_STATUS_LABELS[s]}
              </option>
            ))}
          </select>
        </label>
        {subject ? (
          <button type="button" className="btn btn--ghost" onClick={() => setParam('subject', '')}>
            Felhasználó-szűrő törlése
          </button>
        ) : null}
      </div>

      {error ? <ErrorBox error={error} onRetry={reload} /> : null}
      {loading && !data ? <Loading label="Előfizetések betöltése…" /> : null}

      {data && !error ? (
        data.subscriptions.length === 0 ? (
          <Empty
            title="Nincs találat"
            hint="Módosítsd a szűrőket, vagy adj ki előfizetést egy felhasználó adatlapján."
          />
        ) : (
          <div className="panel">
            <table className="table">
              <thead>
                <tr>
                  <th>Felhasználó</th>
                  <th>Csomag</th>
                  <th>Állapot</th>
                  <th>Időszak vége</th>
                  <th className="num">Ár</th>
                  <th className="right">Műveletek</th>
                </tr>
              </thead>
              <tbody>
                {data.subscriptions.map((s) => (
                  <tr key={s.id}>
                    <td>
                      <Link className="link" to={`/felhasznalok/${encodeURIComponent(s.subject)}`}>
                        {s.user.display_name || s.user.email || s.subject}
                      </Link>
                      {s.user.email ? <span className="muted"> · {s.user.email}</span> : null}
                    </td>
                    <td>
                      <PlanBadge plan={s.plan} />
                    </td>
                    <td>
                      <SubscriptionBadge
                        status={s.status}
                        cancelAtPeriodEnd={s.cancel_at_period_end}
                      />
                    </td>
                    <td className="muted">{formatDate(s.current_period_end)}</td>
                    <td className="num">{formatMoney(s.price_minor, s.currency)}</td>
                    <td className="right">
                      {plans.data ? (
                        <SubscriptionActions
                          subscription={s}
                          plans={plans.data.plans}
                          currency={plans.data.currency}
                          onDone={done}
                          compact
                        />
                      ) : null}
                    </td>
                  </tr>
                ))}
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
