# Ranking Performa Kurir Implementation Plan

> Kerjakan task satu per satu sesuai urutan dependensinya. Centang langkah (`- [ ]`) setelah selesai.

**Goal:** Endpoint `GET /api/v1/courier-performance` (ranking, total, dan tren harian per kurir) di backend, dan dashboard `/courier-performance` bergaya "Hasil Lomba Lari" di frontend, di atas boilerplate yang sudah ada di `main`.

**Architecture:** Kontrak dibekukan dulu di `backend/api/openapi.yaml` bersama perubahan konvensi zona waktu di `CLAUDE.md` (Task 0). Setelah itu BE dan FE jalan paralel: BE menambah aturan WIB, endpoint ranking, dan seed data contoh; FE menguji terhadap fixture yang angkanya sama persis dengan seed. Integrasi lewat `docker compose`.

**Tech Stack:** Mengikuti boilerplate. BE: Go, `net/http`, pgx, `testing` + `testify` + mock `mockery` (`make mocks`). FE: React, TypeScript, Vite, TanStack Query, Vitest, Testing Library, MSW, ditambah `@fontsource/barlow` dan `@fontsource/barlow-condensed`. Grafik digambar sendiri dengan SVG.

**Spec:** `docs/specs/2026-11-18-courier-performance.md`
**Desain:** `docs/design/2026-11-18-courier-performance.md` (wajib dibaca sebelum F1, F3 sampai F5)

## Global Constraints
- Ikuti konvensi di `CLAUDE.md`: respons `{data, meta}`, error `{error:{code,message}}`, parameter `start_date`/`end_date`, field snake_case, KPI dihitung backend.
- Penyimpanan UTC; batas hari dan batas SLA dihitung dalam WIB (`time.FixedZone("WIB", 7*3600)`).
- Cut-off SLA: ditugaskan sebelum 16:00 WIB berarti batas hari itu 23:59:59 WIB; pukul 16:00 atau lebih berarti besok 16:00:00 WIB. `due_at` dari sumber tetap diutamakan.
- Ranking: `success_rate` menurun, lalu `courier_id` menaik; `rank` global; `success_rate_gap` dari backend.
- `daily`: satu entri per tanggal WIB dalam periode, hari kosong bernilai 0, jumlah sama dengan total.
- Format FE: `formatPercent` yang sudah ada (`90,0%`), selisih `−4,3 pp` (minus U+2212), rank 1 `Pemimpin`.
- Tidak me-log atau menampilkan nomor HP, NIK, atau alamat. Data seed sintetis.
- BE hanya menyentuh `backend/` (plus `docker-compose.yml` untuk seed), FE hanya `web/`. Kontrak hanya diubah lewat PR yang direview kedua sisi.

## Review Focus
1. Batas hari WIB vs UTC di semua tempat: worker, repository agregasi, service, dan zero-fill `daily`. Satu tempat yang masih UTC membuat event 00:00 sampai 06:59 WIB masuk hari sebelumnya. Dikunci di B1 (`TestDayStartWIB`, `TestAggregateDaily_WIBBoundary` integration) dan B2 (`TestFillDaily`).
2. `rank` dan `success_rate_gap` di halaman 2: kalau dihitung per halaman, rank 51 jadi 1 dan gap jadi 0. Dikunci di B2 (`TestListCourierPerformance_SecondPage` integration) dan F3 (test halaman 2 tanpa podium).
3. Hari kosong di `daily`: kalau hanya hari yang ada datanya dikirim, strip FE bergeser. Dikunci di B2 (`TestFillDaily`) dan F4 (`DailyStrip` hari kosong tanpa NaN).
4. Event tanpa `due_at` dan tanpa `assigned_at`: tidak pernah tepat SLA (perilaku boilerplate) dan tidak boleh panic. Dikunci di B1 (`TestDeliveryEvent_WithDefaultDueAt`).
5. Layar 390px dengan 31 hari dan 40 kurir: tidak boleh scroll horizontal halaman. Data uji `JKT02` dari seed (B3), diperiksa lewat screenshot di F5.

## Pembagian kerja dan model

| Task | Sisi | Pemilik | Model | Effort | Bergantung ke |
|---|---|---|---|---|---|
| 0. Kontrak OpenAPI dan konvensi zona waktu | bersama | BE + FE | sonnet | medium | - |
| B1. Aturan WIB: hari laporan dan cut-off SLA | BE | tim BE | sonnet | high | 0 |
| B2. Endpoint ranking performa kurir | BE | tim BE | sonnet | high | B1 |
| B3. Seed data contoh | BE | tim BE | sonnet | medium | B2 |
| F1. Token desain, font, shell, `formatGap` | FE | tim FE | sonnet | medium | 0 |
| F2. Tipe, client, dan hook `useCourierPerformance` | FE | tim FE | sonnet | medium | F1 |
| F3. Halaman: podium, tabel hasil, state | FE | tim FE | sonnet | medium | F2 |
| F4. Strip harian dan panel detail kurir | FE | tim FE | sonnet | high | F3 |
| F5. Gerak "garis finis", responsif, QA desain | FE | tim FE | sonnet | high | F4 |
| I1. Integrasi lokal | bersama | BE + FE | haiku | low | B3, F5 |

Saran PR: Task 0 sendiri; B1; B2+B3; F1; F2+F3; F4+F5. Tiap PR di bawah 400 baris di luar mock hasil generate dan lockfile.

---

### Task 0: Kontrak OpenAPI dan konvensi zona waktu

**Files:**
- Modify: `backend/api/openapi.yaml`
- Modify: `CLAUDE.md` (bagian Backend, Konvensi API)

**Interfaces:**
- Produces: path `/api/v1/courier-performance` dan schema `CourierPerformance`, `DailyCount` yang menjadi acuan B2 dan F2.

- [ ] **Step 1: Tambah path** `GET /api/v1/courier-performance` dengan parameter dan respons persis seperti bagian Kontrak API di spec (200 `{data: CourierPerformance[], meta: PageMeta}`, 400 dan 500 memakai schema `Error`).
- [ ] **Step 2: Tambah schema** `CourierPerformance` (semua field wajib: `rank`, `courier_id`, `courier_name`, `branch_id`, `total_packages`, `delivered`, `failed`, `on_time`, `success_rate`, `sla_rate`, `success_rate_gap`, `daily`) dan `DailyCount` (`date`, `total_packages`, `delivered`, `failed`, `on_time`), dengan `description` untuk rumus dan urutan. Ubah deskripsi `start_date` di `/courier-metrics` dari "(UTC)" menjadi "(tanggal WIB)".
- [ ] **Step 3: Ubah `CLAUDE.md`**: ganti kalimat "Semua waktu disimpan dalam UTC; konversi zona waktu dilakukan di frontend." menjadi: "Semua waktu disimpan dalam UTC. Batas hari laporan dan batas SLA adalah aturan bisnis yang dihitung backend dalam WIB (UTC+7); `start_date` dan `end_date` adalah tanggal WIB."
- [ ] **Step 4: Validasi YAML**

Run: `python3 -c "import yaml,sys; yaml.safe_load(open('backend/api/openapi.yaml')); print('ok')"`
Expected: `ok`.

- [ ] **Step 5: Commit, PR, minta review dari satu orang BE dan satu orang FE**

```bash
git add backend/api/openapi.yaml CLAUDE.md
git commit -m "docs(contracts): kontrak ranking performa kurir dan aturan zona waktu WIB"
```

---

### Task B1: Aturan WIB: hari laporan dan cut-off SLA

**Files:**
- Create: `backend/internal/domain/timezone.go`
- Test: `backend/internal/domain/timezone_test.go`
- Modify: `backend/internal/domain/delivery.go`, `backend/internal/domain/delivery_test.go`
- Modify: `backend/internal/adapter/kafka/delivery_consumer.go`, `delivery_consumer_test.go`
- Modify: `backend/internal/adapter/postgres/metric_repository.go`, `repository_integration_test.go`
- Modify: `backend/internal/service/metric_service.go`, `metric_service_test.go`
- Modify: `backend/cmd/worker/main.go`

**Interfaces:**
- Produces:
  - `var WIB = time.FixedZone("WIB", 7*60*60)`
  - `func DayStartWIB(t time.Time) time.Time`: 00:00 WIB pada tanggal WIB dari `t`.
  - `func SLADeadline(assignedAt time.Time) time.Time`: aturan cut-off di spec, hasil dalam `WIB`.
  - Field baru `AssignedAt time.Time` di `DeliveryEvent`, dan `func (e DeliveryEvent) WithDefaultDueAt() DeliveryEvent`: bila `DueAt` nol dan `AssignedAt` tidak nol, isi `DueAt = SLADeadline(AssignedAt).UTC()`; selain itu kembalikan apa adanya.
  - Pesan Kafka menerima field opsional `assigned_at`.
  - `AggregateDaily` (service dan repository) menerima tanggal dan mengagregasi `[DayStartWIB, +1 hari)`, kolom `date` diisi tanggal WIB.

- [ ] **Step 1: Tulis `TestSLADeadline`** (table-driven, bandingkan dengan `Equal`): 7 Nov 09:00 WIB menghasilkan 7 Nov 23:59:59 WIB; 15:59:59 menghasilkan 7 Nov 23:59:59; 16:00:00 menghasilkan 8 Nov 16:00:00; 21:00 menghasilkan 8 Nov 16:00:00; 30 Nov 17:00 menghasilkan 1 Des 16:00:00; `2026-11-07T09:00:00Z` (= 16:00 WIB) menghasilkan 8 Nov 16:00:00 WIB.
- [ ] **Step 2: Tulis `TestDayStartWIB`**: `2026-11-17T16:59:00Z` menghasilkan 17 Nov 00:00 WIB; `2026-11-17T17:00:00Z` menghasilkan 18 Nov 00:00 WIB.
- [ ] **Step 3: Tulis `TestDeliveryEvent_WithDefaultDueAt`**: `AssignedAt` 7 Nov 09:00 WIB tanpa `DueAt` mengisi `DueAt` 7 Nov 23:59:59 WIB; `DueAt` dari sumber tidak diubah; tanpa keduanya `DueAt` tetap nol dan `IsOnTime()` false. Tambah kasus AC6: ditugaskan 7 Nov 09:00, `OccurredAt` 8 Nov 00:00 WIB tidak tepat SLA; ditugaskan 7 Nov 16:00, `OccurredAt` 8 Nov 16:00 WIB tepat SLA.
- [ ] **Step 4: Ubah test consumer**: pesan dengan `assigned_at` tanpa `due_at` menghasilkan event dengan `DueAt` hasil cut-off.
- [ ] **Step 5: Ubah test service** `AggregateDaily`: input `2026-11-17T17:30:00Z` meneruskan 18 Nov 00:00 WIB ke repository (sebelumnya 17 Nov UTC).
- [ ] **Step 6: Tambah `TestAggregateDaily_WIBBoundary`** di integration test: event `2026-11-17T16:59:00Z` dan `2026-11-17T17:00:00Z`; agregasi 17 Nov hanya menghitung yang pertama, agregasi 18 Nov hanya yang kedua.
- [ ] **Step 7: Jalankan, pastikan gagal**

Run: `cd backend && go test ./internal/... && make test-integration`
Expected: FAIL pada test baru (`undefined: SLADeadline`, batas UTC).

- [ ] **Step 8: Implementasi.** Repository memakai `start := DayStartWIB(date)`, `end := start.AddDate(0, 0, 1)`, kolom `date` diisi `start.Format(time.DateOnly)`. Service memakai `DayStartWIB`, bukan `Truncate(24*time.Hour)`. Worker menghitung "hari ini" dan "kemarin" dari `time.Now().In(domain.WIB)`. Consumer memanggil `WithDefaultDueAt()` di `toDomain`. Perbarui komentar yang menyebut UTC.
- [ ] **Step 9: Jalankan semua test, coverage, lint**

Run: `cd backend && make test-race && make test-integration && make cover && make lint`
Expected: PASS, coverage `internal/domain` dan `internal/service` minimal 80%, total minimal 60%, lint bersih.

- [ ] **Step 10: Commit dan buka PR B1**

```bash
git add backend/
git commit -m "feat(backend): batas hari dan cut-off SLA 16:00 dalam WIB"
```

---

### Task B2: Endpoint ranking performa kurir

**Files:**
- Create: `backend/internal/domain/performance.go`, `backend/internal/domain/performance_test.go`
- Modify: `backend/internal/port/port.go` (lalu `make mocks`)
- Modify: `backend/internal/adapter/postgres/metric_repository.go`, `repository_integration_test.go`
- Modify: `backend/internal/service/metric_service.go`, `metric_service_test.go`
- Create: `backend/internal/adapter/httpapi/courier_performance_handler.go`, `courier_performance_handler_test.go`
- Modify: `backend/internal/adapter/httpapi/router.go`, `backend/cmd/api/main.go`

**Interfaces:**
- Consumes: `MetricFilter` (validasi, pagination, batas 31 hari) dan `parseMetricFilter` yang sudah ada.
- Produces:
  - `type DailyCount struct { Date time.Time; TotalPackages, Delivered, Failed, OnTime int }` (`Date` = tanggal sipil, dibandingkan lewat `Format(time.DateOnly)`).
  - `type CourierPerformance struct { Rank int; CourierID, CourierName, BranchID string; TotalPackages, Delivered, Failed, OnTime int; LeaderSuccessRate float64; Daily []DailyCount }` dengan method `SuccessRate()`, `SLARate()` (rumus sama dengan `CourierMetric`), dan `SuccessRateGap()` = `LeaderSuccessRate - SuccessRate()`, minimal 0.
  - `type CourierPerformancePage struct { Items []CourierPerformance; Total int }`
  - `func FillDaily(counts []DailyCount, start, end time.Time) []DailyCount`: satu entri per tanggal dari `start` sampai `end` inklusif, urut menaik, hari tanpa data bernilai 0.
  - `MetricRepository.ListCourierPerformance(ctx context.Context, filter domain.MetricFilter) (domain.CourierPerformancePage, error)`
  - `MetricService.ListCourierPerformance(ctx context.Context, filter domain.MetricFilter) (domain.CourierPerformancePage, error)`: `WithDefaults`, `Validate`, lalu repository. Tanpa cache untuk endpoint ini dulu.
  - Route `GET /api/v1/courier-performance`.

- [ ] **Step 1: Tulis test domain**:
  - `TestFillDaily`: data hanya 2 dan 5 Nov, periode 1 sampai 7 Nov, menghasilkan 7 entri berurutan dengan 1, 3, 4, 6, 7 Nov bernilai 0; periode 1 hari menghasilkan 1 entri; `counts` nil menghasilkan entri nol, bukan nil.
  - `TestCourierPerformance_Rates`: 40/36/4/33 dengan leader 0.9 menghasilkan `SuccessRate` 0.9, `SLARate` 33/36, gap 0; 4/0/4/0 dengan leader 0.9 menghasilkan rate 0, gap 0.9; total 0 menghasilkan rate 0 tanpa NaN.
- [ ] **Step 2: Tulis test service** dengan mock repository: filter tidak valid (periode 32 hari) mengembalikan `ErrInvalidFilter` tanpa memanggil repository; error repository dibungkus dengan konteks.
- [ ] **Step 3: Tulis test handler** dengan `httptest` dan service palsu: 200 dengan bentuk persis contoh di spec (termasuk `rank`, `success_rate_gap`, `daily[].date` berformat `YYYY-MM-DD`); data kosong menghasilkan `"data":[]`; `start_date=abc` menghasilkan 400 `invalid_query`; `ErrInvalidFilter` menghasilkan 400 `invalid_filter`; error lain menghasilkan 500 `internal_error` tanpa pesan asli.
- [ ] **Step 4: Tulis integration test repository**: masukkan `courier_daily_metrics` JKT01 sesuai tabel Data contoh di spec, lalu:
  - `TestListCourierPerformance_JKT01`: 4 item urut C001, C003, C004, C002; total dan `daily` sama dengan tabel; `meta.total` 4.
  - `TestListCourierPerformance_SecondPage`: `page_size` 2, `page` 2 menghasilkan C004 rank 3 dan C002 rank 4, `LeaderSuccessRate` 0.9.
  - `TestListCourierPerformance_BranchFilter`: `branch_id` BDG01 menghasilkan 0 item.
- [ ] **Step 5: Jalankan, pastikan gagal**

Run: `cd backend && go test ./internal/... && make test-integration`
Expected: FAIL, simbol belum ada.

- [ ] **Step 6: Implementasi.**
  - Query 1 (halaman): agregasi `courier_daily_metrics` per `courier_id` dalam periode dan cabang, hitung kunci urut dengan rumus yang sama dengan domain, `ROW_NUMBER()` untuk `rank`, `MAX(...) OVER ()` untuk rate pemimpin, `COUNT(*) OVER ()` untuk total, lalu `LIMIT/OFFSET`. `branch_id` per kurir = cabang pada tanggal terakhirnya di periode.
  - Query 2 (harian): baris harian untuk `courier_id = ANY($ids)` dalam periode, lalu `FillDaily` di Go. Dua query per request, tanpa N+1.
  - Handler memetakan ke DTO snake_case, slice kosong di-encode `[]`, error lewat `writeDomainError`.
  - Jalankan `make mocks` setelah interface `port` berubah.
- [ ] **Step 7: Jalankan semua test, coverage, lint**

Run: `cd backend && make test-race && make test-integration && make cover && make lint`
Expected: PASS, coverage sesuai standar, lint bersih.

- [ ] **Step 8: Commit**

```bash
git add backend/
git commit -m "feat(backend): endpoint ranking performa kurir dengan tren harian"
```

---

### Task B3: Seed data contoh

**Files:**
- Create: `backend/seeds/demo.sql`
- Modify: `backend/Makefile` (target `seed`)
- Modify: `docker-compose.yml` (service `seed`)

**Interfaces:**
- Produces: data sintetis di `couriers` dan `courier_daily_metrics`:
  - `JKT01`, 1 sampai 7 Nov 2026: persis tabel Data contoh di spec.
  - `JKT02`, 1 sampai 31 Okt 2026: 40 kurir (`K201` sampai `K240`, nama "Kurir 01" sampai "Kurir 40"), angka harian deterministik dari `generate_series` (misalnya `total = (kurir * 7 + hari * 3) % 9`, `delivered` dan `on_time` turunan yang konsisten: `on_time <= delivered <= total`, `failed = total - delivered`).
  - `BDG01`: tanpa data.

- [ ] **Step 1: Tulis `demo.sql`** yang idempoten: hapus dulu baris seed untuk kurir `C001` sampai `C004` dan `K201` sampai `K240`, lalu insert. Data pribadi tidak ada.
- [ ] **Step 2: Tambah target** `seed` di Makefile (psql lewat image `postgres:16-alpine` dengan `--network host` dan `DATABASE_URL`) dan service `seed` di compose (image `postgres:16-alpine`, `depends_on: migrate: service_completed_successfully`, menjalankan `psql -f /seeds/demo.sql` lalu selesai).
- [ ] **Step 3: Verifikasi**

Run: `docker compose up -d postgres migrate seed backend` lalu `curl -s 'localhost:8080/api/v1/courier-performance?start_date=2026-11-01&end_date=2026-11-07&branch_id=JKT01&page_size=50'`
Expected: 4 item dengan angka sama dengan tabel Data contoh; untuk `JKT02` periode 1 sampai 31 Okt, `meta.total` 40 dan setiap item punya 31 entri `daily`.

- [ ] **Step 4: Commit dan buka PR B2+B3**

```bash
git add backend/seeds backend/Makefile docker-compose.yml
git commit -m "chore(backend): seed data contoh untuk dashboard performa kurir"
```

---

### Task F1: Token desain, font, shell, dan `formatGap`

**Files:**
- Modify: `web/package.json` (tambah `@fontsource/barlow`, `@fontsource/barlow-condensed`; cek versi stabil terbaru di npm saat memasang)
- Modify: `web/src/app/app.css` (token desain), `web/src/app/main.tsx` (import font), `web/src/app/Layout.tsx` (shell)
- Create: `web/src/lib/format-gap.ts`
- Test: `web/src/lib/format-gap.test.ts`

**Interfaces:**
- Produces:
  - Token CSS (nama baru, tanpa warna hardcode di komponen): `--color-shell`, `--color-shell-ink`, `--color-bg`, `--color-surface`, `--color-ink`, `--color-ink-muted`, `--color-rule`, `--color-lane`, `--color-lane-soft`, `--color-gold`, `--color-silver`, `--color-bronze`, `--color-error`, `--font-ui`, `--font-figure`, `--text-xs` sampai `--text-display`, `--space-1` sampai `--space-8`. Token terang di `:root`, token gelap di `@media (prefers-color-scheme: dark)` dengan peran yang sama. Variabel lama (`--bg`, `--accent`, dan lainnya) diganti di semua pemakaiannya.
  - `export function formatGap(gap: number): string`

- [ ] **Step 1: Tulis test `formatGap`** dengan `it.each`: `0.04285714` menjadi `"−4,3 pp"`, `0.9` menjadi `"−90,0 pp"`, `0` menjadi `"0,0 pp"`.
- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `cd web && npm test -- format-gap`
Expected: FAIL, modul tidak ditemukan.

- [ ] **Step 3: Implementasi.** `formatGap` memakai `Intl.NumberFormat("id-ID", { minimumFractionDigits: 1, maximumFractionDigits: 1 })` atas `gap * 100`, diberi awalan `−` bila gap lebih dari 0. Shell: top bar `--color-shell` dengan teks `--color-shell-ink`, nama produk memakai `--font-figure`. Nilai token warna lolos kontras AA terhadap `--color-surface` dan `--color-shell` di kedua tema.
- [ ] **Step 4: Jalankan semua pemeriksaan**

Run: `cd web && npm run test:coverage && npm run lint && npm run typecheck`
Expected: semua lolos, termasuk test yang sudah ada.

- [ ] **Step 5: Commit dan buka PR F1** (sertakan screenshot shell terang dan gelap)

```bash
git add web/
git commit -m "feat(web): token desain, font Barlow, dan shell dashboard"
```

---

### Task F2: Tipe, client, dan hook `useCourierPerformance`

**Files:**
- Modify: `web/src/api/types.ts`
- Create: `web/src/api/courier-performance.ts`
- Test: `web/src/api/courier-performance.test.ts`
- Modify: `web/src/test/fixtures.ts`

**Interfaces:**
- Consumes: `apiGet`, `ApiError` (`client.ts`), `CourierMetricFilter`, `PageMetaDTO`, `Paginated` (`types.ts`).
- Produces:
  - DTO: `DailyCountDTO { date; total_packages; delivered; failed; on_time }`, `CourierPerformanceDTO` (field persis schema OpenAPI).
  - Tipe aplikasi (camelCase): `DailyCount { date; totalPackages; delivered; failed; onTime }`, `CourierPerformance { rank; courierID; courierName; branchID; totalPackages; delivered; failed; onTime; successRate; slaRate; successRateGap; daily: DailyCount[] }`.
  - `export function toCourierPerformance(dto: CourierPerformanceDTO): CourierPerformance`
  - `export async function fetchCourierPerformance(filter: CourierMetricFilter, signal?: AbortSignal): Promise<Paginated<CourierPerformance>>`
  - `export const courierPerformanceKey = (filter: CourierMetricFilter) => ['courier-performance', filter] as const`
  - `export function useCourierPerformance(filter: CourierMetricFilter)` dengan `placeholderData: keepPreviousData`.
  - `fixtures.ts`: `courierPerformanceDTOs: CourierPerformanceDTO[]` berisi 4 kurir JKT01 persis tabel Data contoh di spec, dengan `success_rate`, `sla_rate`, dan `success_rate_gap` hasil rumus (misalnya Dewi `success_rate` 30/35, gap `0.9 - 30/35`).

- [ ] **Step 1: Tulis test** mengikuti pola `courier-metrics.test.ts`: parameter query (`start_date`, `end_date`, `branch_id`, `page`, `page_size`) terkirim benar; respons dipetakan ke camelCase termasuk `daily`; error 400 `invalid_filter` dilempar sebagai `ApiError` dengan `code` dan `message` dari backend.
- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `cd web && npm test -- courier-performance`
Expected: FAIL, modul tidak ditemukan.

- [ ] **Step 3: Implementasi** sesuai Interfaces.
- [ ] **Step 4: Jalankan, pastikan lolos**

Run: `cd web && npm run test:coverage`
Expected: PASS, coverage `src/api` minimal 80%.

- [ ] **Step 5: Commit**

```bash
git add web/
git commit -m "feat(web): client dan hook ranking performa kurir"
```

---

### Task F3: Halaman: podium, tabel hasil, dan state

**Files:**
- Modify: `web/src/pages/courier-performance/CourierPerformancePage.tsx`, `CourierPerformancePage.test.tsx`
- Modify: `web/src/features/courier-performance/use-courier-performance-filter.ts` (`PAGE_SIZE` 50)
- Modify: `web/src/components/QueryState.tsx` (prop opsional `loadingFallback?: ReactNode`, tetap di dalam `role="status"`)
- Create: `web/src/features/courier-performance/Podium.tsx`, `ResultsTable.tsx`, `courier-performance.css`
- Delete: `web/src/features/courier-performance/CourierPerformanceTable.tsx` (tabel per hari tidak dipakai lagi; `useCourierMetrics` tetap ada di `api/`)
- Modify: `web/src/app/router.test.tsx` (mock endpoint baru)

**Interfaces:**
- Consumes: `useCourierPerformance`, `CourierPerformance` (F2), `formatPercent`, `formatNumber`, `formatGap` (F1), `DateRangeFilter`, `Pagination`, `QueryState`.
- Produces:
  - `Podium` props `{ items: CourierPerformance[]; onSelect: (courierID: string) => void }`: `<ol aria-label="Podium">` berisi item dengan `rank` 1 sampai 3, tiap kartu sebuah `<button>` berisi nomor posisi, nama, keberhasilan, dan tepat SLA. Emas, perak, perunggu hanya di sini.
  - `ResultsTable` props `{ items: CourierPerformance[]; onSelect: (courierID: string) => void }`: `<table>` dengan caption "Hasil lengkap", kolom persis Pos, Kurir, Harian, Total paket, Terkirim, Gagal, Keberhasilan, Tepat SLA, Selisih. Kolom Harian diisi F4. Nama kurir adalah `<button>`. Tiap baris punya `data-finish-key={courierID}` untuk F5. Rank 1 menampilkan `Pemimpin` di kolom Selisih.
  - `CourierPerformancePage` menyimpan `selectedCourierID: string | null` di state.

- [ ] **Step 1: Tulis ulang test halaman** (pola test yang sudah ada, `today` 18 Nov 2026):
  - loading: `role="status"` berisi "Memuat data".
  - error dan Coba lagi: perilaku lama tetap jalan.
  - podium dan tabel (respons `courierPerformanceDTOs`): `Podium` berisi Budi Santoso, Dewi Lestari, Agus Pratama berurutan; tabel 4 baris; baris Budi memuat `90,0%`, `91,7%`, `Pemimpin`; baris Dewi `85,7%`, `100,0%`, `−4,3 pp`; baris Sari `0,0%`, `−90,0 pp`.
  - halaman 2 (respons `meta.page` 2, satu item rank 51): podium tidak dirender, baris memuat posisi `51`.
  - kosong: "Belum ada data performa kurir untuk periode ini."
  - filter: mengubah Cabang mengirim request dengan `branch_id` baru dan `page` 1.
  - netral: baris Sari tidak punya class atau atribut bernada negatif.
- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `cd web && npm test -- CourierPerformancePage router`
Expected: FAIL.

- [ ] **Step 3: Implementasi.** Layout sesuai bagian Viewport pertama di dokumen desain. Skeleton berbentuk podium dan baris dikirim lewat `loadingFallback`. Podium hanya bila `meta.page === 1`.
- [ ] **Step 4: Jalankan semua pemeriksaan**

Run: `cd web && npm run test:coverage && npm run lint && npm run typecheck`
Expected: semua lolos, coverage total minimal 60%.

- [ ] **Step 5: Commit dan buka PR F2+F3** (sertakan screenshot desktop dengan data, kosong, dan error; beri keterangan "data contoh")

```bash
git add web/
git commit -m "feat(web): podium dan tabel hasil ranking performa kurir"
```

---

### Task F4: Strip harian dan panel detail kurir

**Files:**
- Create: `web/src/features/courier-performance/DailyStrip.tsx`, `DailyStrip.test.tsx`
- Create: `web/src/features/courier-performance/LaneChart.tsx`
- Create: `web/src/features/courier-performance/CourierDetailPanel.tsx`, `CourierDetailPanel.test.tsx`
- Modify: `Podium.tsx`, `ResultsTable.tsx`, `CourierPerformancePage.tsx`, `courier-performance.css`

**Interfaces:**
- Consumes: `DailyCount`, `CourierPerformance` (F2), `formatPercent`, `formatNumber`, `formatDate` (`lib/date.ts`), `selectedCourierID` (F3).
- Produces:
  - `DailyStrip` props `{ daily: DailyCount[]; size: 'sm' | 'md' }`: `null` bila `daily.length <= 1`. Akar `role="img"` dengan `aria-label` "Tren harian {n} hari: terkirim {Σdelivered} dari {ΣtotalPackages}". Satu sel per hari dengan `data-day={date}`, grid `repeat(n, minmax(0, 1fr))`. Tinggi isi = `delivered / totalPackages` (0 bila `totalPackages` 0); bagian tepat SLA `data-part="on-time"`, bagian telat `data-part="late"` (pola garis). Hari tanpa paket `data-empty="true"`.
  - `LaneChart` props `{ daily: DailyCount[] }`: SVG `role="img"` dengan `aria-label`; seri terkirim (garis penuh) dan tepat SLA (garis putus), sumbu Y mulai 0.
  - `CourierDetailPanel` props `{ item: CourierPerformance | null; onClose: () => void }`: `role="dialog"`, `aria-modal="true"`, `aria-labelledby` ke judul nama kurir; isi posisi, angka total, `LaneChart`, dan tabel harian (Tanggal, Total paket, Terkirim, Gagal, Tepat SLA) sebagai alternatif teks. Tutup lewat tombol "Tutup" dan Esc; fokus kembali ke pemicu; Tab terkunci di dalam panel.

- [ ] **Step 1: Tulis test `DailyStrip`** (data Budi dari fixture): 7 elemen `[data-day]`; `aria-label` "Tren harian 7 hari: terkirim 36 dari 40"; sel `2026-11-03` tanpa bagian `late` (5 terkirim, 5 tepat SLA); sel `2026-11-04` punya bagian `late` (5 terkirim, 4 tepat SLA). Data Sari: sel `2026-11-03` `data-empty="true"` dan tidak ada `NaN` di atribut `style`. Satu entri: tidak merender apa pun.
- [ ] **Step 2: Tulis test `CourierDetailPanel`** lewat halaman: klik "Dewi Lestari" di tabel membuka dialog bernama "Dewi Lestari" berisi tabel harian 7 baris dan grafik `role="img"`; Esc menutup dan fokus kembali ke tombol "Dewi Lestari"; klik kartu podium "Budi Santoso" membuka dialog Budi.
- [ ] **Step 3: Jalankan, pastikan gagal**

Run: `cd web && npm test -- DailyStrip CourierDetailPanel`
Expected: FAIL, modul tidak ditemukan.

- [ ] **Step 4: Implementasi.** Strip `md` di kartu podium, `sm` di kolom Harian. Desktop: panel dari kanan di atas backdrop.
- [ ] **Step 5: Jalankan semua pemeriksaan**

Run: `cd web && npm run test:coverage && npm run lint && npm run typecheck`
Expected: semua lolos.

- [ ] **Step 6: Commit**

```bash
git add web/
git commit -m "feat(web): strip harian dan panel detail kurir"
```

---

### Task F5: Gerak "garis finis", responsif, dan QA desain

**Files:**
- Create: `web/src/lib/use-media-query.ts`
- Create: `web/src/features/courier-performance/use-finish-line.ts`, `use-finish-line.test.tsx`
- Create: `web/src/features/courier-performance/ResultsList.tsx`, `FilterSheet.tsx`
- Create: `web/src/pages/courier-performance/CourierPerformancePage.mobile.test.tsx`
- Modify: `ResultsTable.tsx`, `Podium.tsx`, `CourierDetailPanel.tsx`, `CourierPerformancePage.tsx`, `courier-performance.css`

**Interfaces:**
- Produces:
  - `export function useMediaQuery(query: string): boolean` (`false` bila `matchMedia` tidak ada).
  - `export function useFinishLine(container: RefObject<HTMLElement | null>, keys: string[]): void`: FLIP; saat urutan `keys` berubah, elemen `[data-finish-key]` dianimasikan dari posisi lama ke baru dengan `element.animate` (transform, 450 sampai 600 ms, easing keluar). Tidak memanggil `animate` bila `prefers-reduced-motion: reduce`.
  - Lebar `(max-width: 640px)`: `FilterSheet` (tombol ringkasan, misalnya "JKT01 · 1 sampai 7 Nov", membuka sheet berisi filter), podium carousel `scroll-snap-type: x mandatory`, `ResultsList` (`<ol aria-label="Hasil lengkap">` berisi posisi, nama, keberhasilan, strip `sm`) menggantikan `ResultsTable`, dan panel detail jadi sheet dari bawah.

- [ ] **Step 1: Tulis test `useFinishLine`**: `keys` `["C001","C003"]` berubah ke `["C003","C001"]` dengan `animate` dan `getBoundingClientRect` di-mock: `animate` terpanggil untuk kedua elemen; dengan reduced motion tidak terpanggil.
- [ ] **Step 2: Tulis test mobile** dengan `matchMedia` sempit: tombol ringkasan filter tampil dan membuka sheet berisi input Cabang; data JKT01 tampil sebagai `<ol aria-label="Hasil lengkap">` 4 item tanpa `<table>`.
- [ ] **Step 3: Jalankan, pastikan gagal**

Run: `cd web && npm test -- use-finish-line mobile`
Expected: FAIL.

- [ ] **Step 4: Implementasi.** Pasang `useFinishLine` di tabel, daftar, dan podium dengan `keys` = urutan `courierID` yang tampil. Kolom Selisih memudar masuk setelah baris berhenti. Tabel di layar menengah dibungkus kontainer `overflow-x: auto` sendiri; halaman tidak scroll horizontal.
- [ ] **Step 5: Jalankan semua pemeriksaan**

Run: `cd web && npm run test:coverage && npm run lint && npm run typecheck`
Expected: semua lolos.

- [ ] **Step 6: QA desain, satu putaran plus satu konfirmasi.** `docker compose up --build`, buka `/courier-performance`, ambil screenshot ke `.impeccable/review/`: `desktop.png` (1440px, JKT01 1 sampai 7 Nov), `mobile.png` (390px, JKT01), `dense-desktop.png` dan `dense-mobile.png` (JKT02 1 sampai 31 Okt), plus `desktop-dark.png`. Jalankan `~/.claude/skills/impeccable/scripts/impeccable detect --json web/src`. Perbaiki semua temuan dalam satu batch, ambil ulang screenshot sekali. Kriteria: tanpa scroll horizontal di 390px, podium terbaca tanpa zoom, strip 31 hari terbaca, kontras AA di kedua tema, reduced motion mematikan gerak.
- [ ] **Step 7: Review akhir desain** oleh reviewer terpisah (impeccable finish reviewer) dengan dokumen desain sebagai kontrak, lalu tulis `DESIGN.md` dari hasil build.
- [ ] **Step 8: Commit dan buka PR F4+F5** (sertakan screenshot, keterangan "data contoh")

```bash
git add web/ DESIGN.md
git commit -m "feat(web): gerak garis finis dan tampilan responsif performa kurir"
```

---

### Task I1: Integrasi lokal

**Files:** tidak ada perubahan kode. Ketidakcocokan diperbaiki lewat PR terpisah ke sisi yang salah; kontrak hanya diubah atas persetujuan kedua sisi.

- [ ] **Step 1:** `cp .env.example .env` (isi `GO_VERSION` dan `NODE_VERSION`), lalu `docker compose up --build`.
- [ ] **Step 2:** Buka `http://localhost:5173/courier-performance`.
- [ ] **Step 3: Cek empat skenario**:
  - JKT01 1 sampai 7 Nov: podium Budi, Dewi, Agus; angka sama dengan tabel Data contoh.
  - JKT02 1 sampai 31 Okt: 40 kurir, strip 31 sel, tanpa scroll horizontal di 390px.
  - BDG01: pesan data kosong.
  - Periode lebih dari 31 hari: pesan error `invalid_filter` dari backend.
- [ ] **Step 4:** Catat hasil di deskripsi PR terakhir yang di-merge, dengan screenshot keempat skenario.
