//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/postgres"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

func TestBranchKPITargets_Constraints(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	_, err := pool.Exec(ctx, "TRUNCATE branch_kpi_targets")
	require.NoError(t, err)

	const insert = `INSERT INTO branch_kpi_targets (branch_id, month, success_rate, sla_rate, delivered_packages)
		VALUES ($1, $2::date, $3, $4, $5)`

	t.Run("baris valid diterima, termasuk batas 0 dan 1", func(t *testing.T) {
		_, err := pool.Exec(ctx, insert, "JKT01", "2026-11-01", 0.9, 0.85, 3000)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, insert, "JKT01", "2026-12-01", 1, 0, 0)
		require.NoError(t, err)
	})

	rejected := []struct {
		name      string
		branch    string
		month     string
		success   float64
		sla       float64
		delivered int
	}{
		{"success_rate di atas 1", "JKT02", "2026-11-01", 1.1, 0.85, 100},
		{"success_rate negatif", "JKT02", "2026-11-01", -0.1, 0.85, 100},
		{"sla_rate di atas 1", "JKT02", "2026-11-01", 0.9, 1.1, 100},
		{"sla_rate negatif", "JKT02", "2026-11-01", 0.9, -0.1, 100},
		{"delivered_packages negatif", "JKT02", "2026-11-01", 0.9, 0.85, -1},
		{"month bukan hari pertama bulan", "JKT02", "2026-11-15", 0.9, 0.85, 100},
		{"primary key ganda (branch_id, month)", "JKT01", "2026-11-01", 0.8, 0.8, 200},
	}
	for _, tc := range rejected {
		t.Run("ditolak: "+tc.name, func(t *testing.T) {
			_, err := pool.Exec(ctx, insert, tc.branch, tc.month, tc.success, tc.sla, tc.delivered)
			assert.Error(t, err)
		})
	}

	var n int
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM branch_kpi_targets").Scan(&n))
	assert.Equal(t, 2, n, "baris yang ditolak tidak tersimpan")
}

func TestTargetRepository_GetAndList(t *testing.T) {
	pool := openTestPool(t)
	ctx := context.Background()
	repo := postgres.NewTargetRepository(pool)

	const insert = `INSERT INTO branch_kpi_targets (branch_id, month, success_rate, sla_rate, delivered_packages)
		VALUES ($1, $2::date, $3, $4, $5)`
	rows := []struct {
		branch, month string
		success, sla  float64
		delivered     int
	}{
		{"JKT01", "2026-10-01", 0.88, 0.8, 2900},
		{"JKT01", "2026-11-01", 0.9, 0.85, 3000},
		{"JKT02", "2026-10-01", 0.87, 0.82, 2000},
		{"JKT02", "2026-11-01", 0.91, 0.86, 2100},
	}
	for _, r := range rows {
		_, err := pool.Exec(ctx, insert, r.branch, r.month, r.success, r.sla, r.delivered)
		require.NoError(t, err)
	}
	oct := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	nov := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)

	t.Run("get ada", func(t *testing.T) {
		got, found, err := repo.GetTarget(ctx, "JKT01", nov)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, domain.KPITarget{BranchID: "JKT01", Month: nov, SuccessRate: 0.9, SLARate: 0.85, DeliveredPackages: 3000}, got)
	})
	t.Run("get tidak ada", func(t *testing.T) {
		_, found, err := repo.GetTarget(ctx, "BDG01", nov)
		require.NoError(t, err)
		assert.False(t, found)
		_, found, err = repo.GetTarget(ctx, "JKT01", time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		assert.False(t, found)
	})
	t.Run("list dua cabang dua bulan terurut", func(t *testing.T) {
		got, err := repo.ListTargets(ctx, []string{"JKT02", "JKT01", "BDG01"}, oct, nov)
		require.NoError(t, err)
		require.Len(t, got, 4)
		assert.Equal(t, []string{"JKT01", "JKT01", "JKT02", "JKT02"}, []string{got[0].BranchID, got[1].BranchID, got[2].BranchID, got[3].BranchID})
		assert.Equal(t, oct, got[0].Month)
		assert.Equal(t, nov, got[1].Month)
	})
	t.Run("list satu bulan dan satu cabang", func(t *testing.T) {
		got, err := repo.ListTargets(ctx, []string{"JKT02"}, nov, nov)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, 2100, got[0].DeliveredPackages)
	})
	t.Run("list tanpa cabang kosong bukan nil", func(t *testing.T) {
		got, err := repo.ListTargets(ctx, nil, oct, nov)
		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})
}
