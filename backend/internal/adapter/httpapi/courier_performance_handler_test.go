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

// fakeRateTargets meniru service.TargetService: target selalu untuk bulan end_date (WIB).
type fakeRateTargets struct {
	target *domain.KPITarget
	err    error
	calls  int
	gotID  string
	gotEnd time.Time
}

func (f *fakeRateTargets) RateTarget(_ context.Context, branchID string, end time.Time) (*domain.KPITarget, error) {
	f.calls++
	f.gotID, f.gotEnd = branchID, end
	if f.err != nil || f.target == nil {
		return nil, f.err
	}
	t := *f.target
	t.BranchID = branchID
	t.Month = domain.MonthOf(end)
	return &t, nil
}

func newPerformanceServer(lister *fakePerformanceLister) http.Handler {
	return newPerformanceServerWithTargets(lister, &fakeRateTargets{})
}

func newPerformanceServerWithTargets(lister *fakePerformanceLister, targets httpapi.RateTargetFinder) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return httpapi.NewRouter(httpapi.Handlers{
		CourierMetrics:     httpapi.NewCourierMetricHandler(&fakeLister{}, logger),
		CourierPerformance: httpapi.NewCourierPerformanceHandler(lister, targets, logger),
	}, logger)
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
				 "success_rate":0.75,"sla_rate":1,"success_rate_gap":0.15000000000000002,"success_rate_status":null,
				 "daily":[
				   {"date":"2026-11-01","total_packages":3,"delivered":3,"failed":0,"on_time":3},
				   {"date":"2026-11-02","total_packages":1,"delivered":0,"failed":1,"on_time":0}]},
				{"rank":4,"courier_id":"C002","courier_name":"Sari Wulandari","branch_id":"JKT01",
				 "total_packages":0,"delivered":0,"failed":0,"on_time":0,
				 "success_rate":0,"sla_rate":0,"success_rate_gap":0.9,"success_rate_status":null,
				 "daily":[
				   {"date":"2026-11-01","total_packages":0,"delivered":0,"failed":0,"on_time":0},
				   {"date":"2026-11-02","total_packages":0,"delivered":0,"failed":0,"on_time":0}]}],
				"meta":{"page":2,"page_size":2,"total":4,"branch_target":null}}`,
		},
		{
			name:       "data kosong tetap mengembalikan array",
			query:      "?start_date=2026-11-01&end_date=2026-11-07",
			lister:     &fakePerformanceLister{},
			wantStatus: http.StatusOK,
			wantBody:   `{"data":[],"meta":{"page":1,"page_size":20,"total":0,"branch_target":null}}`,
		},
		{
			name:       "daily nil tetap di-encode sebagai array",
			query:      "?start_date=2026-11-01&end_date=2026-11-07",
			lister:     &fakePerformanceLister{page: domain.CourierPerformancePage{Total: 1, Items: []domain.CourierPerformance{{Rank: 1, CourierID: "C001"}}}},
			wantStatus: http.StatusOK,
			wantBody: `{"data":[{"rank":1,"courier_id":"C001","courier_name":"","branch_id":"","total_packages":0,
				"delivered":0,"failed":0,"on_time":0,"success_rate":0,"sla_rate":0,"success_rate_gap":0,"success_rate_status":null,"daily":[]}],
				"meta":{"page":1,"page_size":20,"total":1,"branch_target":null}}`,
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

func targetPerformancePage() domain.CourierPerformancePage {
	return domain.CourierPerformancePage{
		Total: 3,
		Items: []domain.CourierPerformance{
			{Rank: 1, CourierID: "C001", BranchID: "JKT01", TotalPackages: 10, Delivered: 9, OnTime: 9, LeaderSuccessRate: 0.9},
			{Rank: 2, CourierID: "C002", BranchID: "JKT01", TotalPackages: 100, Delivered: 85, OnTime: 80, LeaderSuccessRate: 0.9},
			{Rank: 3, CourierID: "C003", BranchID: "JKT01", TotalPackages: 100000, Delivered: 84996, OnTime: 80000, LeaderSuccessRate: 0.9},
		},
	}
}

func getPerformance(t *testing.T, lister *fakePerformanceLister, targets *fakeRateTargets, query string) (int, map[string]any) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/courier-performance"+query, nil)
	newPerformanceServerWithTargets(lister, targets).ServeHTTP(rec, req)
	var body map[string]any
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	return rec.Code, body
}

func statuses(t *testing.T, body map[string]any) []any {
	t.Helper()
	var out []any
	for _, it := range body["data"].([]any) {
		out = append(out, it.(map[string]any)["success_rate_status"])
	}
	return out
}

func TestCourierPerformanceHandler_Target(t *testing.T) {
	target := &domain.KPITarget{SuccessRate: 0.85, SLARate: 0.8, DeliveredPackages: 3000}

	t.Run("dengan target mengisi meta dan status per kurir", func(t *testing.T) {
		targets := &fakeRateTargets{target: target}
		code, body := getPerformance(t, &fakePerformanceLister{page: targetPerformancePage()}, targets,
			"?start_date=2026-11-01&end_date=2026-11-30&branch_id=JKT01")

		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, 1, targets.calls)
		assert.Equal(t, "JKT01", targets.gotID)
		assert.Equal(t, []any{"met", "met", "below"}, statuses(t, body))
		meta := body["meta"].(map[string]any)
		assert.Equal(t, map[string]any{"month": "2026-11", "success_rate": 0.85, "sla_rate": 0.8}, meta["branch_target"])
		assert.Equal(t, 1.0, meta["page"])
		assert.Equal(t, 20.0, meta["page_size"])
		assert.Equal(t, 3.0, meta["total"])
	})

	t.Run("bulan target mengikuti end_date", func(t *testing.T) {
		tests := []struct {
			end       string
			wantMonth string
		}{
			{end: "2026-11-30", wantMonth: "2026-11"},
			{end: "2026-12-01", wantMonth: "2026-12"},
		}
		for _, tt := range tests {
			targets := &fakeRateTargets{target: target}
			_, body := getPerformance(t, &fakePerformanceLister{page: targetPerformancePage()}, targets,
				"?start_date=2026-11-25&end_date="+tt.end+"&branch_id=JKT01")

			assert.Equal(t, tt.end, targets.gotEnd.Format(time.DateOnly))
			bt := body["meta"].(map[string]any)["branch_target"].(map[string]any)
			assert.Equal(t, tt.wantMonth, bt["month"])
		}
	})

	t.Run("cabang tanpa target menghasilkan null", func(t *testing.T) {
		targets := &fakeRateTargets{}
		code, body := getPerformance(t, &fakePerformanceLister{page: targetPerformancePage()}, targets,
			"?start_date=2026-11-01&end_date=2026-11-30&branch_id=BDG01")

		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, 1, targets.calls)
		assert.Equal(t, []any{nil, nil, nil}, statuses(t, body))
		meta := body["meta"].(map[string]any)
		v, present := meta["branch_target"]
		assert.True(t, present)
		assert.Nil(t, v)
	})

	t.Run("tanpa branch_id target tidak dicari", func(t *testing.T) {
		targets := &fakeRateTargets{target: target}
		code, body := getPerformance(t, &fakePerformanceLister{page: targetPerformancePage()}, targets,
			"?start_date=2026-11-01&end_date=2026-11-30")

		assert.Equal(t, http.StatusOK, code)
		assert.Equal(t, 0, targets.calls)
		assert.Equal(t, []any{nil, nil, nil}, statuses(t, body))
		v, present := body["meta"].(map[string]any)["branch_target"]
		assert.True(t, present)
		assert.Nil(t, v)
	})

	t.Run("error target menjadi 500 tanpa bocor", func(t *testing.T) {
		targets := &fakeRateTargets{err: errors.New("connection refused")}
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet,
			"/api/v1/courier-performance?start_date=2026-11-01&end_date=2026-11-30&branch_id=JKT01", nil)
		newPerformanceServerWithTargets(&fakePerformanceLister{page: targetPerformancePage()}, targets).ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"error":{"code":"internal_error","message":"terjadi kesalahan pada server"}}`, rec.Body.String())
	})

	t.Run("filter tidak valid tidak mencari target", func(t *testing.T) {
		targets := &fakeRateTargets{target: target}
		code, _ := getPerformance(t, &fakePerformanceLister{}, targets, "?branch_id=JKT01")

		assert.Equal(t, http.StatusBadRequest, code)
		assert.Equal(t, 0, targets.calls)
	})

	t.Run("field lama sama dengan golden sebelum B9", func(t *testing.T) {
		const golden = `{"rank":1,"courier_id":"C001","courier_name":"","branch_id":"JKT01",
			"total_packages":10,"delivered":9,"failed":0,"on_time":9,
			"success_rate":0.9,"sla_rate":1,"success_rate_gap":0,"daily":[]}`
		_, body := getPerformance(t, &fakePerformanceLister{page: targetPerformancePage()}, &fakeRateTargets{target: target},
			"?start_date=2026-11-01&end_date=2026-11-30&branch_id=JKT01")

		item := body["data"].([]any)[0].(map[string]any)
		delete(item, "success_rate_status")
		got, err := json.Marshal(item)
		assert.NoError(t, err)
		assert.JSONEq(t, golden, string(got))

		meta := body["meta"].(map[string]any)
		delete(meta, "branch_target")
		assert.Equal(t, map[string]any{"page": 1.0, "page_size": 20.0, "total": 3.0}, meta)
	})
}
