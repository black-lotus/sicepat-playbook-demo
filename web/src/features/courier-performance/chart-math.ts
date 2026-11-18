/** Langkah sumbu yang enak dibaca; dikalikan pangkat sepuluh. */
const NICE_STEPS = [1, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10];

/** Skala sumbu Y bilangan bulat dengan tiga interval (0, step, 2 step, 3 step) yang mencakup `max`. */
export function niceScale(max: number): { step: number; top: number } {
  const raw = Math.max(1, max) / 3;
  const magnitude = 10 ** Math.floor(Math.log10(raw));
  const factor = NICE_STEPS.find((candidate) => candidate >= raw / magnitude) ?? 10;
  const step = Math.max(1, Math.ceil(factor * magnitude));
  return { step, top: step * 3 };
}

export interface CompositionInput {
  delivered: number;
  onTime: number;
  failed: number;
}

export type CompositionKey = 'on-time' | 'late' | 'failed';

export interface CompositionShare {
  key: CompositionKey;
  label: string;
  value: number;
  /** Proporsi 0..1 terhadap jumlah ketiga segmen; 0 bila jumlahnya 0 (tidak pernah NaN). */
  share: number;
}

/** Memecah paket menjadi tepat SLA, terlambat, dan gagal dari angka API. Hanya proporsi tampilan yang dihitung di sini. */
export function compositionShares({ delivered, onTime, failed }: CompositionInput): CompositionShare[] {
  const onTimeValue = Math.max(0, Math.min(onTime, delivered));
  const lateValue = Math.max(0, delivered - onTimeValue);
  const failedValue = Math.max(0, failed);
  const sum = onTimeValue + lateValue + failedValue;
  const share = (value: number) => (sum > 0 ? value / sum : 0);
  return [
    { key: 'on-time', label: 'Tepat SLA', value: onTimeValue, share: share(onTimeValue) },
    { key: 'late', label: 'Terlambat', value: lateValue, share: share(lateValue) },
    { key: 'failed', label: 'Gagal', value: failedValue, share: share(failedValue) },
  ];
}
