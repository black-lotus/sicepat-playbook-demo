import type { DailyCount } from '../api/types';
import type { TargetRates } from './target-labels';

/** Satu batang grafik harian. Semua proporsi adalah persen tampilan (0..100) yang dihitung dari angka API. */
export interface DailyBar {
  date: string;
  total: number;
  delivered: number;
  onTime: number;
  late: number;
  failed: number;
  /** Tinggi batang relatif terhadap hari dengan paket terbanyak di strip yang sama. */
  height: number;
  /** Tinggi tiap segmen sebagai persen dari tinggi batang. */
  parts: { onTime: number; late: number; failed: number };
  /** Terkirim dibagi total (0..1). Null bila hari itu tanpa paket. */
  success: number | null;
  empty: boolean;
}

function round(value: number): number {
  return Math.round(value * 100) / 100;
}

function share(part: number, whole: number): number {
  if (whole <= 0) return 0;
  return round(Math.min(100, Math.max(0, (part / whole) * 100)));
}

/** Total yang dipakai sebagai skala: total API, atau jumlah segmen bila data tidak konsisten, supaya segmen tak melewati batang. */
function volume(day: DailyCount): number {
  const onTime = Math.min(day.onTime, day.delivered);
  return Math.max(day.totalPackages, day.delivered + day.failed, onTime);
}

export function dailyBars(daily: DailyCount[]): DailyBar[] {
  const peak = daily.reduce((max, day) => Math.max(max, volume(day)), 0);
  return daily.map((day) => {
    const onTime = Math.max(0, Math.min(day.onTime, day.delivered));
    const late = Math.max(0, day.delivered - onTime);
    const failed = Math.max(0, day.failed);
    const base = volume(day);
    return {
      date: day.date,
      total: day.totalPackages,
      delivered: day.delivered,
      onTime,
      late,
      failed,
      height: share(base, peak),
      parts: { onTime: share(onTime, base), late: share(late, base), failed: share(failed, base) },
      success: day.totalPackages > 0 ? day.delivered / day.totalPackages : null,
      empty: base === 0,
    };
  });
}

/**
 * Posisi garis target di dalam satu batang, persen dari tinggi batang itu.
 * Keberhasilan = terkirim/total, jadi targetnya setinggi `successRate` dari total.
 * Tepat SLA = tepat/terkirim, jadi targetnya setinggi `slaRate` dari jumlah terkirim.
 */
export function targetMarkers(bar: DailyBar, targets: TargetRates): { success: number; sla: number } | null {
  const base = Math.max(bar.total, bar.delivered + bar.failed);
  if (bar.empty || base <= 0) return null;
  return {
    success: share(targets.successRate * bar.total, base),
    sla: share(targets.slaRate * bar.delivered, base),
  };
}

export function dailySummary(daily: DailyCount[]): { total: number; delivered: number; onTime: number; late: number; failed: number } {
  const bars = dailyBars(daily);
  return bars.reduce(
    (sum, bar) => ({
      total: sum.total + bar.total,
      delivered: sum.delivered + bar.delivered,
      onTime: sum.onTime + bar.onTime,
      late: sum.late + bar.late,
      failed: sum.failed + bar.failed,
    }),
    { total: 0, delivered: 0, onTime: 0, late: 0, failed: 0 },
  );
}
