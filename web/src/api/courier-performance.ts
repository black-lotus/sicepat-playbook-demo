import { keepPreviousData, useQuery } from '@tanstack/react-query';

import { apiGet } from './client';
import type {
  CourierMetricFilter,
  CourierPerformance,
  CourierPerformanceDTO,
  DailyCount,
  DailyCountDTO,
  PageMetaDTO,
  Paginated,
} from './types';

function toDailyCount(dto: DailyCountDTO): DailyCount {
  return {
    date: dto.date,
    totalPackages: dto.total_packages,
    delivered: dto.delivered,
    failed: dto.failed,
    onTime: dto.on_time,
  };
}

export function toCourierPerformance(dto: CourierPerformanceDTO): CourierPerformance {
  return {
    rank: dto.rank,
    courierID: dto.courier_id,
    courierName: dto.courier_name,
    branchID: dto.branch_id,
    totalPackages: dto.total_packages,
    delivered: dto.delivered,
    failed: dto.failed,
    onTime: dto.on_time,
    successRate: dto.success_rate,
    slaRate: dto.sla_rate,
    successRateGap: dto.success_rate_gap,
    daily: dto.daily.map(toDailyCount),
  };
}

export async function fetchCourierPerformance(
  filter: CourierMetricFilter,
  signal?: AbortSignal,
): Promise<Paginated<CourierPerformance>> {
  const body = await apiGet<CourierPerformanceDTO[], PageMetaDTO>(
    '/api/v1/courier-performance',
    {
      start_date: filter.startDate,
      end_date: filter.endDate,
      branch_id: filter.branchID,
      page: filter.page,
      page_size: filter.pageSize,
    },
    signal,
  );
  return {
    items: body.data.map(toCourierPerformance),
    meta: { page: body.meta.page, pageSize: body.meta.page_size, total: body.meta.total },
  };
}

export const courierPerformanceKey = (filter: CourierMetricFilter) =>
  ['courier-performance', filter] as const;

/** Ranking kurir untuk satu periode. Data halaman sebelumnya tetap tampil saat pindah halaman. */
export function useCourierPerformance(filter: CourierMetricFilter) {
  return useQuery({
    queryKey: courierPerformanceKey(filter),
    queryFn: ({ signal }) => fetchCourierPerformance(filter, signal),
    placeholderData: keepPreviousData,
  });
}
