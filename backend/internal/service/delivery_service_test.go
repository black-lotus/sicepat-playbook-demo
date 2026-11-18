package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/port/mocks"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/service"
)

func TestDeliveryService_RecordEvent(t *testing.T) {
	valid := domain.DeliveryEvent{
		AWBNumber:  "002345678901",
		CourierID:  "C001",
		BranchID:   "JKT01",
		Status:     domain.DeliveryStatusDelivered,
		OccurredAt: time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC),
	}
	errDB := errors.New("db down")

	tests := []struct {
		name    string
		event   domain.DeliveryEvent
		setup   func(repo *mocks.DeliveryRepository)
		wantErr error
	}{
		{
			name:  "event valid disimpan",
			event: valid,
			setup: func(repo *mocks.DeliveryRepository) {
				repo.EXPECT().SaveEvent(mock.Anything, valid).Return(nil)
			},
		},
		{
			name:    "event tidak valid tidak disimpan",
			event:   domain.DeliveryEvent{},
			setup:   func(*mocks.DeliveryRepository) {},
			wantErr: domain.ErrInvalidEvent,
		},
		{
			name:  "error repository dikembalikan",
			event: valid,
			setup: func(repo *mocks.DeliveryRepository) {
				repo.EXPECT().SaveEvent(mock.Anything, valid).Return(errDB)
			},
			wantErr: errDB,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewDeliveryRepository(t)
			tt.setup(repo)

			err := service.NewDeliveryService(repo).RecordEvent(context.Background(), tt.event)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
