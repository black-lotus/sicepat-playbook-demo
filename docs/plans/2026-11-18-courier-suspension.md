# Suspend Kurir karena Fraud Implementation Plan

> Kerjakan task satu per satu sesuai urutan dependensinya. Centang langkah (`- [ ]`) setelah selesai.

**Goal:** Pencatatan kasus fraud kurir dengan bukti, approval satu tahap, banding sederhana, dan daftar kurir yang sedang disuspend. Backend menyediakan `/api/v1/suspensions`, frontend menyediakan halaman `/suspensions`.

**Architecture:** Kontrak dibekukan dulu di `backend/api/openapi.yaml` (Task 0). Setelah itu BE dan FE jalan paralel. BE: domain murni (status, aturan transisi, `allowed_actions`), repository PostgreSQL dengan transisi status bersyarat dalam satu transaksi, penyimpanan bukti di filesystem lewat port, dan middleware identitas dari header. FE: konteks pengguna "Masuk sebagai", client yang mengirim header identitas, lalu tiga halaman. Tidak ada worker: status berbasis tanggal dihitung saat data dibaca.

**Tech Stack:** Mengikuti boilerplate. BE: Go, `net/http`, pgx, `testing` + `testify` + mock `mockery` (`make mocks`). FE: React, TypeScript, Vite, React Router, TanStack Query, Vitest, Testing Library, MSW. Tidak ada dependensi baru.

**Spec:** `docs/specs/2026-11-18-courier-suspension.md`

## Global Constraints
- Ikuti konvensi di `CLAUDE.md`: respons `{data, meta}`, error `{error:{code,message}}`, field snake_case, path kebab-case jamak.
- Waktu kejadian disimpan UTC. `start_date` dan `end_date` adalah tanggal WIB (`time.FixedZone("WIB", 7*3600)`); "hari ini" selalu dihitung dalam WIB dari jam yang di-inject, bukan `time.Now()` langsung.
- Durasi hanya 3, 7, 14, 30. `end_date` inklusif = `start_date + duration_days - 1`.
- Status `SCHEDULED`, `ACTIVE`, dan `COMPLETED` tidak pernah disimpan; dihitung dari status keputusan dan tanggal.
- Kode error persis tabel di spec.
- Frontend tidak menghitung ulang status atau hak aksi. Tombol hanya berasal dari `allowed_actions`.
- Tidak me-log `summary`, catatan, alasan banding, nama file, atau isi file. Data contoh dan fixture sintetis.
- Tidak menyentuh KPI, `courier_daily_metrics`, atau endpoint performa.
- BE hanya menyentuh `backend/` (plus `docker-compose.yml` untuk volume bukti), FE hanya `web/`. Kontrak hanya diubah lewat PR yang direview kedua sisi.

## Review Focus
1. **Transisi status saat bersamaan.** Dua approver yang menekan bersamaan tidak boleh menghasilkan dua keputusan. Update status wajib bersyarat (`WHERE id = $1 AND status = 'SUBMITTED'`) dalam transaksi yang sama dengan insert riwayat; 0 baris berarti 409. Dikunci di B2 (`TestApprove_Concurrent` integration).
2. **Batas hari WIB.** Kasus mulai 20 Nov harus `ACTIVE` pada 19 Nov 17:00 UTC (= 20 Nov 00:00 WIB). Kalau "hari ini" diambil dari tanggal UTC, kurir terlihat belum disuspend selama 7 jam pertama. Dikunci di B1 (`TestEffectiveStatus`).
3. **Satu kasus terbuka per kurir.** Dua pengajuan bersamaan untuk kurir yang sama tidak boleh sama-sama lolos. Kunci baris kurir (`SELECT ... FROM couriers WHERE id = $1 FOR UPDATE`) sebelum mengecek kasus terbuka. Dikunci di B2 (`TestCreate_OpenCaseConcurrent` integration).
4. **Pemisahan tugas dan akses bukti.** Pengaju tidak bisa approve/reject kasusnya sendiri; `viewer` tidak bisa mengunduh bukti dan tidak menerima daftar bukti. Dikunci di B1 (`TestAllowedActions`) dan B4 (test handler 403).
5. **Validasi file dari isi.** File `.exe` yang diganti nama `.pdf` harus ditolak; batas ukuran dicek tanpa membaca seluruh body ke memori (`http.MaxBytesReader`). Dikunci di B4 (`TestCreate_EvidenceValidation`).

## Pembagian kerja dan model

| Task | Sisi | Pemilik | Model | Effort | Bergantung ke |
|---|---|---|---|---|---|
| 0. Kontrak OpenAPI | bersama | BE + FE | sonnet | medium | - |
| B1. Domain: status, aturan, `allowed_actions` | BE | tim BE | sonnet | high | 0 |
| B2. Migrasi, repository, penyimpanan bukti | BE | tim BE | sonnet | high | B1 |
| B3. Service dan port | BE | tim BE | sonnet | high | B2 |
| B4. Handler HTTP, middleware identitas, wiring | BE | tim BE | sonnet | medium | B3 |
| F1. Identitas pengguna, client, tipe, hook | FE | tim FE | sonnet | medium | 0 |
| F2. Halaman daftar | FE | tim FE | sonnet | medium | F1 |
| F3. Halaman form pengajuan | FE | tim FE | sonnet | medium | F1 |
| F4. Halaman detail dan aksi | FE | tim FE | sonnet | high | F1 |
| I1. Integrasi lokal | bersama | BE + FE | haiku | low | B4, F2, F3, F4 |

Saran PR: Task 0 sendiri; B1; B2; B3+B4; F1+F2; F3; F4. Tiap PR di bawah 400 baris di luar mock hasil generate dan lockfile.

Plan performa kurir juga mengubah `backend/internal/adapter/httpapi/router.go`, `backend/cmd/api/main.go`, dan `web/src/app/router.tsx`. Siapa pun yang merge belakangan melakukan rebase dan menggabungkan registrasi route; jangan menghapus route milik fitur lain. Bila `domain.WIB` dari plan itu sudah ada di `main`, pakai itu; bila belum, buat `backend/internal/domain/timezone.go` berisi `var WIB = time.FixedZone("WIB", 7*60*60)` saja, supaya bisa digabung tanpa duplikasi.

---

### Task 0: Kontrak OpenAPI

**Files:**
- Modify: `backend/api/openapi.yaml`

**Interfaces:**
- Produces: path `/api/v1/suspensions` beserta turunannya, schema `Suspension`, `SuspensionDetail`, `SuspensionAppeal`, `SuspensionEvidence`, `SuspensionHistory`, dan security header `X-User-ID`/`X-User-Role` yang menjadi acuan B4 dan F1.

- [ ] **Step 1: Tambah parameter bersama** `XUserID` dan `XUserRole` (`in: header`, wajib, enum peran) dan pakai di semua operasi suspend.
- [ ] **Step 2: Tambah path** sesuai tabel Kontrak API di spec, termasuk `multipart/form-data` untuk `POST /api/v1/suspensions` (`evidence` berupa array `string` `format: binary`, `minItems: 1`, `maxItems: 5`) dan respons file untuk unduh bukti.
- [ ] **Step 3: Tambah schema** dengan semua field wajib (`nullable: true` untuk `appeal_status`, `decision`, `appeal`, dan `note`), enum `status`, `appeal_status`, `fraud_type`, `duration_days` (`[3, 7, 14, 30]`), `history[].action`, `allowed_actions[]`. Daftarkan semua kode error dari spec di `description` respons 4xx.
- [ ] **Step 4: Validasi YAML**

Run: `python3 -c "import yaml; yaml.safe_load(open('backend/api/openapi.yaml')); print('ok')"`
Expected: `ok`.

- [ ] **Step 5: Commit, PR, minta review dari satu orang BE dan satu orang FE**

```bash
git add backend/api/openapi.yaml
git commit -m "docs(contracts): kontrak API suspend kurir"
```

---

### Task B1: Domain: status, aturan, dan `allowed_actions`

**Files:**
- Create: `backend/internal/domain/suspension.go`, `backend/internal/domain/suspension_test.go`
- Create (bila belum ada): `backend/internal/domain/timezone.go`
- Modify: `backend/internal/domain/errors.go`

**Interfaces:**
- Produces:
  - `type Role string` dengan `RoleInvestigator`, `RoleApprover`, `RoleViewer`; `type Actor struct { ID string; Role Role }`.
  - `type FraudType string` (5 konstanta sesuai spec) dan `func (t FraudType) Valid() bool`.
  - `type DecisionStatus string`: `DecisionSubmitted`, `DecisionApproved`, `DecisionRejected`, `DecisionCancelled` (yang disimpan).
  - `type SuspensionStatus string`: `PENDING_APPROVAL`, `REJECTED`, `SCHEDULED`, `ACTIVE`, `COMPLETED`, `CANCELLED` (yang dikirim ke API).
  - `type AppealStatus string`: `AppealPending`, `AppealAccepted`, `AppealRejected` (string kosong = belum ada banding).
  - `type Suspension struct` berisi semua field spec, termasuk `StartDate time.Time` (tengah malam WIB), `Decision *Decision`, `Appeal *Appeal`, `Evidence []Evidence`, `History []HistoryEntry`.
  - `func NewSuspension(input NewSuspensionInput, courier Courier, today time.Time) (Suspension, error)`: validasi durasi, jenis, panjang `summary`, dan `start_date >= today`; menghitung `EndDate`.
  - `func (s Suspension) EndDate() time.Time` = `StartDate.AddDate(0, 0, DurationDays-1)`.
  - `func (s Suspension) EffectiveStatus(now time.Time) SuspensionStatus`: konversi `now` ke WIB, ambil tanggalnya, bandingkan dengan `StartDate` dan `EndDate`.
  - `func (s Suspension) AllowedActions(actor Actor, now time.Time) []Action` dengan `ActionApprove`, `ActionReject`, `ActionSubmitAppeal`, `ActionDecideAppeal`; selalu non-nil.
  - `func (s Suspension) CanApprove(actor Actor, now time.Time) error` dan padanannya untuk reject, ajukan banding, putuskan banding, mengembalikan error domain yang dipetakan B4 ke kode di spec.
  - Error baru: `ErrForbidden`, `ErrSelfApproval`, `ErrInvalidRequest`, `ErrStartDatePassed`, `ErrInvalidTransition`, `ErrAppealAlreadySubmitted`, `ErrNotFound`, `ErrCourierNotFound`, `ErrCourierHasOpenCase`, `ErrInvalidEvidence`.
  - `func TodayWIB(now time.Time) time.Time`: tengah malam WIB untuk tanggal `now` di WIB.

- [ ] **Step 1: Tulis `TestNewSuspension`** (table-driven, `today` = 20 Nov 2026 WIB): durasi 3/7/14/30 valid dan 5 tidak valid (`ErrInvalidRequest`); `fraud_type` kosong dan tidak dikenal ditolak; `summary` 19 dan 2001 karakter ditolak, 20 dan 2000 diterima; `start_date` 19 Nov menghasilkan `ErrStartDatePassed`, 20 Nov diterima; durasi 3 mulai 20 Nov menghasilkan `EndDate` 22 Nov; durasi 30 mulai 20 Nov menghasilkan 19 Des.
- [ ] **Step 2: Tulis `TestEffectiveStatus`** untuk kasus disetujui 3 hari mulai 20 Nov, dengan `now` dalam UTC:
  - 19 Nov 16:59:59 UTC (19 Nov 23:59:59 WIB) menghasilkan `SCHEDULED`
  - 19 Nov 17:00:00 UTC (20 Nov 00:00 WIB) menghasilkan `ACTIVE`
  - 22 Nov 16:59:59 UTC (22 Nov 23:59:59 WIB) menghasilkan `ACTIVE`
  - 22 Nov 17:00:00 UTC (23 Nov 00:00 WIB) menghasilkan `COMPLETED`
  - `DecisionSubmitted`, `DecisionRejected`, `DecisionCancelled` menghasilkan `PENDING_APPROVAL`, `REJECTED`, `CANCELLED` di tanggal mana pun.
- [ ] **Step 3: Tulis `TestAllowedActions`** (table-driven, aktor × status):
  - `PENDING_APPROVAL`: approver lain mendapat `[APPROVE, REJECT]`; approver yang juga pengaju mendapat `[]`; investigator dan viewer mendapat `[]`.
  - `SCHEDULED` dan `ACTIVE` tanpa banding: investigator mendapat `[SUBMIT_APPEAL]`; approver dan viewer mendapat `[]`.
  - Banding `PENDING`: approver mendapat `[DECIDE_APPEAL]`; investigator mendapat `[]`.
  - Banding sudah diputuskan, `COMPLETED`, `REJECTED`, `CANCELLED`: semua mendapat `[]`.
  - Hasil tidak pernah `nil`.
- [ ] **Step 4: Tulis test `Can*`** untuk memastikan error yang tepat: viewer approve menghasilkan `ErrForbidden`; pengaju approve menghasilkan `ErrSelfApproval`; approve kasus `ACTIVE` menghasilkan `ErrInvalidTransition`; approve kasus yang `start_date`-nya sudah lewat menghasilkan `ErrStartDatePassed`; banding kedua menghasilkan `ErrAppealAlreadySubmitted`; banding saat `COMPLETED` menghasilkan `ErrInvalidTransition`.
- [ ] **Step 5: Jalankan, pastikan gagal**

Run: `cd backend && go test ./internal/domain/...`
Expected: FAIL, simbol belum ada.

- [ ] **Step 6: Implementasi.** Domain hanya meng-import stdlib. `AllowedActions` dibangun dari fungsi `Can*` supaya aturan hanya ditulis sekali.
- [ ] **Step 7: Jalankan dengan coverage**

Run: `cd backend && go test -cover ./internal/domain/...`
Expected: PASS, coverage minimal 80%.

- [ ] **Step 8: Commit**

```bash
git add backend/internal/domain/
git commit -m "feat(backend): domain suspend kurir dengan status berbasis tanggal WIB"
```

---

### Task B2: Migrasi, repository, dan penyimpanan bukti

**Files:**
- Create: `backend/migrations/<nomor berikutnya>_suspensions.up.sql`, `.down.sql` (pakai nomor berikutnya yang belum terpakai di `main` saat itu)
- Create: `backend/internal/adapter/postgres/suspension_repository.go`, `suspension_repository_integration_test.go`
- Create: `backend/internal/adapter/filestore/evidence_store.go`, `evidence_store_test.go`

**Interfaces:**
- Consumes: tipe domain dari B1, tabel `couriers` yang sudah ada.
- Produces:
  - Tabel `suspension_cases` (status keputusan `CHECK IN ('SUBMITTED','APPROVED','REJECTED','CANCELLED')`, `duration_days CHECK IN (3,7,14,30)`, `start_date DATE`, kolom keputusan dan banding, `courier_name` dan `branch_id` sebagai snapshot), `suspension_evidence`, `suspension_history`. Index `(branch_id, start_date)`, `(courier_id)`, `(submitted_at DESC)`.
  - `SuspensionRepository` dengan method:
    - `Create(ctx, s domain.Suspension, evidence []domain.Evidence, today time.Time) (int64, error)`: dalam satu transaksi, kunci baris kurir `FOR UPDATE` (`ErrCourierNotFound` bila tidak ada), cek kasus terbuka (`SUBMITTED`, atau `APPROVED` dengan tanggal akhir >= `today`, lalu `ErrCourierHasOpenCase`), insert kasus, bukti, dan riwayat `SUBMITTED`.
    - `Get(ctx, id) (domain.Suspension, error)`: termasuk bukti dan riwayat; `ErrNotFound`.
    - `List(ctx, filter domain.SuspensionFilter, now time.Time) (domain.SuspensionPage, error)`: filter status efektif dihitung di SQL dengan tanggal WIB hari ini sebagai parameter; `COUNT(*) OVER ()` untuk total.
    - `Transition(ctx, id int64, from, to domain.DecisionStatus, entry domain.HistoryEntry, apply func(*domain.Suspension)) error`: `UPDATE ... WHERE id = $1 AND status = $from`; 0 baris menghasilkan `ErrInvalidTransition`; insert riwayat di transaksi yang sama.
    - `SubmitAppeal` dan `DecideAppeal` dengan syarat yang sama (`appeal_status IS NULL` / `= 'PENDING'`).
    - `GetEvidence(ctx, caseID, evidenceID int64) (domain.Evidence, error)`.
  - `filestore.EvidenceStore` dengan `Save(ctx, key string, r io.Reader) error`, `Open(ctx, key string) (io.ReadCloser, error)`, `Delete(ctx, key string) error`. Key dibuat backend (`<case>/<16 byte acak dari crypto/rand dalam hex>`), nama file asli tidak dipakai sebagai path.

- [ ] **Step 1: Tulis migrasi** dan pastikan `make migrate-up` lalu `make migrate-down` berjalan bersih di database lokal.
- [ ] **Step 2: Tulis integration test** (`//go:build integration`, pola sama dengan `repository_integration_test.go`):
  - `TestCreateAndGet`: kasus dengan 2 bukti tersimpan; `Get` mengembalikan bukti dan satu riwayat `SUBMITTED`.
  - `TestCreate_CourierNotFound` dan `TestCreate_OpenCase`: kasus kedua untuk kurir yang sama saat kasus pertama `SUBMITTED` menghasilkan `ErrCourierHasOpenCase`; setelah kasus pertama `REJECTED`, pengajuan kedua diterima.
  - `TestCreate_OpenCaseConcurrent`: dua goroutine mengajukan untuk kurir yang sama; tepat satu sukses.
  - `TestApprove_Concurrent`: dua goroutine `Transition(SUBMITTED → APPROVED)`; tepat satu sukses, satu `ErrInvalidTransition`, riwayat berisi satu `APPROVED`.
  - `TestList_StatusFilter`: dengan `now` = 21 Nov 03:00 UTC, kasus mulai 20 Nov (3 hari) muncul di `ACTIVE`, kasus mulai 25 Nov muncul di `SCHEDULED`, kasus mulai 10 Nov (3 hari) muncul di `COMPLETED`; filter `branch_id` bekerja; daftar kosong menghasilkan slice non-nil.
- [ ] **Step 3: Tulis unit test `evidence_store_test.go`** dengan `t.TempDir()`: simpan, buka, hapus; key berisi `..` ditolak.
- [ ] **Step 4: Jalankan, pastikan gagal**

Run: `cd backend && go test ./internal/adapter/filestore/... && make test-integration`
Expected: FAIL, simbol belum ada.

- [ ] **Step 5: Implementasi.** Semua query memakai parameter, tanpa N+1: `Get` memakai tiga query (kasus, bukti, riwayat), `List` satu query.
- [ ] **Step 6: Jalankan**

Run: `cd backend && make test-race && make test-integration && make lint`
Expected: PASS, lint bersih.

- [ ] **Step 7: Commit**

```bash
git add backend/migrations/ backend/internal/adapter/postgres/ backend/internal/adapter/filestore/
git commit -m "feat(backend): penyimpanan kasus suspend dan bukti investigasi"
```

---

### Task B3: Service dan port

**Files:**
- Modify: `backend/internal/port/port.go` (lalu `make mocks`)
- Create: `backend/internal/service/suspension_service.go`, `suspension_service_test.go`

**Interfaces:**
- Consumes: domain dari B1, repository dan store dari B2.
- Produces:
  - `port.SuspensionRepository`, `port.EvidenceStore`, dan `port.Clock` (`Now() time.Time`).
  - `SuspensionService` dengan method `Create(ctx, actor, input, files []EvidenceUpload)`, `Get(ctx, actor, id)`, `List(ctx, actor, filter)`, `Approve(ctx, actor, id, note)`, `Reject(ctx, actor, id, note)`, `SubmitAppeal(ctx, actor, id, reason)`, `DecideAppeal(ctx, actor, id, decision, note)`, dan `OpenEvidence(ctx, actor, caseID, evidenceID) (domain.Evidence, io.ReadCloser, error)`. Semua aksi mengembalikan `SuspensionView` = `Suspension` + `Status` efektif + `AllowedActions` untuk aktor itu.

- [ ] **Step 1: Tulis test service** dengan mock repository, store, dan clock tetap (21 Nov 03:00 UTC):
  - `Create`: peran selain investigator menghasilkan `ErrForbidden` tanpa menyimpan apa pun; file disimpan lebih dulu, lalu bila repository gagal, semua file yang sudah tersimpan dihapus dan error dikembalikan.
  - `Approve`: aturan `Can*` dipanggil dengan `now` dari clock; `ErrSelfApproval` tidak memanggil repository; sukses memanggil `Transition(SUBMITTED → APPROVED)` dengan entri riwayat berisi aktor dan catatan.
  - `Reject` tanpa catatan menghasilkan `ErrInvalidRequest`.
  - `DecideAppeal` `ACCEPTED` mengubah status keputusan menjadi `CANCELLED`; `REJECTED` tidak mengubahnya.
  - `Get` untuk `viewer` mengembalikan `Evidence` kosong; `OpenEvidence` untuk `viewer` menghasilkan `ErrForbidden`.
- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `cd backend && go test ./internal/service/...`
Expected: FAIL.

- [ ] **Step 3: Implementasi.** Service tidak memanggil `time.Now()` langsung. Kegagalan menghapus file cadangan hanya di-log dengan key, tanpa nama file asli.
- [ ] **Step 4: Jalankan dengan coverage**

Run: `cd backend && make mocks && go test -race -cover ./internal/service/...`
Expected: PASS, coverage minimal 80%.

- [ ] **Step 5: Commit**

```bash
git add backend/internal/port/ backend/internal/service/
git commit -m "feat(backend): service pengajuan, approval, dan banding suspend kurir"
```

---

### Task B4: Handler HTTP, middleware identitas, dan wiring

**Files:**
- Create: `backend/internal/adapter/httpapi/actor.go`, `actor_test.go`
- Create: `backend/internal/adapter/httpapi/suspension_handler.go`, `suspension_handler_test.go`
- Modify: `backend/internal/adapter/httpapi/response.go` (pemetaan error baru)
- Modify: `backend/internal/adapter/httpapi/router.go`, `backend/cmd/api/main.go`
- Modify: `backend/internal/config/config.go` (`EVIDENCE_DIR`, default `./data/evidence`)
- Modify: `docker-compose.yml` (volume bernama untuk `EVIDENCE_DIR` di service `backend`), `backend/.env.example`

**Interfaces:**
- Consumes: `SuspensionService` dari B3.
- Produces:
  - Middleware `requireActor` yang membaca `X-User-ID` dan `X-User-Role` ke context; 401 `unauthenticated` bila kosong atau tidak dikenal. Hanya dipasang di route suspend.
  - Route sesuai tabel Kontrak API di spec, dengan DTO snake_case.

- [ ] **Step 1: Tulis test handler** dengan `httptest` dan service palsu:
  - `TestCreate_EvidenceValidation`: 0 file, 6 file, file 10 MB + 1 byte, dan file berisi header `MZ` bernama `bukti.pdf` menghasilkan 400 `invalid_evidence`; PDF, JPG, dan PNG valid diteruskan ke service.
  - Form tidak valid menghasilkan 400 `invalid_request`; `duration_days=abc` tidak menyebabkan 500.
  - Pemetaan setiap error domain ke status dan kode di tabel spec; error tak dikenal menghasilkan 500 `internal_error` tanpa pesan asli.
  - Bentuk JSON detail persis contoh di spec, termasuk `"appeal": null`, `"evidence": []` untuk viewer, dan `allowed_actions` selalu array.
  - Unduh bukti mengirim `Content-Type` dari data tersimpan dan `Content-Disposition: attachment`.
  - Request tanpa header identitas menghasilkan 401.
- [ ] **Step 2: Tulis test log**: jalankan request create dengan logger ke buffer, lalu pastikan buffer tidak berisi `summary` dan nama file.
- [ ] **Step 3: Jalankan, pastikan gagal**

Run: `cd backend && go test ./internal/adapter/httpapi/...`
Expected: FAIL.

- [ ] **Step 4: Implementasi.** Batasi body create dengan `http.MaxBytesReader` (5 × 10 MB + 1 MB untuk field). Jenis file diperiksa dengan `http.DetectContentType` pada 512 byte pertama. Wiring di `cmd/api/main.go` memakai clock nyata dan `filestore` dengan `EVIDENCE_DIR`.
- [ ] **Step 5: Jalankan semua test, coverage, lint**

Run: `cd backend && make test-race && make test-integration && make cover && make lint`
Expected: PASS, coverage sesuai standar, lint bersih.

- [ ] **Step 6: Commit**

```bash
git add backend/ docker-compose.yml
git commit -m "feat(backend): endpoint suspend kurir dan identitas pengguna sementara"
```

---

### Task F1: Identitas pengguna, client, tipe, dan hook

**Files:**
- Create: `web/src/features/auth/actor-context.tsx`, `web/src/features/auth/ActorSwitcher.tsx`, `ActorSwitcher.test.tsx`
- Modify: `web/src/api/client.ts`, `client.test.ts` (header identitas, `apiPost`, `apiPostForm`, `apiDownloadURL`)
- Create: `web/src/api/suspension-types.ts`, `web/src/api/suspensions.ts`, `suspensions.test.ts`
- Modify: `web/src/app/Layout.tsx` (pasang `ActorSwitcher`), `web/src/app/providers.tsx`
- Create: `web/src/test/suspension-fixtures.ts`

**Interfaces:**
- Produces:
  - `ActorProvider`, `useActor()` mengembalikan `{ actor: { id: string; role: Role } | null, setActor }`. Disimpan di `sessionStorage` (dibungkus try/catch). Ini satu-satunya tempat yang perlu diganti saat SSO tersedia.
  - Client mengirim `X-User-ID` dan `X-User-Role` di semua request bila actor ada. Respons 401 dilempar sebagai `ApiError` biasa.
  - Tipe camelCase: `Suspension`, `SuspensionDetail`, `SuspensionStatus`, `AppealStatus`, `FraudType`, `SuspensionAction`, `SuspensionFilter`, beserta DTO snake_case dan fungsi pemetaan.
  - Hook: `useSuspensions(filter)`, `useSuspension(id)`, `useCreateSuspension()`, `useApproveSuspension(id)`, `useRejectSuspension(id)`, `useSubmitAppeal(id)`, `useDecideAppeal(id)`. Mutasi meng-invalidate `['suspensions']` dan `['suspension', id]`.
  - Label Indonesia: `fraudTypeLabel`, `suspensionStatusLabel`, `appealStatusLabel` di `web/src/features/suspensions/labels.ts`.
  - Unduh bukti memakai `fetch` lewat client (supaya header identitas ikut), lalu membuat object URL; bukan tautan langsung.

- [ ] **Step 1: Tulis test**: client mengirim kedua header; `apiPostForm` mengirim `FormData` tanpa mengatur `Content-Type` manual; pemetaan DTO ke camelCase untuk detail lengkap (dengan `appeal: null` dan dengan banding); `ActorSwitcher` menyimpan pilihan dan memulihkannya setelah render ulang.
- [ ] **Step 2: Jalankan, pastikan gagal**

Run: `cd web && npm test`
Expected: FAIL.

- [ ] **Step 3: Implementasi.**
- [ ] **Step 4: Jalankan**

Run: `cd web && npm run typecheck && npm run lint && npm run test:coverage`
Expected: PASS, coverage `api/` dan `lib/` minimal 80%.

- [ ] **Step 5: Commit**

```bash
git add web/
git commit -m "feat(web): identitas pengguna sementara dan client API suspend kurir"
```

---

### Task F2: Halaman daftar

**Files:**
- Create: `web/src/pages/suspensions/SuspensionListPage.tsx`, `SuspensionListPage.test.tsx`
- Create: `web/src/features/suspensions/SuspensionTable.tsx`, `SuspensionStatusBadge.tsx`
- Modify: `web/src/app/router.tsx`, `web/src/app/Layout.tsx` (menu "Suspend Kurir")
- Modify: `web/src/lib/date.ts`, `date.test.ts` (`formatDateRange`, misalnya `20–22 Nov 2026`)

**Interfaces:**
- Consumes: hook dan tipe dari F1, `QueryState`, `DataTable`, `Pagination` yang sudah ada.

- [ ] **Step 1: Tulis test halaman** dengan MSW: default memanggil `status=ACTIVE`; ganti filter status dan isi cabang mengirim parameter yang sesuai dan kembali ke halaman 1; state loading, error (menampilkan `error.message`), dan kosong; kolom Periode `20–22 Nov 2026`; tombol "Ajukan suspend" hanya untuk investigator; klik baris pindah ke `/suspensions/12`.
- [ ] **Step 2: Jalankan, pastikan gagal, implementasi, lalu jalankan**

Run: `cd web && npm run typecheck && npm run lint && npm test`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add web/
git commit -m "feat(web): halaman daftar suspend kurir"
```

---

### Task F3: Halaman form pengajuan

**Files:**
- Create: `web/src/pages/suspensions/SuspensionCreatePage.tsx`, `SuspensionCreatePage.test.tsx`
- Create: `web/src/features/suspensions/EvidenceInput.tsx`
- Modify: `web/src/app/router.tsx`

**Interfaces:**
- Consumes: `useCreateSuspension` dari F1.

- [ ] **Step 1: Tulis test**: durasi berupa 4 radio (3, 7, 14, 30 hari) tanpa input bebas; `start_date` default hari ini dan `min` hari ini; peringatan data pribadi tampil di dekat `summary`; validasi sisi klien untuk jumlah, ukuran, dan jenis file (pesan per file); tombol kirim nonaktif selama mengirim; error backend (misalnya `courier_has_open_case`) tampil dengan `error.message`; sukses pindah ke halaman detail; non-investigator yang membuka URL ini melihat pesan "Hanya investigator yang bisa mengajukan suspend".
- [ ] **Step 2: Jalankan, pastikan gagal, implementasi, lalu jalankan**

Run: `cd web && npm run typecheck && npm run lint && npm test`
Expected: PASS.

- [ ] **Step 3: Commit**

```bash
git add web/
git commit -m "feat(web): form pengajuan suspend kurir"
```

---

### Task F4: Halaman detail dan aksi

**Files:**
- Create: `web/src/pages/suspensions/SuspensionDetailPage.tsx`, `SuspensionDetailPage.test.tsx`
- Create: `web/src/features/suspensions/SuspensionActions.tsx`, `NoteDialog.tsx`, `SuspensionHistory.tsx`, `EvidenceList.tsx`
- Modify: `web/src/app/router.tsx`
- Modify: `web/src/lib/date.ts`, `date.test.ts` (`formatDateTimeWIB`, misalnya `19 Nov 2026 10.15 WIB`)

**Interfaces:**
- Consumes: hook mutasi dari F1, `allowed_actions` dari backend.

- [ ] **Step 1: Tulis test**:
  - Tombol yang tampil sama persis dengan `allowed_actions` untuk setiap kombinasi di fixture (`[]`, `[APPROVE, REJECT]`, `[SUBMIT_APPEAL]`, `[DECIDE_APPEAL]`).
  - Reject, banding, dan keputusan banding membuka dialog yang tidak bisa dikirim tanpa catatan; approve langsung dengan konfirmasi.
  - Setelah aksi sukses, data detail dimuat ulang dan riwayat bertambah; 409 `invalid_status_transition` menampilkan pesan dan memuat ulang data.
  - Daftar bukti tidak tampil saat `evidence` kosong; waktu riwayat ditampilkan dalam WIB.
  - State loading, error, dan 404 (`not_found`).
- [ ] **Step 2: Jalankan, pastikan gagal, implementasi, lalu jalankan**

Run: `cd web && npm run typecheck && npm run lint && npm run test:coverage && npm run build`
Expected: PASS, coverage keseluruhan minimal 60%.

- [ ] **Step 3: Commit**

```bash
git add web/
git commit -m "feat(web): halaman detail, approval, dan banding suspend kurir"
```

---

### Task I1: Integrasi lokal

**Files:** tidak ada perubahan kode. Ketidakcocokan diperbaiki lewat PR terpisah ke sisi yang salah; kontrak hanya diubah atas persetujuan kedua sisi.

- [ ] **Step 1:** `cp .env.example .env`, lalu `docker compose up --build`. Pastikan tabel `couriers` berisi minimal dua kurir contoh (sintetis).
- [ ] **Step 2: Jalankan alur lengkap** di `http://localhost:5173/suspensions`:
  - Masuk sebagai `inv-01` (investigator), ajukan suspend 3 hari mulai hari ini dengan satu PDF.
  - Masuk sebagai `inv-01` dengan peran approver: tombol Setujui tidak tampil.
  - Masuk sebagai `apv-01` (approver), setujui. Kasus muncul di daftar default "Sedang disuspend".
  - Masuk sebagai `inv-01`, ajukan banding. Masuk sebagai `apv-01`, terima banding. Status menjadi `CANCELLED`.
  - Masuk sebagai `kc-01` (viewer): detail terlihat tanpa daftar bukti.
  - Ajukan kasus kedua untuk kurir dengan kasus terbuka: pesan `courier_has_open_case`.
- [ ] **Step 3:** Periksa `docker compose logs backend`: tidak ada isi `summary`, catatan, atau nama file.
- [ ] **Step 4:** Catat hasil di deskripsi PR terakhir yang di-merge, dengan screenshot setiap langkah.
