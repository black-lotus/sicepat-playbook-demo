import { keepPreviousData, useQuery } from '@tanstack/react-query';

import { apiGet } from './client';
import type {
  CourierMetric,
  CourierMetricDTO,
  CourierMetricFilter,
  PageMetaDTO,
  Paginated,
} from './types';

export function toCourierMetric(dto: CourierMetricDTO): CourierMetric {
  return {
    courierID: dto.courier_id,
    courierName: dto.courier_name,
    branchID: dto.branch_id,
    date: dto.date,
    totalPackages: dto.total_packages,
    delivered: dto.delivered,
    failed: dto.failed,
    onTime: dto.on_time,
    successRate: dto.success_rate,
    slaRate: dto.sla_rate,
  };
}

export async function fetchCourierMetrics(
  filter: CourierMetricFilter,
  signal?: AbortSignal,
): Promise<Paginated<CourierMetric>> {
  const body = await apiGet<CourierMetricDTO[], PageMetaDTO>(
    '/api/v1/courier-metrics',
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
    items: body.data.map(toCourierMetric),
    meta: { page: body.meta.page, pageSize: body.meta.page_size, total: body.meta.total },
  };
}

export const courierMetricsKey = (filter: CourierMetricFilter) =>
  ['courier-metrics', filter] as const;

/** KPI kurir per hari. Data halaman sebelumnya tetap tampil saat pindah halaman. */
export function useCourierMetrics(filter: CourierMetricFilter) {
  return useQuery({
    queryKey: courierMetricsKey(filter),
    queryFn: ({ signal }) => fetchCourierMetrics(filter, signal),
    placeholderData: keepPreviousData,
  });
}
