//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/postgres"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// branchFixtureRow adalah satu baris courier_daily_metrics: kurir, cabang, hari (1 Nov 2026),
// lalu total, delivered, failed, on_time.
type branchFixtureRow struct {
	courier, branch string
	day             int
	vals            [4]int
}

// Tiga cabang:
//   - BDG01: C10 (8/8), success rate 1.0, satu kurir.
//   - JKT01: C01 hari 1 (10/9) dan hari 3 (10/8), C02 hari 1 (10/7): 30 paket, 24 delivered, rate 0.8, tiga baris, dua kurir.
//   - SBY01: C20 (10/6) dan C21 hari 2 (10/6): rate 0.6, dua kurir.
//
// Tidak ada cabang lain, jadi "cabang tanpa data" tidak muncul.
var branchFixture = []branchFixtureRow{
	{"C10", "BDG01", 2, [4]int{8, 8, 0, 8}},
	{"C01", "JKT01", 1, [4]int{10, 9, 1, 8}},
	{"C01", "JKT01", 3, [4]int{10, 8, 2, 6}},
	{"C02", "JKT01", 1, [4]int{10, 7, 3, 7}},
	{"C20", "SBY01", 1, [4]int{10, 6, 4, 5}},
	{"C21", "SBY01", 2, [4]int{10, 6, 4, 6}},
	// di luar periode, harus diabaikan
	{"C10", "BDG01", 20, [4]int{50, 1, 49, 1}},
}

func seedBranchFixture(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := openTestPool(t)
	ctx := context.Background()
	for _, r := range branchFixture {
		_, err := pool.Exec(ctx, `INSERT INTO courier_daily_metrics
			(courier_id, branch_id, date, total_packages, delivered, failed, on_time)
			VALUES ($1, $2, $3::date, $4, $5, $6, $7)`,
			r.courier, r.branch, time.Date(2026, 11, r.day, 0, 0, 0, 0, time.UTC).Format(time.DateOnly),
			r.vals[0], r.vals[1], r.vals[2], r.vals[3])
		require.NoError(t, err)
	}
	return pool
}

func branchFilter() domain.MetricFilter {
	return domain.MetricFilter{
		StartDate: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC),
		Page:      1,
		PageSize:  20,
	}
}

func TestBranchPerformance_RankingAndDaily(t *testing.T) {
	metrics := postgres.NewMetricRepository(seedBranchFixture(t))

	page, err := metrics.ListBranchPerformance(context.Background(), branchFilter())
	require.NoError(t, err)
	assert.Equal(t, 3, page.Total)
	require.Len(t, page.Items, 3)

	want := []struct {
		id       string
		couriers int
		totals   [4]int
	}{
		{"BDG01", 1, [4]int{8, 8, 0, 8}},
		{"JKT01", 2, [4]int{30, 24, 6, 21}},
		{"SBY01", 2, [4]int{20, 12, 8, 11}},
	}
	for i, got := range page.Items {
		assert.Equal(t, want[i].id, got.BranchID)
		assert.Equal(t, i+1, got.Rank)
		assert.Equal(t, want[i].couriers, got.CourierCount, got.BranchID)
		assert.Equal(t, want[i].totals, [4]int{got.TotalPackages, got.Delivered, got.Failed, got.OnTime}, got.BranchID)

		// daily: satu entri per hari WIB, hari kosong 0, jumlah sama dengan total, data di luar periode diabaikan.
		require.Len(t, got.Daily, 7, got.BranchID)
		var sum [4]int
		for d, day := range got.Daily {
			assert.Equal(t, time.Date(2026, 11, d+1, 0, 0, 0, 0, time.UTC).Format(time.DateOnly), day.Date.Format(time.DateOnly))
			sum[0] += day.TotalPackages
			sum[1] += day.Delivered
			sum[2] += day.Failed
			sum[3] += day.OnTime
		}
		assert.Equal(t, want[i].totals, sum, got.BranchID)
	}
	// JKT01: hari 1 digabung dua kurir (20/16), hari 2 kosong, hari 3 (10/8).
	assert.Equal(t, domain.DailyCount{Date: branchDay(1), TotalPackages: 20, Delivered: 16, Failed: 4, OnTime: 15}, page.Items[1].Daily[0])
	assert.Equal(t, domain.DailyCount{Date: branchDay(2)}, page.Items[1].Daily[1])
	assert.Equal(t, domain.DailyCount{Date: branchDay(3), TotalPackages: 10, Delivered: 8, Failed: 2, OnTime: 6}, page.Items[1].Daily[2])
}

func TestBranchPerformance_IgnoresBranchFilter(t *testing.T) {
	metrics := postgres.NewMetricRepository(seedBranchFixture(t))

	f := branchFilter()
	f.BranchID = "JKT01"
	page, err := metrics.ListBranchPerformance(context.Background(), f)
	require.NoError(t, err)
	assert.Equal(t, 3, page.Total)
	assert.Len(t, page.Items, 3)
}

func TestBranchPerformance_TieBreakByBranchID(t *testing.T) {
	pool := seedBranchFixture(t)
	ctx := context.Background()
	// AAA01 punya success rate 1.0 sama dengan BDG01, jadi urutan ditentukan branch_id.
	_, err := pool.Exec(ctx, `INSERT INTO courier_daily_metrics
		(courier_id, branch_id, date, total_packages, delivered, failed, on_time)
		VALUES ('C30', 'AAA01', '2026-11-04', 5, 5, 0, 5)`)
	require.NoError(t, err)

	page, err := postgres.NewMetricRepository(pool).ListBranchPerformance(ctx, branchFilter())
	require.NoError(t, err)
	require.Len(t, page.Items, 4)
	got := make([]string, len(page.Items))
	for i, it := range page.Items {
		got[i] = it.BranchID
	}
	assert.Equal(t, []string{"AAA01", "BDG01", "JKT01", "SBY01"}, got)
}

func TestBranchPerformance_SecondPage(t *testing.T) {
	metrics := postgres.NewMetricRepository(seedBranchFixture(t))

	f := branchFilter()
	f.PageSize = 2
	f.Page = 2
	page, err := metrics.ListBranchPerformance(context.Background(), f)
	require.NoError(t, err)
	assert.Equal(t, 3, page.Total)
	require.Len(t, page.Items, 1)
	assert.Equal(t, "SBY01", page.Items[0].BranchID)
	assert.Equal(t, 3, page.Items[0].Rank, "rank global, bukan posisi dalam halaman")
	assert.Len(t, page.Items[0].Daily, 7)
}

func TestBranchPerformance_PagePastEndKeepsTotal(t *testing.T) {
	metrics := postgres.NewMetricRepository(seedBranchFixture(t))

	f := branchFilter()
	f.PageSize = 2
	f.Page = 3
	page, err := metrics.ListBranchPerformance(context.Background(), f)
	require.NoError(t, err)
	assert.NotNil(t, page.Items)
	assert.Empty(t, page.Items)
	assert.Equal(t, 3, page.Total, "total adalah jumlah cabang dalam periode, bukan jumlah baris halaman")
}

func TestBranchPerformance_Empty(t *testing.T) {
	metrics := postgres.NewMetricRepository(openTestPool(t))

	page, err := metrics.ListBranchPerformance(context.Background(), branchFilter())
	require.NoError(t, err)
	assert.NotNil(t, page.Items)
	assert.Empty(t, page.Items)
	assert.Zero(t, page.Total)
}

func branchDay(day int) time.Time {
	return time.Date(2026, 11, day, 0, 0, 0, 0, time.UTC)
}
