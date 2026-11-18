---
name: review-plan
description: Review implementation plan sebelum agent mulai mengeksekusi, memakai checklist tim yang sama setiap kali. Pakai saat plan baru ditulis atau diubah, saat diminta "review plan", "cek plan", "plan siap dieksekusi?", atau tepat sebelum menjalankan plan lewat agent atau workflow. Hasilnya keputusan siap atau belum, dengan bukti per butir.
---

# Review plan

Checklist ini dipakai sebelum agent mulai mengeksekusi plan. Jalankan setiap kali, dengan urutan yang sama. Jangan mulai eksekusi bila ada butir yang gagal, kecuali user menyetujuinya secara eksplisit setelah melihat hasilnya.

## Masukan

- Plan yang direview, biasanya di `docs/plans/`.
- Spec yang menjadi acuannya (tertulis di header plan, biasanya di `docs/specs/`).
- `CLAUDE.md` untuk konvensi tim.

Baca spec dan plan langsung dari disk, lengkap. Jangan mengandalkan ringkasan atau ingatan dari percakapan.

## Checklist

### 1. Langkah sesuai spec, tidak menambah fitur
- Petakan setiap task di plan ke bagian spec atau acceptance criteria yang dilayaninya.
- Task, endpoint, field, layar, atau perilaku yang tidak punya dasar di spec dianggap penambahan fitur. Tandai sebagai gagal, kecuali plan menulis alasannya dan user menyetujuinya.
- Sebaliknya, cek juga bagian spec yang tidak dilayani task mana pun.

### 2. File yang akan disentuh masuk akal
- Setiap task menyebut file yang dibuat, diubah, atau dihapus.
- File sesuai struktur dan konvensi di `CLAUDE.md` (misalnya FE hanya di `web/`, BE hanya di `backend/`, kontrak API di `backend/api/openapi.yaml`).
- Task yang berjalan paralel tidak menyentuh file yang sama. Kalau menyentuh, plan harus mengatur urutannya atau isolasi worktree.
- Tidak ada file di luar scope fitur tanpa alasan.

### 3. Ada rencana test untuk setiap kriteria selesai
- Setiap acceptance criteria di spec punya minimal satu test otomatis di plan, dengan nama atau deskripsi yang jelas, dan task yang menulisnya.
- Kriteria yang tidak bisa diuji otomatis (misalnya tampilan di lebar layar tertentu) punya cek manual yang konkret: langkah, alat, dan hasil yang diharapkan.
- Test harus bisa gagal bila perilakunya rusak. Assertion yang selalu benar tidak dihitung.

### 4. Tidak ada perubahan berisiko tanpa alasan
Perubahan berisiko antara lain:
- mengubah perilaku yang sudah ada di `main`,
- migrasi atau perubahan data, termasuk data lama,
- kontrak API,
- zona waktu dan batas tanggal,
- autentikasi,
- konkurensi,
- dependency baru,
- operasi yang tidak bisa dibatalkan.

Untuk setiap perubahan berisiko, plan harus menulis:
- risikonya,
- alasan tetap dilakukan,
- mitigasinya, termasuk test, backfill, atau rollback,
- dan siapa yang menyetujui bila menyentuh bagian tim lain.

## Keluaran

Laporkan dalam tabel, satu baris per butir:

| Butir | Status | Bukti | Yang harus diperbaiki |
|---|---|---|---|
| 1. Langkah sesuai spec | lolos / gagal | task dan bagian spec yang dicek | ... |
| 2. File masuk akal | lolos / gagal | ... | ... |
| 3. Rencana test per kriteria | lolos / gagal | AC yang belum punya test | ... |
| 4. Perubahan berisiko | lolos / gagal | perubahan dan alasan atau mitigasinya | ... |

Tutup dengan satu kalimat keputusan: **siap dieksekusi** atau **belum siap**, beserta butir yang menahannya. Bila belum siap, tawarkan untuk memperbaiki plan dulu.
