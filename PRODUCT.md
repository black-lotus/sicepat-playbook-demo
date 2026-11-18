# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users
Supervisor cabang SiCepat. Mereka membuka dashboard di dua situasi: di laptop kantor cabang, dan dari
ponsel saat berada di hub atau gudang. Karena itu dashboard harus nyaman dipakai di desktop maupun
ponsel.

## Product Purpose
Dashboard performa SiGesit (kurir SiCepat) per cabang dan per periode. Tujuan utama yang harus langsung
terasa saat dibuka: ranking dan apresiasi kurir, yaitu siapa yang performanya terbaik sehingga layak
diapresiasi, dengan data yang cukup untuk melihat siapa yang perlu dibantu.

Berhasil berarti supervisor dalam beberapa detik tahu siapa kurir terbaik di cabangnya untuk periode itu
dan kenapa, serta bisa menelusuri tren harian seorang kurir.

## Operating Context
- Supervisor memilih rentang tanggal (default 7 hari terakhir, maksimal 31 hari) dan cabang (opsional); data
  diambil otomatis saat filter berubah.
- Metrik per kurir: total paket, terkirim, gagal, tepat SLA, keberhasilan (`success_rate`), tepat SLA
  (`sla_rate`), posisi ranking, dan selisih dari pemimpin.
- Tepat SLA memakai cut-off 16:00 WIB: paket yang ditugaskan sebelum 16:00 harus terkirim hari itu
  sebelum 23:59:59, paket mulai 16:00 harus terkirim besok sebelum 16:00.
- Hari laporan memakai tanggal WIB; data disimpan dalam UTC dan backend yang menghitung batas harinya.
- Detail aturan dan kontrak API: `docs/specs/2026-11-18-courier-performance.md`.

## Capabilities and Constraints
- Stack: backend Go (hexagonal, PostgreSQL, Redis, Kafka), frontend React, TypeScript strict, Vite,
  React Router, TanStack Query (lihat `CLAUDE.md`).
- Data hanya dari REST API backend. Frontend tidak menghitung ulang KPI, hanya memformat.
- Kontrak API memberi total dan tren harian (`daily`) per kurir. Lihat spec untuk definisinya.
- Belum ada: daftar cabang dari API, autentikasi, target KPI per cabang. Alasan gagal punya halaman
  sendiri (Gagal Kirim).

## Brand Commitments
Belum ada aset atau aturan brand yang ditetapkan untuk dashboard ini (logo, warna, tipografi).
Jangan mengklaim identitas visual resmi SiCepat tanpa konfirmasi.

## Evidence on Hand
Belum ada data produksi. Data contoh dari repository in-memory backend bersifat sintetis dan harus
diberi label bila tampil di demo atau screenshot.

## Product Principles
- Apresiasi dulu, tanpa mempermalukan: ranking menonjolkan yang terbaik; kurir dengan angka rendah
  terlihat untuk dibantu, bukan dipajang.
- Angka harus bisa dipercaya: setiap angka berasal dari API dan bisa ditelusuri ke definisinya.
- Sekilas di ponsel, mendalam di laptop: jawaban utama terbaca di layar kecil tanpa zoom.

## Accessibility & Inclusion
- Tidak menampilkan atau me-log nomor HP, NIK, atau alamat kurir dan penerima.
- Bahasa antarmuka: Indonesia. Angka dan persen memakai format `id-ID`.
