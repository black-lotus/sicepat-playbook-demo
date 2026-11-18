// @vitest-environment node
/// <reference types="node" />
import { readFileSync } from 'node:fs';

// Vitest mengosongkan hasil impor `?raw` untuk file CSS, jadi stylesheet dibaca langsung dari disk.
const featureCSS = readFileSync(new URL('./courier-performance.css', import.meta.url), 'utf8');
const appCSS = readFileSync(new URL('../../app/app.css', import.meta.url), 'utf8');

describe('stylesheet performa kurir', () => {
  it('backdrop memakai token --color-scrim, bukan warna mentah', () => {
    expect(featureCSS).not.toMatch(/rgb\(\s*0 0 0/);
    expect(featureCSS).toContain('var(--color-scrim)');
  });

  it('--color-scrim didefinisikan untuk tema terang dan gelap', () => {
    expect(appCSS.match(/--color-scrim:/g)).toHaveLength(2);
  });

  it('tidak ada selector global section atau .table-wrap tanpa prefix halaman', () => {
    expect(featureCSS).not.toMatch(/^\s*section\s*\{/m);
    expect(featureCSS).not.toMatch(/^\s*\.table-wrap\s*\{/m);
    expect(featureCSS).toContain('.courier-performance-page .table-wrap');
  });

  it('Selisih memudar masuk berdasarkan data-settled, bukan animasi saat mount', () => {
    expect(featureCSS).toContain('[data-settled="true"] > .gap-cell');
    expect(featureCSS).not.toMatch(/\.gap-cell\s*\{[^}]*animation/);
    expect(featureCSS).not.toContain('@keyframes gap-in');
  });
});
