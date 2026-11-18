import { useEffect, useRef } from 'react';
import type { ReactNode } from 'react';

import { formatDateRange } from '../lib/date';
import { CalendarIcon, ChevronDownIcon } from './icons';

interface DateRangeFilterProps {
  startDate: string;
  endDate: string;
  onChange: (range: { startDate: string; endDate: string }) => void;
  /** `pill` menampilkan ringkasan periode sebagai pil yang membuka input tanggal; `inline` menampilkan input langsung. */
  variant?: 'inline' | 'pill';
}

export function DateRangeFilter({ startDate, endDate, onChange, variant = 'inline' }: DateRangeFilterProps) {
  const fields = (
    <fieldset className="date-range">
      <legend className="visually-hidden">Periode</legend>
      <label>
        Dari
        <input
          type="date"
          value={startDate}
          max={endDate}
          onChange={(e) => onChange({ startDate: e.target.value, endDate })}
        />
      </label>
      <label>
        Sampai
        <input
          type="date"
          value={endDate}
          min={startDate}
          onChange={(e) => onChange({ startDate, endDate: e.target.value })}
        />
      </label>
    </fieldset>
  );

  if (variant === 'inline') return fields;
  return <PeriodPill summary={formatDateRange(startDate, endDate)}>{fields}</PeriodPill>;
}

function PeriodPill({ summary, children }: { summary: string; children: ReactNode }) {
  const ref = useRef<HTMLDetailsElement>(null);

  // Menutup popover saat klik di luar atau menekan Esc, supaya tidak menutupi hasil.
  useEffect(() => {
    const close = () => {
      if (ref.current) ref.current.open = false;
    };
    const onPointerDown = (event: PointerEvent) => {
      if (ref.current?.open && event.target instanceof Node && !ref.current.contains(event.target)) close();
    };
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && ref.current?.open) {
        close();
        ref.current.querySelector('summary')?.focus();
      }
    };
    document.addEventListener('pointerdown', onPointerDown);
    document.addEventListener('keydown', onKeyDown);
    return () => {
      document.removeEventListener('pointerdown', onPointerDown);
      document.removeEventListener('keydown', onKeyDown);
    };
  }, []);

  return (
    <details ref={ref} className="period-pill">
      <summary>
        <CalendarIcon />
        <span className="visually-hidden">Periode: </span>
        <span className="period-pill-value">{summary}</span>
        <ChevronDownIcon className="period-pill-chevron" />
      </summary>
      <div className="period-pill-popover">{children}</div>
    </details>
  );
}
