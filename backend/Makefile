DATABASE_URL ?= postgres://sigesit:sigesit@localhost:5432/sigesit?sslmode=disable
MIGRATE_IMAGE := migrate/migrate:v4.18.3

.PHONY: run-api run-worker run-consumer test test-race test-integration cover lint mocks migrate-up migrate-down

run-api:
	DATABASE_URL=$(DATABASE_URL) go run ./cmd/api

run-worker:
	DATABASE_URL=$(DATABASE_URL) go run ./cmd/worker

run-consumer:
	DATABASE_URL=$(DATABASE_URL) go run ./cmd/consumer

test:
	go test ./...

test-race:
	go test -race ./...

# PERINGATAN: test ini men-TRUNCATE tabel delivery_events, courier_daily_metrics, dan couriers.
# Jangan arahkan DATABASE_URL ke database yang berisi data yang ingin disimpan.
test-integration:
	TEST_DATABASE_URL=$(DATABASE_URL) go test -tags=integration ./...

# Mock hasil generate tidak dihitung dalam cakupan.
cover:
	go test -coverprofile=coverage.out $$(go list ./... | grep -v /mocks)
	go tool cover -func=coverage.out | tail -1

lint:
	golangci-lint run

mocks:
	go tool mockery

migrate-up:
	docker run --rm --network host -v $(CURDIR)/migrations:/migrations $(MIGRATE_IMAGE) -path=/migrations -database "$(DATABASE_URL)" up

migrate-down:
	docker run --rm --network host -v $(CURDIR)/migrations:/migrations $(MIGRATE_IMAGE) -path=/migrations -database "$(DATABASE_URL)" down 1
