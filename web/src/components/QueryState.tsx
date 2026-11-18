import type { ReactNode } from 'react';

interface QueryStateProps {
  isPending: boolean;
  error: Error | null;
  isEmpty: boolean;
  emptyMessage: string;
  onRetry: () => void;
  children: ReactNode;
}

/** Menampilkan state loading, error, dan data kosong secara seragam di semua halaman. */
export function QueryState({ isPending, error, isEmpty, emptyMessage, onRetry, children }: QueryStateProps) {
  if (isPending) {
    return (
      <p className="state" role="status">
        Memuat data…
      </p>
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
