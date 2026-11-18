import { formatGap } from './format-gap';

describe('formatGap', () => {
  it.each([
    [0.04285714, '−4,3 pp'],
    [0.9, '−90,0 pp'],
    [0, '0,0 pp'],
    [0.0004, '0,0 pp'],
    [0.00049, '0,0 pp'],
    [0.0005, '−0,1 pp'],
    [-0.0004, '0,0 pp'],
  ])('%s menjadi %s', (input, expected) => {
    expect(formatGap(input)).toBe(expected);
  });
});
