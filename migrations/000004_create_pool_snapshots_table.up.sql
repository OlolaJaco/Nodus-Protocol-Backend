CREATE TABLE IF NOT EXISTS pool_snapshots (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id     VARCHAR(60) NOT NULL,
    reserve_0       VARCHAR(40) NOT NULL,
    reserve_1       VARCHAR(40) NOT NULL,
    token_0         VARCHAR(12) NOT NULL,
    token_1         VARCHAR(12) NOT NULL,
    lp_total_supply VARCHAR(40) NOT NULL DEFAULT '0',
    timestamp_last  BIGINT      NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_pool_snapshots_contract_id  ON pool_snapshots(contract_id);
CREATE INDEX IF NOT EXISTS idx_pool_snapshots_created_at   ON pool_snapshots(created_at DESC);
