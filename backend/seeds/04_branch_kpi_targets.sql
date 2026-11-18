-- branch_kpi_targets: target KPI bulanan 2026 (12 bulan) untuk semua cabang seed kecuali BDG01, yang sengaja
-- tanpa target (menguji cabang tanpa target). Deterministik: rumus tanpa random, jadi dua kali jalan
-- menghasilkan isi tabel yang sama. Dijalankan lewat seed.sql (butuh temp table seed_branch).
--
-- Target dasar per cabang dipilih sebagian di atas dan sebagian di bawah angka aktual seed supaya penanda
-- met dan below sama-sama muncul. Rate naik 0,5 poin persen per bulan dari Januari (maksimal 0,99);
-- volume bulanan = dasar x (0,90 + 0,01 x nomor bulan), dibulatkan.

DELETE FROM branch_kpi_targets
WHERE branch_id IN (SELECT branch_id FROM seed_branch)
  AND month BETWEEN DATE '2026-01-01' AND DATE '2026-12-01';

INSERT INTO branch_kpi_targets (branch_id, month, success_rate, sla_rate, delivered_packages)
SELECT b.branch_id,
       m.month,
       LEAST(b.success_rate + 0.005 * (EXTRACT(MONTH FROM m.month) - 1), 0.99),
       LEAST(b.sla_rate     + 0.005 * (EXTRACT(MONTH FROM m.month) - 1), 0.99),
       ROUND(b.delivered_packages * (0.90 + 0.01 * EXTRACT(MONTH FROM m.month)))::int
FROM (VALUES
    --  cabang   success_rate  sla_rate  delivered_packages (dasar)
    ('JKT01', 0.7500::numeric, 0.9000::numeric,  300),
    ('JKT02', 0.7800::numeric, 0.8000::numeric, 4000),
    ('JKT03', 0.7000::numeric, 0.8500::numeric, 7000),
    ('SBY01', 0.8500::numeric, 0.8000::numeric,  320),
    ('MDN01', 0.6500::numeric, 0.9500::numeric,  280),
    ('BDG02', 0.8500::numeric, 0.8500::numeric,  900),
    ('SMG01', 0.7500::numeric, 0.8000::numeric, 1300),
    ('DPS01', 0.6000::numeric, 0.7000::numeric,  500)
) AS b(branch_id, success_rate, sla_rate, delivered_packages)
CROSS JOIN generate_series(DATE '2026-01-01', DATE '2026-12-01', INTERVAL '1 month') AS g(month_ts)
CROSS JOIN LATERAL (SELECT g.month_ts::date AS month) AS m
WHERE b.branch_id IN (SELECT branch_id FROM seed_branch)
ORDER BY b.branch_id, m.month;
