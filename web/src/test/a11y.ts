const INTERACTIVE = 'button, a[href], input, select, textarea, summary, [tabindex], [role="button"], [role="link"]';
const NATIVE_CONTROL = 'button, a[href], summary, [role="button"], [role="link"]';

/**
 * Elemen interaktif di dalam kontrol native (tombol atau tautan) tidak valid dan tidak bisa difokus
 * dengan benar. Mengembalikan deskripsi setiap elemen interaktif yang bersarang di dalam kontrol lain.
 */
export function nestedInteractive(root: ParentNode): string[] {
  const found: string[] = [];
  for (const el of root.querySelectorAll<HTMLElement>(INTERACTIVE)) {
    const host = el.parentElement?.closest(NATIVE_CONTROL);
    if (host && root.contains(host)) {
      found.push(`<${el.tagName.toLowerCase()}> di dalam <${host.tagName.toLowerCase()}>`);
    }
  }
  return found;
}
