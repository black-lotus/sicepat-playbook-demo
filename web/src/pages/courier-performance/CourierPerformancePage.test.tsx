import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';

import { courierMetricDTO } from '../../test/fixtures';
import { renderWithQuery } from '../../test/render';
import { server } from '../../test/server';
import { CourierPerformancePage } from './CourierPerformancePage';

const today = new Date(2026, 10, 18);
const endpoint = '/api/v1/courier-metrics';

function renderPage() {
  return renderWithQuery(<CourierPerformancePage today={today} />);
}

describe('CourierPerformancePage', () => {
  it('menampilkan state loading', async () => {
    server.use(
      http.get(endpoint, async () => {
        await delay('infinite');
        return HttpResponse.json({});
      }),
    );

    renderPage();

    expect(await screen.findByRole('status')).toHaveTextContent('Memuat data');
  });

  it('menampilkan state error dan bisa mencoba lagi', async () => {
    server.use(
      http.get(endpoint, () => HttpResponse.json({ error: { code: 'internal_error', message: 'terjadi kesalahan pada server' } }, { status: 500 }), { once: true }),
      http.get(endpoint, () => HttpResponse.json({ data: [courierMetricDTO], meta: { page: 1, page_size: 20, total: 1 } })),
    );

    renderPage();

    expect(await screen.findByRole('alert')).toHaveTextContent('terjadi kesalahan pada server');
    await userEvent.click(screen.getByRole('button', { name: 'Coba lagi' }));
    expect(await screen.findByRole('cell', { name: 'Budi' })).toBeInTheDocument();
  });

  it('menampilkan state data kosong', async () => {
    server.use(http.get(endpoint, () => HttpResponse.json({ data: [], meta: { page: 1, page_size: 20, total: 0 } })));

    renderPage();

    expect(await screen.findByText('Belum ada data performa kurir untuk periode ini.')).toBeInTheDocument();
  });

  it('menampilkan data terisi dengan angka yang diformat, tanpa menghitung ulang KPI', async () => {
    let params: URLSearchParams | undefined;
    server.use(
      http.get(endpoint, ({ request }) => {
        params = new URL(request.url).searchParams;
        return HttpResponse.json({ data: [courierMetricDTO], meta: { page: 1, page_size: 20, total: 1 } });
      }),
    );

    renderPage();

    const row = (await screen.findByRole('cell', { name: 'Budi' })).closest('tr');
    expect(row).not.toBeNull();
    const cells = within(row as HTMLTableRowElement).getAllByRole('cell').map((c) => c.textContent);
    expect(cells).toEqual(['17 Nov 2026', 'Budi', 'JKT01', '1.250', '1.000', '250', '80,0%', '75,0%']);
    expect(params?.get('start_date')).toBe('2026-11-12');
    expect(params?.get('end_date')).toBe('2026-11-18');
    expect(screen.getByText('Halaman 1 dari 1')).toBeInTheDocument();
  });

  it('pindah halaman dan filter cabang mengirim parameter yang sesuai', async () => {
    const requests: URLSearchParams[] = [];
    server.use(
      http.get(endpoint, ({ request }) => {
        const params = new URL(request.url).searchParams;
        requests.push(params);
        const page = Number(params.get('page'));
        return HttpResponse.json({
          data: [{ ...courierMetricDTO, courier_id: `C00${page}`, courier_name: `Kurir ${page}` }],
          meta: { page, page_size: 20, total: 45 },
        });
      }),
    );
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText('Halaman 1 dari 3')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Berikutnya' }));
    expect(await screen.findByRole('cell', { name: 'Kurir 2' })).toBeInTheDocument();
    expect(screen.getByText('Halaman 2 dari 3')).toBeInTheDocument();

    await user.type(screen.getByLabelText('Cabang'), 'jkt01');
    expect(await screen.findByText('Halaman 1 dari 3')).toBeInTheDocument();
    expect(requests.at(-1)?.get('branch_id')).toBe('JKT01');
    expect(requests.at(-1)?.get('page')).toBe('1');
  });

  it('mengubah periode mengirim tanggal baru', async () => {
    const requests: URLSearchParams[] = [];
    server.use(
      http.get(endpoint, ({ request }) => {
        requests.push(new URL(request.url).searchParams);
        return HttpResponse.json({ data: [], meta: { page: 1, page_size: 20, total: 0 } });
      }),
    );

    renderPage();
    await screen.findByText('Belum ada data performa kurir untuk periode ini.');

    fireEvent.change(screen.getByLabelText('Dari'), { target: { value: '2026-11-01' } });
    fireEvent.change(screen.getByLabelText('Sampai'), { target: { value: '2026-11-10' } });

    await waitFor(() => {
      expect(requests.at(-1)?.get('start_date')).toBe('2026-11-01');
      expect(requests.at(-1)?.get('end_date')).toBe('2026-11-10');
    });
  });
});
