CREATE TABLE assets (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    canonical_key       TEXT NOT NULL UNIQUE,
    display_name        TEXT NOT NULL,
    hostname            TEXT NOT NULL DEFAULT '',
    ip_address          TEXT NOT NULL DEFAULT '',
    asset_risk_score    INT NOT NULL DEFAULT 0 CHECK (asset_risk_score BETWEEN 0 AND 100),
    finding_count       INT NOT NULL DEFAULT 0,
    finding_count_by_source JSONB NOT NULL DEFAULT '{}',
    critical_count      INT NOT NULL DEFAULT 0,
    source_count        INT NOT NULL DEFAULT 0,
    is_public           BOOLEAN NOT NULL DEFAULT FALSE,
    has_edr             BOOLEAN NOT NULL DEFAULT FALSE,
    is_encrypted        BOOLEAN NOT NULL DEFAULT TRUE,
    owner               TEXT NOT NULL DEFAULT '',
    owner_is_privileged BOOLEAN NOT NULL DEFAULT FALSE,
    owner_mfa_enabled   BOOLEAN NOT NULL DEFAULT TRUE,
    first_seen          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE asset_references (
    id            BIGSERIAL PRIMARY KEY,
    asset_id      UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    entity_type   TEXT NOT NULL DEFAULT 'finding' CHECK (entity_type IN ('finding','device','cloud_finding','identity_user')),
    entity_id     UUID NOT NULL,
    source_tool   TEXT NOT NULL DEFAULT '',
    source_vendor TEXT NOT NULL DEFAULT '',
    severity      TEXT NOT NULL DEFAULT 'info',
    status        TEXT NOT NULL DEFAULT '',
    title         TEXT NOT NULL DEFAULT '',
    occurred_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (asset_id, entity_type, entity_id)
);

CREATE INDEX idx_assets_risk ON assets(asset_risk_score DESC);
CREATE INDEX idx_asset_references_asset ON asset_references(asset_id, occurred_at DESC);
