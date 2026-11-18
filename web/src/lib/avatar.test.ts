import { avatarHue, initials } from './avatar';

describe('initials', () => {
  it('memakai huruf pertama dua kata pertama dalam huruf kapital', () => {
    expect(initials('Budi Santoso')).toBe('BS');
    expect(initials('dewi lestari wulandari')).toBe('DL');
  });

  it('memakai dua huruf pertama bila hanya satu kata', () => {
    expect(initials('Agus')).toBe('AG');
    expect(initials('C001')).toBe('C0');
  });

  it('mengembalikan tanda tanya untuk nama kosong', () => {
    expect(initials('   ')).toBe('?');
  });
});

describe('avatarHue', () => {
  it('deterministik dan berada di rentang 0 sampai 359', () => {
    expect(avatarHue('C001')).toBe(avatarHue('C001'));
    for (const id of ['C001', 'C002', 'C003', 'JKT03-0042', '']) {
      const hue = avatarHue(id);
      expect(hue).toBeGreaterThanOrEqual(0);
      expect(hue).toBeLessThan(360);
    }
  });

  it('ID berbeda umumnya menghasilkan hue berbeda', () => {
    expect(avatarHue('C001')).not.toBe(avatarHue('C002'));
  });
});
