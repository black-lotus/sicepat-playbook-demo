import type { CSSProperties } from 'react';

import type { DailyCount } from '../api/types';
import { formatNumber } from '../lib/format-number';
import { targetLabels, type TargetRates } from '../lib/target-labels';

interface DailyStripProps {
  daily: DailyCount[];
  size: 'sm' | 'md';
  /** Bila true, strip hanya dekorasi: disembunyikan dari teknologi bantu (dipakai di dalam tombol podium). */
  decorative?: boolean;
  /** Target bulanan cabang (0 sampai 1). Hanya diberikan bila cabang punya target; menggambar dua garis horizontal. */
  targets?: TargetRates;
}


/** Di atas jumlah hari ini, strip ukuran sm memakai lebar minimum per sel supaya pola telat tetap terbaca. */
const DENSE_FROM_DAYS = 14;

function percent(part: number, whole: number): number {
  if (whole <= 0) return 0;
  return Math.min(100, Math.max(0, (part / whole) * 100));
}

export function DailyStrip({ daily, size, decorative = false, targets }: DailyStripProps) {
  if (daily.length <= 1) return null;

  const delivered = daily.reduce((sum, day) => sum + day.delivered, 0);
  const total = daily.reduce((sum, day) => sum + day.totalPackages, 0);
  const style = { gridTemplateColumns: `repeat(${daily.length}, minmax(0, 1fr))`, '--days': daily.length } as CSSProperties;
  const dense = size === 'sm' && daily.length > DENSE_FROM_DAYS;
  const labels = targets ? targetLabels(targets) : null;
  const a11y = decorative
    ? { 'aria-hidden': true as const }
    : { role: 'img' as const, 'aria-label': `Tren harian ${daily.length} hari: terkirim ${formatNumber(delivered)} dari ${formatNumber(total)}${labels ? `. ${labels.success}. ${labels.sla}` : ''}` };

  return (
    <span className={`daily-strip daily-strip-${size}${dense ? ' daily-strip-dense' : ''}`} style={style} {...a11y}>
      {daily.map((day) => {
        const onTime = Math.min(day.onTime, day.delivered);
        const late = day.delivered - onTime;
        return (
          <span key={day.date} className="daily-day" data-day={day.date} data-empty={day.totalPackages === 0 ? 'true' : undefined}>
            {late > 0 && <span className="daily-part daily-late" data-part="late" style={{ height: `${percent(late, day.totalPackages)}%` }} />}
            {onTime > 0 && <span className="daily-part daily-on-time" data-part="on-time" style={{ height: `${percent(onTime, day.totalPackages)}%` }} />}
          </span>
        );
      })}
      {targets && labels && (
        <>
          <span className="daily-target" data-target="success" data-line="solid" style={{ bottom: `${percent(targets.successRate, 1)}%` }}>
            <span className="daily-target-label">{labels.success}</span>
          </span>
          <span className="daily-target" data-target="sla" data-line="dashed" style={{ bottom: `${percent(targets.slaRate, 1)}%` }}>
            <span className="daily-target-label">{labels.sla}</span>
          </span>
        </>
      )}
    </span>
  );
}
