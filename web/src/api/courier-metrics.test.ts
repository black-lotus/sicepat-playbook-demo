import { http, HttpResponse } from 'msw';

import { courierMetricDTO } from '../test/fixtures';
import { server } from '../test/server';
import { fetchCourierMetrics } from './courier-metrics';

describe('fetchCourierMetrics', () => {
  it('memetakan filter ke query API dan response ke camelCase', async () => {
    let params: URLSearchParams | undefined;
    server.use(
      http.get('/api/v1/courier-metrics', ({ request }) => {
        params = new URL(request.url).searchParams;
        return HttpResponse.json({ data: [courierMetricDTO], meta: { page: 2, page_size: 10, total: 11 } });
      }),
    );

    const result = await fetchCourierMetrics({
      startDate: '2026-11-01',
      endDate: '2026-11-17',
      branchID: 'JKT01',
      page: 2,
      pageSize: 10,
    });

    expect(Object.fromEntries(params ?? [])).toEqual({
      start_date: '2026-11-01',
      end_date: '2026-11-17',
      branch_id: 'JKT01',
      page: '2',
      page_size: '10',
    });
    expect(result).toEqual({
      items: [
        {
          courierID: 'C001',
          courierName: 'Budi',
          branchID: 'JKT01',
          date: '2026-11-17',
          totalPackages: 1250,
          delivered: 1000,
          failed: 250,
          onTime: 750,
          successRate: 0.8,
          slaRate: 0.75,
        },
      ],
      meta: { page: 2, pageSize: 10, total: 11 },
    });
  });
});
