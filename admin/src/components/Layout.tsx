import { NavLink, Outlet } from 'react-router-dom';
import { useAuth } from '../AuthProvider';

const NAV = [
  { to: '/', label: 'Vezérlőpult', end: true },
  { to: '/felhasznalok', label: 'Felhasználók', end: false },
  { to: '/elofizetesek', label: 'Előfizetések', end: false },
  { to: '/szamlazas', label: 'Számlázás', end: false },
];

export function Layout() {
  const { profile, logout, accountUrl } = useAuth();

  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <span className="brand__mark" aria-hidden="true" />
          <span>
            <strong>Snitt</strong>
            <span className="brand__sub">admin</span>
          </span>
        </div>

        <nav className="nav">
          {NAV.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) => `nav__item ${isActive ? 'is-active' : ''}`}
            >
              {item.label}
            </NavLink>
          ))}
        </nav>

        <div className="sidebar__foot">
          <div className="who">
            <strong>{profile?.name ?? '–'}</strong>
            <span className="muted">{profile?.email}</span>
          </div>
          <a className="btn btn--ghost" href={accountUrl} target="_blank" rel="noreferrer">
            Fiók
          </a>
          <button type="button" className="btn btn--ghost" onClick={logout}>
            Kijelentkezés
          </button>
        </div>
      </aside>

      <main className="main">
        <Outlet />
      </main>
    </div>
  );
}

export function Page({
  title,
  description,
  actions,
  children,
}: {
  title: string;
  description?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
}) {
  return (
    <>
      <header className="page-head">
        <div>
          <h1>{title}</h1>
          {description ? <p className="muted">{description}</p> : null}
        </div>
        {actions ? <div className="page-head__actions">{actions}</div> : null}
      </header>
      {children}
    </>
  );
}
