CREATE TABLE posture_snapshots (
    snapshot_date       DATE PRIMARY KEY,
    overall_score       INT NOT NULL CHECK (overall_score BETWEEN 0 AND 100),
    vulnerability_score INT NOT NULL CHECK (vulnerability_score BETWEEN 0 AND 100),
    identity_score      INT NOT NULL CHECK (identity_score BETWEEN 0 AND 100),
    endpoint_score      INT NOT NULL CHECK (endpoint_score BETWEEN 0 AND 100),
    cloud_score         INT NOT NULL CHECK (cloud_score BETWEEN 0 AND 100),
    compliance_score    INT NOT NULL CHECK (compliance_score BETWEEN 0 AND 100),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posture_snapshots_date ON posture_snapshots(snapshot_date DESC);
