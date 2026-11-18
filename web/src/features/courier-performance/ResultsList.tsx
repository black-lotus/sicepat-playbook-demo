import { useRef } from 'react';

import type { CourierPerformance } from '../../api/types';
import { formatPercent } from '../../lib/format-percent';
import { DailyStrip } from './DailyStrip';
import { useFinishLine } from './use-finish-line';

interface ResultsListProps {
  items: CourierPerformance[];
  onSelect: (courierID: string) => void;
}

/** Daftar peringkat untuk layar sempit, pengganti `ResultsTable`. */
export function ResultsList({ items, onSelect }: ResultsListProps) {
  const listRef = useRef<HTMLOListElement>(null);
  useFinishLine(listRef, items.map((item) => item.courierID));

  return (
    <ol ref={listRef} className="results-list" aria-label="Hasil lengkap">
      {items.map((item) => (
        <li key={item.courierID} className="results-item" data-finish-key={item.courierID}>
          <span className="results-pos figure">{item.rank}</span>
          <div className="results-body">
            <button type="button" className="link-button results-name" onClick={() => onSelect(item.courierID)}>
              {item.courierName || item.courierID}
            </button>
            <DailyStrip daily={item.daily} size="sm" />
          </div>
          <span className="results-rate">
            <span className="results-rate-figure">{formatPercent(item.successRate)}</span>
            <span className="results-rate-label">Keberhasilan</span>
          </span>
        </li>
      ))}
    </ol>
  );
}
