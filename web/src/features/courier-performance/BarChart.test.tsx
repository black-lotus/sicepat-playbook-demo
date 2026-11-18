import { render } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { toCourierPerformance } from '../../api/courier-performance';
import type { DailyCount } from '../../api/types';
import { courierPerformanceDTOs } from '../../test/fixtures';
import { BarChart } from './BarChart';

const budiDTO = courierPerformanceDTOs.find((dto) => dto.courier_name === 'Budi Santoso');
if (!budiDTO) throw new Error('fixture Budi Santoso tidak ada');
const budi = toCourierPerformance(budiDTO);

function day(date: string, delivered: number): DailyCount {
  return { date, totalPackages: delivered, delivered, failed: 0, onTime: delivered };
}

describe('BarChart', () => {
  it('merender satu batang per hari', () => {
    const { container } = render(<BarChart daily={budi.daily} />);

    expect(container.querySelectorAll('rect.bar[data-day]')).toHaveLength(budi.daily.length);
  });

  it('batang dengan nilai tertinggi tersorot dan memegang tooltip angka serta tanggal', () => {
    const { container } = render(<BarChart daily={budi.daily} />);

    // Budi: terkirim terbanyak 6 pada 1 Nov (pertama bila seri).
    const peak = container.querySelector('rect.bar[data-peak="true"]');
    expect(peak).toHaveAttribute('data-day', '2026-11-01');
    expect(peak).toHaveAttribute('data-active', 'true');
    expect(container.querySelectorAll('rect.bar[data-active="true"]')).toHaveLength(1);
    const tooltip = container.querySelector('.bar-tooltip');
    expect(tooltip).toHaveTextContent('1 Nov 2026');
    expect(tooltip).toHaveTextContent('6 terkirim');
  });

  it('sumbu Y punya empat garis bantu bilangan bulat termasuk nol dan mencakup nilai tertinggi', () => {
    const { container } = render(<BarChart daily={[day('2026-11-01', 36), day('2026-11-02', 20)]} />);

    const ticks = Array.from(container.querySelectorAll('text.bar-tick'))
      .map((node) => node.textContent ?? '')
      .filter((text) => /^\d+$/.test(text));
    expect(ticks).toEqual(['0', '15', '30', '45']);
  });

  it('semua hari bernilai nol: tidak ada batang tersorot, tidak ada tooltip, tidak ada NaN', () => {
    const { container } = render(<BarChart daily={[day('2026-11-01', 0), day('2026-11-02', 0)]} />);

    expect(container.querySelector('[data-peak]')).toBeNull();
    expect(container.querySelector('.bar-tooltip')).toBeNull();
    expect(container.innerHTML).not.toContain('NaN');
  });

  it('periode panjang: batang tetap satu per hari dan label sumbu X dijarangkan', () => {
    const daily = Array.from({ length: 31 }, (_, i) => day(`2026-10-${String(i + 1).padStart(2, '0')}`, (i % 7) + 1));
    const { container } = render(<BarChart daily={daily} />);

    expect(container.querySelectorAll('rect.bar[data-day]')).toHaveLength(31);
    const dateLabels = Array.from(container.querySelectorAll('text.bar-tick')).filter((node) => /Okt/.test(node.textContent ?? ''));
    expect(dateLabels.length).toBeLessThanOrEqual(8);
    expect(dateLabels.length).toBeGreaterThan(2);
  });

  it('hover pada hari lain memindahkan tooltip dan kembali ke puncak saat pointer keluar', async () => {
    const user = userEvent.setup();
    const { container } = render(<BarChart daily={budi.daily} />);
    const hits = container.querySelectorAll('rect.bar-hit');

    await user.hover(hits[1] as Element);
    expect(container.querySelector('.bar-tooltip')).toHaveAttribute('data-day', '2026-11-02');

    await user.unhover(hits[1] as Element);
    expect(container.querySelector('.bar-tooltip')).toHaveAttribute('data-day', '2026-11-01');
  });
});
