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

// CourierPerformanceHandler menyajikan endpoint ranking performa kurir.
type CourierPerformanceHandler struct {
	svc    CourierPerformanceLister
	logger *slog.Logger
}

// NewCourierPerformanceHandler membuat CourierPerformanceHandler.
func NewCourierPerformanceHandler(svc CourierPerformanceLister, logger *slog.Logger) *CourierPerformanceHandler {
	return &CourierPerformanceHandler{svc: svc, logger: logger}
}

type dailyCountResponse struct {
	Date          string `json:"date"`
	TotalPackages int    `json:"total_packages"`
	Delivered     int    `json:"delivered"`
	Failed        int    `json:"failed"`
	OnTime        int    `json:"on_time"`
}

type courierPerformanceResponse struct {
	Rank           int                  `json:"rank"`
	CourierID      string               `json:"courier_id"`
	CourierName    string               `json:"courier_name"`
	BranchID       string               `json:"branch_id"`
	TotalPackages  int                  `json:"total_packages"`
	Delivered      int                  `json:"delivered"`
	Failed         int                  `json:"failed"`
	OnTime         int                  `json:"on_time"`
	SuccessRate    float64              `json:"success_rate"`
	SLARate        float64              `json:"sla_rate"`
	SuccessRateGap float64              `json:"success_rate_gap"`
	Daily          []dailyCountResponse `json:"daily"`
}

func toCourierPerformanceResponse(p domain.CourierPerformance) courierPerformanceResponse {
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
		Rank:           p.Rank,
		CourierID:      p.CourierID,
		CourierName:    p.CourierName,
		BranchID:       p.BranchID,
		TotalPackages:  p.TotalPackages,
		Delivered:      p.Delivered,
		Failed:         p.Failed,
		OnTime:         p.OnTime,
		SuccessRate:    p.SuccessRate(),
		SLARate:        p.SLARate(),
		SuccessRateGap: p.SuccessRateGap(),
		Daily:          daily,
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
	items := make([]courierPerformanceResponse, 0, len(page.Items))
	for _, p := range page.Items {
		items = append(items, toCourierPerformanceResponse(p))
	}
	filter = filter.WithDefaults()
	writeData(w, items, pageMeta{Page: filter.Page, PageSize: filter.PageSize, Total: page.Total})
}
