package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/port/mocks"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/service"
)

func TestAggregate_UsesWIBDays(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	// 2026-11-17 17:30 UTC = 2026-11-18 00:30 WIB: "hari ini" adalah 18 Nov, "kemarin" 17 Nov (WIB).
	now := time.Date(2026, 11, 17, 17, 30, 0, 0, time.UTC)

	tests := []struct {
		name    string
		repoErr error
	}{
		{name: "berhasil"},
		{name: "error agregasi tidak menghentikan hari berikutnya", repoErr: errors.New("db down")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMetricRepository(t)
			repo.EXPECT().AggregateDaily(mock.Anything, time.Date(2026, 11, 17, 0, 0, 0, 0, domain.WIB)).Return(tt.repoErr).Once()
			repo.EXPECT().AggregateDaily(mock.Anything, time.Date(2026, 11, 18, 0, 0, 0, 0, domain.WIB)).Return(tt.repoErr).Once()
			svc := service.NewMetricService(repo, mocks.NewMetricCache(t), time.Minute, logger)

			aggregate(context.Background(), svc, now, logger)
		})
	}
}
