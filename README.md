# Monitoring Performa SiGesit

Dashboard untuk memantau performa SiGesit (kurir SiCepat). Konvensi tim ada di [CLAUDE.md](CLAUDE.md).

- `backend/`: API, worker, dan consumer (Go).
- `web/`: dashboard frontend (React, TypeScript, Vite).

## Menjalankan secara lokal

Dengan Docker Compose (detail di bagian "Local development" di [CLAUDE.md](CLAUDE.md)):
```sh
cp .env.example .env
docker compose up --build   # lalu buka http://localhost:5173/courier-performance
```

Tanpa Docker untuk aplikasinya (infrastruktur tetap dari Compose):
```sh
docker compose up -d postgres redis kafka migrate
cd backend && make run-api        # juga: make run-worker, make run-consumer
cd web && nvm use && npm install && npm run dev
```
Bila port 8080 sudah dipakai, jalankan API dengan `HTTP_ADDR=:18080 make run-api`
dan frontend dengan `API_PROXY_TARGET=http://localhost:18080 npm run dev`.

## Test

```sh
cd backend && make test && make test-integration   # integration butuh PostgreSQL yang sudah dimigrasi
cd web && npm test && npm run lint && npm run typecheck
```
