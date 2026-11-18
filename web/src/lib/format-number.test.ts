import { formatNumber } from './format-number';

describe('formatNumber', () => {
  it.each([
    [0, '0'],
    [999, '999'],
    [12500, '12.500'],
    [1250000, '1.250.000'],
  ])('%s menjadi %s', (input, expected) => {
    expect(formatNumber(input)).toBe(expected);
  });
});
