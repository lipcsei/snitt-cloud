import { Component, type ErrorInfo, type ReactNode } from 'react';
import type { Breadcrumb, ErrorEvent } from '@sentry/react';

/**
 * Opcionális hibajelentés a saját GlitchTipünkre (Sentry-kompatibilis). A DSN
 * build-időben épül be (VITE_SENTRY_DSN), és az érték lehet ÜRES sztring is: ez
 * a kikapcsolt állapot, ilyenkor az SDK el sem indul, és az oldal pontosan úgy
 * működik, mintha ez a fájl nem is lenne.
 *
 * A landing nyilvános oldal, ezért az SDK NEM része a fő csomagnak: csak
 * beállított DSN mellett, külön csomagként (dinamikus import) töltődik be, a
 * betöltés alatt keletkező hibákat viszont még nem látja. A DSN build-időben
 * konstans, ezért üres értéknél a különálló csomag el sem készül; ezért nincs
 * itt trim() sem: az elfedné a konstanst a csomagoló elől.
 */
const dsn = import.meta.env.VITE_SENTRY_DSN;

type Sdk = typeof import('./sentrySdk');
let sdk: Promise<Sdk | null> | undefined;

/**
 * Az URL a lekérdezés és a fragment nélkül: a Keycloak visszairányítása a
 * fragmentben hoz kódot, ez nem való a hibajelentésbe.
 */
function stripUrl(url: string): string {
  return url.split(/[?#]/)[0];
}

function beforeSend(event: ErrorEvent): ErrorEvent {
  if (event.request) {
    if (event.request.url) event.request.url = stripUrl(event.request.url);
    delete event.request.query_string;
  }
  return event;
}

function beforeBreadcrumb(crumb: Breadcrumb): Breadcrumb {
  // A navigációs és a hálózati morzsák URL-jei ugyanígy kaphatnak lekérdezést.
  const data = crumb.data;
  if (data) {
    for (const key of ['url', 'from', 'to']) {
      if (typeof data[key] === 'string') data[key] = stripUrl(data[key]);
    }
  }
  return crumb;
}

/** Betölti és elindítja az SDK-t (egyszer); ha a betöltés nem sikerül, null. */
function load(): Promise<Sdk | null> {
  if (!sdk) {
    sdk = import('./sentrySdk').then(
      (Sentry) => {
        Sentry.init({
          dsn,
          environment: import.meta.env.VITE_SENTRY_ENVIRONMENT || import.meta.env.MODE,
          // Személyes adat nélkül; nyomkövetés (tracing) és session replay sincs: nincs
          // hozzáadva a browserTracingIntegration és a replayIntegration.
          sendDefaultPii: false,
          beforeSend,
          beforeBreadcrumb,
        });
        return Sentry;
      },
      () => null,
    );
  }
  return sdk;
}

export function initSentry(): void {
  if (dsn) void load();
}

/**
 * A render közben dobott hibákat jelenti (React 18: ErrorBoundary). Kikapcsolt
 * hibajelentésnél átlátszó: semmi nem kerül a komponensfába, a viselkedés
 * változatlan. Hiba esetén üres oldalt ad, ami ugyanaz, mint amit a React
 * boundary nélkül is csinálna (leszedi a fát).
 */
export function SentryBoundary({ children }: { children: ReactNode }) {
  if (!dsn) return <>{children}</>;
  return <Boundary>{children}</Boundary>;
}

class Boundary extends Component<{ children: ReactNode }, { failed: boolean }> {
  state = { failed: false };

  static getDerivedStateFromError() {
    return { failed: true };
  }

  componentDidCatch(error: unknown, info: ErrorInfo) {
    // Kezeletlen (handled: false): a boundary nem ad fallbacket, az oldal üres marad.
    void load().then((Sentry) =>
      Sentry?.withScope((scope) => {
        scope.setContext('react', { componentStack: info.componentStack });
        Sentry.captureException(error, { mechanism: { handled: false } });
      }),
    );
  }

  render() {
    return this.state.failed ? null : this.props.children;
  }
}
