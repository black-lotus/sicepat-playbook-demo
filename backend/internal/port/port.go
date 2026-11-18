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

// TargetRepository membaca target KPI bulanan per cabang.
type TargetRepository interface {
	// ListTargets mengembalikan target cabang dalam branchIDs untuk bulan from sampai to (inklusif).
	ListTargets(ctx context.Context, branchIDs []string, from, to time.Time) ([]domain.KPITarget, error)
	// GetTarget mengembalikan target satu cabang pada satu bulan; found false bila tidak ada.
	GetTarget(ctx context.Context, branchID string, month time.Time) (domain.KPITarget, bool, error)
}

// FailedRepository membaca ringkasan gagal kirim dari delivery_events.
type FailedRepository interface {
	// FailedSummary menghitung total paket, total gagal, alasan, dan kurir gagal terbanyak untuk satu periode.
	// PreviousTotalFailed tidak diisi. Reasons dan TopCouriers tidak pernah nil.
	FailedSummary(ctx context.Context, f domain.FailedFilter) (domain.FailedSummary, error)
	// CountFailed menghitung event gagal dalam periode dan filter cabang.
	CountFailed(ctx context.Context, f domain.FailedFilter) (int, error)
}
