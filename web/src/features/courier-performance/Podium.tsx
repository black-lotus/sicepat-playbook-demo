import type { CourierPerformance } from '../../api/types';
import { formatNumber } from '../../lib/format-number';
import { formatPercent } from '../../lib/format-percent';
import { DailyStrip } from './DailyStrip';

interface PodiumProps {
  items: CourierPerformance[];
  onSelect: (courierID: string) => void;
}

const medal: Record<number, string> = { 1: 'gold', 2: 'silver', 3: 'bronze' };

export function Podium({ items, onSelect }: PodiumProps) {
  const top = items.filter((item) => item.rank >= 1 && item.rank <= 3);

  return (
    <ol className="podium" aria-label="Podium">
      {top.map((item) => (
        <li key={item.courierID} className="podium-item">
          <button type="button" className={`podium-card podium-${medal[item.rank]}`} onClick={() => onSelect(item.courierID)}>
            <span className="podium-pos">{item.rank}</span>
            <span className="podium-name">{item.courierName || item.courierID}</span>
            <span className="podium-figure">{formatPercent(item.successRate)}</span>
            <span className="podium-label">Keberhasilan</span>
            <span className="podium-meta">
              Tepat SLA {formatPercent(item.slaRate)} · {formatNumber(item.totalPackages)} paket
            </span>
            <DailyStrip daily={item.daily} size="md" decorative />
          </button>
        </li>
      ))}
    </ol>
  );
}
