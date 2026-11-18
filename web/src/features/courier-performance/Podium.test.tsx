import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { toCourierPerformance } from '../../api/courier-performance';
import { nestedInteractive } from '../../test/a11y';
import { courierPerformanceDTOs } from '../../test/fixtures';
import { Podium } from './Podium';

const items = courierPerformanceDTOs.map(toCourierPerformance);

function cardOf(name: RegExp): HTMLElement {
  const card = screen.getByRole('button', { name }).closest<HTMLElement>('.podium-card');
  if (!card) throw new Error('kartu podium tidak ada');
  return card;
}

describe('Podium', () => {
  it('peringkat 1 menjadi kartu hero dengan label, success rate, dan chip Tepat SLA', () => {
    render(<Podium items={items} onSelect={() => undefined} />);

    const hero = cardOf(/Budi Santoso/);
    expect(hero).toHaveClass('podium-hero');
    expect(hero).toHaveTextContent('Kurir terbaik periode ini');
    expect(hero).toHaveTextContent('90,0%');
    expect(hero).toHaveTextContent('Tepat SLA 91,7%');
    expect(hero).toHaveTextContent('40 paket');
  });

  it('peringkat 2 dan 3 menjadi kartu kecil dengan posisi dan tanpa label hero', () => {
    render(<Podium items={items} onSelect={() => undefined} />);

    const second = cardOf(/Dewi Lestari/);
    expect(second).not.toHaveClass('podium-hero');
    expect(second).toHaveTextContent('Pos 2');
    expect(second).not.toHaveTextContent('Kurir terbaik periode ini');
    expect(cardOf(/Agus Pratama/)).toHaveTextContent('Pos 3');
  });

  it('hanya merender tiga peringkat teratas dan memanggil onSelect dengan courier_id', async () => {
    const onSelect = vi.fn();
    render(<Podium items={items} onSelect={onSelect} />);

    const list = screen.getByRole('list', { name: 'Podium' });
    expect(within(list).getAllByRole('button')).toHaveLength(3);
    await userEvent.click(screen.getByRole('button', { name: /Dewi Lestari/ }));
    expect(onSelect).toHaveBeenCalledWith('C003');
  });

  it('kartu adalah kontainer biasa dengan satu tombol bernama kurir, bukan tombol pembungkus', () => {
    render(<Podium items={items} onSelect={() => undefined} />);

    for (const card of document.querySelectorAll<HTMLElement>('.podium-card')) {
      expect(card.tagName).toBe('DIV');
      expect(within(card).getAllByRole('button')).toHaveLength(1);
    }
    expect(screen.getByRole('button', { name: /Budi Santoso/ })).toHaveAccessibleName(/Budi Santoso/);
  });

  it('setiap kartu punya grafik harian bersegmen dan tidak ada elemen interaktif bersarang', () => {
    render(<Podium items={items} onSelect={() => undefined} />);

    const list = screen.getByRole('list', { name: 'Podium' });
    expect(within(list).getAllByRole('group', { name: /Tren harian/ })).toHaveLength(3);
    expect(nestedInteractive(list)).toEqual([]);
  });

  it('tombol nama tidak memuat grafik sehingga nama aksesibelnya bersih', () => {
    render(<Podium items={items} onSelect={() => undefined} />);

    expect(screen.getByRole('button', { name: /Budi Santoso/ })).not.toHaveAccessibleName(/Tren harian/);
  });
});
