import { Link } from 'react-router-dom';
import { api } from '../api';
import { Page } from '../components/Layout';
import { ErrorBox, Loading } from '../components/States';
import { formatMoney } from '../format';
import { useAsync } from '../hooks';

function Stat({
  label,
  value,
  hint,
  tone,
}: {
  label: string;
  value: string;
  hint?: string;
  tone?: 'accent' | 'warn';
}) {
  return (
    <div className={`stat ${tone ? `stat--${tone}` : ''}`}>
      <span className="stat__label">{label}</span>
      <strong className="stat__value">{value}</strong>
      {hint ? <span className="stat__hint">{hint}</span> : null}
    </div>
  );
}

export function Dashboard() {
  const { data, loading, error, reload } = useAsync(() => api.overview(), []);

  if (loading) return <Loading label="Összesítők betöltése…" />;
  if (error) return <ErrorBox error={error} onRetry={reload} />;
  if (!data) return null;

  const planName = (key: string) => data.plans.find((p) => p.key === key)?.name ?? key;

  return (
    <Page title="Vezérlőpult" description="A felhő oldal aktuális állapota egy képernyőn.">
      {data.warnings.map((w) => (
        <div key={w} className="banner banner--warn" role="status">
          <span>{w}</span>
        </div>
      ))}

      <section className="stats">
        <Stat
          label="Regisztrált felhasználók"
          value={data.user_count === null ? '–' : String(data.user_count)}
          hint="Keycloak realm"
        />
        <Stat
          label="Felhő profilok"
          value={String(data.profile_count)}
          hint={`+${data.new_profiles_30d} az elmúlt 30 napban`}
        />
        <Stat
          label="Aktív előfizetések"
          value={String(data.active_subscription_count)}
          hint={`+${data.new_subscriptions_30d} az elmúlt 30 napban`}
          tone="accent"
        />
        <Stat
          label="MRR"
          value={formatMoney(data.mrr_minor, data.currency)}
          hint="Aktív előfizetések havi díja"
          tone="accent"
        />
        <Stat
          label="Nyitott számlák"
          value={String(data.unpaid_invoice_count)}
          hint={formatMoney(data.unpaid_invoice_total_minor, data.currency)}
          tone={data.unpaid_invoice_count > 0 ? 'warn' : undefined}
        />
      </section>

      <section className="panel">
        <div className="panel__head">
          <h2>Aktív előfizetések csomagonként</h2>
          <Link className="btn btn--ghost" to="/elofizetesek">
            Előfizetések
          </Link>
        </div>

        {data.active_by_plan.length === 0 ? (
          <p className="muted panel__body">Jelenleg nincs aktív előfizetés.</p>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Csomag</th>
                <th className="num">Darab</th>
                <th className="num">Havi díj</th>
                <th className="num">Havi bevétel</th>
              </tr>
            </thead>
            <tbody>
              {data.active_by_plan.map((row) => {
                const plan = data.plans.find((p) => p.key === row.plan);
                const price = plan?.price_minor ?? 0;
                return (
                  <tr key={row.plan}>
                    <td>
                      <strong>{planName(row.plan)}</strong>
                      <span className="muted"> · {row.plan}</span>
                    </td>
                    <td className="num">{row.count}</td>
                    <td className="num">{formatMoney(price, data.currency)}</td>
                    <td className="num">{formatMoney(price * row.count, data.currency)}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
        <p className="panel__note">
          A listaár szerinti bevétel tájékoztató: az MRR a ténylegesen rögzített, akár egyedi
          árakból számol, ezért eltérhet ettől a sortól.
        </p>
      </section>

      <section className="panel">
        <div className="panel__head">
          <h2>Csomagok</h2>
        </div>
        <table className="table">
          <thead>
            <tr>
              <th>Csomag</th>
              <th>Leírás</th>
              <th>Feloldott jogosultságok</th>
              <th className="num">Listaár</th>
            </tr>
          </thead>
          <tbody>
            {data.plans.map((p) => (
              <tr key={p.key}>
                <td>
                  <strong>{p.name}</strong>
                  <span className="muted"> · {p.key}</span>
                </td>
                <td className="muted">{p.description}</td>
                <td>
                  {p.feature_keys.map((f) => (
                    <code key={f}>{f}</code>
                  ))}
                </td>
                <td className="num">{formatMoney(p.price_minor, data.currency)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </Page>
  );
}
