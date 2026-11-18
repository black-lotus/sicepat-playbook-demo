---
name: stacked-pr
description: Memecah perubahan besar menjadi PR bertumpuk (stacked) berukuran kecil, membukanya dengan base berantai, me-merge-nya dengan aman ke main, dan memulihkan bila PR ter-merge ke branch perantara. Pakai saat sebuah branch melewati sekitar 400 baris, saat diminta "pecah PR", "stacked PR", "PR bertumpuk", "merge tumpukan PR", atau saat isi PR yang sudah di-merge ternyata tidak ada di main.
---

# Stacked PR

PR bertumpuk membuat setiap PR kecil (sekitar 400 baris atau kurang, `CLAUDE.md`), tapi punya satu jebakan: bila base PR berikutnya tidak dipindah ke `main` sebelum di-merge, isinya masuk ke branch perantara, bukan ke `main`. Skill ini menutup jebakan itu di setiap tahap.

Push, pembukaan PR, dan merge tetap menunggu konfirmasi user.

## 1. Pecah

1. Simpan sumber: `git branch backup/<nama> <branch-sumber>`, catat SHA tip-nya.
2. Tentukan langkah logis, satu perubahan per langkah. Test ikut di langkah yang sama dengan kode yang diujinya. Perbaikan hasil review digabung ke langkah pemilik kodenya, bukan commit terpisah.
3. Bangun ulang dari base: satu branch per langkah, tepat satu commit di atas branch sebelumnya. Nama: `<prefix>-01-<topik>`, `<prefix>-02-<topik>`, dan seterusnya.
4. Ukuran tiap commit: `git show --numstat --format= <sha> | grep -v -E 'package-lock|/mocks/' | awk '{a+=$1+$2} END {print a}'`. Lebih dari sekitar 400 baris: pecah lagi.
5. Setiap langkah harus hijau sendiri (FE: `npm run test:coverage`, `npm run lint`, `npm run typecheck`; BE: `make test-race`, integration test ke DB uji, `make lint`). Langkah antara boleh berisi versi sementara yang masuk akal (misalnya stub), dan wajib disebut di deskripsi PR.
6. Kriteria akhir: `git diff backup/<nama> <tip-baru>` harus kosong.

## 2. Buka PR

- Base berantai: PR pertama ke `main`, PR berikutnya ke branch PR sebelumnya. `gh pr create --base <base> --head <branch>`.
- Deskripsi mengikuti skill `pull-request`. Tambah bagian **Urutan merge** di setiap PR:
  > Bagian dari tumpukan X, urutan N dari M. Merge dari bawah ke atas. Setelah PR di bawahnya di-merge, ubah base PR ini ke `main` sebelum merge.
- Rujuk PR lain dengan nomor GitHub (`#12`), bukan nomor urut internal.

## 3. Merge dengan aman

Untuk setiap PR, dari bawah ke atas:

1. Pastikan base-nya `main`: `gh pr view <n> --json baseRefName`. Bila belum, `gh pr edit <n> --base main`.
2. Merge PR itu.
3. Verifikasi isinya benar-benar di `main`: `git fetch origin && git diff --quiet <sha-commit-PR> origin/main -- <folder>`. Kosong berarti aman.
4. Lanjut ke PR berikutnya: ulangi dari langkah 1.

Cara mencegah sejak awal: pemilik repo mengaktifkan **Settings, General, Automatically delete head branches**. Setelah branch PR di bawahnya terhapus, GitHub otomatis memindahkan base PR berikutnya ke `main`.

## 4. Pulihkan bila ter-merge ke branch perantara

Gejala: semua PR berstatus merged, tapi fiturnya tidak ada di `main`.

1. Daftar base: `gh pr list --state merged --json number,baseRefName,headRefName`.
2. Cari branch yang memuat hasil akhir: untuk setiap branch tumpukan yang masih ada, `git diff --quiet <tip-akhir-yang-direview> origin/<branch> -- <folder>`. Yang kosong adalah branch berisi isi lengkap.
3. Buka satu PR susulan dari branch itu ke `main`. Deskripsi: isinya gabungan PR #a sampai #b yang sudah direview, tanpa perubahan tambahan, plus checklist bahwa diff sama dengan tip yang direview dan test lolos.
4. Setelah di-merge, ulangi verifikasi langkah 3 di bagian Merge dengan aman.

## Keluaran

Laporkan tabel: nomor PR, branch, base, jumlah baris, status (draft, terbuka, merged), dan untuk tahap merge, hasil verifikasi isi di `main` per PR.
