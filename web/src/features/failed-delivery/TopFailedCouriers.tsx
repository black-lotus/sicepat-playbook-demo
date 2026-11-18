import type { FailedCourier } from '../../api/types';
import { Avatar } from '../courier-performance/Avatar';
import { formatNumber } from '../../lib/format-number';
import { formatPercent } from '../../lib/format-percent';

/** Daftar kurir dengan gagal terbanyak (maksimal 5 dari API). */
export function TopFailedCouriers({ couriers }: { couriers: FailedCourier[] }) {
  return (
    <section className="card" aria-labelledby="failed-couriers-title">
      <h2 id="failed-couriers-title" className="card-title">Kurir dengan gagal terbanyak</h2>
      <ol className="failed-couriers">
        {couriers.map((courier) => {
          const name = courier.courierName || courier.courierID;
          return (
            <li key={courier.courierID}>
              <span className="failed-courier-rank">{courier.rank}</span>
              <Avatar courierID={courier.courierID} name={name} />
              <span className="failed-courier-who">
                <strong>{name}</strong>
                <span>{courier.branchID}</span>
              </span>
              <span className="failed-courier-count">
                {formatNumber(courier.failed)} <span>dari {formatNumber(courier.totalPackages)} paket</span>
              </span>
              <span className="failed-courier-rate">{formatPercent(courier.failedRate)}</span>
            </li>
          );
        })}
      </ol>
    </section>
  );
}
