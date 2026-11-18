import { render, screen, within } from '@testing-library/react';

import { DonutChart } from './DonutChart';
import type { DonutSegment } from './DonutChart';

function segments(onTime: number, late: number, failed: number): DonutSegment[] {
  const sum = onTime + late + failed;
  const share = (value: number) => (sum > 0 ? value / sum : 0);
  return [
    { key: 'on-time', label: 'Tepat SLA', value: onTime, share: share(onTime), pattern: 'solid' },
    { key: 'late', label: 'Terlambat', value: late, share: share(late), pattern: 'diagonal' },
    { key: 'failed', label: 'Gagal', value: failed, share: share(failed), pattern: 'dots' },
  ];
}

describe('DonutChart', () => {
  it('menampilkan total dan label di tengah serta persen per segmen', () => {
    // Budi Santoso: 40 paket, 33 tepat SLA, 3 telat, 4 gagal => 82,5%, 7,5%, 10,0%.
    const { container } = render(<DonutChart total={40} label="Total paket" segments={segments(33, 3, 4)} />);

    expect(container.querySelector('.donut-total')).toHaveTextContent('40');
    expect(container.querySelector('.donut-total-label')).toHaveTextContent('Total paket');
    const legend = container.querySelector('.donut-legend') as HTMLElement;
    expect(within(legend).getByText('Tepat SLA').closest('li')).toHaveTextContent('82,5%');
    expect(within(legend).getByText('Terlambat').closest('li')).toHaveTextContent('7,5%');
    expect(within(legend).getByText('Gagal').closest('li')).toHaveTextContent('10,0%');
    expect(screen.getByRole('img', { name: 'Total paket 40: Tepat SLA 82,5%, Terlambat 7,5%, Gagal 10,0%' })).toBeInTheDocument();
  });

  it('proporsi arc sesuai dan berjumlah penuh', () => {
    const { container } = render(<DonutChart total={40} label="Total paket" segments={segments(33, 3, 4)} />);

    const shares = Array.from(container.querySelectorAll('circle.donut-arc')).map((node) => Number(node.getAttribute('data-share')));
    expect(shares).toHaveLength(3);
    expect(shares[0]).toBeCloseTo(33 / 40);
    expect(shares[1]).toBeCloseTo(3 / 40);
    expect(shares[2]).toBeCloseTo(4 / 40);
    expect(shares.reduce((a, b) => a + b, 0)).toBeCloseTo(1);
  });

  it('segmen bernilai 0 tidak digambar dan persennya 0,0% tanpa NaN', () => {
    const { container } = render(<DonutChart total={35} label="Total paket" segments={segments(30, 0, 5)} />);

    expect(container.querySelector('circle[data-segment="late"]')).toBeNull();
    expect(container.querySelector('li[data-segment="late"]')).toHaveTextContent('0,0%');
    expect(container.innerHTML).not.toContain('NaN');
  });

  it('semua nol: hanya lintasan kosong, total 0, tidak ada NaN', () => {
    const { container } = render(<DonutChart total={0} label="Total paket" segments={segments(0, 0, 0)} />);

    expect(container.querySelectorAll('circle.donut-arc')).toHaveLength(0);
    expect(container.querySelector('.donut-total')).toHaveTextContent('0');
    expect(container.querySelectorAll('.legend-percent')).toHaveLength(3);
    for (const node of container.querySelectorAll('.legend-percent')) expect(node).toHaveTextContent('0,0%');
    expect(container.innerHTML).not.toContain('NaN');
  });

  it('satu segmen penuh menggambar satu arc tanpa celah', () => {
    const { container } = render(<DonutChart total={5} label="Total paket" segments={segments(5, 0, 0)} />);

    const arcs = container.querySelectorAll('circle.donut-arc');
    expect(arcs).toHaveLength(1);
    expect(arcs[0]).toHaveAttribute('stroke-dasharray', '100 0');
  });

  it('pola default solid dan legenda memakai penanda sesuai pola', () => {
    const { container } = render(
      <DonutChart
        total={10}
        label="Total gagal"
        segments={[
          { key: 'a', label: 'Alasan A', value: 6, share: 0.6 },
          { key: 'b', label: 'Alasan B', value: 4, share: 0.4, pattern: 'dots' },
        ]}
      />,
    );

    expect(container.querySelector('li[data-segment="a"] .legend-mark')).toHaveClass('legend-mark-solid');
    expect(container.querySelector('li[data-segment="b"] .legend-mark')).toHaveClass('legend-mark-dots');
    expect(container.querySelector('.donut-total-label')).toHaveTextContent('Total gagal');
  });
});
