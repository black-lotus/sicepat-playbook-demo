import { formatPercent } from './format-percent';

describe('formatPercent', () => {
  it.each([
    [0, '0,0%'],
    [0.8, '80,0%'],
    [0.6666, '66,7%'],
    [1, '100,0%'],
  ])('%s menjadi %s', (input, expected) => {
    expect(formatPercent(input)).toBe(expected);
  });
});
