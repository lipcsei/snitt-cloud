import Keycloak from 'keycloak-js';
import type { KeycloakInitOptions } from 'keycloak-js';

export const keycloakConfig = {
  url: import.meta.env.VITE_KEYCLOAK_URL ?? 'http://localhost:8081',
  realm: import.meta.env.VITE_KEYCLOAK_REALM ?? 'snitt',
  clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID ?? 'snitt-admin',
};

export const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8090').replace(
  /\/$/,
  '',
);

/** Ez a szerep nyitja az /api/v1/admin/* végpontokat; a szerver is ezt nézi. */
export const ADMIN_ROLE = 'admin';

export const keycloak = new Keycloak(keycloakConfig);

export const initOptions: KeycloakInitOptions = {
  // A landinggel ellentétben itt nincs névtelen mód: az admin felületnek
  // minden oldala bejelentkezést igényel, ezért azonnal a Keycloakra megyünk.
  onLoad: 'login-required',
  pkceMethod: 'S256',
  silentCheckSsoRedirectUri: `${window.location.origin}/silent-check-sso.html`,
  checkLoginIframe: false,
  enableLogging: false,
};

export type AdminProfile = {
  name: string;
  email: string;
  username: string;
};

type TokenClaims = {
  name?: string;
  given_name?: string;
  family_name?: string;
  preferred_username?: string;
  email?: string;
  realm_access?: { roles?: string[] };
};

function claims(): TokenClaims | undefined {
  return (keycloak.tokenParsed ?? keycloak.idTokenParsed) as TokenClaims | undefined;
}

export function readProfile(): AdminProfile | null {
  const c = claims();
  if (!c) return null;

  const composed = [c.given_name, c.family_name].filter(Boolean).join(' ').trim();
  return {
    name: c.name || composed || c.preferred_username || c.email || 'Felhasználó',
    email: c.email ?? '',
    username: c.preferred_username ?? '',
  };
}

export function hasAdminRole(): boolean {
  return claims()?.realm_access?.roles?.includes(ADMIN_ROLE) ?? false;
}

export function accountConsoleUrl(): string {
  return `${keycloakConfig.url.replace(/\/$/, '')}/realms/${encodeURIComponent(
    keycloakConfig.realm,
  )}/account/`;
}
