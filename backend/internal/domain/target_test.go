package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

func civil(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func kpiTarget(year int, month time.Month, successRate, slaRate float64, delivered int) domain.KPITarget {
	return domain.KPITarget{
		BranchID: "JKT01", Month: civil(year, month, 1),
		SuccessRate: successRate, SLARate: slaRate, DeliveredPackages: delivered,
	}
}

func TestMonthOf(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{"akhir November", civil(2026, 11, 30), civil(2026, 11, 1)},
		{"awal Desember", civil(2026, 12, 1), civil(2026, 12, 1)},
		{"23:59 WIB masih bulan itu", time.Date(2026, 11, 30, 23, 59, 0, 0, domain.WIB), civil(2026, 11, 1)},
		{"00:00 WIB bulan berikutnya", time.Date(2026, 12, 1, 0, 0, 0, 0, domain.WIB), civil(2026, 12, 1)},
		{"UTC sore sudah bulan berikutnya di WIB", time.Date(2026, 11, 30, 18, 0, 0, 0, time.UTC), civil(2026, 12, 1)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.MonthOf(tc.in)
			assert.True(t, tc.want.Equal(got), "want %s got %s", tc.want, got)
			assert.Equal(t, time.UTC, got.Location())
		})
	}
}

func TestTargetForPeriod(t *testing.T) {
	cases := []struct {
		name       string
		targets    []domain.KPITarget
		start, end time.Time
		want       *domain.PeriodTarget
	}{
		{
			name:    "satu bulan, 7 hari dari 3000 jadi 700",
			targets: []domain.KPITarget{kpiTarget(2026, 11, 0.9, 0.85, 3000)},
			start:   civil(2026, 11, 1), end: civil(2026, 11, 7),
			want: &domain.PeriodTarget{SuccessRate: 0.9, SLARate: 0.85, DeliveredPackages: 700},
		},
		{
			name: "lintas Oktober dan November: rate dari bulan end, volume per hari",
			targets: []domain.KPITarget{
				kpiTarget(2026, 10, 0.8, 0.7, 3100),
				kpiTarget(2026, 11, 0.9, 0.85, 3000),
			},
			start: civil(2026, 10, 30), end: civil(2026, 11, 2),
			// 2 hari Okt (3100/31 = 100 per hari) + 2 hari Nov (3000/30 = 100 per hari) = 400
			want: &domain.PeriodTarget{SuccessRate: 0.9, SLARate: 0.85, DeliveredPackages: 400},
		},
		{
			name:    "bulan awal tanpa target menyumbang 0",
			targets: []domain.KPITarget{kpiTarget(2026, 11, 0.9, 0.85, 3000)},
			start:   civil(2026, 10, 30), end: civil(2026, 11, 2),
			want: &domain.PeriodTarget{SuccessRate: 0.9, SLARate: 0.85, DeliveredPackages: 200},
		},
		{
			name:    "bulan end tanpa target menghasilkan nil",
			targets: []domain.KPITarget{kpiTarget(2026, 10, 0.8, 0.7, 3100)},
			start:   civil(2026, 10, 30), end: civil(2026, 11, 2),
			want: nil,
		},
		{
			name:    "tanpa target sama sekali menghasilkan nil",
			targets: nil,
			start:   civil(2026, 11, 1), end: civil(2026, 11, 7),
			want: nil,
		},
		{
			name:    "pembulatan tepat 700,5 jadi 701",
			targets: []domain.KPITarget{kpiTarget(2026, 2, 0.9, 0.85, 2802)},
			start:   civil(2026, 2, 1), end: civil(2026, 2, 7),
			// 7 * 2802 / 28 = 700,5
			want: &domain.PeriodTarget{SuccessRate: 0.9, SLARate: 0.85, DeliveredPackages: 701},
		},
		{
			name:    "lintas tahun memakai target Desember dan Januari",
			targets: []domain.KPITarget{kpiTarget(2026, 12, 0.8, 0.7, 3100), kpiTarget(2027, 1, 0.85, 0.75, 3100)},
			start:   civil(2026, 12, 31), end: civil(2027, 1, 1),
			want: &domain.PeriodTarget{SuccessRate: 0.85, SLARate: 0.75, DeliveredPackages: 200},
		},
		{
			name:    "end di hari terakhir bulan memakai bulan itu",
			targets: []domain.KPITarget{kpiTarget(2026, 11, 0.9, 0.85, 3000), kpiTarget(2026, 12, 0.95, 0.9, 3100)},
			start:   civil(2026, 11, 24), end: civil(2026, 11, 30),
			want: &domain.PeriodTarget{SuccessRate: 0.9, SLARate: 0.85, DeliveredPackages: 700},
		},
		{
			name:    "end di hari pertama bulan berikutnya memakai bulan berikutnya",
			targets: []domain.KPITarget{kpiTarget(2026, 11, 0.9, 0.85, 3000), kpiTarget(2026, 12, 0.95, 0.9, 3100)},
			start:   civil(2026, 11, 25), end: civil(2026, 12, 1),
			// 6 hari Nov * 100 + 1 hari Des * 100 = 700
			want: &domain.PeriodTarget{SuccessRate: 0.95, SLARate: 0.9, DeliveredPackages: 700},
		},
		{
			name:    "target nol menghasilkan volume nol bukan nil",
			targets: []domain.KPITarget{kpiTarget(2026, 11, 0, 0, 0)},
			start:   civil(2026, 11, 1), end: civil(2026, 11, 7),
			want: &domain.PeriodTarget{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := domain.TargetForPeriod(tc.targets, tc.start, tc.end)
			if tc.want == nil {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.InDelta(t, tc.want.SuccessRate, got.SuccessRate, 1e-9)
			assert.InDelta(t, tc.want.SLARate, got.SLARate, 1e-9)
			assert.Equal(t, tc.want.DeliveredPackages, got.DeliveredPackages)
		})
	}
}

func TestCompareTarget(t *testing.T) {
	cases := []struct {
		name           string
		actual, target float64
		want           domain.TargetStatus
	}{
		{"sama berarti met", 0.85, 0.85, domain.TargetMet},
		{"di atas met", 0.9, 0.85, domain.TargetMet},
		{"0,84996 terhadap 0,85 below", 0.84996, 0.85, domain.TargetBelow},
		{"nol terhadap nol met", 0, 0, domain.TargetMet},
		{"nol terhadap target positif below", 0, 0.5, domain.TargetBelow},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, domain.CompareTarget(tc.actual, tc.target))
		})
	}
	assert.Equal(t, domain.TargetStatus("met"), domain.TargetMet)
	assert.Equal(t, domain.TargetStatus("below"), domain.TargetBelow)
}

func TestEvaluateTarget(t *testing.T) {
	perf := domain.BranchPerformance{BranchID: "JKT01", TotalPackages: 1000, Delivered: 833, OnTime: 700}
	// success 0,833; sla 700/833 = 0,84034

	t.Run("nil bila tanpa target", func(t *testing.T) {
		assert.Nil(t, domain.EvaluateTarget(perf, nil))
	})
	t.Run("campuran met dan below", func(t *testing.T) {
		pt := &domain.PeriodTarget{SuccessRate: 0.85, SLARate: 0.8, DeliveredPackages: 900}
		got := domain.EvaluateTarget(perf, pt)
		require.NotNil(t, got)
		assert.Equal(t, *pt, got.Target)
		assert.Equal(t, domain.TargetBelow, got.SuccessRateStatus)
		assert.Equal(t, domain.TargetMet, got.SLARateStatus)
		assert.Equal(t, domain.TargetBelow, got.DeliveredStatus)
	})
	t.Run("volume sama dengan target berarti met", func(t *testing.T) {
		got := domain.EvaluateTarget(perf, &domain.PeriodTarget{SuccessRate: 0.5, SLARate: 0.5, DeliveredPackages: 833})
		require.NotNil(t, got)
		assert.Equal(t, domain.TargetMet, got.DeliveredStatus)
		assert.Equal(t, domain.TargetMet, got.SuccessRateStatus)
	})
	t.Run("cabang tanpa paket", func(t *testing.T) {
		got := domain.EvaluateTarget(domain.BranchPerformance{}, &domain.PeriodTarget{SuccessRate: 0.9, SLARate: 0.9, DeliveredPackages: 10})
		require.NotNil(t, got)
		assert.Equal(t, domain.TargetBelow, got.SuccessRateStatus)
		assert.Equal(t, domain.TargetBelow, got.DeliveredStatus)
	})
}
