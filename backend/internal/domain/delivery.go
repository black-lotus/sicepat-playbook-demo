package domain

import (
	"fmt"
	"time"
)

// DeliveryStatus adalah status akhir percobaan pengiriman paket.
type DeliveryStatus string

const (
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
)

// DeliveryEvent adalah satu percobaan pengiriman paket oleh kurir.
// Event ini sengaja tidak menyimpan data pribadi penerima.
type DeliveryEvent struct {
	AWBNumber    string
	CourierID    string
	BranchID     string
	Status       DeliveryStatus
	FailedReason string
	OccurredAt   time.Time
	DueAt        time.Time
}

// Validate memastikan event lengkap sebelum disimpan.
func (e DeliveryEvent) Validate() error {
	switch {
	case e.AWBNumber == "":
		return fmt.Errorf("%w: awb_number wajib diisi", ErrInvalidEvent)
	case e.CourierID == "" || e.BranchID == "":
		return fmt.Errorf("%w: courier_id dan branch_id wajib diisi", ErrInvalidEvent)
	case e.OccurredAt.IsZero():
		return fmt.Errorf("%w: occurred_at wajib diisi", ErrInvalidEvent)
	}
	switch e.Status {
	case DeliveryStatusDelivered:
		return nil
	case DeliveryStatusFailed:
		if e.FailedReason == "" {
			return fmt.Errorf("%w: failed_reason wajib diisi untuk status failed", ErrInvalidEvent)
		}
		return nil
	default:
		return fmt.Errorf("%w: status %q tidak dikenal", ErrInvalidEvent, e.Status)
	}
}

// IsOnTime menandakan paket terkirim sebelum atau tepat pada batas SLA.
func (e DeliveryEvent) IsOnTime() bool {
	return e.Status == DeliveryStatusDelivered && !e.DueAt.IsZero() && !e.OccurredAt.After(e.DueAt)
}
