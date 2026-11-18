---
name: pull-request
description: Cek pull request sebelum meminta review, memakai checklist tim yang sama setiap kali, lalu menyusun deskripsi PR. Pakai saat akan membuka PR, saat diminta "siapkan PR", "cek PR", "PR siap direview?", atau sebelum meminta review silang ke pasangan.
---

# Pull request

Checklist ini dipakai sebelum meminta review. Jalankan setiap kali, dengan urutan yang sama. PR baru boleh dibuka atau diminta review bila semua butir lolos, atau bila user menyetujui pengecualian secara eksplisit setelah melihat hasilnya. Push dan pembukaan PR tetap menunggu konfirmasi user.

## Masukan

- Branch PR dan base-nya (biasanya `main`).
- Spec dan plan yang dilayani PR (`docs/specs/`, `docs/plans/`).
- Aturan PR di `CLAUDE.md` bagian Commit dan pull request.

## Checklist

### 1. Deskripsi menjelaskan apa dan kenapa
Susun deskripsi dengan bagian berikut:
- **Tujuan:** apa yang berubah dan kenapa. Rujuk task di plan dan acceptance criteria di spec.
- **Perubahan:** daftar singkat per area atau file penting.
- **Cara mengetes:** perintah yang dijalankan dan hasilnya.
- **Dampak ke metrik atau dashboard.**
- **Screenshot** untuk perubahan UI. Beri keterangan "data contoh" bila memakai data seed.
- **Keputusan yang perlu dicek reviewer**, bila ada.

Deskripsi yang hanya mengulang judul commit dianggap gagal.

### 2. Test baru atau yang diubah ikut di PR
- Jalankan `git diff --stat <base>...HEAD` dan pastikan setiap perubahan perilaku punya test di diff yang sama.
- Cocokkan dengan tabel rencana test di plan: setiap acceptance criteria yang disentuh PR harus menunjuk ke test yang ada di PR atau sudah ada di base.
- Untuk perubahan penting, buktikan test-nya mengunci perilaku: cabut perubahannya sementara, pastikan test gagal, lalu pulihkan.

### 3. Lint dan test lolos
Jalankan sesuai bagian yang disentuh:
- **Frontend (`web/`):** `npm run test:coverage`, `npm run lint`, `npm run typecheck`.
- **Backend (`backend/`):**
  - `make test-race`
  - integration test, ke database uji (bukan database berisi data penting, karena test men-TRUNCATE tabel)
  - `make cover`
  - `make lint`

Laporkan hasil nyata (jumlah test, coverage, jumlah temuan lint). Temuan lint yang sudah ada di base harus disebut terpisah dari temuan baru. PR tidak boleh menambah temuan baru.

### 4. Ukuran perubahan cukup kecil untuk direview
- Hitung baris yang berubah di luar lockfile, mock, dan kode hasil generate. Contohnya: `git diff --numstat <base>...HEAD`, lalu buang `package-lock.json` dan `mocks/`.
- Batasnya sekitar 400 baris dan satu perubahan logis per PR (`CLAUDE.md`).
- Bila lewat batas, usulkan pemecahan PR (misalnya per commit atau per task plan), atau tulis alasannya di deskripsi.

### 5. Kamu sudah membaca seluruh diff sendiri
- Butir ini untuk manusia pemilik PR, bukan untuk agent. Agent tidak boleh mencentangnya sendiri.
- Tampilkan ringkasan file yang berubah. Lalu tanyakan langsung ke user apakah seluruh diff sudah dibaca. Bila belum, butir ini gagal.

## Keluaran

Laporkan dalam tabel, satu baris per butir:

| Butir | Status | Bukti |
|---|---|---|
| 1. Deskripsi apa dan kenapa | lolos / gagal | draf deskripsi terlampir |
| 2. Test ikut di PR | lolos / gagal | test yang menutup setiap AC yang disentuh |
| 3. Lint dan test lolos | lolos / gagal | perintah dan hasilnya |
| 4. Ukuran cukup kecil | lolos / gagal | jumlah baris di luar lockfile dan mock |
| 5. Diff sudah dibaca sendiri | menunggu konfirmasi user | - |

Sertakan draf deskripsi PR lengkap di bawah tabel.

## Review silang

Saat mereview PR pasangan lain, beri minimal satu komentar yang mengubah keputusan: perilaku, desain, risiko, atau test. Komentar soal gaya penulisan saja tidak dihitung.
