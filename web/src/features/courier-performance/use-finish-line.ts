import { useLayoutEffect, useRef } from 'react';
import type { RefObject } from 'react';

interface Position {
  top: number;
  left: number;
}

const SLIDE_DURATION_MS = 520;
const SLIDE_EASING = 'cubic-bezier(0.22, 1, 0.36, 1)';
const ENTER_OFFSET_PX = 12;
const ENTER_DURATION_MS = 220;
const ENTER_STAGGER_MS = 20;
const ENTER_MAX_DELAY_MS = 80;

function prefersReducedMotion(): boolean {
  return typeof window.matchMedia === 'function' && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
}

function settle(node: HTMLElement, settled: boolean) {
  node.dataset.settled = settled ? 'true' : 'false';
}

/**
 * Gerak "garis finis". Saat urutan `keys` berubah, setiap elemen `[data-finish-key]`:
 * - yang sudah punya posisi lama meluncur (FLIP) dari posisi lamanya ke posisi barunya;
 * - yang baru (tanpa posisi lama) masuk dari offset kecil dengan opacity 0, bertahap per indeks.
 *
 * Elemen diberi `data-settled="false"` selama beranimasi dan `data-settled="true"` setelah animasinya
 * selesai, atau langsung bila tidak ada animasi (reduced motion, atau `animate` tidak tersedia).
 * CSS kolom Selisih memudar masuk berdasarkan `[data-settled="true"]`.
 *
 * Posisi direkam di setiap commit supaya tidak basi (mis. setelah resize), tetapi animasi hanya
 * dijalankan saat urutan `keys` berubah. Posisi diukur relatif terhadap kontainer sehingga scroll
 * halaman tidak dihitung sebagai perpindahan.
 */
export function useFinishLine(container: RefObject<HTMLElement | null>, keys: string[]): void {
  const positions = useRef<Map<string, Position>>(new Map());
  const lastKeys = useRef<string | null>(null);
  const running = useRef<Map<HTMLElement, Animation>>(new Map());
  const keyString = keys.join('|');

  // Tanpa deps: posisi harus direkam pada setiap commit.
  useLayoutEffect(() => {
    const root = container.current;
    if (!root) return;

    // Animasi yang masih berjalan membuat hasil ukur ikut tergeser, jadi hentikan dulu dan anggap selesai.
    for (const [node, animation] of running.current) {
      animation.onfinish = null;
      animation.cancel?.();
      settle(node, true);
    }
    running.current.clear();

    const origin = root.getBoundingClientRect();
    const nodes = Array.from(root.querySelectorAll<HTMLElement>('[data-finish-key]'));
    const next = new Map<string, Position>();
    for (const node of nodes) {
      const rect = node.getBoundingClientRect();
      next.set(node.dataset.finishKey ?? '', { top: rect.top - origin.top, left: rect.left - origin.left });
    }

    const before = positions.current;
    positions.current = next;
    const orderChanged = keyString !== lastKeys.current;
    lastKeys.current = keyString;

    const canAnimate = !prefersReducedMotion();
    nodes.forEach((node, index) => {
      const key = node.dataset.finishKey ?? '';
      const from = before.get(key);
      const to = next.get(key);
      const animatable = orderChanged && canAnimate && to !== undefined && typeof node.animate === 'function';

      let animation: Animation | undefined;
      if (animatable && !from) {
        animation = node.animate(
          [
            { transform: `translateY(${ENTER_OFFSET_PX}px)`, opacity: 0 },
            { transform: 'translateY(0px)', opacity: 1 },
          ],
          {
            duration: ENTER_DURATION_MS,
            delay: Math.min(index * ENTER_STAGGER_MS, ENTER_MAX_DELAY_MS),
            easing: SLIDE_EASING,
            fill: 'backwards',
          },
        );
      } else if (animatable && from && to && (from.left !== to.left || from.top !== to.top)) {
        animation = node.animate(
          [{ transform: `translate(${from.left - to.left}px, ${from.top - to.top}px)` }, { transform: 'translate(0px, 0px)' }],
          { duration: SLIDE_DURATION_MS, easing: SLIDE_EASING },
        );
      }

      if (!animation) {
        settle(node, true);
        return;
      }
      settle(node, false);
      running.current.set(node, animation);
      animation.onfinish = () => {
        running.current.delete(node);
        settle(node, true);
      };
    });
  });
}
