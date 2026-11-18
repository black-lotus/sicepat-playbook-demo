# Spec: Ranking performa kurir

## Tujuan
Supervisor cabang bisa melihat ranking kurir untuk rentang tanggal tertentu: siapa yang terbaik
(untuk diapresiasi), angka totalnya, dan tren hariannya.

Spec ini mengikuti konvensi dan boilerplate yang sudah ada di `main` (lihat `CLAUDE.md`).
Desain FE: `docs/design/2026-11-18-courier-performance.md`.

## Di luar scope
- Breakdown alasan gagal (sudah ada halaman Gagal Kirim terpisah).
- Target KPI per cabang, autentikasi, export.
- Daftar cabang dari API. FE memakai input teks untuk `branch_id`.

## Data yang dipakai
Mengikuti boilerplate:
- `delivery_events`: satu baris per percobaan pengiriman, status `delivered` atau `failed`,
  dengan `occurred_at` dan `due_at` (batas SLA), disimpan dalam UTC.
- `courier_daily_metrics`: agregat harian per kurir, diisi worker.

Rumus per kurir (sudah ada di `domain.CourierMetric`):
- `total_packages` = jumlah event; `delivered`, `failed` = jumlah per status.
- `on_time` = jumlah `delivered` dengan `occurred_at <= due_at`.
- `success_rate` = `delivered / total_packages`, 0 bila `total_packages` 0.
- `sla_rate` = `on_time / delivered`, 0 bila `delivered` 0.
- Rate dikirim apa adanya (0..1), tanpa pembulatan. FE memformat.

## Aturan zona waktu (keputusan baru, berlaku untuk seluruh backend)
Penyimpanan tetap UTC. Batas hari dan batas SLA adalah aturan bisnis, jadi dihitung backend dalam WIB
(UTC+7, tanpa DST).

- **Hari laporan:** event masuk ke tanggal `occurred_at` dalam WIB. Hari D = `[D 00:00 WIB, D+1 00:00 WIB)`.
  Worker mengagregasi `courier_daily_metrics` per hari WIB. `start_date` dan `end_date` adalah tanggal WIB.
- **Batas SLA (cut-off 16:00 WIB):** dihitung dari waktu paket ditugaskan.
  - Ditugaskan sebelum 16:00 WIB pada hari D: batas = hari D 23:59:59 WIB.
  - Ditugaskan pukul 16:00 WIB atau lebih: batas = hari D+1 16:00:00 WIB.
  - Event Kafka mendapat field opsional `assigned_at`. Bila `due_at` kosong dan `assigned_at` ada,
    consumer mengisi `due_at` dengan aturan ini. Bila `due_at` dikirim sumber, nilainya dipakai apa adanya.

| Ditugaskan (WIB) | Batas SLA (WIB) |
|---|---|
| 7 Nov 09:00 | 7 Nov 23:59:59 |
| 7 Nov 15:59:59 | 7 Nov 23:59:59 |
| 7 Nov 16:00 | 8 Nov 16:00:00 |
| 7 Nov 21:00 | 8 Nov 16:00:00 |

`CLAUDE.md` bagian Konvensi API perlu diubah: kalimat "konversi zona waktu dilakukan di frontend"
diganti aturan di atas.

**Transisi data lama.** Baris `courier_daily_metrics` yang dibuat sebelum aturan ini memakai batas hari
UTC, dan worker hanya menghitung ulang hari ini dan kemarin. Agar riwayat tidak memakai dua aturan
sekaligus, semua hari sebelum perubahan dihitung ulang sekali dari `delivery_events` (sumber kebenaran)
dengan batas hari WIB. Penghitungan ulang satu hari selalu mengganti seluruh baris hari itu, sehingga
baris basi yang tidak lagi punya event ikut hilang. Prosesnya boleh dijalankan berulang dengan hasil
yang sama.

## Kontrak API

Endpoint baru, didokumentasikan di `backend/api/openapi.yaml`. Endpoint `/api/v1/courier-metrics`
yang sudah ada tidak berubah.

`GET /api/v1/courier-performance?start_date=2026-11-01&end_date=2026-11-07&branch_id=JKT01&page=1&page_size=50`

| Parameter | Wajib | Keterangan |
|---|---|---|
| `start_date`, `end_date` | ya | `YYYY-MM-DD`, tanggal WIB, inklusif, maksimal 31 hari |
| `branch_id` | tidak | kosong berarti semua cabang |
| `page` | tidak | default 1 |
| `page_size` | tidak | default 20, maksimal 100 |

Validasi memakai `MetricFilter` yang sudah ada: format salah menghasilkan 400 `invalid_query`,
periode atau pagination tidak valid menghasilkan 400 `invalid_filter`, error lain 500 `internal_error`.

Respons 200:
```json
{
  "data": [
    {
      "rank": 1,
      "courier_id": "C001",
      "courier_name": "Budi Santoso",
      "branch_id": "JKT01",
      "total_packages": 40,
      "delivered": 36,
      "failed": 4,
      "on_time": 33,
      "success_rate": 0.9,
      "sla_rate": 0.9166666666666666,
      "success_rate_gap": 0,
      "daily": [
        { "date": "2026-11-01", "total_packages": 6, "delivered": 6, "failed": 0, "on_time": 5 }
      ]
    }
  ],
  "meta": { "page": 1, "page_size": 50, "total": 4 }
}
```
- Urutan: `success_rate` menurun, lalu `courier_id` menaik. `rank` adalah posisi global (1 sampai
  `meta.total`), stabil di semua halaman.
- `success_rate_gap` = `success_rate` peringkat 1 dikurangi `success_rate` kurir ini (0 untuk peringkat 1),
  dihitung backend karena peringkat 1 bisa ada di halaman lain.
- `daily`: satu entri per tanggal WIB dari `start_date` sampai `end_date`, urut menaik, hari tanpa event
  bernilai 0. Jumlah tiap field sama dengan totalnya. Contoh di atas dipotong.
- Kurir yang tidak punya event di periode itu tidak muncul. Tidak ada data: `data` berisi `[]`.

## Data contoh
Data sintetis dimasukkan lewat seed pengembangan (bukan migrasi). Screenshot dan demo yang memakainya diberi keterangan "data contoh".

Cabang `JKT01`, 1 sampai 7 Nov 2026. Nilai `daily` ditulis `total/delivered/failed/on_time`:

| Rank | Kurir | Total t/d/f/o | `success_rate` | `sla_rate` | `daily` 1 sampai 7 Nov |
|---|---|---|---|---|---|
| 1 | C001 Budi Santoso | 40/36/4/33 | 0.9 | 0.91667 | 6/6/0/5, 5/5/0/5, 6/5/1/5, 6/5/1/4, 5/5/0/5, 6/5/1/5, 6/5/1/4 |
| 2 | C003 Dewi Lestari | 35/30/5/30 | 0.85714 | 1 | 5/5/0/5, 5/4/1/4, 5/4/1/4, 5/5/0/5, 5/4/1/4, 5/4/1/4, 5/4/1/4 |
| 3 | C004 Agus Pratama | 21/15/6/12 | 0.71429 | 0.8 | 3/3/0/3, 3/2/1/2, 3/2/1/1, 3/2/1/2, 3/2/1/1, 3/2/1/2, 3/2/1/1 |
| 4 | C002 Sari Wulandari | 4/0/4/0 | 0 | 0 | 1/0/1/0, 1/0/1/0, 0/0/0/0, 1/0/1/0, 0/0/0/0, 1/0/1/0, 0/0/0/0 |

Cabang `JKT02`: 40 kurir, 1 sampai 31 Okt 2026, untuk uji tampilan padat. Cabang `BDG01`: tanpa data.

## Acceptance criteria
- AC1: Request valid mengembalikan 200 `{data, meta}` dengan satu item per kurir yang punya event, angka sesuai rumus.
- AC2: Kurir dengan `total_packages` 0 di suatu hari atau `delivered` 0 mendapat rate 0, bukan error atau NaN.
- AC3: Event `occurred_at` 2026-11-01 05:00 WIB dan 2026-11-07 23:30 WIB terhitung untuk periode 1 sampai 7 Nov; event 2026-10-31 23:59 WIB dan 2026-11-08 00:00 WIB tidak.
- AC4: Parameter tidak valid menghasilkan 400 dengan `invalid_query` atau `invalid_filter`, pesan dari backend.
- AC5: Tidak ada data menghasilkan 200 dengan `data: []`.
- AC6: Batas SLA mengikuti cut-off 16:00 WIB: ditugaskan 7 Nov 09:00 dan terkirim 8 Nov 00:00 WIB tidak tepat SLA; ditugaskan 7 Nov 16:00 dan terkirim 8 Nov 16:00 WIB tepat SLA.
- AC7: `rank` dan `success_rate_gap` benar di halaman 2 (peringkat 1 ada di halaman 1).
- AC8: `daily` berisi tepat satu entri per hari dalam periode, jumlahnya sama dengan total.
- AC9: Halaman `/courier-performance` memakai filter yang sudah ada (periode default 7 hari terakhir, cabang opsional) dan mengambil data otomatis saat filter berubah.
- AC10: Halaman menampilkan **podium** (rank 1 sampai 3, hanya di halaman 1) dan **tabel hasil** dengan kolom Pos, Kurir, Harian, Total paket, Terkirim, Gagal, Keberhasilan, Tepat SLA, Selisih. Rate memakai `formatPercent` yang sudah ada (`90,0%`); Selisih dalam poin persen (`−4,3 pp`), rank 1 menampilkan `Pemimpin`.
- AC11: State loading, error (pesan dari API, tombol Coba lagi), dan kosong tetap ada.
- AC12: Klik baris atau kartu podium membuka panel detail kurir berisi tren harian. Panel ditutup dengan tombol Tutup atau Esc.
- AC13: Di lebar 390px halaman tidak scroll horizontal; podium menjadi carousel geser, tabel menjadi daftar peringkat.
- AC14: Peringkat bawah tampil netral (tanpa merah, tanpa label negatif). Tepat SLA dan telat dibedakan bentuk, bukan hanya warna.
