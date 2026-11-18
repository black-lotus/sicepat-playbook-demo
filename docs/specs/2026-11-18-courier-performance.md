# Spec: Laporan performa kurir per cabang

## Tujuan
Supervisor cabang bisa melihat performa setiap kurir di cabangnya untuk rentang tanggal tertentu:
berapa paket yang ditugaskan, terkirim, gagal, dan seberapa banyak yang tepat SLA.

## Di luar scope
- Penyimpanan PostgreSQL, Kafka, Redis, dan worker agregasi. BE memakai repository in-memory berisi data contoh.
- Autentikasi, pagination, export, grafik.
- Daftar cabang dari API. FE memakai input teks untuk `branch_id`.

## Definisi
Satu `Delivery` adalah satu paket yang ditugaskan ke satu kurir.

| Field | Tipe | Keterangan |
|---|---|---|
| `AWB` | string | nomor resi |
| `CourierID`, `CourierName` | string | kurir yang ditugaskan |
| `BranchID` | string | cabang kurir |
| `Status` | `DELIVERED`, `FAILED`, `ON_PROCESS` | status terakhir |
| `AssignedAt` | waktu | kapan paket ditugaskan, dipakai untuk filter periode |
| `DeliveredAt` | waktu, boleh kosong | terisi hanya bila `DELIVERED` |
| `SLADeadline` | waktu | batas waktu kirim |

Perhitungan per kurir, dari delivery yang `AssignedAt`-nya ada di dalam periode:
- `assigned` = jumlah semua delivery (termasuk `ON_PROCESS`)
- `delivered` = jumlah `DELIVERED`
- `failed` = jumlah `FAILED`
- `on_time` = jumlah `DELIVERED` dengan `DeliveredAt <= SLADeadline`
- `success_rate` = `delivered / assigned`, 0 bila `assigned` = 0
- `on_time_rate` = `on_time / delivered`, 0 bila `delivered` = 0
- Kedua rate dibulatkan 4 angka desimal (0.916666 menjadi 0.9167)

Periode memakai zona waktu Asia/Jakarta (WIB). `from` dan `to` inklusif:
periode = `[from 00:00 WIB, to+1 hari 00:00 WIB)`.

## Kontrak API

`GET /api/v1/courier-performance?branch_id=JKT01&from=2026-11-01&to=2026-11-07`

Validasi, dicek berurutan:
| Kondisi | Status | `error.code` |
|---|---|---|
| `branch_id` kosong | 400 | `MISSING_BRANCH_ID` |
| `from` atau `to` kosong atau bukan format `YYYY-MM-DD` | 400 | `INVALID_DATE` |
| `from` > `to`, atau periode lebih dari 31 hari (inklusif) | 400 | `INVALID_DATE_RANGE` |

Respons 200:
```json
{
  "branch_id": "JKT01",
  "from": "2026-11-01",
  "to": "2026-11-07",
  "items": [
    {
      "courier_id": "C001",
      "courier_name": "Budi Santoso",
      "assigned": 40,
      "delivered": 36,
      "failed": 3,
      "on_time": 33,
      "success_rate": 0.9,
      "on_time_rate": 0.9167
    }
  ]
}
```
- `items` diurutkan `success_rate` menurun, lalu `courier_id` menaik.
- Tidak ada data: `items` berisi `[]`, bukan `null`.

Respons 400:
```json
{ "error": { "code": "INVALID_DATE_RANGE", "message": "Periode maksimal 31 hari dan from tidak boleh setelah to." } }
```

## Acceptance criteria
- AC1: Request valid mengembalikan 200 dengan satu item per kurir di cabang itu, angka sesuai perhitungan di atas.
- AC2: Kurir dengan `assigned` 0 atau `delivered` 0 mendapat rate 0, bukan error.
- AC3: Paket yang ditugaskan 2026-11-07 23:30 WIB masuk ke `to=2026-11-07`; paket 2026-11-08 00:00 WIB tidak.
- AC4: Setiap kondisi validasi mengembalikan 400 dengan `error.code` yang sesuai tabel.
- AC5: Cabang tanpa data mengembalikan 200 dengan `items: []`.
- AC6: Halaman `/courier-performance` punya input cabang, tanggal awal, tanggal akhir, dan tombol Tampilkan. Request baru dikirim saat tombol ditekan.
- AC7: Halaman menampilkan tabel kolom: Kurir, Ditugaskan, Terkirim, Gagal, Success rate, Tepat SLA. Rate ditampilkan sebagai persen format Indonesia satu desimal (0.9167 menjadi `91,7%`).
- AC8: Halaman menampilkan state loading, pesan error dari `error.message` API, dan pesan "Belum ada data untuk periode ini." saat `items` kosong.
