package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/config"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		check   func(t *testing.T, cfg config.Config)
		wantErr bool
	}{
		{
			name: "default",
			env:  map[string]string{"DATABASE_URL": "postgres://x"},
			check: func(t *testing.T, cfg config.Config) {
				assert.Equal(t, ":8080", cfg.HTTPAddr)
				assert.Equal(t, []string{"localhost:9092"}, cfg.KafkaBrokers)
				assert.Equal(t, 5*time.Minute, cfg.CacheTTL)
			},
		},
		{
			name: "override",
			env:  map[string]string{"DATABASE_URL": "postgres://x", "KAFKA_BROKERS": "k1:9092,k2:9092", "CACHE_TTL": "30s"},
			check: func(t *testing.T, cfg config.Config) {
				assert.Equal(t, []string{"k1:9092", "k2:9092"}, cfg.KafkaBrokers)
				assert.Equal(t, 30*time.Second, cfg.CacheTTL)
			},
		},
		{name: "tanpa DATABASE_URL", env: map[string]string{}, wantErr: true},
		{name: "durasi salah", env: map[string]string{"DATABASE_URL": "postgres://x", "AGGREGATE_INTERVAL": "lima"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{"DATABASE_URL", "KAFKA_BROKERS", "CACHE_TTL", "AGGREGATE_INTERVAL", "HTTP_ADDR"} {
				t.Setenv(k, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg, err := config.Load()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, cfg)
		})
	}
}

func TestNewLogger(t *testing.T) {
	assert.NotNil(t, config.NewLogger("debug"))
	assert.NotNil(t, config.NewLogger("tidak-dikenal"))
}
