import { useState } from 'react';

import { useBranchPerformance } from '../../api/branch-performance';
import { DateRangeFilter } from '../../components/DateRangeFilter';
import { Pagination } from '../../components/Pagination';
import { QueryState } from '../../components/QueryState';
import '../../features/branch-summary/branch-summary.css';
import { BranchCard } from '../../features/branch-summary/BranchCard';
import '../../features/courier-performance/courier-performance.css';
import { lastSevenDays } from '../../lib/date';

const PAGE_SIZE = 10;

interface BranchSummaryPageProps {
  today?: Date;
}

export function BranchSummaryPage({ today }: BranchSummaryPageProps) {
  const [range, setRange] = useState(() => lastSevenDays(today ?? new Date()));
  const [page, setPage] = useState(1);
  const query = useBranchPerformance({ ...range, page, pageSize: PAGE_SIZE });

  return (
    <section className="branch-summary-page">
      <header className="page-header">
        <h1>Ringkasan Cabang</h1>
        <div className="filters filters-pill">
          <DateRangeFilter
            variant="pill"
            startDate={range.startDate}
            endDate={range.endDate}
            onChange={(next) => {
              setRange(next);
              setPage(1);
            }}
          />
        </div>
      </header>

      <QueryState
        isPending={query.isPending}
        error={query.error}
        isEmpty={query.data?.meta.total === 0}
        emptyMessage="Tidak ada data cabang untuk periode ini."
        onRetry={() => void query.refetch()}
      >
        {query.data && (
          <>
            <ol className="branch-list" aria-label="Ranking cabang">
              {query.data.items.map((branch) => (
                <li key={branch.branchID}>
                  <BranchCard branch={branch} startDate={range.startDate} endDate={range.endDate} />
                </li>
              ))}
            </ol>
            <Pagination page={page} pageSize={PAGE_SIZE} total={query.data.meta.total} onPageChange={setPage} />
          </>
        )}
      </QueryState>
    </section>
  );
}
