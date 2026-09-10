import { useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { api } from '../api';
import { PlanBadge, SubscriptionBadge } from '../components/Badges';
import { Page } from '../components/Layout';
import { Pagination } from '../components/Pagination';
import { Empty, ErrorBox, Loading } from '../components/States';
import { formatDate } from '../format';
import { useAsync, useDebounced } from '../hooks';

const PAGE_SIZE = 25;

export function Users() {
  const [params, setParams] = useSearchParams();
  const page = Number(params.get('page') ?? '1') || 1;
  const [search, setSearch] = useState(params.get('q') ?? '');
  const debounced = useDebounced(search);

  const setPage = (next: number) => {
    const p = new URLSearchParams(params);
    p.set('page', String(next));
    setParams(p);
  };

  const onSearch = (value: string) => {
    setSearch(value);
    const p = new URLSearchParams(params);
    if (value) p.set('q', value);
    else p.delete('q');
    // Szűréskor mindig az első oldalra ugrunk, különben üres lapon köthetne ki.
    p.set('page', '1');
    setParams(p, { replace: true });
  };

  const { data, loading, error, reload } = useAsync(
    () => api.users({ q: debounced || undefined, page, page_size: PAGE_SIZE }),
    [debounced, page],
  );

  return (
    <Page
      title="Felhasználók"
      description="Az identitás a Keycloakból jön, a profil és az előfizetés a saját adatbázisból."
    >
      <div className="toolbar">
        <label className="field field--search">
          <span className="sr-only">Keresés</span>
          <input
            type="search"
            value={search}
            placeholder="Keresés e-mail, felhasználónév vagy név szerint…"
            onChange={(e) => onSearch(e.target.value)}
          />
        </label>
        <button type="button" className="btn btn--ghost" onClick={reload}>
          Frissítés
        </button>
      </div>

      {error ? <ErrorBox error={error} onRetry={reload} /> : null}
      {loading && !data ? <Loading label="Felhasználók betöltése…" /> : null}

      {data && !error ? (
        data.users.length === 0 ? (
          <Empty
            title="Nincs találat"
            hint={debounced ? `A(z) „${debounced}” keresésre nincs felhasználó.` : undefined}
          />
        ) : (
          <div className="panel">
            <table className="table table--hover">
              <thead>
                <tr>
                  <th>Felhasználó</th>
                  <th>E-mail</th>
                  <th>Állapot</th>
                  <th>Előfizetés</th>
                  <th>Regisztrált</th>
                </tr>
              </thead>
              <tbody>
                {data.users.map((u) => (
                  <tr key={u.subject}>
                    <td>
                      <Link className="link" to={`/felhasznalok/${encodeURIComponent(u.subject)}`}>
                        {u.full_name || u.username || u.subject}
                      </Link>
                      <span className="muted mono"> · {u.username}</span>
                    </td>
                    <td>
                      {u.email || <span className="muted">–</span>}
                      {u.email && !u.email_verified ? (
                        <span className="pill pill--warn">nem megerősített</span>
                      ) : null}
                    </td>
                    <td>
                      {u.enabled ? (
                        <span className="pill pill--ok">Engedélyezve</span>
                      ) : (
                        <span className="pill pill--danger">Letiltva</span>
                      )}
                      {!u.has_profile ? (
                        <span className="pill pill--muted">nincs felhő profil</span>
                      ) : null}
                    </td>
                    <td>
                      {u.subscription ? (
                        <span className="badges">
                          <PlanBadge plan={u.subscription.plan} />
                          <SubscriptionBadge
                            status={u.subscription.status}
                            cancelAtPeriodEnd={u.subscription.cancel_at_period_end}
                          />
                        </span>
                      ) : (
                        <span className="muted">Ingyenes</span>
                      )}
                    </td>
                    <td className="muted">{formatDate(u.registered_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            <Pagination
              page={data.page}
              pageSize={data.page_size}
              total={data.total}
              onPage={setPage}
            />
          </div>
        )
      ) : null}
    </Page>
  );
}
