// Package port berisi interface yang menghubungkan service dengan infrastruktur.
package port

import (
	"context"
	"time"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// MetricRepository membaca dan menghitung agregat KPI kurir.
type MetricRepository interface {
	ListCourierMetrics(ctx context.Context, filter domain.MetricFilter) (domain.CourierMetricPage, error)
	ListCourierPerformance(ctx context.Context, filter domain.MetricFilter) (domain.CourierPerformancePage, error)
	AggregateDaily(ctx context.Context, date time.Time) error
}

// DeliveryRepository menyimpan event pengiriman paket.
type DeliveryRepository interface {
	SaveEvent(ctx context.Context, event domain.DeliveryEvent) error
}

// MetricCache menyimpan hasil query metrik untuk sementara.
type MetricCache interface {
	GetCourierMetrics(ctx context.Context, key string) (page domain.CourierMetricPage, found bool, err error)
	SetCourierMetrics(ctx context.Context, key string, page domain.CourierMetricPage, ttl time.Duration) error
}
