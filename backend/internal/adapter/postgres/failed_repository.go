package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// topFailedCouriers adalah jumlah maksimal kurir di daftar gagal terbanyak.
const topFailedCouriers = 5

// FailedRepository mengimplementasikan port.FailedRepository.
type FailedRepository struct {
	pool *pgxpool.Pool
}

// NewFailedRepository membuat FailedRepository.
func NewFailedRepository(pool *pgxpool.Pool) *FailedRepository {
	return &FailedRepository{pool: pool}
}

// Semua query memakai $1 = awal periode (00:00 WIB), $2 = akhir periode eksklusif (00:00 WIB hari sesudah
// end_date), $3 = cabang ('' berarti semua). Batas dihitung di Go agar tidak bergantung timezone sesi database.

const failedTotalsSQL = `
SELECT COUNT(*)::int,
       COUNT(*) FILTER (WHERE status = 'failed')::int
FROM delivery_events
WHERE occurred_at >= $1 AND occurred_at < $2
  AND ($3::text = '' OR branch_id = $3::text)`

// failedReasonsSQL memakai indeks parsial idx_delivery_events_failed_branch_occurred. COLLATE "C" membuat
// urutan alasan sama di semua database.
const failedReasonsSQL = `
SELECT reason, cnt FROM (
    SELECT COALESCE(NULLIF(btrim(failed_reason), ''), '` + domain.NoReasonLabel + `') AS reason,
           COUNT(*)::int AS cnt
    FROM delivery_events
    WHERE status = 'failed'
      AND occurred_at >= $1 AND occurred_at < $2
      AND ($3::text = '' OR branch_id = $3::text)
    GROUP BY 1
) r
ORDER BY cnt DESC, reason COLLATE "C" ASC`

// failedTopCouriersSQL memilih lima kurir gagal terbanyak dulu, baru menghitung total paket masing-masing.
const failedTopCouriersSQL = `
WITH top AS (
    SELECT courier_id,
           COUNT(*)::int AS failed,
           (array_agg(branch_id ORDER BY occurred_at DESC, id DESC))[1] AS branch_id
    FROM delivery_events
    WHERE status = 'failed'
      AND occurred_at >= $1 AND occurred_at < $2
      AND ($3::text = '' OR branch_id = $3::text)
    GROUP BY courier_id
    ORDER BY failed DESC, courier_id ASC
    LIMIT $4
)
SELECT t.courier_id, COALESCE(c.name, ''), t.branch_id,
       (SELECT COUNT(*)::int FROM delivery_events e
         WHERE e.courier_id = t.courier_id
           AND e.occurred_at >= $1 AND e.occurred_at < $2
           AND ($3::text = '' OR e.branch_id = $3::text)) AS total_packages,
       t.failed
FROM top t
LEFT JOIN couriers c ON c.id = t.courier_id
ORDER BY t.failed DESC, t.courier_id ASC`

const countFailedSQL = `
SELECT COUNT(*)::int
FROM delivery_events
WHERE status = 'failed'
  AND occurred_at >= $1 AND occurred_at < $2
  AND ($3::text = '' OR branch_id = $3::text)`

// failedBounds mengubah tanggal sipil filter menjadi [00:00 WIB start, 00:00 WIB sesudah end).
func failedBounds(f domain.FailedFilter) (time.Time, time.Time) {
	return domain.DayStartWIB(f.StartDate), domain.DayStartWIB(f.EndDate).AddDate(0, 0, 1)
}

// FailedSummary mengambil ringkasan gagal kirim dengan tiga query agregat.
func (r *FailedRepository) FailedSummary(ctx context.Context, f domain.FailedFilter) (domain.FailedSummary, error) {
	from, to := failedBounds(f)
	s := domain.FailedSummary{Reasons: []domain.FailedReason{}, TopCouriers: []domain.FailedCourier{}}

	if err := r.pool.QueryRow(ctx, failedTotalsSQL, from, to, f.BranchID).Scan(&s.TotalPackages, &s.TotalFailed); err != nil {
		return domain.FailedSummary{}, fmt.Errorf("hitung total gagal kirim: %w", err)
	}
	if s.TotalFailed == 0 {
		return s, nil
	}

	rows, err := r.pool.Query(ctx, failedReasonsSQL, from, to, f.BranchID)
	if err != nil {
		return domain.FailedSummary{}, fmt.Errorf("query alasan gagal kirim: %w", err)
	}
	var reason domain.FailedReason
	if _, err := pgx.ForEachRow(rows, []any{&reason.Reason, &reason.Count}, func() error {
		s.Reasons = append(s.Reasons, reason)
		return nil
	}); err != nil {
		return domain.FailedSummary{}, fmt.Errorf("scan alasan gagal kirim: %w", err)
	}

	rows, err = r.pool.Query(ctx, failedTopCouriersSQL, from, to, f.BranchID, topFailedCouriers)
	if err != nil {
		return domain.FailedSummary{}, fmt.Errorf("query kurir gagal terbanyak: %w", err)
	}
	var c domain.FailedCourier
	if _, err := pgx.ForEachRow(rows, []any{&c.CourierID, &c.CourierName, &c.BranchID, &c.TotalPackages, &c.Failed}, func() error {
		c.Rank = len(s.TopCouriers) + 1
		s.TopCouriers = append(s.TopCouriers, c)
		return nil
	}); err != nil {
		return domain.FailedSummary{}, fmt.Errorf("scan kurir gagal terbanyak: %w", err)
	}
	return s, nil
}

// CountFailed menghitung event gagal dalam periode dan filter cabang.
func (r *FailedRepository) CountFailed(ctx context.Context, f domain.FailedFilter) (int, error) {
	from, to := failedBounds(f)
	var n int
	if err := r.pool.QueryRow(ctx, countFailedSQL, from, to, f.BranchID).Scan(&n); err != nil {
		return 0, fmt.Errorf("hitung gagal kirim: %w", err)
	}
	return n, nil
}
