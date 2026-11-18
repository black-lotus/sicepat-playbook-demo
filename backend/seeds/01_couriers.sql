-- Kurir dan cabang yang dikelola seed. Ruang ID yang dikelola seed: id di seed_courier (tabel couriers)
-- dan cabang di seed_branch (termasuk BDG01 yang sengaja tanpa kurir dan tanpa data).
-- Dijalankan lewat seed.sql (butuh transaksi yang sama dengan 02 dan 03).

-- k = nomor urut kurir di cabangnya, dipakai sebagai pengubah deterministik pada rumus data.
CREATE TEMP TABLE seed_courier ON COMMIT DROP AS
SELECT id, name, branch_id, k FROM (
    -- JKT01: 4 kurir. Profil: Budi tinggi, Dewi tinggi, Agus sedang, Sari rendah.
    SELECT * FROM (VALUES
        ('C001', 'Budi Santoso',   'JKT01', 1),
        ('C002', 'Sari Wulandari', 'JKT01', 2),
        ('C003', 'Dewi Lestari',   'JKT01', 3),
        ('C004', 'Agus Pratama',   'JKT01', 4)
    ) AS v(id, name, branch_id, k)
    UNION ALL
    -- JKT02: 40 kurir (K201 sampai K240) untuk tampilan padat.
    SELECT 'K' || (200 + k), 'Kurir ' || lpad(k::text, 2, '0'), 'JKT02', k
    FROM generate_series(1, 40) AS k
    UNION ALL
    -- JKT03: 60 kurir (K301 sampai K360) untuk pagination (page_size 50).
    SELECT 'K' || (300 + k), 'Kurir JKT03 ' || lpad(k::text, 2, '0'), 'JKT03', k
    FROM generate_series(1, 60) AS k
    UNION ALL
    -- SBY01: 2 kurir (podium kurang dari 3).
    SELECT * FROM (VALUES
        ('S001', 'Rina Kusuma', 'SBY01', 1),
        ('S002', 'Joko Susilo', 'SBY01', 2)
    ) AS v(id, name, branch_id, k)
    UNION ALL
    -- MDN01: 3 kurir; M101 dan M102 berangka identik (tie-break courier_id).
    SELECT * FROM (VALUES
        ('M101', 'Andi Saputra', 'MDN01', 1),
        ('M102', 'Bima Prakoso', 'MDN01', 2),
        ('M103', 'Citra Ayu',    'MDN01', 3)
    ) AS v(id, name, branch_id, k)
    UNION ALL
    -- Cabang pembanding untuk Ringkasan Cabang: BDG02 (8 kurir), SMG01 (12 kurir), DPS01 (6 kurir).
    SELECT b.prefix || (100 + k), n.first_names[(k - 1) % 12 + 1] || ' ' || n.last_names[(k * 5) % 10 + 1], b.branch_id, k
    FROM (VALUES ('BDG02', 'B', 8), ('SMG01', 'G', 12), ('DPS01', 'P', 6)) AS b(branch_id, prefix, n)
    CROSS JOIN LATERAL generate_series(1, b.n) AS k
    CROSS JOIN (SELECT
        ARRAY['Rizky', 'Dian', 'Fajar', 'Putri', 'Eko', 'Maya', 'Hendra', 'Wulan', 'Tono', 'Lina', 'Yoga', 'Nia'] AS first_names,
        ARRAY['Hidayat', 'Wijaya', 'Nugroho', 'Santoso', 'Rahayu', 'Setiawan', 'Kurniawan', 'Maharani', 'Permana', 'Utami'] AS last_names
    ) AS n
) AS all_couriers;

CREATE TEMP TABLE seed_branch ON COMMIT DROP AS
SELECT branch_id FROM seed_courier GROUP BY branch_id
UNION ALL SELECT 'BDG01';

DELETE FROM couriers WHERE id IN (SELECT id FROM seed_courier);

INSERT INTO couriers (id, name, branch_id, created_at)
SELECT id, name, branch_id, TIMESTAMPTZ '2026-01-01 00:00:00+07'
FROM seed_courier
ORDER BY branch_id, id;
