import Keycloak from 'keycloak-js';
import type { KeycloakInitOptions } from 'keycloak-js';

export const keycloakConfig = {
  url: import.meta.env.VITE_KEYCLOAK_URL ?? 'http://localhost:8081',
  realm: import.meta.env.VITE_KEYCLOAK_REALM ?? 'snitt',
  clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID ?? 'snitt-landing',
};

export const keycloak = new Keycloak(keycloakConfig);

export const initOptions: KeycloakInitOptions = {
  onLoad: 'check-sso',
  pkceMethod: 'S256',
  // Rejtett iframe-ben nézzük meg a meglévő session-t, hogy a marketing oldal
  // soha ne irányítódjon át a Keycloakra betöltéskor. Ha a Keycloak nem él,
  // az iframe egyszerűen elhasal, az oldal viszont marad.
  silentCheckSsoRedirectUri: `${window.location.origin}/silent-check-sso.html`,
  silentCheckSsoFallback: false,
  checkLoginIframe: false,
  enableLogging: false,
};

/** Ha a Keycloak nem válaszol, ennyi idő után feladjuk és névtelenként megyünk tovább. */
export const AUTH_INIT_TIMEOUT_MS = 4000;

export type Profile = {
  name: string;
  email: string;
};

type IdClaims = {
  name?: string;
  given_name?: string;
  family_name?: string;
  preferred_username?: string;
  email?: string;
};

export function readProfile(): Profile | null {
  const claims = (keycloak.idTokenParsed ?? keycloak.tokenParsed) as IdClaims | undefined;
  if (!claims) return null;

  const composed = [claims.given_name, claims.family_name].filter(Boolean).join(' ').trim();
  const fullName =
    claims.name || composed || claims.preferred_username || claims.email || 'Felhasználó';

  return { name: fullName, email: claims.email ?? '' };
}

export function accountConsoleUrl(): string {
  return `${keycloakConfig.url.replace(/\/$/, '')}/realms/${encodeURIComponent(
    keycloakConfig.realm,
  )}/account/`;
}
