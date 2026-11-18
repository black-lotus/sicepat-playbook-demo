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

// jkt01Seed adalah tabel Data contoh JKT01 di spec: tiap hari ditulis total/delivered/failed/on_time,
// urut 1 sampai 7 Nov 2026. Hari 0/0/0/0 tidak punya baris, sama seperti hasil worker.
var jkt01Seed = []struct {
	id, name string
	daily    [7][4]int
}{
	{"C001", "Budi Santoso", [7][4]int{{6, 6, 0, 5}, {5, 5, 0, 5}, {6, 5, 1, 5}, {6, 5, 1, 4}, {5, 5, 0, 5}, {6, 5, 1, 5}, {6, 5, 1, 4}}},
	{"C003", "Dewi Lestari", [7][4]int{{5, 5, 0, 5}, {5, 4, 1, 4}, {5, 4, 1, 4}, {5, 5, 0, 5}, {5, 4, 1, 4}, {5, 4, 1, 4}, {5, 4, 1, 4}}},
	{"C004", "Agus Pratama", [7][4]int{{3, 3, 0, 3}, {3, 2, 1, 2}, {3, 2, 1, 1}, {3, 2, 1, 2}, {3, 2, 1, 1}, {3, 2, 1, 2}, {3, 2, 1, 1}}},
	{"C002", "Sari Wulandari", [7][4]int{{1, 0, 1, 0}, {1, 0, 1, 0}, {0, 0, 0, 0}, {1, 0, 1, 0}, {0, 0, 0, 0}, {1, 0, 1, 0}, {0, 0, 0, 0}}},
}

func seedJKT01(t *testing.T) *postgres.MetricRepository {
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
	for _, c := range jkt01Seed {
		_, err = pool.Exec(ctx, "INSERT INTO couriers (id, name, branch_id) VALUES ($1, $2, 'JKT01')", c.id, c.name)
		require.NoError(t, err)
		for i, d := range c.daily {
			if d[0] == 0 {
				continue
			}
			_, err = pool.Exec(ctx, `INSERT INTO courier_daily_metrics
				(courier_id, branch_id, date, total_packages, delivered, failed, on_time)
				VALUES ($1, 'JKT01', $2::date, $3, $4, $5, $6)`,
				c.id, time.Date(2026, 11, i+1, 0, 0, 0, 0, time.UTC).Format(time.DateOnly), d[0], d[1], d[2], d[3])
			require.NoError(t, err)
		}
	}
	return postgres.NewMetricRepository(pool)
}

func jkt01Filter() domain.MetricFilter {
	return domain.MetricFilter{
		StartDate: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC),
		BranchID:  "JKT01",
		Page:      1,
		PageSize:  50,
	}
}

func TestListCourierPerformance_JKT01(t *testing.T) {
	metrics := seedJKT01(t)

	page, err := metrics.ListCourierPerformance(context.Background(), jkt01Filter())
	require.NoError(t, err)
	assert.Equal(t, 4, page.Total)
	require.Len(t, page.Items, 4)

	wantOrder := []string{"C001", "C003", "C004", "C002"}
	wantTotals := [][4]int{{40, 36, 4, 33}, {35, 30, 5, 30}, {21, 15, 6, 12}, {4, 0, 4, 0}}
	for i, got := range page.Items {
		assert.Equal(t, wantOrder[i], got.CourierID)
		assert.Equal(t, i+1, got.Rank)
		assert.Equal(t, "JKT01", got.BranchID)
		assert.Equal(t, wantTotals[i], [4]int{got.TotalPackages, got.Delivered, got.Failed, got.OnTime}, got.CourierID)
		assert.InDelta(t, 0.9, got.LeaderSuccessRate, 1e-12)

		// daily: tepat satu entri per hari WIB, hari kosong bernilai 0, jumlah sama dengan total.
		require.Len(t, got.Daily, 7, got.CourierID)
		var sum [4]int
		for d, day := range got.Daily {
			assert.Equal(t, time.Date(2026, 11, d+1, 0, 0, 0, 0, time.UTC).Format(time.DateOnly), day.Date.Format(time.DateOnly))
			assert.Equal(t, jkt01Seed[seedIndex(got.CourierID)].daily[d], [4]int{day.TotalPackages, day.Delivered, day.Failed, day.OnTime})
			sum[0] += day.TotalPackages
			sum[1] += day.Delivered
			sum[2] += day.Failed
			sum[3] += day.OnTime
		}
		assert.Equal(t, wantTotals[i], sum, got.CourierID)
	}
	assert.Equal(t, "Budi Santoso", page.Items[0].CourierName)
	assert.Zero(t, page.Items[0].SuccessRateGap())
	assert.InDelta(t, 0.9-15.0/21.0, page.Items[2].SuccessRateGap(), 1e-12)
}

func seedIndex(courierID string) int {
	for i, c := range jkt01Seed {
		if c.id == courierID {
			return i
		}
	}
	return -1
}

func TestListCourierPerformance_SecondPage(t *testing.T) {
	metrics := seedJKT01(t)

	f := jkt01Filter()
	f.PageSize = 2
	f.Page = 2
	page, err := metrics.ListCourierPerformance(context.Background(), f)
	require.NoError(t, err)
	assert.Equal(t, 4, page.Total)
	require.Len(t, page.Items, 2)

	assert.Equal(t, "C004", page.Items[0].CourierID)
	assert.Equal(t, 3, page.Items[0].Rank)
	assert.Equal(t, "C002", page.Items[1].CourierID)
	assert.Equal(t, 4, page.Items[1].Rank)
	for _, it := range page.Items {
		assert.InDelta(t, 0.9, it.LeaderSuccessRate, 1e-12, it.CourierID)
		assert.Len(t, it.Daily, 7, it.CourierID)
	}
	assert.InDelta(t, 0.9, page.Items[1].SuccessRateGap(), 1e-12)
}

func TestListCourierPerformance_BranchFilter(t *testing.T) {
	metrics := seedJKT01(t)

	f := jkt01Filter()
	f.BranchID = "BDG01"
	page, err := metrics.ListCourierPerformance(context.Background(), f)
	require.NoError(t, err)
	assert.Zero(t, page.Total)
	assert.NotNil(t, page.Items)
	assert.Empty(t, page.Items)
}

func TestListCourierPerformance_TieBreakByCourierID(t *testing.T) {
	metrics := seedJKT01(t)
	url := os.Getenv("TEST_DATABASE_URL")
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	// C000 dan C005 punya success rate 0.9 yang sama persis dengan C001, jadi urutan ditentukan courier_id.
	for _, id := range []string{"C005", "C000"} {
		_, err = pool.Exec(ctx, `INSERT INTO courier_daily_metrics
			(courier_id, branch_id, date, total_packages, delivered, failed, on_time)
			VALUES ($1, 'JKT01', '2026-11-03', 10, 9, 1, 9)`, id)
		require.NoError(t, err)
	}

	page, err := metrics.ListCourierPerformance(ctx, jkt01Filter())
	require.NoError(t, err)
	require.Len(t, page.Items, 6)
	got := make([]string, len(page.Items))
	for i, it := range page.Items {
		got[i] = it.CourierID
	}
	assert.Equal(t, []string{"C000", "C001", "C005", "C003", "C004", "C002"}, got)
}

func TestListCourierPerformance_PagePastEndKeepsTotal(t *testing.T) {
	metrics := seedJKT01(t)

	f := jkt01Filter()
	f.PageSize = 2
	f.Page = 3
	page, err := metrics.ListCourierPerformance(context.Background(), f)
	require.NoError(t, err)
	assert.NotNil(t, page.Items)
	assert.Empty(t, page.Items)
	assert.Equal(t, 4, page.Total, "total adalah jumlah kurir dalam filter, bukan jumlah baris halaman")
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

// AC3 dari event sampai ranking: batas hari WIB menentukan event mana yang terhitung.
func TestCourierPerformance_AC3_EventToRanking(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	_, err := pool.Exec(ctx, "INSERT INTO couriers (id, name, branch_id) VALUES ('C001', 'Budi', 'JKT01')")
	require.NoError(t, err)

	wib := func(month time.Month, day, hour, minute int) time.Time {
		return time.Date(2026, month, day, hour, minute, 0, 0, domain.WIB)
	}
	deliveries := postgres.NewDeliveryRepository(pool)
	for _, e := range []domain.DeliveryEvent{
		{AWBNumber: "X0", OccurredAt: wib(10, 31, 23, 59)}, // sehari sebelum periode, tidak terhitung
		{AWBNumber: "X1", OccurredAt: wib(11, 1, 5, 0)},    // terhitung
		{AWBNumber: "X2", OccurredAt: wib(11, 7, 23, 30)},  // terhitung
		{AWBNumber: "X3", OccurredAt: wib(11, 8, 0, 0)},    // sehari setelah periode, tidak terhitung
	} {
		e.CourierID, e.BranchID, e.Status = "C001", "JKT01", domain.DeliveryStatusDelivered
		require.NoError(t, deliveries.SaveEvent(ctx, e))
	}

	metrics := postgres.NewMetricRepository(pool)
	// Agregasi semua hari terkait, termasuk 31 Okt dan 8 Nov, supaya batas benar-benar diuji.
	for d := 0; d < 9; d++ {
		require.NoError(t, metrics.AggregateDaily(ctx, time.Date(2026, 10, 31+d, 0, 0, 0, 0, domain.WIB)))
	}

	page, err := metrics.ListCourierPerformance(ctx, domain.MetricFilter{
		StartDate: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC),
		Page:      1,
		PageSize:  20,
	})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	got := page.Items[0]
	assert.Equal(t, 2, got.TotalPackages)
	assert.Equal(t, 2, got.Delivered)
	require.Len(t, got.Daily, 7)
	assert.Equal(t, 1, got.Daily[0].TotalPackages, "1 Nov")
	assert.Equal(t, 1, got.Daily[6].TotalPackages, "7 Nov")
	for _, d := range got.Daily[1:6] {
		assert.Zero(t, d.TotalPackages, d.Date.Format(time.DateOnly))
	}
}

// Kurir pindah cabang di dalam periode: branch_id item adalah cabang pada tanggal terakhir,
// dan filter cabang hanya menghitung hari di cabang itu.
func TestListCourierPerformance_CourierMovesBranch(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	for _, r := range []struct {
		day    int
		branch string
		total  int
	}{{1, "JKT01", 4}, {2, "JKT01", 3}, {3, "JKT02", 5}} {
		_, err := pool.Exec(ctx, `INSERT INTO courier_daily_metrics
			(courier_id, branch_id, date, total_packages, delivered, failed, on_time)
			VALUES ('C001', $1, make_date(2026, 11, $2), $3, $3, 0, $3)`, r.branch, r.day, r.total)
		require.NoError(t, err)
	}
	metrics := postgres.NewMetricRepository(pool)

	cases := []struct {
		name      string
		branch    string
		wantItems int
		wantTotal int
		wantDaily [7]int
		wantBr    string
	}{
		{"semua cabang", "", 12, 12, [7]int{4, 3, 5, 0, 0, 0, 0}, "JKT02"},
		{"cabang lama", "JKT01", 7, 7, [7]int{4, 3, 0, 0, 0, 0, 0}, "JKT01"},
		{"cabang baru", "JKT02", 5, 5, [7]int{0, 0, 5, 0, 0, 0, 0}, "JKT02"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := jkt01Filter()
			f.BranchID = c.branch
			page, err := metrics.ListCourierPerformance(ctx, f)
			require.NoError(t, err)
			assert.Equal(t, 1, page.Total)
			require.Len(t, page.Items, 1)
			got := page.Items[0]
			assert.Equal(t, c.wantBr, got.BranchID)
			assert.Equal(t, c.wantTotal, got.TotalPackages)
			require.Len(t, got.Daily, 7)
			var daily [7]int
			for i, d := range got.Daily {
				daily[i] = d.TotalPackages
			}
			assert.Equal(t, c.wantDaily, daily)
		})
	}
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
