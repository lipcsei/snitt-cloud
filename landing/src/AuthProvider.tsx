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
import { useLocation } from 'react-router-dom';
import {
  AUTH_INIT_TIMEOUT_MS,
  accountConsoleUrl,
  authConfigured,
  initOptions,
  keycloak,
  readProfile,
  type Profile,
} from './auth';
import { PATHS, absoluteUrl, langFromPathname } from './i18n';

export type AuthState = {
  /** false, amíg a check-sso fut; utána akkor is true, ha a Keycloak nem élt. */
  ready: boolean;
  authenticated: boolean;
  profile: Profile | null;
  login: () => void;
  register: () => void;
  logout: () => void;
  accountUrl: string;
};

const AuthContext = createContext<AuthState | null>(null);

/**
 * Ha az init elhasalt, a keycloak-js adapter nincs felállítva, és a login/logout
 * hívás dob. A marketing oldal ettől nem eshet szét, ezért mind el van kapva.
 */
function safely(fn: () => Promise<unknown> | void) {
  if (!authConfigured) return;
  try {
    Promise.resolve(fn()).catch((err) => console.warn('Keycloak nem érhető el', err));
  } catch (err) {
    console.warn('Keycloak nem érhető el', err);
  }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [ready, setReady] = useState(false);
  const [authenticated, setAuthenticated] = useState(false);
  const [profile, setProfile] = useState<Profile | null>(null);
  // A React 18 StrictMode kétszer futtatja az effektet, a keycloak.init() viszont
  // csak egyszer hívható meg.
  const started = useRef(false);

  useEffect(() => {
    if (started.current) return;
    started.current = true;

    // Keycloak nélküli build (a fiókok még nincsenek élesben): meg sem
    // próbáljuk - az adapter üres realmmel a saját domainünkre lőne.
    if (!authConfigured) {
      setReady(true);
      return;
    }

    const finish = (ok: boolean) => {
      setAuthenticated(ok);
      setProfile(ok ? readProfile() : null);
      setReady(true);
    };

    const timeout = new Promise<boolean>((resolve) => {
      window.setTimeout(() => resolve(false), AUTH_INIT_TIMEOUT_MS);
    });

    Promise.race([keycloak.init(initOptions), timeout])
      .then(finish)
      .catch(() => finish(false));

    keycloak.onAuthSuccess = () => finish(true);
    keycloak.onAuthRefreshSuccess = () => setProfile(readProfile());
    keycloak.onAuthLogout = () => finish(false);
    keycloak.onTokenExpired = () => {
      void keycloak.updateToken(30).catch(() => finish(false));
    };
  }, []);

  // A Keycloak felülete is azon a nyelven jöjjön, amelyiken a látogató épp olvas.
  const lang = langFromPathname(useLocation().pathname);
  const home = absoluteUrl(PATHS[lang].home);
  // Bejelentkezés után arra az oldalra jövünk vissza, ahonnan indult (pl. /profil).
  const here = window.location.href;

  const login = useCallback(() => {
    safely(() => keycloak.login({ redirectUri: here, locale: lang }));
  }, [here, lang]);

  const register = useCallback(() => {
    safely(() => keycloak.register({ redirectUri: here, locale: lang }));
  }, [here, lang]);

  const logout = useCallback(() => {
    safely(() => keycloak.logout({ redirectUri: home }));
  }, [home]);

  const value = useMemo<AuthState>(
    () => ({
      ready,
      authenticated,
      profile,
      login,
      register,
      logout,
      accountUrl: accountConsoleUrl(),
    }),
    [ready, authenticated, profile, login, register, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthState {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth() csak AuthProvider-en belül használható');
  return ctx;
}
