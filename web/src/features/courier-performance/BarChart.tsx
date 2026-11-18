import { useId, useState } from 'react';

import type { DailyCount } from '../../api/types';
import { formatDate, formatDayMonth } from '../../lib/date';
import { formatNumber } from '../../lib/format-number';
import { niceScale } from './chart-math';

interface BarChartProps {
  daily: DailyCount[];
}

const WIDTH = 480;
const HEIGHT = 250;
const PAD = { top: 56, right: 8, bottom: 30, left: 34 };
const TIP = { width: 108, height: 42 };
const MAX_X_LABELS = 7;

/** Grafik batang paket terkirim per hari. Batang tertinggi tersorot gradien dengan tooltip; hover memindahkan tooltip. */
export function BarChart({ daily }: BarChartProps) {
  const gradientID = `bar-gradient-${useId().replace(/:/g, '')}`;
  const [hovered, setHovered] = useState<number | null>(null);

  const max = Math.max(0, ...daily.map((day) => day.delivered));
  const { step, top } = niceScale(max);
  const innerW = WIDTH - PAD.left - PAD.right;
  const innerH = HEIGHT - PAD.top - PAD.bottom;
  const slot = daily.length > 0 ? innerW / daily.length : innerW;
  const barW = Math.min(slot * 0.62, 28);
  const y = (value: number) => PAD.top + innerH - (value / top) * innerH;
  const peak = max > 0 ? daily.findIndex((day) => day.delivered === max) : -1;
  const active = hovered ?? (peak >= 0 ? peak : null);
  const labelEvery = Math.max(1, Math.ceil(daily.length / MAX_X_LABELS));

  const first = daily[0];
  const last = daily[daily.length - 1];
  const label = first && last
    ? `Grafik batang paket terkirim per hari, ${formatDate(first.date)} sampai ${formatDate(last.date)}`
    : 'Grafik batang paket terkirim per hari';

  const barX = (index: number) => PAD.left + index * slot + (slot - barW) / 2;
  const activeDay = active === null ? undefined : daily[active];
  const tipX = active === null ? 0 : Math.min(Math.max(barX(active) + barW / 2 - TIP.width / 2, PAD.left), WIDTH - PAD.right - TIP.width);
  const tipY = activeDay ? Math.max(2, y(activeDay.delivered) - TIP.height - 10) : 0;

  return (
    <svg className="bar-chart" viewBox={`0 0 ${WIDTH} ${HEIGHT}`} role="img" aria-label={label}>
      <defs>
        <linearGradient id={gradientID} x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="var(--color-accent-to)" />
          <stop offset="1" stopColor="var(--color-accent-from)" />
        </linearGradient>
      </defs>
      {[0, 1, 2, 3].map((tick) => (
        <g key={tick}>
          <line className={tick === 0 ? 'bar-axis' : 'bar-grid'} x1={PAD.left} x2={WIDTH - PAD.right} y1={y(tick * step)} y2={y(tick * step)} />
          <text className="bar-tick" x={PAD.left - 8} y={y(tick * step)} textAnchor="end" dominantBaseline="middle">{formatNumber(tick * step)}</text>
        </g>
      ))}
      {daily.map((day, index) => {
        const height = (day.delivered / top) * innerH;
        const isActive = index === active;
        return (
          <g key={day.date}>
            <rect
              className={`bar${isActive ? ' bar-active' : ''}`}
              data-day={day.date}
              data-peak={index === peak ? 'true' : undefined}
              data-active={isActive ? 'true' : undefined}
              x={barX(index)}
              y={y(day.delivered)}
              width={barW}
              height={height}
              rx={Math.min(6, barW / 2)}
              style={isActive ? { fill: `url(#${gradientID})` } : undefined}
            />
            {index % labelEvery === 0 && (
              <text className="bar-tick" x={barX(index) + barW / 2} y={HEIGHT - 8} textAnchor="middle">{formatDayMonth(day.date)}</text>
            )}
            <rect
              className="bar-hit"
              x={PAD.left + index * slot}
              y={PAD.top}
              width={slot}
              height={innerH}
              fill="transparent"
              onPointerEnter={() => setHovered(index)}
              onPointerLeave={() => setHovered(null)}
            />
          </g>
        );
      })}
      {activeDay && (
        <g className="bar-tooltip" data-day={activeDay.date} transform={`translate(${tipX} ${tipY})`} pointerEvents="none">
          <rect width={TIP.width} height={TIP.height} rx={12} />
          <text className="bar-tooltip-date" x={TIP.width / 2} y={17} textAnchor="middle">{formatDate(activeDay.date)}</text>
          <text className="bar-tooltip-value" x={TIP.width / 2} y={34} textAnchor="middle">{formatNumber(activeDay.delivered)} terkirim</text>
        </g>
      )}
    </svg>
  );
}
