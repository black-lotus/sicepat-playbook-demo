//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/postgres"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// Test ini butuh database yang sudah dimigrasi, misalnya dari docker compose:
// TEST_DATABASE_URL=postgres://sigesit:sigesit@localhost:5432/sigesit?sslmode=disable
func TestRepositories_AggregateAndList(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL tidak diisi")
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, "TRUNCATE delivery_events, courier_daily_metrics, couriers")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "INSERT INTO couriers (id, name, branch_id) VALUES ('C001', 'Budi', 'JKT01')")
	require.NoError(t, err)

	day := time.Date(2026, 11, 17, 0, 0, 0, 0, domain.WIB)
	due := day.Add(17 * time.Hour)
	events := []domain.DeliveryEvent{
		{AWBNumber: "A1", CourierID: "C001", BranchID: "JKT01", Status: domain.DeliveryStatusDelivered, OccurredAt: day.Add(9 * time.Hour), DueAt: due},
		{AWBNumber: "A2", CourierID: "C001", BranchID: "JKT01", Status: domain.DeliveryStatusDelivered, OccurredAt: day.Add(18 * time.Hour), DueAt: due},
		{AWBNumber: "A3", CourierID: "C001", BranchID: "JKT01", Status: domain.DeliveryStatusFailed, FailedReason: "alamat tidak ditemukan", OccurredAt: day.Add(10 * time.Hour)},
	}
	deliveries := postgres.NewDeliveryRepository(pool)
	for _, e := range events {
		require.NoError(t, deliveries.SaveEvent(ctx, e))
	}
	// Event duplikat diabaikan.
	require.NoError(t, deliveries.SaveEvent(ctx, events[0]))

	metrics := postgres.NewMetricRepository(pool)
	require.NoError(t, metrics.AggregateDaily(ctx, day))

	page, err := metrics.ListCourierMetrics(ctx, domain.MetricFilter{StartDate: day, EndDate: day, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	assert.Equal(t, 1, page.Total)

	got := page.Items[0]
	assert.Equal(t, "Budi", got.CourierName)
	assert.Equal(t, 3, got.TotalPackages)
	assert.Equal(t, 2, got.Delivered)
	assert.Equal(t, 1, got.Failed)
	assert.Equal(t, 1, got.OnTime)
}

func TestAggregateDaily_WIBBoundary(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL tidak diisi")
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, "TRUNCATE delivery_events, courier_daily_metrics, couriers")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "INSERT INTO couriers (id, name, branch_id) VALUES ('C001', 'Budi', 'JKT01')")
	require.NoError(t, err)

	// 16:59 UTC = 23:59 WIB tanggal 17, 17:00 UTC = 00:00 WIB tanggal 18.
	deliveries := postgres.NewDeliveryRepository(pool)
	for _, e := range []domain.DeliveryEvent{
		{AWBNumber: "B1", CourierID: "C001", BranchID: "JKT01", Status: domain.DeliveryStatusDelivered, OccurredAt: time.Date(2026, 11, 17, 16, 59, 0, 0, time.UTC)},
		{AWBNumber: "B2", CourierID: "C001", BranchID: "JKT01", Status: domain.DeliveryStatusDelivered, OccurredAt: time.Date(2026, 11, 17, 17, 0, 0, 0, time.UTC)},
	} {
		require.NoError(t, deliveries.SaveEvent(ctx, e))
	}

	metrics := postgres.NewMetricRepository(pool)
	for _, day := range []time.Time{
		time.Date(2026, 11, 17, 0, 0, 0, 0, domain.WIB),
		time.Date(2026, 11, 18, 0, 0, 0, 0, domain.WIB),
	} {
		require.NoError(t, metrics.AggregateDaily(ctx, day))
	}

	for _, day := range []string{"2026-11-17", "2026-11-18"} {
		var total int
		err := pool.QueryRow(ctx,
			"SELECT total_packages FROM courier_daily_metrics WHERE courier_id = 'C001' AND date = $1::date", day).Scan(&total)
		require.NoError(t, err, day)
		assert.Equal(t, 1, total, day)
	}
}

// openTestPool membuka pool ke database uji dan mengosongkan tabel.
func openTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL tidak diisi")
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	_, err = pool.Exec(ctx, "TRUNCATE delivery_events, courier_daily_metrics, couriers")
	require.NoError(t, err)
	return pool
}

// PK courier_daily_metrics adalah (courier_id, date): kurir dengan event di dua cabang pada hari WIB
// yang sama tetap menghasilkan satu baris, dengan cabang dari event terakhir hari itu.
func TestAggregateDaily_CourierInTwoBranchesSameDay(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()

	day := time.Date(2026, 11, 17, 0, 0, 0, 0, domain.WIB)
	deliveries := postgres.NewDeliveryRepository(pool)
	for _, e := range []domain.DeliveryEvent{
		{AWBNumber: "M1", BranchID: "JKT01", OccurredAt: day.Add(9 * time.Hour)},
		{AWBNumber: "M2", BranchID: "JKT02", OccurredAt: day.Add(15 * time.Hour)},
		{AWBNumber: "M3", BranchID: "JKT01", OccurredAt: day.Add(11 * time.Hour)},
	} {
		e.CourierID, e.Status = "C001", domain.DeliveryStatusDelivered
		require.NoError(t, deliveries.SaveEvent(ctx, e))
	}
	// Kurir lain di hari yang sama tetap teragregasi.
	require.NoError(t, deliveries.SaveEvent(ctx, domain.DeliveryEvent{
		AWBNumber: "M4", CourierID: "C002", BranchID: "JKT01", Status: domain.DeliveryStatusFailed, OccurredAt: day.Add(10 * time.Hour),
	}))

	require.NoError(t, postgres.NewMetricRepository(pool).AggregateDaily(ctx, day))

	var branch string
	var total int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT branch_id, total_packages FROM courier_daily_metrics WHERE courier_id = 'C001' AND date = '2026-11-17'").Scan(&branch, &total))
	assert.Equal(t, "JKT02", branch)
	assert.Equal(t, 3, total)

	var rows int
	require.NoError(t, pool.QueryRow(ctx, "SELECT COUNT(*) FROM courier_daily_metrics").Scan(&rows))
	assert.Equal(t, 2, rows)
}
