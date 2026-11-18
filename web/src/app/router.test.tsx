import { render, screen } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { createMemoryRouter, RouterProvider } from 'react-router-dom';

import { server } from '../test/server';
import { AppProviders } from './providers';
import { routes } from './router';

function renderAt(path: string) {
  const router = createMemoryRouter(routes, { initialEntries: [path] });
  return render(
    <AppProviders>
      <RouterProvider router={router} />
    </AppProviders>,
  );
}

describe('router', () => {
  beforeEach(() => {
    server.use(http.get('/api/v1/courier-metrics', () => HttpResponse.json({ data: [], meta: { page: 1, page_size: 20, total: 0 } })));
  });

  it('mengarahkan root ke halaman performa kurir', async () => {
    renderAt('/');
    expect(await screen.findByRole('heading', { name: 'Performa Kurir' })).toBeInTheDocument();
  });

  it.each([
    ['/branch-summary', 'Ringkasan Cabang'],
    ['/failed-delivery', 'Gagal Kirim'],
  ])('membuka %s', async (path, heading) => {
    renderAt(path);
    expect(await screen.findByRole('heading', { name: heading })).toBeInTheDocument();
  });
});
