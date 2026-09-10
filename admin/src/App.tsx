import { Navigate, Route, Routes } from 'react-router-dom';
import { useAuth } from './AuthProvider';
import { Layout } from './components/Layout';
import { Loading } from './components/States';
import { Audit } from './pages/Audit';
import { Dashboard } from './pages/Dashboard';
import { Invoices } from './pages/Invoices';
import { Subscriptions } from './pages/Subscriptions';
import { UserDetail } from './pages/UserDetail';
import { Users } from './pages/Users';

/** Bejelentkezett, de admin szerep nélküli felhasználó zárt képernyője. */
function NoAccess() {
  const { profile, logout, accountUrl } = useAuth();
  return (
    <div className="gate">
      <div className="gate__card">
        <span className="gate__mark" aria-hidden="true" />
        <h1>Nincs jogosultságod</h1>
        <p>
          A(z) <strong>{profile?.email || profile?.name || 'bejelentkezett fiók'}</strong> nem
          rendelkezik <code>admin</code> szereppel, ezért a Snitt admin felület nem érhető el.
        </p>
        <p className="muted">
          Ha ez tévedés, kérd meg a realm adminisztrátort, hogy adja hozzá a fiókhoz az{' '}
          <code>admin</code> realm szerepet a Keycloakban.
        </p>
        <div className="gate__actions">
          <a className="btn btn--ghost" href={accountUrl} target="_blank" rel="noreferrer">
            Fiókom
          </a>
          <button type="button" className="btn btn--primary" onClick={logout}>
            Kijelentkezés
          </button>
        </div>
      </div>
    </div>
  );
}

/** A Keycloak elérhetetlenségekor: itt nincs mit renderelni token nélkül. */
function AuthUnavailable({ message }: { message: string }) {
  return (
    <div className="gate">
      <div className="gate__card">
        <span className="gate__mark gate__mark--warn" aria-hidden="true" />
        <h1>A bejelentkezés nem érhető el</h1>
        <p>{message}</p>
        <p className="muted">
          Ellenőrizd, hogy fut-e a Keycloak, és hogy a <code>VITE_KEYCLOAK_URL</code> a megfelelő
          címre mutat-e.
        </p>
        <div className="gate__actions">
          <button type="button" className="btn btn--primary" onClick={() => window.location.reload()}>
            Újratöltés
          </button>
        </div>
      </div>
    </div>
  );
}

export default function App() {
  const { ready, authenticated, isAdmin, error } = useAuth();

  if (!ready) {
    return (
      <div className="gate">
        <Loading label="Bejelentkezés ellenőrzése…" />
      </div>
    );
  }
  if (error) return <AuthUnavailable message={error} />;
  // login-required mellett ide nem jutunk el bejelentkezés nélkül, de ha a
  // Keycloak mégis visszaenged, ne egy félig működő felületet mutassunk.
  if (!authenticated) return <AuthUnavailable message="A munkamenet nem jött létre." />;
  if (!isAdmin) return <NoAccess />;

  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<Dashboard />} />
        <Route path="felhasznalok" element={<Users />} />
        <Route path="felhasznalok/:subject" element={<UserDetail />} />
        <Route path="elofizetesek" element={<Subscriptions />} />
        <Route path="szamlazas" element={<Invoices />} />
        <Route path="naplo" element={<Audit />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}
