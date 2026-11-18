import { render, screen } from '@testing-library/react';

import { StatusBadge } from './StatusBadge';

describe('StatusBadge', () => {
  it.each([
    ['met', 'Tercapai'],
    ['below', 'Di bawah target'],
    ['above', 'Di atas target'],
  ] as const)('status %s menampilkan ikon svg aria-hidden dan teks terbaca', (status, label) => {
    const { container } = render(<StatusBadge status={status} label={label} />);

    expect(screen.getByText(label)).toBeInTheDocument();
    const svg = container.querySelector('svg');
    expect(svg).not.toBeNull();
    expect(svg).toHaveAttribute('aria-hidden', 'true');
    expect(container.firstElementChild).toHaveAttribute('data-status', status);
  });
});
