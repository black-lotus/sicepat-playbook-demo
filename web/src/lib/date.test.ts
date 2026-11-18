import { formatDate, formatDateRange, formatDayMonth, lastSevenDays, toISODate } from './date';

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

describe('formatDateRange', () => {
  it('meringkas rentang dalam satu bulan', () => {
    expect(formatDateRange('2026-11-01', '2026-11-07')).toBe('1 sampai 7 Nov');
  });

  it('menyebut kedua bulan bila rentang melewati batas bulan', () => {
    expect(formatDateRange('2026-10-28', '2026-11-03')).toBe('28 Okt sampai 3 Nov');
  });

  it('menampilkan satu tanggal bila periode satu hari', () => {
    expect(formatDateRange('2026-11-05', '2026-11-05')).toBe('5 Nov');
  });

  it.each([
    ['', '2026-11-18'],
    ['2026-11-18', ''],
    ['', ''],
    ['2026-13-45', '2026-11-18'],
    ['bukan-tanggal', '2026-11-18'],
  ])('mengembalikan "Pilih periode" bila tanggal tidak valid (%j, %j)', (start, end) => {
    expect(formatDateRange(start, end)).toBe('Pilih periode');
  });
});

describe('formatDayMonth', () => {
  it('menampilkan tanggal dan bulan pendek', () => {
    expect(formatDayMonth('2026-11-12')).toBe('12 Nov');
    expect(formatDayMonth('2026-10-01')).toBe('1 Okt');
  });
});
