package service

import (
	"context"
	"fmt"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/port"
)

// FailedDeliveryService menyajikan ringkasan gagal kirim.
type FailedDeliveryService struct {
	repo port.FailedRepository
}

// NewFailedDeliveryService membuat FailedDeliveryService.
func NewFailedDeliveryService(repo port.FailedRepository) *FailedDeliveryService {
	return &FailedDeliveryService{repo: repo}
}

// Summary memvalidasi filter, mengambil ringkasan, lalu mengisi total gagal periode pembanding.
func (s *FailedDeliveryService) Summary(ctx context.Context, f domain.FailedFilter) (domain.FailedSummary, error) {
	if err := f.Validate(); err != nil {
		return domain.FailedSummary{}, err
	}
	sum, err := s.repo.FailedSummary(ctx, f)
	if err != nil {
		return domain.FailedSummary{}, fmt.Errorf("ambil ringkasan gagal kirim: %w", err)
	}
	prevStart, prevEnd := domain.PreviousPeriod(f.StartDate, f.EndDate)
	prev, err := s.repo.CountFailed(ctx, domain.FailedFilter{StartDate: prevStart, EndDate: prevEnd, BranchID: f.BranchID})
	if err != nil {
		return domain.FailedSummary{}, fmt.Errorf("hitung gagal periode pembanding: %w", err)
	}
	sum.PreviousTotalFailed = prev
	return sum, nil
}
