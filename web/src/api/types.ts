/** Bentuk JSON dari backend (snake_case). Hanya dipakai di dalam src/api/. */
export interface CourierMetricDTO {
  courier_id: string;
  courier_name: string;
  branch_id: string;
  date: string;
  total_packages: number;
  delivered: number;
  failed: number;
  on_time: number;
  success_rate: number;
  sla_rate: number;
}

export interface PageMetaDTO {
  page: number;
  page_size: number;
  total: number;
}

/** KPI harian satu kurir. Angka rasio sudah dihitung backend (0..1). */
export interface CourierMetric {
  courierID: string;
  courierName: string;
  branchID: string;
  date: string;
  totalPackages: number;
  delivered: number;
  failed: number;
  onTime: number;
  successRate: number;
  slaRate: number;
}

export interface PageMeta {
  page: number;
  pageSize: number;
  total: number;
}

export interface Paginated<T> {
  items: T[];
  meta: PageMeta;
}

export interface CourierMetricFilter {
  startDate: string;
  endDate: string;
  branchID?: string;
  page: number;
  pageSize: number;
}
