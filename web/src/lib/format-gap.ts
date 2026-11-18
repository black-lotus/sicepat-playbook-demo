const gapFormatter = new Intl.NumberFormat('id-ID', {
  minimumFractionDigits: 1,
  maximumFractionDigits: 1,
});

/** Memformat selisih rasio 0..1 dari pemimpin menjadi poin persen, misalnya 0.0429 menjadi "−4,3 pp". */
export function formatGap(gap: number): string {
  // Tanda ditentukan dari nilai yang sudah dibulatkan, supaya gap 0.0004 menjadi "0,0 pp", bukan "−0,0 pp".
  const rounded = Math.round(gap * 1000) / 10;
  if (rounded === 0) return `${gapFormatter.format(0)} pp`;
  const prefix = rounded > 0 ? '−' : '';
  return `${prefix}${gapFormatter.format(rounded)} pp`;
}
