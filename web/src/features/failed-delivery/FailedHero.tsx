import { ArrowDownIcon, ArrowUpIcon } from '../../components/icons';
import { formatNumber } from '../../lib/format-number';
import { formatPercent } from '../../lib/format-percent';

interface FailedHeroProps {
  totalFailed: number;
  failedRate: number;
  previousTotalFailed: number;
  failedChangeRate: number | null;
}

function ChangeChip({ changeRate }: { changeRate: number | null }) {
  if (changeRate === null) {
    return <span className="change-chip" data-trend="none">Tidak ada pembanding</span>;
  }
  if (changeRate > 0) {
    return (
      <span className="change-chip" data-trend="up">
        <ArrowUpIcon width={14} height={14} />
        Naik {formatPercent(changeRate)}
      </span>
    );
  }
  if (changeRate < 0) {
    return (
      <span className="change-chip" data-trend="down">
        <ArrowDownIcon width={14} height={14} />
        Turun {formatPercent(Math.abs(changeRate))}
      </span>
    );
  }
  return <span className="change-chip" data-trend="flat">Tidak berubah</span>;
}

/** Kartu hero total paket gagal dengan chip perubahan terhadap periode pembanding. Semua angka dari API. */
export function FailedHero({ totalFailed, failedRate, previousTotalFailed, failedChangeRate }: FailedHeroProps) {
  return (
    <section className="card failed-hero" aria-labelledby="failed-hero-title">
      <h2 id="failed-hero-title" className="card-title">Total gagal kirim</h2>
      <p className="failed-hero-total">{formatNumber(totalFailed)}</p>
      <div className="failed-hero-meta">
        <ChangeChip changeRate={failedChangeRate} />
        <span>{formatPercent(failedRate)} dari total paket</span>
        <span>Periode sebelumnya: {formatNumber(previousTotalFailed)}</span>
      </div>
    </section>
  );
}
