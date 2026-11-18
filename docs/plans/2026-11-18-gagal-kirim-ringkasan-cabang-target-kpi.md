# Gagal Kirim, Ringkasan Cabang, dan Target KPI Implementation Plan

> Kerjakan task sesuai urutan dependensinya. Centang langkah (`- [ ]`) setelah selesai. Sebelum eksekusi, jalankan skill `review-plan`.

**Goal:** Tiga fitur: halaman Gagal Kirim, halaman Ringkasan Cabang, dan target KPI bulanan per cabang yang dipakai Ringkasan Cabang dan Performa Kurir (status target per kurir dihitung backend).

**Architecture:** Kontrak empat endpoint (tiga baru, satu tambahan field di `courier-performance`) dibekukan dulu di OpenAPI (Task 0). BE dibangun di atas branch seed `feat/seed-data` (PR #20) karena target diisi lewat seed. FE dibangun di atas redesign gelap `feat/fe-redesign-dark` (belum punya PR) karena memakai donut, strip harian, dan token barunya. Setiap PR sekitar 400 baris atau kurang, ditumpuk; base PR dipindah ke `main` setelah PR di bawahnya di-merge.

**Tech Stack:** sama dengan plan ranking performa kurir. Tidak ada dependency baru di `go.mod` atau `web/package.json`. Lighthouse CLI dipakai lewat `npx` hanya di I2 (alat cek lokal, tidak masuk repo).

**Spec:** `docs/specs/2026-11-18-gagal-kirim-ringkasan-cabang-target-kpi.md`

## Global Constraints
- Batas hari WIB, cut-off SLA 16:00 WIB, dan rumus rate sama dengan spec ranking performa kurir.
- Respons list `{data, meta}`; respons satu objek (`failed-deliveries/summary`, `branch-kpi-targets`) `{data}` tanpa `meta`. Error `invalid_query`, `invalid_filter`, `target_not_found`, `internal_error`. Filter maksimal 31 hari.
- KPI, persen, status target (termasuk `success_rate_status` per kurir), dan prorata dihitung backend. FE hanya memformat dan menghitung proporsi tampilan (busur donut, lebar bar, posisi garis target).
- Bulan target = bulan `end_date`. Pembulatan target volume `math.Round` sekali di akhir. Status dibandingkan dengan nilai mentah.
- Target hanya dibaca. Tidak ada endpoint tulis.
- Tidak ada data pribadi di seed, fixture, log, atau UI.
- Setiap PR: test ikut di PR, `make lint` dan test lolos, sekitar 400 baris atau kurang di luar lockfile dan mock.
- Setiap task yang mengubah `backend/internal/port/port.go` menjalankan `make mocks` (`.mockery.yml` memakai `all: true`, jadi mock baru ikut ter-generate) dan meng-commit `internal/port/mocks/`.
- Integration test memakai fixture Go yang ditulis di test (pola `seedJKT01` di `repository_integration_test.go`), bukan `backend/seeds/*.sql`, karena `openTestPool` men-TRUNCATE tabel dan `seed.sql` butuh `psql` (`\ir`). Seed SQL diverifikasi terpisah di B6a.

## Review Focus
1. Prorata target volume lintas bulan, bulan tanpa target, dan pembulatan .5. Dikunci di B6a (`TestTargetForPeriod`).
2. Periode pembanding Gagal Kirim lintas bulan atau tahun. Dikunci di B7a (`TestPreviousPeriod`).
3. Jumlah alasan sama dengan total gagal, termasuk `failed_reason` NULL atau kosong menjadi "Tanpa alasan". Dikunci di B7a integration.
4. `rank` cabang global lintas halaman. Dikunci di B8a integration.
5. `success_rate_status` per kurir dan `meta.branch_target` di `courier-performance` tanpa mengubah field lama. Dikunci di B9.
6. Filter Performa Kurir baca dan tulis URL, back bekerja, default 7 hari tanpa query. Dikunci di F11.

## Perubahan berisiko dan mitigasinya

| Perubahan | Risiko | Alasan | Mitigasi | Penyetuju |
|---|---|---|---|---|
| Kontrak API: tiga endpoint baru dan tambahan `meta.branch_target` serta `success_rate_status` di `courier-performance` | FE dan BE tidak sinkron; klien lama `courier-performance` rusak | Spec butuh endpoint baru; keputusan user: status per kurir dihitung BE | Task 0 dibekukan dulu dan di-review kedua sisi; hanya menambah field (tidak menghapus atau mengganti nama); test handler B9 memastikan field lama tetap sama | Lead FE (review PR Task 0) |
| Migrasi baru `branch_kpi_targets` | Migrasi gagal di lingkungan lain | Target butuh tabel sendiri | File `down`; langkah naik, turun, naik konkret di B6a; integ test constraint | Lead BE |
| Indeks parsial `delivery_events` | `CREATE INDEX` biasa mengunci tulis dari consumer | Query gagal kirim butuh indeks | `CREATE INDEX CONCURRENTLY IF NOT EXISTS` di file migrasi yang hanya berisi satu pernyataan (golang-migrate menjalankan banyak pernyataan dalam satu transaksi, dan `CONCURRENTLY` tidak boleh di dalam transaksi); down `DROP INDEX CONCURRENTLY IF EXISTS`; bila build gagal, indeks `INVALID` dibuang dengan down lalu ulang | Lead BE |
| Query agregat atas `delivery_events` (sekitar 240 ribu baris seed) per request | Lambat untuk 31 hari semua cabang | Gagal kirim butuh data per paket | Indeks di atas; `EXPLAIN ANALYZE` dicatat di PR B7a | Lead BE |
| Zona waktu dan batas tanggal: periode pembanding, prorata per hari, bulan target | Salah hari di sekitar tengah malam WIB atau pergantian bulan | Aturan bisnis WIB | Semua perhitungan tanggal di domain memakai tanggal sipil WIB (`DayStartWIB`); test 23:59 dan 00:00 WIB (B7a), lintas bulan dan tahun (B7a), bulan `end_date` di batas bulan (B6a, B9) | Lead BE |
| Filter Performa Kurir baca dan tulis query URL | Mengubah perilaku awal halaman di `main`; riwayat browser bertambah tiap ubah filter | Keputusan user: tautan bisa dibagikan, back bekerja | Tanpa query perilaku lama tetap (dikunci test); nilai tidak valid jatuh ke default; test back mengembalikan filter sebelumnya | Lead FE |
| PR ditumpuk di atas branch belum di-merge (`feat/seed-data` PR #20, `feat/fe-redesign-dark` belum ber-PR) | Konflik atau base berubah | Butuh seed dan komponen redesign | Urutan merge: PR #20, lalu PR redesign dibuka dan di-merge, lalu Task 0; setiap PR di atasnya di-rebase ke `main` setelah base di-merge, test diulang setelah rebase | Pemilik PR #20 dan redesign |
| Target tanpa autentikasi | Data target terbaca siapa pun yang bisa mengakses API, sama seperti endpoint lain | Keputusan user: seed saja | Endpoint tulis di luar scope | User (sudah) |

## Rencana test per kriteria selesai

| AC | Test | Task |
|---|---|---|
| G1 | integ `TestFailedSummary_Totals` (fixture: JKT01 dan JKT02, beberapa alasan, kurir lintas cabang), `TestFailedSummary_WIBBoundary` (event 23:59 dan 00:00 WIB), `TestFailedSummary_NullReason` (NULL dan `''` jadi "Tanpa alasan"), `TestFailedSummary_TopCouriers` (maks 5, tie-break `courier_id`); unit handler bentuk respons tanpa `meta` | B7a, B7b |
| G2 | integ `TestFailedSummary_ReasonsSumToTotal`, `TestFailedSummary_DeliveredOnly` (ada paket tanpa gagal: `reasons: []`, rate 0), `TestFailedSummary_Empty` (tanpa paket); unit handler `reasons` berupa `[]` bukan `null` | B7a, B7b |
| G3 | unit `TestPreviousPeriod` (7 hari, lintas bulan, lintas tahun), `TestFailedChangeRate` (naik, turun, nil bila lalu 0); unit service `CountFailed` dipanggil dengan periode pembanding | B7a, B7b |
| G4 | FE `FailedDeliveryPage.test.tsx`: hero, chip naik, turun, null; donut dengan legenda nama, jumlah, persen; daftar kurir; loading, error, kosong | F9 |
| R1 | integ `TestBranchPerformance_RankingAndDaily`, `_SecondPage` (rank global), `_PagePastEndKeepsTotal` | B8a |
| R2 | unit `TestTargetForPeriod` (satu bulan; lintas Okt dan Nov; bulan awal tanpa target; bulan `end` tanpa target nil; pembulatan 700,5 jadi 701), `TestCompareTarget` (sama berarti met; 0,84996 terhadap 0,85 below), `TestEvaluateTarget`; unit service `TestBranchPerformanceService_List` (target null untuk cabang tanpa target, `ListTargets` dipanggil sekali) | B6a, B8b |
| R3 | FE `BranchSummaryPage.test.tsx`: kartu terurut, tiga penanda dengan ikon dan teks, klik kartu menavigasi ke `/courier-performance?branch_id=…&start_date=…&end_date=…` | F10 |
| T1 | integ `TestBranchKPITargets_Constraints` (insert rate 1,1, rate -0,1, volume -1 ditolak; PK ganda ditolak); langkah migrasi naik, turun, naik; cek seed `make seed` dua kali dengan query hitungan | B6a |
| T2 | unit handler `TestKPITargetHandler` (200 dengan `month` `YYYY-MM`, 404 `target_not_found`, 400 `invalid_query` untuk month salah format, month kosong, branch kosong); integ `TestTargetRepository_GetAndList` | B6b |
| T3 | FE `CourierPerformancePage.test.tsx`: garis target keberhasilan dan tepat SLA tampil, penanda per kurir dengan ikon dan teks; `meta.branch_target` null atau tanpa cabang: garis dan penanda tidak tampil | F12 |
| T4 | unit domain `TestCourierPerformance_SuccessRateStatus`; unit handler `TestCourierPerformanceHandler_Target` (dengan target, tanpa target, tanpa `branch_id`, field lama sama dengan golden sebelumnya, bulan target dari `end_date` 2026-11-30 dan 2026-12-01) | B9 |
| U1 | FE `use-courier-performance-filter.test.tsx`: baca query, tulis query saat ubah periode, cabang, halaman; back mengembalikan filter sebelumnya; tanpa query default 7 hari; query tidak valid jatuh ke default | F11 |
| X1 | FE: assertion penanda status punya ikon (`svg` `aria-hidden`) dan teks yang terbaca ("Tercapai", "Di bawah target", "Di atas target") di F9, F10, F12; I2: `scrollWidth <= innerWidth` di 390px untuk tiga halaman, Lighthouse audit `color-contrast` lolos (nol pelanggaran) di tiga halaman, screenshot 1440px dan 390px | F9, F10, F12, I2 |

## Pembagian kerja dan model

| Task | Sisi | Model | Effort | Bergantung ke |
|---|---|---|---|---|
| 0. Kontrak OpenAPI empat endpoint | bersama | sonnet | medium | PR #20 dan redesign di-merge |
| B6a. Migrasi, seed, dan domain target | BE | sonnet | high | 0 |
| B6b. Repo, service, handler target, router `Handlers` | BE | sonnet | medium | B6a |
| B7a. Domain, query, dan indeks gagal kirim | BE | sonnet | high | B6b |
| B7b. Service dan handler gagal kirim | BE | sonnet | medium | B7a |
| B8a. Query ranking cabang | BE | sonnet | high | B7b |
| B8b. Service dan handler ranking cabang dengan target | BE | sonnet | medium | B8a |
| B9. Target cabang dan status per kurir di `courier-performance` | BE | sonnet | high | B8b |
| F8. Tipe, client, hook, dan helper render dengan router | FE | sonnet | medium | 0 |
| F9. Halaman Gagal Kirim dan `DonutChart` generik | FE | sonnet | medium | F8 |
| F10. Halaman Ringkasan Cabang dan `DailyStrip` ke `components/` | FE | sonnet | medium | F9 |
| F11. Filter Performa Kurir baca dan tulis URL | FE | sonnet | medium | F8 |
| F12. Garis target dan penanda per kurir di Performa Kurir | FE | sonnet | medium | F10, F11 |
| I2. Integrasi lokal, cek 390px, dan kontras | bersama | haiku | low | semua |

BE berurutan B6a, B6b, B7a, B7b, B8a, B8b, B9 di satu branch tumpuk, karena semuanya menyentuh `port.go`, `router.go`, `cmd/api/main.go`, dan nomor migrasi. FE: F8 dulu; F9 lalu F10 berurutan (keduanya memindah komponen dari `features/courier-performance/` dan menyentuh `CourierDetailPanel.tsx` serta `ResultsList.tsx`); F11 paralel dengan F9 dan F10 (hanya `use-courier-performance-filter.ts` dan `CourierPerformancePage.tsx`); F12 setelah F10 dan F11. FE bisa mulai paralel dengan BE setelah Task 0 karena memakai MSW.

---

### Task 0: Kontrak OpenAPI
**Files:** Modify `backend/api/openapi.yaml`.
- [ ] Tambah path `failed-deliveries/summary`, `branch-performance`, `branch-kpi-targets` dan schema `FailedSummary`, `FailedReason`, `FailedCourier`, `BranchPerformance`, `BranchTarget` (objek `target`), `KPITarget` persis spec, termasuk 400, 404, 500. Respons summary dan target `{data}` tanpa `meta`; `month` `pattern: '^\d{4}-\d{2}$'`.
- [ ] Ubah `courier-performance`: tambah `success_rate_status` (`enum [met, below]`, `nullable`) di item dan `branch_target` (`nullable`, `{month, success_rate, sla_rate}`) di `meta`. Field lama tidak diubah.
- [ ] Validasi YAML: `python3 -c 'import sys,yaml; yaml.safe_load(open(sys.argv[1]))' backend/api/openapi.yaml` tanpa error. Minta review lead FE di PR.
- [ ] Commit `docs(contracts): kontrak gagal kirim, ringkasan cabang, target KPI, dan target di performa kurir`.

### Task B6a: Migrasi, seed, dan domain target
**Files:** Create `backend/migrations/000002_branch_kpi_targets.up.sql`, `000002_branch_kpi_targets.down.sql`, `backend/seeds/04_branch_kpi_targets.sql`, `backend/internal/domain/target.go`, `target_test.go`, `backend/internal/adapter/postgres/target_integration_test.go`. Modify `backend/seeds/seed.sql` (tambah `\ir 04_branch_kpi_targets.sql` sebelum `COMMIT`), `backend/internal/domain/errors.go` (`ErrTargetNotFound`).
**Interfaces (`internal/domain/target.go`):**
- `type KPITarget struct { BranchID string; Month time.Time; SuccessRate, SLARate float64; DeliveredPackages int }` (`Month` = tanggal sipil hari pertama bulan, lokasi UTC seperti `DailyCount.Date`).
- `func MonthOf(t time.Time) time.Time` (hari pertama bulan dari tanggal sipil WIB `t`).
- `type PeriodTarget struct { SuccessRate, SLARate float64; DeliveredPackages int }`.
- `func TargetForPeriod(targets []KPITarget, start, end time.Time) *PeriodTarget`. Prasyarat: `targets` satu cabang. Nil bila bulan `end` tanpa target. Volume = `int(math.Round(Σ hari (target bulan hari itu / jumlah hari bulan itu)))`.
- `type TargetStatus string` dengan `TargetMet = "met"`, `TargetBelow = "below"`; `func CompareTarget(actual, target float64) TargetStatus` (`actual >= target` berarti met, nilai mentah).
- `var ErrTargetNotFound = errors.New("target tidak ditemukan")`.
- [ ] Test dulu: `TestTargetForPeriod` dan `TestCompareTarget` sesuai tabel test (R2), `TestMonthOf` (2026-11-30 dan 2026-12-01).
- [ ] Migrasi: tabel sesuai spec, `CHECK (success_rate BETWEEN 0 AND 1)`, sama untuk `sla_rate`, `CHECK (delivered_packages >= 0)`, `CHECK (month = date_trunc('month', month)::date)`, PK `(branch_id, month)`. Down: `DROP TABLE IF EXISTS branch_kpi_targets;`.
- [ ] Integ `TestBranchKPITargets_Constraints` (build tag `integration`, pakai `openTestPool`, TRUNCATE `branch_kpi_targets` di awal test).
- [ ] Seed `04`: `DELETE FROM branch_kpi_targets WHERE branch_id IN (SELECT branch_id FROM seed_branch) AND month BETWEEN '2026-01-01' AND '2026-12-01'`, lalu insert deterministik 12 bulan untuk `seed_branch` kecuali BDG01, rumus tanpa random; sebagian cabang di atas, sebagian di bawah aktual.
- [ ] Verifikasi migrasi (dari `backend/`, DB uji kosong): `make migrate-up`; `psql "$DATABASE_URL" -c '\d branch_kpi_targets'` menampilkan PK dan tiga CHECK; `make migrate-down` (turun 1) lalu `\d branch_kpi_targets` menampilkan "Did not find any relation"; `make migrate-up` lagi berhasil.
- [ ] Verifikasi seed: `make seed` dua kali; setelah tiap kali jalankan `SELECT branch_id, count(*), md5(string_agg(month::text||success_rate||sla_rate||delivered_packages, ',' ORDER BY month)) FROM branch_kpi_targets GROUP BY branch_id ORDER BY 1;`. Hasil: 12 baris per cabang seed selain BDG01, BDG01 tidak muncul, hash kedua jalan identik. Tempel hasil di deskripsi PR.
- [ ] `make test-race`, `make test-integration`, `make lint` tanpa temuan baru.
- [ ] Commit `feat(backend): tabel dan aturan target KPI bulanan per cabang`.

### Task B6b: Repo, service, handler target, router `Handlers`
**Files:** Modify `backend/internal/port/port.go`, `internal/port/mocks/` (generate), `backend/internal/adapter/httpapi/router.go`, `response.go`, `backend/cmd/api/main.go`, `backend/internal/adapter/postgres/repository_integration_test.go` (`openTestPool` ikut TRUNCATE `branch_kpi_targets`). Create `backend/internal/adapter/postgres/target_repository.go`, `backend/internal/service/target_service.go`, `target_service_test.go`, `backend/internal/adapter/httpapi/kpi_target_handler.go`, `kpi_target_handler_test.go`; tambah `TestTargetRepository_GetAndList` di `target_integration_test.go`.
**Interfaces:**
- Port `TargetRepository { ListTargets(ctx context.Context, branchIDs []string, from, to time.Time) ([]domain.KPITarget, error); GetTarget(ctx context.Context, branchID string, month time.Time) (domain.KPITarget, bool, error) }` (`from`, `to` = bulan, inklusif).
- `postgres.NewTargetRepository(pool *pgxpool.Pool) *TargetRepository`.
- `service.NewTargetService(repo port.TargetRepository) *TargetService` dengan `GetTarget(ctx, branchID string, month time.Time) (domain.KPITarget, error)` (`domain.ErrTargetNotFound` bila tidak ada) dan `RateTarget(ctx, branchID string, end time.Time) (*domain.KPITarget, error)` (target bulan `MonthOf(end)`, nil bila tidak ada).
- `httpapi.NewKPITargetHandler(svc KPITargetGetter, logger *slog.Logger) *KPITargetHandler`, `KPITargetGetter { GetTarget(ctx, branchID string, month time.Time) (domain.KPITarget, error) }`. Respons `writeData(w, body, nil)` sehingga `{data}` tanpa `meta`; `month` diformat `2006-01`.
- Router: `type Handlers struct { CourierMetrics *CourierMetricHandler; CourierPerformance *CourierPerformanceHandler; KPITargets *KPITargetHandler }` dan `func NewRouter(h Handlers, logger *slog.Logger) http.Handler`. B7b dan B8b menambah field. Route `GET /api/v1/branch-kpi-targets`.
- `writeDomainError` memetakan `domain.ErrTargetNotFound` ke 404 `target_not_found`.
- [ ] `make mocks`.
- [ ] Test handler `TestKPITargetHandler` sesuai tabel (T2); test service dengan mock; integ repo dengan fixture Go dua cabang dan dua bulan.
- [ ] Update `main.go` ke `NewRouter(httpapi.Handlers{...}, logger)`; test router lama tetap lolos.
- [ ] Commit `feat(backend): endpoint target KPI per cabang`.

### Task B7a: Domain, query, dan indeks gagal kirim
**Files:** Create `backend/internal/domain/failed.go`, `failed_test.go`, `backend/internal/adapter/postgres/failed_repository.go`, `failed_integration_test.go`, `backend/migrations/000003_failed_events_index.up.sql`, `000003_failed_events_index.down.sql`. Modify `backend/internal/domain/metric.go` (ekstrak `validatePeriod`), `backend/internal/port/port.go`, `internal/port/mocks/` (generate).
**Interfaces (`internal/domain/failed.go`):**
- `const NoReasonLabel = "Tanpa alasan"`.
- `type FailedFilter struct { StartDate, EndDate time.Time; BranchID string }` dengan `Validate() error` yang memanggil `validatePeriod(start, end time.Time) error` (diekstrak dari `MetricFilter.Validate`, yang juga memanggilnya; test `MetricFilter` lama harus tetap lolos). Dipilih validasi periode terpisah, bukan `WithDefaults` lalu `Validate`, supaya endpoint tanpa pagination tidak membawa field page.
- `func PreviousPeriod(start, end time.Time) (time.Time, time.Time)`.
- `type FailedReason struct { Reason string; Count int }`.
- `type FailedCourier struct { Rank int; CourierID, CourierName, BranchID string; TotalPackages, Failed int }` dengan `FailedRate() float64` (0 bila total 0).
- `type FailedSummary struct { TotalPackages, TotalFailed, PreviousTotalFailed int; Reasons []FailedReason; TopCouriers []FailedCourier }` dengan `FailedRate() float64`, `FailedChangeRate() *float64`, `Share(r FailedReason) float64` (0 bila `TotalFailed` 0).
- Port `FailedRepository { FailedSummary(ctx context.Context, f domain.FailedFilter) (domain.FailedSummary, error); CountFailed(ctx context.Context, f domain.FailedFilter) (int, error) }`. `FailedSummary` tidak mengisi `PreviousTotalFailed`. Reasons memakai `COALESCE(NULLIF(btrim(failed_reason), ''), 'Tanpa alasan')`. `Reasons` dan `TopCouriers` tidak pernah nil.
- `postgres.NewFailedRepository(pool *pgxpool.Pool) *FailedRepository`.
- [ ] Test dulu: `TestPreviousPeriod`, `TestFailedChangeRate`, `TestFailedFilter_Validate`, lalu integ G1 dan G2 sesuai tabel dengan fixture Go `seedFailedFixture(t)` (event sintetis JKT01 dan JKT02, termasuk NULL dan `''`, satu kurir pindah cabang, cabang `DLV01` hanya delivered).
- [ ] Migrasi up berisi tepat satu pernyataan: `CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_delivery_events_failed_branch_occurred ON delivery_events (branch_id, occurred_at) WHERE status = 'failed'`. Down satu pernyataan: `DROP INDEX CONCURRENTLY IF EXISTS idx_delivery_events_failed_branch_occurred`. Tidak ada komentar yang memuat titik koma.
- [ ] Verifikasi: `make migrate-up`, `\d delivery_events` menampilkan indeks valid; `make migrate-down` menghapusnya; `make migrate-up` lagi. `EXPLAIN ANALYZE` query summary 31 hari semua cabang di DB ber-seed, tempel di PR.
- [ ] `make mocks`, `make test-race`, `make test-integration`, `make lint`.
- [ ] Commit `feat(backend): query ringkasan gagal kirim`.

### Task B7b: Service dan handler gagal kirim
**Files:** Create `backend/internal/service/failed_delivery_service.go`, `failed_delivery_service_test.go`, `backend/internal/adapter/httpapi/failed_delivery_handler.go`, `failed_delivery_handler_test.go`. Modify `router.go` (field `FailedDeliveries *FailedDeliveryHandler`, route `GET /api/v1/failed-deliveries/summary`), `cmd/api/main.go`.
**Interfaces:**
- `service.NewFailedDeliveryService(repo port.FailedRepository) *FailedDeliveryService` dengan `Summary(ctx context.Context, f domain.FailedFilter) (domain.FailedSummary, error)`: `f.Validate()`, `repo.FailedSummary`, lalu `repo.CountFailed` untuk `PreviousPeriod` dengan cabang yang sama.
- `httpapi.NewFailedDeliveryHandler(svc FailedSummarizer, logger *slog.Logger) *FailedDeliveryHandler`, `FailedSummarizer { Summary(ctx, f domain.FailedFilter) (domain.FailedSummary, error) }`. Parsing tanggal memakai `parseDate`; respons `writeData(w, body, nil)` (`successBody.Meta` sudah `omitempty`, jadi keluar `{data}`).
- [ ] Test service (mock: periode pembanding benar, error dibungkus) dan handler (bentuk, `reasons: []`, `failed_change_rate: null`, tanpa `meta`, 400 `invalid_query`, 400 `invalid_filter`).
- [ ] Commit `feat(backend): endpoint ringkasan gagal kirim`.

### Task B8a: Query ranking cabang
**Files:** Create `backend/internal/domain/branch_performance.go`, `branch_performance_test.go`, `backend/internal/adapter/postgres/branch_performance_integration_test.go`. Modify `backend/internal/adapter/postgres/metric_repository.go`, `backend/internal/port/port.go`, `internal/port/mocks/` (generate).
**Interfaces:**
- `type BranchPerformance struct { Rank int; BranchID string; CourierCount, TotalPackages, Delivered, Failed, OnTime int; Daily []DailyCount }` dengan `SuccessRate()`, `SLARate()` (rumus sama dengan `CourierMetric`).
- `type BranchPerformancePage struct { Items []BranchPerformance; Total int }`.
- Port `MetricRepository` tambah `ListBranchPerformance(ctx context.Context, filter domain.MetricFilter) (domain.BranchPerformancePage, error)`: satu query ranking dengan `COUNT(*) OVER ()` dan rank global, satu query harian untuk cabang di halaman, lalu `FillDaily`. `BranchID` di filter diabaikan.
- [ ] Integ R1 sesuai tabel dengan fixture Go tiga cabang.
- [ ] `make mocks`, test, lint. Commit `feat(backend): query ranking performa cabang`.

### Task B8b: Service dan handler ranking cabang dengan target
**Files:** Create `backend/internal/service/branch_performance_service.go`, `branch_performance_service_test.go`, `backend/internal/adapter/httpapi/branch_performance_handler.go`, `branch_performance_handler_test.go`. Modify `backend/internal/domain/target.go` dan `target_test.go`, `router.go` (field `BranchPerformance *BranchPerformanceHandler`, route `GET /api/v1/branch-performance`), `cmd/api/main.go`.
**Interfaces:**
- `type BranchTargetResult struct { Target PeriodTarget; SuccessRateStatus, SLARateStatus, DeliveredStatus TargetStatus }`.
- `func EvaluateTarget(p BranchPerformance, t *PeriodTarget) *BranchTargetResult` (nil bila `t` nil; volume membandingkan `Delivered` dengan `t.DeliveredPackages`).
- `type BranchPerformanceWithTarget struct { BranchPerformance; Target *BranchTargetResult }` dan `type BranchPerformanceResultPage struct { Items []BranchPerformanceWithTarget; Total int }`.
- `service.NewBranchPerformanceService(metrics port.MetricRepository, targets port.TargetRepository) *BranchPerformanceService` dengan `List(ctx, f domain.MetricFilter) (domain.BranchPerformanceResultPage, error)`: `WithDefaults`, `Validate`, ranking, satu kali `ListTargets` untuk cabang di halaman dari `MonthOf(start)` sampai `MonthOf(end)`, kelompokkan per cabang, `TargetForPeriod`, `EvaluateTarget`.
- `httpapi.NewBranchPerformanceHandler(svc BranchPerformanceLister, logger *slog.Logger) *BranchPerformanceHandler`.
- [ ] Test `TestEvaluateTarget`, `TestBranchPerformanceService_List`, handler (bentuk, `target: null`, `meta`, 400).
- [ ] Commit `feat(backend): endpoint ranking performa cabang dengan target`.

### Task B9: Target cabang dan status per kurir di `courier-performance`
**Files:** Modify `backend/internal/domain/performance.go`, `performance_test.go`, `backend/internal/adapter/httpapi/courier_performance_handler.go`, `courier_performance_handler_test.go`, `cmd/api/main.go`. `response.go` tidak diubah: meta khusus didefinisikan di handler, `type courierPerformanceMeta struct { pageMeta; BranchTarget *branchTargetMeta \`json:"branch_target"\` }`, supaya endpoint lain tidak ikut mengirim `branch_target`.
**Interfaces:**
- `func (p CourierPerformance) SuccessRateStatus(t *KPITarget) *TargetStatus` (nil bila `t` nil; memakai `CompareTarget`).
- `type RateTargetFinder interface { RateTarget(ctx context.Context, branchID string, end time.Time) (*domain.KPITarget, error) }` di `httpapi`, diimplementasikan `service.TargetService`.
- `httpapi.NewCourierPerformanceHandler(svc CourierPerformanceLister, targets RateTargetFinder, logger *slog.Logger)`. Setelah `ListCourierPerformance` sukses dan `filter.BranchID != ""`, handler memanggil `RateTarget(filter.BranchID, filter.EndDate)`; item mendapat `success_rate_status`, `meta.branch_target` diisi `{month "YYYY-MM", success_rate, sla_rate}` atau `null`.
- [ ] Test T4 sesuai tabel; golden respons lama: semua field lama bernilai sama seperti sebelum B9.
- [ ] Commit `feat(backend): target cabang dan status per kurir di performa kurir`.

### Task F8: Tipe, client, hook, dan helper render dengan router
**Files:** Modify `web/src/api/types.ts`, `web/src/api/courier-performance.ts`, `courier-performance.test.ts`, `web/src/test/fixtures.ts`, `web/src/test/render.tsx`. Create `web/src/api/failed-deliveries.ts`, `failed-deliveries.test.ts`, `web/src/api/branch-performance.ts`, `branch-performance.test.ts`.
**Interfaces:**
- Tipe DTO snake_case dan tipe camelCase: `FailedSummary`, `FailedReason`, `FailedCourier`, `BranchPerformance`, `BranchTarget`, `TargetStatus = 'met' | 'below'`, `CourierPerformance.successRateStatus: TargetStatus | null`, `PageMeta.branchTarget?: { month: string; successRate: number; slaRate: number } | null`.
- `useFailedSummary(filter: { startDate: string; endDate: string; branchID: string })`, `useBranchPerformance(filter: { startDate: string; endDate: string; page: number; pageSize: number })`. `branch-kpi-targets` tidak diberi hook karena tidak ada konsumen FE di scope ini (target Performa Kurir datang dari `meta.branch_target`).
- `renderWithProviders(ui, { route?: string })` di `test/render.tsx`: QueryClient tanpa retry dan `MemoryRouter` dengan `initialEntries`. `renderWithQuery` lama tetap.
- [ ] Test pemetaan DTO ke tipe untuk tiga client, termasuk `failed_change_rate: null` dan `branch_target: null`. Fixture sintetis dari contoh spec.
- [ ] Commit `feat(web): client dan hook gagal kirim, ringkasan cabang, dan target`.

### Task F9: Halaman Gagal Kirim dan `DonutChart` generik
**Files:** Move `web/src/features/courier-performance/DonutChart.tsx` dan `DonutChart.test.tsx` ke `web/src/components/` lalu generalisasi. Modify `web/src/features/courier-performance/CourierDetailPanel.tsx` (memetakan `compositionShares` ke `segments`), `web/src/pages/failed-delivery/FailedDeliveryPage.tsx`. Create `FailedDeliveryPage.test.tsx`, `web/src/features/failed-delivery/FailedHero.tsx`, `FailedReasonLegend.tsx`, `TopFailedCouriers.tsx`, `failed-delivery.css`, `web/src/components/StatusBadge.tsx`, `StatusBadge.test.tsx`.
**Interfaces:**
- `DonutChart({ total, segments, label })` dengan `segments: { key: string; label: string; value: number; share: number; pattern?: 'solid' | 'diagonal' | 'dots' }[]`. Tidak meng-import dari `features/`. Test donut lama diadaptasi, tampilan panel detail kurir tidak berubah.
- `StatusBadge({ status: 'met' | 'below' | 'above', label: string })`: ikon `svg` `aria-hidden` plus teks. Dipakai F10 dan F12.
- [ ] Hero dengan chip naik (panah atas, "Naik x%"), turun (panah bawah, "Turun x%"), null ("Tidak ada pembanding"); legenda donut berisi alasan, jumlah, persen; daftar 5 kurir; loading, error, kosong.
- [ ] Commit `feat(web): halaman gagal kirim`.

### Task F10: Halaman Ringkasan Cabang dan `DailyStrip` ke `components/`
**Files:** Move `web/src/features/courier-performance/DailyStrip.tsx` dan `DailyStrip.test.tsx` ke `web/src/components/`; update import di `Podium.tsx`, `ResultsList.tsx`, `ResultsTable.tsx`, `CourierDetailPanel.tsx`, `BarChart.tsx` bila meng-import. Modify `web/src/pages/branch-summary/BranchSummaryPage.tsx`. Create `BranchSummaryPage.test.tsx`, `web/src/features/branch-summary/BranchCard.tsx`, `BranchCard.test.tsx`, `branch-summary.css`.
- [ ] Filter periode (`DateRangeFilter`), kartu terurut berisi rank, cabang, keberhasilan, tepat SLA, paket terkirim, jumlah kurir, `DailyStrip`, tiga `StatusBadge`, `Pagination`; loading, error, kosong. Kartu adalah `Link` ke `/courier-performance?branch_id=…&start_date=…&end_date=…`.
- [ ] Test R3 dan X1 sesuai tabel.
- [ ] Commit `feat(web): halaman ringkasan cabang`.

### Task F11: Filter Performa Kurir baca dan tulis URL
**Files:** Modify `web/src/features/courier-performance/use-courier-performance-filter.ts`, `web/src/pages/courier-performance/CourierPerformancePage.tsx`, `CourierPerformancePage.test.tsx`, `CourierPerformancePage.mobile.test.tsx` (render lewat `renderWithProviders`). Create `use-courier-performance-filter.test.tsx`.
- [ ] Hook memakai `useSearchParams` sebagai sumber kebenaran: baca `branch_id`, `start_date`, `end_date`, `page`; nilai tidak valid (format tanggal salah, rentang lebih dari 31 hari, `page` < 1) jatuh ke default per parameter. Setter menulis query lewat `setSearchParams` (push, entri riwayat baru); ubah periode atau cabang mengembalikan `page` ke 1. Tanpa query tidak menulis apa pun saat mount, default 7 hari.
- [ ] Test U1 sesuai tabel.
- [ ] Commit `feat(web): filter performa kurir dari dan ke URL`.

### Task F12: Garis target dan penanda per kurir di Performa Kurir
**Files:** Modify `web/src/components/DailyStrip.tsx`, `DailyStrip.test.tsx`, `web/src/features/courier-performance/ResultsList.tsx`, `ResultsTable.tsx`, `results.css`, `web/src/pages/courier-performance/CourierPerformancePage.tsx`, `CourierPerformancePage.test.tsx`.
- [ ] `DailyStrip` menerima `targets?: { successRate: number; slaRate: number }` dan menggambar dua garis horizontal: keberhasilan garis utuh, tepat SLA garis putus, masing-masing dengan label teks ("Target keberhasilan 90%", "Target tepat SLA 85%"). Hanya diberikan bila `meta.branchTarget` ada.
- [ ] Setiap baris kurir menampilkan `StatusBadge` dari `successRateStatus` ("Di atas target" atau "Di bawah target"); tidak tampil bila null.
- [ ] Test T3 dan X1 sesuai tabel.
- [ ] Commit `feat(web): target cabang di performa kurir`.

### Task I2: Integrasi lokal, cek 390px, dan kontras
**Files:** tidak ada perubahan kode; hasil dicatat di deskripsi PR terakhir.
- [ ] `docker compose up --build` (migrate dan seed otomatis). Buka `/failed-delivery`, `/branch-summary`, `/courier-performance?branch_id=JKT01&start_date=2026-11-01&end_date=2026-11-07`.
- [ ] Di DevTools, lebar 390px dan 1440px: `document.documentElement.scrollWidth <= window.innerWidth` bernilai `true` di tiga halaman; simpan screenshot.
- [ ] Kontras: `npx lighthouse@<versi stabil terbaru, ditulis eksplisit> <url> --only-audits=color-contrast --output=json --chrome-flags=--headless` untuk tiga URL; skor audit `color-contrast` 1 (nol pelanggaran). Catat versi dan hasil di PR.
- [ ] Klik kartu cabang membuka Performa Kurir dengan filter yang sama; tombol back kembali ke Ringkasan Cabang.
