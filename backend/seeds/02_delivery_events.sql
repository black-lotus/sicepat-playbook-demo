-- delivery_events: satu event sintetis per paket untuk semua cabang, 1 Jan sampai 31 Des 2026 (hari WIB).
-- Alurnya: (1) seed_plan menghitung jumlah paket per kurir per hari dengan rumus, (2) setiap paket
-- diperluas menjadi satu event dengan awb_number unik, waktu, due_at sesuai cut-off SLA, dan alasan gagal.
-- courier_daily_metrics (03) lalu diagregasi dari event ini. Dijalankan lewat seed.sql.

-- ---------------------------------------------------------------------------------------------------
-- Bersihkan: event dengan awb_number seed dan semua event 2026 milik kurir yang dikelola seed
-- (termasuk sisa uji, supaya agregat di 03 hanya berisi data seed).
-- ---------------------------------------------------------------------------------------------------
DELETE FROM delivery_events
WHERE awb_number LIKE 'SG-2026%'
   OR (courier_id IN (SELECT id FROM seed_courier)
       AND occurred_at >= TIMESTAMP '2026-01-01 00:00:00' AT TIME ZONE 'Asia/Jakarta'
       AND occurred_at <  TIMESTAMP '2027-01-01 00:00:00' AT TIME ZONE 'Asia/Jakarta');

-- ---------------------------------------------------------------------------------------------------
-- Alasan gagal: daftar tetap, bobot (persen, jumlah 100) berbeda per cabang.
-- Urutan: Alamat tidak ditemukan, Penerima tidak di tempat, Penerima menolak paket, Nomor telepon tidak aktif,
-- Paket rusak, Cuaca buruk, Kendaraan bermasalah.
-- ---------------------------------------------------------------------------------------------------
CREATE TEMP TABLE seed_reason ON COMMIT DROP AS
SELECT branch_id, reasons[i] AS reason,
       SUM(weights[i]) OVER (PARTITION BY branch_id ORDER BY i) - weights[i] AS lo,
       SUM(weights[i]) OVER (PARTITION BY branch_id ORDER BY i) AS hi
FROM (
    SELECT w.branch_id, w.weights, r.reasons
    FROM (VALUES
        ('JKT01', ARRAY[15, 35, 10, 20, 5,  5, 10]),
        ('JKT02', ARRAY[20, 30, 10, 15, 5, 10, 10]),
        ('JKT03', ARRAY[10, 25, 15, 25, 5, 10, 10]),
        ('SBY01', ARRAY[25, 20, 10, 10, 5, 25,  5]),
        ('MDN01', ARRAY[30, 20, 10, 10, 5, 15, 10]),
        ('BDG02', ARRAY[15, 30, 15, 20, 10, 5,  5]),
        ('SMG01', ARRAY[20, 25, 10, 15, 5, 20,  5]),
        ('DPS01', ARRAY[10, 40, 10, 10, 5, 10, 15])
    ) AS w(branch_id, weights)
    CROSS JOIN (SELECT ARRAY['Alamat tidak ditemukan', 'Penerima tidak di tempat', 'Penerima menolak paket',
                             'Nomor telepon tidak aktif', 'Paket rusak', 'Cuaca buruk', 'Kendaraan bermasalah'] AS reasons) AS r
) AS x
CROSS JOIN LATERAL generate_subscripts(x.weights, 1) AS i;

-- ---------------------------------------------------------------------------------------------------
-- Rencana jumlah paket per kurir per hari WIB: total, delivered, failed, on_time (on_time <= delivered).
-- ---------------------------------------------------------------------------------------------------
CREATE TEMP TABLE seed_day ON COMMIT DROP AS
SELECT DATE '2026-01-01' + g AS d,
       extract(doy FROM DATE '2026-01-01' + g)::int AS doy,
       extract(day FROM DATE '2026-01-01' + g)::int AS dom
FROM generate_series(0, 364) AS g;

CREATE TEMP TABLE seed_plan (
    courier_id TEXT NOT NULL,
    branch_id  TEXT NOT NULL,
    date       DATE NOT NULL,
    total      INT  NOT NULL,
    delivered  INT  NOT NULL,
    failed     INT  NOT NULL,
    on_time    INT  NOT NULL
) ON COMMIT DROP;

-- JKT01, 1 sampai 7 Nov 2026: persis tabel "Data contoh" di docs/specs/2026-11-18-courier-performance.md
-- (hari tanpa paket tidak punya baris, sama seperti worker).
INSERT INTO seed_plan
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

-- Kurir berprofil (JKT01 di luar jendela spec, dan cabang pembanding). Paket per hari = base + (doy * k) % span.
-- failed ~ pf persen dari total dan terlambat ~ pl persen dari delivered, dengan variasi harian deterministik.
CREATE TEMP TABLE seed_profile ON COMMIT DROP AS
SELECT c.id AS courier_id, c.branch_id, c.k, p.base, p.span, p.pf, p.pl
FROM seed_courier c
JOIN (VALUES
    ('C001', 5, 3, 10,  8),   -- Budi: tinggi
    ('C003', 4, 3, 15,  0),   -- Dewi: tinggi, selalu tepat SLA
    ('C004', 3, 2, 30, 20),   -- Agus: sedang
    ('C002', 0, 3, 85,  0)    -- Sari: rendah, sering tanpa paket
) AS p(courier_id, base, span, pf, pl) ON p.courier_id = c.id
UNION ALL
-- Cabang pembanding: BDG02 kuat, SMG01 menengah, DPS01 lemah; tiap kurir sedikit berbeda.
SELECT c.id, c.branch_id, c.k, 3, 5,
       b.pf + (c.k * 7) % 20,
       b.pl + (c.k * 5) % 15
FROM seed_courier c
JOIN (VALUES ('BDG02', 8, 5), ('SMG01', 18, 12), ('DPS01', 30, 20)) AS b(branch_id, pf, pl) ON b.branch_id = c.branch_id;

INSERT INTO seed_plan
SELECT p.courier_id, p.branch_id, d.d, t.t, t.t - f.f, f.f, (t.t - f.f) - l.l
FROM seed_profile p
CROSS JOIN seed_day d
CROSS JOIN LATERAL (SELECT p.base + (d.doy * p.k) % p.span AS t) AS t
CROSS JOIN LATERAL (SELECT LEAST(t.t, (t.t * p.pf + (d.doy * 7 + p.k * 13) % 100) / 100) AS f) AS f
CROSS JOIN LATERAL (SELECT LEAST(t.t - f.f, ((t.t - f.f) * p.pl + (d.doy * 11 + p.k * 5) % 100) / 100) AS l) AS l
WHERE t.t > 0
  AND NOT (p.branch_id = 'JKT01' AND d.d BETWEEN DATE '2026-11-01' AND DATE '2026-11-07');

-- JKT02: 40 kurir. Rumus yang sama dengan seed lama (Oktober 2026 tidak berubah), kini seluruh tahun.
INSERT INTO seed_plan
SELECT c.id, c.branch_id, d.d, t.t, t.t - f.f, f.f, (t.t - f.f) - LEAST(t.t - f.f, (c.k + 2 * d.dom) % 3)
FROM seed_courier c
CROSS JOIN seed_day d
CROSS JOIN LATERAL (SELECT (c.k * 7 + d.dom * 3) % 9 AS t) AS t
CROSS JOIN LATERAL (SELECT LEAST(t.t, (c.k + d.dom) % 3) AS f) AS f
WHERE c.branch_id = 'JKT02' AND t.t > 0;

-- JKT03: 60 kurir, selalu ada paket.
INSERT INTO seed_plan
SELECT c.id, c.branch_id, d.d, t.t, t.t - f.f, f.f, (t.t - f.f) - LEAST(t.t - f.f, (c.k + d.doy) % 2)
FROM seed_courier c
CROSS JOIN seed_day d
CROSS JOIN LATERAL (SELECT 3 + (c.k * 5 + d.doy) % 6 AS t) AS t
CROSS JOIN LATERAL (SELECT LEAST(t.t, (c.k * 3 + d.doy) % 4) AS f) AS f
WHERE c.branch_id = 'JKT03';

-- SBY01: 2 kurir, 6 paket per hari.
INSERT INTO seed_plan
SELECT c.id, c.branch_id, d.d, 6, 6 - c.f, c.f, 6 - c.f - 1
FROM (VALUES ('S001', 'SBY01', 0), ('S002', 'SBY01', 2)) AS c(id, branch_id, f)
CROSS JOIN seed_day d;

-- MDN01: M101 dan M102 berangka identik setiap hari (tie-break courier_id), M103 lebih rendah.
INSERT INTO seed_plan
SELECT c.id, c.branch_id, d.d, 5, 5 - c.f, c.f, 5 - c.f
FROM (VALUES ('M101', 'MDN01', 1), ('M102', 'MDN01', 1), ('M103', 'MDN01', 3)) AS c(id, branch_id, f)
CROSS JOIN seed_day d;

-- ---------------------------------------------------------------------------------------------------
-- Perluas rencana menjadi event: satu baris per paket.
-- Paket ke-i dikelompokkan (diputar per hari supaya status tidak selalu berurutan sama) menjadi:
--   A = terkirim tepat SLA, L = terkirim terlambat, F = gagal.
-- Waktu (menit sejak 00:00 WIB hari D) dan waktu ditugaskan dipilih sesuai kelas, lalu due_at dihitung dari
-- waktu ditugaskan dengan cut-off 16:00 WIB:
--   ditugaskan sebelum 16:00 hari X  -> batas X 23:59:59 WIB
--   ditugaskan 16:00 atau lebih      -> batas X+1 16:00:00 WIB
-- ---------------------------------------------------------------------------------------------------
INSERT INTO delivery_events (awb_number, courier_id, branch_id, status, failed_reason, occurred_at, due_at, created_at)
SELECT 'SG-' || to_char(e.date, 'YYYYMMDD') || '-' || e.courier_id || '-' || lpad(e.i::text, 2, '0'),
       e.courier_id,
       e.branch_id,
       CASE WHEN e.cls = 'F' THEN 'failed' ELSE 'delivered' END,
       CASE WHEN e.cls = 'F' THEN r.reason END,
       e.occurred_wib AT TIME ZONE 'Asia/Jakarta',
       CASE WHEN e.assigned_wib::time < TIME '16:00'
            THEN date_trunc('day', e.assigned_wib) + INTERVAL '23 hours 59 minutes 59 seconds'
            ELSE date_trunc('day', e.assigned_wib) + INTERVAL '1 day 16 hours'
       END AT TIME ZONE 'Asia/Jakarta',
       e.occurred_wib AT TIME ZONE 'Asia/Jakarta'
FROM (
    SELECT x.*,
           x.date::timestamp + make_interval(mins => x.m) AS occurred_wib,
           CASE x.cls
               -- A genap: ditugaskan 2 jam sebelum selesai di hari yang sama. A ganjil: ditugaskan kemarin 16:30+,
               -- selesai paling lambat 15:29 sehingga masih sebelum batas 16:00.
               WHEN 'A' THEN CASE WHEN x.i % 2 = 0 THEN x.date::timestamp + make_interval(mins => x.m - 120)
                                  ELSE x.date::timestamp - INTERVAL '1 day' + INTERVAL '16 hours 30 minutes' END
               -- L genap: ditugaskan kemarin 09:00 sampai 15:59, batas kemarin 23:59:59. L ganjil: ditugaskan kemarin
               -- 17:00, batas hari ini 16:00, selesai setelah 16:30.
               WHEN 'L' THEN CASE WHEN x.i % 2 = 0 THEN x.date::timestamp - INTERVAL '1 day' + make_interval(mins => 540 + x.h % 420)
                                  ELSE x.date::timestamp - INTERVAL '1 day' + INTERVAL '17 hours' END
               ELSE x.date::timestamp + make_interval(mins => x.m - 120)
           END AS assigned_wib
    FROM (
        SELECT p.*, c.k, i, d.doy,
               (d.doy * 37 + c.k * 19 + i * 41) AS h,
               CASE WHEN (i - 1 + d.doy + c.k) % p.total < p.on_time THEN 'A'
                    WHEN (i - 1 + d.doy + c.k) % p.total < p.delivered THEN 'L'
                    ELSE 'F' END AS cls
        FROM seed_plan p
        JOIN seed_courier c ON c.id = p.courier_id
        JOIN seed_day d ON d.d = p.date
        CROSS JOIN LATERAL generate_series(1, p.total) AS i
    ) AS y
    CROSS JOIN LATERAL (
        SELECT y.*,
               CASE
                   WHEN y.cls = 'A' AND y.i % 2 = 1 THEN 480 + y.h % 450    -- 08:00 sampai 15:29
                   WHEN y.cls = 'L' AND y.i % 2 = 1 THEN 990 + y.h % 240    -- 16:30 sampai 20:29
                   ELSE 480 + y.h % 720                                     -- 08:00 sampai 19:59
               END AS m
    ) AS x
) AS e
LEFT JOIN seed_reason r
       ON e.cls = 'F' AND r.branch_id = e.branch_id AND (e.h % 100) >= r.lo AND (e.h % 100) < r.hi
ORDER BY e.branch_id, e.courier_id, e.date, e.i;
