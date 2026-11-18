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

func TestDeliveryEvent_WithDefaultDueAt(t *testing.T) {
	wib := func(month time.Month, day, hour int) time.Time {
		return time.Date(2026, month, day, hour, 0, 0, 0, domain.WIB)
	}
	base := domain.DeliveryEvent{Status: domain.DeliveryStatusDelivered}

	t.Run("mengisi DueAt dari AssignedAt", func(t *testing.T) {
		e := base
		e.AssignedAt = wib(11, 7, 9)
		got := e.WithDefaultDueAt()
		assert.True(t, time.Date(2026, 11, 7, 23, 59, 59, 0, domain.WIB).Equal(got.DueAt))
		assert.Equal(t, time.UTC, got.DueAt.Location())
	})

	t.Run("DueAt dari sumber tidak diubah", func(t *testing.T) {
		e := base
		e.AssignedAt = wib(11, 7, 9)
		e.DueAt = wib(11, 9, 12)
		assert.Equal(t, e, e.WithDefaultDueAt())
	})

	t.Run("tanpa AssignedAt dan DueAt tidak pernah tepat SLA", func(t *testing.T) {
		e := base
		e.OccurredAt = wib(11, 7, 10)
		got := e.WithDefaultDueAt()
		assert.True(t, got.DueAt.IsZero())
		assert.False(t, got.IsOnTime())
	})

	t.Run("AC6 SLA dihitung dari cut-off", func(t *testing.T) {
		tests := []struct {
			name       string
			assignedAt time.Time
			occurredAt time.Time
			wantOnTime bool
		}{
			{name: "ditugaskan pagi, tiba lewat tengah malam", assignedAt: wib(11, 7, 9), occurredAt: wib(11, 8, 0), wantOnTime: false},
			{name: "ditugaskan 16:00, tiba tepat 16:00 besok", assignedAt: wib(11, 7, 16), occurredAt: wib(11, 8, 16), wantOnTime: true},
			{name: "ditugaskan pagi, tiba 23:59:59", assignedAt: wib(11, 7, 9), occurredAt: time.Date(2026, 11, 7, 23, 59, 59, 0, domain.WIB), wantOnTime: true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				e := base
				e.AssignedAt = tt.assignedAt
				e.OccurredAt = tt.occurredAt
				assert.Equal(t, tt.wantOnTime, e.WithDefaultDueAt().IsOnTime())
			})
		}
	})
}
