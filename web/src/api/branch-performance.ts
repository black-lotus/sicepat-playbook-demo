import { keepPreviousData, useQuery } from '@tanstack/react-query';

import { apiGet } from './client';
import { toDailyCount } from './courier-performance';
import type {
  BranchPerformance,
  BranchPerformanceDTO,
  BranchPerformanceFilter,
  PageMetaDTO,
  Paginated,
} from './types';

export function toBranchPerformance(dto: BranchPerformanceDTO): BranchPerformance {
  return {
    rank: dto.rank,
    branchID: dto.branch_id,
    courierCount: dto.courier_count,
    totalPackages: dto.total_packages,
    delivered: dto.delivered,
    failed: dto.failed,
    onTime: dto.on_time,
    successRate: dto.success_rate,
    slaRate: dto.sla_rate,
    target: dto.target
      ? {
          successRate: dto.target.success_rate,
          slaRate: dto.target.sla_rate,
          deliveredPackages: dto.target.delivered_packages,
          successRateStatus: dto.target.success_rate_status,
          slaRateStatus: dto.target.sla_rate_status,
          deliveredStatus: dto.target.delivered_status,
        }
      : null,
    daily: dto.daily.map(toDailyCount),
  };
}

export async function fetchBranchPerformance(
  filter: BranchPerformanceFilter,
  signal?: AbortSignal,
): Promise<Paginated<BranchPerformance>> {
  const body = await apiGet<BranchPerformanceDTO[], PageMetaDTO>(
    '/api/v1/branch-performance',
    {
      start_date: filter.startDate,
      end_date: filter.endDate,
      page: filter.page,
      page_size: filter.pageSize,
    },
    signal,
  );
  return {
    items: body.data.map(toBranchPerformance),
    meta: {
      page: body.meta.page,
      pageSize: body.meta.page_size,
      total: body.meta.total,
    },
  };
}

export const branchPerformanceKey = (filter: BranchPerformanceFilter) =>
  ['branch-performance', filter] as const;

/** Ranking cabang untuk satu periode. Data halaman sebelumnya tetap tampil saat pindah halaman. */
export function useBranchPerformance(filter: BranchPerformanceFilter) {
  return useQuery({
    queryKey: branchPerformanceKey(filter),
    queryFn: ({ signal }) => fetchBranchPerformance(filter, signal),
    placeholderData: keepPreviousData,
  });
}
