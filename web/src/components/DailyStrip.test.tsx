import { fireEvent, render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { toCourierPerformance } from '../api/courier-performance';
import type { DailyCount } from '../api/types';
import { courierPerformanceDTOs } from '../test/fixtures';
import { nestedInteractive } from '../test/a11y';
import { DailyStrip } from './DailyStrip';

function courierByName(name: string) {
  const dto = courierPerformanceDTOs.find((item) => item.courier_name === name);
  if (!dto) throw new Error(`fixture ${name} tidak ada`);
  return toCourierPerformance(dto);
}

const budi = courierByName('Budi Santoso');
const sari = courierByName('Sari Wulandari');

function bar(container: HTMLElement, day: string): HTMLElement {
  const el = container.querySelector<HTMLElement>(`[data-day="${day}"]`);
  expect(el).not.toBeNull();
  return el as HTMLElement;
}

function stackHeight(el: HTMLElement): string {
  return el.querySelector<HTMLElement>('.daily-stack')?.style.height ?? '';
}

const uneven: DailyCount[] = [
  { date: '2026-11-10', totalPackages: 10, delivered: 8, failed: 2, onTime: 6 },
  { date: '2026-11-11', totalPackages: 20, delivered: 15, failed: 5, onTime: 10 },
  { date: '2026-11-12', totalPackages: 0, delivered: 0, failed: 0, onTime: 0 },
  { date: '2026-11-13', totalPackages: 5, delivered: 5, failed: 0, onTime: 5 },
];

describe('DailyStrip', () => {
  it('merender satu batang per hari dengan ringkasan teks yang aksesibel', () => {
    const { container } = render(<DailyStrip daily={budi.daily} size="md" />);

    expect(container.querySelectorAll('[data-day]')).toHaveLength(7);
    expect(
      screen.getByRole('group', { name: 'Tren harian 7 hari: terkirim 36 dari 40, tepat SLA 33, terlambat 3, gagal 4' }),
    ).toBeInTheDocument();
  });

  it('tinggi batang relatif terhadap hari terbanyak: terbanyak 100%, hari kosong tanpa tinggi', () => {
    const { container } = render(<DailyStrip daily={uneven} size="md" />);

    expect(stackHeight(bar(container, '2026-11-11'))).toBe('100%');
    expect(stackHeight(bar(container, '2026-11-10'))).toBe('50%');
    expect(stackHeight(bar(container, '2026-11-13'))).toBe('25%');
    expect(bar(container, '2026-11-12').querySelector('.daily-stack')).toBeNull();
    expect(bar(container, '2026-11-12')).toHaveAttribute('data-empty', 'true');
  });

  it('membagi batang menjadi tepat SLA, terlambat, dan gagal sesuai angka', () => {
    const { container } = render(<DailyStrip daily={uneven} size="md" />);

    const day = bar(container, '2026-11-11');
    const part = (name: string) => day.querySelector<HTMLElement>(`[data-part="${name}"]`);
    expect(part('on-time')).toHaveAttribute('data-count', '10');
    expect(part('late')).toHaveAttribute('data-count', '5');
    expect(part('failed')).toHaveAttribute('data-count', '5');
    expect(part('on-time')?.style.height).toBe('50%');
    expect(part('late')?.style.height).toBe('25%');
    expect(part('failed')?.style.height).toBe('25%');
  });

  it('tidak menggambar segmen yang bernilai nol dan tidak menghasilkan NaN', () => {
    const { container } = render(<DailyStrip daily={sari.daily} size="sm" />);

    const onlyFailed = bar(container, '2026-11-01');
    expect(onlyFailed.querySelector('[data-part="on-time"]')).toBeNull();
    expect(onlyFailed.querySelector('[data-part="late"]')).toBeNull();
    expect(onlyFailed.querySelector('[data-part="failed"]')).not.toBeNull();
    for (const el of container.querySelectorAll('[style]')) {
      expect(el.getAttribute('style')).not.toMatch(/NaN/);
    }
  });

  it('strip tanpa paket sama sekali tidak menghasilkan NaN', () => {
    const empty = uneven.map((d) => ({ ...d, totalPackages: 0, delivered: 0, failed: 0, onTime: 0 }));
    const { container } = render(<DailyStrip daily={empty} size="md" />);

    expect(container.querySelectorAll('[data-empty="true"]')).toHaveLength(4);
    expect(container.innerHTML).not.toMatch(/NaN/);
  });

  it('tidak merender apa pun untuk satu hari', () => {
    const { container } = render(<DailyStrip daily={budi.daily.slice(0, 1)} size="md" />);

    expect(container).toBeEmptyDOMElement();
  });

  describe('varian', () => {
    it('md memberi label tanggal di bawah batang dan legenda tiga segmen', () => {
      const { container } = render(<DailyStrip daily={budi.daily} size="md" />);

      const axis = container.querySelector('.daily-axis');
      expect(axis).toHaveTextContent('1 Nov');
      expect(axis).toHaveTextContent('7');
      const legend = container.querySelector<HTMLElement>('.daily-legend');
      expect(legend).toHaveTextContent('Tepat SLA');
      expect(legend).toHaveTextContent('Terlambat');
      expect(legend).toHaveTextContent('Gagal');
    });

    it('sm tetap tersegmen tetapi tanpa label dan legenda', () => {
      const { container } = render(<DailyStrip daily={budi.daily} size="sm" />);

      expect(container.querySelector('.daily-axis')).toBeNull();
      expect(container.querySelector('.daily-legend')).toBeNull();
      expect(container.querySelectorAll('[data-part="on-time"]').length).toBeGreaterThan(0);
    });

    it('strip md 31 hari menjarangkan label sumbu', () => {
      const daily = Array.from({ length: 31 }, (_, i) => ({
        date: `2026-10-${String(i + 1).padStart(2, '0')}`,
        totalPackages: 5,
        delivered: 4,
        failed: 1,
        onTime: 3,
      }));
      const { container } = render(<DailyStrip daily={daily} size="md" />);

      const labels = Array.from(container.querySelectorAll('.daily-axis > span')).filter((el) => el.textContent);
      expect(labels.length).toBeLessThanOrEqual(9);
      expect(labels.length).toBeGreaterThan(3);
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

      const root = container.querySelector<HTMLElement>('.daily-chart');
      expect(root?.style.getPropertyValue('--days')).toBe('31');
      expect(root).toHaveClass('daily-chart-dense');
    });

    it('strip sm 7 hari memasang --days tanpa kelas padat', () => {
      const { container } = render(<DailyStrip daily={budi.daily} size="sm" />);

      const root = container.querySelector<HTMLElement>('.daily-chart');
      expect(root?.style.getPropertyValue('--days')).toBe('7');
      expect(root).not.toHaveClass('daily-chart-dense');
    });
  });

  describe('tooltip', () => {
    it('tidak tampil sebelum ada interaksi', () => {
      render(<DailyStrip daily={budi.daily} size="md" />);

      expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
    });

    it('tampil saat hover dengan tanggal, total, rincian segmen, dan keberhasilan hari itu', async () => {
      const user = userEvent.setup();
      const { container } = render(<DailyStrip daily={budi.daily} size="md" />);

      await user.hover(bar(container, '2026-11-04'));

      const tip = screen.getByRole('tooltip');
      expect(tip).toHaveTextContent('Rab, 4 Nov');
      expect(within(tip).getByText('Total paket').nextSibling).toHaveTextContent('6');
      expect(within(tip).getByText('Tepat SLA').nextSibling).toHaveTextContent('4');
      expect(within(tip).getByText('Terlambat').nextSibling).toHaveTextContent('1');
      expect(within(tip).getByText('Gagal').nextSibling).toHaveTextContent('1');
      expect(within(tip).getByText('Keberhasilan').nextSibling).toHaveTextContent('83,3%');

      await user.unhover(bar(container, '2026-11-04'));
      expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
    });

    it('tampil saat batang difokus dengan keyboard dan tertutup saat blur atau Escape', async () => {
      const user = userEvent.setup();
      const { container } = render(<DailyStrip daily={budi.daily} size="md" />);

      await user.tab();
      expect(bar(container, '2026-11-01')).toHaveFocus();
      expect(screen.getByRole('tooltip')).toHaveTextContent('Min, 1 Nov');

      await user.keyboard('{Escape}');
      expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();

      await user.tab();
      expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
    });

    it('tampil lewat tap di layar sentuh dan menutup saat menyentuh di luar strip', () => {
      const { container } = render(
        <div>
          <DailyStrip daily={budi.daily} size="sm" />
          <p>di luar</p>
        </div>,
      );

      fireEvent.click(bar(container, '2026-11-03'));
      expect(screen.getByRole('tooltip')).toHaveTextContent('Sel, 3 Nov');

      fireEvent.pointerDown(screen.getByText('di luar'));
      expect(screen.queryByRole('tooltip')).not.toBeInTheDocument();
    });

    it('hari tanpa paket tetap punya keterangan tanpa persen keberhasilan', async () => {
      const user = userEvent.setup();
      const { container } = render(<DailyStrip daily={uneven} size="md" />);

      await user.hover(bar(container, '2026-11-12'));

      const tip = screen.getByRole('tooltip');
      expect(within(tip).getByText('Total paket').nextSibling).toHaveTextContent('0');
      expect(within(tip).getByText('Keberhasilan').nextSibling).toHaveTextContent('-');
      expect(tip).not.toHaveTextContent('NaN');
    });

    it('dirender di body supaya tidak terpotong tepi kartu dan dijepit di dalam layar', async () => {
      const user = userEvent.setup();
      const { container } = render(<DailyStrip daily={budi.daily} size="md" />);
      const first = bar(container, '2026-11-01');
      first.getBoundingClientRect = () => ({ left: 0, right: 20, top: 200, bottom: 240, width: 20, height: 40, x: 0, y: 200, toJSON: () => ({}) });

      await user.hover(first);

      const tip = screen.getByRole('tooltip');
      expect(container.contains(tip)).toBe(false);
      expect(document.body.contains(tip)).toBe(true);
      expect(Number.parseFloat(tip.style.left)).toBeGreaterThanOrEqual(8);
    });

    it('menambahkan target ke keterangan bila targets diberikan', async () => {
      const user = userEvent.setup();
      const { container } = render(<DailyStrip daily={budi.daily} size="sm" targets={{ successRate: 0.9, slaRate: 0.85 }} />);

      await user.hover(bar(container, '2026-11-01'));

      expect(screen.getByRole('tooltip')).toHaveTextContent('Target keberhasilan 90%');
      expect(screen.getByRole('tooltip')).toHaveTextContent('Target tepat SLA 85%');
    });
  });

  describe('keyboard', () => {
    it('hanya satu batang yang menjadi tab stop dan panah berpindah antar batang', async () => {
      const user = userEvent.setup();
      const { container } = render(<DailyStrip daily={budi.daily} size="md" />);

      const bars = Array.from(container.querySelectorAll<HTMLElement>('[data-day]'));
      expect(bars.filter((el) => el.tabIndex === 0)).toHaveLength(1);

      await user.tab();
      await user.keyboard('{ArrowRight}');
      expect(bars[1]).toHaveFocus();
      expect(screen.getByRole('tooltip')).toHaveTextContent('Sen, 2 Nov');
      await user.keyboard('{End}');
      expect(bars[6]).toHaveFocus();
      await user.keyboard('{ArrowLeft}');
      expect(bars[5]).toHaveFocus();
      await user.keyboard('{Home}');
      expect(bars[0]).toHaveFocus();
    });
  });

  describe('garis target', () => {
    const targets = { successRate: 0.9, slaRate: 0.8 };

    it('menggambar penanda keberhasilan dan tepat SLA di tiap batang berisi, relatif terhadap tinggi batang itu', () => {
      const { container } = render(<DailyStrip daily={uneven} size="md" targets={targets} />);

      const day = bar(container, '2026-11-11');
      const success = day.querySelector<HTMLElement>('[data-target="success"]');
      const sla = day.querySelector<HTMLElement>('[data-target="sla"]');
      expect(success?.style.bottom).toBe('90%');
      expect(sla?.style.bottom).toBe('60%');
      expect(success).toHaveAttribute('data-line', 'solid');
      expect(sla).toHaveAttribute('data-line', 'dashed');
      expect(bar(container, '2026-11-12').querySelector('[data-target]')).toBeNull();
    });

    it('menyebut target pada ringkasan teks', () => {
      render(<DailyStrip daily={uneven} size="md" targets={targets} />);

      const group = screen.getByRole('group');
      expect(group).toHaveAccessibleName(/Target keberhasilan 90%\. Target tepat SLA 80%/);
    });

    it('legenda md ikut menyebut target', () => {
      const { container } = render(<DailyStrip daily={uneven} size="md" targets={targets} />);

      expect(container.querySelector('.daily-legend')).toHaveTextContent('Target keberhasilan 90%');
    });

    it('tanpa targets tidak ada penanda', () => {
      const { container } = render(<DailyStrip daily={uneven} size="md" />);

      expect(container.querySelector('[data-target]')).toBeNull();
    });
  });

  it('tidak membuat elemen interaktif bersarang', () => {
    const { container } = render(<DailyStrip daily={budi.daily} size="md" targets={{ successRate: 0.9, slaRate: 0.8 }} />);

    expect(nestedInteractive(container)).toEqual([]);
  });
});
