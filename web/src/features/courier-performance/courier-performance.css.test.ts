// @vitest-environment node
/// <reference types="node" />
import { readFileSync } from 'node:fs';

// Vitest mengosongkan hasil impor `?raw` untuk file CSS, jadi stylesheet dibaca langsung dari disk.
const featureCSS = readFileSync(new URL('./courier-performance.css', import.meta.url), 'utf8');

describe('stylesheet performa kurir', () => {
  it('tidak ada selector global section atau .table-wrap tanpa prefix halaman', () => {
    expect(featureCSS).not.toMatch(/^\s*section\s*\{/m);
    expect(featureCSS).not.toMatch(/^\s*\.table-wrap\s*\{/m);
    expect(featureCSS).toContain('.courier-performance-page .table-wrap');
  });
});
