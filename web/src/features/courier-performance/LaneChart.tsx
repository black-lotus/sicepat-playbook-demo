import type { DailyCount } from '../../api/types';
import { formatDate } from '../../lib/date';
import { formatNumber } from '../../lib/format-number';

interface LaneChartProps {
  daily: DailyCount[];
}

const WIDTH = 480;
const HEIGHT = 200;
const PAD = { top: 12, right: 12, bottom: 28, left: 36 };

export function LaneChart({ daily }: LaneChartProps) {
  const max = Math.max(1, ...daily.map((day) => Math.max(day.delivered, day.onTime)));
  const innerW = WIDTH - PAD.left - PAD.right;
  const innerH = HEIGHT - PAD.top - PAD.bottom;
  const x = (index: number) => PAD.left + (daily.length <= 1 ? innerW / 2 : (index / (daily.length - 1)) * innerW);
  const y = (value: number) => PAD.top + innerH - (value / max) * innerH;
  const line = (pick: (day: DailyCount) => number) =>
    daily.map((day, index) => `${index === 0 ? 'M' : 'L'}${x(index).toFixed(1)} ${y(pick(day)).toFixed(1)}`).join(' ');

  const first = daily[0];
  const last = daily[daily.length - 1];
  const label = first && last
    ? `Grafik lintasan harian ${formatDate(first.date)} sampai ${formatDate(last.date)}: paket terkirim dan tepat SLA per hari`
    : 'Grafik lintasan harian';

  return (
    <svg className="lane-chart" viewBox={`0 0 ${WIDTH} ${HEIGHT}`} role="img" aria-label={label}>
      <line className="lane-axis" x1={PAD.left} x2={WIDTH - PAD.right} y1={y(0)} y2={y(0)} />
      <line className="lane-grid" x1={PAD.left} x2={WIDTH - PAD.right} y1={y(max)} y2={y(max)} />
      <text className="lane-tick" x={PAD.left - 6} y={y(0)} textAnchor="end" dominantBaseline="middle">0</text>
      <text className="lane-tick" x={PAD.left - 6} y={y(max)} textAnchor="end" dominantBaseline="middle">{formatNumber(max)}</text>
      {first && (
        <text className="lane-tick" x={PAD.left} y={HEIGHT - 8} textAnchor="start">{formatDate(first.date)}</text>
      )}
      {last && daily.length > 1 && (
        <text className="lane-tick" x={WIDTH - PAD.right} y={HEIGHT - 8} textAnchor="end">{formatDate(last.date)}</text>
      )}
      <path className="lane-line lane-line-on-time" d={line((day) => day.onTime)} fill="none" />
      <path className="lane-line lane-line-delivered" d={line((day) => day.delivered)} fill="none" />
      {daily.map((day, index) => (
        <circle key={day.date} className="lane-dot" cx={x(index)} cy={y(day.delivered)} r={3} />
      ))}
    </svg>
  );
}
