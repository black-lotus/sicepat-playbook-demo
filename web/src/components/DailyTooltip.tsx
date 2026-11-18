import { useLayoutEffect, useRef } from 'react';
import { createPortal } from 'react-dom';

import type { DailyBar } from '../lib/daily-bars';
import { formatWeekdayDate } from '../lib/date';
import { formatNumber } from '../lib/format-number';
import { formatPercent } from '../lib/format-percent';
import { targetLabels, type TargetRates } from '../lib/target-labels';

interface DailyTooltipProps {
  id: string;
  bar: DailyBar;
  /** Batang yang ditunjuk; tooltip diposisikan relatif terhadapnya. */
  anchor: HTMLElement;
  targets?: TargetRates;
}

const EDGE = 8;
const GAP = 10;

/**
 * Keterangan satu hari. Dirender di body dengan posisi fixed supaya tidak terpotong `overflow`
 * kartu atau tabel, lalu dijepit ke dalam layar sehingga batang paling kiri dan kanan tetap terbaca.
 * Posisi ditulis langsung ke elemen (bukan state) karena bergantung pada ukuran hasil render.
 */
export function DailyTooltip({ id, bar, anchor, targets }: DailyTooltipProps) {
  const ref = useRef<HTMLDivElement>(null);

  useLayoutEffect(() => {
    const tip = ref.current;
    if (!tip) return;
    const a = anchor.getBoundingClientRect();
    const t = tip.getBoundingClientRect();
    const center = a.left + a.width / 2;
    const maxLeft = Math.max(EDGE, window.innerWidth - t.width - EDGE);
    const left = Math.min(Math.max(center - t.width / 2, EDGE), maxLeft);
    const above = a.top - t.height - GAP >= EDGE;
    tip.style.left = `${left}px`;
    tip.style.top = `${above ? a.top - t.height - GAP : a.bottom + GAP}px`;
    tip.style.setProperty('--arrow-x', `${Math.min(Math.max(center - left, 14), Math.max(14, t.width - 14))}px`);
    tip.dataset.side = above ? 'top' : 'bottom';
    tip.style.visibility = 'visible';
  }, [anchor, bar, targets]);

  const labels = targets ? targetLabels(targets) : null;

  return createPortal(
    <div ref={ref} id={id} role="tooltip" className="daily-tooltip">
      <strong className="daily-tooltip-date">{formatWeekdayDate(bar.date)}</strong>
      <dl className="daily-tooltip-rows">
        <div>
          <dt>Total paket</dt>
          <dd>{formatNumber(bar.total)}</dd>
        </div>
        <div>
          <dt><span className="daily-swatch daily-on-time" aria-hidden="true" />Tepat SLA</dt>
          <dd>{formatNumber(bar.onTime)}</dd>
        </div>
        <div>
          <dt><span className="daily-swatch daily-late" aria-hidden="true" />Terlambat</dt>
          <dd>{formatNumber(bar.late)}</dd>
        </div>
        <div>
          <dt><span className="daily-swatch daily-failed" aria-hidden="true" />Gagal</dt>
          <dd>{formatNumber(bar.failed)}</dd>
        </div>
        <div className="daily-tooltip-total">
          <dt>Keberhasilan</dt>
          <dd>{bar.success === null ? '-' : formatPercent(bar.success)}</dd>
        </div>
      </dl>
      {labels && (
        <p className="daily-tooltip-targets">
          {labels.success}. {labels.sla}.
        </p>
      )}
    </div>,
    document.body,
  );
}
