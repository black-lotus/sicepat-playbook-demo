// Package redis berisi implementasi cache dengan Redis (go-redis).
package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
)

// MetricCache mengimplementasikan port.MetricCache.
type MetricCache struct {
	client *goredis.Client
}

// NewMetricCache membuat MetricCache.
func NewMetricCache(client *goredis.Client) *MetricCache {
	return &MetricCache{client: client}
}

// GetCourierMetrics membaca halaman metrik dari cache. found bernilai false bila key tidak ada.
func (c *MetricCache) GetCourierMetrics(ctx context.Context, key string) (domain.CourierMetricPage, bool, error) {
	raw, err := c.client.Get(ctx, key).Bytes()
	if errors.Is(err, goredis.Nil) {
		return domain.CourierMetricPage{}, false, nil
	}
	if err != nil {
		return domain.CourierMetricPage{}, false, fmt.Errorf("redis get %s: %w", key, err)
	}

	var page domain.CourierMetricPage
	if err := json.Unmarshal(raw, &page); err != nil {
		return domain.CourierMetricPage{}, false, fmt.Errorf("decode cache %s: %w", key, err)
	}
	return page, true, nil
}

// SetCourierMetrics menyimpan halaman metrik ke cache dengan TTL.
func (c *MetricCache) SetCourierMetrics(ctx context.Context, key string, page domain.CourierMetricPage, ttl time.Duration) error {
	raw, err := json.Marshal(page)
	if err != nil {
		return fmt.Errorf("encode cache %s: %w", key, err)
	}
	if err := c.client.Set(ctx, key, raw, ttl).Err(); err != nil {
		return fmt.Errorf("redis set %s: %w", key, err)
	}
	return nil
}
