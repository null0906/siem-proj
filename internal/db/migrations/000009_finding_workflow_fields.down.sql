DROP TABLE IF EXISTS finding_events;
DROP INDEX IF EXISTS idx_findings_due_date;
DROP INDEX IF EXISTS idx_findings_assignee;
ALTER TABLE findings DROP COLUMN IF EXISTS workflow_updated_at, DROP COLUMN IF EXISTS note, DROP COLUMN IF EXISTS due_date, DROP COLUMN IF EXISTS assignee;
