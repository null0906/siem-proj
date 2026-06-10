package remediation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seccomply/seccomply/internal/models"
)

type Update struct {
	Actor    string     `json:"actor"`
	Assignee *string    `json:"assignee"`
	DueDate  *time.Time `json:"due_date"`
	Status   *string    `json:"status"`
	Note     *string    `json:"note"`
}

type Service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func ApplySLA(finding *models.Finding, now time.Time) {
	due := finding.FirstSeen.Add(slaDuration(finding.Severity))
	if finding.DueDate != nil {
		due = *finding.DueDate
	}
	finding.SLADueAt = due
	if finding.Status == models.StatusResolved || string(finding.Status) == "risk_accepted" {
		finding.SLAStatus = "closed"
		return
	}
	remaining := due.Sub(now)
	switch {
	case remaining < 0:
		finding.SLAStatus = "breached"
	case remaining <= 72*time.Hour:
		finding.SLAStatus = "due_soon"
	default:
		finding.SLAStatus = "on_track"
	}
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, update Update) (models.Finding, error) {
	actor := strings.TrimSpace(update.Actor)
	if actor == "" {
		actor = "analyst@seccomply.demo"
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return models.Finding{}, err
	}
	defer tx.Rollback(ctx)

	var current struct {
		Assignee string
		DueDate  *time.Time
		Status   string
		Note     string
	}
	err = tx.QueryRow(ctx, `SELECT assignee, due_date, status::text, note FROM findings WHERE id = $1 FOR UPDATE`, id).
		Scan(&current.Assignee, &current.DueDate, &current.Status, &current.Note)
	if err != nil {
		return models.Finding{}, err
	}

	assignee := current.Assignee
	dueDate := current.DueDate
	status := current.Status
	note := current.Note
	if update.Assignee != nil {
		assignee = strings.TrimSpace(*update.Assignee)
		if assignee != current.Assignee {
			if err := event(ctx, tx, id, "assigned", actor, current.Assignee, assignee, ""); err != nil {
				return models.Finding{}, err
			}
		}
	}
	if update.DueDate != nil {
		dueDate = update.DueDate
		from := ""
		if current.DueDate != nil {
			from = current.DueDate.Format(time.RFC3339)
		}
		if err := event(ctx, tx, id, "due_date_changed", actor, from, dueDate.Format(time.RFC3339), ""); err != nil {
			return models.Finding{}, err
		}
	}
	if update.Status != nil {
		status = strings.TrimSpace(*update.Status)
		if !validStatus(status) {
			return models.Finding{}, fmt.Errorf("invalid workflow status %q", status)
		}
		if status != current.Status {
			if err := event(ctx, tx, id, "status_changed", actor, current.Status, status, ""); err != nil {
				return models.Finding{}, err
			}
		}
	}
	if update.Note != nil && strings.TrimSpace(*update.Note) != "" {
		note = strings.TrimSpace(*update.Note)
		if err := event(ctx, tx, id, "note_added", actor, "", "", note); err != nil {
			return models.Finding{}, err
		}
	}

	_, err = tx.Exec(ctx, `
		UPDATE findings SET assignee = $2, due_date = $3, status = $4,
		       note = $5, workflow_updated_at = NOW()
		WHERE id = $1`, id, assignee, dueDate, status, note)
	if err != nil {
		return models.Finding{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return models.Finding{}, err
	}
	return s.Get(ctx, id)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (models.Finding, error) {
	var f models.Finding
	var raw []byte
	err := s.db.QueryRow(ctx, `
		SELECT id, source_tool, source_vendor, external_id, severity, title, description,
		       affected_asset, first_seen, last_seen, status, assignee, due_date, note,
		       raw_payload, ingested_at, source_file_id
		FROM findings WHERE id = $1`, id).Scan(
		&f.ID, &f.SourceTool, &f.SourceVendor, &f.ExternalID, &f.Severity, &f.Title, &f.Description,
		&f.AffectedAsset, &f.FirstSeen, &f.LastSeen, &f.Status, &f.Assignee, &f.DueDate, &f.Note,
		&raw, &f.IngestedAt, &f.SourceFileID,
	)
	if err != nil {
		return models.Finding{}, err
	}
	_ = json.Unmarshal(raw, &f.RawPayload)
	rows, err := s.db.Query(ctx, `
		SELECT id, event_type, actor, from_value, to_value, note, created_at
		FROM finding_events WHERE finding_id = $1 ORDER BY created_at DESC, id DESC`, id)
	if err != nil {
		return models.Finding{}, err
	}
	defer rows.Close()
	f.Events = []models.FindingEvent{}
	for rows.Next() {
		var item models.FindingEvent
		if err := rows.Scan(&item.ID, &item.EventType, &item.Actor, &item.FromValue, &item.ToValue, &item.Note, &item.CreatedAt); err != nil {
			return models.Finding{}, err
		}
		f.Events = append(f.Events, item)
	}
	ApplySLA(&f, time.Now())
	return f, rows.Err()
}

func event(ctx context.Context, tx pgx.Tx, id uuid.UUID, eventType, actor, from, to, note string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO finding_events (finding_id, event_type, actor, from_value, to_value, note)
		VALUES ($1,$2,$3,$4,$5,$6)`, id, eventType, actor, from, to, note)
	return err
}

func validStatus(status string) bool {
	switch status {
	case "open", "in_progress", "resolved", "risk_accepted":
		return true
	default:
		return false
	}
}

func slaDuration(severity models.Severity) time.Duration {
	switch severity {
	case models.SeverityCritical:
		return 7 * 24 * time.Hour
	case models.SeverityHigh:
		return 30 * 24 * time.Hour
	case models.SeverityMedium:
		return 60 * 24 * time.Hour
	default:
		return 90 * 24 * time.Hour
	}
}
