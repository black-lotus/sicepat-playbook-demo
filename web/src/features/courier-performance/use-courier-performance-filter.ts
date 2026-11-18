import { useCallback, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';

import type { CourierMetricFilter } from '../../api/types';
import { lastSevenDays } from '../../lib/date';

const PAGE_SIZE = 50;
const MAX_RANGE_DAYS = 31;
const DAY_MS = 24 * 60 * 60 * 1000;

function parseISODate(value: string | null): number | null {
  if (!value || !/^\d{4}-\d{2}-\d{2}$/.test(value)) return null;
  const ms = Date.parse(`${value}T00:00:00Z`);
  if (Number.isNaN(ms)) return null;
  return new Date(ms).toISOString().slice(0, 10) === value ? ms : null;
}

function parsePage(value: string | null): number {
  if (!value || !/^\d+$/.test(value)) return 1;
  const page = Number(value);
  return page >= 1 ? page : 1;
}

/**
 * Filter halaman performa kurir dengan URL (`branch_id`, `start_date`, `end_date`, `page`) sebagai sumber kebenaran.
 * Nilai tidak valid jatuh ke default per parameter. Setter menambah entri riwayat; mengubah periode atau cabang kembali ke halaman 1.
 */
export function useCourierPerformanceFilter(today: Date = new Date()) {
  const [params, setParams] = useSearchParams();

  const filter = useMemo<CourierMetricFilter>(() => {
    const defaults = lastSevenDays(today);
    let startDate = defaults.startDate;
    let endDate = defaults.endDate;
    const rawStart = params.get('start_date');
    const rawEnd = params.get('end_date');
    const start = parseISODate(rawStart);
    const end = parseISODate(rawEnd);
    // String kosong dipertahankan: itu keadaan sementara saat input tanggal dikosongkan, UI meminta periode.
    if (start !== null || rawStart === '') startDate = rawStart as string;
    if (end !== null || rawEnd === '') endDate = rawEnd as string;
    const s = Date.parse(`${startDate}T00:00:00Z`);
    const e = Date.parse(`${endDate}T00:00:00Z`);
    if (Number.isNaN(s) || Number.isNaN(e)) {
      // salah satu kosong: biarkan
    } else if (e < s || (e - s) / DAY_MS + 1 > MAX_RANGE_DAYS) {
      startDate = defaults.startDate;
      endDate = defaults.endDate;
    }
    return {
      startDate,
      endDate,
      branchID: params.get('branch_id') ?? '',
      page: parsePage(params.get('page')),
      pageSize: PAGE_SIZE,
    };
  }, [params, today]);

  const write = useCallback(
    (changes: Record<string, string | null>) => {
      const next = new URLSearchParams(params);
      for (const [key, value] of Object.entries(changes)) {
        if (value === null) next.delete(key);
        else next.set(key, value);
      }
      setParams(next);
    },
    [params, setParams],
  );

  return {
    filter,
    setPeriod: (range: { startDate: string; endDate: string }) =>
      write({ start_date: range.startDate, end_date: range.endDate, page: null }),
    setBranch: (branchID: string) => write({ branch_id: branchID || null, page: null }),
    setPage: (page: number) => write({ page: page > 1 ? String(page) : null }),
  };
}
