// Package kafka berisi consumer event pengiriman dari Kafka (segmentio/kafka-go).
package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// EventRecorder adalah use case yang dibutuhkan consumer.
type EventRecorder interface {
	RecordEvent(ctx context.Context, event domain.DeliveryEvent) error
}

// DeliveryConsumer membaca event pengiriman dari Kafka lalu mencatatnya.
type DeliveryConsumer struct {
	reader   *kafkago.Reader
	recorder EventRecorder
	logger   *slog.Logger
}

// NewDeliveryConsumer membuat DeliveryConsumer dengan consumer group.
func NewDeliveryConsumer(brokers []string, topic, groupID string, recorder EventRecorder, logger *slog.Logger) *DeliveryConsumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	return &DeliveryConsumer{reader: reader, recorder: recorder, logger: logger}
}

// deliveryMessage adalah bentuk JSON pesan di topic Kafka.
type deliveryMessage struct {
	AWBNumber    string     `json:"awb_number"`
	CourierID    string     `json:"courier_id"`
	BranchID     string     `json:"branch_id"`
	Status       string     `json:"status"`
	FailedReason string     `json:"failed_reason"`
	OccurredAt   time.Time  `json:"occurred_at"`
	DueAt        *time.Time `json:"due_at"`
}

func (m deliveryMessage) toDomain() domain.DeliveryEvent {
	e := domain.DeliveryEvent{
		AWBNumber:    m.AWBNumber,
		CourierID:    m.CourierID,
		BranchID:     m.BranchID,
		Status:       domain.DeliveryStatus(m.Status),
		FailedReason: m.FailedReason,
		OccurredAt:   m.OccurredAt.UTC(),
	}
	if m.DueAt != nil {
		e.DueAt = m.DueAt.UTC()
	}
	return e
}

// Run memproses pesan sampai ctx dibatalkan.
// Pesan yang tidak valid dicatat lalu di-commit agar tidak memblokir partisi.
// Error simpan dikembalikan tanpa commit, sehingga pesan diproses ulang setelah consumer restart.
func (c *DeliveryConsumer) Run(ctx context.Context) error {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("fetch pesan kafka: %w", err)
		}

		if err := c.handle(ctx, msg); err != nil {
			return err
		}
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("commit offset %d: %w", msg.Offset, err)
		}
	}
}

func (c *DeliveryConsumer) handle(ctx context.Context, msg kafkago.Message) error {
	var payload deliveryMessage
	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		c.logger.WarnContext(ctx, "pesan kafka bukan JSON valid, dilewati",
			slog.Int("partition", msg.Partition), slog.Int64("offset", msg.Offset))
		return nil
	}

	err := c.recorder.RecordEvent(ctx, payload.toDomain())
	if errors.Is(err, domain.ErrInvalidEvent) {
		c.logger.WarnContext(ctx, "event tidak valid, dilewati",
			slog.Int64("offset", msg.Offset), slog.Any("error", err))
		return nil
	}
	if err != nil {
		return fmt.Errorf("proses offset %d: %w", msg.Offset, err)
	}
	return nil
}

// Close menutup koneksi reader.
func (c *DeliveryConsumer) Close() error {
	return c.reader.Close()
}
