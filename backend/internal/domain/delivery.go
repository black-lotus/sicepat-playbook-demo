// Package domain berisi entity dan aturan bisnis monitoring performa SiGesit, tanpa dependensi ke infrastruktur.
package domain

import (
	"fmt"
	"time"
)

// DeliveryStatus adalah status akhir percobaan pengiriman paket.
type DeliveryStatus string

const (
	// DeliveryStatusDelivered berarti paket diterima.
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	// DeliveryStatusFailed berarti percobaan pengiriman gagal.
	DeliveryStatusFailed DeliveryStatus = "failed"
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
	// AssignedAt adalah waktu paket ditugaskan ke kurir. Opsional; dipakai menghitung DueAt
	// bila sumber tidak mengirim batas SLA.
	AssignedAt time.Time
	DueAt      time.Time
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

// WithDefaultDueAt mengisi DueAt dari cut-off SLA (SLADeadline) bila DueAt kosong dan AssignedAt ada.
// DueAt dari sumber tidak pernah diubah. Tanpa keduanya DueAt tetap kosong, sehingga paket tidak
// dihitung tepat SLA.
func (e DeliveryEvent) WithDefaultDueAt() DeliveryEvent {
	if e.DueAt.IsZero() && !e.AssignedAt.IsZero() {
		e.DueAt = SLADeadline(e.AssignedAt).UTC()
	}
	return e
}
