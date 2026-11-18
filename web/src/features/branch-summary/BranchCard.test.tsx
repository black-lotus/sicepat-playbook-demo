import { screen, within } from '@testing-library/react';

import { toBranchPerformance } from '../../api/branch-performance';
import { nestedInteractive } from '../../test/a11y';
import { branchPerformanceDTOs } from '../../test/fixtures';
import { renderWithProviders } from '../../test/render';
import { BranchCard } from './BranchCard';

function branchDTO(index: number) {
  const dto = branchPerformanceDTOs[index];
  if (!dto) throw new Error('fixture cabang tidak ada');
  return dto;
}

const range = { startDate: '2026-11-01', endDate: '2026-11-07' };

describe('BranchCard', () => {
  it('menampilkan rank, angka, dan tiga penanda target dengan ikon dan teks', () => {
    const branch = toBranchPerformance(branchDTO(0));
    renderWithProviders(<BranchCard branch={branch} {...range} />);

    const card = screen.getByRole('article', { name: 'BDG02' });
    expect(card).toHaveTextContent('1');
    expect(card).toHaveTextContent('83,3%');
    expect(card).toHaveTextContent('84,0%');
    expect(card).toHaveTextContent('833');
    expect(card).toHaveTextContent('8 kurir');

    const badges = within(card).getAllByText(/Tercapai|Di bawah target/);
    expect(badges.map((b) => b.textContent)).toEqual(['Di bawah target', 'Tercapai', 'Di bawah target']);
    for (const badge of badges) {
      expect(badge.closest('.status-badge')?.querySelector('svg[aria-hidden="true"]')).not.toBeNull();
    }
  });

  it('tanpa target tidak menampilkan penanda', () => {
    const branch = toBranchPerformance(branchDTO(1));
    renderWithProviders(<BranchCard branch={branch} {...range} />);

    expect(screen.queryByText(/Tercapai|Di bawah target/)).toBeNull();
    expect(screen.getByText('Belum ada target')).toBeInTheDocument();
  });

  it('menautkan ke performa kurir dengan filter yang sama', () => {
    const branch = toBranchPerformance(branchDTO(1));
    renderWithProviders(<BranchCard branch={branch} {...range} />);

    expect(screen.getByRole('link')).toHaveAttribute(
      'href',
      '/courier-performance?branch_id=JKT01&start_date=2026-11-01&end_date=2026-11-07',
    );
  });

  it('tautan hanya membungkus nama cabang, bukan grafik harian, dan tidak ada elemen interaktif bersarang', () => {
    const branch = toBranchPerformance({ ...branchDTO(0), daily: manyDays() });
    const { container } = renderWithProviders(<BranchCard branch={branch} {...range} />);

    const link = screen.getByRole('link', { name: 'BDG02' });
    expect(link.querySelector('.daily-chart')).toBeNull();
    expect(screen.getByRole('group', { name: /Tren harian 7 hari/ })).toBeInTheDocument();
    expect(nestedInteractive(container)).toEqual([]);
  });

  it('menampilkan grafik harian bersegmen dengan label tanggal dan legenda', () => {
    const branch = toBranchPerformance({ ...branchDTO(0), daily: manyDays() });
    const { container } = renderWithProviders(<BranchCard branch={branch} {...range} />);

    expect(container.querySelectorAll('[data-day]')).toHaveLength(7);
    expect(container.querySelector('.daily-axis')).toHaveTextContent('1 Nov');
    expect(container.querySelector('.daily-legend')).toHaveTextContent('Terlambat');
  });
});

function manyDays() {
  return Array.from({ length: 7 }, (_, i) => ({
    date: `2026-11-0${i + 1}`,
    total_packages: 100 + i * 10,
    delivered: 90 + i * 8,
    failed: 10 + i * 2,
    on_time: 70 + i * 6,
  }));
}
