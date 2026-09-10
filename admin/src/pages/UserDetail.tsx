import { useCallback, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { api } from '../api';
import { InvoiceBadge, PlanBadge, SubscriptionBadge } from '../components/Badges';
import { CreateInvoiceDialog, InvoiceActions } from '../components/InvoiceActions';
import { Page } from '../components/Layout';
import { GrantDialog, SubscriptionActions } from '../components/SubscriptionActions';
import { Banner, Empty, ErrorBox, Loading } from '../components/States';
import { formatDate, formatDateTime, formatMoney } from '../format';
import { useAsync, usePlans } from '../hooks';

export function UserDetail() {
  const { subject = '' } = useParams();
  const [notice, setNotice] = useState<string | null>(null);
  const [dialog, setDialog] = useState<'grant' | 'invoice' | null>(null);

  const { data, loading, error, reload } = useAsync(() => api.user(subject), [subject]);
  const plans = usePlans();

  const done = useCallback(
    (message: string) => {
      setNotice(message);
      reload();
    },
    [reload],
  );

  if (loading && !data) return <Loading label="Felhasználó betöltése…" />;
  if (error) return <ErrorBox error={error} onRetry={reload} />;
  if (!data) return null;

  const { user } = data;
  const label = user.full_name || user.email || user.username || user.subject;
  const live = data.subscriptions.find((s) => s.status === 'active' || s.status === 'past_due');
  const currency = plans.data?.currency ?? live?.currency ?? 'HUF';

  return (
    <Page
      title={label}
      description={user.email || user.subject}
      actions={
        <>
          <Link className="btn btn--ghost" to="/felhasznalok">
            Vissza a listához
          </Link>
          {!live && plans.data ? (
            <button type="button" className="btn btn--primary" onClick={() => setDialog('grant')}>
              Előfizetés kiadása
            </button>
          ) : null}
          <button type="button" className="btn btn--ghost" onClick={() => setDialog('invoice')}>
            Számla kiállítása
          </button>
        </>
      }
    >
      {notice ? <Banner kind="ok" message={notice} onClose={() => setNotice(null)} /> : null}
      {plans.error ? <ErrorBox error={plans.error} onRetry={plans.reload} /> : null}

      <section className="grid grid--2">
        <div className="panel">
          <div className="panel__head">
            <h2>Profil</h2>
          </div>
          <dl className="kv">
            <dt>Subject (Keycloak ID)</dt>
            <dd className="mono">{user.subject}</dd>
            <dt>Felhasználónév</dt>
            <dd>{user.username || '–'}</dd>
            <dt>E-mail</dt>
            <dd>
              {user.email || '–'}{' '}
              {user.email ? (
                user.email_verified ? (
                  <span className="pill pill--ok">megerősítve</span>
                ) : (
                  <span className="pill pill--warn">nem megerősített</span>
                )
              ) : null}
            </dd>
            <dt>Fiók állapota</dt>
            <dd>
              {user.enabled ? (
                <span className="pill pill--ok">Engedélyezve</span>
              ) : (
                <span className="pill pill--danger">Letiltva</span>
              )}
            </dd>
            <dt>Regisztrált</dt>
            <dd>{formatDateTime(user.registered_at)}</dd>
            <dt>Megjelenítendő név</dt>
            <dd>{user.has_profile ? user.display_name || '–' : <span className="muted">nincs felhő profil</span>}</dd>
            <dt>Nyelv</dt>
            <dd>{user.locale || '–'}</dd>
            <dt>Első felhő-belépés</dt>
            <dd>{formatDateTime(user.profile_since)}</dd>
          </dl>
        </div>

        <div className="panel">
          <div className="panel__head">
            <h2>Jogosultságok</h2>
          </div>
          {data.entitlements.length === 0 ? (
            <p className="muted panel__body">
              Nincs kiadott jogosultság – a desktop app az ingyenes, hirdetéses módban működik.
            </p>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>Kulcs</th>
                  <th>Kiadva</th>
                  <th>Lejár</th>
                </tr>
              </thead>
              <tbody>
                {data.entitlements.map((e) => {
                  const expired = e.expires_at !== null && new Date(e.expires_at) <= new Date();
                  return (
                    <tr key={e.feature_key} className={expired ? 'is-dim' : ''}>
                      <td>
                        <code>{e.feature_key}</code>
                      </td>
                      <td className="muted">{formatDate(e.granted_at)}</td>
                      <td>
                        {e.expires_at ? formatDate(e.expires_at) : <span className="muted">soha</span>}
                        {expired ? <span className="pill pill--muted">lejárt</span> : null}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      </section>

      <section className="panel">
        <div className="panel__head">
          <h2>Előfizetések</h2>
        </div>
        {data.subscriptions.length === 0 ? (
          <Empty title="Nincs előfizetés" hint="Ez a felhasználó az ingyenes csomagot használja." />
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Csomag</th>
                <th>Állapot</th>
                <th>Időszak</th>
                <th className="num">Ár</th>
                <th className="right">Műveletek</th>
              </tr>
            </thead>
            <tbody>
              {data.subscriptions.map((s) => (
                <tr key={s.id}>
                  <td>
                    <PlanBadge plan={s.plan} />
                  </td>
                  <td>
                    <SubscriptionBadge status={s.status} cancelAtPeriodEnd={s.cancel_at_period_end} />
                  </td>
                  <td className="muted">
                    {formatDate(s.current_period_start)} – {formatDate(s.current_period_end)}
                  </td>
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
        )}
      </section>

      <section className="panel">
        <div className="panel__head">
          <h2>Számlák</h2>
          <span className="muted">
            {data.invoice_totals.count} db · nyitott:{' '}
            {formatMoney(data.invoice_totals.open_minor, currency)} · kifizetve:{' '}
            {formatMoney(data.invoice_totals.paid_minor, currency)}
          </span>
        </div>
        {data.invoices.length === 0 ? (
          <Empty title="Nincs számla" hint="Ehhez a felhasználóhoz még nem állítottunk ki számlát." />
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Számlaszám</th>
                <th>Állapot</th>
                <th>Kiállítva</th>
                <th>Fizetve</th>
                <th className="num">Összeg</th>
                <th className="right">Műveletek</th>
              </tr>
            </thead>
            <tbody>
              {data.invoices.map((inv) => (
                <tr key={inv.id}>
                  <td className="mono">{inv.number}</td>
                  <td>
                    <InvoiceBadge status={inv.status} />
                  </td>
                  <td className="muted">{formatDate(inv.issued_at)}</td>
                  <td className="muted">{formatDate(inv.paid_at)}</td>
                  <td className="num">{formatMoney(inv.amount_minor, inv.currency)}</td>
                  <td className="right">
                    <InvoiceActions invoice={inv} onDone={done} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      {dialog === 'grant' && plans.data ? (
        <GrantDialog
          subject={user.subject}
          subjectLabel={label}
          plans={plans.data.plans}
          currency={plans.data.currency}
          onClose={() => setDialog(null)}
          onDone={done}
        />
      ) : null}
      {dialog === 'invoice' ? (
        <CreateInvoiceDialog
          subject={user.subject}
          subjectLabel={label}
          subscriptionId={live?.id ?? null}
          currency={currency}
          onClose={() => setDialog(null)}
          onDone={done}
        />
      ) : null}
    </Page>
  );
}
