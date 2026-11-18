import { useId } from 'react';

import { formatNumber } from '../lib/format-number';
import { formatPercent } from '../lib/format-percent';
import './donut-chart.css';

export type DonutPattern = 'solid' | 'diagonal' | 'dots';

export interface DonutSegment {
  key: string;
  label: string;
  value: number;
  /** Proporsi 0..1 yang sudah dihitung pemanggil dari angka API; tidak pernah NaN. */
  share: number;
  /** Pola isi segmen agar tidak hanya dibedakan warna. Default `solid`. */
  pattern?: DonutPattern;
}

interface DonutChartProps {
  /** Total dari API, ditulis di tengah donut. */
  total: number;
  segments: DonutSegment[];
  /** Keterangan total, ditulis di bawah angka tengah dan di label aksesibel, misalnya "Total paket". */
  label: string;
}

const SIZE = 160;
const CENTER = SIZE / 2;
const RADIUS = 62;
const STROKE = 22;
/** Celah antarsegmen dalam satuan persen lingkaran (pathLength = 100). */
const GAP = 1.2;

/** Donut generik. Segmen dibedakan warna dan pola (polos, garis diagonal, titik); legenda memakai bentuk berbeda. */
export function DonutChart({ total, segments, label }: DonutChartProps) {
  const uid = useId().replace(/:/g, '');
  const gradientID = `donut-gradient-${uid}`;
  const diagonalID = `donut-late-${uid}`;
  const dotsID = `donut-failed-${uid}`;
  const shares = segments;
  const visible = shares.filter((segment) => segment.share > 0);
  const gap = visible.length > 1 ? GAP : 0;
  const stroke: Record<DonutPattern, string> = {
    solid: `url(#${gradientID})`,
    diagonal: `url(#${diagonalID})`,
    dots: `url(#${dotsID})`,
  };

  const arcs = visible.map((segment, index) => {
    const before = visible.slice(0, index).reduce((sum, previous) => sum + previous.share * 100, 0);
    return { segment, length: Math.max(0, segment.share * 100 - gap), offset: -(before + gap / 2) };
  });

  const summary = shares.map((segment) => `${segment.label} ${formatPercent(segment.share)}`).join(', ');

  return (
    <div className="donut">
      <svg className="donut-chart" viewBox={`0 0 ${SIZE} ${SIZE}`} role="img" aria-label={`${label} ${formatNumber(total)}: ${summary}`}>
        <defs>
          <linearGradient id={gradientID} x1="0" y1="0" x2="1" y2="1">
            <stop offset="0" stopColor="var(--color-accent-from)" />
            <stop offset="1" stopColor="var(--color-accent-to)" />
          </linearGradient>
          <pattern id={diagonalID} width="6" height="6" patternUnits="userSpaceOnUse" patternTransform="rotate(45)">
            <rect className="donut-late-bg" width="6" height="6" />
            <rect className="donut-late-fg" width="3" height="6" />
          </pattern>
          <pattern id={dotsID} width="6" height="6" patternUnits="userSpaceOnUse">
            <rect className="donut-failed-bg" width="6" height="6" />
            <circle className="donut-failed-fg" cx="3" cy="3" r="1.5" />
          </pattern>
        </defs>
        <circle className="donut-track" cx={CENTER} cy={CENTER} r={RADIUS} fill="none" strokeWidth={STROKE} />
        {arcs.map(({ segment, length, offset }) => (
          <circle
            key={segment.key}
            className="donut-arc"
            data-segment={segment.key}
            data-share={segment.share}
            cx={CENTER}
            cy={CENTER}
            r={RADIUS}
            fill="none"
            stroke={stroke[segment.pattern ?? 'solid']}
            strokeWidth={STROKE}
            pathLength={100}
            strokeDasharray={`${length} ${100 - length}`}
            strokeDashoffset={offset}
            transform={`rotate(-90 ${CENTER} ${CENTER})`}
          />
        ))}
        <text className="donut-total" x={CENTER} y={CENTER - 2} textAnchor="middle">{formatNumber(total)}</text>
        <text className="donut-total-label" x={CENTER} y={CENTER + 16} textAnchor="middle">{label}</text>
      </svg>
      <ul className="donut-legend">
        {shares.map((segment) => (
          <li key={segment.key} data-segment={segment.key}>
            <span className={`legend-mark legend-mark-${segment.pattern ?? 'solid'}`} aria-hidden="true" />
            <span className="legend-name">{segment.label}</span>
            <span className="legend-count">{formatNumber(segment.value)}</span>
            <span className="legend-percent">{formatPercent(segment.share)}</span>
          </li>
        ))}
      </ul>
    </div>
  );
}
