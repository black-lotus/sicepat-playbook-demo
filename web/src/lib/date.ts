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

const monthFormatter = new Intl.DateTimeFormat('id-ID', { month: 'short', timeZone: 'UTC' });

function dayAndMonth(isoDate: string): { day: number; month: string } {
  const date = new Date(`${isoDate}T00:00:00Z`);
  return { day: date.getUTCDate(), month: monthFormatter.format(date) };
}

function isValidISODate(isoDate: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(isoDate)) return false;
  return !Number.isNaN(new Date(`${isoDate}T00:00:00Z`).getTime());
}

/**
 * Rentang tanggal ringkas, misalnya "1 sampai 7 Nov" atau "28 Okt sampai 3 Nov".
 * Bila salah satu tanggal kosong atau tidak valid (input date yang dikosongkan), mengembalikan "Pilih periode".
 */
export function formatDateRange(startDate: string, endDate: string): string {
  if (!isValidISODate(startDate) || !isValidISODate(endDate)) return 'Pilih periode';
  const start = dayAndMonth(startDate);
  const end = dayAndMonth(endDate);
  if (startDate === endDate) return `${end.day} ${end.month}`;
  if (startDate.slice(0, 7) === endDate.slice(0, 7)) return `${start.day} sampai ${end.day} ${end.month}`;
  return `${start.day} ${start.month} sampai ${end.day} ${end.month}`;
}
