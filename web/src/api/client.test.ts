import { http, HttpResponse } from 'msw';

import { server } from '../test/server';
import { ApiError, apiGet } from './client';

describe('apiGet', () => {
  it('mengirim query tanpa parameter kosong dan mengembalikan envelope', async () => {
    let requestedURL = '';
    server.use(
      http.get('/api/v1/things', ({ request }) => {
        requestedURL = request.url;
        return HttpResponse.json({ data: [1, 2], meta: { total: 2 } });
      }),
    );

    const body = await apiGet<number[], { total: number }>('/api/v1/things', { a: 'x', b: '', c: undefined, page: 2 });

    expect(body).toEqual({ data: [1, 2], meta: { total: 2 } });
    expect(new URL(requestedURL).search).toBe('?a=x&page=2');
  });

  it('melempar ApiError dengan kode dari backend', async () => {
    server.use(
      http.get('/api/v1/things', () =>
        HttpResponse.json({ error: { code: 'invalid_filter', message: 'periode maksimal 31 hari' } }, { status: 400 }),
      ),
    );

    const error = await apiGet('/api/v1/things').catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ status: 400, code: 'invalid_filter', message: 'periode maksimal 31 hari' });
  });

  it('melempar ApiError umum bila body error tidak sesuai format', async () => {
    server.use(http.get('/api/v1/things', () => HttpResponse.json({ oops: true }, { status: 502 })));

    await expect(apiGet('/api/v1/things')).rejects.toMatchObject({ status: 502, code: 'unknown_error' });
  });

  it('melempar ApiError bila response bukan JSON', async () => {
    server.use(http.get('/api/v1/things', () => new HttpResponse('Bad Gateway', { status: 502 })));

    await expect(apiGet('/api/v1/things')).rejects.toMatchObject({ code: 'invalid_response' });
  });

  it('melempar ApiError bila response sukses tanpa data', async () => {
    server.use(http.get('/api/v1/things', () => HttpResponse.json({ items: [] })));

    await expect(apiGet('/api/v1/things')).rejects.toMatchObject({ code: 'invalid_response' });
  });
});
