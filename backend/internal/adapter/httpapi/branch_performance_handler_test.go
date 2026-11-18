package httpapi_test

import (
	"context"
	"errors"
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

type fakeBranchLister struct {
	gotFilter domain.MetricFilter
	page      domain.BranchPerformanceResultPage
	err       error
}

func (f *fakeBranchLister) List(_ context.Context, filter domain.MetricFilter) (domain.BranchPerformanceResultPage, error) {
	f.gotFilter = filter
	if f.err != nil {
		return domain.BranchPerformanceResultPage{}, f.err
	}
	if err := filter.WithDefaults().Validate(); err != nil {
		return domain.BranchPerformanceResultPage{}, err
	}
	return f.page, nil
}

func newBranchPerformanceServer(lister *fakeBranchLister) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return httpapi.NewRouter(httpapi.Handlers{
		BranchPerformance: httpapi.NewBranchPerformanceHandler(lister, logger),
	}, logger)
}

func TestBranchPerformanceHandler_List(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 11, d, 0, 0, 0, 0, time.UTC) }
	page := domain.BranchPerformanceResultPage{
		Total: 2,
		Items: []domain.BranchPerformanceWithTarget{
			{
				BranchPerformance: domain.BranchPerformance{
					Rank: 1, BranchID: "BDG02", CourierCount: 8,
					TotalPackages: 1000, Delivered: 800, Failed: 200, OnTime: 400,
					Daily: []domain.DailyCount{{Date: day(1), TotalPackages: 1000, Delivered: 800, Failed: 200, OnTime: 400}},
				},
				Target: &domain.BranchTargetResult{
					Target:            domain.PeriodTarget{SuccessRate: 0.85, SLARate: 0.5, DeliveredPackages: 900},
					SuccessRateStatus: domain.TargetBelow, SLARateStatus: domain.TargetMet, DeliveredStatus: domain.TargetBelow,
				},
			},
			{BranchPerformance: domain.BranchPerformance{Rank: 2, BranchID: "BDG01"}},
		},
	}

	tests := []struct {
		name       string
		query      string
		lister     *fakeBranchLister
		wantStatus int
		wantBody   string
	}{
		{
			name:       "sukses dengan target dan target null",
			query:      "?start_date=2026-11-01&end_date=2026-11-01&page=1&page_size=2&branch_id=IGNORED",
			lister:     &fakeBranchLister{page: page},
			wantStatus: http.StatusOK,
			wantBody: `{"data":[
				{"rank":1,"branch_id":"BDG02","courier_count":8,"total_packages":1000,"delivered":800,"failed":200,"on_time":400,
				 "success_rate":0.8,"sla_rate":0.5,
				 "target":{"success_rate":0.85,"sla_rate":0.5,"delivered_packages":900,
				   "success_rate_status":"below","sla_rate_status":"met","delivered_status":"below"},
				 "daily":[{"date":"2026-11-01","total_packages":1000,"delivered":800,"failed":200,"on_time":400}]},
				{"rank":2,"branch_id":"BDG01","courier_count":0,"total_packages":0,"delivered":0,"failed":0,"on_time":0,
				 "success_rate":0,"sla_rate":0,"target":null,"daily":[]}],
				"meta":{"page":1,"page_size":2,"total":2}}`,
		},
		{
			name:       "data kosong tetap array",
			query:      "?start_date=2026-11-01&end_date=2026-11-07",
			lister:     &fakeBranchLister{},
			wantStatus: http.StatusOK,
			wantBody:   `{"data":[],"meta":{"page":1,"page_size":20,"total":0}}`,
		},
		{
			name:       "format tanggal salah",
			query:      "?start_date=abc&end_date=2026-11-07",
			lister:     &fakeBranchLister{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_query","message":"parameter start_date tidak valid"}}`,
		},
		{
			name:       "page bukan angka",
			query:      "?start_date=2026-11-01&end_date=2026-11-07&page=x",
			lister:     &fakeBranchLister{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_query","message":"parameter page tidak valid"}}`,
		},
		{
			name:       "filter ditolak domain",
			query:      "",
			lister:     &fakeBranchLister{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":{"code":"invalid_filter","message":"filter tidak valid: start_date dan end_date wajib diisi"}}`,
		},
		{
			name:       "error internal tidak bocor",
			query:      "?start_date=2026-11-01&end_date=2026-11-07",
			lister:     &fakeBranchLister{err: errors.New("password authentication failed")},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":{"code":"internal_error","message":"terjadi kesalahan pada server"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/branch-performance"+tt.query, nil)
			newBranchPerformanceServer(tt.lister).ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			assert.JSONEq(t, tt.wantBody, rec.Body.String())
		})
	}
}
