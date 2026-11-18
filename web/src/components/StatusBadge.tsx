import { ArrowDownIcon, ArrowUpIcon, CheckIcon } from './icons';
import './status-badge.css';

export type TargetBadgeStatus = 'met' | 'below' | 'above';

interface StatusBadgeProps {
  status: TargetBadgeStatus;
  /** Teks status yang terbaca, misalnya "Tercapai". Status tidak pernah hanya disampaikan lewat warna. */
  label: string;
}

const icons = { met: CheckIcon, below: ArrowDownIcon, above: ArrowUpIcon } as const;

/** Penanda status target: ikon dekoratif plus teks. */
export function StatusBadge({ status, label }: StatusBadgeProps) {
  const StatusIcon = icons[status];
  return (
    <span className="status-badge" data-status={status}>
      <StatusIcon width={14} height={14} />
      <span>{label}</span>
    </span>
  );
}
