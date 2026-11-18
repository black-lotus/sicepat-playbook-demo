// Package service berisi use case aplikasi monitoring performa SiGesit.
package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/port"
)

// MetricService menyajikan dan menghitung KPI performa kurir.
type MetricService struct {
	repo     port.MetricRepository
	cache    port.MetricCache
	cacheTTL time.Duration
	logger   *slog.Logger
}

// NewMetricService membuat MetricService.
func NewMetricService(repo port.MetricRepository, cache port.MetricCache, cacheTTL time.Duration, logger *slog.Logger) *MetricService {
	return &MetricService{repo: repo, cache: cache, cacheTTL: cacheTTL, logger: logger}
}

// ListCourierMetrics mengembalikan KPI kurir sesuai filter, dari cache bila ada.
// Kegagalan cache hanya dicatat di log dan tidak menggagalkan request.
func (s *MetricService) ListCourierMetrics(ctx context.Context, filter domain.MetricFilter) (domain.CourierMetricPage, error) {
	filter = filter.WithDefaults()
	if err := filter.Validate(); err != nil {
		return domain.CourierMetricPage{}, err
	}

	key := filter.CacheKey()
	page, found, err := s.cache.GetCourierMetrics(ctx, key)
	if err != nil {
		s.logger.WarnContext(ctx, "baca cache metrik kurir gagal", slog.String("key", key), slog.Any("error", err))
	} else if found {
		return page, nil
	}

	page, err = s.repo.ListCourierMetrics(ctx, filter)
	if err != nil {
		return domain.CourierMetricPage{}, fmt.Errorf("ambil metrik kurir: %w", err)
	}

	if err := s.cache.SetCourierMetrics(ctx, key, page, s.cacheTTL); err != nil {
		s.logger.WarnContext(ctx, "simpan cache metrik kurir gagal", slog.String("key", key), slog.Any("error", err))
	}
	return page, nil
}

// ListCourierPerformance mengembalikan ranking performa kurir sesuai filter.
// Endpoint ini belum memakai cache.
func (s *MetricService) ListCourierPerformance(ctx context.Context, filter domain.MetricFilter) (domain.CourierPerformancePage, error) {
	filter = filter.WithDefaults()
	if err := filter.Validate(); err != nil {
		return domain.CourierPerformancePage{}, err
	}
	page, err := s.repo.ListCourierPerformance(ctx, filter)
	if err != nil {
		return domain.CourierPerformancePage{}, fmt.Errorf("ambil performa kurir: %w", err)
	}
	return page, nil
}

// AggregateDaily menghitung ulang KPI harian semua kurir untuk tanggal WIB dari date.
func (s *MetricService) AggregateDaily(ctx context.Context, date time.Time) error {
	day := domain.DayStartWIB(date)
	if err := s.repo.AggregateDaily(ctx, day); err != nil {
		return fmt.Errorf("agregasi metrik harian %s: %w", day.Format(time.DateOnly), err)
	}
	return nil
}
