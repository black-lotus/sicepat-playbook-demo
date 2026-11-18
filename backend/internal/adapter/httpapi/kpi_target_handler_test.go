package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/httpapi"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

type fakeTargetGetter struct {
	gotBranch string
	gotMonth  time.Time
	calls     int
	target    domain.KPITarget
	err       error
}

func (f *fakeTargetGetter) GetTarget(_ context.Context, branchID string, month time.Time) (domain.KPITarget, error) {
	f.calls++
	f.gotBranch, f.gotMonth = branchID, month
	return f.target, f.err
}

func TestKPITargetHandler(t *testing.T) {
	nov := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	ok := domain.KPITarget{BranchID: "JKT01", Month: nov, SuccessRate: 0.9, SLARate: 0.85, DeliveredPackages: 3000}

	tests := []struct {
		name       string
		query      string
		getter     *fakeTargetGetter
		wantStatus int
		wantCode   string
		wantCalls  int
	}{
		{"200", "?branch_id=JKT01&month=2026-11", &fakeTargetGetter{target: ok}, http.StatusOK, "", 1},
		{"404 target_not_found", "?branch_id=BDG01&month=2026-11", &fakeTargetGetter{err: domain.ErrTargetNotFound}, http.StatusNotFound, "target_not_found", 1},
		{"400 month format salah", "?branch_id=JKT01&month=2026-13", &fakeTargetGetter{}, http.StatusBadRequest, "invalid_query", 0},
		{"400 month tanggal penuh", "?branch_id=JKT01&month=2026-11-01", &fakeTargetGetter{}, http.StatusBadRequest, "invalid_query", 0},
		{"400 month kosong", "?branch_id=JKT01", &fakeTargetGetter{}, http.StatusBadRequest, "invalid_query", 0},
		{"400 branch kosong", "?month=2026-11", &fakeTargetGetter{}, http.StatusBadRequest, "invalid_query", 0},
		{"500 error tak dikenal", "?branch_id=JKT01&month=2026-11", &fakeTargetGetter{err: errors.New("db down")}, http.StatusInternalServerError, "internal_error", 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			srv := httpapi.NewRouter(httpapi.Handlers{KPITargets: httpapi.NewKPITargetHandler(tc.getter, logger)}, logger)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/branch-kpi-targets"+tc.query, nil))

			assert.Equal(t, tc.wantStatus, rec.Code)
			assert.Equal(t, tc.wantCalls, tc.getter.calls)

			var body map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			if tc.wantCode != "" {
				assert.Equal(t, tc.wantCode, body["error"].(map[string]any)["code"])
				return
			}
			assert.NotContains(t, body, "meta", "respons satu objek tanpa meta")
			assert.Equal(t, map[string]any{
				"branch_id": "JKT01", "month": "2026-11", "success_rate": 0.9, "sla_rate": 0.85, "delivered_packages": float64(3000),
			}, body["data"])
			assert.Equal(t, "JKT01", tc.getter.gotBranch)
			assert.Equal(t, nov, tc.getter.gotMonth)
		})
	}
}
