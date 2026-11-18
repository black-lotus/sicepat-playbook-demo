import { useEffect, useId, useRef, useState, type CSSProperties, type KeyboardEvent } from 'react';

import type { DailyCount } from '../api/types';
import { dailyBars, dailySummary, targetMarkers, type DailyBar } from '../lib/daily-bars';
import { formatDayLabel, formatWeekdayDate } from '../lib/date';
import { formatNumber } from '../lib/format-number';
import { formatPercent } from '../lib/format-percent';
import { targetLabels, type TargetRates } from '../lib/target-labels';
import { DailyTooltip } from './DailyTooltip';
import './daily-strip.css';

interface DailyStripProps {
  daily: DailyCount[];
  size: 'sm' | 'md';
  /** Target bulanan cabang (0 sampai 1). Hanya diberikan bila cabang punya target; menggambar penanda di tiap batang. */
  targets?: TargetRates;
}

/** Di atas jumlah hari ini, strip ukuran sm memakai lebar minimum per batang supaya segmen tetap terbaca. */
const DENSE_FROM_DAYS = 14;
/** Jumlah label sumbu maksimum pada varian md; sisanya dijarangkan. */
const MAX_AXIS_LABELS = 8;

interface Active {
  index: number;
  el: HTMLElement;
}

function describeBar(bar: DailyBar): string {
  const success = bar.success === null ? 'tanpa paket' : `keberhasilan ${formatPercent(bar.success)}`;
  return `${formatWeekdayDate(bar.date)}: ${formatNumber(bar.total)} paket, tepat SLA ${formatNumber(bar.onTime)}, terlambat ${formatNumber(bar.late)}, gagal ${formatNumber(bar.failed)}, ${success}`;
}

/**
 * Grafik batang bertumpuk per hari. Tinggi batang = total paket hari itu relatif terhadap hari terbanyak
 * di strip yang sama; dari bawah: tepat SLA (penuh), terlambat (garis miring), gagal (titik).
 * Semua angka dari API, komponen ini hanya menghitung proporsi tampilan.
 */
export function DailyStrip({ daily, size, targets }: DailyStripProps) {
  const tooltipID = useId();
  const root = useRef<HTMLDivElement>(null);
  const [active, setActive] = useState<Active | null>(null);
  const [tabStop, setTabStop] = useState(0);
  const isActive = active !== null;

  // Sentuh di luar strip, gulir, atau ubah ukuran layar menutup keterangan (posisinya fixed, jadi akan basi).
  useEffect(() => {
    if (!isActive) return;
    const close = () => setActive(null);
    const onPointerDown = (event: PointerEvent) => {
      if (!root.current?.contains(event.target as Node)) close();
    };
    document.addEventListener('pointerdown', onPointerDown);
    window.addEventListener('scroll', close, true);
    window.addEventListener('resize', close);
    return () => {
      document.removeEventListener('pointerdown', onPointerDown);
      window.removeEventListener('scroll', close, true);
      window.removeEventListener('resize', close);
    };
  }, [isActive]);

  if (daily.length <= 1) return null;

  const bars = dailyBars(daily);
  const sum = dailySummary(daily);
  const labels = targets ? targetLabels(targets) : null;
  const dense = size === 'sm' && daily.length > DENSE_FROM_DAYS;
  const style = { '--days': daily.length } as CSSProperties;
  const summary = `Tren harian ${daily.length} hari: terkirim ${formatNumber(sum.delivered)} dari ${formatNumber(sum.total)}, tepat SLA ${formatNumber(sum.onTime)}, terlambat ${formatNumber(sum.late)}, gagal ${formatNumber(sum.failed)}${labels ? `. ${labels.success}. ${labels.sla}` : ''}`;
  const axisStep = Math.max(1, Math.ceil(bars.length / MAX_AXIS_LABELS));
  const activeBar = active ? bars[active.index] : undefined;
  const focusIndex = Math.min(tabStop, bars.length - 1);

  const show = (index: number) => (event: { currentTarget: HTMLElement }) => setActive({ index, el: event.currentTarget });

  const onKeyDown = (event: KeyboardEvent<HTMLElement>) => {
    if (event.key === 'Escape') {
      setActive(null);
      return;
    }
    const all = Array.from(event.currentTarget.parentElement?.querySelectorAll<HTMLElement>('[data-day]') ?? []);
    const at = all.indexOf(event.currentTarget);
    const next =
      event.key === 'ArrowRight' ? at + 1 : event.key === 'ArrowLeft' ? at - 1 : event.key === 'Home' ? 0 : event.key === 'End' ? all.length - 1 : null;
    if (next === null) return;
    event.preventDefault();
    all[Math.min(Math.max(next, 0), all.length - 1)]?.focus();
  };

  return (
    <div
      ref={root}
      className={`daily-chart daily-chart-${size}${dense ? ' daily-chart-dense' : ''}`}
      style={style}
      role="group"
      aria-label={summary}
    >
      <div className="daily-strip">
        {bars.map((bar, index) => {
          const marks = targets ? targetMarkers(bar, targets) : null;
          return (
            <span
              key={bar.date}
              className="daily-day"
              role="img"
              aria-label={describeBar(bar)}
              tabIndex={index === focusIndex ? 0 : -1}
              data-day={bar.date}
              data-empty={bar.empty ? 'true' : undefined}
              data-active={active?.index === index ? 'true' : undefined}
              onMouseEnter={show(index)}
              onMouseLeave={() => setActive(null)}
              onFocus={(event) => {
                setTabStop(index);
                show(index)(event);
              }}
              onBlur={() => setActive(null)}
              onClick={show(index)}
              onKeyDown={onKeyDown}
            >
              {!bar.empty && (
                <span className="daily-stack" style={{ height: `${bar.height}%` }}>
                  {bar.onTime > 0 && <span className="daily-part daily-on-time" data-part="on-time" data-count={bar.onTime} style={{ height: `${bar.parts.onTime}%` }} />}
                  {bar.late > 0 && <span className="daily-part daily-late" data-part="late" data-count={bar.late} style={{ height: `${bar.parts.late}%` }} />}
                  {bar.failed > 0 && <span className="daily-part daily-failed" data-part="failed" data-count={bar.failed} style={{ height: `${bar.parts.failed}%` }} />}
                  {marks && (
                    <>
                      <span className="daily-target" data-target="success" data-line="solid" style={{ bottom: `${marks.success}%` }} />
                      <span className="daily-target" data-target="sla" data-line="dashed" style={{ bottom: `${marks.sla}%` }} />
                    </>
                  )}
                </span>
              )}
            </span>
          );
        })}
      </div>
      {size === 'md' && (
        <>
          <div className="daily-axis" aria-hidden="true">
            {bars.map((bar, index) => (
              <span key={bar.date}>{index % axisStep === 0 ? formatDayLabel(bar.date, index === 0 || bar.date.endsWith('-01')) : ''}</span>
            ))}
          </div>
          <div className="daily-legend">
            <span><span className="daily-swatch daily-on-time" aria-hidden="true" />Tepat SLA</span>
            <span><span className="daily-swatch daily-late" aria-hidden="true" />Terlambat</span>
            <span><span className="daily-swatch daily-failed" aria-hidden="true" />Gagal</span>
            {labels && (
              <>
                <span><span className="daily-key" data-line="solid" aria-hidden="true" />{labels.success}</span>
                <span><span className="daily-key" data-line="dashed" aria-hidden="true" />{labels.sla}</span>
              </>
            )}
          </div>
        </>
      )}
      {active && activeBar && <DailyTooltip id={tooltipID} bar={activeBar} anchor={active.el} targets={targets} />}
    </div>
  );
}
