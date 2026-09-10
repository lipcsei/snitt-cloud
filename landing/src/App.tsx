import { Navigate, Route, Routes } from 'react-router-dom';
import Header from './components/Header';
import Footer from './components/Footer';
import Landing from './pages/Landing';
import Install from './pages/Install';
import Profile from './pages/Profile';
import { NotYetProvider } from './NotYetProvider';

export default function App() {
  return (
    <NotYetProvider>
      <div className="app">
        <Header />
        <Routes>
          <Route path="/" element={<Landing />} />
          <Route path="/telepites" element={<Install />} />
          <Route path="/profil" element={<Profile />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
        <Footer />
      </div>
    </NotYetProvider>
  );
}
