-- Target KPI bulanan per cabang. Hanya dibaca aplikasi, diisi lewat seed dan migrasi.
CREATE TABLE branch_kpi_targets (
    branch_id          TEXT NOT NULL,
    month              DATE NOT NULL CHECK (month = date_trunc('month', month)::date),
    success_rate       NUMERIC(5, 4) NOT NULL CHECK (success_rate BETWEEN 0 AND 1),
    sla_rate           NUMERIC(5, 4) NOT NULL CHECK (sla_rate BETWEEN 0 AND 1),
    delivered_packages INTEGER NOT NULL CHECK (delivered_packages >= 0),
    PRIMARY KEY (branch_id, month)
);
