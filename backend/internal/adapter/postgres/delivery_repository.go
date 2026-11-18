package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// DeliveryRepository mengimplementasikan port.DeliveryRepository.
type DeliveryRepository struct {
	pool *pgxpool.Pool
}

// NewDeliveryRepository membuat DeliveryRepository.
func NewDeliveryRepository(pool *pgxpool.Pool) *DeliveryRepository {
	return &DeliveryRepository{pool: pool}
}

// Event yang sama (awb, status, waktu) diabaikan agar consumer aman memproses ulang pesan.
const saveEventSQL = `
INSERT INTO delivery_events (awb_number, courier_id, branch_id, status, failed_reason, occurred_at, due_at)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7)
ON CONFLICT (awb_number, status, occurred_at) DO NOTHING`

// SaveEvent menyimpan satu event pengiriman secara idempoten.
func (r *DeliveryRepository) SaveEvent(ctx context.Context, e domain.DeliveryEvent) error {
	var dueAt *time.Time
	if !e.DueAt.IsZero() {
		dueAt = &e.DueAt
	}
	if _, err := r.pool.Exec(ctx, saveEventSQL,
		e.AWBNumber, e.CourierID, e.BranchID, string(e.Status), e.FailedReason, e.OccurredAt, dueAt,
	); err != nil {
		return fmt.Errorf("insert delivery_events: %w", err)
	}
	return nil
}
