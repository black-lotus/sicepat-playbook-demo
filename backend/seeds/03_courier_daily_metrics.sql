-- courier_daily_metrics diisi dari delivery_events (sumber kebenaran), bukan dari angka literal.
-- Logika agregasi sama dengan aggregateDailySQL di internal/adapter/postgres/metric_repository.go:
-- hari dihitung WIB, satu baris per kurir per hari, branch_id dari event terakhir hari itu, on_time =
-- delivered dengan occurred_at <= due_at. Bedanya hanya satu pernyataan untuk seluruh tahun, bukan satu
-- per hari, dan updated_at deterministik (akhir hari WIB) supaya isi tabel identik di tiap jalan.
-- Dijalankan lewat seed.sql.

-- Hapus metrik kurir yang dikelola seed dan seluruh metrik 2026 di cabang yang dikelola (termasuk sisa
-- uji atau seed lama), lalu isi ulang.
DELETE FROM courier_daily_metrics
WHERE courier_id IN (SELECT id FROM seed_courier)
   OR (branch_id IN (SELECT branch_id FROM seed_branch) AND date BETWEEN DATE '2026-01-01' AND DATE '2026-12-31');

INSERT INTO courier_daily_metrics (courier_id, branch_id, date, total_packages, delivered, failed, on_time, updated_at)
SELECT courier_id,
       (array_agg(branch_id ORDER BY occurred_at DESC, id DESC))[1],
       (occurred_at AT TIME ZONE 'Asia/Jakarta')::date,
       COUNT(*),
       COUNT(*) FILTER (WHERE status = 'delivered'),
       COUNT(*) FILTER (WHERE status = 'failed'),
       COUNT(*) FILTER (WHERE status = 'delivered' AND due_at IS NOT NULL AND occurred_at <= due_at),
       ((occurred_at AT TIME ZONE 'Asia/Jakarta')::date + 1)::timestamp AT TIME ZONE 'Asia/Jakarta'
FROM delivery_events
WHERE courier_id IN (SELECT id FROM seed_courier)
  AND occurred_at >= TIMESTAMP '2026-01-01 00:00:00' AT TIME ZONE 'Asia/Jakarta'
  AND occurred_at <  TIMESTAMP '2027-01-01 00:00:00' AT TIME ZONE 'Asia/Jakarta'
GROUP BY courier_id, (occurred_at AT TIME ZONE 'Asia/Jakarta')::date
ORDER BY courier_id, 3;
