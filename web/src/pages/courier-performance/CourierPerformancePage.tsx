import { useState } from 'react';

import { useCourierPerformance } from '../../api/courier-performance';
import { DateRangeFilter } from '../../components/DateRangeFilter';
import { PinIcon } from '../../components/icons';
import { Pagination } from '../../components/Pagination';
import { QueryState } from '../../components/QueryState';
import { formatDateRange } from '../../lib/date';
import { useMediaQuery } from '../../lib/use-media-query';
import '../../features/courier-performance/courier-performance.css';
import '../../features/courier-performance/podium.css';
import '../../features/courier-performance/results.css';
import '../../features/courier-performance/detail-panel.css';
import { CourierDetailPanel } from '../../features/courier-performance/CourierDetailPanel';
import { FilterSheet } from '../../features/courier-performance/FilterSheet';
import { Podium } from '../../features/courier-performance/Podium';
import { ResultsList } from '../../features/courier-performance/ResultsList';
import { ResultsTable } from '../../features/courier-performance/ResultsTable';
import { useCourierPerformanceFilter } from '../../features/courier-performance/use-courier-performance-filter';

const NARROW_QUERY = '(max-width: 640px)';

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

  const isNarrow = useMediaQuery(NARROW_QUERY);

  const selected = query.data?.items.find((item) => item.courierID === selectedCourierID) ?? null;

  const branchInput = (
    <input
      type="text"
      placeholder="Semua cabang"
      value={filter.branchID}
      onChange={(e) => setBranch(e.target.value.trim().toUpperCase())}
    />
  );

  // Desktop: kontrol bergaya pil dengan label tersembunyi. Layar sempit: kontrol standar di dalam FilterSheet.
  const pillControls = (
    <>
      <DateRangeFilter variant="pill" startDate={filter.startDate} endDate={filter.endDate} onChange={setPeriod} />
      <label className="pill-field">
        <span className="visually-hidden">Cabang</span>
        <PinIcon />
        {branchInput}
      </label>
    </>
  );
  const sheetControls = (
    <>
      <DateRangeFilter startDate={filter.startDate} endDate={filter.endDate} onChange={setPeriod} />
      <label>
        Cabang
        {branchInput}
      </label>
    </>
  );

  return (
    <section className="courier-performance-page">
      <header className="page-header">
        <h1>Performa Kurir</h1>
        {!isNarrow && <div className="filters filters-pill">{pillControls}</div>}
      </header>
      {isNarrow && (
        <FilterSheet summary={`${filter.branchID || 'Semua cabang'} · ${formatDateRange(filter.startDate, filter.endDate)}`}>
          {sheetControls}
        </FilterSheet>
      )}

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
            <div className="card results-card">
              <h2 className="card-title">Hasil lengkap</h2>
              {isNarrow ? (
                <ResultsList items={query.data.items} onSelect={setSelectedCourierID} />
              ) : (
                <ResultsTable items={query.data.items} onSelect={setSelectedCourierID} />
              )}
              <Pagination
                page={query.data.meta.page}
                pageSize={query.data.meta.pageSize}
                total={query.data.meta.total}
                onPageChange={setPage}
              />
            </div>
          </div>
        )}
      </QueryState>
      <CourierDetailPanel item={selected} onClose={() => setSelectedCourierID(null)} />
    </section>
  );
}
