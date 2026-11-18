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
