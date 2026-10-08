/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_KEYCLOAK_URL?: string;
  readonly VITE_KEYCLOAK_REALM?: string;
  readonly VITE_KEYCLOAK_CLIENT_ID?: string;
  readonly VITE_DOWNLOAD_BASE_URL?: string;
  readonly VITE_SENTRY_DSN?: string;
  readonly VITE_SENTRY_ENVIRONMENT?: string;
  readonly VITE_GA_MEASUREMENT_ID?: string;
  /** A közös jogi szolgáltatás nyilvános címe; üresen a beégetett jogi szövegek látszanak (lib/jogi.ts). */
  readonly VITE_JOGI_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
