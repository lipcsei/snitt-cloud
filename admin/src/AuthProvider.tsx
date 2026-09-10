import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import {
  accountConsoleUrl,
  hasAdminRole,
  initOptions,
  keycloak,
  readProfile,
  type AdminProfile,
} from './auth';

export type AuthState = {
  /** false, amíg a Keycloak init fut. */
  ready: boolean;
  authenticated: boolean;
  isAdmin: boolean;
  /** Ha a Keycloak egyáltalán nem érhető el, ez mondja meg, miért. */
  error: string | null;
  profile: AdminProfile | null;
  logout: () => void;
  accountUrl: string;
};

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(false);
  const [authenticated, setAuthenticated] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);
  const [profile, setProfile] = useState<AdminProfile | null>(null);
  const [error, setError] = useState<string | null>(null);
  // A React 18 StrictMode kétszer futtatja az effektet, a keycloak.init()
  // viszont csak egyszer hívható meg.
  const started = useRef(false);

  useEffect(() => {
    if (started.current) return;
    started.current = true;

    const sync = (ok: boolean) => {
      setAuthenticated(ok);
      setIsAdmin(ok && hasAdminRole());
      setProfile(ok ? readProfile() : null);
      setReady(true);
    };

    keycloak
      .init(initOptions)
      .then(sync)
      .catch((err) => {
        // login-required mellett ide csak akkor jutunk, ha a Keycloak nem
        // válaszol - ilyenkor az egész felület használhatatlan, mondjuk ki.
        setError(err instanceof Error ? err.message : 'A Keycloak nem érhető el.');
        sync(false);
      });

    keycloak.onAuthSuccess = () => sync(true);
    keycloak.onAuthRefreshSuccess = () => sync(true);
    keycloak.onAuthLogout = () => sync(false);
    keycloak.onTokenExpired = () => {
      void keycloak.updateToken(30).catch(() => sync(false));
    };
  }, []);

  const logout = useCallback(() => {
    void keycloak.logout({ redirectUri: `${window.location.origin}/` }).catch(() => undefined);
  }, []);

  const value = useMemo<AuthState>(
    () => ({
      ready,
      authenticated,
      isAdmin,
      error,
      profile,
      logout,
      accountUrl: accountConsoleUrl(),
    }),
    [ready, authenticated, isAdmin, error, profile, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth() csak AuthProvider-en belül használható');
  return ctx;
}
