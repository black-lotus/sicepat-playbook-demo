import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';

import { failedSummaryDTO } from '../../test/fixtures';
import { renderWithProviders } from '../../test/render';
import { server } from '../../test/server';
import { FailedDeliveryPage } from './FailedDeliveryPage';

const endpoint = '/api/v1/failed-deliveries/summary';
const today = new Date(2026, 10, 18);

function renderPage() {
  return renderWithProviders(<FailedDeliveryPage today={today} />);
}

function respond(overrides: Partial<typeof failedSummaryDTO>) {
  server.use(http.get(endpoint, () => HttpResponse.json({ data: { ...failedSummaryDTO, ...overrides } })));
}

describe('FailedDeliveryPage', () => {
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
      http.get(endpoint, () => HttpResponse.json({ data: failedSummaryDTO })),
    );

    renderPage();

    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent('Gagal memuat data');
    alert.querySelector('button')?.click();
    expect(await screen.findByText('Penerima tidak di tempat', { selector: '.legend-name' })).toBeInTheDocument();
  });

  it('menampilkan state kosong bila tidak ada paket gagal', async () => {
    respond({ total_failed: 0, failed_rate: 0, previous_total_failed: 0, failed_change_rate: null, reasons: [], top_couriers: [] });

    renderPage();

    expect(await screen.findByText('Tidak ada paket gagal kirim untuk periode ini.')).toBeInTheDocument();
  });

  it('menampilkan hero, chip turun, donut dengan legenda, dan daftar kurir', async () => {
    respond({});
    renderPage();

    const hero = await screen.findByRole('region', { name: 'Total gagal kirim' });
    expect(hero).toHaveTextContent('180');
    expect(hero).toHaveTextContent('15,0%');
    const chip = within(hero).getByText('Turun 10,0%');
    expect(chip.closest('[data-trend]')).toHaveAttribute('data-trend', 'down');
    expect(chip.closest('[data-trend]')?.querySelector('svg[aria-hidden="true"]')).not.toBeNull();

    const legend = within(screen.getByRole('region', { name: 'Alasan gagal kirim' })).getByRole('list');
    const row = within(legend).getByText('Penerima tidak di tempat').closest('li') as HTMLElement;
    expect(row).toHaveTextContent('60');
    expect(row).toHaveTextContent('33,3%');
    expect(within(legend).getByText('Tanpa alasan').closest('li')).toHaveTextContent('11,1%');

    const couriers = within(screen.getByRole('region', { name: 'Kurir dengan gagal terbanyak' })).getByRole('list');
    const item = within(couriers).getByText('Sari Wulandari').closest('li') as HTMLElement;
    expect(item).toHaveTextContent('JKT01');
    expect(item).toHaveTextContent('30');
    expect(item).toHaveTextContent('75,0%');
  });

  it('chip naik memakai panah atas dan teks Naik', async () => {
    respond({ failed_change_rate: 0.25 });

    renderPage();

    const chip = (await screen.findByText('Naik 25,0%')).closest('[data-trend]') as HTMLElement;
    expect(chip).toHaveAttribute('data-trend', 'up');
    expect(chip.querySelector('svg[aria-hidden="true"]')).not.toBeNull();
  });

  it('chip tanpa pembanding hanya berisi teks tanpa panah', async () => {
    respond({ failed_change_rate: null, previous_total_failed: 0 });

    renderPage();

    const chip = (await screen.findByText('Tidak ada pembanding')).closest('[data-trend]') as HTMLElement;
    expect(chip).toHaveAttribute('data-trend', 'none');
    expect(chip.querySelector('svg')).toBeNull();
  });

  it('mengirim filter cabang ke API dalam huruf besar', async () => {
    const branches: (string | null)[] = [];
    server.use(
      http.get(endpoint, ({ request }) => {
        branches.push(new URL(request.url).searchParams.get('branch_id'));
        return HttpResponse.json({ data: failedSummaryDTO });
      }),
    );
    renderPage();
    await screen.findByRole('region', { name: 'Total gagal kirim' });

    await userEvent.type(screen.getByLabelText('Cabang'), 'jkt01');

    await waitFor(() => expect(branches.at(-1)).toBe('JKT01'));
    expect(branches[0]).toBeNull();
  });
});
