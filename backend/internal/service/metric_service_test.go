package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/port/mocks"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/service"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestMetricService_ListCourierMetrics(t *testing.T) {
	const ttl = 5 * time.Minute
	filter := domain.MetricFilter{
		StartDate: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC),
	}
	withDefaults := filter.WithDefaults()
	key := withDefaults.CacheKey()
	page := domain.CourierMetricPage{
		Items: []domain.CourierMetric{{CourierID: "C001", TotalPackages: 10, Delivered: 9}},
		Total: 1,
	}
	errDB := errors.New("db down")
	errCache := errors.New("redis down")

	tests := []struct {
		name    string
		filter  domain.MetricFilter
		setup   func(repo *mocks.MetricRepository, cache *mocks.MetricCache)
		want    domain.CourierMetricPage
		wantErr error
	}{
		{
			name:   "cache hit tidak memanggil repository",
			filter: filter,
			setup: func(_ *mocks.MetricRepository, cache *mocks.MetricCache) {
				cache.EXPECT().GetCourierMetrics(mock.Anything, key).Return(page, true, nil)
			},
			want: page,
		},
		{
			name:   "cache miss mengambil dari repository lalu menyimpan ke cache",
			filter: filter,
			setup: func(repo *mocks.MetricRepository, cache *mocks.MetricCache) {
				cache.EXPECT().GetCourierMetrics(mock.Anything, key).Return(domain.CourierMetricPage{}, false, nil)
				repo.EXPECT().ListCourierMetrics(mock.Anything, withDefaults).Return(page, nil)
				cache.EXPECT().SetCourierMetrics(mock.Anything, key, page, ttl).Return(nil)
			},
			want: page,
		},
		{
			name:   "cache error tetap mengambil dari repository",
			filter: filter,
			setup: func(repo *mocks.MetricRepository, cache *mocks.MetricCache) {
				cache.EXPECT().GetCourierMetrics(mock.Anything, key).Return(domain.CourierMetricPage{}, false, errCache)
				repo.EXPECT().ListCourierMetrics(mock.Anything, withDefaults).Return(page, nil)
				cache.EXPECT().SetCourierMetrics(mock.Anything, key, page, ttl).Return(errCache)
			},
			want: page,
		},
		{
			name:   "repository error dikembalikan",
			filter: filter,
			setup: func(repo *mocks.MetricRepository, cache *mocks.MetricCache) {
				cache.EXPECT().GetCourierMetrics(mock.Anything, key).Return(domain.CourierMetricPage{}, false, nil)
				repo.EXPECT().ListCourierMetrics(mock.Anything, withDefaults).Return(domain.CourierMetricPage{}, errDB)
			},
			wantErr: errDB,
		},
		{
			name:    "filter tidak valid ditolak sebelum menyentuh cache",
			filter:  domain.MetricFilter{},
			setup:   func(*mocks.MetricRepository, *mocks.MetricCache) {},
			wantErr: domain.ErrInvalidFilter,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMetricRepository(t)
			cache := mocks.NewMetricCache(t)
			tt.setup(repo, cache)
			svc := service.NewMetricService(repo, cache, ttl, discardLogger)

			got, err := svc.ListCourierMetrics(context.Background(), tt.filter)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMetricService_AggregateDaily(t *testing.T) {
	input := time.Date(2026, 11, 17, 17, 30, 0, 0, time.UTC) // 2026-11-18 00:30 WIB
	wantDay := time.Date(2026, 11, 18, 0, 0, 0, 0, domain.WIB)
	errDB := errors.New("db down")

	tests := []struct {
		name    string
		repoErr error
	}{
		{name: "berhasil"},
		{name: "repository error", repoErr: errDB},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMetricRepository(t)
			repo.EXPECT().AggregateDaily(mock.Anything, wantDay).Return(tt.repoErr)
			svc := service.NewMetricService(repo, mocks.NewMetricCache(t), time.Minute, discardLogger)

			err := svc.AggregateDaily(context.Background(), input)
			if tt.repoErr != nil {
				assert.ErrorIs(t, err, tt.repoErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
