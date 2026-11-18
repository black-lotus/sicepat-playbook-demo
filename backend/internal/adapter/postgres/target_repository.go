package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// TargetRepository mengimplementasikan port.TargetRepository.
type TargetRepository struct {
	pool *pgxpool.Pool
}

// NewTargetRepository membuat TargetRepository.
func NewTargetRepository(pool *pgxpool.Pool) *TargetRepository {
	return &TargetRepository{pool: pool}
}

const targetColumns = `branch_id, month, success_rate::float8, sla_rate::float8, delivered_packages`

const listTargetsSQL = `
SELECT ` + targetColumns + `
FROM branch_kpi_targets
WHERE branch_id = ANY($1) AND month BETWEEN $2::date AND $3::date
ORDER BY branch_id, month`

const getTargetSQL = `
SELECT ` + targetColumns + `
FROM branch_kpi_targets
WHERE branch_id = $1 AND month = $2::date`

// ListTargets mengembalikan target cabang untuk bulan from sampai to (inklusif).
func (r *TargetRepository) ListTargets(ctx context.Context, branchIDs []string, from, to time.Time) ([]domain.KPITarget, error) {
	if len(branchIDs) == 0 {
		return []domain.KPITarget{}, nil
	}
	rows, err := r.pool.Query(ctx, listTargetsSQL, branchIDs, from.Format(time.DateOnly), to.Format(time.DateOnly))
	if err != nil {
		return nil, fmt.Errorf("query branch_kpi_targets: %w", err)
	}
	defer rows.Close()

	targets := []domain.KPITarget{}
	for rows.Next() {
		t, err := scanTarget(rows)
		if err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca branch_kpi_targets: %w", err)
	}
	return targets, nil
}

// GetTarget mengembalikan target satu cabang pada satu bulan.
func (r *TargetRepository) GetTarget(ctx context.Context, branchID string, month time.Time) (domain.KPITarget, bool, error) {
	t, err := scanTarget(r.pool.QueryRow(ctx, getTargetSQL, branchID, month.Format(time.DateOnly)))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.KPITarget{}, false, nil
	}
	if err != nil {
		return domain.KPITarget{}, false, err
	}
	return t, true, nil
}

func scanTarget(row pgx.Row) (domain.KPITarget, error) {
	var t domain.KPITarget
	var month time.Time
	if err := row.Scan(&t.BranchID, &month, &t.SuccessRate, &t.SLARate, &t.DeliveredPackages); err != nil {
		return domain.KPITarget{}, fmt.Errorf("scan branch_kpi_targets: %w", err)
	}
	t.Month = time.Date(month.Year(), month.Month(), month.Day(), 0, 0, 0, 0, time.UTC)
	return t, nil
}
