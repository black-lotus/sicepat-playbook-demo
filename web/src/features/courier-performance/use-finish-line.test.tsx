import { render } from '@testing-library/react';
import { useRef } from 'react';

import { useFinishLine } from './use-finish-line';

function Harness({ keys }: { keys: string[] }) {
  const ref = useRef<HTMLUListElement>(null);
  useFinishLine(ref, keys);
  return (
    <ul ref={ref}>
      {keys.map((key) => (
        <li key={key} data-finish-key={key}>
          {key}
        </li>
      ))}
    </ul>
  );
}

let layoutOffset = 0;

function mockLayout() {
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockImplementation(function (this: HTMLElement) {
    const index = this.parentElement ? Array.from(this.parentElement.children).indexOf(this) : 0;
    const top = this.tagName === 'LI' ? index * 40 + layoutOffset : 0;
    return { top, left: 0, right: 100, bottom: top + 40, width: 100, height: 40, x: 0, y: top, toJSON: () => ({}) };
  });
}

function mockReducedMotion(reduce: boolean) {
  vi.stubGlobal(
    'matchMedia',
    vi.fn().mockImplementation((query: string) => ({
      matches: reduce && query === '(prefers-reduced-motion: reduce)',
      media: query,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    })),
  );
}

describe('useFinishLine', () => {
  let animate: ReturnType<typeof vi.fn>;
  let animations: { onfinish: (() => void) | null; cancel: ReturnType<typeof vi.fn> }[];

  const rows = (container: HTMLElement) => Array.from(container.querySelectorAll<HTMLElement>('li'));
  const isSlide = (call: unknown[]) => (call[1] as KeyframeAnimationOptions).duration === 520;
  const isEnter = (call: unknown[]) => !isSlide(call);
  const finishAll = () => {
    for (const animation of animations) animation.onfinish?.();
  };

  beforeEach(() => {
    layoutOffset = 0;
    animations = [];
    animate = vi.fn(() => {
      const animation = { onfinish: null as (() => void) | null, cancel: vi.fn() };
      animations.push(animation);
      return animation;
    });
    Object.defineProperty(HTMLElement.prototype, 'animate', { value: animate, configurable: true, writable: true });
    mockLayout();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
    Reflect.deleteProperty(HTMLElement.prototype, 'animate');
  });

  it('menganimasikan elemen dari posisi lama ke posisi baru saat urutan berubah', () => {
    mockReducedMotion(false);
    const { rerender } = render(<Harness keys={['C001', 'C003']} />);
    animate.mockClear();

    rerender(<Harness keys={['C003', 'C001']} />);

    expect(animate).toHaveBeenCalledTimes(2);
    const [frames, options] = animate.mock.calls[0] as [Keyframe[], KeyframeAnimationOptions];
    expect(frames[0]?.transform).toContain('translate');
    expect(frames[1]?.transform).toContain('translate(0');
    expect(options.duration).toBeGreaterThanOrEqual(450);
    expect(options.duration).toBeLessThanOrEqual(600);
  });

  it('baris baru tanpa posisi lama masuk dari offset kecil dengan opacity 0, bertahap per indeks', () => {
    mockReducedMotion(false);
    const { rerender } = render(<Harness keys={['C001']} />);
    animate.mockClear();

    rerender(<Harness keys={['C001', 'C005', 'C006']} />);

    const calls = animate.mock.calls;
    expect(calls).toHaveLength(2);
    expect(calls.every(isEnter)).toBe(true);
    const [frames, options] = calls[0] as [Keyframe[], KeyframeAnimationOptions];
    expect(frames[0]).toMatchObject({ transform: 'translateY(12px)', opacity: 0 });
    expect(frames[1]).toMatchObject({ transform: 'translateY(0px)', opacity: 1 });
    expect(options.fill).toBe('backwards');
    const delays = calls.map((call) => Number((call[1] as KeyframeAnimationOptions).delay));
    expect(delays[1]).toBeGreaterThan(delays[0] as number);
    for (const call of calls) {
      const o = call[1] as KeyframeAnimationOptions;
      expect(Number(o.delay) + Number(o.duration)).toBeLessThanOrEqual(300);
    }
  });

  it('data pertama yang tampil masuk dengan animasi, total waktu per baris tidak lebih dari 300 ms', () => {
    mockReducedMotion(false);
    const keys = Array.from({ length: 50 }, (_, i) => `C${i}`);

    render(<Harness keys={keys} />);

    expect(animate).toHaveBeenCalledTimes(50);
    for (const call of animate.mock.calls) {
      const o = call[1] as KeyframeAnimationOptions;
      expect(Number(o.delay) + Number(o.duration)).toBeLessThanOrEqual(300);
    }
  });

  it('data-settled diset true setelah animasi baris selesai', () => {
    mockReducedMotion(false);
    const { container } = render(<Harness keys={['C001', 'C003']} />);

    for (const row of rows(container)) expect(row).toHaveAttribute('data-settled', 'false');

    finishAll();

    for (const row of rows(container)) expect(row).toHaveAttribute('data-settled', 'true');
  });

  it('baris yang meluncur menjadi belum settled lagi sampai animasinya selesai', () => {
    mockReducedMotion(false);
    const { container, rerender } = render(<Harness keys={['C001', 'C003']} />);
    finishAll();

    rerender(<Harness keys={['C003', 'C001']} />);

    for (const row of rows(container)) expect(row).toHaveAttribute('data-settled', 'false');
    finishAll();
    for (const row of rows(container)) expect(row).toHaveAttribute('data-settled', 'true');
  });

  it('baris yang tidak bergerak langsung settled', () => {
    mockReducedMotion(false);
    const { container, rerender } = render(<Harness keys={['C001', 'C003']} />);
    finishAll();
    animate.mockClear();

    rerender(<Harness keys={['C001', 'C003', 'C009']} />);

    const [first, second, third] = rows(container);
    expect(first).toHaveAttribute('data-settled', 'true');
    expect(second).toHaveAttribute('data-settled', 'true');
    expect(third).toHaveAttribute('data-settled', 'false');
    expect(animate).toHaveBeenCalledTimes(1);
  });

  it('tidak menganimasikan bila urutan tidak berubah', () => {
    mockReducedMotion(false);
    const { rerender } = render(<Harness keys={['C001', 'C003']} />);
    animate.mockClear();

    rerender(<Harness keys={['C001', 'C003']} />);

    expect(animate).not.toHaveBeenCalled();
  });

  it('memakai posisi dari commit terakhir, bukan dari saat urutan terakhir berubah', () => {
    mockReducedMotion(false);
    const { rerender } = render(<Harness keys={['C001', 'C003']} />);
    layoutOffset = 100;
    rerender(<Harness keys={['C001', 'C003']} />);
    animate.mockClear();

    rerender(<Harness keys={['C003', 'C001']} />);

    const transforms = animate.mock.calls.filter(isSlide).map((call) => (call[0] as Keyframe[])[0]?.transform);
    expect(transforms).toContain('translate(0px, 40px)');
    expect(transforms).toContain('translate(0px, -40px)');
  });

  it('menghentikan animasi yang masih berjalan sebelum mengukur ulang', () => {
    mockReducedMotion(false);
    const { rerender } = render(<Harness keys={['C001']} />);
    const pending = animations[0];

    rerender(<Harness keys={['C003', 'C001']} />);

    expect(pending?.cancel).toHaveBeenCalled();
  });

  it('reduced motion: tanpa animasi dan data-settled langsung true', () => {
    mockReducedMotion(true);
    const { container, rerender } = render(<Harness keys={['C001', 'C003']} />);

    for (const row of rows(container)) expect(row).toHaveAttribute('data-settled', 'true');

    rerender(<Harness keys={['C003', 'C001', 'C007']} />);

    expect(animate).not.toHaveBeenCalled();
    for (const row of rows(container)) expect(row).toHaveAttribute('data-settled', 'true');
  });

  it('aman bila matchMedia dan animate tidak tersedia, dan langsung settled', () => {
    Reflect.deleteProperty(HTMLElement.prototype, 'animate');
    const { container, rerender } = render(<Harness keys={['C001', 'C003']} />);

    expect(() => rerender(<Harness keys={['C003', 'C001']} />)).not.toThrow();
    for (const row of rows(container)) expect(row).toHaveAttribute('data-settled', 'true');
  });
});
