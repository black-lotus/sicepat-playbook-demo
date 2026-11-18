import { useRef } from 'react';

import type { CourierPerformance } from '../../api/types';
import { formatGap } from '../../lib/format-gap';
import type { TargetRates } from '../../lib/target-labels';
import { formatNumber } from '../../lib/format-number';
import { formatPercent } from '../../lib/format-percent';
import { Avatar } from './Avatar';
import { DailyStrip } from '../../components/DailyStrip';
import { StatusBadge } from '../../components/StatusBadge';
import { useFinishLine } from './use-finish-line';

interface ResultsTableProps {
  items: CourierPerformance[];
  onSelect: (courierID: string) => void;
  /** Target cabang; bila ada, strip menggambar garis target. */
  targets?: TargetRates;
}

const medal: Record<number, 'gold' | 'silver' | 'bronze'> = { 1: 'gold', 2: 'silver', 3: 'bronze' };

/**
 * Tabel semantik yang ditata sebagai baris kartu (gaya riwayat transaksi). Role ditulis eksplisit karena
 * `display` bukan table membuat sebagian browser membuang semantik tabel.
 */
export function ResultsTable({ items, onSelect, targets }: ResultsTableProps) {
  const bodyRef = useRef<HTMLTableSectionElement>(null);
  useFinishLine(bodyRef, items.map((item) => item.courierID));

  return (
    <div className="table-wrap">
      <table className="data-table results-table" role="table">
        <caption className="visually-hidden">Hasil lengkap</caption>
        <thead className="visually-hidden" role="rowgroup">
          <tr role="row">
            <th scope="col" role="columnheader" className="num">Pos</th>
            <th scope="col" role="columnheader">Kurir</th>
            <th scope="col" role="columnheader">Harian</th>
            <th scope="col" role="columnheader" className="num">Total paket</th>
            <th scope="col" role="columnheader" className="num">Terkirim</th>
            <th scope="col" role="columnheader" className="num">Gagal</th>
            <th scope="col" role="columnheader" className="num">Keberhasilan</th>
            <th scope="col" role="columnheader" className="num">Tepat SLA</th>
            <th scope="col" role="columnheader" className="num">Selisih</th>
          </tr>
        </thead>
        <tbody ref={bodyRef} role="rowgroup">
          {items.map((item) => {
            const name = item.courierName || item.courierID;
            return (
              <tr key={item.courierID} role="row" data-finish-key={item.courierID}>
                <td role="cell" className="num cell-hidden">{item.rank}</td>
                <td role="cell" className="cell-courier">
                  <Avatar courierID={item.courierID} name={name} medal={medal[item.rank]} />
                  <span className="courier-text">
                    <button type="button" className="link-button" onClick={() => onSelect(item.courierID)}>
                      {name}
                    </button>
                    <span className="courier-sub" aria-hidden="true">
                      Pos {item.rank} · {formatNumber(item.totalPackages)} paket
                    </span>
                  </span>
                </td>
                <td role="cell" className="daily-cell"><DailyStrip daily={item.daily} size="sm" targets={targets} /></td>
                <td role="cell" className="num cell-hidden">{formatNumber(item.totalPackages)}</td>
                <td role="cell" className="num cell-stat cell-delivered">
                  <span className="stat-label" aria-hidden="true">Terkirim</span>
                  {formatNumber(item.delivered)}
                </td>
                <td role="cell" className="num cell-stat cell-failed">
                  <span className="stat-label" aria-hidden="true">Gagal</span>
                  {formatNumber(item.failed)}
                </td>
                <td role="cell" className="num cell-rate">{formatPercent(item.successRate)}
                  {item.successRateStatus && (
                    <StatusBadge status={item.successRateStatus} label={item.successRateStatus === 'met' ? 'Di atas target' : 'Di bawah target'} />
                  )}
                </td>
                <td role="cell" className="num cell-stat cell-sla">
                  <span className="stat-label" aria-hidden="true">Tepat SLA</span>
                  {formatPercent(item.slaRate)}
                </td>
                <td role="cell" className="num gap-cell">{item.rank === 1 ? 'Pemimpin' : formatGap(item.successRateGap)}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
