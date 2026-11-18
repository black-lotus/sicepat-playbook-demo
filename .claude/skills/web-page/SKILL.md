---
name: web-page
description: Membuat atau mengisi halaman dashboard di frontend (web/), berurutan dari tipe dan hook query di api/, komponen di features/, halaman di pages/, sampai route, lengkap dengan test MSW untuk state loading, error, kosong, dan terisi. Pakai saat diminta "buat halaman", "isi halaman ringkasan cabang / gagal kirim", "tampilkan endpoint X di dashboard", atau mengerjakan task frontend di plan yang menambah halaman atau hook baru.
---

# Halaman dashboard

Halaman dibangun dari lapisan data ke tampilan: `api/` → `features/` → `pages/` → `app/router.tsx`. Halaman memanggil hook di `api/`, dan komponen hanya menerima data lewat props. Semua perintah dijalankan dari folder `web/`.

## Masukan

- Kontrak endpoint di `backend/api/openapi.yaml`. Frontend hanya memakai field yang ada di kontrak.
- Spec, desain, dan task plan yang dilayani (`docs/specs/`, `docs/design/`, `docs/plans/`), terutama acceptance criteria, teks state kosong, dan data contoh.
- Konvensi frontend di `CLAUDE.md` bagian **Frontend**.

**Berhenti dan tanya user bila field yang dibutuhkan tampilan tidak ada di kontrak.** Jangan menghitung KPI sendiri di frontend untuk menutupi kekurangan itu. Angka KPI datang dari API, frontend hanya memformat.

## Contoh rujukan di repo

Baca dulu sebelum menulis, lalu ikuti polanya. Jangan mengubah file ini kecuali task memintanya.

| Lapisan | File contoh |
|---|---|
| Client HTTP dan `ApiError` | `src/api/client.ts` |
| DTO dan tipe domain | `src/api/types.ts` |
| Fetch, mapper, query key, hook | `src/api/courier-performance.ts` dan test-nya |
| State loading, error, kosong | `src/components/QueryState.tsx` |
| Komponen generik | `src/components/DataTable.tsx`, `DateRangeFilter.tsx`, `Pagination.tsx` |
| Hook state filter | `src/features/courier-performance/use-courier-performance-filter.ts` |
| Halaman dan test-nya | `src/pages/courier-performance/CourierPerformancePage.tsx`, `.test.tsx` |
| Helper format | `src/lib/format-percent.ts`, `format-number.ts`, `date.ts` |
| Infrastruktur test | `src/test/server.ts`, `render.tsx`, `fixtures.ts` |
| Route dan navigasi | `src/app/router.tsx`, `Layout.tsx`, `router.test.tsx` |

Placeholder yang menunggu diisi: `src/pages/branch-summary/BranchSummaryPage.tsx` dan `src/pages/failed-delivery/FailedDeliveryPage.tsx`.

## Langkah

### 1. Tipe (`src/api/types.ts`)
- Tambah `XxxDTO` yang sama persis dengan JSON backend (snake_case). DTO hanya dipakai di dalam `src/api/`.
- Tambah tipe domain camelCase (`Xxx`) dan tipe filter (`XxxFilter`) yang dipakai komponen. Akronim pada tipe dan konstanta ditulis kapital (`AWB`, `SLA`), field memakai camelCase (`courierID`, `awbNumber`, `slaRate`).
- Pakai ulang `PageMetaDTO`, `PageMeta`, dan `Paginated<T>`.

### 2. Fetch dan hook (`src/api/<nama-endpoint>.ts`)
- `toXxx(dto)` memetakan DTO ke tipe domain tanpa menghitung angka baru.
- `fetchXxx(filter, signal)` memanggil `apiGet<XxxDTO[], PageMetaDTO>('/api/v1/...', {snake_case params}, signal)`. Jangan memanggil `fetch` langsung.
- `xxxKey(filter)` mengembalikan query key `['<nama>', filter] as const`.
- `useXxx(filter)` memakai `useQuery` dengan `signal`. Untuk list berhalaman, tambah `placeholderData: keepPreviousData`.
- Data server hanya disimpan di cache TanStack Query, bukan di state global atau context.
- Test (`<nama-endpoint>.test.ts`) dengan MSW `server.use(http.get(...))`:
  - parameter query terkirim dengan benar (snake_case, parameter kosong tidak dikirim)
  - mapping ke camelCase
  - response error menjadi `ApiError` dengan `status` dan `code`
  - query key stabil

### 3. Fixture (`src/test/fixtures.ts`)
- Tambah `xxxDTOs` berisi data sintetis. Bila spec punya tabel data contoh, samakan fixture dengan tabel itu dan beri komentar rujukannya.
- Jangan memakai nama, nomor HP, NIK, atau alamat asli.

### 4. Komponen fitur (`src/features/<nama-fitur>/`)
- Komponen (`PascalCase.tsx`) menerima data dan callback lewat props, tanpa memanggil hook query.
- Hook state lokal (filter, pilihan baris) diletakkan di file kebab-case `use-xxx.ts`. Mengubah filter mengembalikan `page` ke 1.
- Format angka hanya lewat helper di `src/lib/` (`formatPercent`, `formatNumber`, tanggal). Bila butuh helper baru, buat file kebab-case di `lib/` beserta test-nya.
- Komponen yang dipakai lintas fitur dipindah ke `src/components/`. `components/` tidak boleh meng-import dari `features/` atau `pages/`.
- Data pribadi kurir dan penerima tidak ditampilkan mentah dan tidak di-`console.log`.
- Tabel, tombol, dan dialog harus bisa diakses: tabel diberi `aria-label`, tombol diberi nama yang jelas. Test memakai query berbasis role.

### 5. Halaman (`src/pages/<nama-halaman>/XxxPage.tsx`)
- Susunannya: hook filter → `useXxx(filter)` → `<QueryState isPending error isEmpty emptyMessage onRetry>` → komponen fitur.
- Keempat state wajib ada. Teks state kosong diambil dari spec bila ada.
- Halaman yang bergantung pada tanggal menerima prop opsional `today?: Date` supaya test deterministik.
- Pagination memakai `<Pagination>` dengan `meta` dari API.

### 6. Route dan navigasi (`src/app/`)
- Daftarkan path kebab-case di `routes` (`router.tsx`). Bila halaman baru, tambah `NavLink` di `Layout.tsx`. Bila mengganti placeholder, route-nya sudah ada.
- Perbarui `router.test.tsx`: tambah handler MSW untuk endpoint baru di `beforeEach`, dan pastikan heading halaman muncul di path tersebut.

### 7. Test halaman (`XxxPage.test.tsx`)
Render dengan `renderWithQuery` (tanpa retry) dan mock API lewat `server.use`. Kasus wajib:
- **loading:** handler `await delay('infinite')`, lalu `findByRole('status')` berisi "Memuat data".
- **error:** handler 500 `{error:{code,message}}` dengan `{ once: true }` diikuti handler sukses. Pastikan `role="alert"` muncul, klik "Coba lagi", lalu data tampil.
- **kosong:** `data: []`, lalu teks state kosong tampil.
- **terisi:** angka tampil dalam format dari `lib/` sesuai fixture.
- Interaksi utama: ganti filter mengirim parameter baru dan kembali ke halaman 1, pindah halaman, buka detail.

### 8. Verifikasi
Jalankan dan laporkan hasil nyatanya:
- `npm test`
- `npm run test:coverage` (minimal 80% untuk `lib/` dan `api/`, 60% untuk seluruh frontend)
- `npm run lint`
- `npm run typecheck` (tanpa `any` dan `@ts-ignore`; pakai `unknown` lalu persempit tipenya)
- Cek visual bila stack jalan (`docker compose up`): buka `http://localhost:5173/<path>` dan periksa keempat state serta layar sempit (≤ 640px). Ambil screenshot untuk deskripsi PR dan beri keterangan "data contoh".

## Keluaran

Laporkan dalam tabel, satu baris per lapisan:

| Lapisan | File | Test | Status |
|---|---|---|---|
| Tipe dan DTO | `api/types.ts` | typecheck | lolos / gagal |
| Fetch dan hook | `api/xxx.ts` | nama test | ... |
| Fixture | `test/fixtures.ts` | - | ... |
| Komponen fitur | `features/...` | ... | ... |
| Halaman (4 state) | `pages/...` | loading / error / kosong / terisi | ... |
| Route dan nav | `app/...` | `router.test.tsx` | ... |

Di bawah tabel, cantumkan hasil coverage, lint, dan typecheck, serta field kontrak yang belum ada atau tidak cocok. Kekurangan kontrak diusulkan ke user untuk disepakati dengan backend, bukan diakali di frontend.

Commit memakai `feat(web): ...`. Commit dan push tetap menunggu permintaan user.
