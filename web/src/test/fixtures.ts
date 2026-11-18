import type { CourierMetricDTO } from '../api/types';

export const courierMetricDTO: CourierMetricDTO = {
  courier_id: 'C001',
  courier_name: 'Budi',
  branch_id: 'JKT01',
  date: '2026-11-17',
  total_packages: 1250,
  delivered: 1000,
  failed: 250,
  on_time: 750,
  success_rate: 0.8,
  sla_rate: 0.75,
};
