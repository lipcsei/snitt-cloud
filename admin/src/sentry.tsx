import type { ReactNode } from 'react';
import * as Sentry from '@sentry/react';

/**
 * Opcionális hibajelentés a saját GlitchTipünkre (Sentry-kompatibilis). A DSN
 * build-időben épül be (VITE_SENTRY_DSN), és az érték lehet ÜRES sztring is: ez
 * a kikapcsolt állapot, ilyenkor az SDK el sem indul, és az app pontosan úgy
 * működik, mintha ez a fájl nem is lenne. A DSN build-időben konstans, ezért
 * üres értéknél az SDK a csomagból is kimarad; ezért nincs itt trim() sem: az
 * elfedné a konstanst a csomagoló elől.
 */
const dsn = import.meta.env.VITE_SENTRY_DSN;

/**
 * Az URL a lekérdezés és a fragment nélkül. Az admin szűrői az URL-ben élnek
 * (pl. a felhasználók listáján `?q=<e-mail>`), a Keycloak visszairányítása
 * pedig a fragmentben hoz kódot: ezek nem valók a hibajelentésbe.
 */
function stripUrl(url: string): string {
  return url.split(/[?#]/)[0];
}

export function initSentry(): void {
  if (!dsn) return;
  Sentry.init({
    dsn,
    environment: import.meta.env.VITE_SENTRY_ENVIRONMENT || import.meta.env.MODE,
    // Személyes adat nélkül; nyomkövetés (tracing) és session replay sincs: nincs
    // hozzáadva a browserTracingIntegration és a replayIntegration.
    sendDefaultPii: false,
    beforeSend(event) {
      if (event.request) {
        if (event.request.url) event.request.url = stripUrl(event.request.url);
        delete event.request.query_string;
      }
      return event;
    },
    beforeBreadcrumb(crumb) {
      // A navigációs és a hálózati morzsák URL-jei ugyanígy kaphatnak lekérdezést.
      const data = crumb.data;
      if (data) {
        for (const key of ['url', 'from', 'to']) {
          if (typeof data[key] === 'string') data[key] = stripUrl(data[key]);
        }
      }
      return crumb;
    },
  });
}

/**
 * A render közben dobott hibákat jelenti (React 18: ErrorBoundary). Kikapcsolt
 * hibajelentésnél átlátszó: semmi nem kerül a komponensfába, a viselkedés
 * változatlan. Hiba esetén a fallback nélküli boundary üres oldalt ad, ami
 * ugyanaz, mint amit a React boundary nélkül is csinálna (leszedi a fát).
 */
export function SentryBoundary({ children }: { children: ReactNode }) {
  if (!dsn) return <>{children}</>;
  return <Sentry.ErrorBoundary>{children}</Sentry.ErrorBoundary>;
}
