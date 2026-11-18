// Package postgres berisi implementasi repository dengan PostgreSQL (pgx).
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// NewPool membuka koneksi pool ke PostgreSQL dan memastikan database bisa dijangkau.
func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("buka pool postgres: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}

// MetricRepository mengimplementasikan port.MetricRepository.
type MetricRepository struct {
	pool *pgxpool.Pool
}

// NewMetricRepository membuat MetricRepository.
func NewMetricRepository(pool *pgxpool.Pool) *MetricRepository {
	return &MetricRepository{pool: pool}
}

const listCourierMetricsSQL = `
SELECT m.courier_id, COALESCE(c.name, ''), m.branch_id, m.date,
       m.total_packages, m.delivered, m.failed, m.on_time,
       COUNT(*) OVER () AS total
FROM courier_daily_metrics m
LEFT JOIN couriers c ON c.id = m.courier_id
WHERE m.date BETWEEN $1::date AND $2::date
  AND ($3::text = '' OR m.branch_id = $3::text)
ORDER BY m.date DESC, m.courier_id
LIMIT $4 OFFSET $5`

// ListCourierMetrics mengambil satu halaman KPI kurir dalam satu query, beserta total baris.
func (r *MetricRepository) ListCourierMetrics(ctx context.Context, f domain.MetricFilter) (domain.CourierMetricPage, error) {
	rows, err := r.pool.Query(ctx, listCourierMetricsSQL, f.StartDate, f.EndDate, f.BranchID, f.PageSize, f.Offset())
	if err != nil {
		return domain.CourierMetricPage{}, fmt.Errorf("query courier_daily_metrics: %w", err)
	}

	page := domain.CourierMetricPage{Items: []domain.CourierMetric{}}
	var m domain.CourierMetric
	scan := []any{&m.CourierID, &m.CourierName, &m.BranchID, &m.Date, &m.TotalPackages, &m.Delivered, &m.Failed, &m.OnTime, &page.Total}
	if _, err := pgx.ForEachRow(rows, scan, func() error {
		page.Items = append(page.Items, m)
		return nil
	}); err != nil {
		return domain.CourierMetricPage{}, fmt.Errorf("scan courier_daily_metrics: %w", err)
	}
	return page, nil
}

// courierAggCTE mengagregasi per kurir untuk filter periode dan cabang ($1 sampai $3). Success rate memakai
// rumus yang sama dengan domain (delivered / total, 0 bila total 0). Dipakai query halaman dan query hitung.
const courierAggCTE = `
agg AS (
    SELECT m.courier_id,
           (array_agg(m.branch_id ORDER BY m.date DESC))[1] AS branch_id,
           SUM(m.total_packages)::int AS total_packages,
           SUM(m.delivered)::int      AS delivered,
           SUM(m.failed)::int         AS failed,
           SUM(m.on_time)::int        AS on_time,
           COALESCE(SUM(m.delivered)::float8 / NULLIF(SUM(m.total_packages), 0), 0) AS success_rate
    FROM courier_daily_metrics m
    WHERE m.date BETWEEN $1::date AND $2::date
      AND ($3::text = '' OR m.branch_id = $3::text)
    GROUP BY m.courier_id
)`

// listCourierPerformanceSQL memberi peringkat global lalu memotong satu halaman.
// Rank, rate pemimpin, dan total dihitung sebelum LIMIT sehingga stabil di semua halaman.
const listCourierPerformanceSQL = `
WITH` + courierAggCTE + `, ranked AS (
    SELECT agg.*,
           ROW_NUMBER() OVER (ORDER BY success_rate DESC, courier_id ASC) AS rank,
           MAX(success_rate) OVER ()                                      AS leader_success_rate,
           COUNT(*) OVER ()                                               AS total
    FROM agg
)
SELECT r.rank, r.courier_id, COALESCE(c.name, ''), r.branch_id,
       r.total_packages, r.delivered, r.failed, r.on_time,
       r.leader_success_rate, r.total
FROM ranked r
LEFT JOIN couriers c ON c.id = r.courier_id
ORDER BY r.rank
LIMIT $4 OFFSET $5`

// countCourierPerformanceSQL menghitung kurir dalam filter. Hanya dipakai saat halaman yang diminta kosong
// di luar halaman pertama, karena total dari baris halaman tidak tersedia.
const countCourierPerformanceSQL = `
WITH` + courierAggCTE + `
SELECT COUNT(*) FROM agg`

const listCourierDailySQL = `
SELECT m.courier_id, m.date, m.total_packages, m.delivered, m.failed, m.on_time
FROM courier_daily_metrics m
WHERE m.courier_id = ANY($1::text[])
  AND m.date BETWEEN $2::date AND $3::date
  AND ($4::text = '' OR m.branch_id = $4::text)
ORDER BY m.courier_id, m.date`

// ListCourierPerformance mengambil satu halaman ranking kurir dan tren hariannya dengan dua query,
// tanpa N+1. Hari tanpa data diisi nol di Go lewat domain.FillDaily.
func (r *MetricRepository) ListCourierPerformance(ctx context.Context, f domain.MetricFilter) (domain.CourierPerformancePage, error) {
	rows, err := r.pool.Query(ctx, listCourierPerformanceSQL, f.StartDate, f.EndDate, f.BranchID, f.PageSize, f.Offset())
	if err != nil {
		return domain.CourierPerformancePage{}, fmt.Errorf("query peringkat kurir: %w", err)
	}
	page := domain.CourierPerformancePage{Items: []domain.CourierPerformance{}}
	var p domain.CourierPerformance
	scan := []any{&p.Rank, &p.CourierID, &p.CourierName, &p.BranchID,
		&p.TotalPackages, &p.Delivered, &p.Failed, &p.OnTime, &p.LeaderSuccessRate, &page.Total}
	if _, err := pgx.ForEachRow(rows, scan, func() error {
		page.Items = append(page.Items, p)
		return nil
	}); err != nil {
		return domain.CourierPerformancePage{}, fmt.Errorf("scan peringkat kurir: %w", err)
	}
	if len(page.Items) == 0 {
		if f.Offset() > 0 {
			if err := r.pool.QueryRow(ctx, countCourierPerformanceSQL, f.StartDate, f.EndDate, f.BranchID).Scan(&page.Total); err != nil {
				return domain.CourierPerformancePage{}, fmt.Errorf("hitung peringkat kurir: %w", err)
			}
		}
		return page, nil
	}

	ids := make([]string, len(page.Items))
	for i, it := range page.Items {
		ids[i] = it.CourierID
	}
	dailyRows, err := r.pool.Query(ctx, listCourierDailySQL, ids, f.StartDate, f.EndDate, f.BranchID)
	if err != nil {
		return domain.CourierPerformancePage{}, fmt.Errorf("query tren harian kurir: %w", err)
	}
	byCourier := make(map[string][]domain.DailyCount, len(ids))
	var courierID string
	var d domain.DailyCount
	dailyScan := []any{&courierID, &d.Date, &d.TotalPackages, &d.Delivered, &d.Failed, &d.OnTime}
	if _, err := pgx.ForEachRow(dailyRows, dailyScan, func() error {
		byCourier[courierID] = append(byCourier[courierID], d)
		return nil
	}); err != nil {
		return domain.CourierPerformancePage{}, fmt.Errorf("scan tren harian kurir: %w", err)
	}
	for i := range page.Items {
		page.Items[i].Daily = domain.FillDaily(byCourier[page.Items[i].CourierID], f.StartDate, f.EndDate)
	}
	return page, nil
}

const aggregateDailySQL = `
INSERT INTO courier_daily_metrics (courier_id, branch_id, date, total_packages, delivered, failed, on_time, updated_at)
SELECT courier_id,
       (array_agg(branch_id ORDER BY occurred_at DESC, id DESC))[1],
       $1::date,
       COUNT(*),
       COUNT(*) FILTER (WHERE status = 'delivered'),
       COUNT(*) FILTER (WHERE status = 'failed'),
       COUNT(*) FILTER (WHERE status = 'delivered' AND due_at IS NOT NULL AND occurred_at <= due_at),
       now()
FROM delivery_events
WHERE occurred_at >= $2 AND occurred_at < $3
GROUP BY courier_id
ON CONFLICT (courier_id, date) DO UPDATE SET
    branch_id      = EXCLUDED.branch_id,
    total_packages = EXCLUDED.total_packages,
    delivered      = EXCLUDED.delivered,
    failed         = EXCLUDED.failed,
    on_time        = EXCLUDED.on_time,
    updated_at     = EXCLUDED.updated_at`

// AggregateDaily menghitung ulang KPI semua kurir untuk satu tanggal WIB dalam satu query agregat.
// Event yang dihitung berada di [00:00 WIB, 00:00 WIB hari berikutnya), dan kolom date diisi tanggal WIB.
// Satu baris per kurir per hari (PK courier_id, date); branch_id diambil dari event terakhir hari itu.
// Batas hari dihitung di Go agar hasilnya tidak bergantung pada timezone sesi database.
func (r *MetricRepository) AggregateDaily(ctx context.Context, date time.Time) error {
	start := domain.DayStartWIB(date)
	end := start.AddDate(0, 0, 1)
	if _, err := r.pool.Exec(ctx, aggregateDailySQL, start.Format(time.DateOnly), start, end); err != nil {
		return fmt.Errorf("upsert courier_daily_metrics: %w", err)
	}
	return nil
}
