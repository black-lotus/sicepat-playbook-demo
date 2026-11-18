import { Link } from 'react-router-dom';

import type { BranchPerformance, TargetStatus } from '../../api/types';
import { DailyStrip } from '../../components/DailyStrip';
import { StatusBadge } from '../../components/StatusBadge';
import { formatNumber } from '../../lib/format-number';
import { formatPercent } from '../../lib/format-percent';

interface BranchCardProps {
  branch: BranchPerformance;
  startDate: string;
  endDate: string;
}

const statusLabel: Record<TargetStatus, string> = { met: 'Tercapai', below: 'Di bawah target' };

/** Kartu satu cabang. Seluruh kartu adalah tautan ke Performa Kurir dengan cabang dan periode yang sama. */
export function BranchCard({ branch, startDate, endDate }: BranchCardProps) {
  const query = new URLSearchParams({ branch_id: branch.branchID, start_date: startDate, end_date: endDate });
  const target = branch.target;
  const badges = target
    ? [
        { key: 'success', caption: 'Keberhasilan', status: target.successRateStatus },
        { key: 'sla', caption: 'Tepat SLA', status: target.slaRateStatus },
        { key: 'volume', caption: 'Paket terkirim', status: target.deliveredStatus },
      ]
    : [];

  return (
    <Link className="branch-card" to={`/courier-performance?${query.toString()}`}>
      <span className="branch-card-head">
        <span className="branch-card-rank" aria-label={`Peringkat ${branch.rank}`}>
          {branch.rank}
        </span>
        <span className="branch-card-name">{branch.branchID}</span>
        <span className="branch-card-couriers">{formatNumber(branch.courierCount)} kurir</span>
      </span>
      <span className="branch-card-stats">
        <span>
          <span className="branch-card-label">Keberhasilan</span>
          <strong>{formatPercent(branch.successRate)}</strong>
        </span>
        <span>
          <span className="branch-card-label">Tepat SLA</span>
          <strong>{formatPercent(branch.slaRate)}</strong>
        </span>
        <span>
          <span className="branch-card-label">Terkirim</span>
          <strong>{formatNumber(branch.delivered)}</strong>
        </span>
      </span>
      <DailyStrip daily={branch.daily} size="md" decorative />
      <span className="branch-card-targets">
        {target ? (
          badges.map((badge) => (
            <span key={badge.key} className="branch-card-target">
              <span className="branch-card-label">{badge.caption}</span>
              <StatusBadge status={badge.status} label={statusLabel[badge.status]} />
            </span>
          ))
        ) : (
          <span className="branch-card-label">Belum ada target</span>
        )}
      </span>
    </Link>
  );
}
