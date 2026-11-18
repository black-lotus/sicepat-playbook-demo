package service

import (
	"context"
	"fmt"
	"time"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/port"
)

// TargetService menyajikan target KPI bulanan per cabang.
type TargetService struct {
	repo port.TargetRepository
}

// NewTargetService membuat TargetService.
func NewTargetService(repo port.TargetRepository) *TargetService {
	return &TargetService{repo: repo}
}

// GetTarget mengembalikan target cabang pada bulan tertentu, atau domain.ErrTargetNotFound.
func (s *TargetService) GetTarget(ctx context.Context, branchID string, month time.Time) (domain.KPITarget, error) {
	t, found, err := s.repo.GetTarget(ctx, branchID, month)
	if err != nil {
		return domain.KPITarget{}, fmt.Errorf("ambil target kpi: %w", err)
	}
	if !found {
		return domain.KPITarget{}, domain.ErrTargetNotFound
	}
	return t, nil
}

// RateTarget mengembalikan target bulan tanggal end (WIB), atau nil bila cabang tidak punya target.
func (s *TargetService) RateTarget(ctx context.Context, branchID string, end time.Time) (*domain.KPITarget, error) {
	t, found, err := s.repo.GetTarget(ctx, branchID, domain.MonthOf(end))
	if err != nil {
		return nil, fmt.Errorf("ambil target kpi: %w", err)
	}
	if !found {
		return nil, nil
	}
	return &t, nil
}
