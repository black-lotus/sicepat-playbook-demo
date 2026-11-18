import { Link } from 'react-router-dom';

/** Pengaman tingkat router: ditampilkan bila render sebuah halaman melempar error. */
export function RouteError() {
  return (
    <main className="route-error" role="alert">
      <p>Terjadi kesalahan pada halaman ini.</p>
      <Link to="/courier-performance">Kembali ke Performa Kurir</Link>
    </main>
  );
}
