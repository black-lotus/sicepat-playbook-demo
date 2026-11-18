# Gagal Kirim, Ringkasan Cabang, dan Target KPI Implementation Plan

> Kerjakan task sesuai urutan dependensinya. Centang langkah (`- [ ]`) setelah selesai. Sebelum eksekusi, jalankan skill `review-plan`.

**Goal:** Tiga fitur: halaman Gagal Kirim, halaman Ringkasan Cabang, dan target KPI bulanan per cabang yang dipakai Ringkasan Cabang dan Performa Kurir.

**Architecture:** Kontrak tiga endpoint dibekukan dulu di OpenAPI (Task 0). BE dibangun di atas branch seed (`feat/seed-data`) karena target diisi lewat seed. FE dibangun di atas redesign gelap (`feat/fe-redesign-dark`) karena memakai komponen donut, kartu glass, dan token barunya. Setiap PR sekitar 400 baris atau kurang, ditumpuk; base PR dipindah ke `main` setelah PR di bawahnya di-merge.

**Tech Stack:** sama dengan plan ranking performa kurir. Tidak ada dependency baru.

**Spec:** `docs/specs/2026-11-18-gagal-kirim-ringkasan-cabang-target-kpi.md`

## Global Constraints
- Batas hari WIB, cut-off SLA 16:00 WIB, dan rumus rate sama dengan spec ranking performa kurir.
- Respons `{data, meta}`; error `invalid_query`, `invalid_filter`, `target_not_found`, `internal_error`. Filter maksimal 31 hari.
- KPI, persen, status target, dan prorata dihitung backend. FE hanya memformat dan menghitung proporsi tampilan (busur donut, lebar bar).
- Target hanya dibaca. Tidak ada endpoint tulis.
- Tidak ada data pribadi di seed, log, atau UI.
- Setiap PR: test ikut di PR, lint dan test lolos, sekitar 400 baris atau kurang di luar lockfile dan mock.

## Review Focus
1. Prorata target volume lintas bulan dan bulan tanpa target. Dikunci di B6a (`TestTargetForPeriod`).
2. Periode pembanding Gagal Kirim untuk periode yang melewati batas bulan atau tahun. Dikunci di B7a (`TestPreviousPeriod`).
3. Jumlah alasan sama dengan total gagal, termasuk alasan kosong dari data lama. Dikunci di B7a integration.
4. `rank` cabang global lintas halaman. Dikunci di B8a integration.
5. Halaman Performa Kurir yang dibuka dari Ringkasan Cabang memakai filter dari URL, bukan default 7 hari. Dikunci di F10.

## Perubahan berisiko dan mitigasinya

| Perubahan | Risiko | Alasan | Mitigasi |
|---|---|---|---|
| Migrasi baru `branch_kpi_targets` dan indeks event gagal | Migrasi gagal di lingkungan lain | Target butuh tabel sendiri; query gagal kirim butuh indeks | File `down` untuk rollback, integration test menjalankan migrasi bersih |
| Query agregat atas `delivery_events` (sekitar 240 ribu baris seed) per request | Lambat untuk rentang 31 hari semua cabang | Gagal kirim butuh data per paket | Indeks parsial `(branch_id, occurred_at) WHERE status = 'failed'`; cek `EXPLAIN` di B7a |
| Target tanpa autentikasi | Tidak ada, karena hanya dibaca | Keputusan user: seed saja | Endpoint tulis di luar scope |
| Filter Performa Kurir dibaca dari URL | Mengubah perilaku awal halaman yang sudah ada | Navigasi dari Ringkasan Cabang | Tanpa query URL perilaku lama tetap (default 7 hari), dikunci test |

## Rencana test per kriteria selesai

| AC | Test | Task |
|---|---|---|
| G1 | integ `TestFailedSummary_JKT01Nov` (angka dari seed), unit handler bentuk respons | B7a, B7b |
| G2 | integ `TestFailedSummary_ReasonsSumToTotal`, `TestFailedSummary_NoFailures` | B7a |
| G3 | unit `TestPreviousPeriod` (lintas bulan dan tahun), `TestFailedChangeRate` (null bila lalu 0) | B7a |
| G4 | FE `GagalKirimPage`: hero, donut, daftar kurir, loading, error, kosong | F9 |
| R1 | integ `TestBranchPerformance_RankingAndDaily`, `TestBranchPerformance_SecondPage` | B8a |
| R2 | unit `TestTargetForPeriod` (bulan end_date, prorata lintas bulan, bulan tanpa target, status), integ `TestBranchPerformance_TargetNullForBDG01` | B6a, B8b |
| R3 | FE `BranchSummaryPage`: kartu, penanda target berbentuk dan bertekst, klik membuka URL Performa Kurir dengan filter | F10 |
| T1 | integ migrasi bersih naik dan turun, cek constraint; verifikasi seed target | B6a |
| T2 | unit handler 200, 404 `target_not_found`, 400 `invalid_query` | B6b |
| T3 | FE penanda per kurir dan garis target; tanpa target tidak tampil | F11 |
| X1 | Cek browser 390px `scrollWidth <= innerWidth` untuk tiga halaman, plus screenshot | I2 |

## Pembagian kerja dan model

| Task | Sisi | Model | Effort | Bergantung ke |
|---|---|---|---|---|
| 0. Kontrak OpenAPI tiga endpoint | bersama | sonnet | medium | - |
| B6a. Migrasi, seed, dan domain target | BE | sonnet | high | 0, seed |
| B6b. Repo, service, handler target | BE | sonnet | medium | B6a |
| B7a. Domain dan query gagal kirim | BE | sonnet | high | 0, seed |
| B7b. Service dan handler gagal kirim | BE | sonnet | medium | B7a |
| B8a. Query ranking cabang | BE | sonnet | high | 0 |
| B8b. Service dan handler ranking cabang dengan target | BE | sonnet | medium | B6a, B8a |
| F8. Tipe, client, dan hook tiga endpoint | FE | sonnet | medium | 0, redesign |
| F9. Halaman Gagal Kirim | FE | sonnet | medium | F8 |
| F10. Halaman Ringkasan Cabang dan filter dari URL | FE | sonnet | medium | F8 |
| F11. Target di halaman Performa Kurir | FE | sonnet | medium | F8 |
| I2. Integrasi lokal dan cek 390px | bersama | haiku | low | semua |

B6, B7, B8 menyentuh berkas port, router, dan main yang sama, jadi dikerjakan berurutan atau di worktree terpisah lalu digabung. F9, F10, F11 menyentuh halaman berbeda dan bisa paralel, kecuali F10 dan F11 yang sama-sama mengubah halaman Performa Kurir: F11 setelah F10.

---

### Task 0: Kontrak OpenAPI
**Files:** Modify `backend/api/openapi.yaml`.
- [ ] Tambah tiga path dan schema persis bagian Kontrak API di spec (`FailedSummary`, `FailedReason`, `FailedCourier`, `BranchPerformance`, `BranchTarget`, `KPITarget`), termasuk 400, 404, dan 500.
- [ ] Validasi YAML, lalu commit `docs(contracts): kontrak gagal kirim, ringkasan cabang, dan target KPI`.

### Task B6a: Migrasi, seed, dan domain target
**Files:** Create `backend/migrations/000002_branch_kpi_targets.up.sql` dan `.down.sql`, `backend/seeds/04_branch_kpi_targets.sql` (dipanggil dari `seed.sql`), `backend/internal/domain/target.go` dan `target_test.go`.
**Interfaces:**
- `type KPITarget struct { BranchID string; Month time.Time; SuccessRate, SLARate float64; DeliveredPackages int }`
- `type PeriodTarget struct { SuccessRate, SLARate float64; DeliveredPackages int }` dan `func TargetForPeriod(targets []KPITarget, start, end time.Time) *PeriodTarget` (nil bila bulan `end` tanpa target; prorata sesuai spec).
- `type TargetStatus string` (`met`, `below`) dan `func CompareTarget(actual, target float64) TargetStatus`.
- [ ] Test dulu: `TestTargetForPeriod` (periode dalam satu bulan; lintas Okt dan Nov dengan target berbeda; bulan awal tanpa target; bulan `end` tanpa target menghasilkan nil; pembulatan), `TestCompareTarget` (sama dengan target berarti met).
- [ ] Migrasi dengan constraint `CHECK (success_rate BETWEEN 0 AND 1)`, sama untuk `sla_rate`, `delivered_packages >= 0`, PK `(branch_id, month)`.
- [ ] Seed target deterministik untuk 12 bulan 2026, semua cabang seed kecuali BDG01; nilai target dibuat sebagian tercapai dan sebagian tidak.
- [ ] Verifikasi: migrasi naik dan turun di DB uji bersih; seed dua kali identik; `make test-race`, lint tanpa temuan baru.
- [ ] Commit `feat(backend): tabel dan aturan target KPI bulanan per cabang`.

### Task B6b: Repo, service, handler target
**Files:** port, `adapter/postgres/target_repository.go` dan integ test, service, `adapter/httpapi/kpi_target_handler.go` dan test, router, `cmd/api/main.go`.
**Interfaces:** `TargetRepository.ListTargets(ctx, branchIDs []string, from, to time.Time) ([]domain.KPITarget, error)` dan `GetTarget(ctx, branchID string, month time.Time) (domain.KPITarget, bool, error)`; route `GET /api/v1/branch-kpi-targets`.
- [ ] Test handler 200, 404 `target_not_found`, 400 `invalid_query` (month bukan `YYYY-MM`, branch kosong); integ repo membaca seed.
- [ ] Commit `feat(backend): endpoint target KPI per cabang`.

### Task B7a: Domain dan query gagal kirim
**Files:** `domain/failed.go` dan test, `adapter/postgres/failed_repository.go` dan integ test, `backend/migrations/000003_failed_events_index.up.sql` dan `.down.sql`.
**Interfaces:** `func PreviousPeriod(start, end time.Time) (time.Time, time.Time)`; `type FailedSummary struct { TotalPackages, TotalFailed, PreviousTotalFailed int; Reasons []FailedReason; TopCouriers []FailedCourier }` dengan method `FailedRate()` dan `FailedChangeRate() *float64`; `FailedRepository.FailedSummary(ctx, branchID string, start, end time.Time) (domain.FailedSummary, error)`.
- [ ] Test dulu: `TestPreviousPeriod` (7 hari, lintas bulan, lintas tahun), `TestFailedChangeRate`, integ `TestFailedSummary_JKT01Nov`, `_ReasonsSumToTotal`, `_NoFailures` (BDG01), batas WIB event 23:59 dan 00:00.
- [ ] Indeks parsial `(branch_id, occurred_at) WHERE status = 'failed'`; catat hasil `EXPLAIN` di deskripsi PR.
- [ ] Commit `feat(backend): query ringkasan gagal kirim`.

### Task B7b: Service dan handler gagal kirim
- [ ] Service memvalidasi filter (tanpa pagination) dan memanggil repo; handler `GET /api/v1/failed-deliveries/summary` sesuai kontrak; test service dan handler (bentuk, `reasons: []`, `failed_change_rate: null`).
- [ ] Commit `feat(backend): endpoint ringkasan gagal kirim`.

### Task B8a: Query ranking cabang
**Interfaces:** `type BranchPerformance struct { Rank int; BranchID string; CourierCount, TotalPackages, Delivered, Failed, OnTime int; Daily []DailyCount }` dengan `SuccessRate()`, `SLARate()`; `MetricRepository.ListBranchPerformance(ctx, filter MetricFilter) (BranchPerformancePage, error)` (dua query, tanpa N+1, `FillDaily`).
- [ ] Integ `TestBranchPerformance_RankingAndDaily`, `_SecondPage`, `_PagePastEndKeepsTotal`.
- [ ] Commit `feat(backend): query ranking performa cabang`.

### Task B8b: Service dan handler ranking cabang dengan target
- [ ] Service mengambil target untuk cabang di halaman (`ListTargets` sekali), menghitung `TargetForPeriod` dan status per metrik; handler `GET /api/v1/branch-performance`; test termasuk `target: null` untuk BDG01.
- [ ] Commit `feat(backend): endpoint ranking performa cabang dengan target`.

### Task F8: Tipe, client, dan hook
- [ ] DTO dan tipe camelCase untuk tiga endpoint, `useFailedSummary`, `useBranchPerformance`, `useBranchTarget` (`enabled` hanya bila cabang terpilih; 404 menjadi `null`, bukan error); fixture dari angka seed; test pemetaan dan 404.
- [ ] Commit `feat(web): client dan hook gagal kirim, ringkasan cabang, dan target`.

### Task F9: Halaman Gagal Kirim
- [ ] Ganti placeholder `FailedDeliveryPage`: filter periode dan cabang, hero total gagal dengan chip perubahan (ikon panah dan teks), donut alasan (pakai komponen donut redesign), daftar 5 kurir; state loading, error, kosong; test halaman.
- [ ] Commit `feat(web): halaman gagal kirim`.

### Task F10: Halaman Ringkasan Cabang dan filter dari URL
- [ ] Ganti placeholder `BranchSummaryPage`: filter periode, kartu cabang terurut dengan strip harian dan tiga penanda target (ikon dan teks), pagination; klik kartu ke `/courier-performance?branch_id=&start_date=&end_date=`.
- [ ] `use-courier-performance-filter` membaca nilai awal dari query URL dan menulis balik saat filter berubah; tanpa query, default 7 hari tetap.
- [ ] Test halaman, test navigasi, test filter dari URL dan tanpa URL.
- [ ] Commit `feat(web): halaman ringkasan cabang`.

### Task F11: Target di halaman Performa Kurir
- [ ] Bila cabang terpilih dan target ada: tampilkan target keberhasilan dan tepat SLA di header hasil, dan penanda di atas atau di bawah target per kurir (ikon dan teks). Tanpa target atau semua cabang: penanda tidak tampil.
- [ ] Test dengan target, tanpa target (404), dan tanpa cabang.
- [ ] Commit `feat(web): target cabang di performa kurir`.

### Task I2: Integrasi lokal
- [ ] Jalankan migrasi dan seed, API, dan web. Cek tiga halaman di 1440px dan 390px (`scrollWidth <= innerWidth`), simpan screenshot, catat di deskripsi PR terakhir.
