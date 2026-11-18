//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/postgres"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

const (
	reasonAbsent  = "Penerima tidak di tempat"
	reasonAddress = "Alamat tidak ditemukan"
	reasonRefused = "Penerima menolak"
	reasonWeather = "Cuaca buruk"
)

func failedFilter(branch string) domain.FailedFilter {
	return domain.FailedFilter{
		StartDate: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC),
		BranchID:  branch,
	}
}

// seedFailedFixture mengisi event sintetis (tanpa data pribadi) dan mengembalikan repository.
//
// Periode 2026-11-01..2026-11-07 (WIB), total 35 event, 15 gagal:
//   - JKT01: C001 (9 event, 5 gagal), C002 (9, 4), C006 (3, 1) = 21 event, 10 gagal
//   - JKT02: C003 (5, 2), C004 (3, 1), C006 (2, 1), C007 (1, 1) = 11 event, 5 gagal
//   - DLV01: C005 hanya delivered, 3 event
//
// C006 pindah dari JKT01 (1 Nov) ke JKT02 (5 Nov). Alasan NULL, ” dan spasi menjadi "Tanpa alasan".
// Dua event gagal tepat di luar periode (23:59 WIB 31 Okt dan 00:00 WIB 8 Nov) tidak boleh terhitung.
func seedFailedFixture(t *testing.T) *postgres.FailedRepository {
	t.Helper()
	pool := openTestPool(t)
	ctx := context.Background()

	for _, c := range []struct{ id, name, branch string }{
		{"C001", "Kurir Satu", "JKT01"}, {"C002", "Kurir Dua", "JKT01"}, {"C003", "Kurir Tiga", "JKT02"},
		{"C004", "Kurir Empat", "JKT02"}, {"C005", "Kurir Lima", "DLV01"}, {"C006", "Kurir Enam", "JKT02"},
		{"C007", "Kurir Tujuh", "JKT02"},
	} {
		_, err := pool.Exec(ctx, "INSERT INTO couriers (id, name, branch_id) VALUES ($1, $2, $3)", c.id, c.name, c.branch)
		require.NoError(t, err)
	}

	n := 0
	ins := func(courier, branch, status string, reason *string, at time.Time) {
		n++
		_, err := pool.Exec(ctx, `INSERT INTO delivery_events
			(awb_number, courier_id, branch_id, status, failed_reason, occurred_at)
			VALUES ($1, $2, $3, $4, $5, $6)`,
			fmt.Sprintf("AWB%04d", n), courier, branch, status, reason, at)
		require.NoError(t, err)
	}
	str := func(s string) *string { return &s }
	at := func(day, hour int) time.Time { return time.Date(2026, 11, day, hour, 0, 0, 0, domain.WIB) }
	repeat := func(k int, f func(i int)) {
		for i := 0; i < k; i++ {
			f(i)
		}
	}

	repeat(3, func(i int) { ins("C001", "JKT01", "failed", str(reasonAbsent), at(2, 9+i)) })
	ins("C001", "JKT01", "failed", str(reasonAddress), at(3, 9))
	ins("C001", "JKT01", "failed", nil, at(3, 10))
	repeat(4, func(i int) { ins("C001", "JKT01", "delivered", nil, at(4, 9+i)) })

	repeat(2, func(i int) { ins("C002", "JKT01", "failed", str(reasonAbsent), at(2, 9+i)) })
	ins("C002", "JKT01", "failed", str(""), at(3, 9))
	ins("C002", "JKT01", "failed", str("   "), at(3, 10))
	repeat(5, func(i int) { ins("C002", "JKT01", "delivered", nil, at(4, 9+i)) })

	repeat(2, func(i int) { ins("C003", "JKT02", "failed", str(reasonAddress), at(2, 9+i)) })
	repeat(3, func(i int) { ins("C003", "JKT02", "delivered", nil, at(4, 9+i)) })

	ins("C004", "JKT02", "failed", str(reasonRefused), at(2, 9))
	repeat(2, func(i int) { ins("C004", "JKT02", "delivered", nil, at(4, 9+i)) })

	repeat(3, func(i int) { ins("C005", "DLV01", "delivered", nil, at(4, 9+i)) })

	ins("C006", "JKT01", "failed", str(reasonRefused), at(1, 9))
	repeat(2, func(i int) { ins("C006", "JKT01", "delivered", nil, at(1, 10+i)) })
	ins("C006", "JKT02", "failed", str(reasonRefused), at(5, 9))
	ins("C006", "JKT02", "delivered", nil, at(5, 10))

	ins("C007", "JKT02", "failed", str(reasonWeather), at(6, 9))

	// Tepat di luar periode: 23:59 WIB 31 Okt dan 00:00 WIB 8 Nov.
	ins("C001", "JKT01", "failed", str(reasonAbsent), time.Date(2026, 10, 31, 23, 59, 0, 0, domain.WIB))
	ins("C001", "JKT01", "failed", str(reasonAbsent), time.Date(2026, 11, 8, 0, 0, 0, 0, domain.WIB))

	return postgres.NewFailedRepository(pool)
}

func reasonsOf(s domain.FailedSummary) []domain.FailedReason { return s.Reasons }

func TestFailedSummary_Totals(t *testing.T) {
	repo := seedFailedFixture(t)
	ctx := context.Background()

	t.Run("semua cabang", func(t *testing.T) {
		got, err := repo.FailedSummary(ctx, failedFilter(""))
		require.NoError(t, err)
		assert.Equal(t, 35, got.TotalPackages)
		assert.Equal(t, 15, got.TotalFailed)
		assert.Zero(t, got.PreviousTotalFailed, "diisi service, bukan repo")
		assert.InDelta(t, 15.0/35, got.FailedRate(), 1e-9)
		assert.Equal(t, []domain.FailedReason{
			{Reason: reasonAbsent, Count: 5},
			{Reason: reasonAddress, Count: 3},
			{Reason: reasonRefused, Count: 3},
			{Reason: domain.NoReasonLabel, Count: 3},
			{Reason: reasonWeather, Count: 1},
		}, got.Reasons)
	})

	t.Run("satu cabang", func(t *testing.T) {
		got, err := repo.FailedSummary(ctx, failedFilter("JKT01"))
		require.NoError(t, err)
		assert.Equal(t, 21, got.TotalPackages)
		assert.Equal(t, 10, got.TotalFailed)
		assert.Equal(t, []domain.FailedReason{
			{Reason: reasonAbsent, Count: 5},
			{Reason: domain.NoReasonLabel, Count: 3},
			{Reason: reasonAddress, Count: 1},
			{Reason: reasonRefused, Count: 1},
		}, got.Reasons)

		got, err = repo.FailedSummary(ctx, failedFilter("JKT02"))
		require.NoError(t, err)
		assert.Equal(t, 11, got.TotalPackages)
		assert.Equal(t, 5, got.TotalFailed)
	})
}

func TestFailedSummary_WIBBoundary(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	repo := postgres.NewFailedRepository(pool)

	// Batas WIB: 17:00 UTC = 00:00 WIB hari berikutnya.
	events := []time.Time{
		time.Date(2026, 10, 31, 16, 59, 0, 0, time.UTC), // 23:59 WIB 31 Okt, di luar
		time.Date(2026, 10, 31, 17, 0, 0, 0, time.UTC),  // 00:00 WIB 1 Nov, masuk
		time.Date(2026, 11, 7, 16, 59, 0, 0, time.UTC),  // 23:59 WIB 7 Nov, masuk
		time.Date(2026, 11, 7, 17, 0, 0, 0, time.UTC),   // 00:00 WIB 8 Nov, di luar
	}
	for i, at := range events {
		_, err := pool.Exec(ctx, `INSERT INTO delivery_events (awb_number, courier_id, branch_id, status, failed_reason, occurred_at)
			VALUES ($1, 'C001', 'JKT01', 'failed', 'x', $2)`, fmt.Sprintf("B%d", i), at)
		require.NoError(t, err)
	}

	got, err := repo.FailedSummary(ctx, failedFilter(""))
	require.NoError(t, err)
	assert.Equal(t, 2, got.TotalFailed)
	assert.Equal(t, 2, got.TotalPackages)

	n, err := repo.CountFailed(ctx, failedFilter(""))
	require.NoError(t, err)
	assert.Equal(t, 2, n)
}

func TestFailedSummary_NullReason(t *testing.T) {
	repo := seedFailedFixture(t)
	got, err := repo.FailedSummary(context.Background(), failedFilter("JKT01"))
	require.NoError(t, err)

	var tanpa int
	for _, r := range got.Reasons {
		assert.NotEmpty(t, r.Reason)
		if r.Reason == domain.NoReasonLabel {
			tanpa = r.Count
		}
	}
	assert.Equal(t, 3, tanpa, "NULL, kosong, dan spasi digabung jadi satu alasan")
}

func TestFailedSummary_TopCouriers(t *testing.T) {
	repo := seedFailedFixture(t)
	ctx := context.Background()

	t.Run("maksimal 5, tie-break courier_id", func(t *testing.T) {
		got, err := repo.FailedSummary(ctx, failedFilter(""))
		require.NoError(t, err)
		// Enam kurir gagal; C007 (failed 1, id terbesar) terpotong.
		assert.Equal(t, []domain.FailedCourier{
			{Rank: 1, CourierID: "C001", CourierName: "Kurir Satu", BranchID: "JKT01", TotalPackages: 9, Failed: 5},
			{Rank: 2, CourierID: "C002", CourierName: "Kurir Dua", BranchID: "JKT01", TotalPackages: 9, Failed: 4},
			{Rank: 3, CourierID: "C003", CourierName: "Kurir Tiga", BranchID: "JKT02", TotalPackages: 5, Failed: 2},
			{Rank: 4, CourierID: "C006", CourierName: "Kurir Enam", BranchID: "JKT02", TotalPackages: 5, Failed: 2},
			{Rank: 5, CourierID: "C004", CourierName: "Kurir Empat", BranchID: "JKT02", TotalPackages: 3, Failed: 1},
		}, got.TopCouriers)
	})

	t.Run("kurir pindah cabang dihitung per filter cabang", func(t *testing.T) {
		got, err := repo.FailedSummary(ctx, failedFilter("JKT01"))
		require.NoError(t, err)
		require.Len(t, got.TopCouriers, 3)
		last := got.TopCouriers[2]
		assert.Equal(t, domain.FailedCourier{Rank: 3, CourierID: "C006", CourierName: "Kurir Enam", BranchID: "JKT01", TotalPackages: 3, Failed: 1}, last)

		got, err = repo.FailedSummary(ctx, failedFilter("JKT02"))
		require.NoError(t, err)
		require.Len(t, got.TopCouriers, 4)
		assert.Equal(t, []string{"C003", "C004", "C006", "C007"}, []string{
			got.TopCouriers[0].CourierID, got.TopCouriers[1].CourierID, got.TopCouriers[2].CourierID, got.TopCouriers[3].CourierID,
		})
		assert.Equal(t, 2, got.TopCouriers[2].TotalPackages, "C006 di JKT02: 1 gagal dan 1 delivered")
	})

	t.Run("hanya kurir dengan gagal", func(t *testing.T) {
		got, err := repo.FailedSummary(ctx, failedFilter(""))
		require.NoError(t, err)
		for _, c := range got.TopCouriers {
			assert.NotEqual(t, "C005", c.CourierID)
			assert.Positive(t, c.Failed)
		}
	})
}

func TestFailedSummary_ReasonsSumToTotal(t *testing.T) {
	repo := seedFailedFixture(t)
	for _, branch := range []string{"", "JKT01", "JKT02", "DLV01"} {
		t.Run("cabang "+branch, func(t *testing.T) {
			got, err := repo.FailedSummary(context.Background(), failedFilter(branch))
			require.NoError(t, err)
			var sum int
			var shares float64
			for _, r := range reasonsOf(got) {
				sum += r.Count
				shares += got.Share(r)
			}
			assert.Equal(t, got.TotalFailed, sum)
			if got.TotalFailed > 0 {
				assert.InDelta(t, 1.0, shares, 1e-9)
			}
		})
	}
}

func TestFailedSummary_DeliveredOnly(t *testing.T) {
	repo := seedFailedFixture(t)
	got, err := repo.FailedSummary(context.Background(), failedFilter("DLV01"))
	require.NoError(t, err)
	assert.Equal(t, 3, got.TotalPackages)
	assert.Zero(t, got.TotalFailed)
	assert.Zero(t, got.FailedRate())
	assert.NotNil(t, got.Reasons)
	assert.Empty(t, got.Reasons)
	assert.NotNil(t, got.TopCouriers)
	assert.Empty(t, got.TopCouriers)
}

func TestFailedSummary_Empty(t *testing.T) {
	repo := seedFailedFixture(t)
	got, err := repo.FailedSummary(context.Background(), failedFilter("TIDAKADA"))
	require.NoError(t, err)
	assert.Zero(t, got.TotalPackages)
	assert.Zero(t, got.TotalFailed)
	assert.Zero(t, got.FailedRate())
	assert.NotNil(t, got.Reasons)
	assert.Empty(t, got.Reasons)
	assert.NotNil(t, got.TopCouriers)
	assert.Empty(t, got.TopCouriers)
}

func TestFailedRepository_CountFailed(t *testing.T) {
	repo := seedFailedFixture(t)
	ctx := context.Background()

	n, err := repo.CountFailed(ctx, failedFilter(""))
	require.NoError(t, err)
	assert.Equal(t, 15, n)

	n, err = repo.CountFailed(ctx, failedFilter("JKT01"))
	require.NoError(t, err)
	assert.Equal(t, 10, n)

	n, err = repo.CountFailed(ctx, failedFilter("DLV01"))
	require.NoError(t, err)
	assert.Zero(t, n)
}
