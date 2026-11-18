package service

import (
	"context"
	"fmt"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/port"
)

// BranchPerformanceService menyajikan ranking performa cabang beserta evaluasi target.
type BranchPerformanceService struct {
	metrics port.MetricRepository
	targets port.TargetRepository
}

// NewBranchPerformanceService membuat BranchPerformanceService.
func NewBranchPerformanceService(metrics port.MetricRepository, targets port.TargetRepository) *BranchPerformanceService {
	return &BranchPerformanceService{metrics: metrics, targets: targets}
}

// List mengambil ranking cabang, lalu mengevaluasi target periode untuk cabang di halaman
// dengan satu kali ListTargets (bulan start sampai bulan end).
func (s *BranchPerformanceService) List(ctx context.Context, f domain.MetricFilter) (domain.BranchPerformanceResultPage, error) {
	f = f.WithDefaults()
	if err := f.Validate(); err != nil {
		return domain.BranchPerformanceResultPage{}, err
	}
	page, err := s.metrics.ListBranchPerformance(ctx, f)
	if err != nil {
		return domain.BranchPerformanceResultPage{}, fmt.Errorf("ambil ranking cabang: %w", err)
	}
	result := domain.BranchPerformanceResultPage{
		Items: make([]domain.BranchPerformanceWithTarget, 0, len(page.Items)),
		Total: page.Total,
	}
	if len(page.Items) == 0 {
		return result, nil
	}

	ids := make([]string, 0, len(page.Items))
	for _, item := range page.Items {
		ids = append(ids, item.BranchID)
	}
	targets, err := s.targets.ListTargets(ctx, ids, domain.MonthOf(f.StartDate), domain.MonthOf(f.EndDate))
	if err != nil {
		return domain.BranchPerformanceResultPage{}, fmt.Errorf("ambil target cabang: %w", err)
	}
	byBranch := make(map[string][]domain.KPITarget, len(ids))
	for _, t := range targets {
		byBranch[t.BranchID] = append(byBranch[t.BranchID], t)
	}
	for _, item := range page.Items {
		pt := domain.TargetForPeriod(byBranch[item.BranchID], f.StartDate, f.EndDate)
		result.Items = append(result.Items, domain.BranchPerformanceWithTarget{
			BranchPerformance: item,
			Target:            domain.EvaluateTarget(item, pt),
		})
	}
	return result, nil
}
