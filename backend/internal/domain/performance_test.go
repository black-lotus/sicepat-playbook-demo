package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

func nov(day int) time.Time {
	return time.Date(2026, 11, day, 0, 0, 0, 0, time.UTC)
}

func TestFillDaily(t *testing.T) {
	counts := []domain.DailyCount{
		{Date: nov(5), TotalPackages: 6, Delivered: 5, Failed: 1, OnTime: 4},
		{Date: nov(2), TotalPackages: 3, Delivered: 3, OnTime: 2},
	}

	t.Run("hari kosong diisi nol dan urut menaik", func(t *testing.T) {
		got := domain.FillDaily(counts, nov(1), nov(7))
		require.Len(t, got, 7)
		for i, d := range got {
			assert.Equal(t, nov(i+1).Format(time.DateOnly), d.Date.Format(time.DateOnly))
		}
		assert.Equal(t, domain.DailyCount{Date: nov(2), TotalPackages: 3, Delivered: 3, OnTime: 2}, got[1])
		assert.Equal(t, domain.DailyCount{Date: nov(5), TotalPackages: 6, Delivered: 5, Failed: 1, OnTime: 4}, got[4])
		for _, i := range []int{0, 2, 3, 5, 6} {
			assert.Equal(t, domain.DailyCount{Date: nov(i + 1)}, got[i], "hari ke-%d", i+1)
		}
	})

	t.Run("periode satu hari menghasilkan satu entri", func(t *testing.T) {
		got := domain.FillDaily(counts, nov(5), nov(5))
		require.Len(t, got, 1)
		assert.Equal(t, 6, got[0].TotalPackages)
	})

	t.Run("counts nil menghasilkan entri nol, bukan nil", func(t *testing.T) {
		got := domain.FillDaily(nil, nov(1), nov(3))
		require.NotNil(t, got)
		require.Len(t, got, 3)
		for i, d := range got {
			assert.Equal(t, domain.DailyCount{Date: nov(i + 1)}, d)
		}
	})

	t.Run("data di luar periode diabaikan", func(t *testing.T) {
		got := domain.FillDaily(counts, nov(3), nov(4))
		require.Len(t, got, 2)
		assert.Zero(t, got[0].TotalPackages)
		assert.Zero(t, got[1].TotalPackages)
	})

	t.Run("end sebelum start menghasilkan slice kosong", func(t *testing.T) {
		got := domain.FillDaily(counts, nov(7), nov(1))
		assert.NotNil(t, got)
		assert.Empty(t, got)
	})
}

func TestCourierPerformance_Rates(t *testing.T) {
	tests := []struct {
		name        string
		perf        domain.CourierPerformance
		wantSuccess float64
		wantSLA     float64
		wantGap     float64
	}{
		{
			name:        "pemimpin tanpa selisih",
			perf:        domain.CourierPerformance{TotalPackages: 40, Delivered: 36, Failed: 4, OnTime: 33, LeaderSuccessRate: 0.9},
			wantSuccess: 0.9,
			wantSLA:     33.0 / 36.0,
			wantGap:     0,
		},
		{
			name:        "semua gagal punya selisih penuh",
			perf:        domain.CourierPerformance{TotalPackages: 4, Failed: 4, LeaderSuccessRate: 0.9},
			wantSuccess: 0,
			wantSLA:     0,
			wantGap:     0.9,
		},
		{
			name:        "total nol tidak menghasilkan NaN",
			perf:        domain.CourierPerformance{LeaderSuccessRate: 0.5},
			wantSuccess: 0,
			wantSLA:     0,
			wantGap:     0.5,
		},
		{
			name:        "selisih tidak pernah negatif",
			perf:        domain.CourierPerformance{TotalPackages: 10, Delivered: 10, OnTime: 10, LeaderSuccessRate: 0.9},
			wantSuccess: 1,
			wantSLA:     1,
			wantGap:     0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.wantSuccess, tt.perf.SuccessRate(), 1e-12)
			assert.InDelta(t, tt.wantSLA, tt.perf.SLARate(), 1e-12)
			assert.InDelta(t, tt.wantGap, tt.perf.SuccessRateGap(), 1e-12)
		})
	}
}

func TestCourierPerformance_SuccessRateStatus(t *testing.T) {
	target := &domain.KPITarget{SuccessRate: 0.85}
	tests := []struct {
		name   string
		perf   domain.CourierPerformance
		target *domain.KPITarget
		want   *domain.TargetStatus
	}{
		{name: "tanpa target nil", perf: domain.CourierPerformance{TotalPackages: 10, Delivered: 10}, target: nil, want: nil},
		{name: "sama dengan target met", perf: domain.CourierPerformance{TotalPackages: 100, Delivered: 85}, target: target, want: ptrStatus(domain.TargetMet)},
		{name: "di atas target met", perf: domain.CourierPerformance{TotalPackages: 10, Delivered: 10}, target: target, want: ptrStatus(domain.TargetMet)},
		{name: "0,84996 di bawah 0,85 below", perf: domain.CourierPerformance{TotalPackages: 100000, Delivered: 84996}, target: target, want: ptrStatus(domain.TargetBelow)},
		{name: "tanpa paket di bawah target", perf: domain.CourierPerformance{}, target: target, want: ptrStatus(domain.TargetBelow)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.perf.SuccessRateStatus(tt.target))
		})
	}
}

func ptrStatus(s domain.TargetStatus) *domain.TargetStatus { return &s }
