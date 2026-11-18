package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/port/mocks"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/service"
)

func TestBranchPerformanceService_List(t *testing.T) {
	day := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }
	filter := domain.MetricFilter{StartDate: day(10, 30), EndDate: day(11, 2)}
	defaulted := filter.WithDefaults()
	errDB := errors.New("db down")
	page := domain.BranchPerformancePage{
		Total: 2,
		Items: []domain.BranchPerformance{
			{Rank: 1, BranchID: "JKT01", TotalPackages: 100, Delivered: 90, OnTime: 80},
			{Rank: 2, BranchID: "BDG01", TotalPackages: 100, Delivered: 50, OnTime: 40},
		},
	}

	t.Run("target null untuk cabang tanpa target dan ListTargets sekali", func(t *testing.T) {
		metrics := mocks.NewMetricRepository(t)
		targets := mocks.NewTargetRepository(t)
		metrics.EXPECT().ListBranchPerformance(mock.Anything, defaulted).Return(page, nil)
		targets.EXPECT().ListTargets(mock.Anything, []string{"JKT01", "BDG01"}, day(10, 1), day(11, 1)).Return([]domain.KPITarget{
			{BranchID: "JKT01", Month: day(10, 1), SuccessRate: 0.9, SLARate: 0.8, DeliveredPackages: 3100},
			{BranchID: "JKT01", Month: day(11, 1), SuccessRate: 0.95, SLARate: 0.7, DeliveredPackages: 3000},
		}, nil).Once()

		got, err := service.NewBranchPerformanceService(metrics, targets).List(context.Background(), filter)
		require.NoError(t, err)
		assert.Equal(t, 2, got.Total)
		require.Len(t, got.Items, 2)

		jkt := got.Items[0]
		assert.Equal(t, "JKT01", jkt.BranchID)
		require.NotNil(t, jkt.Target)
		// rate dari bulan end (November); volume: 2 hari Okt (3100/31) + 2 hari Nov (3000/30) = 200 + 200
		assert.InDelta(t, 0.95, jkt.Target.Target.SuccessRate, 1e-9)
		assert.Equal(t, 400, jkt.Target.Target.DeliveredPackages)
		assert.Equal(t, domain.TargetBelow, jkt.Target.SuccessRateStatus)
		assert.Equal(t, domain.TargetMet, jkt.Target.SLARateStatus)
		assert.Equal(t, domain.TargetBelow, jkt.Target.DeliveredStatus)

		assert.Nil(t, got.Items[1].Target)
		assert.Equal(t, "BDG01", got.Items[1].BranchID)
	})
	t.Run("halaman kosong tidak memanggil ListTargets", func(t *testing.T) {
		metrics := mocks.NewMetricRepository(t)
		targets := mocks.NewTargetRepository(t)
		metrics.EXPECT().ListBranchPerformance(mock.Anything, defaulted).Return(domain.BranchPerformancePage{Total: 3}, nil)

		got, err := service.NewBranchPerformanceService(metrics, targets).List(context.Background(), filter)
		require.NoError(t, err)
		assert.NotNil(t, got.Items)
		assert.Empty(t, got.Items)
		assert.Equal(t, 3, got.Total)
	})
	t.Run("filter tidak valid", func(t *testing.T) {
		_, err := service.NewBranchPerformanceService(mocks.NewMetricRepository(t), mocks.NewTargetRepository(t)).
			List(context.Background(), domain.MetricFilter{})
		assert.ErrorIs(t, err, domain.ErrInvalidFilter)
	})
	t.Run("error ranking dibungkus", func(t *testing.T) {
		metrics := mocks.NewMetricRepository(t)
		metrics.EXPECT().ListBranchPerformance(mock.Anything, defaulted).Return(domain.BranchPerformancePage{}, errDB)
		_, err := service.NewBranchPerformanceService(metrics, mocks.NewTargetRepository(t)).List(context.Background(), filter)
		assert.ErrorIs(t, err, errDB)
	})
	t.Run("error target dibungkus", func(t *testing.T) {
		metrics := mocks.NewMetricRepository(t)
		targets := mocks.NewTargetRepository(t)
		metrics.EXPECT().ListBranchPerformance(mock.Anything, defaulted).Return(page, nil)
		targets.EXPECT().ListTargets(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errDB)
		_, err := service.NewBranchPerformanceService(metrics, targets).List(context.Background(), filter)
		assert.ErrorIs(t, err, errDB)
	})
}
