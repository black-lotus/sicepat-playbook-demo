import { http, HttpResponse } from 'msw';

import { courierPerformanceDTOs } from '../test/fixtures';
import { server } from '../test/server';
import { ApiError } from './client';
import { courierPerformanceKey, fetchCourierPerformance, toCourierPerformance } from './courier-performance';

const filter = {
  startDate: '2026-11-01',
  endDate: '2026-11-07',
  branchID: 'JKT01',
  page: 1,
  pageSize: 50,
};

describe('fetchCourierPerformance', () => {
  it('mengirim parameter query dengan benar', async () => {
    let params: URLSearchParams | undefined;
    server.use(
      http.get('/api/v1/courier-performance', ({ request }) => {
        params = new URL(request.url).searchParams;
        return HttpResponse.json({ data: [], meta: { page: 1, page_size: 50, total: 0 } });
      }),
    );

    await fetchCourierPerformance(filter);

    expect(Object.fromEntries(params ?? [])).toEqual({
      start_date: '2026-11-01',
      end_date: '2026-11-07',
      branch_id: 'JKT01',
      page: '1',
      page_size: '50',
    });
  });

  it('memetakan respons ke camelCase termasuk daily', async () => {
    server.use(
      http.get('/api/v1/courier-performance', () =>
        HttpResponse.json({ data: courierPerformanceDTOs, meta: { page: 1, page_size: 50, total: 4 } }),
      ),
    );

    const result = await fetchCourierPerformance(filter);

    expect(result.meta).toEqual({ page: 1, pageSize: 50, total: 4, branchTarget: null });
    expect(result.items).toHaveLength(4);
    expect(result.items[1]).toEqual({
      rank: 2,
      courierID: 'C003',
      courierName: 'Dewi Lestari',
      branchID: 'JKT01',
      totalPackages: 35,
      delivered: 30,
      failed: 5,
      onTime: 30,
      successRate: 30 / 35,
      slaRate: 1,
      successRateGap: 0.9 - 30 / 35,
      successRateStatus: null,
      daily: [
        { date: '2026-11-01', totalPackages: 5, delivered: 5, failed: 0, onTime: 5 },
        { date: '2026-11-02', totalPackages: 5, delivered: 4, failed: 1, onTime: 4 },
        { date: '2026-11-03', totalPackages: 5, delivered: 4, failed: 1, onTime: 4 },
        { date: '2026-11-04', totalPackages: 5, delivered: 5, failed: 0, onTime: 5 },
        { date: '2026-11-05', totalPackages: 5, delivered: 4, failed: 1, onTime: 4 },
        { date: '2026-11-06', totalPackages: 5, delivered: 4, failed: 1, onTime: 4 },
        { date: '2026-11-07', totalPackages: 5, delivered: 4, failed: 1, onTime: 4 },
      ],
    });
    expect(courierPerformanceDTOs.map(toCourierPerformance)[0]?.daily).toHaveLength(7);
  });

  it('memetakan meta.branch_target dan success_rate_status', async () => {
    server.use(
      http.get('/api/v1/courier-performance', () =>
        HttpResponse.json({
          data: [{ ...courierPerformanceDTOs[0], success_rate_status: 'met' }],
          meta: {
            page: 1,
            page_size: 50,
            total: 1,
            branch_target: { month: '2026-11', success_rate: 0.9, sla_rate: 0.85 },
          },
        }),
      ),
    );

    const result = await fetchCourierPerformance(filter);

    expect(result.meta.branchTarget).toEqual({ month: '2026-11', successRate: 0.9, slaRate: 0.85 });
    expect(result.items[0]?.successRateStatus).toBe('met');
  });

  it('melempar ApiError dengan code dan message dari backend untuk 400 invalid_filter', async () => {
    server.use(
      http.get('/api/v1/courier-performance', () =>
        HttpResponse.json(
          { error: { code: 'invalid_filter', message: 'start_date harus sebelum end_date' } },
          { status: 400 },
        ),
      ),
    );

    const error = await fetchCourierPerformance(filter).catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({
      status: 400,
      code: 'invalid_filter',
      message: 'start_date harus sebelum end_date',
    });
  });
});

describe('courierPerformanceKey', () => {
  it('memuat filter agar cache terpisah per filter', () => {
    expect(courierPerformanceKey(filter)).toEqual(['courier-performance', filter]);
  });
});
