import { render, screen } from '@testing-library/react';

import { toCourierPerformance } from '../api/courier-performance';
import { courierPerformanceDTOs } from '../test/fixtures';
import { DailyStrip } from './DailyStrip';

function courierByName(name: string) {
  const dto = courierPerformanceDTOs.find((item) => item.courier_name === name);
  if (!dto) throw new Error(`fixture ${name} tidak ada`);
  return toCourierPerformance(dto);
}

const budi = courierByName('Budi Santoso');
const sari = courierByName('Sari Wulandari');

function cell(container: HTMLElement, day: string): HTMLElement {
  const el = container.querySelector<HTMLElement>(`[data-day="${day}"]`);
  expect(el).not.toBeNull();
  return el as HTMLElement;
}

describe('DailyStrip', () => {
  it('merender satu sel per hari dengan label ringkasan', () => {
    const { container } = render(<DailyStrip daily={budi.daily} size="md" />);

    expect(container.querySelectorAll('[data-day]')).toHaveLength(7);
    expect(screen.getByRole('img', { name: 'Tren harian 7 hari: terkirim 36 dari 40' })).toBeInTheDocument();
  });

  it('menggambar bagian telat hanya bila ada paket terkirim yang telat', () => {
    const { container } = render(<DailyStrip daily={budi.daily} size="sm" />);

    const onTimeDay = cell(container, '2026-11-03');
    expect(onTimeDay.querySelector('[data-part="late"]')).toBeNull();
    expect(onTimeDay.querySelector('[data-part="on-time"]')).not.toBeNull();

    const lateDay = cell(container, '2026-11-04');
    expect(lateDay.querySelector('[data-part="late"]')).not.toBeNull();
    expect(lateDay.querySelector('[data-part="on-time"]')).not.toBeNull();
  });

  it('menandai hari tanpa paket sebagai kosong tanpa NaN', () => {
    const { container } = render(<DailyStrip daily={sari.daily} size="sm" />);

    expect(cell(container, '2026-11-03')).toHaveAttribute('data-empty', 'true');
    expect(cell(container, '2026-11-01')).not.toHaveAttribute('data-empty');
    for (const el of container.querySelectorAll('[style]')) {
      expect(el.getAttribute('style')).not.toMatch(/NaN/);
    }
  });

  it('mode dekoratif menyembunyikan strip dari aksesibilitas tanpa role img', () => {
    const { container } = render(<DailyStrip daily={budi.daily} size="md" decorative />);

    const root = container.querySelector<HTMLElement>('.daily-strip');
    expect(root).not.toBeNull();
    expect(root).toHaveAttribute('aria-hidden', 'true');
    expect(root).not.toHaveAttribute('role');
    expect(root).not.toHaveAttribute('aria-label');
    expect(screen.queryByRole('img')).not.toBeInTheDocument();
  });

  it('tidak merender apa pun untuk satu hari', () => {
    const { container } = render(<DailyStrip daily={budi.daily.slice(0, 1)} size="md" />);

    expect(container).toBeEmptyDOMElement();
  });

  it('memasang --days pada strip sm 31 hari dan menandainya padat', () => {
    const daily = Array.from({ length: 31 }, (_, i) => ({
      date: `2026-10-${String(i + 1).padStart(2, '0')}`,
      totalPackages: 5,
      delivered: 4,
      failed: 1,
      onTime: 3,
    }));
    const { container } = render(<DailyStrip daily={daily} size="sm" />);

    const root = container.querySelector<HTMLElement>('.daily-strip');
    expect(root?.style.getPropertyValue('--days')).toBe('31');
    expect(root).toHaveClass('daily-strip-dense');
  });

  it('strip sm 7 hari memasang --days tanpa kelas padat', () => {
    const { container } = render(<DailyStrip daily={budi.daily} size="sm" />);

    const root = container.querySelector<HTMLElement>('.daily-strip');
    expect(root?.style.getPropertyValue('--days')).toBe('7');
    expect(root).not.toHaveClass('daily-strip-dense');
  });
  it('menggambar dua garis target dengan label teks bila targets diberikan', () => {
    const { container } = render(<DailyStrip daily={budi.daily} size="md" targets={{ successRate: 0.9, slaRate: 0.85 }} />);

    const success = container.querySelector<HTMLElement>('[data-target="success"]');
    const sla = container.querySelector<HTMLElement>('[data-target="sla"]');
    expect(success).toHaveTextContent('Target keberhasilan 90%');
    expect(sla).toHaveTextContent('Target tepat SLA 85%');
    expect(success?.style.bottom).toBe('90%');
    expect(sla?.style.bottom).toBe('85%');
    expect(success).toHaveAttribute('data-line', 'solid');
    expect(sla).toHaveAttribute('data-line', 'dashed');
  });

  it('tidak menggambar garis target tanpa targets', () => {
    const { container } = render(<DailyStrip daily={budi.daily} size="md" />);

    expect(container.querySelector('[data-target]')).toBeNull();
  });
});
