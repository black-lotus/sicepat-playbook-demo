package httpapi_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/httpapi"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

type fakePerformanceLister struct {
	gotFilter domain.MetricFilter
	page      domain.CourierPerformancePage
	err       error
}

func (f *fakePerformanceLister) ListCourierPerformance(_ context.Context, filter domain.MetricFilter) (domain.CourierPerformancePage, error) {
	f.gotFilter = filter
	if f.err != nil {
		return domain.CourierPerformancePage{}, f.err
	}
	if err := filter.WithDefaults().Validate(); err != nil {
		return domain.CourierPerformancePage{}, err
	}
	return f.page, nil
}

func newPerformanceServer(lister *fakePerformanceLister) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return httpapi.NewRouter(
		httpapi.NewCourierMetricHandler(&fakeLister{}, logger),
		httpapi.NewCourierPerformanceHandler(lister, logger),
		logger,
	)
}

func TestCourierPerformanceHandler_List(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 11, d, 0, 0, 0, 0, time.UTC) }
	page := domain.CourierPerformancePage{
		Total: 4,
		Items: []domain.CourierPerformance{
			{
				Rank: 3, CourierID: "C004", CourierName: "Agus Pratama", BranchID: "JKT01",
				TotalPackages: 4, Delivered: 3, Failed: 1, OnTime: 3, LeaderSuccessRate: 0.9,
				Daily: []domain.DailyCount{
					{Date: day(1), TotalPackages: 3, Delivered: 3, OnTime: 3},
					{Date: day(2), TotalPackages: 1, Failed: 1},
				},
			},
			{
				Rank: 4, CourierID: "C002", CourierName: "Sari Wulandari", BranchID: "JKT01",
				LeaderSuccessRate: 0.9, Daily: []domain.DailyCount{{Date: day(1)}, {Date: day(2)}},
			},
		},
	}

	tests := []struct {
		name       string
		query      string
		lister     *fakePerformanceLister
		wantStatus int
		wantBody   string
	}{
		{
			name:       "sukses dengan rank, gap, dan daily",
			query:      "?start_date=2026-11-01&end_date=2026-11-02&branch_id=JKT01&page=2&page_size=2",
			lister:     &fakePerformanceLister{page: page},
			wantStatus: http.StatusOK,
			wantBody: `{"data":[
				{"rank":3,"courier_id":"C004","courier_name":"Agus Pratama","branch_id":"JKT01",
				 "total_packages":4,"delivered":3,"failed":1,"on_time":3,
				 "success_rate":0.75,"sla_rate":1,"success_rate_gap":0.15000000000000002,
				 "daily":[
				   {"date":"2026-11-01","total_packages":3,"delivered":3,"failed":0,"on_time":3},
				   {"date":"2026-11-02","total_packages":1,"delivered":0,"failed":1,"on_time":0}]},
				{"rank":4,"courier_id":"C002","courier_name":"Sari Wulandari","branch_id":"JKT01",
				 "total_packages":0,"delivered":0,"failed":0,"on_time":0,
				 "success_rate":0,"sla_rate":0,"success_rate_gap":0.9,
				 "daily":[
				   {"date":"2026-11-01","total_packages":0,"delivered":0,"failed":0,"on_time":0},
				   {"date":"2026-11-02","total_packages":0,"delivered":0,"failed":0,"on_time":0}]}],
				"meta":{"page":2,"page_size":2,"total":4}}`,
		},
		{
			name:       "data kosong tetap mengembalikan array",
			query:      "?start_date=2026-11-01&end_date=2026-11-07",
			lister:     &fakePerformanceLister{},
			wantStatus: http.StatusOK,
			wantBody:   `{"data":[],"meta":{"page":1,"page_size":20,"total":0}}`,
		},
		{
			name:       "daily nil tetap di-encode sebagai array",
			query:      "?start_date=2026-11-01&end_date=2026-11-07",
			lister:     &fakePerformanceLister{page: domain.CourierPerformancePage{Total: 1, Items: []domain.CourierPerformance{{Rank: 1, CourierID: "C001"}}}},
			wantStatus: http.StatusOK,
			wantBody: `{"data":[{"rank":1,"courier_id":"C001","courier_name":"","branch_id":"","total_packages":0,
				"delivered":0,"failed":0,"on_time":0,"success_rate":0,"sla_rate":0,"success_rate_gap":0,"daily":[]}],
				"meta":{"page":1,"page_size":20,"total":1}}`,
		},
		{
			name:       "format tanggal salah",
			query:      "?start_date=abc&end_date=2026-11-07",
			lister:     &fakePerformanceLister{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_query","message":"parameter start_date tidak valid"}}`,
		},
		{
			name:       "filter ditolak domain",
			query:      "",
			lister:     &fakePerformanceLister{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_filter","message":"filter tidak valid: start_date dan end_date wajib diisi"}}`,
		},
		{
			name:       "error internal tidak bocor ke client",
			query:      "?start_date=2026-11-01&end_date=2026-11-07",
			lister:     &fakePerformanceLister{err: fmt.Errorf("ambil performa kurir: %w", errors.New("password authentication failed"))},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":{"code":"internal_error","message":"terjadi kesalahan pada server"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/courier-performance"+tt.query, nil)
			newPerformanceServer(tt.lister).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.wantBody, rec.Body.String())
		})
	}
}

func TestCourierPerformanceHandler_ParsesFilter(t *testing.T) {
	lister := &fakePerformanceLister{}
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
		"/api/v1/courier-performance?start_date=2026-11-01&end_date=2026-11-07&branch_id=BDG02&page=3&page_size=50", nil)
	newPerformanceServer(lister).ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, domain.MetricFilter{
		StartDate: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC),
		BranchID:  "BDG02",
		Page:      3,
		PageSize:  50,
	}, lister.gotFilter)
}
