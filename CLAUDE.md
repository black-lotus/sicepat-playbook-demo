# Konvensi tim

## Tentang aplikasi
Aplikasi ini adalah dashboard untuk memantau performa SiGesit (kurir SiCepat): jumlah paket terkirim,
tingkat keberhasilan delivery, ketepatan waktu (SLA), paket gagal kirim beserta alasannya, dan
produktivitas per kurir, per cabang, dan per periode.

Repo ini hanya berisi frontend. Data diambil dari REST API backend (Go) yang dikelola terpisah.

## Arsitektur singkat
- Bahasa: TypeScript (mode `strict`). Versi Node dan dependensi mengikuti `package.json`.
- Stack: React, Vite, React Router, TanStack Query untuk data fetching dan cache.
- Struktur folder di `src/`:
  - `app/` berisi entry, router, dan provider global.
  - `pages/` berisi satu folder per halaman (`courier-performance`, `branch-summary`, `failed-delivery`).
  - `features/` berisi logika per fitur: komponen, hook, dan tipe yang khusus untuk fitur itu.
  - `components/` berisi komponen UI generik yang dipakai lintas fitur (tabel, kartu KPI, filter tanggal).
  - `api/` berisi client HTTP, tipe response, dan hook query (`useCourierMetrics`).
  - `lib/` berisi helper murni, misalnya format angka, persen, dan tanggal.
- Alur data: halaman memanggil hook query di `api/`, hook memanggil client HTTP, komponen hanya menerima data lewat props.
- Base URL API dibaca dari environment variable `VITE_API_BASE_URL`.

## Konvensi kode
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

## Test
Framework dan cara menjalankannya:
- Vitest dengan React Testing Library. Response API di-mock dengan MSW.
- Unit dan komponen test dijalankan dengan `npm test`.
- Lint dan cek tipe dijalankan dengan `npm run lint` dan `npm run typecheck` sebelum push.

Standar cakupan minimal:
- 80% untuk `lib/` dan `api/`.
- 60% untuk seluruh project. Cek dengan `npm run test:coverage`.
- Setiap halaman diuji untuk state loading, error, data kosong, dan data terisi.

## Commit dan pull request
Format pesan commit:
- Conventional Commits: `<type>(<scope>): <ringkasan>`, misalnya `feat(courier-performance): tambah filter cabang`.
- Type yang dipakai: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `perf`, `style`.
- Sertakan ID task ClickUp di body commit bila ada.

Ukuran PR yang wajar:
- Maksimal sekitar 400 baris perubahan, di luar file hasil generate dan lockfile.
- Satu PR untuk satu perubahan logis, disertai test.
- Deskripsi PR berisi tujuan perubahan, cara mengetesnya, dan screenshot untuk perubahan UI.
