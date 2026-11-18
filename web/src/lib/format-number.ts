const numberFormatter = new Intl.NumberFormat('id-ID');

/** Memformat bilangan bulat dengan pemisah ribuan, misalnya 12500 menjadi "12.500". */
export function formatNumber(value: number): string {
  return numberFormatter.format(value);
}
