# Monitoring Performa SiGesit

Dashboard untuk memantau performa SiGesit (kurir SiCepat). Konvensi tim ada di [CLAUDE.md](CLAUDE.md).

- `backend/`: API, worker, dan consumer (Go).
- `web/`: dashboard frontend (React, TypeScript, Vite).

## Menjalankan secara lokal

Dengan Docker Compose (detail di bagian "Local development" di [CLAUDE.md](CLAUDE.md)):
```sh
cp .env.example .env
docker compose up --build   # lalu buka http://localhost:5173/courier-performance
```

Tanpa Docker untuk aplikasinya (infrastruktur tetap dari Compose):
```sh
docker compose up -d postgres redis kafka migrate
cd backend && make run-api        # juga: make run-worker, make run-consumer
cd web && nvm use && npm install && npm run dev
```
Bila port 8080 sudah dipakai, jalankan API dengan `HTTP_ADDR=:18080 make run-api`
dan frontend dengan `API_PROXY_TARGET=http://localhost:18080 npm run dev`.

## Data contoh

Seed pengembangan membuat data sintetis (tanpa data pribadi) yang sama di setiap mesin. `delivery_events`
adalah sumber kebenaran: satu event per paket untuk semua cabang, 1 Jan sampai 31 Des 2026, dengan
`due_at` sesuai cut-off SLA 16:00 WIB. `courier_daily_metrics` diisi dari agregasi event itu dengan logika
yang sama seperti worker (hari WIB). Alasan gagal memakai daftar tetap (Alamat tidak ditemukan, Penerima
tidak di tempat, Penerima menolak paket, Nomor telepon tidak aktif, Paket rusak, Cuaca buruk, Kendaraan
bermasalah) dengan sebaran berbeda per cabang.

```sh
cd backend && make seed   # butuh migrasi sudah jalan; `docker compose up` menjalankannya otomatis lewat service seed
```

Seed deterministik (tanpa random) dan idempoten: dijalankan berulang menghasilkan isi tabel yang sama.
Ia menghapus dulu data di ruang ID yang ia kelola (kurir di tabel di bawah, event dan metrik kurir itu pada 2026,
metrik 2026 di cabang tersebut) lalu mengisi ulang dalam satu transaksi. Data di luar itu tidak disentuh.
Berkas ada di `backend/seeds/` (`seed.sql` memanggil `01_couriers.sql`, `02_delivery_events.sql`,
`03_courier_daily_metrics.sql`).

| Cabang | Kurir | Skenario |
|---|---|---|
| `JKT01` | 4 (`C001` sampai `C004`) | 1 sampai 7 Nov 2026 persis tabel di spec ranking kurir. Hari lain bervariasi: Budi tinggi, Dewi tinggi, Agus sedang, Sari rendah |
| `JKT02` | 40 (`K201` sampai `K240`) | Tampilan padat; Okt 2026 31 entri harian per kurir |
| `JKT03` | 60 (`K301` sampai `K360`) | Pagination (`page_size` 50, halaman 2 berisi 10 kurir) |
| `SBY01` | 2 (`S001`, `S002`) | Podium kurang dari 3 |
| `MDN01` | 3 (`M101` sampai `M103`) | `M101` dan `M102` berangka identik, urutan ditentukan `courier_id` |
| `BDG01` | 0 | Tanpa data (state kosong) |
| `BDG02` | 8 | Cabang pembanding, performa tinggi |
| `SMG01` | 12 | Cabang pembanding, performa menengah |
| `DPS01` | 6 | Cabang pembanding, performa rendah |

## Test

```sh
cd backend && make test && make test-integration   # integration butuh PostgreSQL yang sudah dimigrasi
cd web && npm test && npm run lint && npm run typecheck
```
