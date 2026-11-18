-- Seed data contoh (sintetis) untuk dashboard performa kurir. Idempoten: aman dijalankan berulang.
-- Tidak ada data pribadi (nomor HP, NIK, alamat).
BEGIN;

-- Bersihkan semua metrik di cabang demo pada rentang demo, termasuk kurir lain yang tersisa dari uji atau seed lama.
DELETE FROM courier_daily_metrics
WHERE branch_id IN ('JKT01', 'JKT02', 'BDG01')
  AND date BETWEEN DATE '2026-10-01' AND DATE '2026-11-07';

DELETE FROM courier_daily_metrics
WHERE courier_id IN ('C001', 'C002', 'C003', 'C004')
   OR courier_id IN (SELECT 'K' || n FROM generate_series(201, 240) AS n);

DELETE FROM couriers
WHERE id IN ('C001', 'C002', 'C003', 'C004')
   OR id IN (SELECT 'K' || n FROM generate_series(201, 240) AS n);

-- JKT01: 4 kurir, 1 sampai 7 Nov 2026 (hari tanpa paket tidak disimpan, sama seperti worker).
INSERT INTO couriers (id, name, branch_id) VALUES
    ('C001', 'Budi Santoso', 'JKT01'),
    ('C002', 'Sari Wulandari', 'JKT01'),
    ('C003', 'Dewi Lestari', 'JKT01'),
    ('C004', 'Agus Pratama', 'JKT01');

INSERT INTO courier_daily_metrics (courier_id, branch_id, date, total_packages, delivered, failed, on_time)
SELECT v.courier_id, 'JKT01', DATE '2026-11-01' + (v.day - 1), v.total, v.delivered, v.failed, v.on_time
FROM (VALUES
    ('C001', 1, 6, 6, 0, 5), ('C001', 2, 5, 5, 0, 5), ('C001', 3, 6, 5, 1, 5), ('C001', 4, 6, 5, 1, 4),
    ('C001', 5, 5, 5, 0, 5), ('C001', 6, 6, 5, 1, 5), ('C001', 7, 6, 5, 1, 4),
    ('C003', 1, 5, 5, 0, 5), ('C003', 2, 5, 4, 1, 4), ('C003', 3, 5, 4, 1, 4), ('C003', 4, 5, 5, 0, 5),
    ('C003', 5, 5, 4, 1, 4), ('C003', 6, 5, 4, 1, 4), ('C003', 7, 5, 4, 1, 4),
    ('C004', 1, 3, 3, 0, 3), ('C004', 2, 3, 2, 1, 2), ('C004', 3, 3, 2, 1, 1), ('C004', 4, 3, 2, 1, 2),
    ('C004', 5, 3, 2, 1, 1), ('C004', 6, 3, 2, 1, 2), ('C004', 7, 3, 2, 1, 1),
    ('C002', 1, 1, 0, 1, 0), ('C002', 2, 1, 0, 1, 0), ('C002', 4, 1, 0, 1, 0), ('C002', 6, 1, 0, 1, 0)
) AS v(courier_id, day, total, delivered, failed, on_time);

-- JKT02: 40 kurir (K201 sampai K240), 1 sampai 31 Okt 2026, angka deterministik.
INSERT INTO couriers (id, name, branch_id)
SELECT 'K' || (200 + k), 'Kurir ' || lpad(k::text, 2, '0'), 'JKT02'
FROM generate_series(1, 40) AS k;

INSERT INTO courier_daily_metrics (courier_id, branch_id, date, total_packages, delivered, failed, on_time)
SELECT 'K' || (200 + k), 'JKT02', DATE '2026-10-01' + (d - 1),
       t, t - f, f, (t - f) - LEAST(t - f, (k + 2 * d) % 3)
FROM generate_series(1, 40) AS k
CROSS JOIN generate_series(1, 31) AS d
CROSS JOIN LATERAL (SELECT (k * 7 + d * 3) % 9 AS t) AS tt
CROSS JOIN LATERAL (SELECT LEAST(t, (k + d) % 3) AS f) AS ff
WHERE t > 0;

-- BDG01: sengaja tanpa data.
COMMIT;
