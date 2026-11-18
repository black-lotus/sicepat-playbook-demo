import type { CourierMetric } from '../../api/types';
import { DataTable, type Column } from '../../components/DataTable';
import { formatDate } from '../../lib/date';
import { formatNumber } from '../../lib/format-number';
import { formatPercent } from '../../lib/format-percent';

const columns: Column<CourierMetric>[] = [
  { key: 'date', header: 'Tanggal', render: (m) => formatDate(m.date) },
  { key: 'courier', header: 'Kurir', render: (m) => m.courierName || m.courierID },
  { key: 'branch', header: 'Cabang', render: (m) => m.branchID },
  { key: 'total', header: 'Total paket', align: 'right', render: (m) => formatNumber(m.totalPackages) },
  { key: 'delivered', header: 'Terkirim', align: 'right', render: (m) => formatNumber(m.delivered) },
  { key: 'failed', header: 'Gagal', align: 'right', render: (m) => formatNumber(m.failed) },
  { key: 'success', header: 'Keberhasilan', align: 'right', render: (m) => formatPercent(m.successRate) },
  { key: 'sla', header: 'Tepat SLA', align: 'right', render: (m) => formatPercent(m.slaRate) },
];

interface CourierPerformanceTableProps {
  metrics: CourierMetric[];
}

export function CourierPerformanceTable({ metrics }: CourierPerformanceTableProps) {
  return (
    <DataTable
      caption="Performa harian kurir"
      columns={columns}
      rows={metrics}
      rowKey={(m) => `${m.courierID}-${m.date}`}
    />
  );
}
