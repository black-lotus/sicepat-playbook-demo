package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type fakeLister struct {
	gotFilter domain.MetricFilter
	page      domain.CourierMetricPage
	err       error
}

func (f *fakeLister) ListCourierMetrics(_ context.Context, filter domain.MetricFilter) (domain.CourierMetricPage, error) {
	f.gotFilter = filter
	if f.err != nil {
		return domain.CourierMetricPage{}, f.err
	}
	if err := filter.WithDefaults().Validate(); err != nil {
		return domain.CourierMetricPage{}, err
	}
	return f.page, nil
}

func newServer(lister *fakeLister) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return httpapi.NewRouter(httpapi.Handlers{
		CourierMetrics:     httpapi.NewCourierMetricHandler(lister, logger),
		CourierPerformance: httpapi.NewCourierPerformanceHandler(&fakePerformanceLister{}, &fakeRateTargets{}, logger),
	}, logger)
}

func TestCourierMetricHandler_List(t *testing.T) {
	metric := domain.CourierMetric{
		CourierID:     "C001",
		CourierName:   "Budi",
		BranchID:      "JKT01",
		Date:          time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC),
		TotalPackages: 10,
		Delivered:     8,
		Failed:        2,
		OnTime:        6,
	}

	tests := []struct {
		name       string
		query      string
		lister     *fakeLister
		wantStatus int
		wantBody   string
	}{
		{
			name:       "sukses dengan pagination",
			query:      "?start_date=2026-11-01&end_date=2026-11-17&branch_id=JKT01&page=2&page_size=10",
			lister:     &fakeLister{page: domain.CourierMetricPage{Items: []domain.CourierMetric{metric}, Total: 11}},
			wantStatus: http.StatusOK,
			wantBody: `{"data":[{"courier_id":"C001","courier_name":"Budi","branch_id":"JKT01","date":"2026-11-17",
				"total_packages":10,"delivered":8,"failed":2,"on_time":6,"success_rate":0.8,"sla_rate":0.75}],
				"meta":{"page":2,"page_size":10,"total":11}}`,
		},
		{
			name:       "data kosong tetap mengembalikan array",
			query:      "?start_date=2026-11-01&end_date=2026-11-07",
			lister:     &fakeLister{},
			wantStatus: http.StatusOK,
			wantBody:   `{"data":[],"meta":{"page":1,"page_size":20,"total":0}}`,
		},
		{
			name:       "format tanggal salah",
			query:      "?start_date=01-11-2026&end_date=2026-11-07",
			lister:     &fakeLister{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_query","message":"parameter start_date tidak valid"}}`,
		},
		{
			name:       "filter ditolak domain",
			query:      "",
			lister:     &fakeLister{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_filter","message":"filter tidak valid: start_date dan end_date wajib diisi"}}`,
		},
		{
			name:       "error internal tidak bocor ke client",
			query:      "?start_date=2026-11-01&end_date=2026-11-07",
			lister:     &fakeLister{err: fmt.Errorf("ambil metrik kurir: %w", errors.New("password authentication failed"))},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":{"code":"internal_error","message":"terjadi kesalahan pada server"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/courier-metrics"+tt.query, nil)

			newServer(tt.lister).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestCourierMetricHandler_ParsesFilter(t *testing.T) {
	lister := &fakeLister{}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/courier-metrics?start_date=2026-11-01&end_date=2026-11-07&branch_id=BDG02&page=3&page_size=50", nil)

	newServer(lister).ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, domain.MetricFilter{
		StartDate: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC),
		BranchID:  "BDG02",
		Page:      3,
		PageSize:  50,
	}, lister.gotFilter)
}

func TestHealthz(t *testing.T) {
	rec := httptest.NewRecorder()
	newServer(&fakeLister{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "ok", body["status"])
}
