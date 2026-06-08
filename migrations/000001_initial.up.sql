CREATE TYPE source_tool_enum AS ENUM ('firewall', 'edr', 'scanner', 'dlp', 'other');
CREATE TYPE severity_enum AS ENUM ('critical', 'high', 'medium', 'low', 'info');
CREATE TYPE finding_status_enum AS ENUM ('open', 'resolved', 'suppressed');
CREATE TYPE parse_status_enum AS ENUM ('pending', 'success', 'failed', 'partial');

CREATE TABLE deployment (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    version      TEXT NOT NULL DEFAULT '0.1.0',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE source_files (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename       TEXT NOT NULL,
    sha256         TEXT NOT NULL UNIQUE,
    row_count      INT NOT NULL DEFAULT 0,
    parse_status   parse_status_enum NOT NULL DEFAULT 'pending',
    error_log      TEXT NOT NULL DEFAULT '',
    vendor_matched TEXT NOT NULL DEFAULT '',
    ingested_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE findings (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_tool    source_tool_enum NOT NULL DEFAULT 'other',
    source_vendor  TEXT NOT NULL DEFAULT '',
    external_id    TEXT NOT NULL DEFAULT '',
    severity       severity_enum NOT NULL DEFAULT 'info',
    title          TEXT NOT NULL,
    description    TEXT NOT NULL DEFAULT '',
    affected_asset TEXT NOT NULL DEFAULT '',
    first_seen     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status         finding_status_enum NOT NULL DEFAULT 'open',
    raw_payload    JSONB NOT NULL DEFAULT '{}',
    ingested_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    source_file_id UUID REFERENCES source_files(id) ON DELETE SET NULL
);

CREATE INDEX idx_findings_severity       ON findings(severity);
CREATE INDEX idx_findings_source_tool    ON findings(source_tool);
CREATE INDEX idx_findings_status         ON findings(status);
CREATE INDEX idx_findings_ingested_at    ON findings(ingested_at DESC);
CREATE INDEX idx_findings_source_file_id ON findings(source_file_id);
CREATE INDEX idx_findings_first_seen     ON findings(first_seen DESC);

-- pipeline_events tracks state transitions for the ingestion pipeline strip
CREATE TABLE pipeline_events (
    id         BIGSERIAL PRIMARY KEY,
    event_type TEXT NOT NULL, -- 'landed', 'parsed', 'normalized', 'indexed'
    filename   TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pipeline_events_created ON pipeline_events(created_at DESC);
