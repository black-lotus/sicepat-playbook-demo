# Menjalankan seluruh stack lokal dari root repo. Lihat `make help`.
# Dua cara:
#   make dev  : Postgres dan Redis di container dengan port tidak bentrok, API dan web jalan di host (hot reload).
#   make up   : semua service lewat docker compose (butuh port 5432, 6379, 9092, 8080, 5173 kosong).

PG_PORT        ?= 55432
REDIS_PORT     ?= 56379
API_PORT       ?= 18080
WEB_PORT       ?= 5173
NETWORK        ?= sigesit-test
PG_CONTAINER   ?= sigesit-test-pg
REDIS_CONTAINER?= sigesit-test-redis
PG_IMAGE       ?= postgres:16-alpine
REDIS_IMAGE    ?= redis:7-alpine
MIGRATE_IMAGE  ?= migrate/migrate:v4.18.3

DB_URL_HOST      = postgres://sigesit:sigesit@localhost:$(PG_PORT)/sigesit?sslmode=disable
DB_URL_CONTAINER = postgres://sigesit:sigesit@$(PG_CONTAINER):5432/sigesit?sslmode=disable

.PHONY: help dev infra migrate seed api web stop-infra clean-infra up down test lint

help: ## Tampilkan daftar perintah
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "} {printf "  make %-12s %s\n", $$1, $$2}'

dev: infra migrate seed ## Siapkan DB lalu jalankan API dan web sekaligus (Ctrl+C menghentikan keduanya)
	@echo "API http://localhost:$(API_PORT)  |  Web http://localhost:$(WEB_PORT)/courier-performance"
	@trap 'kill 0' INT TERM; \
	  $(MAKE) --no-print-directory api & \
	  $(MAKE) --no-print-directory web & \
	  wait

infra: ## Nyalakan container Postgres dan Redis (dibuat bila belum ada)
	@docker network inspect $(NETWORK) >/dev/null 2>&1 || docker network create $(NETWORK) >/dev/null
	@if docker ps -a --format '{{.Names}}' | grep -qx $(PG_CONTAINER); then docker start $(PG_CONTAINER) >/dev/null; \
	  else docker run -d --name $(PG_CONTAINER) --network $(NETWORK) -e POSTGRES_USER=sigesit -e POSTGRES_PASSWORD=sigesit -e POSTGRES_DB=sigesit -p $(PG_PORT):5432 $(PG_IMAGE) >/dev/null; fi
	@if docker ps -a --format '{{.Names}}' | grep -qx $(REDIS_CONTAINER); then docker start $(REDIS_CONTAINER) >/dev/null; \
	  else docker run -d --name $(REDIS_CONTAINER) --network $(NETWORK) -p $(REDIS_PORT):6379 $(REDIS_IMAGE) >/dev/null; fi
	@echo "Menunggu Postgres siap..."; \
	  for i in $$(seq 1 30); do docker exec $(PG_CONTAINER) pg_isready -U sigesit >/dev/null 2>&1 && exit 0; sleep 1; done; \
	  echo "Postgres tidak siap dalam 30 detik"; exit 1

migrate: ## Terapkan migrasi ke Postgres lokal
	docker run --rm --network $(NETWORK) -v "$(CURDIR)/backend/migrations:/migrations:ro" $(MIGRATE_IMAGE) -path=/migrations -database "$(DB_URL_CONTAINER)" up

seed: ## Isi ulang data contoh 2026 (menghapus data seed lama)
	docker run --rm --network $(NETWORK) -v "$(CURDIR)/backend/seeds:/seeds:ro" $(PG_IMAGE) psql -q -v ON_ERROR_STOP=1 "$(DB_URL_CONTAINER)" -f /seeds/seed.sql

api: ## Jalankan API Go di host
	cd backend && DATABASE_URL="$(DB_URL_HOST)" REDIS_ADDR=localhost:$(REDIS_PORT) HTTP_ADDR=:$(API_PORT) go run ./cmd/api

web: ## Jalankan dev server Vite di host
	cd web && ([ -d node_modules ] || npm install) && API_PROXY_TARGET=http://localhost:$(API_PORT) npm run dev -- --port $(WEB_PORT)

stop-infra: ## Hentikan container Postgres dan Redis (data tetap)
	-docker stop $(PG_CONTAINER) $(REDIS_CONTAINER)

clean-infra: ## Hapus container dan network lokal (data hilang)
	-docker rm -f $(PG_CONTAINER) $(REDIS_CONTAINER)
	-docker network rm $(NETWORK)

up: ## Jalankan semua service lewat docker compose
	docker compose up --build

down: ## Hentikan semua service docker compose
	docker compose down

test: ## Jalankan test BE dan FE
	cd backend && go test ./...
	cd web && npm test

lint: ## Jalankan lint BE dan FE
	cd backend && golangci-lint run
	cd web && npm run lint && npm run typecheck
