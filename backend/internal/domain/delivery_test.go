package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

func TestDeliveryEvent_Validate(t *testing.T) {
	now := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	valid := domain.DeliveryEvent{
		AWBNumber:  "002345678901",
		CourierID:  "C001",
		BranchID:   "JKT01",
		Status:     domain.DeliveryStatusDelivered,
		OccurredAt: now,
	}

	tests := []struct {
		name    string
		modify  func(e *domain.DeliveryEvent)
		wantErr bool
	}{
		{name: "delivered valid", modify: func(*domain.DeliveryEvent) {}},
		{name: "failed dengan alasan", modify: func(e *domain.DeliveryEvent) {
			e.Status = domain.DeliveryStatusFailed
			e.FailedReason = "penerima tidak di tempat"
		}},
		{name: "failed tanpa alasan", modify: func(e *domain.DeliveryEvent) { e.Status = domain.DeliveryStatusFailed }, wantErr: true},
		{name: "tanpa awb", modify: func(e *domain.DeliveryEvent) { e.AWBNumber = "" }, wantErr: true},
		{name: "tanpa courier", modify: func(e *domain.DeliveryEvent) { e.CourierID = "" }, wantErr: true},
		{name: "tanpa waktu", modify: func(e *domain.DeliveryEvent) { e.OccurredAt = time.Time{} }, wantErr: true},
		{name: "status tidak dikenal", modify: func(e *domain.DeliveryEvent) { e.Status = "lost" }, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := valid
			tt.modify(&e)
			err := e.Validate()
			if tt.wantErr {
				assert.ErrorIs(t, err, domain.ErrInvalidEvent)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestDeliveryEvent_IsOnTime(t *testing.T) {
	due := time.Date(2026, 11, 18, 17, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		event domain.DeliveryEvent
		want  bool
	}{
		{name: "sebelum due", event: domain.DeliveryEvent{Status: domain.DeliveryStatusDelivered, OccurredAt: due.Add(-time.Hour), DueAt: due}, want: true},
		{name: "tepat due", event: domain.DeliveryEvent{Status: domain.DeliveryStatusDelivered, OccurredAt: due, DueAt: due}, want: true},
		{name: "terlambat", event: domain.DeliveryEvent{Status: domain.DeliveryStatusDelivered, OccurredAt: due.Add(time.Minute), DueAt: due}},
		{name: "tanpa due", event: domain.DeliveryEvent{Status: domain.DeliveryStatusDelivered, OccurredAt: due}},
		{name: "gagal kirim", event: domain.DeliveryEvent{Status: domain.DeliveryStatusFailed, OccurredAt: due.Add(-time.Hour), DueAt: due}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.event.IsOnTime())
		})
	}
}
