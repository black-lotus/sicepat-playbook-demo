package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

func date(s string) time.Time {
	t, err := time.Parse(time.DateOnly, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestCourierMetric_Rates(t *testing.T) {
	tests := []struct {
		name        string
		metric      domain.CourierMetric
		wantSuccess float64
		wantSLA     float64
	}{
		{name: "tanpa paket", metric: domain.CourierMetric{}, wantSuccess: 0, wantSLA: 0},
		{
			name:        "sebagian terkirim",
			metric:      domain.CourierMetric{TotalPackages: 10, Delivered: 8, Failed: 2, OnTime: 6},
			wantSuccess: 0.8,
			wantSLA:     0.75,
		},
		{
			name:        "semua gagal",
			metric:      domain.CourierMetric{TotalPackages: 4, Failed: 4},
			wantSuccess: 0,
			wantSLA:     0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.wantSuccess, tt.metric.SuccessRate(), 1e-9)
			assert.InDelta(t, tt.wantSLA, tt.metric.SLARate(), 1e-9)
		})
	}
}

func TestMetricFilter_Validate(t *testing.T) {
	valid := domain.MetricFilter{StartDate: date("2026-11-01"), EndDate: date("2026-11-07"), Page: 1, PageSize: 20}

	tests := []struct {
		name    string
		modify  func(f *domain.MetricFilter)
		wantErr bool
	}{
		{name: "valid", modify: func(*domain.MetricFilter) {}},
		{name: "periode tepat 31 hari", modify: func(f *domain.MetricFilter) { f.EndDate = date("2026-12-01") }},
		{name: "tanpa start_date", modify: func(f *domain.MetricFilter) { f.StartDate = time.Time{} }, wantErr: true},
		{name: "end sebelum start", modify: func(f *domain.MetricFilter) { f.EndDate = date("2026-10-31") }, wantErr: true},
		{name: "periode lebih dari 31 hari", modify: func(f *domain.MetricFilter) { f.EndDate = date("2026-12-02") }, wantErr: true},
		{name: "page 0", modify: func(f *domain.MetricFilter) { f.Page = 0 }, wantErr: true},
		{name: "page_size terlalu besar", modify: func(f *domain.MetricFilter) { f.PageSize = 101 }, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := valid
			tt.modify(&f)
			err := f.Validate()
			if tt.wantErr {
				assert.ErrorIs(t, err, domain.ErrInvalidFilter)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestMetricFilter_WithDefaultsAndOffset(t *testing.T) {
	f := domain.MetricFilter{}.WithDefaults()
	assert.Equal(t, 1, f.Page)
	assert.Equal(t, domain.DefaultPageSize, f.PageSize)
	assert.Equal(t, 0, f.Offset())

	f.Page = 3
	assert.Equal(t, 40, f.Offset())
}

func TestMetricFilter_CacheKey(t *testing.T) {
	f := domain.MetricFilter{StartDate: date("2026-11-01"), EndDate: date("2026-11-07"), BranchID: "JKT01", Page: 2, PageSize: 50}
	assert.Equal(t, "courier-metrics:2026-11-01:2026-11-07:JKT01:2:50", f.CacheKey())
}
