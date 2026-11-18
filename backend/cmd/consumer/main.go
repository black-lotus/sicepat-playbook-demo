// Command consumer membaca event pengiriman dari Kafka dan menyimpannya ke PostgreSQL.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/kafka"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/adapter/postgres"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/config"
	"github.com/black-lotus/sicepat-playbook-demo/backend/internal/service"
)

func main() {
	if err := run(); err != nil {
		slog.Error("consumer berhenti dengan error", slog.Any("error", err))
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

	consumer := kafka.NewDeliveryConsumer(
		cfg.KafkaBrokers,
		cfg.KafkaTopic,
		cfg.KafkaGroupID,
		service.NewDeliveryService(postgres.NewDeliveryRepository(pool)),
		logger,
	)
	defer consumer.Close()

	logger.Info("consumer berjalan", slog.String("topic", cfg.KafkaTopic))
	return consumer.Run(ctx)
}
