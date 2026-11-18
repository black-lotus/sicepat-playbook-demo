# Spec: Gagal Kirim, Ringkasan Cabang, dan Target KPI per cabang

## Tujuan
Melengkapi dua halaman yang masih placeholder dan menambah target KPI:
- **Gagal Kirim:** supervisor tahu berapa paket gagal, alasan terbanyak, kurir dengan gagal terbanyak, dan
  apakah kondisinya membaik dibanding periode sebelumnya.
- **Ringkasan Cabang:** manajer membandingkan cabang dalam satu periode dan melihat cabang mana yang
  memenuhi target.
- **Target KPI per cabang:** setiap cabang punya target bulanan untuk keberhasilan, tepat SLA, dan volume
  paket terkirim. Target dipakai Ringkasan Cabang dan halaman Performa Kurir.

Mengikuti konvensi `CLAUDE.md` dan aturan spec `2026-11-18-courier-performance.md` (batas hari WIB,
cut-off SLA 16:00 WIB, rumus rate, respons `{data, meta}`, filter maksimal 31 hari).

## Di luar scope
- Form atau endpoint untuk mengubah target. Target hanya diisi lewat seed dan migrasi (belum ada autentikasi).
- Target per kurir, notifikasi, export.

## Data
- Sumber gagal kirim: `delivery_events` (status `failed`, `failed_reason`), periode berdasarkan `occurred_at`
  dalam hari WIB.
- Sumber Ringkasan Cabang: `courier_daily_metrics`, dikelompokkan per `branch_id` baris harian.
- Target: tabel baru `branch_kpi_targets`:

| Kolom | Tipe | Keterangan |
|---|---|---|
| `branch_id` | text | cabang |
| `month` | date | hari pertama bulan |
| `success_rate` | numeric(5,4) | target 0..1 |
| `sla_rate` | numeric(5,4) | target 0..1 |
| `delivered_packages` | integer | target paket terkirim sebulan, minimal 0 |

Primary key `(branch_id, month)`. Seed mengisi target 12 bulan 2026 untuk semua cabang seed kecuali `BDG01`
(menguji cabang tanpa target).

## Aturan target untuk sebuah periode
- **Target rate** (`success_rate`, `sla_rate`): target bulan yang memuat `end_date`.
- **Target volume periode:** jumlah dari setiap hari dalam periode, masing-masing `delivered_packages` bulan
  hari itu dibagi jumlah hari di bulan itu, dibulatkan ke bilangan bulat terdekat di akhir.
  Contoh: target November 3.000 dan periode 1 sampai 7 Nov, maka target volume = round(7 × 3.000 / 30) = 700.
- Bulan tanpa target menyumbang 0 ke target volume. Bila bulan `end_date` tidak punya target, objek `target`
  bernilai `null`.
- **Status terhadap target:** `met` bila nilai aktual sama atau di atas target, `below` bila di bawah.

## Kontrak API (didokumentasikan di `backend/api/openapi.yaml`)

### `GET /api/v1/failed-deliveries/summary?start_date&end_date&branch_id`
`branch_id` opsional (kosong = semua cabang). Validasi sama dengan `MetricFilter` tanpa pagination.

```json
{
  "data": {
    "total_packages": 1200,
    "total_failed": 180,
    "failed_rate": 0.15,
    "previous_total_failed": 200,
    "failed_change_rate": -0.1,
    "reasons": [
      { "reason": "Penerima tidak di tempat", "count": 60, "share": 0.3333333333 }
    ],
    "top_couriers": [
      { "rank": 1, "courier_id": "C002", "courier_name": "Sari Wulandari", "branch_id": "JKT01",
        "total_packages": 40, "failed": 30, "failed_rate": 0.75 }
    ]
  }
}
```
- Periode pembanding: rentang dengan panjang hari sama, tepat sebelum `start_date`.
- `failed_change_rate` = (`total_failed` - `previous_total_failed`) / `previous_total_failed`; `null` bila
  `previous_total_failed` 0.
- `reasons` urut `count` menurun lalu `reason` menaik; `share` = `count / total_failed`; jumlah `count` sama
  dengan `total_failed`. Tanpa gagal: `reasons` `[]`, `failed_rate` 0.
- `top_couriers`: maksimal 5, urut `failed` menurun lalu `courier_id` menaik, hanya kurir dengan `failed` > 0.

### `GET /api/v1/branch-performance?start_date&end_date&page&page_size`
```json
{
  "data": [
    {
      "rank": 1, "branch_id": "BDG02", "courier_count": 8,
      "total_packages": 1000, "delivered": 833, "failed": 167, "on_time": 700,
      "success_rate": 0.833, "sla_rate": 0.8403,
      "target": {
        "success_rate": 0.85, "sla_rate": 0.8, "delivered_packages": 900,
        "success_rate_status": "below", "sla_rate_status": "met", "delivered_status": "below"
      },
      "daily": [ { "date": "2026-11-01", "total_packages": 140, "delivered": 120, "failed": 20, "on_time": 101 } ]
    }
  ],
  "meta": { "page": 1, "page_size": 20, "total": 8 }
}
```
- Urut `success_rate` menurun lalu `branch_id` menaik; `rank` global. Cabang tanpa data di periode tidak muncul.
- `courier_count` = jumlah kurir berbeda dengan baris metrik di periode.
- `daily` satu entri per hari WIB, hari kosong 0, jumlah sama dengan total.

### `GET /api/v1/branch-kpi-targets?branch_id&month`
`month` berformat `YYYY-MM`, keduanya wajib. 200 `{ "data": { "branch_id", "month", "success_rate", "sla_rate", "delivered_packages" } }`.
Tidak ada target: 404 `target_not_found`. Format salah: 400 `invalid_query`.

## Tampilan (mengikuti gaya redesign gelap di `feat/fe-redesign-dark`)
- **Gagal Kirim (`/failed-delivery`):** filter periode dan cabang; kartu hero total gagal dengan chip
  perubahan (naik = waspada, turun = membaik; dibedakan ikon panah dan teks, bukan hanya warna); donut alasan
  gagal dengan legenda persen; daftar 5 kurir dengan gagal terbanyak.
- **Ringkasan Cabang (`/branch-summary`):** filter periode; daftar kartu cabang terurut berisi rank, cabang,
  keberhasilan, tepat SLA, paket terkirim, jumlah kurir, strip harian, dan tiga penanda target (tercapai atau
  di bawah, dengan ikon dan teks). Klik kartu membuka `/courier-performance` dengan cabang dan periode yang sama
  lewat query URL.
- **Performa Kurir:** bila cabang dipilih dan target ada, tampil garis target keberhasilan dan tepat SLA cabang;
  setiap kurir diberi penanda di atas atau di bawah target keberhasilan cabang. Halaman membaca `branch_id`,
  `start_date`, `end_date` dari query URL bila ada.

## Acceptance criteria
- G1: Ringkasan gagal kirim menghitung `total_failed`, `failed_rate`, `reasons`, dan `top_couriers` sesuai aturan, dengan batas hari WIB.
- G2: Jumlah `count` semua alasan sama dengan `total_failed`; tanpa gagal menghasilkan `reasons: []` dan rate 0.
- G3: Periode pembanding dan `failed_change_rate` benar; `null` bila periode lalu tanpa gagal.
- G4: Halaman Gagal Kirim menampilkan hero, donut, dan daftar kurir, plus state loading, error, kosong.
- R1: Ranking cabang urut dan `rank` global benar lintas halaman; `daily` lengkap per hari.
- R2: Target rate memakai bulan `end_date`; target volume diprorata per hari; status `met`/`below` benar; cabang tanpa target menghasilkan `target: null`.
- R3: Halaman Ringkasan Cabang menampilkan kartu cabang dan penanda target, dan klik kartu membuka Performa Kurir dengan filter yang sama.
- T1: Migrasi membuat `branch_kpi_targets` dengan constraint rate 0..1 dan volume minimal 0; seed mengisi target 2026 untuk semua cabang kecuali BDG01.
- T2: `GET /branch-kpi-targets` mengembalikan target, 404 `target_not_found`, atau 400 `invalid_query`.
- T3: Halaman Performa Kurir menampilkan target cabang dan penanda per kurir; tanpa target, penanda tidak tampil.
- X1: Semua tampilan baru tanpa scroll horizontal di 390px, kontras AA, status dibedakan bentuk dan teks.
