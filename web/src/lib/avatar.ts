/** Inisial untuk avatar: huruf pertama dua kata pertama, atau dua huruf pertama bila hanya satu kata. */
export function initials(name: string): string {
  const words = name.trim().split(/\s+/).filter(Boolean);
  if (words.length === 0) return '?';
  const first = words[0] ?? '';
  const second = words[1];
  const letters = second ? `${first.charAt(0)}${second.charAt(0)}` : first.slice(0, 2);
  return letters.toUpperCase();
}

/** Hue 0..359 yang deterministik dari ID, supaya avatar kurir yang sama selalu berwarna sama. */
export function avatarHue(id: string): number {
  let hash = 0;
  for (const char of id) {
    hash = (hash * 31 + char.charCodeAt(0)) >>> 0;
  }
  return hash % 360;
}
