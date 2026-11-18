import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { delay, http, HttpResponse } from 'msw';
import { Route, Routes, useLocation } from 'react-router-dom';

import { branchPerformanceDTOs } from '../../test/fixtures';
import { renderWithProviders } from '../../test/render';
import { server } from '../../test/server';
import { BranchSummaryPage } from './BranchSummaryPage';

const endpoint = '/api/v1/branch-performance';
const today = new Date(2026, 10, 7);

function respond(data = branchPerformanceDTOs, meta = { page: 1, page_size: 10, total: data.length }) {
  server.use(http.get(endpoint, () => HttpResponse.json({ data, meta })));
}

function card(cards: HTMLElement[], index: number): HTMLElement {
  const el = cards[index];
  if (!el) throw new Error('kartu tidak ada');
  return el;
}

function LocationProbe() {
  const location = useLocation();
  return <p data-testid="location">{location.pathname + location.search}</p>;
}

function renderPage() {
  return renderWithProviders(
    <Routes>
      <Route path="/" element={<BranchSummaryPage today={today} />} />
      <Route path="/courier-performance" element={<LocationProbe />} />
    </Routes>,
  );
}

describe('BranchSummaryPage', () => {
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
      http.get(endpoint, () => HttpResponse.json({ data: branchPerformanceDTOs, meta: { page: 1, page_size: 10, total: 2 } })),
    );
    renderPage();
    const alert = await screen.findByRole('alert');
    expect(alert).toHaveTextContent('Gagal memuat data');
    alert.querySelector('button')?.click();
    expect(await screen.findByText('BDG02')).toBeInTheDocument();
  });

  it('menampilkan state kosong', async () => {
    respond([], { page: 1, page_size: 10, total: 0 });
    renderPage();
    expect(await screen.findByText('Tidak ada data cabang untuk periode ini.')).toBeInTheDocument();
  });

  it('menampilkan kartu terurut dengan penanda target dan paginasi', async () => {
    respond();
    renderPage();

    const list = await screen.findByRole('list', { name: 'Ranking cabang' });
    const cards = within(list).getAllByRole('listitem');
    expect(cards).toHaveLength(2);
    expect(card(cards, 0)).toHaveTextContent('BDG02');
    expect(card(cards, 1)).toHaveTextContent('JKT01');
    expect(within(card(cards, 0)).getAllByText(/Tercapai|Di bawah target/)).toHaveLength(3);
    expect(within(card(cards, 1)).queryByText(/Tercapai|Di bawah target/)).toBeNull();
    expect(screen.getByRole('navigation', { name: 'Pagination' })).toBeInTheDocument();
  });

  it('klik kartu menavigasi ke performa kurir dengan filter yang sama', async () => {
    respond();
    renderPage();

    await userEvent.click(await screen.findByRole('link', { name: /BDG02/ }));

    expect(await screen.findByTestId('location')).toHaveTextContent(
      '/courier-performance?branch_id=BDG02&start_date=2026-11-01&end_date=2026-11-07',
    );
  });
});
