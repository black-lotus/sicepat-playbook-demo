import { useState } from 'react';

import { useFailedSummary } from '../../api/failed-deliveries';
import { DateRangeFilter } from '../../components/DateRangeFilter';
import { PinIcon } from '../../components/icons';
import { QueryState } from '../../components/QueryState';
import '../../features/courier-performance/courier-performance.css';
import '../../features/failed-delivery/failed-delivery.css';
import { FailedHero } from '../../features/failed-delivery/FailedHero';
import { FailedReasonLegend } from '../../features/failed-delivery/FailedReasonLegend';
import { TopFailedCouriers } from '../../features/failed-delivery/TopFailedCouriers';
import { lastSevenDays } from '../../lib/date';

interface FailedDeliveryPageProps {
  today?: Date;
}

export function FailedDeliveryPage({ today }: FailedDeliveryPageProps) {
  const [filter, setFilter] = useState(() => ({ ...lastSevenDays(today ?? new Date()), branchID: '' }));
  const query = useFailedSummary(filter);

  return (
    <section className="failed-delivery-page">
      <header className="page-header">
        <h1>Gagal Kirim</h1>
        <div className="filters filters-pill">
          <DateRangeFilter
            variant="pill"
            startDate={filter.startDate}
            endDate={filter.endDate}
            onChange={(range) => setFilter((f) => ({ ...f, ...range }))}
          />
          <label className="pill-field">
            <span className="visually-hidden">Cabang</span>
            <PinIcon />
            <input
              type="text"
              placeholder="Semua cabang"
              value={filter.branchID}
              onChange={(e) => {
                const branchID = e.target.value.trim().toUpperCase();
                setFilter((f) => ({ ...f, branchID }));
              }}
            />
          </label>
        </div>
      </header>

      <QueryState
        isPending={query.isPending}
        error={query.error}
        isEmpty={query.data?.totalFailed === 0}
        emptyMessage="Tidak ada paket gagal kirim untuk periode ini."
        onRetry={() => void query.refetch()}
      >
        {query.data && (
          <div className="failed-content">
            <FailedHero
              totalFailed={query.data.totalFailed}
              failedRate={query.data.failedRate}
              previousTotalFailed={query.data.previousTotalFailed}
              failedChangeRate={query.data.failedChangeRate}
            />
            <div className="failed-grid">
              <FailedReasonLegend totalFailed={query.data.totalFailed} reasons={query.data.reasons} />
              <TopFailedCouriers couriers={query.data.topCouriers} />
            </div>
          </div>
        )}
      </QueryState>
    </section>
  );
}
