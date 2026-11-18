package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// FailedSummarizer adalah use case yang dibutuhkan handler ringkasan gagal kirim.
type FailedSummarizer interface {
	Summary(ctx context.Context, f domain.FailedFilter) (domain.FailedSummary, error)
}

// FailedDeliveryHandler menyajikan endpoint ringkasan gagal kirim.
type FailedDeliveryHandler struct {
	svc    FailedSummarizer
	logger *slog.Logger
}

// NewFailedDeliveryHandler membuat FailedDeliveryHandler.
func NewFailedDeliveryHandler(svc FailedSummarizer, logger *slog.Logger) *FailedDeliveryHandler {
	return &FailedDeliveryHandler{svc: svc, logger: logger}
}

type failedReasonResponse struct {
	Reason string  `json:"reason"`
	Count  int     `json:"count"`
	Share  float64 `json:"share"`
}

type failedCourierResponse struct {
	Rank          int     `json:"rank"`
	CourierID     string  `json:"courier_id"`
	CourierName   string  `json:"courier_name"`
	BranchID      string  `json:"branch_id"`
	TotalPackages int     `json:"total_packages"`
	Failed        int     `json:"failed"`
	FailedRate    float64 `json:"failed_rate"`
}

type failedSummaryResponse struct {
	TotalPackages       int                     `json:"total_packages"`
	TotalFailed         int                     `json:"total_failed"`
	FailedRate          float64                 `json:"failed_rate"`
	PreviousTotalFailed int                     `json:"previous_total_failed"`
	FailedChangeRate    *float64                `json:"failed_change_rate"`
	Reasons             []failedReasonResponse  `json:"reasons"`
	TopCouriers         []failedCourierResponse `json:"top_couriers"`
}

func toFailedSummaryResponse(s domain.FailedSummary) failedSummaryResponse {
	reasons := make([]failedReasonResponse, 0, len(s.Reasons))
	for _, r := range s.Reasons {
		reasons = append(reasons, failedReasonResponse{Reason: r.Reason, Count: r.Count, Share: s.Share(r)})
	}
	couriers := make([]failedCourierResponse, 0, len(s.TopCouriers))
	for _, c := range s.TopCouriers {
		couriers = append(couriers, failedCourierResponse{
			Rank: c.Rank, CourierID: c.CourierID, CourierName: c.CourierName, BranchID: c.BranchID,
			TotalPackages: c.TotalPackages, Failed: c.Failed, FailedRate: c.FailedRate(),
		})
	}
	return failedSummaryResponse{
		TotalPackages:       s.TotalPackages,
		TotalFailed:         s.TotalFailed,
		FailedRate:          s.FailedRate(),
		PreviousTotalFailed: s.PreviousTotalFailed,
		FailedChangeRate:    s.FailedChangeRate(),
		Reasons:             reasons,
		TopCouriers:         couriers,
	}
}

// Summary menangani GET /api/v1/failed-deliveries/summary.
func (h *FailedDeliveryHandler) Summary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var f domain.FailedFilter
	var err error
	if f.StartDate, err = parseDate(q.Get("start_date")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", queryError{"start_date"}.Error())
		return
	}
	if f.EndDate, err = parseDate(q.Get("end_date")); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", queryError{"end_date"}.Error())
		return
	}
	f.BranchID = q.Get("branch_id")

	sum, err := h.svc.Summary(r.Context(), f)
	if err != nil {
		writeDomainError(w, r, h.logger, err)
		return
	}
	writeData(w, toFailedSummaryResponse(sum), nil)
}
