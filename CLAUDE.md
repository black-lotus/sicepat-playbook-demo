# Konvensi tim

## Tentang aplikasi
Aplikasi ini memantau performa SiGesit (kurir SiCepat): jumlah paket terkirim, tingkat keberhasilan
delivery, ketepatan waktu (SLA), paket gagal kirim beserta alasannya, dan produktivitas per kurir,
per cabang, dan per periode.

## Struktur repo
Mono repo dengan dua bagian:
- `web/` berisi frontend dashboard. Konvensinya ada di bagian **Frontend**.
- `backend/` berisi API, worker, dan consumer. Konvensinya ada di bagian **Backend**.

Alur data secara umum: event status paket dan aktivitas kurir masuk ke backend, backend menghitung
agregat KPI, lalu frontend menampilkannya lewat REST API.

Saat bekerja di satu bagian, ikuti hanya konvensi bagian itu. Perubahan yang menyentuh kontrak API
(endpoint, bentuk request atau response) harus disepakati kedua sisi.

## Local development (Docker Compose)
Seluruh stack lokal dijalankan dengan satu perintah dari root repo. Prasyarat: Docker dengan Compose v2.

Setup pertama kali: `cp .env.example .env`, lalu isi `GO_VERSION` dan `NODE_VERSION` (samakan dengan `backend/go.mod` dan `web/package.json`). Compose menolak jalan bila keduanya kosong. Dockerfile dev ada di `docker/`.

Service di `docker-compose.yml`:

| Service | Isi | Port host | Perintah di dalam container |
|---|---|---|---|
| `backend` | REST API Go | `8080` | `go run ./cmd/api` |
| `worker` | Agregasi KPI harian berkala | - | `go run ./cmd/worker` |
| `consumer` | Membaca event pengiriman dari Kafka | - | `go run ./cmd/consumer` |
| `web` | Dev server Vite | `5173` | `npm run dev -- --host` |
| `postgres` | PostgreSQL 16 | `5432` | - |
| `migrate` | Menjalankan `backend/migrations` sekali lalu berhenti | - | `migrate up` |
| `redis` | Redis 7 untuk cache | `6379` | - |
| `kafka` | Kafka (KRaft) | `9092` | - |

Service backend (`backend`, `worker`, `consumer`) baru jalan setelah `migrate` selesai. Bila port host default sudah dipakai aplikasi lain, ubah lewat `.env` (`BACKEND_HOST_PORT`, `WEB_HOST_PORT`, `POSTGRES_HOST_PORT`, `REDIS_HOST_PORT`, `KAFKA_HOST_PORT`); port di dalam Compose tidak berubah. Di dalam Compose, Kafka diakses lewat `kafka:29092`; dari host lewat `localhost:9092`.

Perintah umum:
```bash
docker compose up --build        # jalankan semua service, lalu buka http://localhost:5173/courier-performance
docker compose up backend        # hanya BE, untuk tim BE
docker compose logs -f backend   # ikuti log satu service
docker compose down              # hentikan dan hapus container
docker compose down -v           # sekaligus hapus volume (reset node_modules dan cache)
```

Konvensi:
- Source `backend/` dan `web/` di-mount sebagai volume supaya perubahan kode langsung terbaca (hot reload di `web`, restart `backend` dengan `docker compose restart backend`).
- `node_modules` memakai named volume sendiri, jangan di-mount dari host, supaya binary host (macOS) dan container (Linux) tidak bentrok.
- Versi image Go dan Node diatur lewat `.env` dan mengikuti `backend/go.mod` dan `web/package.json`. Jangan pakai tag `latest`.
- `web` memanggil API lewat proxy Vite `/api`, bukan langsung ke host lain, jadi tidak perlu CORS dan `VITE_API_BASE_URL` dibiarkan kosong. Di dalam Compose, target proxy adalah `http://backend:8080` (bukan `localhost`), diatur lewat environment variable `API_PROXY_TARGET` yang dibaca `web/vite.config.ts` dengan default `http://localhost:8080` untuk jalan tanpa Docker.
- Konfigurasi lewat environment variable di `docker-compose.yml` (`HTTP_ADDR` untuk `backend`). Rahasia dan nilai khusus mesin ditaruh di `.env` yang tidak di-commit; sediakan `.env.example` bila ada variabel baru.
- Pakai data contoh atau data sintetis. Jangan memasukkan data pribadi kurir atau penerima (nomor HP, NIK, alamat) ke seed, fixture, maupun volume lokal.
- Menjalankan test tetap lewat perintah di bagian Test masing-masing. Untuk di dalam container: `docker compose exec web npm test` atau `docker compose exec backend go test ./...`.
- Compose dipakai untuk development lokal saja, bukan untuk deployment.

## Frontend

### Arsitektur singkat
- Bahasa: TypeScript (mode `strict`). Versi Node mengikuti `web/.nvmrc`, dependensi mengikuti `web/package.json`.
- Stack: React, Vite, React Router, TanStack Query untuk data fetching dan cache.
- Struktur folder di `web/src/`:
  - `app/` berisi entry, router, dan provider global.
  - `pages/` berisi satu folder per halaman (`courier-performance`, `branch-summary`, `failed-delivery`).
  - `features/` berisi logika per fitur: komponen, hook, dan tipe yang khusus untuk fitur itu.
  - `components/` berisi komponen UI generik yang dipakai lintas fitur (tabel, kartu KPI, filter tanggal).
  - `api/` berisi client HTTP, tipe response, dan hook query (`useCourierMetrics`).
  - `lib/` berisi helper murni, misalnya format angka, persen, dan tanggal.
- Alur data: halaman memanggil hook query di `api/`, hook memanggil client HTTP, komponen hanya menerima data lewat props.
- Base URL API dibaca dari environment variable `VITE_API_BASE_URL`. Saat development, biarkan kosong: Vite meneruskan `/api` ke `API_PROXY_TARGET` (default `http://localhost:8080`).

### Konvensi kode
Gaya penamaan:
- Komponen dan tipe memakai PascalCase (`CourierTable`, `PerformanceMetric`). Fungsi, variabel, dan hook memakai camelCase, hook diawali `use`.
- File komponen memakai PascalCase (`CourierTable.tsx`). File lain memakai kebab-case (`format-percent.ts`). File test memakai akhiran `.test.ts` atau `.test.tsx` di sebelah file yang diuji.
- Akronim ditulis kapital penuh pada tipe dan konstanta (`AWB`, `SLA`), camelCase biasa pada variabel (`awbNumber`).

Struktur folder:
- Ikuti struktur di atas. `components/` tidak boleh meng-import dari `features/` atau `pages/`.
- Fetch data hanya lewat hook di `api/`, jangan `fetch` langsung dari komponen.

Pola yang dihindari:
- Tipe `any` dan `@ts-ignore`. Pakai `unknown` lalu persempit tipenya.
- Menghitung ulang KPI di frontend. Angka KPI berasal dari API; frontend hanya memformat.
- Menyimpan data server di state global. Data server dikelola TanStack Query.
- Menampilkan atau me-log data pribadi kurir dan penerima (nomor HP, NIK, alamat) dalam bentuk mentah.
- Halaman tanpa state loading, error, dan data kosong.

### Test
Framework dan cara menjalankannya (dari folder `web/`):
- Vitest dengan React Testing Library. Response API di-mock dengan MSW.
- Unit dan komponen test dijalankan dengan `npm test`.
- Lint dan cek tipe dijalankan dengan `npm run lint` dan `npm run typecheck` sebelum push.

Standar cakupan minimal:
- 80% untuk `lib/` dan `api/`.
- 60% untuk seluruh frontend. Cek dengan `npm run test:coverage`.
- Setiap halaman diuji untuk state loading, error, data kosong, dan data terisi.

## Backend

### Arsitektur singkat
Bahasa dan versi:
- Go, dengan versi yang tertulis di `backend/go.mod`.

Modul utama dan alurnya:
- Arsitektur hexagonal (ports and adapters) di `backend/`:
  - `cmd/` berisi entrypoint per binary (`api`, `worker`, `consumer`).
  - `internal/domain/` berisi entity dan aturan bisnis (misalnya `Courier`, `Delivery`, `PerformanceMetric`), tanpa dependensi ke infrastruktur.
  - `internal/port/` berisi interface untuk repository, publisher, dan cache.
  - `internal/service/` berisi use case, misalnya menghitung skor performa atau merangkum KPI harian.
  - `internal/adapter/` berisi implementasi port: `httpapi` (handler REST), `postgres`, `redis`, `kafka`.
- Alur data:
  1. Event status paket dan aktivitas kurir masuk lewat Kafka.
  2. Consumer menyimpannya ke PostgreSQL.
  3. Worker menghitung agregat KPI secara berkala, lalu menyimpan hasilnya di PostgreSQL dan cache Redis.
  4. API menyajikan data KPI ke frontend dashboard.

Dependensi penting:
- PostgreSQL (`pgx`), Redis (`go-redis`), Kafka, OpenTelemetry untuk tracing, dan logging terstruktur (`log/slog`).

### Konvensi API
- REST dengan JSON, diberi prefix versi (`/api/v1/...`). Path memakai kebab-case dan kata benda jamak (`/api/v1/couriers/{id}/metrics`).
- Field JSON memakai snake_case (`awb_number`, `success_rate`).
- Response sukses berbentuk `{"data": ..., "meta": ...}`. `meta` berisi info pagination (`page`, `page_size`, `total`).
- Response error berbentuk `{"error": {"code": "...", "message": "..."}}` dengan status HTTP yang sesuai (400, 404, 422, 500). Pesan error internal tidak dikirim ke client.
- Filter periode memakai `start_date` dan `end_date` dengan format ISO 8601. Semua waktu disimpan dalam UTC. Batas hari laporan dan batas SLA adalah aturan bisnis yang dihitung backend dalam WIB (UTC+7); `start_date` dan `end_date` adalah tanggal WIB.
- Angka KPI dihitung di backend, karena frontend hanya memformat.
- Perubahan kontrak API (endpoint, bentuk request atau response) harus disepakati dengan tim frontend dan didokumentasikan di spesifikasi OpenAPI (`backend/api/openapi.yaml`).

### Konvensi kode
Gaya penamaan:
- Nama package memakai huruf kecil, satu kata, tanpa underscore (`courier`, `metric`).
- Nama yang diekspor memakai PascalCase, yang tidak diekspor memakai camelCase. Akronim ditulis kapital penuh (`AWB`, `ID`, `HTTP`).
- Interface diberi nama sesuai perilakunya (`CourierRepository`, `MetricPublisher`), tanpa prefix `I`.
- Nama file memakai snake_case (`courier_repository.go`). File test memakai akhiran `_test.go`.

Struktur folder:
- Ikuti struktur hexagonal di atas. `internal/domain` tidak boleh meng-import `internal/adapter`.
- Konfigurasi dibaca dari environment variable di `internal/config`.
- Migrasi database disimpan di `backend/migrations/`.

Pola yang dihindari:
- `panic` di kode aplikasi. Selalu kembalikan `error` dan bungkus dengan konteks (`fmt.Errorf("hitung skor kurir: %w", err)`).
- Variabel global yang bisa diubah dan `init()` yang punya efek samping. Gunakan dependency injection lewat konstruktor.
- Mengabaikan `context.Context`. Setiap fungsi yang melakukan I/O menerima `ctx` sebagai parameter pertama.
- Query N+1 saat menghitung metrik per kurir. Gunakan query agregat atau batch.
- Endpoint list tanpa pagination atau tanpa batas periode.
- Me-log data pribadi kurir dan penerima (nomor HP, NIK, alamat) dalam bentuk mentah.

### Test
Framework dan cara menjalankannya (dari folder `backend/`):
- Memakai `testing` bawaan Go dengan pola table-driven, ditambah `testify` untuk assertion dan mock dari `mockery`. Mock di `internal/port/mocks` dibuat ulang dengan `make mocks` (`go tool mockery`, versinya terkunci di `go.mod`) setiap kali interface di `internal/port` berubah.
- Unit test dijalankan dengan `go test ./...`, dan race detector dengan `go test -race ./...`.
- Integration test (PostgreSQL, Redis, Kafka) diberi build tag `integration` dan dijalankan dengan `go test -tags=integration ./...`.
- Handler HTTP diuji dengan `net/http/httptest`, termasuk bentuk response sukses dan error.
- Lint dijalankan dengan `golangci-lint run` sebelum push.

Standar cakupan minimal:
- 80% untuk `internal/domain` dan `internal/service`.
- 60% untuk seluruh backend. Cek dengan `go test -cover ./...`.

## Commit dan pull request
Format pesan commit:
- Conventional Commits: `<type>(<scope>): <ringkasan>`, misalnya `feat(web): tambah filter cabang di halaman performa kurir`.
- Scope minimal menyebut bagiannya: `web`, `backend`, atau nama modul di dalamnya.
- Type yang dipakai: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`, `style`.
- Sertakan ID task ClickUp di body commit bila ada.

Ukuran PR yang wajar:
- Maksimal sekitar 400 baris perubahan, di luar file hasil generate, mock, dan lockfile.
- Satu PR untuk satu perubahan logis, disertai test.
- Deskripsi PR berisi tujuan perubahan, cara mengetesnya, dan dampaknya ke metrik atau dashboard. Sertakan screenshot untuk perubahan UI.
