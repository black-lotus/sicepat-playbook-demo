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

func TestTargetService_GetTarget(t *testing.T) {
	nov := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	target := domain.KPITarget{BranchID: "JKT01", Month: nov, SuccessRate: 0.9, SLARate: 0.85, DeliveredPackages: 3000}
	errDB := errors.New("db down")

	tests := []struct {
		name    string
		setup   func(*mocks.TargetRepository)
		want    domain.KPITarget
		wantErr error
	}{
		{"ada", func(r *mocks.TargetRepository) {
			r.EXPECT().GetTarget(mock.Anything, "JKT01", nov).Return(target, true, nil)
		}, target, nil},
		{"tidak ada", func(r *mocks.TargetRepository) {
			r.EXPECT().GetTarget(mock.Anything, "JKT01", nov).Return(domain.KPITarget{}, false, nil)
		}, domain.KPITarget{}, domain.ErrTargetNotFound},
		{"error repo dibungkus", func(r *mocks.TargetRepository) {
			r.EXPECT().GetTarget(mock.Anything, "JKT01", nov).Return(domain.KPITarget{}, false, errDB)
		}, domain.KPITarget{}, errDB},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := mocks.NewTargetRepository(t)
			tc.setup(repo)
			got, err := service.NewTargetService(repo).GetTarget(context.Background(), "JKT01", nov)
			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTargetService_RateTarget(t *testing.T) {
	dec := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	target := domain.KPITarget{BranchID: "JKT01", Month: dec, SuccessRate: 0.9}

	t.Run("bulan end memakai tanggal WIB", func(t *testing.T) {
		repo := mocks.NewTargetRepository(t)
		// 2026-11-30 17:30 UTC adalah 1 Desember 00:30 WIB.
		end := time.Date(2026, 11, 30, 17, 30, 0, 0, time.UTC)
		repo.EXPECT().GetTarget(mock.Anything, "JKT01", dec).Return(target, true, nil)
		got, err := service.NewTargetService(repo).RateTarget(context.Background(), "JKT01", end)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, target, *got)
	})
	t.Run("tanpa target nil", func(t *testing.T) {
		repo := mocks.NewTargetRepository(t)
		repo.EXPECT().GetTarget(mock.Anything, "BDG01", dec).Return(domain.KPITarget{}, false, nil)
		got, err := service.NewTargetService(repo).RateTarget(context.Background(), "BDG01", dec)
		require.NoError(t, err)
		assert.Nil(t, got)
	})
	t.Run("error repo", func(t *testing.T) {
		repo := mocks.NewTargetRepository(t)
		errDB := errors.New("db down")
		repo.EXPECT().GetTarget(mock.Anything, "JKT01", dec).Return(domain.KPITarget{}, false, errDB)
		_, err := service.NewTargetService(repo).RateTarget(context.Background(), "JKT01", dec)
		assert.ErrorIs(t, err, errDB)
	})
}
