CREATE TABLE IF NOT EXISTS batches (
    id         TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS disbursements (
    batch_id        TEXT NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    worker_id       TEXT NOT NULL,
    ord             INT  NOT NULL,
    status          TEXT NOT NULL,
    provider_txn_id TEXT,
    error           TEXT,
    PRIMARY KEY (batch_id, worker_id)
);
