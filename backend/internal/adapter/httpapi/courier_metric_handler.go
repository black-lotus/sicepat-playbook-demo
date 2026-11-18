package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// CourierMetricLister adalah use case yang dibutuhkan handler metrik kurir.
type CourierMetricLister interface {
	ListCourierMetrics(ctx context.Context, filter domain.MetricFilter) (domain.CourierMetricPage, error)
}

// CourierMetricHandler menyajikan endpoint KPI kurir.
type CourierMetricHandler struct {
	svc    CourierMetricLister
	logger *slog.Logger
}

// NewCourierMetricHandler membuat CourierMetricHandler.
func NewCourierMetricHandler(svc CourierMetricLister, logger *slog.Logger) *CourierMetricHandler {
	return &CourierMetricHandler{svc: svc, logger: logger}
}

type courierMetricResponse struct {
	CourierID     string  `json:"courier_id"`
	CourierName   string  `json:"courier_name"`
	BranchID      string  `json:"branch_id"`
	Date          string  `json:"date"`
	TotalPackages int     `json:"total_packages"`
	Delivered     int     `json:"delivered"`
	Failed        int     `json:"failed"`
	OnTime        int     `json:"on_time"`
	SuccessRate   float64 `json:"success_rate"`
	SLARate       float64 `json:"sla_rate"`
}

func toCourierMetricResponse(m domain.CourierMetric) courierMetricResponse {
	return courierMetricResponse{
		CourierID:     m.CourierID,
		CourierName:   m.CourierName,
		BranchID:      m.BranchID,
		Date:          m.Date.Format(time.DateOnly),
		TotalPackages: m.TotalPackages,
		Delivered:     m.Delivered,
		Failed:        m.Failed,
		OnTime:        m.OnTime,
		SuccessRate:   m.SuccessRate(),
		SLARate:       m.SLARate(),
	}
}

// List menangani GET /api/v1/courier-metrics.
func (h *CourierMetricHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseMetricFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", err.Error())
		return
	}

	page, err := h.svc.ListCourierMetrics(r.Context(), filter)
	if err != nil {
		writeDomainError(w, r, h.logger, err)
		return
	}

	items := make([]courierMetricResponse, 0, len(page.Items))
	for _, m := range page.Items {
		items = append(items, toCourierMetricResponse(m))
	}
	filter = filter.WithDefaults()
	writeData(w, items, pageMeta{Page: filter.Page, PageSize: filter.PageSize, Total: page.Total})
}

type queryError struct{ param string }

func (e queryError) Error() string { return "parameter " + e.param + " tidak valid" }

func parseMetricFilter(r *http.Request) (domain.MetricFilter, error) {
	q := r.URL.Query()
	var f domain.MetricFilter
	var err error

	if f.StartDate, err = parseDate(q.Get("start_date")); err != nil {
		return f, queryError{"start_date"}
	}
	if f.EndDate, err = parseDate(q.Get("end_date")); err != nil {
		return f, queryError{"end_date"}
	}
	if f.Page, err = parseOptionalInt(q.Get("page")); err != nil {
		return f, queryError{"page"}
	}
	if f.PageSize, err = parseOptionalInt(q.Get("page_size")); err != nil {
		return f, queryError{"page_size"}
	}
	f.BranchID = q.Get("branch_id")
	return f, nil
}

// parseDate menerima tanggal kosong (divalidasi di domain) atau format YYYY-MM-DD.
func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(time.DateOnly, s)
}

func parseOptionalInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}
