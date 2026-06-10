package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seccomply/seccomply/internal/scoring"
)

const highRiskThreshold = 80

type Event struct {
	ID         int64          `json:"id"`
	EventType  string         `json:"event_type"`
	Title      string         `json:"title"`
	Detail     string         `json:"detail"`
	Severity   string         `json:"severity"`
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	Metadata   map[string]any `json:"metadata"`
	CreatedAt  time.Time      `json:"created_at"`
}

type PeriodMetrics struct {
	Days             int `json:"days"`
	NewFindings      int `json:"new_findings"`
	ResolvedFindings int `json:"resolved_findings"`
	NewCriticals     int `json:"new_criticals"`
	FilesIngested    int `json:"files_ingested"`
	ScoreDelta       int `json:"score_delta"`
}

type PeriodComparison struct {
	Current  PeriodMetrics `json:"current"`
	Previous PeriodMetrics `json:"previous"`
	Delta    PeriodMetrics `json:"delta"`
}

type SinceLastIngest struct {
	Available      bool       `json:"available"`
	Filename       string     `json:"filename"`
	IngestedAt     *time.Time `json:"ingested_at"`
	NewFindings    int        `json:"new_findings"`
	Resolved       int        `json:"resolved"`
	NewCriticals   int        `json:"new_criticals"`
	ScoreDelta     int        `json:"score_delta"`
	HighRiskAssets int        `json:"high_risk_assets"`
}

type Feed struct {
	Events          []Event          `json:"events"`
	Total           int              `json:"total"`
	SinceLastIngest SinceLastIngest  `json:"since_last_ingest"`
	Comparison      PeriodComparison `json:"comparison"`
}

type State struct {
	Score          int
	HighRiskAssets map[string]bool
}

type IngestResult struct {
	SourceFileID uuid.UUID
	Filename     string
	Vendor       string
	Rows         int
	NewFindings  int
	NewCriticals int
	Resolved     int
}

type Service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) CaptureState(ctx context.Context) (State, error) {
	scores, err := scoring.New(s.db).CurrentScores(ctx)
	if err != nil {
		return State{}, err
	}
	rows, err := s.db.Query(ctx, `SELECT canonical_key FROM assets WHERE asset_risk_score >= $1`, highRiskThreshold)
	if err != nil {
		return State{}, err
	}
	defer rows.Close()
	state := State{Score: scores.Overall, HighRiskAssets: map[string]bool{}}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return State{}, err
		}
		state.HighRiskAssets[key] = true
	}
	return state, rows.Err()
}

func (s *Service) RecordIngest(ctx context.Context, before State, result IngestResult) error {
	after, err := s.CaptureState(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	if err := s.record(ctx, "file_ingested", result.Filename+" ingested", fmt.Sprintf("%d records indexed from %s.", result.Rows, result.Vendor), "info", "source_file", result.SourceFileID.String(), map[string]any{
		"filename": result.Filename, "vendor": result.Vendor, "row_count": result.Rows,
	}, now); err != nil {
		return err
	}
	if result.NewFindings > 0 {
		if err := s.record(ctx, "findings_added", fmt.Sprintf("%d new findings identified", result.NewFindings), fmt.Sprintf("%s added new security evidence.", result.Vendor), "warning", "source_file", result.SourceFileID.String(), map[string]any{"count": result.NewFindings}, now.Add(time.Millisecond)); err != nil {
			return err
		}
	}
	if result.Resolved > 0 {
		if err := s.record(ctx, "findings_resolved", fmt.Sprintf("%d findings resolved", result.Resolved), "Previously open findings are no longer active.", "success", "source_file", result.SourceFileID.String(), map[string]any{"count": result.Resolved}, now.Add(2*time.Millisecond)); err != nil {
			return err
		}
	}
	if result.NewCriticals > 0 {
		if err := s.record(ctx, "critical_detected", fmt.Sprintf("%d new critical findings detected", result.NewCriticals), "Immediate review is recommended.", "critical", "source_file", result.SourceFileID.String(), map[string]any{"count": result.NewCriticals}, now.Add(3*time.Millisecond)); err != nil {
			return err
		}
	}
	if delta := after.Score - before.Score; delta != 0 {
		severity := "success"
		title := "Security posture improved"
		if delta < 0 {
			severity = "warning"
			title = "Security posture declined"
		}
		if err := s.record(ctx, "score_changed", title, fmt.Sprintf("Posture moved from %d to %d.", before.Score, after.Score), severity, "", "", map[string]any{"previous_score": before.Score, "score": after.Score, "delta": delta}, now.Add(4*time.Millisecond)); err != nil {
			return err
		}
	}
	for key := range after.HighRiskAssets {
		if before.HighRiskAssets[key] {
			continue
		}
		if err := s.record(ctx, "asset_risk_threshold", "Asset crossed the high-risk threshold", "A correlated asset now requires priority attention.", "critical", "asset", key, map[string]any{"risk_threshold": highRiskThreshold}, now.Add(5*time.Millisecond)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Feed(ctx context.Context, limit, days int) (Feed, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if days != 7 && days != 30 && days != 60 {
		days = 7
	}
	var total int
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM activity_events`).Scan(&total); err != nil {
		return Feed{}, err
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, event_type, title, detail, severity, entity_type, entity_id, metadata, created_at
		FROM activity_events ORDER BY created_at DESC, id DESC LIMIT $1`, limit)
	if err != nil {
		return Feed{}, err
	}
	defer rows.Close()
	events := []Event{}
	for rows.Next() {
		var event Event
		var metadata []byte
		if err := rows.Scan(&event.ID, &event.EventType, &event.Title, &event.Detail, &event.Severity, &event.EntityType, &event.EntityID, &metadata, &event.CreatedAt); err != nil {
			return Feed{}, err
		}
		_ = json.Unmarshal(metadata, &event.Metadata)
		events = append(events, event)
	}
	comparison, err := s.comparison(ctx, days)
	if err != nil {
		return Feed{}, err
	}
	since, err := s.sinceLastIngest(ctx)
	if err != nil {
		return Feed{}, err
	}
	return Feed{Events: events, Total: total, SinceLastIngest: since, Comparison: comparison}, rows.Err()
}

func (s *Service) comparison(ctx context.Context, days int) (PeriodComparison, error) {
	current, err := s.period(ctx, days, 0)
	if err != nil {
		return PeriodComparison{}, err
	}
	previous, err := s.period(ctx, days, days)
	if err != nil {
		return PeriodComparison{}, err
	}
	return PeriodComparison{
		Current:  current,
		Previous: previous,
		Delta: PeriodMetrics{
			Days:             days,
			NewFindings:      current.NewFindings - previous.NewFindings,
			ResolvedFindings: current.ResolvedFindings - previous.ResolvedFindings,
			NewCriticals:     current.NewCriticals - previous.NewCriticals,
			FilesIngested:    current.FilesIngested - previous.FilesIngested,
			ScoreDelta:       current.ScoreDelta - previous.ScoreDelta,
		},
	}, nil
}

func (s *Service) period(ctx context.Context, days, offset int) (PeriodMetrics, error) {
	var metrics PeriodMetrics
	metrics.Days = days
	err := s.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE first_seen >= NOW() - (($1::int + $2::int) * INTERVAL '1 day') AND first_seen < NOW() - ($2::int * INTERVAL '1 day')),
			COUNT(*) FILTER (WHERE status = 'resolved' AND last_seen >= NOW() - (($1::int + $2::int) * INTERVAL '1 day') AND last_seen < NOW() - ($2::int * INTERVAL '1 day')),
			COUNT(*) FILTER (WHERE severity = 'critical' AND first_seen >= NOW() - (($1::int + $2::int) * INTERVAL '1 day') AND first_seen < NOW() - ($2::int * INTERVAL '1 day'))
		FROM findings`, days, offset).Scan(&metrics.NewFindings, &metrics.ResolvedFindings, &metrics.NewCriticals)
	if err != nil {
		return PeriodMetrics{}, err
	}
	if err := s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM source_files
		WHERE ingested_at >= NOW() - (($1::int + $2::int) * INTERVAL '1 day')
		  AND ingested_at < NOW() - ($2::int * INTERVAL '1 day')`, days, offset).Scan(&metrics.FilesIngested); err != nil {
		return PeriodMetrics{}, err
	}
	var start, end int
	err = s.db.QueryRow(ctx, `
		SELECT
			COALESCE((SELECT overall_score FROM posture_snapshots
				WHERE snapshot_date >= CURRENT_DATE - ($1::int + $2::int - 1)
				  AND snapshot_date <= CURRENT_DATE - $2::int
				ORDER BY snapshot_date ASC LIMIT 1), 0),
			COALESCE((SELECT overall_score FROM posture_snapshots
				WHERE snapshot_date >= CURRENT_DATE - ($1::int + $2::int - 1)
				  AND snapshot_date <= CURRENT_DATE - $2::int
				ORDER BY snapshot_date DESC LIMIT 1), 0)`,
		days, offset).Scan(&start, &end)
	if err != nil {
		return PeriodMetrics{}, err
	}
	metrics.ScoreDelta = end - start
	return metrics, nil
}

func (s *Service) sinceLastIngest(ctx context.Context) (SinceLastIngest, error) {
	var summary SinceLastIngest
	var metadata []byte
	var ingestedAt time.Time
	err := s.db.QueryRow(ctx, `
		SELECT metadata, created_at
		FROM activity_events WHERE event_type = 'file_ingested'
		ORDER BY created_at DESC, id DESC LIMIT 1`).Scan(&metadata, &ingestedAt)
	if err != nil {
		return summary, nil
	}
	var values map[string]any
	_ = json.Unmarshal(metadata, &values)
	summary.Available = true
	summary.Filename, _ = values["filename"].(string)
	summary.IngestedAt = &ingestedAt
	err = s.db.QueryRow(ctx, `
		SELECT
			COALESCE(SUM((metadata->>'count')::int) FILTER (WHERE event_type = 'findings_added'), 0),
			COALESCE(SUM((metadata->>'count')::int) FILTER (WHERE event_type = 'findings_resolved'), 0),
			COALESCE(SUM((metadata->>'count')::int) FILTER (WHERE event_type = 'critical_detected'), 0),
			COALESCE(SUM((metadata->>'delta')::int) FILTER (WHERE event_type = 'score_changed'), 0),
			COUNT(*) FILTER (WHERE event_type = 'asset_risk_threshold')
		FROM activity_events WHERE created_at >= $1`, ingestedAt).Scan(
		&summary.NewFindings, &summary.Resolved, &summary.NewCriticals, &summary.ScoreDelta, &summary.HighRiskAssets,
	)
	return summary, err
}

func (s *Service) record(ctx context.Context, eventType, title, detail, severity, entityType, entityID string, metadata map[string]any, createdAt time.Time) error {
	payload, _ := json.Marshal(metadata)
	_, err := s.db.Exec(ctx, `
		INSERT INTO activity_events (event_type, title, detail, severity, entity_type, entity_id, metadata, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		eventType, title, detail, severity, entityType, entityID, payload, createdAt,
	)
	return err
}
