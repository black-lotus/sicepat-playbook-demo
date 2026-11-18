import { fireEvent, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';

import { CourierPerformancePage } from '../../pages/courier-performance/CourierPerformancePage';
import { courierPerformanceDTOs } from '../../test/fixtures';
import { renderWithProviders } from '../../test/render';
import { server } from '../../test/server';

const today = new Date(2026, 10, 18);
const endpoint = '/api/v1/courier-performance';
const meta = { page: 1, page_size: 50, total: 4 };

async function renderPage() {
  server.use(http.get(endpoint, () => HttpResponse.json({ data: courierPerformanceDTOs, meta })));
  renderWithProviders(<CourierPerformancePage today={today} />);
  return screen.findByRole('table', { name: 'Hasil lengkap' });
}

describe('CourierDetailPanel', () => {
  it('membuka dialog dari tabel dengan tabel harian dan grafik', async () => {
    const user = userEvent.setup();
    const table = await renderPage();

    await user.click(within(table).getByRole('button', { name: 'Dewi Lestari' }));

    const dialog = screen.getByRole('dialog', { name: 'Dewi Lestari' });
    expect(dialog).toHaveAttribute('aria-modal', 'true');
    expect(within(dialog).getByRole('button', { name: 'Tutup' })).toHaveFocus();
    // Tabel harian adalah alternatif teks di dalam <details>; dibuka dulu seperti pengguna.
    await user.click(within(dialog).getByText('Rincian harian'));
    const daily = within(dialog).getByRole('table');
    expect(within(daily).getAllByRole('row')).toHaveLength(8);
    expect(within(daily).getByRole('columnheader', { name: 'Tepat SLA' })).toBeInTheDocument();
    expect(within(dialog).getByRole('img', { name: /Grafik batang paket terkirim per hari/ })).toBeInTheDocument();
    expect(within(dialog).getByRole('img', { name: /Komposisi 35 paket/ })).toBeInTheDocument();
    expect(within(dialog).getByRole('heading', { name: 'Komposisi paket' })).toBeInTheDocument();
  });

  it('menutup dengan Esc dan mengembalikan fokus ke pemicu', async () => {
    const user = userEvent.setup();
    const table = await renderPage();
    const trigger = within(table).getByRole('button', { name: 'Dewi Lestari' });

    await user.click(trigger);
    await user.keyboard('{Escape}');

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it('menutup lewat tombol Tutup', async () => {
    const user = userEvent.setup();
    const table = await renderPage();

    await user.click(within(table).getByRole('button', { name: 'Dewi Lestari' }));
    await user.click(screen.getByRole('button', { name: 'Tutup' }));

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('membuka dialog dari kartu podium', async () => {
    const user = userEvent.setup();
    await renderPage();

    await user.click(within(screen.getByRole('list', { name: 'Podium' })).getByRole('button', { name: /Budi Santoso/ }));

    expect(screen.getByRole('dialog', { name: 'Budi Santoso' })).toBeInTheDocument();
  });

  it('Tab dan Shift+Tab berputar di panel (Tutup dan ringkasan rincian harian); fokus tidak lari ke halaman di belakang', async () => {
    const user = userEvent.setup();
    const table = await renderPage();

    await user.click(within(table).getByRole('button', { name: 'Dewi Lestari' }));
    const dialog = screen.getByRole('dialog', { name: 'Dewi Lestari' });
    const close = within(dialog).getByRole('button', { name: 'Tutup' });
    expect(close).toHaveFocus();

    // Elemen fokus panel: Tutup lalu ringkasan "Rincian harian". Tab maju dan mundur berputar di antara keduanya.
    const summary = within(dialog).getByText('Rincian harian');
    await user.tab();
    expect(summary).toHaveFocus();
    await user.tab();
    expect(close).toHaveFocus();
    await user.tab({ shift: true });
    expect(summary).toHaveFocus();
    await user.tab({ shift: true });
    expect(close).toHaveFocus();
  });

  it('klik backdrop menutup dialog bila pointerdown dan click sama-sama di backdrop', async () => {
    const user = userEvent.setup();
    const table = await renderPage();
    await user.click(within(table).getByRole('button', { name: 'Dewi Lestari' }));
    const backdrop = screen.getByRole('dialog', { name: 'Dewi Lestari' }).parentElement as HTMLElement;

    fireEvent.pointerDown(backdrop);
    fireEvent.click(backdrop);

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('pointerdown di dalam panel lalu click di backdrop (drag-select) tidak menutup dialog', async () => {
    const user = userEvent.setup();
    const table = await renderPage();
    await user.click(within(table).getByRole('button', { name: 'Dewi Lestari' }));
    const dialog = screen.getByRole('dialog', { name: 'Dewi Lestari' });
    const backdrop = dialog.parentElement as HTMLElement;

    fireEvent.pointerDown(dialog);
    fireEvent.click(backdrop);

    expect(screen.getByRole('dialog', { name: 'Dewi Lestari' })).toBeInTheDocument();

    // Setelah gestur itu, klik backdrop tanpa pointerdown sendiri juga tidak boleh menutup.
    fireEvent.click(backdrop);
    expect(screen.getByRole('dialog', { name: 'Dewi Lestari' })).toBeInTheDocument();
  });

  it('klik di dalam panel tidak menutup dialog', async () => {
    const user = userEvent.setup();
    const table = await renderPage();
    await user.click(within(table).getByRole('button', { name: 'Dewi Lestari' }));

    await user.click(screen.getByRole('heading', { name: 'Dewi Lestari' }));

    expect(screen.getByRole('dialog', { name: 'Dewi Lestari' })).toBeInTheDocument();
  });

  it('mengunci scroll halaman saat terbuka dan memulihkannya saat tertutup', async () => {
    const user = userEvent.setup();
    document.body.style.overflow = 'auto';
    const table = await renderPage();
    expect(document.body.style.overflow).toBe('auto');

    await user.click(within(table).getByRole('button', { name: 'Dewi Lestari' }));
    expect(document.body.style.overflow).toBe('hidden');

    await user.keyboard('{Escape}');
    expect(document.body.style.overflow).toBe('auto');

    document.body.style.overflow = '';
  });
});
