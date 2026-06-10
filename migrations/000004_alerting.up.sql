CREATE TABLE alert_rules (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name               TEXT NOT NULL UNIQUE,
    description        TEXT NOT NULL DEFAULT '',
    category           TEXT NOT NULL CHECK (category IN ('vulnerability','identity','cloud','endpoint','compliance','posture','asset')),
    condition_type     TEXT NOT NULL CHECK (condition_type IN ('threshold','delta','new_entity','sla')),
    metric             TEXT NOT NULL,
    operator           TEXT CHECK (operator IS NULL OR operator IN ('gt','lt','gte','lte','eq')),
    threshold_value    DOUBLE PRECISION,
    delta_direction    TEXT CHECK (delta_direction IS NULL OR delta_direction IN ('increase','decrease')),
    delta_amount       DOUBLE PRECISION,
    severity           severity_enum NOT NULL DEFAULT 'medium',
    enabled            BOOLEAN NOT NULL DEFAULT TRUE,
    channel_ids        INT[] NOT NULL DEFAULT '{}',
    evaluation_trigger TEXT NOT NULL DEFAULT 'both' CHECK (evaluation_trigger IN ('on_ingest','scheduled','both')),
    cooldown_minutes   INT NOT NULL DEFAULT 60 CHECK (cooldown_minutes >= 0),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE alerts (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id          UUID NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    fingerprint      TEXT NOT NULL,
    status           TEXT NOT NULL DEFAULT 'firing' CHECK (status IN ('firing','acknowledged','resolved')),
    severity         severity_enum NOT NULL,
    title            TEXT NOT NULL,
    context          JSONB NOT NULL DEFAULT '{}',
    triggered_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    acknowledged_at  TIMESTAMPTZ,
    acknowledged_by  TEXT,
    resolved_at      TIMESTAMPTZ,
    auto_resolved    BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE alert_rule_state (
    rule_id           UUID PRIMARY KEY REFERENCES alert_rules(id) ON DELETE CASCADE,
    last_metric_value DOUBLE PRECISION,
    last_evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_alerts_active_fingerprint
    ON alerts(rule_id, fingerprint)
    WHERE status IN ('firing', 'acknowledged');
CREATE INDEX idx_alerts_status_severity ON alerts(status, severity);
CREATE INDEX idx_alerts_rule_triggered ON alerts(rule_id, triggered_at DESC);
