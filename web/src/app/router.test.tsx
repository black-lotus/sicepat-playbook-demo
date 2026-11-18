import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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
    server.use(http.get('/api/v1/courier-performance', () => HttpResponse.json({ data: [], meta: { page: 1, page_size: 50, total: 0 } })));
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

  it('menampilkan pesan dan tautan kembali bila halaman melempar error', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => undefined);
    function Boom(): never {
      throw new Error('rusak');
    }
    const router = createMemoryRouter(
      [
        {
          element: routes[0]?.element,
          errorElement: routes[0]?.errorElement,
          children: [...(routes[0]?.children ?? []), { path: 'boom', element: <Boom /> }],
        },
      ],
      { initialEntries: ['/boom'] },
    );
    render(
      <AppProviders>
        <RouterProvider router={router} />
      </AppProviders>,
    );

    expect(await screen.findByText('Terjadi kesalahan pada halaman ini.')).toBeInTheDocument();
    await userEvent.click(screen.getByRole('link', { name: 'Kembali ke Performa Kurir' }));
    expect(await screen.findByRole('heading', { name: 'Performa Kurir' })).toBeInTheDocument();
    vi.restoreAllMocks();
  });
});
