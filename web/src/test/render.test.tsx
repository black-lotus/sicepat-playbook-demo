import { screen } from '@testing-library/react';
import { useLocation } from 'react-router-dom';

import { renderWithProviders } from './render';

function Where() {
  const { pathname, search } = useLocation();
  return <p>{pathname + search}</p>;
}

describe('renderWithProviders', () => {
  it('memakai route awal dari opsi', () => {
    renderWithProviders(<Where />, {
      route: '/courier-performance?branch_id=JKT01',
    });
    expect(screen.getByText('/courier-performance?branch_id=JKT01')).toBeInTheDocument();
  });

  it('default ke /', () => {
    renderWithProviders(<Where />);
    expect(screen.getByText('/')).toBeInTheDocument();
  });
});
