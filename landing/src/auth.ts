import Keycloak from 'keycloak-js';
import type { KeycloakInitOptions } from 'keycloak-js';

/**
 * A build-időben beégetett érték lehet ÜRES sztring is (nem csak hiányzó), ha
 * a CI-ban nincs kitöltve a repository variable - az `??` erre nem esik vissza.
 * Emiatt indult el korábban a keycloak-js üres realmmel, és kérte végtelen
 * ciklusban a `/realms//protocol/...` címet a saját domainünkről.
 */
function env(value: string | undefined, fallback: string): string {
  const trimmed = (value ?? '').trim();
  return trimmed || fallback;
}

export const keycloakConfig = {
  url: env(import.meta.env.VITE_KEYCLOAK_URL, ''),
  realm: env(import.meta.env.VITE_KEYCLOAK_REALM, ''),
  clientId: env(import.meta.env.VITE_KEYCLOAK_CLIENT_ID, ''),
};

/**
 * A fiókok még nincsenek élesben: a nyilvános oldal Keycloak nélkül épül, és
 * ilyenkor NEM szabad inicializálni sem - enélkül az adapter a saját
 * domainünkre lő üres realmmel, és 404-ekkel szemeteli tele a konzolt.
 */
export const authConfigured =
  keycloakConfig.url !== '' && keycloakConfig.realm !== '' && keycloakConfig.clientId !== '';

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
  if (!authConfigured) return '';
  return `${keycloakConfig.url.replace(/\/$/, '')}/realms/${encodeURIComponent(
    keycloakConfig.realm,
  )}/account/`;
}
