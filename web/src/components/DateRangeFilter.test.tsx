import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import { DateRangeFilter } from './DateRangeFilter';

describe('DateRangeFilter', () => {
  it('varian inline langsung menampilkan input Dari dan Sampai', () => {
    render(<DateRangeFilter startDate="2026-11-12" endDate="2026-11-18" onChange={() => undefined} />);

    expect(screen.getByLabelText('Dari')).toHaveValue('2026-11-12');
    expect(screen.getByLabelText('Sampai')).toHaveValue('2026-11-18');
    expect(document.querySelector('details')).toBeNull();
  });

  it('varian pil meringkas periode dan tetap punya input yang berlabel', () => {
    const onChange = vi.fn();
    render(<DateRangeFilter variant="pill" startDate="2026-11-12" endDate="2026-11-18" onChange={onChange} />);

    expect(screen.getByText('12 sampai 18 Nov')).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText('Dari'), { target: { value: '2026-11-10' } });
    expect(onChange).toHaveBeenCalledWith({ startDate: '2026-11-10', endDate: '2026-11-18' });
  });

  it('varian pil menutup popover dengan Esc dan klik di luar', async () => {
    const user = userEvent.setup();
    render(<DateRangeFilter variant="pill" startDate="2026-11-12" endDate="2026-11-18" onChange={() => undefined} />);
    const details = document.querySelector('details') as HTMLDetailsElement;

    await user.click(screen.getByText('12 sampai 18 Nov'));
    expect(details.open).toBe(true);
    await user.keyboard('{Escape}');
    expect(details.open).toBe(false);

    await user.click(screen.getByText('12 sampai 18 Nov'));
    expect(details.open).toBe(true);
    await user.click(document.body);
    expect(details.open).toBe(false);
  });
});
