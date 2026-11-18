import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';

import { courierPerformanceDTOs } from '../../test/fixtures';
import { renderWithQuery } from '../../test/render';
import { server } from '../../test/server';
import { CourierPerformancePage } from './CourierPerformancePage';

const today = new Date(2026, 10, 18);
const endpoint = '/api/v1/courier-performance';

function renderPage() {
  return renderWithQuery(<CourierPerformancePage today={today} />);
}

const meta = { page: 1, page_size: 50, total: 4 };

function rowOf(name: string): HTMLTableRowElement {
  const table = screen.getByRole('table', { name: 'Hasil lengkap' });
  const row = within(table).getByRole('button', { name }).closest('tr');
  expect(row).not.toBeNull();
  return row as HTMLTableRowElement;
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

  it('merender skeleton loadingFallback di dalam role status', async () => {
    server.use(
      http.get(endpoint, async () => {
        await delay('infinite');
        return HttpResponse.json({});
      }),
    );

    const { container } = renderPage();

    const status = await screen.findByRole('status');
    const skeleton = container.querySelector('.skeleton');
    expect(skeleton).not.toBeNull();
    expect(status).toContainElement(skeleton as HTMLElement);
    expect(status.querySelectorAll('.skeleton-block')).toHaveLength(3);
    expect(status.querySelectorAll('.skeleton-row')).toHaveLength(3);
  });

  it('menampilkan state error dan bisa mencoba lagi', async () => {
    server.use(
      http.get(endpoint, () => HttpResponse.json({ error: { code: 'internal_error', message: 'terjadi kesalahan pada server' } }, { status: 500 }), { once: true }),
      http.get(endpoint, () => HttpResponse.json({ data: courierPerformanceDTOs, meta })),
    );

    renderPage();

    expect(await screen.findByRole('alert')).toHaveTextContent('terjadi kesalahan pada server');
    await userEvent.click(screen.getByRole('button', { name: 'Coba lagi' }));
    expect(await screen.findByRole('table', { name: 'Hasil lengkap' })).toBeInTheDocument();
  });

  it('menampilkan state data kosong', async () => {
    server.use(http.get(endpoint, () => HttpResponse.json({ data: [], meta: { page: 1, page_size: 50, total: 0 } })));

    renderPage();

    expect(await screen.findByText('Belum ada data performa kurir untuk periode ini.')).toBeInTheDocument();
  });

  it('menampilkan podium dan tabel hasil dengan angka dari API', async () => {
    let params: URLSearchParams | undefined;
    server.use(
      http.get(endpoint, ({ request }) => {
        params = new URL(request.url).searchParams;
        return HttpResponse.json({ data: courierPerformanceDTOs, meta });
      }),
    );

    renderPage();

    const podium = await screen.findByRole('list', { name: 'Podium' });
    const names = within(podium)
      .getAllByRole('button')
      .map((b) => b.textContent ?? '');
    expect(names).toHaveLength(3);
    expect(names[0]).toContain('Budi Santoso');
    expect(names[1]).toContain('Dewi Lestari');
    expect(names[2]).toContain('Agus Pratama');

    const table = screen.getByRole('table', { name: 'Hasil lengkap' });
    expect(within(table).getAllByRole('row')).toHaveLength(5);
    expect(rowOf('Budi Santoso')).toHaveTextContent('90,0%');
    expect(rowOf('Budi Santoso')).toHaveTextContent('91,7%');
    expect(rowOf('Budi Santoso')).toHaveTextContent('Pemimpin');
    expect(rowOf('Dewi Lestari')).toHaveTextContent('85,7%');
    expect(rowOf('Dewi Lestari')).toHaveTextContent('100,0%');
    expect(rowOf('Dewi Lestari')).toHaveTextContent('−4,3 pp');
    expect(rowOf('Sari Wulandari')).toHaveTextContent('0,0%');
    expect(rowOf('Sari Wulandari')).toHaveTextContent('−90,0 pp');
    expect(params?.get('start_date')).toBe('2026-11-12');
    expect(params?.get('end_date')).toBe('2026-11-18');
    expect(params?.get('page_size')).toBe('50');
  });

  it('pindah ke halaman 2 tanpa podium dengan rank global, lalu mengubah cabang kembali ke halaman 1', async () => {
    const requests: URLSearchParams[] = [];
    server.use(
      http.get(endpoint, ({ request }) => {
        const params = new URL(request.url).searchParams;
        requests.push(params);
        if (params.get('page') === '2') {
          return HttpResponse.json({
            data: [{ ...courierPerformanceDTOs[1], rank: 51, courier_id: 'C051', courier_name: 'Kurir Lima Puluh Satu' }],
            meta: { page: 2, page_size: 50, total: 51 },
          });
        }
        return HttpResponse.json({ data: courierPerformanceDTOs, meta: { page: 1, page_size: 50, total: 51 } });
      }),
    );
    const user = userEvent.setup();

    renderPage();
    await screen.findByRole('list', { name: 'Podium' });
    expect(requests.at(-1)?.get('page')).toBe('1');

    await user.click(screen.getByRole('button', { name: 'Berikutnya' }));

    await screen.findByRole('button', { name: 'Kurir Lima Puluh Satu' });
    expect(requests.at(-1)?.get('page')).toBe('2');
    expect(screen.queryByRole('list', { name: 'Podium' })).not.toBeInTheDocument();
    expect(within(rowOf('Kurir Lima Puluh Satu')).getAllByRole('cell')[0]).toHaveTextContent('51');
    expect(screen.getByText('Halaman 2 dari 2')).toBeInTheDocument();

    await user.type(screen.getByLabelText('Cabang'), 'jkt01');

    await waitFor(() => {
      expect(requests.at(-1)?.get('branch_id')).toBe('JKT01');
    });
    expect(requests.at(-1)?.get('page')).toBe('1');
    expect(await screen.findByRole('list', { name: 'Podium' })).toBeInTheDocument();
    expect(screen.getByText('Halaman 1 dari 2')).toBeInTheDocument();
  });

  it('menandai hasil lama dengan aria-busy selama request filter baru tertunda', async () => {
    let release: () => void = () => undefined;
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    server.use(
      http.get(endpoint, async ({ request }) => {
        if (new URL(request.url).searchParams.get('branch_id') === 'J') await gate;
        return HttpResponse.json({ data: courierPerformanceDTOs, meta });
      }),
    );
    const user = userEvent.setup();

    renderPage();
    const table = await screen.findByRole('table', { name: 'Hasil lengkap' });
    const results = table.closest('.results');
    expect(results).not.toBeNull();
    expect(results).not.toHaveAttribute('aria-busy');
    expect(results).not.toHaveClass('results-stale');

    await user.type(screen.getByLabelText('Cabang'), 'j');

    await waitFor(() => {
      expect(results).toHaveAttribute('aria-busy', 'true');
    });
    expect(results).toHaveClass('results-stale');
    expect(screen.getByRole('table', { name: 'Hasil lengkap' })).toBeInTheDocument();

    release();

    await waitFor(() => {
      expect(results).not.toHaveAttribute('aria-busy');
    });
    expect(results).not.toHaveClass('results-stale');
  });

  it('mengubah cabang mengirim branch_id baru dan kembali ke halaman 1', async () => {
    const requests: URLSearchParams[] = [];
    server.use(
      http.get(endpoint, ({ request }) => {
        requests.push(new URL(request.url).searchParams);
        return HttpResponse.json({ data: courierPerformanceDTOs, meta });
      }),
    );
    const user = userEvent.setup();

    renderPage();
    await screen.findByRole('table', { name: 'Hasil lengkap' });

    await user.type(screen.getByLabelText('Cabang'), 'jkt01');

    await waitFor(() => {
      expect(requests.at(-1)?.get('branch_id')).toBe('JKT01');
    });
    expect(requests.at(-1)?.get('page')).toBe('1');
  });

  it('mengubah periode mengirim tanggal baru', async () => {
    const requests: URLSearchParams[] = [];
    server.use(
      http.get(endpoint, ({ request }) => {
        requests.push(new URL(request.url).searchParams);
        return HttpResponse.json({ data: [], meta: { page: 1, page_size: 50, total: 0 } });
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

  it('peringkat bawah netral tanpa penanda negatif', async () => {
    server.use(http.get(endpoint, () => HttpResponse.json({ data: courierPerformanceDTOs, meta })));

    renderPage();

    await screen.findByRole('table', { name: 'Hasil lengkap' });
    const row = rowOf('Sari Wulandari');
    expect(row.className).not.toMatch(/neg|bad|danger|error|fail|low/i);
    expect(row.outerHTML).not.toMatch(/neg|danger|error/i);
  });


  it('section halaman memakai kelas halaman agar CSS tidak bocor ke section lain', async () => {
    server.use(http.get(endpoint, () => HttpResponse.json({ data: courierPerformanceDTOs, meta })));

    const { container } = renderPage();
    await screen.findByRole('table', { name: 'Hasil lengkap' });

    expect(container.querySelector('section')).toHaveClass('courier-performance-page');
  });
});
