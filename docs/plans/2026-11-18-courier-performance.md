# Laporan Performa Kurir Implementation Plan

> Kerjakan task satu per satu sesuai urutan dependensinya. Centang langkah (`- [ ]`) setelah selesai.

**Goal:** Endpoint `GET /api/v1/courier-performance` di backend dan halaman `/courier-performance` di frontend, dikerjakan paralel oleh tim BE dan tim FE.

**Architecture:** Kontrak API dibekukan dulu sebagai file contoh JSON di `contracts/` (Task 0). Setelah itu BE (`backend/`, Go hexagonal) dan FE (`web/`, React) jalan paralel tanpa saling menunggu: BE menguji handler terhadap bentuk kontrak, FE memakai file kontrak yang sama sebagai mock MSW. Integrasi di akhir lewat proxy Vite `/api` ke `localhost:8080`, jadi tidak perlu CORS.

**Tech Stack:** BE: Go, `net/http` stdlib (routing pola `GET /path`), `testing` + `testify`. FE: React, TypeScript, Vite, React Router, TanStack Query, Vitest, React Testing Library, MSW.

**Spec:** `docs/specs/2026-11-18-courier-performance.md`

## Asumsi yang harus dikonfirmasi sebelum eksekusi
- Bagian Backend di `CLAUDE.md` masih kosong. Plan ini memakai Go hexagonal sesuai draf konvensi BE yang sudah beredar (`cmd/`, `internal/domain|port|service|adapter`). Tim BE mengisi `CLAUDE.md` bagian Backend dulu; kalau stack-nya beda, Task B1 sampai B3 ditulis ulang, kontrak di Task 0 tetap.
- Versi Go dan versi paket npm dipilih saat scaffold: cek rilis stabil terbaru di web, jangan dari ingatan. Catat versinya di deskripsi PR.

## Global Constraints
- Zona waktu periode: Asia/Jakarta. Periode `[from 00:00 WIB, to+1 hari 00:00 WIB)`.
- Batas SLA dari `AssignedAt` (WIB): sebelum 16:00 berarti hari yang sama 23:59:59; pukul 16:00 atau lebih berarti besok 16:00:00. Tepat SLA bila `DeliveredAt <= batas`.
- Periode maksimal 31 hari inklusif.
- Rate dibulatkan 4 desimal di BE. FE tidak menghitung ulang rate, hanya memformat.
- `items` kosong dikirim sebagai `[]`, bukan `null`.
- Kode error persis: `MISSING_BRANCH_ID`, `INVALID_DATE`, `INVALID_DATE_RANGE`.
- Format persen FE: locale `id-ID`, maksimal 1 desimal (`91,7%`).
- Tidak me-log atau menampilkan nomor HP, NIK, atau alamat.
- BE dan FE hanya menyentuh foldernya sendiri (`backend/`, `web/`). `contracts/` hanya diubah lewat Task 0 atau PR yang direview kedua sisi.

## Review Focus
1. Batas waktu dihitung di WIB, bukan UTC. Kalau batas hari memakai UTC (= 07:00 WIB), paket 00:00 sampai 06:59 WIB di tanggal `from` hilang dan paket dini hari setelah `to` ikut masuk; cut-off SLA 16:00 juga bergeser. Paket 23:30 WIB di hari `to` menangkap kesalahan lain: `to` diperlakukan eksklusif. Dikunci di Task B1 (`TestSLADeadline`) dan Task B2 (`TestReport_PeriodBoundaryWIB`).
2. Pembagian dengan nol: kurir yang paketnya semua `ON_PROCESS` punya `delivered` 0. Rate harus 0, bukan NaN di JSON. Dikunci di Task B1 (`TestSummarize_ZeroDivision`).
3. `items: null`: slice Go nil ter-encode sebagai `null` dan membuat FE crash di `.map`. Dikunci di Task B3 (`TestHandler_EmptyItemsIsArray`) dan Task F3 (state kosong).
4. Error dari API tidak boleh membuat halaman putih: FE menampilkan `error.message`. Dikunci di Task F2 (`ApiError`) dan Task F3 (test state error).
5. Tombol Tampilkan ditekan sebelum input lengkap: FE tidak mengirim request dengan `branch_id` kosong. Dikunci di Task F3 (tombol disabled).

## Pembagian kerja dan model

| Task | Sisi | Pemilik | Model | Effort | Bergantung ke |
|---|---|---|---|---|---|
| 0. Kontrak API | bersama | BE + FE | haiku | low | - |
| B1. Domain dan perhitungan | BE | tim BE | sonnet | medium | 0 |
| B2. Port dan service | BE | tim BE | sonnet | high | B1 |
| B3. Handler HTTP, adapter in-memory, main | BE | tim BE | sonnet | medium | B2 |
| F1. Scaffold web dan `formatPercent` | FE | tim FE | sonnet | medium | 0 |
| F2. API client dan hook query | FE | tim FE | sonnet | medium | F1 |
| F3. Halaman performa kurir | FE | tim FE | sonnet | medium | F2 |
| I1. Integrasi lokal | bersama | BE + FE | haiku | low | B3, F3 |

Jalur B dan F tidak berbagi file, jadi bisa paralel penuh. Saran branch: `feat/contract-courier-performance` (Task 0), `feat/be-courier-performance` (B1-B3, satu PR), `feat/fe-courier-performance` (F1-F3, boleh dua PR bila lewat 400 baris: F1 sendiri, F2+F3).

---

### Task 0: Kontrak API

**Files:**
- Create: `contracts/courier-performance/ok.json`
- Create: `contracts/courier-performance/empty.json`
- Create: `contracts/courier-performance/error-invalid-range.json`

**Interfaces:**
- Produces: tiga file JSON yang dipakai B3 (test bentuk respons) dan F2/F3 (mock MSW).

- [ ] **Step 1: Tulis `ok.json`** persis contoh respons 200 di spec, ditambah item kedua `{"courier_id":"C002","courier_name":"Sari Wulandari","assigned":10,"delivered":0,"failed":0,"on_time":0,"success_rate":0,"on_time_rate":0}` supaya kasus rate 0 ikut terwakili.
- [ ] **Step 2: Tulis `empty.json`** dengan `branch_id` `"BDG01"`, periode sama, `"items": []`.
- [ ] **Step 3: Tulis `error-invalid-range.json`** persis contoh respons 400 di spec.
- [ ] **Step 4: Validasi JSON**

Run: `for f in contracts/courier-performance/*.json; do python3 -m json.tool "$f" > /dev/null && echo "ok $f"; done`
Expected: tiga baris `ok`.

- [ ] **Step 5: Commit, PR, minta review dari satu orang BE dan satu orang FE**

```bash
git add contracts/
git commit -m "docs(contracts): kontrak API performa kurir"
```

---

### Task B1: Domain dan perhitungan performa

**Files:**
- Create: `backend/go.mod` (module `github.com/black-lotus/sicepat-playbook-demo/backend`)
- Create: `backend/internal/domain/delivery.go`
- Create: `backend/internal/domain/courier_performance.go`
- Test: `backend/internal/domain/courier_performance_test.go`

**Interfaces:**
- Produces:
  - `type DeliveryStatus string` dengan konstanta `StatusDelivered = "DELIVERED"`, `StatusFailed = "FAILED"`, `StatusOnProcess = "ON_PROCESS"`
  - `var WIB = time.FixedZone("WIB", 7*3600)` (Indonesia tanpa DST, jadi zona tetap sudah benar dan tidak bergantung tzdata)
  - `type Delivery struct { AWB, CourierID, CourierName, BranchID string; Status DeliveryStatus; AssignedAt time.Time; DeliveredAt *time.Time }`
  - `func SLADeadline(assignedAt time.Time) time.Time`: konversi ke `WIB`, lalu sebelum 16:00 kembalikan hari itu 23:59:59, pukul 16:00 atau lebih kembalikan besok 16:00:00, keduanya di `WIB`.
  - `type CourierPerformance struct { CourierID, CourierName string; Assigned, Delivered, Failed, OnTime int; SuccessRate, OnTimeRate float64 }`
  - `func SummarizeCourierPerformance(deliveries []Delivery) []CourierPerformance`: kelompokkan per `CourierID`, hitung sesuai spec, urutkan `SuccessRate` menurun lalu `CourierID` menaik, selalu kembalikan slice non-nil.

- [ ] **Step 1: Tulis test table-driven `TestSummarize`** dengan kasus:
  - `"campuran"`: C001 dengan 3 DELIVERED (2 sebelum SLA, 1 sesudah), 1 FAILED, 1 ON_PROCESS. Harapan: `Assigned 5, Delivered 3, Failed 1, OnTime 2, SuccessRate 0.6, OnTimeRate 0.6667`.
  - `"urutan"`: C002 rate 1.0, C001 rate 0.5, C003 rate 0.5. Harapan urutan ID: `C002, C001, C003`.
  - `"kosong"`: input nil. Harapan `NotNil` dan `Len 0`.
  - `"tepat di deadline"`: `DeliveredAt == SLADeadline(AssignedAt)` dihitung `OnTime`.
- [ ] **Step 1b: Tulis test table-driven `TestSLADeadline`** (waktu ditulis dengan `time.Date(..., WIB)` kecuali disebut UTC):
  - ditugaskan 7 Nov 09:00 menghasilkan 7 Nov 23:59:59
  - ditugaskan 7 Nov 15:59:59 menghasilkan 7 Nov 23:59:59
  - ditugaskan 7 Nov 16:00:00 menghasilkan 8 Nov 16:00:00
  - ditugaskan 7 Nov 21:00 menghasilkan 8 Nov 16:00:00
  - ditugaskan 30 Nov 17:00 menghasilkan 1 Des 16:00:00 (ganti bulan)
  - ditugaskan 7 Nov 09:00 UTC (= 16:00 WIB) menghasilkan 8 Nov 16:00:00 WIB
  - Bandingkan dengan `Equal`, bukan `==`.
- [ ] **Step 2: Tulis `TestSummarize_ZeroDivision`**: C009 dengan 2 paket `ON_PROCESS`. Harapan `SuccessRate 0`, `OnTimeRate 0`, `Delivered 0`, dan `math.IsNaN` false untuk keduanya.
- [ ] **Step 3: Jalankan, pastikan gagal**

Run: `cd backend && go test ./internal/domain/...`
Expected: FAIL, `undefined: SummarizeCourierPerformance`.

- [ ] **Step 4: Implementasi.** Pembulatan pakai `math.Round(x*10000) / 10000`. Domain hanya meng-import stdlib.
- [ ] **Step 5: Jalankan, pastikan lolos dengan coverage**

Run: `cd backend && go test -cover ./internal/domain/...`
Expected: PASS, coverage minimal 80%.

- [ ] **Step 6: Commit**

```bash
git add backend/
git commit -m "feat(backend): hitung performa kurir di domain"
```

---

### Task B2: Port dan service laporan

**Files:**
- Create: `backend/internal/port/delivery_repository.go`
- Create: `backend/internal/service/courier_performance.go`
- Test: `backend/internal/service/courier_performance_test.go`

**Interfaces:**
- Consumes: `domain.Delivery`, `domain.SummarizeCourierPerformance` (B1).
- Produces:
  - `type DeliveryRepository interface { ListByBranchAndPeriod(ctx context.Context, branchID string, start, end time.Time) ([]domain.Delivery, error) }` dengan `start` inklusif, `end` eksklusif.
  - `type ReportQuery struct { BranchID, From, To string }` (tanggal mentah `YYYY-MM-DD`)
  - `type Report struct { BranchID, From, To string; Items []domain.CourierPerformance }`
  - `type ValidationError struct { Code, Message string }` dengan method `Error() string`
  - Konstanta kode: `CodeMissingBranchID`, `CodeInvalidDate`, `CodeInvalidDateRange` bernilai string persis seperti di spec.
  - `type CourierPerformanceService struct` dengan `func NewCourierPerformanceService(repo port.DeliveryRepository) *CourierPerformanceService` dan `func (s *CourierPerformanceService) Report(ctx context.Context, q ReportQuery) (Report, error)`

- [ ] **Step 1: Buat fake repository di file test** yang menyimpan argumen `start`, `end` terakhir dan mengembalikan delivery yang `AssignedAt` di `[start, end)`.
- [ ] **Step 2: Tulis `TestReport_Validation`** (table-driven, `errors.As` ke `*ValidationError`, cek `Code`):
  - branch kosong menghasilkan `MISSING_BRANCH_ID`
  - `From: "2026-13-01"` menghasilkan `INVALID_DATE`
  - `From: ""` menghasilkan `INVALID_DATE`
  - `From: "2026-11-08", To: "2026-11-07"` menghasilkan `INVALID_DATE_RANGE`
  - `From: "2026-11-01", To: "2026-12-02"` (32 hari) menghasilkan `INVALID_DATE_RANGE`
  - `From: "2026-11-01", To: "2026-12-01"` (31 hari) lolos tanpa error
  - `From: "2026-11-07", To: "2026-11-07"` (1 hari) lolos tanpa error
- [ ] **Step 3: Tulis `TestReport_PeriodBoundaryWIB`**: query `2026-11-01` sampai `2026-11-07`. Harapan argumen ke repo: `start` `Equal` `2026-11-01T00:00:00+07:00`, `end` `Equal` `2026-11-08T00:00:00+07:00`. Seed fake dengan empat paket satu kurir: `2026-11-01T05:00:00+07:00` dan `2026-11-07T23:30:00+07:00` (terhitung), `2026-10-31T23:59:00+07:00` dan `2026-11-08T00:00:00+07:00` (tidak terhitung); harapan `Assigned` total 2.
- [ ] **Step 4: Tulis `TestReport_RepoError`**: fake mengembalikan error; harapan error dibungkus dan `errors.Is` ke error asli true, bukan `ValidationError`.
- [ ] **Step 5: Jalankan, pastikan gagal**

Run: `cd backend && go test ./internal/service/...`
Expected: FAIL, simbol belum ada.

- [ ] **Step 6: Implementasi.** Zona waktu memakai `domain.WIB`. Parse tanggal dengan `time.ParseInLocation("2006-01-02", ...)`. Bungkus error repo: `fmt.Errorf("ambil delivery cabang: %w", err)`.
- [ ] **Step 7: Jalankan, pastikan lolos**

Run: `cd backend && go test -race -cover ./internal/...`
Expected: PASS, coverage `service` minimal 80%.

- [ ] **Step 8: Commit**

```bash
git add backend/
git commit -m "feat(backend): service laporan performa kurir dengan periode WIB"
```

---

### Task B3: Handler HTTP, adapter in-memory, dan entrypoint

**Files:**
- Create: `backend/internal/adapter/memory/delivery_repository.go`
- Create: `backend/internal/adapter/http/courier_performance_handler.go`
- Create: `backend/cmd/api/main.go`
- Test: `backend/internal/adapter/http/courier_performance_handler_test.go`

**Interfaces:**
- Consumes: `service.CourierPerformanceService`, `service.ValidationError`, `port.DeliveryRepository` (B2).
- Produces:
  - `func NewDeliveryRepository(seed []domain.Delivery) *memory.DeliveryRepository` (memenuhi `port.DeliveryRepository`) dan `func SeedData() []domain.Delivery` berisi minimal 3 kurir di `JKT01` dan 0 kurir di `BDG01`, tanggal 2026-11-01 sampai 2026-11-07.
  - `func NewCourierPerformanceHandler(svc *service.CourierPerformanceService) http.Handler` untuk `GET /api/v1/courier-performance`.
  - `cmd/api` mendengarkan di alamat dari env `HTTP_ADDR`, default `:8080`.

- [ ] **Step 1: Tulis `TestHandler_OK`** dengan `httptest`: response 200, `Content-Type: application/json`, dan kunci JSON level atas serta kunci tiap item sama persis dengan `contracts/courier-performance/ok.json` (baca file lewat path relatif `../../../../contracts/courier-performance/ok.json`, bandingkan himpunan kuncinya, bukan nilainya).
- [ ] **Step 2: Tulis `TestHandler_EmptyItemsIsArray`**: `branch_id=BDG01`, body mengandung `"items":[]` dan tidak mengandung `"items":null`.
- [ ] **Step 3: Tulis `TestHandler_ValidationError`**: `from` setelah `to` menghasilkan 400 dan body `{"error":{"code":"INVALID_DATE_RANGE", ...}}`.
- [ ] **Step 4: Tulis `TestHandler_InternalError`**: repo gagal menghasilkan 500 dengan `{"error":{"code":"INTERNAL","message":"Terjadi kesalahan, coba lagi."}}`, tanpa membocorkan pesan error asli.
- [ ] **Step 5: Jalankan, pastikan gagal**

Run: `cd backend && go test ./internal/adapter/...`
Expected: FAIL, simbol belum ada.

- [ ] **Step 6: Implementasi.** Handler memetakan field domain ke JSON snake_case lewat struct DTO di package `http`, bukan tag JSON di domain. Log error 500 dengan `slog` beserta `branch_id`, tanpa data pribadi. `main.go` merangkai repo memory, service, handler, dan menangani shutdown dengan `signal.NotifyContext`.
- [ ] **Step 7: Jalankan semua test dan lint**

Run: `cd backend && go test -race -cover ./... && golangci-lint run`
Expected: PASS, lint bersih, coverage total minimal 60%.

- [ ] **Step 8: Cek manual**

Run: `cd backend && go run ./cmd/api` lalu di terminal lain `curl -s 'localhost:8080/api/v1/courier-performance?branch_id=JKT01&from=2026-11-01&to=2026-11-07'`
Expected: JSON 200 dengan minimal 3 item.

- [ ] **Step 9: Commit dan buka PR BE**

```bash
git add backend/
git commit -m "feat(backend): endpoint GET /api/v1/courier-performance"
```

---

### Task F1: Scaffold web dan `formatPercent`

**Files:**
- Create: `web/` dari template Vite `react-ts`
- Modify: `web/package.json` (script `test`, `test:coverage`, `lint`, `typecheck`)
- Modify: `web/vite.config.ts` (proxy `/api` ke `http://localhost:8080`, konfigurasi Vitest `environment: "jsdom"`, `setupFiles: "./src/test/setup.ts"`)
- Create: `web/src/test/setup.ts`
- Create: `web/src/lib/format-percent.ts`
- Test: `web/src/lib/format-percent.test.ts`

**Interfaces:**
- Produces:
  - `export function formatPercent(rate: number): string`
  - Script npm: `test` (`vitest run`), `test:coverage`, `lint`, `typecheck` (`tsc --noEmit`)
  - `web/src/test/setup.ts` sebagai tempat MSW server di F2.

- [ ] **Step 1: Scaffold** dengan `npm create vite@latest web -- --template react-ts`, lalu pasang `@tanstack/react-query`, `react-router-dom`, dan dev dependency `vitest`, `jsdom`, `@testing-library/react`, `@testing-library/jest-dom`, `@testing-library/user-event`, `msw`. Pastikan `tsconfig` `strict: true`.
- [ ] **Step 2: Tulis test `formatPercent`** dengan `it.each`: `0.9167` menjadi `"91,7%"`, `0.9` menjadi `"90%"`, `0` menjadi `"0%"`, `1` menjadi `"100%"`.
- [ ] **Step 3: Jalankan, pastikan gagal**

Run: `cd web && npm test -- format-percent`
Expected: FAIL, modul tidak ditemukan.

- [ ] **Step 4: Implementasi** dengan `Intl.NumberFormat("id-ID", { style: "percent", maximumFractionDigits: 1 })` yang dibuat sekali di level modul. Kalau output runtime memuat spasi sebelum `%`, sesuaikan test dan catat di PR; jangan memotong string manual.
- [ ] **Step 5: Jalankan semua pemeriksaan**

Run: `cd web && npm test && npm run lint && npm run typecheck`
Expected: semua lolos.

- [ ] **Step 6: Commit**

```bash
git add web/
git commit -m "chore(web): scaffold dashboard dan helper formatPercent"
```

---

### Task F2: API client dan hook query

**Files:**
- Create: `web/src/api/http-client.ts`
- Create: `web/src/api/courier-performance.ts`
- Create: `web/src/test/msw-handlers.ts`
- Modify: `web/src/test/setup.ts` (jalankan MSW `setupServer` dengan `msw-handlers.ts`)
- Test: `web/src/api/courier-performance.test.tsx`

**Interfaces:**
- Consumes: file JSON di `contracts/courier-performance/` (Task 0), di-import langsung dari `../../../contracts/courier-performance/*.json`.
- Produces:
  - `export class ApiError extends Error { constructor(public status: number, public code: string, message: string) }`
  - `export async function getJson<T>(path: string, params: Record<string, string>): Promise<T>`: base URL dari `import.meta.env.VITE_API_BASE_URL ?? ""`; respons non-2xx dengan body `{error:{code,message}}` dilempar sebagai `ApiError`; respons non-2xx tanpa body itu dilempar sebagai `ApiError(status, "UNKNOWN", "Terjadi kesalahan, coba lagi.")`.
  - Tipe `CourierPerformanceItem` dan `CourierPerformanceReport` sesuai kontrak (snake_case, apa adanya dari API).
  - `export type CourierPerformanceQuery = { branchId: string; from: string; to: string }`
  - `export function useCourierPerformance(query: CourierPerformanceQuery | null)`: hook TanStack Query, `enabled: query !== null`, query key `["courier-performance", query]`.
  - `msw-handlers.ts` mengekspor `handlers` (default: `BDG01` mengembalikan `empty.json`, `from > to` mengembalikan 400 `error-invalid-range.json`, selain itu `ok.json`).

- [ ] **Step 1: Tulis test hook** memakai `renderHook` dengan `QueryClientProvider` (retry dimatikan):
  - query `JKT01` menghasilkan `data.items` panjang 2 dan item pertama `courier_id` `"C001"`.
  - query `null` tidak mengirim request (`fetchStatus` `"idle"`).
  - query `from: "2026-11-08", to: "2026-11-07"` menghasilkan `error` instance `ApiError` dengan `code` `"INVALID_DATE_RANGE"` dan `message` sama dengan isi `error-invalid-range.json`.
- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `cd web && npm test -- courier-performance`
Expected: FAIL, modul tidak ditemukan.

- [ ] **Step 3: Implementasi** sesuai Interfaces. Parameter dikirim sebagai `branch_id`, `from`, `to`. Aktifkan `resolveJsonModule` di `tsconfig` agar file kontrak bisa di-import; file kontrak hanya di-import dari kode test, tidak dari kode aplikasi.
- [ ] **Step 4: Jalankan, pastikan lolos dengan coverage**

Run: `cd web && npm run test:coverage`
Expected: PASS, coverage `src/api` dan `src/lib` minimal 80%.

- [ ] **Step 5: Commit**

```bash
git add web/
git commit -m "feat(web): client dan hook query performa kurir"
```

---

### Task F3: Halaman performa kurir

**Files:**
- Create: `web/src/pages/courier-performance/CourierPerformancePage.tsx`
- Create: `web/src/features/courier-performance/CourierPerformanceFilter.tsx`
- Create: `web/src/features/courier-performance/CourierPerformanceTable.tsx`
- Modify: `web/src/app/` router (rute `/courier-performance`, dan `/` redirect ke sana) dan provider `QueryClientProvider`
- Test: `web/src/pages/courier-performance/CourierPerformancePage.test.tsx`

**Interfaces:**
- Consumes: `useCourierPerformance`, `ApiError`, `CourierPerformanceQuery` (F2), `formatPercent` (F1), `handlers` MSW (F2).
- Produces:
  - `CourierPerformanceFilter` props: `{ onSubmit: (q: CourierPerformanceQuery) => void }`; input berlabel "Cabang", "Dari tanggal" (`type="date"`), "Sampai tanggal" (`type="date"`), tombol "Tampilkan" disabled selama salah satu input kosong.
  - `CourierPerformanceTable` props: `{ items: CourierPerformanceItem[] }`; kolom persis: Kurir, Ditugaskan, Terkirim, Gagal, Success rate, Tepat SLA.

- [ ] **Step 1: Tulis test halaman** (render dengan router dan `QueryClientProvider`, interaksi lewat `userEvent`):
  - `"tombol disabled sebelum input lengkap"`: hanya isi Cabang, tombol "Tampilkan" `toBeDisabled()`.
  - `"menampilkan tabel"`: isi `JKT01`, `2026-11-01`, `2026-11-07`, klik Tampilkan; harapan muncul teks "Memuat...", lalu baris "Budi Santoso" dengan `90%` dan `91,7%`, dan baris "Sari Wulandari" dengan `0%`.
  - `"data kosong"`: cabang `BDG01`; harapan teks "Belum ada data untuk periode ini." dan tabel tidak dirender.
  - `"error dari API"`: `from` `2026-11-08`, `to` `2026-11-07`; harapan `role="alert"` berisi pesan dari `error-invalid-range.json`.
- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `cd web && npm test -- CourierPerformancePage`
Expected: FAIL, modul tidak ditemukan.

- [ ] **Step 3: Implementasi.** Halaman menyimpan `CourierPerformanceQuery | null` di state, diisi saat filter submit, lalu memberikannya ke hook. Pesan error: `ApiError.message` bila instance `ApiError`, selain itu "Terjadi kesalahan, coba lagi.". Angka di tabel tidak dihitung ulang.
- [ ] **Step 4: Jalankan semua pemeriksaan**

Run: `cd web && npm run test:coverage && npm run lint && npm run typecheck`
Expected: semua lolos, coverage total minimal 60%.

- [ ] **Step 5: Commit dan buka PR FE** (sertakan screenshot halaman dengan data, kosong, dan error)

```bash
git add web/
git commit -m "feat(web): halaman performa kurir"
```

---

### Task I1: Integrasi lokal

**Files:** tidak ada perubahan kode. Bila ada ketidakcocokan, perbaikannya masuk sebagai PR terpisah ke sisi yang salah, dan `contracts/` hanya diubah atas persetujuan kedua sisi.

- [ ] **Step 1: Jalankan BE**: `cd backend && go run ./cmd/api`
- [ ] **Step 2: Jalankan FE**: `cd web && npm run dev`, buka `http://localhost:5173/courier-performance`
- [ ] **Step 3: Cek tiga skenario**: `JKT01` 2026-11-01 sampai 2026-11-07 menampilkan minimal 3 kurir; `BDG01` menampilkan pesan data kosong; `from` setelah `to` menampilkan pesan error dari BE.
- [ ] **Step 4: Catat hasil** di deskripsi PR terakhir yang di-merge, dengan screenshot ketiga skenario.
