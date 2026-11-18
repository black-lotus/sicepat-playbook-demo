package service

import (
	"context"
	"fmt"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/port"
)

// DeliveryService mencatat event pengiriman paket dari kurir.
type DeliveryService struct {
	repo port.DeliveryRepository
}

// NewDeliveryService membuat DeliveryService.
func NewDeliveryService(repo port.DeliveryRepository) *DeliveryService {
	return &DeliveryService{repo: repo}
}

// RecordEvent memvalidasi lalu menyimpan event pengiriman.
func (s *DeliveryService) RecordEvent(ctx context.Context, event domain.DeliveryEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}
	if err := s.repo.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("simpan event %s: %w", event.AWBNumber, err)
	}
	return nil
}
