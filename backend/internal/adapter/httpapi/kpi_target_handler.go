package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

const monthLayout = "2006-01"

// KPITargetGetter adalah use case yang dibutuhkan handler target KPI.
type KPITargetGetter interface {
	GetTarget(ctx context.Context, branchID string, month time.Time) (domain.KPITarget, error)
}

// KPITargetHandler menyajikan endpoint target KPI per cabang.
type KPITargetHandler struct {
	svc    KPITargetGetter
	logger *slog.Logger
}

// NewKPITargetHandler membuat KPITargetHandler.
func NewKPITargetHandler(svc KPITargetGetter, logger *slog.Logger) *KPITargetHandler {
	return &KPITargetHandler{svc: svc, logger: logger}
}

type kpiTargetResponse struct {
	BranchID          string  `json:"branch_id"`
	Month             string  `json:"month"`
	SuccessRate       float64 `json:"success_rate"`
	SLARate           float64 `json:"sla_rate"`
	DeliveredPackages int     `json:"delivered_packages"`
}

// Get menangani GET /api/v1/branch-kpi-targets.
func (h *KPITargetHandler) Get(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	branchID := q.Get("branch_id")
	if branchID == "" {
		writeError(w, http.StatusBadRequest, "invalid_query", queryError{"branch_id"}.Error())
		return
	}
	month, err := time.Parse(monthLayout, q.Get("month"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_query", queryError{"month"}.Error())
		return
	}
	t, err := h.svc.GetTarget(r.Context(), branchID, month)
	if err != nil {
		writeDomainError(w, r, h.logger, err)
		return
	}
	writeData(w, kpiTargetResponse{
		BranchID:          t.BranchID,
		Month:             t.Month.Format(monthLayout),
		SuccessRate:       t.SuccessRate,
		SLARate:           t.SLARate,
		DeliveredPackages: t.DeliveredPackages,
	}, nil)
}
