import type { ReactNode } from 'react';

interface QueryStateProps {
  isPending: boolean;
  error: Error | null;
  isEmpty: boolean;
  emptyMessage: string;
  onRetry: () => void;
  loadingFallback?: ReactNode;
  children: ReactNode;
}

/** Menampilkan state loading, error, dan data kosong secara seragam di semua halaman. */
export function QueryState({ isPending, error, isEmpty, emptyMessage, onRetry, loadingFallback, children }: QueryStateProps) {
  if (isPending) {
    return (
      <div className="state" role="status">
        <p>Memuat data…</p>
        {loadingFallback}
      </div>
    );
  }
  if (error) {
    return (
      <div className="state state-error" role="alert">
        <p>Gagal memuat data: {error.message}</p>
        <button type="button" onClick={onRetry}>
          Coba lagi
        </button>
      </div>
    );
  }
  if (isEmpty) {
    return <p className="state">{emptyMessage}</p>;
  }
  return <>{children}</>;
}
