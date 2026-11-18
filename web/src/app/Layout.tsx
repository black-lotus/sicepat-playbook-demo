import { NavLink, Outlet } from 'react-router-dom';

export function Layout() {
  return (
    <div className="layout">
      <header className="topbar">
        <span className="brand">Performa SiGesit</span>
        <nav>
          <NavLink to="/courier-performance">Performa Kurir</NavLink>
          <NavLink to="/branch-summary">Ringkasan Cabang</NavLink>
          <NavLink to="/failed-delivery">Gagal Kirim</NavLink>
        </nav>
      </header>
      <main>
        <Outlet />
      </main>
    </div>
  );
}
