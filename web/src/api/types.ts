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

export type TargetStatus = 'met' | 'below';

export interface BranchTargetMetaDTO {
  month: string;
  success_rate: number;
  sla_rate: number;
}

export interface PageMetaDTO {
  page: number;
  page_size: number;
  total: number;
  /** Hanya ada di courier-performance. Null bila tanpa cabang atau tanpa target. */
  branch_target?: BranchTargetMetaDTO | null;
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

export interface BranchTargetMeta {
  month: string;
  successRate: number;
  slaRate: number;
}

export interface PageMeta {
  page: number;
  pageSize: number;
  total: number;
  branchTarget?: BranchTargetMeta | null;
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

export interface DailyCountDTO {
  date: string;
  total_packages: number;
  delivered: number;
  failed: number;
  on_time: number;
}

export interface CourierPerformanceDTO {
  rank: number;
  courier_id: string;
  courier_name: string;
  branch_id: string;
  total_packages: number;
  delivered: number;
  failed: number;
  on_time: number;
  success_rate: number;
  sla_rate: number;
  success_rate_gap: number;
  success_rate_status?: TargetStatus | null;
  daily: DailyCountDTO[];
}

export interface DailyCount {
  date: string;
  totalPackages: number;
  delivered: number;
  failed: number;
  onTime: number;
}

/** Ranking satu kurir untuk seluruh periode. Rasio dan selisih dihitung backend. */
export interface CourierPerformance {
  rank: number;
  courierID: string;
  courierName: string;
  branchID: string;
  totalPackages: number;
  delivered: number;
  failed: number;
  onTime: number;
  successRate: number;
  slaRate: number;
  successRateGap: number;
  successRateStatus: TargetStatus | null;
  daily: DailyCount[];
}

export interface FailedReasonDTO {
  reason: string;
  count: number;
  share: number;
}

export interface FailedCourierDTO {
  rank: number;
  courier_id: string;
  courier_name: string;
  branch_id: string;
  total_packages: number;
  failed: number;
  failed_rate: number;
}

export interface FailedSummaryDTO {
  total_packages: number;
  total_failed: number;
  failed_rate: number;
  previous_total_failed: number;
  failed_change_rate: number | null;
  reasons: FailedReasonDTO[];
  top_couriers: FailedCourierDTO[];
}

export interface FailedReason {
  reason: string;
  count: number;
  share: number;
}

export interface FailedCourier {
  rank: number;
  courierID: string;
  courierName: string;
  branchID: string;
  totalPackages: number;
  failed: number;
  failedRate: number;
}

/** Ringkasan gagal kirim satu periode. Rate dan perubahan dihitung backend. */
export interface FailedSummary {
  totalPackages: number;
  totalFailed: number;
  failedRate: number;
  previousTotalFailed: number;
  failedChangeRate: number | null;
  reasons: FailedReason[];
  topCouriers: FailedCourier[];
}

export interface FailedSummaryFilter {
  startDate: string;
  endDate: string;
  branchID: string;
}

export interface BranchTargetDTO {
  success_rate: number;
  sla_rate: number;
  delivered_packages: number;
  success_rate_status: TargetStatus;
  sla_rate_status: TargetStatus;
  delivered_status: TargetStatus;
}

export interface BranchPerformanceDTO {
  rank: number;
  branch_id: string;
  courier_count: number;
  total_packages: number;
  delivered: number;
  failed: number;
  on_time: number;
  success_rate: number;
  sla_rate: number;
  target: BranchTargetDTO | null;
  daily: DailyCountDTO[];
}

export interface BranchTarget {
  successRate: number;
  slaRate: number;
  deliveredPackages: number;
  successRateStatus: TargetStatus;
  slaRateStatus: TargetStatus;
  deliveredStatus: TargetStatus;
}

/** Performa satu cabang untuk seluruh periode. Target dan statusnya dihitung backend. */
export interface BranchPerformance {
  rank: number;
  branchID: string;
  courierCount: number;
  totalPackages: number;
  delivered: number;
  failed: number;
  onTime: number;
  successRate: number;
  slaRate: number;
  target: BranchTarget | null;
  daily: DailyCount[];
}

export interface BranchPerformanceFilter {
  startDate: string;
  endDate: string;
  page: number;
  pageSize: number;
}
