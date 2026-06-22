import { NavLink, Outlet } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { routes } from '../routes';

const navItems = [
  { to: routes.dashboard, label: 'Dashboard', end: true },
  { to: routes.users, label: 'Users' },
  { to: routes.subscriptions, label: 'Subscriptions' },
  { to: routes.chat, label: 'Chat History' },
  { to: routes.profileRoasts, label: 'Profile Roasts' },
  { to: routes.llm, label: 'LLM Config' },
];

export function Layout() {
  const { logout } = useAuth();

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="brand">Agrobot Admin</div>
        <nav>
          {navItems.map((item) => (
            <NavLink key={item.to} to={item.to} end={item.end} className="nav-link">
              {item.label}
            </NavLink>
          ))}
        </nav>
        <button type="button" className="btn btn-secondary logout" onClick={logout}>
          Logout
        </button>
      </aside>
      <main className="content">
        <Outlet />
      </main>
    </div>
  );
}
