import { render, screen, within } from '@testing-library/react';

import { DonutChart } from './DonutChart';

describe('DonutChart', () => {
  it('menampilkan total di tengah dan persen per segmen dari angka API', () => {
    // Budi Santoso: 40 paket, 36 terkirim, 33 tepat SLA, 4 gagal => tepat 82,5%, telat 7,5%, gagal 10,0%.
    const { container } = render(<DonutChart total={40} delivered={36} onTime={33} failed={4} />);

    expect(container.querySelector('.donut-total')).toHaveTextContent('40');
    const legend = container.querySelector('.donut-legend') as HTMLElement;
    expect(within(legend).getByText('Tepat SLA').closest('li')).toHaveTextContent('82,5%');
    expect(within(legend).getByText('Terlambat').closest('li')).toHaveTextContent('7,5%');
    expect(within(legend).getByText('Gagal').closest('li')).toHaveTextContent('10,0%');
    expect(screen.getByRole('img', { name: 'Komposisi 40 paket: Tepat SLA 82,5%, Terlambat 7,5%, Gagal 10,0%' })).toBeInTheDocument();
  });

  it('proporsi arc sesuai dan berjumlah penuh', () => {
    const { container } = render(<DonutChart total={40} delivered={36} onTime={33} failed={4} />);

    const shares = Array.from(container.querySelectorAll('circle.donut-arc')).map((node) => Number(node.getAttribute('data-share')));
    expect(shares).toHaveLength(3);
    expect(shares[0]).toBeCloseTo(33 / 40);
    expect(shares[1]).toBeCloseTo(3 / 40);
    expect(shares[2]).toBeCloseTo(4 / 40);
    expect(shares.reduce((a, b) => a + b, 0)).toBeCloseTo(1);
  });

  it('segmen bernilai 0 tidak digambar dan persennya 0,0% tanpa NaN', () => {
    // Dewi-like: tidak ada yang telat (on_time = delivered).
    const { container } = render(<DonutChart total={35} delivered={30} onTime={30} failed={5} />);

    expect(container.querySelector('circle[data-segment="late"]')).toBeNull();
    expect(container.querySelector('li[data-segment="late"]')).toHaveTextContent('0,0%');
    expect(container.innerHTML).not.toContain('NaN');
  });

  it('semua nol: hanya lintasan kosong, total 0, tidak ada NaN', () => {
    const { container } = render(<DonutChart total={0} delivered={0} onTime={0} failed={0} />);

    expect(container.querySelectorAll('circle.donut-arc')).toHaveLength(0);
    expect(container.querySelector('.donut-total')).toHaveTextContent('0');
    expect(container.querySelectorAll('.legend-percent')).toHaveLength(3);
    for (const node of container.querySelectorAll('.legend-percent')) expect(node).toHaveTextContent('0,0%');
    expect(container.innerHTML).not.toContain('NaN');
  });

  it('on_time lebih besar dari delivered dibatasi ke delivered sehingga telat tidak negatif', () => {
    const { container } = render(<DonutChart total={10} delivered={8} onTime={9} failed={2} />);

    expect(container.querySelector('li[data-segment="late"]')).toHaveTextContent('0,0%');
    expect(container.querySelector('li[data-segment="on-time"]')).toHaveTextContent('80,0%');
  });

  it('satu segmen penuh menggambar satu arc tanpa celah', () => {
    const { container } = render(<DonutChart total={5} delivered={5} onTime={5} failed={0} />);

    const arcs = container.querySelectorAll('circle.donut-arc');
    expect(arcs).toHaveLength(1);
    expect(arcs[0]).toHaveAttribute('stroke-dasharray', '100 0');
  });
});
