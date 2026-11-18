// Command api menjalankan REST API untuk dashboard monitoring performa SiGesit.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/httpapi"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/postgres"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/redis"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/config"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api berhenti dengan error", slog.Any("error", err))
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

	metricService := service.NewMetricService(
		postgres.NewMetricRepository(pool),
		redis.NewMetricCache(redisClient),
		cfg.CacheTTL,
		logger,
	)
	targetService := service.NewTargetService(postgres.NewTargetRepository(pool))
	failedService := service.NewFailedDeliveryService(postgres.NewFailedRepository(pool))
	router := httpapi.NewRouter(httpapi.Handlers{
		CourierMetrics:     httpapi.NewCourierMetricHandler(metricService, logger),
		CourierPerformance: httpapi.NewCourierPerformanceHandler(metricService, targetService, logger),
		KPITargets:         httpapi.NewKPITargetHandler(targetService, logger),
		FailedDeliveries:   httpapi.NewFailedDeliveryHandler(failedService, logger),
	}, logger)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("api berjalan", slog.String("addr", cfg.HTTPAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	logger.Info("menghentikan api")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}
