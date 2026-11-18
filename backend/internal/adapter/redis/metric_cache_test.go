package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/redis"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

func newCache(t *testing.T) (*redis.MetricCache, *miniredis.Miniredis) {
	t.Helper()
	srv := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: srv.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return redis.NewMetricCache(client), srv
}

func TestMetricCache_RoundTrip(t *testing.T) {
	cache, srv := newCache(t)
	ctx := context.Background()
	page := domain.CourierMetricPage{
		Items: []domain.CourierMetric{{CourierID: "C001", Date: time.Date(2026, 11, 17, 0, 0, 0, 0, time.UTC), TotalPackages: 5}},
		Total: 1,
	}

	_, found, err := cache.GetCourierMetrics(ctx, "k")
	require.NoError(t, err)
	assert.False(t, found, "key belum ada")

	require.NoError(t, cache.SetCourierMetrics(ctx, "k", page, time.Minute))
	got, found, err := cache.GetCourierMetrics(ctx, "k")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, page, got)

	srv.FastForward(2 * time.Minute)
	_, found, err = cache.GetCourierMetrics(ctx, "k")
	require.NoError(t, err)
	assert.False(t, found, "key kedaluwarsa setelah TTL")
}

func TestMetricCache_Errors(t *testing.T) {
	cache, srv := newCache(t)
	ctx := context.Background()

	require.NoError(t, srv.Set("rusak", "bukan-json"))
	_, _, err := cache.GetCourierMetrics(ctx, "rusak")
	assert.Error(t, err)

	srv.Close()
	_, _, err = cache.GetCourierMetrics(ctx, "k")
	assert.Error(t, err)
	assert.Error(t, cache.SetCourierMetrics(ctx, "k", domain.CourierMetricPage{}, time.Minute))
}
