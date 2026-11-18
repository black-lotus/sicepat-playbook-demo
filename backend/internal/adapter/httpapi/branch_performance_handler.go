package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// BranchPerformanceLister adalah use case yang dibutuhkan handler ranking performa cabang.
type BranchPerformanceLister interface {
	List(ctx context.Context, f domain.MetricFilter) (domain.BranchPerformanceResultPage, error)
}

// BranchPerformanceHandler menyajikan endpoint ranking performa cabang.
type BranchPerformanceHandler struct {
	svc    BranchPerformanceLister
	logger *slog.Logger
}

// NewBranchPerformanceHandler membuat BranchPerformanceHandler.
func NewBranchPerformanceHandler(svc BranchPerformanceLister, logger *slog.Logger) *BranchPerformanceHandler {
	return &BranchPerformanceHandler{svc: svc, logger: logger}
}

type branchTargetResponse struct {
	SuccessRate       float64 `json:"success_rate"`
	SLARate           float64 `json:"sla_rate"`
	DeliveredPackages int     `json:"delivered_packages"`
	SuccessRateStatus string  `json:"success_rate_status"`
	SLARateStatus     string  `json:"sla_rate_status"`
	DeliveredStatus   string  `json:"delivered_status"`
}

type branchPerformanceResponse struct {
	Rank          int                   `json:"rank"`
	BranchID      string                `json:"branch_id"`
	CourierCount  int                   `json:"courier_count"`
	TotalPackages int                   `json:"total_packages"`
	Delivered     int                   `json:"delivered"`
	Failed        int                   `json:"failed"`
	OnTime        int                   `json:"on_time"`
	SuccessRate   float64               `json:"success_rate"`
	SLARate       float64               `json:"sla_rate"`
	Target        *branchTargetResponse `json:"target"`
	Daily         []dailyCountResponse  `json:"daily"`
}

func toBranchPerformanceResponse(p domain.BranchPerformanceWithTarget) branchPerformanceResponse {
	daily := make([]dailyCountResponse, 0, len(p.Daily))
	for _, d := range p.Daily {
		daily = append(daily, dailyCountResponse{
			Date:          d.Date.Format(time.DateOnly),
			TotalPackages: d.TotalPackages,
			Delivered:     d.Delivered,
			Failed:        d.Failed,
			OnTime:        d.OnTime,
		})
	}
	var target *branchTargetResponse
	if p.Target != nil {
		target = &branchTargetResponse{
			SuccessRate:       p.Target.Target.SuccessRate,
			SLARate:           p.Target.Target.SLARate,
			DeliveredPackages: p.Target.Target.DeliveredPackages,
			SuccessRateStatus: string(p.Target.SuccessRateStatus),
			SLARateStatus:     string(p.Target.SLARateStatus),
			DeliveredStatus:   string(p.Target.DeliveredStatus),
		}
	}
	return branchPerformanceResponse{
		Rank:          p.Rank,
		BranchID:      p.BranchID,
		CourierCount:  p.CourierCount,
		TotalPackages: p.TotalPackages,
		Delivered:     p.Delivered,
		Failed:        p.Failed,
		OnTime:        p.OnTime,
		SuccessRate:   p.SuccessRate(),
		SLARate:       p.SLARate(),
		Target:        target,
		Daily:         daily,
	}
}

// List menangani GET /api/v1/branch-performance. Parameter branch_id diabaikan.
func (h *BranchPerformanceHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseMetricFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	page, err := h.svc.List(r.Context(), filter)
	if err != nil {
		writeDomainError(w, r, h.logger, err)
		return
	}
	items := make([]branchPerformanceResponse, 0, len(page.Items))
	for _, p := range page.Items {
		items = append(items, toBranchPerformanceResponse(p))
	}
	filter = filter.WithDefaults()
	writeData(w, items, pageMeta{Page: filter.Page, PageSize: filter.PageSize, Total: page.Total})
}
