const percentFormatter = new Intl.NumberFormat('id-ID', {
  style: 'percent',
  minimumFractionDigits: 1,
  maximumFractionDigits: 1,
});

/** Memformat rasio 0..1 dari API menjadi persen, misalnya 0.8 menjadi "80,0%". */
export function formatPercent(ratio: number): string {
  return percentFormatter.format(ratio);
}
