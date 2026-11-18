import { DonutChart } from '../../components/DonutChart';
import type { DonutPattern, DonutSegment } from '../../components/DonutChart';
import type { FailedReason } from '../../api/types';

const patterns: DonutPattern[] = ['solid', 'diagonal', 'dots'];

interface FailedReasonLegendProps {
  totalFailed: number;
  reasons: FailedReason[];
}

/** Donut alasan gagal kirim; legenda berisi nama alasan, jumlah, dan persen dari API. */
export function FailedReasonLegend({ totalFailed, reasons }: FailedReasonLegendProps) {
  const segments: DonutSegment[] = reasons.map((reason, index) => ({
    key: reason.reason,
    label: reason.reason,
    value: reason.count,
    share: reason.share,
    pattern: patterns[index % patterns.length],
  }));
  return (
    <section className="card" aria-labelledby="failed-reasons-title">
      <h2 id="failed-reasons-title" className="card-title">Alasan gagal kirim</h2>
      <DonutChart total={totalFailed} label="Total gagal" segments={segments} />
    </section>
  );
}
