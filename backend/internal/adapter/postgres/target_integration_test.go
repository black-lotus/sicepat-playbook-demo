//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
