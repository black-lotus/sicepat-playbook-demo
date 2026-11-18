import { formatDate, lastSevenDays, toISODate } from './date';

describe('toISODate', () => {
  it('menambahkan nol di depan bulan dan tanggal', () => {
    expect(toISODate(new Date(2026, 0, 5))).toBe('2026-01-05');
  });
});

describe('lastSevenDays', () => {
  it('mencakup hari ini dan 6 hari sebelumnya', () => {
    expect(lastSevenDays(new Date(2026, 10, 18))).toEqual({ startDate: '2026-11-12', endDate: '2026-11-18' });
  });

  it('melewati batas bulan', () => {
    expect(lastSevenDays(new Date(2026, 11, 3))).toEqual({ startDate: '2026-11-27', endDate: '2026-12-03' });
  });
});

describe('formatDate', () => {
  it('memformat tanggal API ke format Indonesia', () => {
    expect(formatDate('2026-11-17')).toBe('17 Nov 2026');
  });
});
