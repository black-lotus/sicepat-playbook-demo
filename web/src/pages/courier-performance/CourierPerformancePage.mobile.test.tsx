import { fireEvent, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';

import { courierPerformanceDTOs } from '../../test/fixtures';
import { renderWithProviders } from '../../test/render';
import { server } from '../../test/server';
import { CourierPerformancePage } from './CourierPerformancePage';

const today = new Date(2026, 10, 18);
const endpoint = '/api/v1/courier-performance';
const meta = { page: 1, page_size: 50, total: 4 };

function stubNarrowViewport() {
  vi.stubGlobal(
    'matchMedia',
    vi.fn().mockImplementation((query: string) => ({
      matches: query === '(max-width: 640px)',
      media: query,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    })),
  );
}

describe('CourierPerformancePage di layar sempit', () => {
  beforeEach(() => {
    stubNarrowViewport();
    server.use(http.get(endpoint, () => HttpResponse.json({ data: courierPerformanceDTOs, meta })));
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('menampilkan ringkasan filter yang membuka sheet berisi input Cabang', async () => {
    const user = userEvent.setup();
    renderWithProviders(<CourierPerformancePage today={today} />);
    await screen.findByRole('list', { name: 'Hasil lengkap' });

    expect(screen.queryByLabelText('Cabang')).not.toBeInTheDocument();
    const trigger = screen.getByRole('button', { name: /Semua cabang · 12 sampai 18 Nov/ });
    expect(trigger).toHaveAttribute('aria-expanded', 'false');

    await user.click(trigger);

    const sheet = screen.getByRole('dialog', { name: 'Filter' });
    expect(within(sheet).getByLabelText('Cabang')).toBeInTheDocument();
    expect(within(sheet).getByLabelText('Dari')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Semua cabang/ })).toHaveAttribute('aria-expanded', 'true');

    await user.keyboard('{Escape}');

    expect(screen.queryByRole('dialog', { name: 'Filter' })).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it('ringkasan filter memuat cabang yang dipilih', async () => {
    const user = userEvent.setup();
    renderWithProviders(<CourierPerformancePage today={today} />);
    await screen.findByRole('list', { name: 'Hasil lengkap' });

    await user.click(screen.getByRole('button', { name: /Semua cabang/ }));
    await user.type(screen.getByLabelText('Cabang'), 'jkt01');
    await user.click(screen.getByRole('button', { name: 'Selesai' }));

    expect(await screen.findByRole('button', { name: /JKT01 · 12 sampai 18 Nov/ })).toBeInTheDocument();
  });

  it('menampilkan hasil sebagai daftar peringkat tanpa tabel', async () => {
    renderWithProviders(<CourierPerformancePage today={today} />);

    const list = await screen.findByRole('list', { name: 'Hasil lengkap' });

    expect(screen.queryByRole('table')).not.toBeInTheDocument();
    const items = within(list).getAllByRole('listitem');
    expect(items).toHaveLength(4);
    expect(items[0]).toHaveTextContent('Budi Santoso');
    expect(items[0]).toHaveTextContent('90,0%');
    expect(items[3]).toHaveTextContent('Sari Wulandari');
    expect(within(items[0] as HTMLElement).getByRole('group', { name: /Tren harian 7 hari/ })).toBeInTheDocument();
  });

  it('membuka panel detail dari daftar', async () => {
    const user = userEvent.setup();
    renderWithProviders(<CourierPerformancePage today={today} />);
    const list = await screen.findByRole('list', { name: 'Hasil lengkap' });

    await user.click(within(list).getByRole('button', { name: 'Dewi Lestari' }));

    expect(screen.getByRole('dialog', { name: 'Dewi Lestari' })).toBeInTheDocument();
  });

  it('mengosongkan input Dari tidak membuat halaman crash dan ringkasan meminta periode', async () => {
    const user = userEvent.setup();
    renderWithProviders(<CourierPerformancePage today={today} />);
    await screen.findByRole('list', { name: 'Hasil lengkap' });

    await user.click(screen.getByRole('button', { name: /Semua cabang/ }));
    const sheet = screen.getByRole('dialog', { name: 'Filter' });
    fireEvent.change(within(sheet).getByLabelText('Dari'), { target: { value: '' } });

    expect(screen.getByRole('dialog', { name: 'Filter' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Semua cabang · Pilih periode/ })).toBeInTheDocument();
  });

  it('Tab dari elemen terakhir sheet kembali ke yang pertama, Shift+Tab dari yang pertama ke yang terakhir', async () => {
    const user = userEvent.setup();
    renderWithProviders(<CourierPerformancePage today={today} />);
    await screen.findByRole('list', { name: 'Hasil lengkap' });

    await user.click(screen.getByRole('button', { name: /Semua cabang/ }));
    const sheet = screen.getByRole('dialog', { name: 'Filter' });
    const first = within(sheet).getByLabelText('Dari');
    const middle = within(sheet).getByLabelText('Cabang');
    const last = within(sheet).getByRole('button', { name: 'Selesai' });
    expect(last).toHaveFocus();

    await user.tab();
    expect(first).toHaveFocus();

    await user.tab({ shift: true });
    expect(last).toHaveFocus();

    await user.tab({ shift: true });
    expect(middle).toHaveFocus();
  });

  it('klik backdrop menutup sheet hanya bila pointerdown dan click sama-sama di backdrop', async () => {
    const user = userEvent.setup();
    renderWithProviders(<CourierPerformancePage today={today} />);
    await screen.findByRole('list', { name: 'Hasil lengkap' });
    await user.click(screen.getByRole('button', { name: /Semua cabang/ }));
    const sheet = screen.getByRole('dialog', { name: 'Filter' });
    const backdrop = sheet.parentElement as HTMLElement;

    fireEvent.pointerDown(sheet);
    fireEvent.click(backdrop);
    expect(screen.getByRole('dialog', { name: 'Filter' })).toBeInTheDocument();

    fireEvent.pointerDown(backdrop);
    fireEvent.click(backdrop);
    expect(screen.queryByRole('dialog', { name: 'Filter' })).not.toBeInTheDocument();
  });

  it('mengunci scroll halaman selama sheet terbuka dan memulihkannya setelah ditutup', async () => {
    const user = userEvent.setup();
    renderWithProviders(<CourierPerformancePage today={today} />);
    await screen.findByRole('list', { name: 'Hasil lengkap' });
    expect(document.body.style.overflow).toBe('');

    await user.click(screen.getByRole('button', { name: /Semua cabang/ }));
    expect(document.body.style.overflow).toBe('hidden');

    await user.click(screen.getByRole('button', { name: 'Selesai' }));
    expect(document.body.style.overflow).toBe('');
  });
});
