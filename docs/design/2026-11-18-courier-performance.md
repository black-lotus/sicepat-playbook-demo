# Desain: Dashboard Performa Kurir, "Hasil Lomba Lari"

Brief ini sudah disetujui. Spec dan kontrak: `docs/specs/2026-11-18-courier-performance.md`.
Konteks produk: `PRODUCT.md`. Kode dasar mengikuti boilerplate di `web/` (layout, router, client API,
`QueryState`, `DateRangeFilter`, `Pagination`, helper format).

## Pengguna dan tugas
- Supervisor cabang, di laptop kantor dan di HP saat di hub. Mode: dashboard kerja.
- Pertanyaan utama: siapa yang layak diapresiasi periode ini, dan kenapa.
- Berhasil bila dalam beberapa detik terlihat top 3 kurir dengan angkanya, lalu tren harian kurir
  mana pun bisa ditelusuri. Semua angka dari API.

## Arah visual
- **Gagasan:** performa dibaca seperti papan hasil marathon: urutan finis, selisih dari pemimpin,
  dan split harian. Yang ditolak: deretan kartu KPI generik di atas tabel abu-abu.
- **Palet:** top bar shell aspal gelap; area kerja putih kertas bib; satu biru "garis lintasan" untuk
  aksen dan seri data utama; emas, perak, perunggu hanya untuk rank 1 sampai 3. Mode terang adalah
  tampilan utama (dipakai di HP di bawah cahaya hub). Boilerplate sudah mengikuti
  `prefers-color-scheme: dark`, jadi token gelap tetap disediakan dengan peran warna yang sama.
- **Tipe:** satu superfamili. Barlow Condensed (tabular) untuk angka dan nomor posisi, Barlow untuk
  teks UI, di-host sendiri lewat paket `@fontsource`. Skala ukuran ketat, angka besar hanya di podium.
- **Gerak khas, "garis finis":** saat data baru tampil, baris meluncur ke posisi finisnya dan kolom
  Selisih muncul setelah baris berhenti. Mengikuti `prefers-reduced-motion`.
- **Disiplin:**
  - Status tidak hanya warna: tepat SLA digambar penuh, telat digambar berpola garis.
  - Strip harian semua kurir memakai skala dan lebar sel yang sama.
  - Satu warna untuk satu fungsi. Merah tidak dipakai untuk peringkat; merah hanya untuk error.

## Viewport pertama
**Desktop:**
- Top bar (`Layout` yang sudah ada, diberi gaya shell aspal): nama produk, navigasi halaman.
- Baris filter: Dari, Sampai (`DateRangeFilter`), Cabang. Data diambil otomatis saat filter berubah.
- Podium (halaman 1 saja): 3 kartu bib besar (nomor posisi, nama, keberhasilan monumental, tepat SLA,
  strip harian).
- Hasil lengkap: tabel Pos, Kurir, Harian, Total paket, Terkirim, Gagal, Keberhasilan, Tepat SLA,
  Selisih, lalu `Pagination`.
- Klik baris atau kartu membuka panel detail di kanan dengan grafik lintasan harian.

**HP (390px):**
- Filter tetap terlihat dalam satu baris ringkas yang bisa dibuka (sheet).
- Podium jadi carousel geser horizontal (scroll snap).
- Tabel jadi daftar peringkat: posisi, nama, keberhasilan, strip harian mini. Panel detail jadi sheet
  dari bawah.

## Strip harian
- Satu sel per hari, lebar sama, urut kiri ke kanan.
- Tinggi isi sel = `delivered / total_packages` hari itu (proporsi tampilan). Bagian tepat SLA diisi
  penuh biru; terkirim tapi telat diisi pola garis.
- Hari tanpa paket: sel kosong bergaris tipis.
- Periode 1 hari: strip disembunyikan.

## State dan rentang
- Memuat: skeleton berbentuk podium dan baris (lewat `QueryState`, tetap `role="status"`).
- Error: pesan dari API dan tombol Coba lagi. Kosong: "Belum ada data performa kurir untuk periode ini."
- Kurir per halaman: `page_size` 50. Kurang dari 3 kurir: podium menampilkan yang ada. Halaman 2 dan
  seterusnya: tanpa podium.
- Periode 1 sampai 31 hari.
- Peringkat bawah netral: tanpa merah, tanpa label negatif.

## Batasan
- Kontrol standar web (input, date, button, tabel, dialog). Tidak ada kostum lomba pada kontrol.
- Tidak ada scroll horizontal tingkat halaman di 390px. Kontras teks minimal WCAG AA di kedua tema.
- Tidak ada PII. Screenshot dan demo dari data seed diberi keterangan "data contoh".
- Grafik digambar sendiri dengan SVG dari data, tanpa library chart.
