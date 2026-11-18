---
name: backend-endpoint
description: Membuat atau mengubah endpoint REST di backend Go, berurutan dari domain, port, mock, service, adapter postgres, sampai handler httpapi, lengkap dengan test di tiap lapisan. Pakai saat diminta "buat endpoint", "tambah API", "endpoint baru untuk ...", mengerjakan task backend di plan yang menambah atau mengubah endpoint, atau saat menambah query baru di repository.
---

# Backend endpoint

Endpoint dibangun dari dalam ke luar, mengikuti arsitektur hexagonal di `backend/`. Setiap lapisan selesai beserta test-nya sebelum lanjut ke lapisan berikutnya. Semua perintah dijalankan dari folder `backend/`.

## Masukan

- Kontrak endpoint di `backend/api/openapi.yaml`: path, parameter, bentuk response sukses dan error.
- Spec dan task plan yang dilayani (`docs/specs/`, `docs/plans/`), terutama acceptance criteria dan tabel rencana test.
- Konvensi backend di `CLAUDE.md` bagian **Backend**.

**Berhenti dan tanya user bila kontrak belum ada di `openapi.yaml` atau berbeda dengan yang diminta.** Perubahan kontrak harus disepakati tim frontend dulu. Jangan mengarang bentuk response.

## Contoh rujukan di repo

Baca dulu sebelum menulis, lalu ikuti polanya. Jangan mengubah file ini kecuali task memintanya.

| Lapisan | File contoh |
|---|---|
| Domain, filter, error | `internal/domain/metric.go`, `internal/domain/performance.go`, `internal/domain/errors.go`, `internal/domain/timezone.go` |
| Port | `internal/port/port.go`, `.mockery.yaml` |
| Service dan test-nya | `internal/service/metric_service.go`, `internal/service/metric_service_test.go` |
| Repository postgres | `internal/adapter/postgres/metric_repository.go`, `internal/adapter/postgres/repository_integration_test.go` |
| Handler dan helper response | `internal/adapter/httpapi/courier_performance_handler.go` dan test-nya, `response.go`, `router.go` |
| Wiring | `cmd/api/main.go` |

## Langkah

### 1. Domain (`internal/domain/`)
- Tambah entity, filter, dan aturan bisnis sebagai struct dan method murni. Domain tidak boleh meng-import `internal/adapter`, `internal/port`, atau library infrastruktur.
- Angka KPI (rate, gap, skor) dihitung di sini sebagai method, misalnya `SuccessRate()`. Jangan menghitungnya di SQL atau di handler kalau bisa dihitung di domain.
- Filter list wajib punya `WithDefaults()` dan `Validate()` yang membatasi periode (`MaxPeriodDays`) dan `page_size` (`MaxPageSize`). Pakai ulang `MetricFilter` bila parameternya sama.
- Error validasi dibungkus dengan sentinel di `errors.go`, contohnya `fmt.Errorf("%w: ...", ErrInvalidFilter)`. Tambah sentinel baru bila butuh status HTTP lain, misalnya `ErrNotFound` untuk 404.
- Aturan waktu: penyimpanan dalam UTC. Batas hari laporan dan cut-off SLA dihitung dalam WIB lewat helper di `timezone.go` (`domain.WIB`, `DayStartWIB`). `start_date` dan `end_date` adalah tanggal WIB.
- Test: table-driven di `*_test.go`, termasuk kasus batas (pembagian dengan nol, batas periode, pergantian hari WIB).

### 2. Port (`internal/port/port.go`)
- Tambah method di interface yang ada atau buat interface baru, dinamai sesuai perilakunya (`XxxRepository`, `XxxPublisher`, `XxxCache`) tanpa prefix `I`.
- Setiap method I/O menerima `ctx context.Context` sebagai parameter pertama.
- Jalankan `make mocks`, lalu pastikan file di `internal/port/mocks/` ikut berubah. Jangan mengedit mock dengan tangan.

### 3. Service (`internal/service/`)
- Use case memanggil `filter.WithDefaults()` lalu `Validate()`, kemudian memanggil repository.
- Bungkus error dengan konteks berbahasa Indonesia: `fmt.Errorf("ambil performa kurir: %w", err)`.
- Bila memakai cache, kegagalan cache hanya di-log (`WarnContext`) dan tidak menggagalkan request. Contohnya ada di `ListCourierMetrics`.
- Dependensi masuk lewat konstruktor `NewXxxService(...)`. Tanpa variabel global dan tanpa `init()`.
- Test: table-driven dengan mock mockery (`repo.EXPECT()...`). Minimal mencakup sukses, filter ditolak (repository tidak dipanggil), dan error repository terbungkus (`assert.ErrorIs`).

### 4. Repository postgres (`internal/adapter/postgres/`)
- SQL disimpan sebagai konstanta (`const listXxxSQL = ...`) dengan parameter `$1..$n`. Jangan menyusun SQL dengan string concat dari input user.
- Hindari N+1: ambil metrik semua kurir dalam satu query agregat atau batch. Total dihitung dengan `COUNT(*) OVER ()` atau query count terpisah. Pagination memakai `LIMIT`/`OFFSET` dari `filter.Offset()`.
- Ranking dan nilai pembanding dihitung sebelum `LIMIT`, supaya stabil di semua halaman.
- Batas hari WIB dihitung di Go dengan helper domain (`DayStartWIB`, `SLADeadline`) lalu dikirim sebagai parameter query. Jangan menulis ulang aturan zona waktu di SQL dengan `AT TIME ZONE`.
- Bila query bergantung pada batas hari, tambah integration test di sekitar 23:59 dan 00:00 WIB (contoh: `TestAggregateDaily_WIBBoundary`).
- Bila butuh tabel atau kolom baru, tambah migrasi bernomor berikutnya di `backend/migrations/` (pasangan `.up.sql` dan `.down.sql`). Seed hanya berisi data sintetis tanpa nomor HP, NIK, atau alamat.
- Test: integration test dengan build tag `integration`, membaca `TEST_DATABASE_URL` dan melakukan `t.Skip` bila kosong. Ingatkan user bahwa test men-TRUNCATE tabel, jadi harus diarahkan ke database uji.

### 5. Handler HTTP (`internal/adapter/httpapi/`)
- File `xxx_handler.go`. Handler bergantung pada interface kecil yang didefinisikan di file handler (contoh `CourierPerformanceLister`), bukan pada service konkret.
- Parse query dengan `parseMetricFilter` atau helper sejenis. Kesalahan format dikirim lewat `writeError(w, 400, "invalid_query", ...)`.
- Error dari service dikirim lewat `writeDomainError`. Bila menambah sentinel domain baru, tambah pemetaannya di `response.go`. Error yang tidak dikenal tetap jadi 500 `internal_error` tanpa membocorkan pesan internal.
- Struct response terpisah dari entity domain, dengan tag JSON snake_case. Slice selalu di-encode sebagai `[]`, bukan `null` (pakai `make(..., 0, n)`).
- Response sukses dikirim lewat `writeData(w, items, pageMeta{...})` sehingga berbentuk `{"data": ..., "meta": {"page","page_size","total"}}`.
- Daftarkan route di `router.go` dengan pola `"GET /api/v1/<kata-benda-jamak-kebab-case>"`, lalu tambah parameter konstruktor `NewRouter` dan perbarui pemanggilnya.
- Jangan me-log query string atau data pribadi kurir dan penerima.
- Test dengan `httptest` lewat `NewRouter`, table-driven, membandingkan body dengan `assert.JSONEq`. Kasus wajib:
  - sukses dengan data dan `meta`
  - data kosong (`"data":[]`)
  - parameter salah format (400 `invalid_query`)
  - filter ditolak domain (400 `invalid_filter`)
  - error internal tidak bocor (500 `internal_error`)
  - parameter query ter-parse ke filter dengan benar

### 6. Wiring (`cmd/api/main.go`)
- Buat repository, service, dan handler lewat konstruktor, lalu masukkan ke `NewRouter`. Konfigurasi baru dibaca di `internal/config` dari environment variable. Bila ada variabel baru, tambahkan ke `.env.example` dan `docker-compose.yml`.

### 7. Verifikasi
Jalankan dan laporkan hasil nyatanya:
- `make test-race`
- `make cover` (minimal 80% untuk `internal/domain` dan `internal/service`, 60% untuk seluruh backend)
- `make lint`
- `make test-integration` bila repository berubah, hanya setelah user mengonfirmasi `DATABASE_URL` mengarah ke database uji
- Cek manual bila stack jalan: `curl "localhost:8080/api/v1/<path>?start_date=...&end_date=..."`, lalu cocokkan dengan `openapi.yaml`

## Keluaran

Laporkan dalam tabel, satu baris per lapisan:

| Lapisan | File | Test | Status |
|---|---|---|---|
| Domain | ... | nama test | lolos / gagal |
| Port dan mock | ... | `make mocks` | ... |
| Service | ... | ... | ... |
| Repository | ... | integration (dijalankan / dilewati, alasannya) | ... |
| Handler dan route | ... | ... | ... |
| Wiring | ... | build | ... |

Di bawah tabel, cantumkan hasil `make cover` dan `make lint`, serta perbedaan dengan `openapi.yaml` bila ada. Perbedaan kontrak tidak boleh diselesaikan sepihak. Usulkan ke user untuk disepakati dengan frontend.

Commit memakai `feat(backend): ...` atau scope modul, misalnya `feat(backend/httpapi): ...`. Commit dan push tetap menunggu permintaan user.
