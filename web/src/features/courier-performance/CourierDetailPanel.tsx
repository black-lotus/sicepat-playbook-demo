import { useRef } from 'react';

import type { CourierPerformance } from '../../api/types';
import { formatDate } from '../../lib/date';
import { formatGap } from '../../lib/format-gap';
import { formatNumber } from '../../lib/format-number';
import { formatPercent } from '../../lib/format-percent';
import { LaneChart } from './LaneChart';
import { useBackdropClose } from './use-backdrop-close';
import { useDialogFocus } from './use-dialog-focus';

interface CourierDetailPanelProps {
  item: CourierPerformance | null;
  onClose: () => void;
}

export function CourierDetailPanel({ item, onClose }: CourierDetailPanelProps) {
  if (!item) return null;
  return <PanelContent key={item.courierID} item={item} onClose={onClose} />;
}

function PanelContent({ item, onClose }: { item: CourierPerformance; onClose: () => void }) {
  const dialogRef = useRef<HTMLDivElement>(null);
  const closeRef = useRef<HTMLButtonElement>(null);

  useDialogFocus(dialogRef, closeRef, onClose);
  const backdropProps = useBackdropClose(onClose);

  const titleID = `detail-title-${item.courierID}`;
  const name = item.courierName || item.courierID;

  return (
    <div className="detail-backdrop" {...backdropProps}>
      <div
        ref={dialogRef}
        className="detail-panel"
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleID}
      >
        <header className="detail-header">
          <div>
            <p className="detail-pos">Posisi {item.rank}</p>
            <h2 id={titleID}>{name}</h2>
            <p className="detail-branch">{item.branchID}</p>
          </div>
          <button ref={closeRef} type="button" onClick={onClose}>Tutup</button>
        </header>

        <dl className="detail-totals">
          <div><dt>Total paket</dt><dd>{formatNumber(item.totalPackages)}</dd></div>
          <div><dt>Terkirim</dt><dd>{formatNumber(item.delivered)}</dd></div>
          <div><dt>Gagal</dt><dd>{formatNumber(item.failed)}</dd></div>
          <div><dt>Keberhasilan</dt><dd>{formatPercent(item.successRate)}</dd></div>
          <div><dt>Tepat SLA</dt><dd>{formatPercent(item.slaRate)}</dd></div>
          <div><dt>Selisih</dt><dd>{item.rank === 1 ? 'Pemimpin' : formatGap(item.successRateGap)}</dd></div>
        </dl>

        <LaneChart daily={item.daily} />
        <p className="detail-legend">
          <span className="legend-delivered">Terkirim</span>
          <span className="legend-on-time">Tepat SLA</span>
        </p>

        <div className="table-wrap">
          <table className="data-table">
            <caption className="visually-hidden">Rincian harian {name}</caption>
            <thead>
              <tr>
                <th scope="col">Tanggal</th>
                <th scope="col" className="num">Total paket</th>
                <th scope="col" className="num">Terkirim</th>
                <th scope="col" className="num">Gagal</th>
                <th scope="col" className="num">Tepat SLA</th>
              </tr>
            </thead>
            <tbody>
              {item.daily.map((day) => (
                <tr key={day.date}>
                  <td>{formatDate(day.date)}</td>
                  <td className="num">{formatNumber(day.totalPackages)}</td>
                  <td className="num">{formatNumber(day.delivered)}</td>
                  <td className="num">{formatNumber(day.failed)}</td>
                  <td className="num">{formatNumber(day.onTime)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
