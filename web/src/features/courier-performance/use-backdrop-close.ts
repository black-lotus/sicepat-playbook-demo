import { useRef } from 'react';
import type { MouseEvent, PointerEvent } from 'react';

/**
 * Props untuk backdrop dialog. Dialog hanya ditutup bila pointerdown dan click sama-sama
 * bertarget backdrop, sehingga drag-select yang dimulai di panel dan berakhir di backdrop tidak menutupnya.
 */
export function useBackdropClose(onClose: () => void) {
  const pressedOnBackdrop = useRef(false);

  return {
    onPointerDown: (event: PointerEvent<HTMLElement>) => {
      pressedOnBackdrop.current = event.target === event.currentTarget;
    },
    onClick: (event: MouseEvent<HTMLElement>) => {
      const shouldClose = pressedOnBackdrop.current && event.target === event.currentTarget;
      pressedOnBackdrop.current = false;
      if (shouldClose) onClose();
    },
  };
}
