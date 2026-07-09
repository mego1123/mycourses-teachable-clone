CREATE TABLE system_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cpu_usage FLOAT, memory_usage FLOAT, disk_usage FLOAT, goroutines INT,
    http_requests JSONB, mongodb_stats JSONB, go_runtime JSONB, integrations JSONB,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_system_metrics_time ON system_metrics(timestamp DESC);

CREATE TABLE daily_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL, dau INT NOT NULL DEFAULT 0, wau INT NOT NULL DEFAULT 0, mau INT NOT NULL DEFAULT 0,
    new_users INT NOT NULL DEFAULT 0, new_tenants INT NOT NULL DEFAULT 0, active_tenants INT NOT NULL DEFAULT 0,
    arr_cents BIGINT NOT NULL DEFAULT 0, mrr_cents BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_daily_metrics_date ON daily_metrics(date);

CREATE TABLE leader_locks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lock_name TEXT NOT NULL, holder_id TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL, last_renewed TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_leader_locks_name ON leader_locks(lock_name);
CREATE INDEX idx_leader_locks_expires ON leader_locks(expires_at);
