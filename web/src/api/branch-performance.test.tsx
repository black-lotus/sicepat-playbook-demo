import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { branchPerformanceDTOs } from '../test/fixtures';
import { server } from '../test/server';
import { ApiError } from './client';
import { branchPerformanceKey, fetchBranchPerformance, useBranchPerformance } from './branch-performance';

const filter = {
  startDate: '2026-11-01',
  endDate: '2026-11-07',
  page: 1,
  pageSize: 20,
};

describe('fetchBranchPerformance', () => {
  it('mengirim parameter query dengan benar', async () => {
    let params: URLSearchParams | undefined;
    server.use(
      http.get('/api/v1/branch-performance', ({ request }) => {
        params = new URL(request.url).searchParams;
        return HttpResponse.json({
          data: [],
          meta: { page: 1, page_size: 20, total: 0 },
        });
      }),
    );

    await fetchBranchPerformance(filter);

    expect(Object.fromEntries(params ?? [])).toEqual({
      start_date: '2026-11-01',
      end_date: '2026-11-07',
      page: '1',
      page_size: '20',
    });
  });

  it('memetakan cabang dengan target dan tanpa target (null)', async () => {
    server.use(
      http.get('/api/v1/branch-performance', () =>
        HttpResponse.json({
          data: branchPerformanceDTOs,
          meta: { page: 1, page_size: 20, total: 2 },
        }),
      ),
    );

    const result = await fetchBranchPerformance(filter);

    expect(result.meta).toEqual({ page: 1, pageSize: 20, total: 2 });
    expect(result.items[0]).toEqual({
      rank: 1,
      branchID: 'BDG02',
      courierCount: 8,
      totalPackages: 1000,
      delivered: 833,
      failed: 167,
      onTime: 700,
      successRate: 0.833,
      slaRate: 0.8403,
      target: {
        successRate: 0.85,
        slaRate: 0.8,
        deliveredPackages: 900,
        successRateStatus: 'below',
        slaRateStatus: 'met',
        deliveredStatus: 'below',
      },
      daily: [
        {
          date: '2026-11-01',
          totalPackages: 140,
          delivered: 120,
          failed: 20,
          onTime: 101,
        },
      ],
    });
    expect(result.items[1]?.target).toBeNull();
  });

  it('melempar ApiError dengan code dan message dari backend', async () => {
    server.use(
      http.get('/api/v1/branch-performance', () =>
        HttpResponse.json(
          { error: { code: 'invalid_filter', message: 'periode tidak valid' } },
          { status: 400 },
        ),
      ),
    );

    const error = await fetchBranchPerformance(filter).catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({
      status: 400,
      code: 'invalid_filter',
      message: 'periode tidak valid',
    });
  });
});

describe('branchPerformanceKey', () => {
  it('memuat filter agar cache terpisah per filter', () => {
    expect(branchPerformanceKey(filter)).toEqual(['branch-performance', filter]);
  });
});

describe('useBranchPerformance', () => {
  it('mengambil data lewat hook', async () => {
    server.use(
      http.get('/api/v1/branch-performance', () =>
        HttpResponse.json({
          data: branchPerformanceDTOs,
          meta: { page: 1, page_size: 20, total: 2 },
        }),
      ),
    );
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const { result } = renderHook(() => useBranchPerformance(filter), {
      wrapper: ({ children }) => (
        <QueryClientProvider client={client}>{children}</QueryClientProvider>
      ),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.items).toHaveLength(2);
  });
});
