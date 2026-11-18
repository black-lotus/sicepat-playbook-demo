import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';

import { failedSummaryDTO } from '../test/fixtures';
import { server } from '../test/server';
import { ApiError } from './client';
import { failedSummaryKey, fetchFailedSummary, useFailedSummary } from './failed-deliveries';

const filter = {
  startDate: '2026-11-01',
  endDate: '2026-11-07',
  branchID: 'JKT01',
};

describe('fetchFailedSummary', () => {
  it('mengirim parameter query dan menghilangkan branch_id kosong', async () => {
    const seen: URLSearchParams[] = [];
    server.use(
      http.get('/api/v1/failed-deliveries/summary', ({ request }) => {
        seen.push(new URL(request.url).searchParams);
        return HttpResponse.json({ data: failedSummaryDTO });
      }),
    );

    await fetchFailedSummary(filter);
    await fetchFailedSummary({ ...filter, branchID: '' });

    expect(Object.fromEntries(seen[0] ?? [])).toEqual({
      start_date: '2026-11-01',
      end_date: '2026-11-07',
      branch_id: 'JKT01',
    });
    expect(seen[1]?.has('branch_id')).toBe(false);
  });

  it('memetakan respons tanpa meta ke camelCase', async () => {
    server.use(
      http.get('/api/v1/failed-deliveries/summary', () =>
        HttpResponse.json({ data: failedSummaryDTO }),
      ),
    );

    const result = await fetchFailedSummary(filter);

    expect(result).toEqual({
      totalPackages: 1200,
      totalFailed: 180,
      failedRate: 0.15,
      previousTotalFailed: 200,
      failedChangeRate: -0.1,
      reasons: [
        { reason: 'Penerima tidak di tempat', count: 60, share: 0.3333333333 },
        { reason: 'Tanpa alasan', count: 20, share: 0.1111111111 },
      ],
      topCouriers: [
        {
          rank: 1,
          courierID: 'C002',
          courierName: 'Sari Wulandari',
          branchID: 'JKT01',
          totalPackages: 40,
          failed: 30,
          failedRate: 0.75,
        },
      ],
    });
  });

  it('mempertahankan failed_change_rate null dan reasons kosong', async () => {
    server.use(
      http.get('/api/v1/failed-deliveries/summary', () =>
        HttpResponse.json({
          data: {
            ...failedSummaryDTO,
            failed_change_rate: null,
            reasons: [],
            top_couriers: [],
          },
        }),
      ),
    );

    const result = await fetchFailedSummary(filter);

    expect(result.failedChangeRate).toBeNull();
    expect(result.reasons).toEqual([]);
    expect(result.topCouriers).toEqual([]);
  });

  it('melempar ApiError untuk 400 invalid_filter', async () => {
    server.use(
      http.get('/api/v1/failed-deliveries/summary', () =>
        HttpResponse.json(
          {
            error: {
              code: 'invalid_filter',
              message: 'rentang terlalu panjang',
            },
          },
          { status: 400 },
        ),
      ),
    );

    const error = await fetchFailedSummary(filter).catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ status: 400, code: 'invalid_filter' });
  });
});

describe('failedSummaryKey', () => {
  it('memuat filter agar cache terpisah per filter', () => {
    expect(failedSummaryKey(filter)).toEqual(['failed-summary', filter]);
  });
});

describe('useFailedSummary', () => {
  it('mengambil data lewat hook', async () => {
    server.use(
      http.get('/api/v1/failed-deliveries/summary', () =>
        HttpResponse.json({ data: failedSummaryDTO }),
      ),
    );
    const client = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    });
    const { result } = renderHook(() => useFailedSummary(filter), {
      wrapper: ({ children }) => (
        <QueryClientProvider client={client}>{children}</QueryClientProvider>
      ),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.totalFailed).toBe(180);
  });
});
