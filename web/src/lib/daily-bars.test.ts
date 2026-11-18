import type { DailyCount } from '../api/types';
import { dailyBars, dailySummary, targetMarkers, type DailyBar } from './daily-bars';

function day(date: string, totalPackages: number, delivered: number, failed: number, onTime: number): DailyCount {
  return { date, totalPackages, delivered, failed, onTime };
}

const week: DailyCount[] = [
  day('2026-11-01', 10, 8, 2, 6),
  day('2026-11-02', 5, 5, 0, 5),
  day('2026-11-03', 0, 0, 0, 0),
  day('2026-11-04', 20, 15, 5, 10),
];

function barAt(daily: DailyCount[], index: number): DailyBar {
  const bar = dailyBars(daily)[index];
  if (!bar) throw new Error(`batang ${index} tidak ada`);
  return bar;
}

describe('dailyBars', () => {
  it('tinggi batang relatif terhadap hari dengan paket terbanyak', () => {
    const bars = dailyBars(week);

    expect(bars.map((bar) => bar.height)).toEqual([50, 25, 0, 100]);
  });

  it('membagi segmen dari angka API: tepat SLA, terlambat, gagal', () => {
    const bar = barAt(week, 3);

    expect(bar).toMatchObject({ total: 20, onTime: 10, late: 5, failed: 5 });
    expect(bar.parts).toEqual({ onTime: 50, late: 25, failed: 25 });
  });

  it('menandai hari tanpa paket sebagai kosong tanpa NaN dan tanpa keberhasilan', () => {
    const bar = barAt(week, 2);

    expect(bar).toMatchObject({ empty: true, height: 0, success: null, parts: { onTime: 0, late: 0, failed: 0 } });
  });

  it('semua hari kosong tidak menghasilkan NaN', () => {
    const bars = dailyBars([day('2026-11-01', 0, 0, 0, 0), day('2026-11-02', 0, 0, 0, 0)]);

    expect(JSON.stringify(bars)).not.toMatch(/NaN|null,"height"/);
    expect(bars.every((bar) => bar.height === 0 && bar.empty)).toBe(true);
  });

  it('keberhasilan hari itu adalah terkirim dibagi total', () => {
    expect(dailyBars(week).map((bar) => bar.success)).toEqual([0.8, 1, null, 0.75]);
  });

  it('menjepit tepat SLA yang melebihi terkirim dan menjaga segmen tidak melewati tinggi batang', () => {
    const bar = barAt([day('2026-11-01', 4, 3, 1, 9)], 0);

    expect(bar).toMatchObject({ onTime: 3, late: 0, failed: 1 });
    expect(bar.parts.onTime + bar.parts.late + bar.parts.failed).toBeLessThanOrEqual(100);
  });

  it('daftar kosong menghasilkan daftar kosong', () => {
    expect(dailyBars([])).toEqual([]);
  });
});

describe('targetMarkers', () => {
  it('menaruh penanda keberhasilan dan tepat SLA relatif terhadap tinggi batang hari itu', () => {
    const bar = barAt(week, 3);

    // 20 paket, 15 terkirim: target keberhasilan 90% = 90% tinggi batang; target SLA 80% dari 15 terkirim = 60% tinggi batang.
    expect(targetMarkers(bar, { successRate: 0.9, slaRate: 0.8 })).toEqual({ success: 90, sla: 60 });
  });

  it('hari kosong tidak punya penanda', () => {
    expect(targetMarkers(barAt(week, 2), { successRate: 0.9, slaRate: 0.8 })).toBeNull();
  });
});

describe('dailySummary', () => {
  it('menjumlahkan seluruh hari', () => {
    expect(dailySummary(week)).toEqual({ total: 35, delivered: 28, onTime: 21, late: 7, failed: 7 });
  });
});
