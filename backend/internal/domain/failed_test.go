package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

func TestPreviousPeriod(t *testing.T) {
	tests := []struct {
		name               string
		start, end         string
		wantStart, wantEnd string
	}{
		{"7 hari", "2026-11-08", "2026-11-14", "2026-11-01", "2026-11-07"},
		{"satu hari", "2026-11-08", "2026-11-08", "2026-11-07", "2026-11-07"},
		{"lintas bulan", "2026-11-03", "2026-11-09", "2026-10-27", "2026-11-02"},
		{"lintas tahun", "2027-01-03", "2027-01-09", "2026-12-27", "2027-01-02"},
		{"31 hari", "2026-12-01", "2026-12-31", "2026-10-31", "2026-11-30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := domain.PreviousPeriod(date(tt.start), date(tt.end))
			assert.Equal(t, date(tt.wantStart), gotStart)
			assert.Equal(t, date(tt.wantEnd), gotEnd)
		})
	}
}

func TestFailedChangeRate(t *testing.T) {
	t.Run("naik", func(t *testing.T) {
		got := domain.FailedSummary{TotalFailed: 150, PreviousTotalFailed: 100}.FailedChangeRate()
		require.NotNil(t, got)
		assert.InDelta(t, 0.5, *got, 1e-9)
	})
	t.Run("turun", func(t *testing.T) {
		got := domain.FailedSummary{TotalFailed: 180, PreviousTotalFailed: 200}.FailedChangeRate()
		require.NotNil(t, got)
		assert.InDelta(t, -0.1, *got, 1e-9)
	})
	t.Run("sama", func(t *testing.T) {
		got := domain.FailedSummary{TotalFailed: 10, PreviousTotalFailed: 10}.FailedChangeRate()
		require.NotNil(t, got)
		assert.Zero(t, *got)
	})
	t.Run("nil bila periode lalu 0", func(t *testing.T) {
		assert.Nil(t, domain.FailedSummary{TotalFailed: 10}.FailedChangeRate())
		assert.Nil(t, domain.FailedSummary{}.FailedChangeRate())
	})
}

func TestFailedSummary_RatesAndShare(t *testing.T) {
	s := domain.FailedSummary{TotalPackages: 1200, TotalFailed: 180}
	assert.InDelta(t, 0.15, s.FailedRate(), 1e-9)
	assert.InDelta(t, 1.0/3, s.Share(domain.FailedReason{Reason: "x", Count: 60}), 1e-9)

	assert.Zero(t, domain.FailedSummary{TotalFailed: 5}.FailedRate())
	assert.Zero(t, domain.FailedSummary{TotalPackages: 5}.Share(domain.FailedReason{Count: 1}))
}

func TestFailedCourier_FailedRate(t *testing.T) {
	assert.InDelta(t, 0.75, domain.FailedCourier{TotalPackages: 40, Failed: 30}.FailedRate(), 1e-9)
	assert.Zero(t, domain.FailedCourier{Failed: 3}.FailedRate())
}

func TestFailedFilter_Validate(t *testing.T) {
	valid := domain.FailedFilter{StartDate: date("2026-11-01"), EndDate: date("2026-11-07")}

	tests := []struct {
		name    string
		modify  func(f *domain.FailedFilter)
		wantErr bool
	}{
		{name: "valid", modify: func(*domain.FailedFilter) {}},
		{name: "dengan cabang", modify: func(f *domain.FailedFilter) { f.BranchID = "JKT01" }},
		{name: "periode tepat 31 hari", modify: func(f *domain.FailedFilter) { f.EndDate = date("2026-12-01") }},
		{name: "tanpa start_date", modify: func(f *domain.FailedFilter) { f.StartDate = time.Time{} }, wantErr: true},
		{name: "tanpa end_date", modify: func(f *domain.FailedFilter) { f.EndDate = time.Time{} }, wantErr: true},
		{name: "end sebelum start", modify: func(f *domain.FailedFilter) { f.EndDate = date("2026-10-31") }, wantErr: true},
		{name: "periode lebih dari 31 hari", modify: func(f *domain.FailedFilter) { f.EndDate = date("2026-12-02") }, wantErr: true},
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
