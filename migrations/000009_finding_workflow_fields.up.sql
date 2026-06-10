ALTER TABLE findings
    ADD COLUMN assignee TEXT NOT NULL DEFAULT '',
    ADD COLUMN due_date TIMESTAMPTZ,
    ADD COLUMN note TEXT NOT NULL DEFAULT '',
    ADD COLUMN workflow_updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
CREATE INDEX idx_findings_assignee ON findings(assignee) WHERE assignee <> '';
CREATE INDEX idx_findings_due_date ON findings(due_date) WHERE due_date IS NOT NULL;
CREATE TABLE finding_events (
    id BIGSERIAL PRIMARY KEY, finding_id UUID NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL CHECK (event_type IN ('assigned','due_date_changed','status_changed','note_added')),
    actor TEXT NOT NULL DEFAULT 'system', from_value TEXT NOT NULL DEFAULT '', to_value TEXT NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '', created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_finding_events_finding ON finding_events(finding_id, created_at DESC);
WITH ranked AS (SELECT id, ROW_NUMBER() OVER (ORDER BY severity, first_seen) AS row_num FROM findings WHERE status = 'open')
UPDATE findings f SET
 assignee = CASE ranked.row_num % 4 WHEN 0 THEN 'security@seccomply.demo' WHEN 1 THEN 'it-ops@seccomply.demo' WHEN 2 THEN 'cloud-team@seccomply.demo' ELSE 'appsec@seccomply.demo' END,
 due_date = CASE ranked.row_num % 3 WHEN 0 THEN NOW() - INTERVAL '2 days' WHEN 1 THEN NOW() + INTERVAL '3 days' ELSE NOW() + INTERVAL '21 days' END,
 status = CASE WHEN ranked.row_num % 5 = 0 THEN 'in_progress'::finding_status_enum ELSE f.status END,
 note = CASE WHEN ranked.row_num % 6 = 0 THEN 'Remediation validation is pending.' ELSE '' END, workflow_updated_at = NOW()
FROM ranked WHERE f.id = ranked.id AND ranked.row_num <= 24;
INSERT INTO finding_events (finding_id,event_type,actor,to_value,created_at) SELECT id,'assigned','system',assignee,workflow_updated_at FROM findings WHERE assignee <> '';
INSERT INTO finding_events (finding_id,event_type,actor,to_value,created_at) SELECT id,'due_date_changed','system',due_date::text,workflow_updated_at + INTERVAL '1 millisecond' FROM findings WHERE due_date IS NOT NULL;
INSERT INTO finding_events (finding_id,event_type,actor,to_value,created_at) SELECT id,'status_changed','system',status::text,workflow_updated_at + INTERVAL '2 milliseconds' FROM findings WHERE status = 'in_progress';
INSERT INTO finding_events (finding_id,event_type,actor,note,created_at) SELECT id,'note_added','system',note,workflow_updated_at + INTERVAL '3 milliseconds' FROM findings WHERE note <> '';
