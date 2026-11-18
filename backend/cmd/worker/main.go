// Command worker menghitung agregat KPI harian kurir secara berkala.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/postgres"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/redis"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/config"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/domain"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("worker berhenti dengan error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := config.NewLogger(cfg.LogLevel)

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	redisClient := goredis.NewClient(&goredis.Options{Addr: cfg.RedisAddr})
	defer redisClient.Close()

	svc := service.NewMetricService(postgres.NewMetricRepository(pool), redis.NewMetricCache(redisClient), cfg.CacheTTL, logger)

	logger.Info("worker berjalan", slog.Duration("interval", cfg.AggregateInterval))
	ticker := time.NewTicker(cfg.AggregateInterval)
	defer ticker.Stop()

	for {
		aggregate(ctx, svc, time.Now(), logger)
		select {
		case <-ctx.Done():
			logger.Info("menghentikan worker")
			return nil
		case <-ticker.C:
		}
	}
}

// aggregate menghitung ulang KPI hari ini dan kemarin (tanggal WIB), karena event bisa datang terlambat.
func aggregate(ctx context.Context, svc *service.MetricService, now time.Time, logger *slog.Logger) {
	now = now.In(domain.WIB)
	for _, day := range []time.Time{now.AddDate(0, 0, -1), now} {
		if err := svc.AggregateDaily(ctx, day); err != nil {
			logger.ErrorContext(ctx, "agregasi gagal", slog.Any("error", err))
		}
	}
}
