import { useRef } from 'react';

import type { CourierPerformance } from '../../api/types';
import type { TargetRates } from '../../lib/target-labels';
import { formatNumber } from '../../lib/format-number';
import { formatPercent } from '../../lib/format-percent';
import { Avatar } from './Avatar';
import { DailyStrip } from '../../components/DailyStrip';
import { StatusBadge } from '../../components/StatusBadge';
import { useFinishLine } from './use-finish-line';

interface ResultsListProps {
  items: CourierPerformance[];
  onSelect: (courierID: string) => void;
  /** Target cabang; bila ada, strip menggambar garis target. */
  targets?: TargetRates;
}

function RateStatus({ status }: { status: CourierPerformance['successRateStatus'] }) {
  if (!status) return null;
  return <StatusBadge status={status} label={status === 'met' ? 'Di atas target' : 'Di bawah target'} />;
}

const medal: Record<number, 'gold' | 'silver' | 'bronze'> = { 1: 'gold', 2: 'silver', 3: 'bronze' };

/** Daftar peringkat untuk layar sempit, pengganti `ResultsTable`. */
export function ResultsList({ items, onSelect, targets }: ResultsListProps) {
  const listRef = useRef<HTMLOListElement>(null);
  useFinishLine(listRef, items.map((item) => item.courierID));

  return (
    <ol ref={listRef} className="results-list" aria-label="Hasil lengkap">
      {items.map((item) => {
        const name = item.courierName || item.courierID;
        return (
          <li key={item.courierID} className="results-item" data-finish-key={item.courierID}>
            <Avatar courierID={item.courierID} name={name} medal={medal[item.rank]} />
            <div className="results-body">
              <button type="button" className="link-button results-name" onClick={() => onSelect(item.courierID)}>
                {name}
              </button>
              <span className="courier-sub">Pos {item.rank} · {formatNumber(item.totalPackages)} paket</span>
              <DailyStrip daily={item.daily} size="sm" targets={targets} />
              <RateStatus status={item.successRateStatus} />
            </div>
            <span className="results-rate">
              <span className="results-rate-figure">{formatPercent(item.successRate)}</span>
              <span className="results-rate-label">Keberhasilan</span>
            </span>
          </li>
        );
      })}
    </ol>
  );
}
