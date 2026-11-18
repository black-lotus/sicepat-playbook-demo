package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// CourierPerformanceLister adalah use case yang dibutuhkan handler ranking performa kurir.
type CourierPerformanceLister interface {
	ListCourierPerformance(ctx context.Context, filter domain.MetricFilter) (domain.CourierPerformancePage, error)
}

// RateTargetFinder mencari target KPI cabang untuk bulan tanggal end (WIB); nil bila tidak ada.
type RateTargetFinder interface {
	RateTarget(ctx context.Context, branchID string, end time.Time) (*domain.KPITarget, error)
}

// CourierPerformanceHandler menyajikan endpoint ranking performa kurir.
type CourierPerformanceHandler struct {
	svc     CourierPerformanceLister
	targets RateTargetFinder
	logger  *slog.Logger
}

// NewCourierPerformanceHandler membuat CourierPerformanceHandler.
func NewCourierPerformanceHandler(svc CourierPerformanceLister, targets RateTargetFinder, logger *slog.Logger) *CourierPerformanceHandler {
	return &CourierPerformanceHandler{svc: svc, targets: targets, logger: logger}
}

// branchTargetMeta adalah target rate cabang untuk bulan end_date.
type branchTargetMeta struct {
	Month       string  `json:"month"`
	SuccessRate float64 `json:"success_rate"`
	SLARate     float64 `json:"sla_rate"`
}

// courierPerformanceMeta menambah branch_target ke meta halaman. Didefinisikan di sini supaya
// endpoint lain yang memakai pageMeta tidak ikut mengirim branch_target.
type courierPerformanceMeta struct {
	pageMeta
	BranchTarget *branchTargetMeta `json:"branch_target"`
}

type dailyCountResponse struct {
	Date          string `json:"date"`
	TotalPackages int    `json:"total_packages"`
	Delivered     int    `json:"delivered"`
	Failed        int    `json:"failed"`
	OnTime        int    `json:"on_time"`
}

type courierPerformanceResponse struct {
	Rank           int     `json:"rank"`
	CourierID      string  `json:"courier_id"`
	CourierName    string  `json:"courier_name"`
	BranchID       string  `json:"branch_id"`
	TotalPackages  int     `json:"total_packages"`
	Delivered      int     `json:"delivered"`
	Failed         int     `json:"failed"`
	OnTime         int     `json:"on_time"`
	SuccessRate    float64 `json:"success_rate"`
	SLARate        float64 `json:"sla_rate"`
	SuccessRateGap float64 `json:"success_rate_gap"`
	// SuccessRateStatus null bila cabang tanpa target atau branch_id kosong.
	SuccessRateStatus *domain.TargetStatus `json:"success_rate_status"`
	Daily             []dailyCountResponse `json:"daily"`
}

func toCourierPerformanceResponse(p domain.CourierPerformance, target *domain.KPITarget) courierPerformanceResponse {
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
	return courierPerformanceResponse{
		Rank:              p.Rank,
		CourierID:         p.CourierID,
		CourierName:       p.CourierName,
		BranchID:          p.BranchID,
		TotalPackages:     p.TotalPackages,
		Delivered:         p.Delivered,
		Failed:            p.Failed,
		OnTime:            p.OnTime,
		SuccessRate:       p.SuccessRate(),
		SLARate:           p.SLARate(),
		SuccessRateGap:    p.SuccessRateGap(),
		SuccessRateStatus: p.SuccessRateStatus(target),
		Daily:             daily,
	}
}

// List menangani GET /api/v1/courier-performance.
func (h *CourierPerformanceHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseMetricFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}
	page, err := h.svc.ListCourierPerformance(r.Context(), filter)
	if err != nil {
		writeDomainError(w, r, h.logger, err)
		return
	}
	var target *domain.KPITarget
	if filter.BranchID != "" {
		target, err = h.targets.RateTarget(r.Context(), filter.BranchID, filter.EndDate)
		if err != nil {
			writeDomainError(w, r, h.logger, err)
			return
		}
	}
	items := make([]courierPerformanceResponse, 0, len(page.Items))
	for _, p := range page.Items {
		items = append(items, toCourierPerformanceResponse(p, target))
	}
	filter = filter.WithDefaults()
	meta := courierPerformanceMeta{pageMeta: pageMeta{Page: filter.Page, PageSize: filter.PageSize, Total: page.Total}}
	if target != nil {
		meta.BranchTarget = &branchTargetMeta{
			Month:       target.Month.Format(monthLayout),
			SuccessRate: target.SuccessRate,
			SLARate:     target.SLARate,
		}
	}
	writeData(w, items, meta)
}
