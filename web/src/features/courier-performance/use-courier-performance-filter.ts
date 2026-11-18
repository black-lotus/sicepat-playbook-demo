import { useState } from 'react';

import type { CourierMetricFilter } from '../../api/types';
import { lastSevenDays } from '../../lib/date';

const PAGE_SIZE = 20;

/** State filter halaman performa kurir. Mengubah periode atau cabang kembali ke halaman 1. */
export function useCourierPerformanceFilter(today: Date = new Date()) {
  const [filter, setFilter] = useState<CourierMetricFilter>(() => ({
    ...lastSevenDays(today),
    branchID: '',
    page: 1,
    pageSize: PAGE_SIZE,
  }));

  return {
    filter,
    setPeriod: (range: { startDate: string; endDate: string }) =>
      setFilter((f) => ({ ...f, ...range, page: 1 })),
    setBranch: (branchID: string) => setFilter((f) => ({ ...f, branchID, page: 1 })),
    setPage: (page: number) => setFilter((f) => ({ ...f, page })),
  };
}
