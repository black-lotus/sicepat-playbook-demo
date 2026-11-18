-- Data kurir. Data pribadi (nomor HP, NIK, alamat) sengaja tidak disimpan di sini.
CREATE TABLE couriers (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    branch_id  TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Satu baris per percobaan pengiriman paket.
CREATE TABLE delivery_events (
    id            BIGSERIAL PRIMARY KEY,
    awb_number    TEXT NOT NULL,
    courier_id    TEXT NOT NULL,
    branch_id     TEXT NOT NULL,
    status        TEXT NOT NULL CHECK (status IN ('delivered', 'failed')),
    failed_reason TEXT,
    occurred_at   TIMESTAMPTZ NOT NULL,
    due_at        TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (awb_number, status, occurred_at)
);

CREATE INDEX idx_delivery_events_occurred_at ON delivery_events (occurred_at);

-- Agregat KPI harian per kurir, diisi oleh worker.
CREATE TABLE courier_daily_metrics (
    courier_id     TEXT NOT NULL,
    branch_id      TEXT NOT NULL,
    date           DATE NOT NULL,
    total_packages INT NOT NULL,
    delivered      INT NOT NULL,
    failed         INT NOT NULL,
    on_time        INT NOT NULL,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (courier_id, date)
);

CREATE INDEX idx_courier_daily_metrics_date_branch ON courier_daily_metrics (date, branch_id);
