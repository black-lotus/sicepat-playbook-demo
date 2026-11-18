-- Entry seed data contoh (sintetis) untuk pengembangan. Idempoten dan deterministik: tanpa random,
-- dua kali jalan menghasilkan isi tabel yang sama. Seluruhnya dalam satu transaksi.
--
-- Jalankan lewat `make seed` (dari backend/) atau service `seed` di docker-compose.yml. Berkas 01 sampai 04
-- berbagi temp table di transaksi ini, jadi dijalankan lewat berkas ini, bukan sendiri-sendiri.
-- Tidak ada data pribadi (nomor HP, NIK, alamat): kurir dan paket semuanya sintetis.
BEGIN;

\ir 01_couriers.sql
\ir 02_delivery_events.sql
\ir 03_courier_daily_metrics.sql
\ir 04_branch_kpi_targets.sql

COMMIT;
