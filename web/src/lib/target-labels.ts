export interface TargetRates {
  successRate: number;
  slaRate: number;
}

/** Label teks garis target, misalnya "Target keberhasilan 90%". */
export function targetLabels(targets: TargetRates): { success: string; sla: string } {
  return {
    success: `Target keberhasilan ${Math.round(targets.successRate * 100)}%`,
    sla: `Target tepat SLA ${Math.round(targets.slaRate * 100)}%`,
  };
}
