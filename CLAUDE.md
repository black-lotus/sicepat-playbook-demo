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
| `backend` | API Go dengan repository in-memory berisi data contoh | `8080` | `go run ./cmd/api` |
| `web` | Dev server Vite | `5173` | `npm run dev -- --host` |

Belum ada database, Kafka, atau Redis. Kalau nanti ditambahkan, tambahkan sebagai service baru di file yang sama.

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
- Bahasa: TypeScript (mode `strict`). Versi Node dan dependensi mengikuti `web/package.json`.
- Stack: React, Vite, React Router, TanStack Query untuk data fetching dan cache.
- Struktur folder di `web/src/`:
  - `app/` berisi entry, router, dan provider global.
  - `pages/` berisi satu folder per halaman (`courier-performance`, `branch-summary`, `failed-delivery`).
  - `features/` berisi logika per fitur: komponen, hook, dan tipe yang khusus untuk fitur itu.
  - `components/` berisi komponen UI generik yang dipakai lintas fitur (tabel, kartu KPI, filter tanggal).
  - `api/` berisi client HTTP, tipe response, dan hook query (`useCourierMetrics`).
  - `lib/` berisi helper murni, misalnya format angka, persen, dan tanggal.
- Alur data: halaman memanggil hook query di `api/`, hook memanggil client HTTP, komponen hanya menerima data lewat props.
- Base URL API dibaca dari environment variable `VITE_API_BASE_URL`.

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
<!-- TODO: diisi oleh tim backend -->

### Arsitektur singkat
Bahasa dan versi:
Modul utama dan alurnya:
Dependensi penting:

### Konvensi kode
Gaya penamaan:
Struktur folder:
Pola yang dihindari:

### Test
Framework dan cara menjalankannya:
Standar cakupan minimal:

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
