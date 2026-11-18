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

func TestFailedDeliveryService_Summary(t *testing.T) {
	day := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 0, 0, 0, 0, time.UTC) }
	filter := domain.FailedFilter{StartDate: day(11, 8), EndDate: day(11, 14), BranchID: "JKT01"}
	errDB := errors.New("db down")

	t.Run("periode pembanding dipakai untuk CountFailed", func(t *testing.T) {
		repo := mocks.NewFailedRepository(t)
		summary := domain.FailedSummary{TotalPackages: 100, TotalFailed: 30}
		repo.EXPECT().FailedSummary(mock.Anything, filter).Return(summary, nil)
		repo.EXPECT().CountFailed(mock.Anything, domain.FailedFilter{StartDate: day(11, 1), EndDate: day(11, 7), BranchID: "JKT01"}).Return(20, nil)

		got, err := service.NewFailedDeliveryService(repo).Summary(context.Background(), filter)
		require.NoError(t, err)
		assert.Equal(t, 20, got.PreviousTotalFailed)
		assert.Equal(t, 30, got.TotalFailed)
	})
	t.Run("filter tidak valid tanpa query", func(t *testing.T) {
		repo := mocks.NewFailedRepository(t)
		_, err := service.NewFailedDeliveryService(repo).Summary(context.Background(), domain.FailedFilter{})
		assert.ErrorIs(t, err, domain.ErrInvalidFilter)
	})
	t.Run("error summary dibungkus", func(t *testing.T) {
		repo := mocks.NewFailedRepository(t)
		repo.EXPECT().FailedSummary(mock.Anything, filter).Return(domain.FailedSummary{}, errDB)
		_, err := service.NewFailedDeliveryService(repo).Summary(context.Background(), filter)
		assert.ErrorIs(t, err, errDB)
	})
	t.Run("error count dibungkus", func(t *testing.T) {
		repo := mocks.NewFailedRepository(t)
		repo.EXPECT().FailedSummary(mock.Anything, filter).Return(domain.FailedSummary{}, nil)
		repo.EXPECT().CountFailed(mock.Anything, mock.Anything).Return(0, errDB)
		_, err := service.NewFailedDeliveryService(repo).Summary(context.Background(), filter)
		assert.ErrorIs(t, err, errDB)
	})
}
