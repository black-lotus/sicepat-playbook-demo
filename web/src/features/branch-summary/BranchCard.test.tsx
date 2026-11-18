import { screen, within } from '@testing-library/react';

import { toBranchPerformance } from '../../api/branch-performance';
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

    const link = screen.getByRole('link');
    expect(link).toHaveTextContent('BDG02');
    expect(link).toHaveTextContent('1');
    expect(link).toHaveTextContent('83,3%');
    expect(link).toHaveTextContent('84,0%');
    expect(link).toHaveTextContent('833');
    expect(link).toHaveTextContent('8 kurir');

    const badges = within(link).getAllByText(/Tercapai|Di bawah target/);
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
});
