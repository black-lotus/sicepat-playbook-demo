import { useState } from 'react';

import { useCourierPerformance } from '../../api/courier-performance';
import { DateRangeFilter } from '../../components/DateRangeFilter';
import { Pagination } from '../../components/Pagination';
import { QueryState } from '../../components/QueryState';
import '../../features/courier-performance/courier-performance.css';
import { CourierDetailPanel } from '../../features/courier-performance/CourierDetailPanel';
import { Podium } from '../../features/courier-performance/Podium';
import { ResultsTable } from '../../features/courier-performance/ResultsTable';
import { useCourierPerformanceFilter } from '../../features/courier-performance/use-courier-performance-filter';

interface CourierPerformancePageProps {
  today?: Date;
}

function LoadingSkeleton() {
  return (
    <div className="skeleton" aria-hidden="true">
      <div className="skeleton-podium">
        <div className="skeleton-block" />
        <div className="skeleton-block" />
        <div className="skeleton-block" />
      </div>
      <div className="skeleton-row" />
      <div className="skeleton-row" />
      <div className="skeleton-row" />
    </div>
  );
}

export function CourierPerformancePage({ today }: CourierPerformancePageProps) {
  const { filter, setPeriod, setBranch, setPage } = useCourierPerformanceFilter(today);
  const query = useCourierPerformance(filter);
  const [selectedCourierID, setSelectedCourierID] = useState<string | null>(null);

  const selected = query.data?.items.find((item) => item.courierID === selectedCourierID) ?? null;

  return (
    <section className="courier-performance-page">
      <h1>Performa Kurir</h1>
      <div className="filters">
        <DateRangeFilter startDate={filter.startDate} endDate={filter.endDate} onChange={setPeriod} />
        <label>
          Cabang
          <input
            type="text"
            placeholder="Semua cabang"
            value={filter.branchID}
            onChange={(e) => setBranch(e.target.value.trim().toUpperCase())}
          />
        </label>
      </div>

      <QueryState
        isPending={query.isPending}
        error={query.error}
        isEmpty={query.data?.items.length === 0}
        emptyMessage="Belum ada data performa kurir untuk periode ini."
        onRetry={() => void query.refetch()}
        loadingFallback={<LoadingSkeleton />}
      >
        {query.data && (
          <div className={query.isPlaceholderData ? 'results results-stale' : 'results'} aria-busy={query.isPlaceholderData ? true : undefined}>
            {query.data.meta.page === 1 && <Podium items={query.data.items} onSelect={setSelectedCourierID} />}
            <ResultsTable items={query.data.items} onSelect={setSelectedCourierID} />
            <Pagination
              page={query.data.meta.page}
              pageSize={query.data.meta.pageSize}
              total={query.data.meta.total}
              onPageChange={setPage}
            />
          </div>
        )}
      </QueryState>
      <CourierDetailPanel item={selected} onClose={() => setSelectedCourierID(null)} />
    </section>
  );
}
