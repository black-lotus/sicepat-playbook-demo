import { act, renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { MemoryRouter, useLocation, useNavigate } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { useCourierPerformanceFilter } from './use-courier-performance-filter';

const today = new Date(2026, 10, 7);

function setup(route: string) {
  const wrapper = ({ children }: { children: ReactNode }) => (
    <MemoryRouter initialEntries={[route]}>{children}</MemoryRouter>
  );
  return renderHook(
    () => ({
      hook: useCourierPerformanceFilter(today),
      location: useLocation(),
      navigate: useNavigate(),
    }),
    { wrapper },
  );
}

describe('useCourierPerformanceFilter', () => {
  it('tanpa query memakai default 7 hari dan tidak menulis URL', () => {
    const { result } = setup('/courier-performance');
    expect(result.current.hook.filter).toMatchObject({
      startDate: '2026-11-01',
      endDate: '2026-11-07',
      branchID: '',
      page: 1,
    });
    expect(result.current.location.search).toBe('');
  });

  it('membaca query dari URL', () => {
    const { result } = setup('/x?branch_id=JKT01&start_date=2026-10-01&end_date=2026-10-10&page=3');
    expect(result.current.hook.filter).toMatchObject({
      branchID: 'JKT01',
      startDate: '2026-10-01',
      endDate: '2026-10-10',
      page: 3,
    });
  });

  it('nilai tidak valid jatuh ke default per parameter', () => {
    const bad = setup('/x?start_date=abc&end_date=2026-11-05&page=0');
    expect(bad.result.current.hook.filter).toMatchObject({
      startDate: '2026-11-01',
      endDate: '2026-11-05',
      page: 1,
    });
    const long = setup('/x?start_date=2026-01-01&end_date=2026-11-07');
    expect(long.result.current.hook.filter).toMatchObject({
      startDate: '2026-11-01',
      endDate: '2026-11-07',
    });
    const badDate = setup('/x?start_date=2026-02-31&end_date=2026-11-07&page=x');
    expect(badDate.result.current.hook.filter.startDate).toBe('2026-11-01');
    expect(badDate.result.current.hook.filter.page).toBe(1);
  });

  it('setter menulis query dan ubah periode atau cabang mengembalikan page ke 1', () => {
    const { result } = setup('/x?page=4');
    act(() => result.current.hook.setPeriod({ startDate: '2026-11-02', endDate: '2026-11-04' }));
    const q = new URLSearchParams(result.current.location.search);
    expect(q.get('start_date')).toBe('2026-11-02');
    expect(q.get('end_date')).toBe('2026-11-04');
    expect(q.get('page')).toBeNull();
    act(() => result.current.hook.setPage(2));
    expect(result.current.hook.filter.page).toBe(2);
    act(() => result.current.hook.setBranch('JKT01'));
    expect(result.current.hook.filter.branchID).toBe('JKT01');
    expect(result.current.hook.filter.page).toBe(1);
  });

  it('back mengembalikan filter sebelumnya', () => {
    const { result } = setup('/x');
    act(() => result.current.hook.setBranch('JKT01'));
    act(() => result.current.hook.setBranch('JKT02'));
    expect(result.current.hook.filter.branchID).toBe('JKT02');
    act(() => result.current.navigate(-1));
    expect(result.current.hook.filter.branchID).toBe('JKT01');
    act(() => result.current.navigate(-1));
    expect(result.current.hook.filter.branchID).toBe('');
  });
});
