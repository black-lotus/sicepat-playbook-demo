import { useCourierMetrics } from '../../api/courier-metrics';
import { DateRangeFilter } from '../../components/DateRangeFilter';
import { Pagination } from '../../components/Pagination';
import { QueryState } from '../../components/QueryState';
import { CourierPerformanceTable } from '../../features/courier-performance/CourierPerformanceTable';
import { useCourierPerformanceFilter } from '../../features/courier-performance/use-courier-performance-filter';

interface CourierPerformancePageProps {
  today?: Date;
}

export function CourierPerformancePage({ today }: CourierPerformancePageProps) {
  const { filter, setPeriod, setBranch, setPage } = useCourierPerformanceFilter(today);
  const query = useCourierMetrics(filter);

  return (
    <section>
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
      >
        {query.data && (
          <>
            <CourierPerformanceTable metrics={query.data.items} />
            <Pagination
              page={query.data.meta.page}
              pageSize={query.data.meta.pageSize}
              total={query.data.meta.total}
              onPageChange={setPage}
            />
          </>
        )}
      </QueryState>
    </section>
  );
}
