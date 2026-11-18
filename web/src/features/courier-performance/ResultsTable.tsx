import { useRef } from 'react';

import type { CourierPerformance } from '../../api/types';
import { formatGap } from '../../lib/format-gap';
import { formatNumber } from '../../lib/format-number';
import { formatPercent } from '../../lib/format-percent';
import { DailyStrip } from './DailyStrip';
import { useFinishLine } from './use-finish-line';

interface ResultsTableProps {
  items: CourierPerformance[];
  onSelect: (courierID: string) => void;
}

export function ResultsTable({ items, onSelect }: ResultsTableProps) {
  const bodyRef = useRef<HTMLTableSectionElement>(null);
  useFinishLine(bodyRef, items.map((item) => item.courierID));

  return (
    <div className="table-wrap">
      <table className="data-table results-table">
        <caption className="visually-hidden">Hasil lengkap</caption>
        <thead>
          <tr>
            <th scope="col" className="num">Pos</th>
            <th scope="col">Kurir</th>
            <th scope="col">Harian</th>
            <th scope="col" className="num">Total paket</th>
            <th scope="col" className="num">Terkirim</th>
            <th scope="col" className="num">Gagal</th>
            <th scope="col" className="num">Keberhasilan</th>
            <th scope="col" className="num">Tepat SLA</th>
            <th scope="col" className="num">Selisih</th>
          </tr>
        </thead>
        <tbody ref={bodyRef}>
          {items.map((item) => (
            <tr key={item.courierID} data-finish-key={item.courierID}>
              <td className="num figure">{item.rank}</td>
              <td>
                <button type="button" className="link-button" onClick={() => onSelect(item.courierID)}>
                  {item.courierName || item.courierID}
                </button>
              </td>
              <td className="daily-cell"><DailyStrip daily={item.daily} size="sm" /></td>
              <td className="num">{formatNumber(item.totalPackages)}</td>
              <td className="num">{formatNumber(item.delivered)}</td>
              <td className="num">{formatNumber(item.failed)}</td>
              <td className="num">{formatPercent(item.successRate)}</td>
              <td className="num">{formatPercent(item.slaRate)}</td>
              <td className="num gap-cell">{item.rank === 1 ? 'Pemimpin' : formatGap(item.successRateGap)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
