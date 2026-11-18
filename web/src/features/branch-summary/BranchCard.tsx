import { useId } from 'react';
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

/**
 * Kartu satu cabang. Nama cabang adalah tautan ke Performa Kurir dengan cabang dan periode yang sama;
 * ::after tautan meregang menutupi kartu (di bawah grafik) sehingga seluruh kartu tetap bisa diklik
 * tanpa membungkus grafik yang bisa difokus di dalam tautan.
 */
export function BranchCard({ branch, startDate, endDate }: BranchCardProps) {
  const titleID = useId();
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
    <article className="branch-card" aria-labelledby={titleID}>
      <div className="branch-card-head">
        <span className="branch-card-rank" aria-label={`Peringkat ${branch.rank}`}>
          {branch.rank}
        </span>
        <Link id={titleID} className="branch-card-name" to={`/courier-performance?${query.toString()}`}>
          {branch.branchID}
        </Link>
        <span className="branch-card-couriers">{formatNumber(branch.courierCount)} kurir</span>
      </div>
      <div className="branch-card-stats">
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
      </div>
      <DailyStrip daily={branch.daily} size="md" />
      <div className="branch-card-targets">
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
      </div>
    </article>
  );
}
