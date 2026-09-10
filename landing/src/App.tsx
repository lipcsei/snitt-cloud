import { useMemo, type ReactNode } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import Header from './components/Header';
import Footer from './components/Footer';
import HeadMeta from './components/HeadMeta';
import Landing from './pages/Landing';
import Install from './pages/Install';
import Profile from './pages/Profile';
import { NotYetProvider } from './NotYetProvider';
import { LanguageProvider, PATHS, prefersEnglishOnEntry } from './i18n';

/**
 * A nyelvsemleges "/" az egyetlen pont, ahol nyelvet tippelünk. Aki már
 * választott (a fejlécben), annak a döntése erősebb a böngésző beállításánál,
 * a /en pedig soha nem irányít vissza - így nem lehet beleragadni egy nyelvbe.
 */
function EntryPoint({ children }: { children: ReactNode }) {
  const toEnglish = useMemo(() => prefersEnglishOnEntry(), []);
  if (toEnglish) return <Navigate to={PATHS.en.home} replace />;
  return <>{children}</>;
}

export default function App() {
  return (
    <LanguageProvider>
      <NotYetProvider>
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

            {/* Ismeretlen /en/... útvonalról az angol főoldalra megyünk, hogy a
                látogató ne essen át a másik nyelvre. */}
            <Route path="/en/*" element={<Navigate to={PATHS.en.home} replace />} />
            <Route path="*" element={<Navigate to={PATHS.hu.home} replace />} />
          </Routes>
          <Footer />
        </div>
      </NotYetProvider>
    </LanguageProvider>
  );
}
