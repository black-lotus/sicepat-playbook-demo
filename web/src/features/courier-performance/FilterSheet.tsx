import { useId, useRef, useState } from 'react';
import type { ReactNode } from 'react';

import { useBackdropClose } from './use-backdrop-close';
import { useDialogFocus } from './use-dialog-focus';

interface FilterSheetProps {
  summary: string;
  children: ReactNode;
}

/** Filter ringkas untuk layar sempit: satu tombol ringkasan yang membuka sheet dari bawah. */
export function FilterSheet({ summary, children }: FilterSheetProps) {
  const [open, setOpen] = useState(false);
  const sheetID = useId();

  return (
    <div className="filter-sheet">
      <button
        type="button"
        className="filter-summary"
        aria-haspopup="dialog"
        aria-expanded={open}
        aria-controls={open ? sheetID : undefined}
        onClick={() => setOpen(true)}
      >
        <span className="filter-summary-label">Filter</span>
        <span className="filter-summary-value">{summary}</span>
      </button>
      {open && (
        <SheetContent id={sheetID} onClose={() => setOpen(false)}>
          {children}
        </SheetContent>
      )}
    </div>
  );
}

function SheetContent({ id, onClose, children }: { id: string; onClose: () => void; children: ReactNode }) {
  const dialogRef = useRef<HTMLDivElement>(null);
  const doneRef = useRef<HTMLButtonElement>(null);
  const titleID = `${id}-title`;

  useDialogFocus(dialogRef, doneRef, onClose);
  const backdropProps = useBackdropClose(onClose);

  return (
    <div className="detail-backdrop" {...backdropProps}>
      <div
        id={id}
        ref={dialogRef}
        className="detail-panel filter-panel"
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleID}
      >
        <h2 id={titleID} className="filter-title">Filter</h2>
        <div className="filters filters-sheet">{children}</div>
        <button ref={doneRef} type="button" className="filter-done" onClick={onClose}>Selesai</button>
      </div>
    </div>
  );
}
