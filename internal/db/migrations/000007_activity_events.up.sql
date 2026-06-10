CREATE TABLE activity_events (
    id         BIGSERIAL PRIMARY KEY,
    event_type TEXT NOT NULL CHECK (event_type IN (
        'file_ingested','findings_added','findings_resolved','critical_detected',
        'score_changed','asset_risk_threshold'
    )),
    title      TEXT NOT NULL,
    detail     TEXT NOT NULL DEFAULT '',
    severity   TEXT NOT NULL DEFAULT 'info' CHECK (severity IN ('info','success','warning','critical')),
    entity_type TEXT NOT NULL DEFAULT '',
    entity_id   TEXT NOT NULL DEFAULT '',
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_events_created ON activity_events(created_at DESC);
CREATE INDEX idx_activity_events_type_created ON activity_events(event_type, created_at DESC);

INSERT INTO activity_events (event_type, title, detail, severity, entity_type, entity_id, metadata, created_at)
SELECT
    'file_ingested',
    filename || ' ingested',
    row_count || ' records indexed from ' || COALESCE(NULLIF(vendor_matched, ''), 'an uploaded source') || '.',
    'info',
    'source_file',
    id::text,
    jsonb_build_object('filename', filename, 'row_count', row_count, 'vendor', vendor_matched),
    ingested_at
FROM source_files
WHERE parse_status = 'success';

INSERT INTO activity_events (event_type, title, detail, severity, entity_type, entity_id, metadata, created_at)
SELECT
    'findings_added',
    COUNT(f.id) || ' new findings identified',
    COALESCE(NULLIF(sf.vendor_matched, ''), 'Uploaded source') || ' added new security evidence.',
    'warning',
    'source_file',
    sf.id::text,
    jsonb_build_object('count', COUNT(f.id)),
    sf.ingested_at + INTERVAL '1 millisecond'
FROM source_files sf
JOIN findings f ON f.source_file_id = sf.id
GROUP BY sf.id, sf.vendor_matched, sf.ingested_at
HAVING COUNT(f.id) > 0;

INSERT INTO activity_events (event_type, title, detail, severity, entity_type, entity_id, metadata, created_at)
SELECT
    'critical_detected',
    COUNT(f.id) || ' new critical findings detected',
    'Immediate review is recommended.',
    'critical',
    'source_file',
    sf.id::text,
    jsonb_build_object('count', COUNT(f.id)),
    sf.ingested_at + INTERVAL '2 milliseconds'
FROM source_files sf
JOIN findings f ON f.source_file_id = sf.id AND f.severity = 'critical'
GROUP BY sf.id, sf.ingested_at
HAVING COUNT(f.id) > 0;

INSERT INTO activity_events (event_type, title, detail, severity, entity_type, entity_id, metadata, created_at)
SELECT
    'findings_resolved',
    COUNT(f.id) || ' findings resolved',
    'Previously open findings are no longer active.',
    'success',
    'source_file',
    sf.id::text,
    jsonb_build_object('count', COUNT(f.id)),
    sf.ingested_at + INTERVAL '3 milliseconds'
FROM source_files sf
JOIN findings f ON f.source_file_id = sf.id AND f.status = 'resolved'
GROUP BY sf.id, sf.ingested_at
HAVING COUNT(f.id) > 0;

INSERT INTO activity_events (event_type, title, detail, severity, entity_type, entity_id, metadata, created_at)
SELECT
    'asset_risk_threshold',
    'Asset crossed the high-risk threshold',
    'A correlated asset now requires priority attention.',
    'critical',
    'asset',
    canonical_key,
    jsonb_build_object('risk_threshold', 80, 'risk_score', asset_risk_score),
    updated_at + INTERVAL '4 milliseconds'
FROM assets
WHERE asset_risk_score >= 80;

INSERT INTO activity_events (event_type, title, detail, severity, metadata, created_at)
SELECT
    'score_changed',
    CASE WHEN overall_score - previous_score > 0 THEN 'Security posture improved' ELSE 'Security posture declined' END,
    'Posture moved from ' || previous_score || ' to ' || overall_score || '.',
    CASE WHEN overall_score - previous_score > 0 THEN 'success' ELSE 'warning' END,
    jsonb_build_object('previous_score', previous_score, 'score', overall_score, 'delta', overall_score - previous_score),
    snapshot_date::timestamptz + INTERVAL '12 hours'
FROM (
    SELECT snapshot_date, overall_score,
           LAG(overall_score) OVER (ORDER BY snapshot_date) AS previous_score
    FROM posture_snapshots
) snapshots
WHERE previous_score IS NOT NULL AND overall_score <> previous_score;
