const dateFormatter = new Intl.DateTimeFormat('id-ID', {
  day: 'numeric',
  month: 'short',
  year: 'numeric',
  timeZone: 'UTC',
});

/** Mengubah Date menjadi string YYYY-MM-DD (format tanggal API) berdasarkan tanggal lokal. */
export function toISODate(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

/** Periode default dashboard: 7 hari terakhir sampai hari ini. */
export function lastSevenDays(today: Date): { startDate: string; endDate: string } {
  const start = new Date(today);
  start.setDate(start.getDate() - 6);
  return { startDate: toISODate(start), endDate: toISODate(today) };
}

/** Memformat tanggal YYYY-MM-DD dari API untuk ditampilkan, misalnya "17 Nov 2026". */
export function formatDate(isoDate: string): string {
  return dateFormatter.format(new Date(`${isoDate}T00:00:00Z`));
}
