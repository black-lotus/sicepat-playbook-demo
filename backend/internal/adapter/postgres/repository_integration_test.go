//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/postgres"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// Test ini butuh database yang sudah dimigrasi, misalnya dari docker compose:
// TEST_DATABASE_URL=postgres://sigesit:sigesit@localhost:5432/sigesit?sslmode=disable
func TestRepositories_AggregateAndList(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL tidak diisi")
	}
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, "TRUNCATE delivery_events, courier_daily_metrics, couriers")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "INSERT INTO couriers (id, name, branch_id) VALUES ('C001', 'Budi', 'JKT01')")
	require.NoError(t, err)

	day := time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC)
	due := day.Add(17 * time.Hour)
	events := []domain.DeliveryEvent{
		{AWBNumber: "A1", CourierID: "C001", BranchID: "JKT01", Status: domain.DeliveryStatusDelivered, OccurredAt: day.Add(9 * time.Hour), DueAt: due},
		{AWBNumber: "A2", CourierID: "C001", BranchID: "JKT01", Status: domain.DeliveryStatusDelivered, OccurredAt: day.Add(18 * time.Hour), DueAt: due},
		{AWBNumber: "A3", CourierID: "C001", BranchID: "JKT01", Status: domain.DeliveryStatusFailed, FailedReason: "alamat tidak ditemukan", OccurredAt: day.Add(10 * time.Hour)},
	}
	deliveries := postgres.NewDeliveryRepository(pool)
	for _, e := range events {
		require.NoError(t, deliveries.SaveEvent(ctx, e))
	}
	// Event duplikat diabaikan.
	require.NoError(t, deliveries.SaveEvent(ctx, events[0]))

	metrics := postgres.NewMetricRepository(pool)
	require.NoError(t, metrics.AggregateDaily(ctx, day))

	page, err := metrics.ListCourierMetrics(ctx, domain.MetricFilter{StartDate: day, EndDate: day, Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	assert.Equal(t, 1, page.Total)

	got := page.Items[0]
	assert.Equal(t, "Budi", got.CourierName)
	assert.Equal(t, 3, got.TotalPackages)
	assert.Equal(t, 2, got.Delivered)
	assert.Equal(t, 1, got.Failed)
	assert.Equal(t, 1, got.OnTime)
}
