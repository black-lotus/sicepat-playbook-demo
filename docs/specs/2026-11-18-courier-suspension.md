# Spec: Suspend kurir karena fraud

## Tujuan
Investigasi fraud kurir saat ini dilakukan manual dan hasilnya tidak tercatat di satu tempat. Fitur ini
mencatat hasil investigasi beserta buktinya, meminta persetujuan sebelum suspend berlaku, dan menampilkan
daftar kurir yang sedang disuspend supaya kepala cabang tidak memberi paket ke kurir tersebut.

Sanksi yang tersedia hanya suspend dengan durasi tetap: 3, 7, 14, atau 30 hari.

## Di luar scope
- Hubungan dengan KPI. Suspend tidak memengaruhi angka performa, dan halaman performa tidak menampilkan status suspend.
- Blokir otomatis di aplikasi SiGesit atau sistem penugasan paket. Penerapan suspend dilakukan kepala cabang
  berdasarkan daftar di dashboard (keputusan "opsi a").
- Deteksi fraud otomatis atau daftar kandidat investigasi.
- Matriks durasi berdasarkan jenis fraud dan jumlah pelanggaran. Durasi dipilih bebas dari empat pilihan.
- Mengubah durasi atau tanggal kasus yang sudah diajukan. Kasus yang salah ditolak lalu diajukan ulang.
- Banding dengan hasil "durasi dikurangi". Hasil banding hanya diterima (dibatalkan) atau ditolak.
- Notifikasi (email, push, WhatsApp) ke kurir, investigator, atau approver.
- Autentikasi dan SSO (lihat "Identitas pengguna").

## Peran

| Peran | Bisa melakukan |
|---|---|
| `investigator` | Mengajukan kasus, mengajukan banding atas nama kurir, melihat daftar dan detail, mengunduh bukti |
| `approver` | Menyetujui atau menolak kasus, memutuskan banding, melihat daftar dan detail, mengunduh bukti |
| `viewer` | Kepala cabang dan Ops: melihat daftar dan detail. **Tidak bisa** mengunduh bukti |

Aturan pemisahan tugas: pengguna yang mengajukan kasus **tidak boleh** menyetujui atau menolak kasus itu.

### Identitas pengguna (sementara)
Aplikasi belum punya autentikasi. Sampai SSO tersedia, identitas dikirim lewat header
`X-User-ID` (wajib, string) dan `X-User-Role` (`investigator`, `approver`, atau `viewer`). Frontend
menyediakan pemilih "Masuk sebagai" di header aplikasi untuk development.

Header ini bisa dipalsukan siapa pun yang bisa memanggil API. **Fitur ini tidak boleh dipakai di production
sebelum identitas diambil dari SSO.** Penggantiannya cukup di satu middleware backend dan satu modul frontend.

## Data

### Kasus suspend
| Field | Keterangan |
|---|---|
| `courier_id` | Wajib ada di tabel `couriers`. Nama dan cabang kurir diambil dari sana dan disimpan saat pengajuan |
| `fraud_type` | `FAKE_POD` (bukti kirim palsu), `PACKAGE_MISAPPROPRIATION` (penggelapan paket), `COD_MISAPPROPRIATION` (penggelapan dana COD), `STATUS_MANIPULATION` (manipulasi status pengiriman), `OTHER` |
| `summary` | Ringkasan hasil investigasi, 20 sampai 2000 karakter |
| `duration_days` | Salah satu dari `3`, `7`, `14`, `30` |
| `start_date` | Tanggal WIB hari pertama suspend. Saat pengajuan dan saat approval tidak boleh sebelum hari ini (WIB) |
| `end_date` | Dihitung backend: hari terakhir suspend (inklusif) = `start_date + duration_days - 1` |
| Bukti | 1 sampai 5 file PDF, JPG, atau PNG, masing-masing maksimal 10 MB |

Contoh: suspend 3 hari mulai 20 Nov berarti kurir disuspend 20, 21, dan 22 Nov (`end_date` 22 Nov), lalu
boleh bertugas lagi 23 Nov.

### Status
Backend menyimpan status keputusan, lalu menghitung status yang ditampilkan berdasarkan hari ini (WIB):

| Status di API | Kondisi |
|---|---|
| `PENDING_APPROVAL` | Diajukan, belum diputuskan |
| `REJECTED` | Ditolak approver |
| `SCHEDULED` | Disetujui, hari ini < `start_date` |
| `ACTIVE` | Disetujui, `start_date` <= hari ini <= `end_date` |
| `COMPLETED` | Disetujui, hari ini > `end_date` |
| `CANCELLED` | Banding diterima |

```
PENDING_APPROVAL ──approve──▶ SCHEDULED ──(start_date)──▶ ACTIVE ──(lewat end_date)──▶ COMPLETED
       │                          │                         │
       └──reject──▶ REJECTED      └─────banding diterima────┴──▶ CANCELLED
```

Tidak ada worker atau cron: status `SCHEDULED`, `ACTIVE`, dan `COMPLETED` dihitung setiap kali data dibaca.

### Banding
- Diajukan oleh `investigator` atas nama kurir, karena kurir tidak memakai dashboard.
- Hanya boleh saat status `SCHEDULED` atau `ACTIVE`, dan hanya **satu kali** per kasus.
- Selama banding diproses, suspend tetap berjalan.
- Diputuskan oleh `approver`: `ACCEPTED` membuat status menjadi `CANCELLED` (kurir boleh bertugas lagi),
  `REJECTED` membuat suspend tetap berjalan sampai `end_date`.
- `appeal_status`: `null` (belum ada banding), `PENDING`, `ACCEPTED`, atau `REJECTED`.

### Aturan lain
- Satu kurir hanya boleh punya satu kasus terbuka (`PENDING_APPROVAL`, `SCHEDULED`, atau `ACTIVE`).
- Penolakan kasus dan keputusan banding wajib disertai catatan (5 sampai 1000 karakter). Approval boleh tanpa catatan.
- Setiap pengajuan, keputusan, dan banding dicatat di riwayat kasus: aksi, pelaku, peran, catatan, waktu.
  Riwayat hanya bisa ditambah, tidak bisa diubah atau dihapus.

### Data pribadi
- Bukti investigasi hampir pasti berisi data pribadi penerima. Bukti hanya bisa diunduh `investigator` dan
  `approver`, tidak pernah ditampilkan inline, dan tidak tercantum di daftar.
- Form menampilkan peringatan agar `summary` dan catatan tidak berisi NIK, nomor HP, atau alamat lengkap.
- Backend tidak me-log isi `summary`, catatan, alasan banding, atau isi file.
- Jenis file diperiksa dari isi file, bukan dari ekstensi.

## Kontrak API
Mengikuti konvensi di `CLAUDE.md`: field snake_case, sukses `{data, meta}`, error `{error:{code,message}}`.
Semua request wajib membawa `X-User-ID` dan `X-User-Role`. Waktu kejadian (`*_at`) dikirim dalam UTC (RFC 3339);
`start_date` dan `end_date` adalah tanggal WIB (`YYYY-MM-DD`).

| Method dan path | Peran | Isi |
|---|---|---|
| `GET /api/v1/suspensions` | semua | Daftar kasus. Query: `status`, `branch_id`, `courier_id`, `page`, `page_size` (default 20, maks 100). Urut `submitted_at` terbaru |
| `POST /api/v1/suspensions` | investigator | `multipart/form-data`: `courier_id`, `fraud_type`, `summary`, `duration_days`, `start_date`, `evidence` (1 sampai 5 file). Respons 201 berisi detail |
| `GET /api/v1/suspensions/{id}` | semua | Detail kasus |
| `POST /api/v1/suspensions/{id}/approve` | approver | JSON `{"note": "..."}` (opsional) |
| `POST /api/v1/suspensions/{id}/reject` | approver | JSON `{"note": "..."}` (wajib) |
| `POST /api/v1/suspensions/{id}/appeal` | investigator | JSON `{"reason": "..."}` (wajib, 5 sampai 1000 karakter) |
| `POST /api/v1/suspensions/{id}/appeal/decision` | approver | JSON `{"decision": "ACCEPTED", "note": "..."}` |
| `GET /api/v1/suspensions/{id}/evidence/{evidence_id}` | investigator, approver | File asli, `Content-Disposition: attachment` |

Semua aksi `POST` mengembalikan 200 dengan detail kasus terbaru.

Item daftar:
```json
{
  "id": 12,
  "courier_id": "C001",
  "courier_name": "Budi Santoso",
  "branch_id": "JKT01",
  "fraud_type": "FAKE_POD",
  "duration_days": 3,
  "start_date": "2026-11-20",
  "end_date": "2026-11-22",
  "status": "ACTIVE",
  "appeal_status": null,
  "submitted_by": "inv-01",
  "submitted_at": "2026-11-19T03:15:00Z"
}
```

Detail = item daftar ditambah:
```json
{
  "summary": "POD 14 paket memakai foto yang sama ...",
  "decision": { "by": "apv-01", "at": "2026-11-19T07:00:00Z", "note": null },
  "appeal": null,
  "evidence": [
    { "id": 31, "file_name": "investigasi.pdf", "content_type": "application/pdf", "size_bytes": 482133 }
  ],
  "history": [
    { "action": "SUBMITTED", "actor_id": "inv-01", "actor_role": "investigator", "note": null, "at": "2026-11-19T03:15:00Z" },
    { "action": "APPROVED", "actor_id": "apv-01", "actor_role": "approver", "note": null, "at": "2026-11-19T07:00:00Z" }
  ],
  "allowed_actions": ["SUBMIT_APPEAL"]
}
```
- `decision` bernilai `null` selama `PENDING_APPROVAL`. `appeal` bernilai `null` sampai banding diajukan, lalu
  berisi `reason`, `submitted_by`, `submitted_at`, `status`, `decided_by`, `decided_at`, `note`.
- `history[].action`: `SUBMITTED`, `APPROVED`, `REJECTED`, `APPEAL_SUBMITTED`, `APPEAL_ACCEPTED`, `APPEAL_REJECTED`. Urut waktu menaik.
- `allowed_actions` dihitung backend untuk pengguna yang memanggil: `APPROVE`, `REJECT`, `SUBMIT_APPEAL`,
  `DECIDE_APPEAL`. Frontend menampilkan tombol hanya berdasarkan daftar ini.
- `evidence` dikirim kosong (`[]`) untuk `viewer`.
- Daftar kosong dikirim sebagai `"data": []`.

Error:
| Kondisi | Status | `error.code` |
|---|---|---|
| Header identitas kosong atau peran tidak dikenal | 401 | `unauthenticated` |
| Peran tidak boleh melakukan aksi ini | 403 | `forbidden` |
| Pengaju kasus mencoba approve atau reject | 403 | `self_approval_not_allowed` |
| Kasus atau file bukti tidak ada | 404 | `not_found` |
| Field form tidak valid (jenis, durasi, panjang teks, tanggal) | 400 | `invalid_request` |
| File bukti: jumlah, ukuran, atau jenis tidak valid | 400 | `invalid_evidence` |
| `start_date` sebelum hari ini (WIB), saat pengajuan atau approval | 400 | `start_date_passed` |
| `courier_id` tidak ada | 422 | `courier_not_found` |
| Kurir masih punya kasus terbuka | 409 | `courier_has_open_case` |
| Aksi tidak sesuai status saat ini, termasuk dua approver menekan bersamaan | 409 | `invalid_status_transition` |
| Banding sudah pernah diajukan | 409 | `appeal_already_submitted` |

## Halaman
- **`/suspensions`**: daftar kasus. Filter status (default `ACTIVE`, judul "Sedang disuspend") dan cabang.
  Kolom: Kurir, Cabang, Jenis fraud, Periode (`20–22 Nov 2026`), Durasi, Status, Banding. Klik baris membuka detail.
  Tombol "Ajukan suspend" hanya tampil untuk `investigator`.
- **`/suspensions/new`**: form pengajuan. Durasi dipilih lewat empat pilihan tetap. `start_date` default hari ini.
  Setelah berhasil, pindah ke halaman detail.
- **`/suspensions/:id`**: data kasus, periode, daftar bukti (tautan unduh hanya untuk peran yang boleh), riwayat,
  dan tombol aksi dari `allowed_actions`. Aksi yang butuh catatan atau alasan membuka dialog dengan field wajib.
- Semua halaman punya state loading, error (menampilkan `error.message`), dan data kosong.

## Acceptance criteria
- AC1: Investigator bisa mengajukan kasus dengan 1 sampai 5 bukti. Kasus muncul dengan status `PENDING_APPROVAL`.
- AC2: Pengajuan ditolak dengan kode error yang sesuai tabel untuk: durasi selain 3/7/14/30, `start_date` kemarin, kurir tidak ada, kurir dengan kasus terbuka, 0 atau 6 file, file 10 MB + 1 byte, dan file `.exe` yang diganti nama menjadi `.pdf`.
- AC3: Approver yang bukan pengaju bisa approve atau reject. Pengaju yang juga punya peran approver mendapat 403 `self_approval_not_allowed`. Reject tanpa catatan ditolak.
- AC4: Kasus 3 hari mulai 20 Nov yang disetujui berstatus `SCHEDULED` pada 19 Nov 23:59 WIB, `ACTIVE` pada 20 Nov 00:00 WIB sampai 22 Nov 23:59 WIB, dan `COMPLETED` pada 23 Nov 00:00 WIB.
- AC5: Banding bisa diajukan sekali saat `SCHEDULED` atau `ACTIVE`. Banding diterima mengubah status menjadi `CANCELLED`; banding ditolak tidak mengubah periode.
- AC6: Dua approve bersamaan pada kasus yang sama menghasilkan satu sukses dan satu 409 `invalid_status_transition`, dengan satu entri `APPROVED` di riwayat.
- AC7: `viewer` melihat daftar dan detail tanpa tautan bukti, dan mendapat 403 saat memanggil endpoint unduh bukti.
- AC8: Halaman `/suspensions` secara default menampilkan kasus `ACTIVE`, dan bisa difilter per cabang.
- AC9: Tombol aksi di halaman detail sama persis dengan `allowed_actions` dari backend.
- AC10: Setiap aksi menambah satu entri riwayat dengan pelaku, peran, catatan, dan waktu yang benar.
- AC11: Log backend tidak berisi `summary`, catatan, alasan banding, atau nama file bukti.

## Pertanyaan terbuka
1. Kapan identitas bisa diambil dari SSO, dan peran apa saja yang tersedia di sana? Sampai terjawab, fitur ini tidak boleh dipakai di production.
2. Penyimpanan bukti di production (object storage mana) dan berapa lama bukti disimpan.
3. Apakah daftar jenis fraud sudah sesuai dengan kategori yang dipakai tim investigasi?
4. Perlukan banding diputuskan oleh approver yang berbeda dari approver kasus? Untuk versi awal tidak diwajibkan.
