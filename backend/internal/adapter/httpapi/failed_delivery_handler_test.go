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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/httpapi"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

type fakeFailedSummarizer struct {
	calls   int
	gotF    domain.FailedFilter
	summary domain.FailedSummary
	err     error
}

func (f *fakeFailedSummarizer) Summary(_ context.Context, flt domain.FailedFilter) (domain.FailedSummary, error) {
	f.calls++
	f.gotF = flt
	return f.summary, f.err
}

func callFailedSummary(t *testing.T, fake *fakeFailedSummarizer, query string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	srv := httpapi.NewRouter(httpapi.Handlers{FailedDeliveries: httpapi.NewFailedDeliveryHandler(fake, logger)}, logger)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/failed-deliveries/summary"+query, nil))
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return rec, body
}

const validFailedQuery = "?start_date=2026-11-08&end_date=2026-11-14&branch_id=JKT01"

func TestFailedDeliveryHandler_Shape(t *testing.T) {
	fake := &fakeFailedSummarizer{summary: domain.FailedSummary{
		TotalPackages: 100, TotalFailed: 30, PreviousTotalFailed: 20,
		Reasons: []domain.FailedReason{{Reason: "Penerima tidak di tempat", Count: 30}},
		TopCouriers: []domain.FailedCourier{
			{Rank: 1, CourierID: "C002", CourierName: "Sari", BranchID: "JKT01", TotalPackages: 40, Failed: 30},
		},
	}}
	rec, body := callFailedSummary(t, fake, validFailedQuery)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotContains(t, body, "meta")
	assert.Equal(t, "JKT01", fake.gotF.BranchID)
	assert.Equal(t, "2026-11-08", fake.gotF.StartDate.Format("2006-01-02"))
	assert.Equal(t, map[string]any{
		"total_packages": float64(100), "total_failed": float64(30), "failed_rate": 0.3,
		"previous_total_failed": float64(20), "failed_change_rate": 0.5,
		"reasons": []any{map[string]any{"reason": "Penerima tidak di tempat", "count": float64(30), "share": 1.0}},
		"top_couriers": []any{map[string]any{
			"rank": float64(1), "courier_id": "C002", "courier_name": "Sari", "branch_id": "JKT01",
			"total_packages": float64(40), "failed": float64(30), "failed_rate": 0.75,
		}},
	}, body["data"])
}

func TestFailedDeliveryHandler_EmptyArraysAndNullChange(t *testing.T) {
	rec, body := callFailedSummary(t, &fakeFailedSummarizer{}, validFailedQuery)
	assert.Equal(t, http.StatusOK, rec.Code)
	data := body["data"].(map[string]any)
	assert.Equal(t, []any{}, data["reasons"])
	assert.Equal(t, []any{}, data["top_couriers"])
	assert.Contains(t, data, "failed_change_rate")
	assert.Nil(t, data["failed_change_rate"])
	assert.Equal(t, 0.0, data["failed_rate"])
}

func TestFailedDeliveryHandler_Errors(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		err        error
		wantStatus int
		wantCode   string
		wantCalls  int
	}{
		{"400 start_date salah", "?start_date=bukan&end_date=2026-11-14", nil, 400, "invalid_query", 0},
		{"400 end_date salah", "?start_date=2026-11-08&end_date=14-11", nil, 400, "invalid_query", 0},
		{"400 invalid_filter", validFailedQuery, domain.ErrInvalidFilter, 400, "invalid_filter", 1},
		{"500", validFailedQuery, errors.New("db down"), 500, "internal_error", 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fake := &fakeFailedSummarizer{err: tc.err}
			rec, body := callFailedSummary(t, fake, tc.query)
			assert.Equal(t, tc.wantStatus, rec.Code)
			assert.Equal(t, tc.wantCalls, fake.calls)
			assert.Equal(t, tc.wantCode, body["error"].(map[string]any)["code"])
		})
	}
}
