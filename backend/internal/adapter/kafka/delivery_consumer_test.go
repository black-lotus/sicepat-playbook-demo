package kafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

type fakeRecorder struct {
	got []domain.DeliveryEvent
	err error
}

func (f *fakeRecorder) RecordEvent(_ context.Context, e domain.DeliveryEvent) error {
	f.got = append(f.got, e)
	return f.err
}

func TestDeliveryConsumer_handle(t *testing.T) {
	errDB := errors.New("db down")
	validJSON := `{"awb_number":"A1","courier_id":"C001","branch_id":"JKT01","status":"delivered",
		"occurred_at":"2026-11-18T09:00:00+07:00","due_at":"2026-11-18T17:00:00+07:00"}`

	tests := []struct {
		name       string
		value      string
		recordErr  error
		wantErr    error
		wantRecord *domain.DeliveryEvent
	}{
		{
			name:  "pesan valid dikonversi ke UTC",
			value: validJSON,
			wantRecord: &domain.DeliveryEvent{
				AWBNumber:  "A1",
				CourierID:  "C001",
				BranchID:   "JKT01",
				Status:     domain.DeliveryStatusDelivered,
				OccurredAt: time.Date(2026, 11, 18, 2, 0, 0, 0, time.UTC),
				DueAt:      time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC),
			},
		},
		{name: "bukan JSON dilewati", value: "bukan-json"},
		{name: "event tidak valid dilewati", value: validJSON, recordErr: fmt.Errorf("%w: x", domain.ErrInvalidEvent)},
		{name: "error simpan dikembalikan", value: validJSON, recordErr: errDB, wantErr: errDB},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := &fakeRecorder{err: tt.recordErr}
			c := &DeliveryConsumer{recorder: recorder, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}

			err := c.handle(context.Background(), kafkago.Message{Value: []byte(tt.value)})

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantRecord != nil {
				assert.Equal(t, []domain.DeliveryEvent{*tt.wantRecord}, recorder.got)
			}
		})
	}
}
