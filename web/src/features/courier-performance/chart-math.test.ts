import { compositionShares, niceScale } from './chart-math';

describe('niceScale', () => {
  it.each([
    [0, 1, 3],
    [1, 1, 3],
    [6, 2, 6],
    [36, 15, 45],
    [100, 40, 120],
    [1234, 500, 1500],
  ])('max %i memberi step %i dan batas atas %i', (max, step, top) => {
    expect(niceScale(max)).toEqual({ step, top });
    expect(niceScale(max).top).toBeGreaterThanOrEqual(max);
  });
});

describe('compositionShares', () => {
  it('memecah paket menjadi tepat SLA, terlambat, dan gagal', () => {
    const shares = compositionShares({ delivered: 36, onTime: 33, failed: 4 });
    expect(shares.map((s) => [s.key, s.value])).toEqual([['on-time', 33], ['late', 3], ['failed', 4]]);
  });

  it('jumlah nol menghasilkan share 0, bukan NaN', () => {
    for (const s of compositionShares({ delivered: 0, onTime: 0, failed: 0 })) expect(s.share).toBe(0);
  });
});
