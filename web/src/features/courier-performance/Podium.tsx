import { useRef } from 'react';

import { CheckIcon } from '../../components/icons';
import type { CourierPerformance } from '../../api/types';
import { formatNumber } from '../../lib/format-number';
import { formatPercent } from '../../lib/format-percent';
import { Avatar } from './Avatar';
import { DailyStrip } from '../../components/DailyStrip';
import { useFinishLine } from './use-finish-line';

interface PodiumProps {
  items: CourierPerformance[];
  onSelect: (courierID: string) => void;
}

const medal: Record<number, 'gold' | 'silver' | 'bronze'> = { 1: 'gold', 2: 'silver', 3: 'bronze' };

export function Podium({ items, onSelect }: PodiumProps) {
  const top = items.filter((item) => item.rank >= 1 && item.rank <= 3);
  const listRef = useRef<HTMLOListElement>(null);
  useFinishLine(listRef, top.map((item) => item.courierID));

  return (
    <ol ref={listRef} className="podium" aria-label="Podium" data-count={top.length}>
      {top.map((item) => {
        const name = item.courierName || item.courierID;
        const tone = medal[item.rank];
        const isHero = item.rank === 1;
        return (
          <li key={item.courierID} className={`podium-item${isHero ? ' podium-item-hero' : ''}`} data-finish-key={item.courierID}>
            {/* Kartu adalah kontainer biasa: nama kurir adalah tombolnya (::after meregang menutupi kartu), jadi grafik harian tidak bersarang di dalam tombol. */}
            <div className={`podium-card podium-${tone}${isHero ? ' podium-hero' : ''}`}>
              {isHero ? (
                <>
                  <span className="podium-eyebrow">Kurir terbaik periode ini</span>
                  <button type="button" className="podium-who podium-select" onClick={() => onSelect(item.courierID)}>
                    <Avatar courierID={item.courierID} name={name} medal={tone} size="lg" />
                    <span className="podium-name">{name}</span>
                  </button>
                  <span className="podium-figure">{formatPercent(item.successRate)}</span>
                  <span className="podium-label">Keberhasilan</span>
                  <span className="podium-chip">
                    <CheckIcon width={14} height={14} />
                    Tepat SLA {formatPercent(item.slaRate)}
                  </span>
                  <span className="podium-meta">{formatNumber(item.totalPackages)} paket</span>
                </>
              ) : (
                <>
                  <button type="button" className="podium-who podium-select" onClick={() => onSelect(item.courierID)}>
                    <Avatar courierID={item.courierID} name={name} medal={tone} size="md" />
                    <span className="podium-who-text">
                      <span className="podium-pos">Pos {item.rank}</span>
                      <span className="podium-name">{name}</span>
                    </span>
                  </button>
                  <span className="podium-figure">{formatPercent(item.successRate)}</span>
                  <span className="podium-label">Keberhasilan</span>
                  <span className="podium-meta">
                    Tepat SLA {formatPercent(item.slaRate)} · {formatNumber(item.totalPackages)} paket
                  </span>
                </>
              )}
              <DailyStrip daily={item.daily} size="md" />
            </div>
          </li>
        );
      })}
    </ol>
  );
}
