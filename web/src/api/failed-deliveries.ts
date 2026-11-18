import { useQuery } from '@tanstack/react-query';

import { apiGet } from './client';
import type { FailedSummary, FailedSummaryDTO, FailedSummaryFilter } from './types';

export function toFailedSummary(dto: FailedSummaryDTO): FailedSummary {
  return {
    totalPackages: dto.total_packages,
    totalFailed: dto.total_failed,
    failedRate: dto.failed_rate,
    previousTotalFailed: dto.previous_total_failed,
    failedChangeRate: dto.failed_change_rate,
    reasons: dto.reasons.map((r) => ({
      reason: r.reason,
      count: r.count,
      share: r.share,
    })),
    topCouriers: dto.top_couriers.map((c) => ({
      rank: c.rank,
      courierID: c.courier_id,
      courierName: c.courier_name,
      branchID: c.branch_id,
      totalPackages: c.total_packages,
      failed: c.failed,
      failedRate: c.failed_rate,
    })),
  };
}

export async function fetchFailedSummary(
  filter: FailedSummaryFilter,
  signal?: AbortSignal,
): Promise<FailedSummary> {
  const body = await apiGet<FailedSummaryDTO, undefined>(
    '/api/v1/failed-deliveries/summary',
    {
      start_date: filter.startDate,
      end_date: filter.endDate,
      branch_id: filter.branchID,
    },
    signal,
  );
  return toFailedSummary(body.data);
}

export const failedSummaryKey = (filter: FailedSummaryFilter) =>
  ['failed-summary', filter] as const;

/** Ringkasan gagal kirim untuk satu periode dan cabang (kosong = semua cabang). */
export function useFailedSummary(filter: FailedSummaryFilter) {
  return useQuery({
    queryKey: failedSummaryKey(filter),
    queryFn: ({ signal }) => fetchFailedSummary(filter, signal),
  });
}
