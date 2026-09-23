import { useMemo, type ReactNode } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import Header from './components/Header';
import Footer from './components/Footer';
import HeadMeta from './components/HeadMeta';
import Landing from './pages/Landing';
import Install from './pages/Install';
import Profile from './pages/Profile';
import PrivacyPolicy from './pages/PrivacyPolicy';
import Terms from './pages/Terms';
import { NotYetProvider } from './NotYetProvider';
import { AnalyticsConsentProvider } from './features/consent/AnalyticsConsentProvider';
import ConsentBanner from './features/consent/ConsentBanner';
import EmailVerifyBanner from './features/emailVerify/EmailVerifyBanner';
import { DEFAULT_LANG, LanguageProvider, PATHS, entryLang } from './i18n';

/**
 * A nyelvsemleges "/" az egyetlen pont, ahol nyelvet tippelünk. Aki már
 * választott (a fejlécben), annak a döntése erősebb a böngésző beállításánál,
 * a /en és a /de pedig soha nem irányít vissza - így nem lehet beleragadni
 * egy nyelvbe.
 */
function EntryPoint({ children }: { children: ReactNode }) {
  const lang = useMemo(() => entryLang(), []);
  if (lang !== DEFAULT_LANG) return <Navigate to={PATHS[lang].home} replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <LanguageProvider>
      <NotYetProvider>
        <AnalyticsConsentProvider>
          <HeadMeta />
          <div className="app">
            <Header />
            <Routes>
              <Route
                path={PATHS.hu.home}
                element={
                  <EntryPoint>
                    <Landing />
                  </EntryPoint>
                }
              />
              <Route path={PATHS.hu.install} element={<Install />} />
              <Route path={PATHS.hu.profile} element={<Profile />} />

              <Route path={PATHS.en.home} element={<Landing />} />
              <Route path={PATHS.en.install} element={<Install />} />
              <Route path={PATHS.en.profile} element={<Profile />} />

              <Route path={PATHS.de.home} element={<Landing />} />
              <Route path={PATHS.de.install} element={<Install />} />
              <Route path={PATHS.de.profile} element={<Profile />} />

              {/* Nyelvsemleges, nyelv-előtag nélküli jogi oldalak: a jogi dokumentumok
                  egyelőre kizárólag magyarul készülnek el, ezért NEM tartoznak a
                  PATHS/RouteKey rendszerhez - ugyanaz az útvonal minden nyelvről. */}
              <Route path="/adatvedelem" element={<PrivacyPolicy />} />
              <Route path="/aszf" element={<Terms />} />

              {/* Ismeretlen /en/... vagy /de/... útvonalról az adott nyelv
                  főoldalára megyünk, hogy a látogató ne essen át egy másikra. */}
              <Route path="/en/*" element={<Navigate to={PATHS.en.home} replace />} />
              <Route path="/de/*" element={<Navigate to={PATHS.de.home} replace />} />
              <Route path="*" element={<Navigate to={PATHS.hu.home} replace />} />
            </Routes>
            <Footer />
            <ConsentBanner />
            <EmailVerifyBanner />
          </div>
        </AnalyticsConsentProvider>
      </NotYetProvider>
    </LanguageProvider>
  );
}
