package alerting

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seccomply/seccomply/internal/scoring"
)

const scheduledInterval = 15 * time.Minute

type Trigger string

const (
	TriggerManual    Trigger = "manual"
	TriggerOnIngest  Trigger = "on_ingest"
	TriggerScheduled Trigger = "scheduled"
)

type Rule struct {
	ID                uuid.UUID
	Name              string
	Category          string
	ConditionType     string
	Metric            string
	Operator          *string
	ThresholdValue    *float64
	DeltaDirection    *string
	DeltaAmount       *float64
	Severity          string
	EvaluationTrigger string
	CooldownMinutes   int
}

type Match struct {
	Scope   string
	Title   string
	Context map[string]any
}

type Evaluation struct {
	RulesEvaluated int `json:"rules_evaluated"`
	Fired          int `json:"fired"`
	Suppressed     int `json:"suppressed"`
	AutoResolved   int `json:"auto_resolved"`
}

type Service struct {
	db  *pgxpool.Pool
	log *slog.Logger
	now func() time.Time
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db, log: slog.Default(), now: time.Now}
}

func RunScheduled(ctx context.Context, db *pgxpool.Pool) {
	service := New(db)
	ticker := time.NewTicker(scheduledInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result, err := service.Evaluate(ctx, TriggerScheduled)
			if err != nil {
				service.log.Error("scheduled alert evaluation failed", "err", err)
				continue
			}
			service.log.Info("scheduled alert evaluation complete", "fired", result.Fired, "resolved", result.AutoResolved, "suppressed", result.Suppressed)
		}
	}
}

func (s *Service) Evaluate(ctx context.Context, trigger Trigger) (Evaluation, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Evaluation{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(7392041)`); err != nil {
		return Evaluation{}, err
	}

	rules, err := loadRules(ctx, tx, trigger)
	if err != nil {
		return Evaluation{}, err
	}

	var result Evaluation
	for _, rule := range rules {
		outcome, err := s.evaluateRule(ctx, tx, rule)
		if err != nil {
			return Evaluation{}, fmt.Errorf("evaluate rule %q: %w", rule.Name, err)
		}
		result.RulesEvaluated++
		result.Fired += outcome.Fired
		result.Suppressed += outcome.Suppressed
		result.AutoResolved += outcome.AutoResolved
	}
	if err := tx.Commit(ctx); err != nil {
		return Evaluation{}, err
	}
	return result, nil
}

func (s *Service) evaluateRule(ctx context.Context, tx pgx.Tx, rule Rule) (Evaluation, error) {
	lastValue, lastEvaluated, hasState, err := loadState(ctx, tx, rule.ID)
	if err != nil {
		return Evaluation{}, err
	}

	matches, currentScopes, currentValue, err := s.matches(ctx, tx, rule, lastValue, lastEvaluated, hasState)
	if err != nil {
		return Evaluation{}, err
	}
	active, err := loadActiveFingerprints(ctx, tx, rule.ID)
	if err != nil {
		return Evaluation{}, err
	}

	var result Evaluation
	for _, match := range matches {
		fingerprint := Fingerprint(rule.ID, match.Scope)
		if active[fingerprint] {
			result.Suppressed++
			continue
		}
		var cooling bool
		err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM alerts
				 WHERE rule_id = $1 AND fingerprint = $2
				   AND triggered_at > $3::timestamptz - make_interval(mins => $4::int)
			)`, rule.ID, fingerprint, s.now(), rule.CooldownMinutes).Scan(&cooling)
		if err != nil {
			return Evaluation{}, err
		}
		if cooling {
			result.Suppressed++
			continue
		}
		payload, _ := json.Marshal(match.Context)
		if _, err := tx.Exec(ctx, `
			INSERT INTO alerts (rule_id, fingerprint, severity, title, context, triggered_at)
			VALUES ($1,$2,$3,$4,$5,$6)`,
			rule.ID, fingerprint, rule.Severity, match.Title, payload, s.now(),
		); err != nil {
			return Evaluation{}, err
		}
		result.Fired++
	}

	for fingerprint := range active {
		if currentScopes[fingerprint] {
			continue
		}
		tag, err := tx.Exec(ctx, `
			UPDATE alerts
			   SET status = 'resolved', resolved_at = $3, auto_resolved = TRUE
			 WHERE rule_id = $1 AND fingerprint = $2
			   AND status IN ('firing','acknowledged')`,
			rule.ID, fingerprint, s.now(),
		)
		if err != nil {
			return Evaluation{}, err
		}
		result.AutoResolved += int(tag.RowsAffected())
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO alert_rule_state (rule_id, last_metric_value, last_evaluated_at, updated_at)
		VALUES ($1,$2,$3,$3)
		ON CONFLICT (rule_id) DO UPDATE SET
			last_metric_value = EXCLUDED.last_metric_value,
			last_evaluated_at = EXCLUDED.last_evaluated_at,
			updated_at = EXCLUDED.updated_at`,
		rule.ID, currentValue, s.now(),
	); err != nil {
		return Evaluation{}, err
	}
	return result, nil
}

func (s *Service) matches(ctx context.Context, tx pgx.Tx, rule Rule, lastValue float64, lastEvaluated time.Time, hasState bool) ([]Match, map[string]bool, float64, error) {
	switch rule.ConditionType {
	case "threshold":
		value, err := metricValue(ctx, tx, rule.Metric)
		if err != nil {
			return nil, nil, 0, err
		}
		scopes := map[string]bool{}
		if rule.ThresholdValue != nil && rule.Operator != nil && Compare(value, *rule.Operator, *rule.ThresholdValue) {
			match := Match{
				Scope:   "global",
				Title:   fmt.Sprintf("%s: %.0f %s %.0f", rule.Name, value, operatorText(*rule.Operator), *rule.ThresholdValue),
				Context: map[string]any{"metric": rule.Metric, "value": value, "operator": *rule.Operator, "threshold": *rule.ThresholdValue},
			}
			scopes[Fingerprint(rule.ID, match.Scope)] = true
			return []Match{match}, scopes, value, nil
		}
		return nil, scopes, value, nil
	case "delta":
		value, err := metricValue(ctx, tx, rule.Metric)
		if err != nil {
			return nil, nil, 0, err
		}
		scopes := map[string]bool{}
		if !hasState || rule.DeltaAmount == nil || rule.DeltaDirection == nil {
			return nil, scopes, value, nil
		}
		delta := value - lastValue
		fires := (*rule.DeltaDirection == "increase" && delta >= *rule.DeltaAmount) ||
			(*rule.DeltaDirection == "decrease" && -delta >= *rule.DeltaAmount)
		if fires {
			match := Match{
				Scope:   "global",
				Title:   fmt.Sprintf("%s: %.0f → %.0f", rule.Name, lastValue, value),
				Context: map[string]any{"metric": rule.Metric, "previous": lastValue, "value": value, "delta": delta},
			}
			scopes[Fingerprint(rule.ID, match.Scope)] = true
			return []Match{match}, scopes, value, nil
		}
		return nil, scopes, value, nil
	case "new_entity":
		return newEntityMatches(ctx, tx, rule, lastEvaluated, hasState)
	case "sla":
		return slaMatches(ctx, tx, rule)
	default:
		return nil, nil, 0, fmt.Errorf("unsupported condition type %q", rule.ConditionType)
	}
}

func newEntityMatches(ctx context.Context, tx pgx.Tx, rule Rule, since time.Time, hasState bool) ([]Match, map[string]bool, float64, error) {
	if !hasState {
		since = time.Now()
	}
	var candidateSQL, currentSQL string
	switch rule.Metric {
	case "new_privileged_without_mfa":
		candidateSQL = `
			SELECT u.id::text, u.user_email
			FROM identity_users u
			LEFT JOIN source_files sf ON sf.id = u.source_file_id
			WHERE u.is_privileged AND NOT u.mfa_enabled
			  AND COALESCE(sf.ingested_at, u.created_at) > $1`
		currentSQL = `SELECT id::text FROM identity_users WHERE is_privileged AND NOT mfa_enabled`
	case "new_public_cloud_exposure":
		candidateSQL = `
			SELECT id::text, COALESCE(NULLIF(affected_asset, ''), title)
			FROM findings
			WHERE status IN ('open','in_progress') AND ingested_at > $1
			  AND (source_vendor ILIKE '%aws%' OR source_vendor ILIKE '%azure%' OR source_vendor ILIKE '%gcp%' OR COALESCE(raw_payload->>'cloud_provider', '') <> '')
			  AND (title ILIKE '%public%' OR COALESCE(raw_payload->>'public', '') ILIKE 'true')`
		currentSQL = `
			SELECT id::text
			FROM findings
			WHERE status IN ('open','in_progress')
			  AND (source_vendor ILIKE '%aws%' OR source_vendor ILIKE '%azure%' OR source_vendor ILIKE '%gcp%' OR COALESCE(raw_payload->>'cloud_provider', '') <> '')
			  AND (title ILIKE '%public%' OR COALESCE(raw_payload->>'public', '') ILIKE 'true')`
	default:
		return nil, nil, 0, fmt.Errorf("unsupported new entity metric %q", rule.Metric)
	}

	rows, err := tx.Query(ctx, candidateSQL, since)
	if err != nil {
		return nil, nil, 0, err
	}

	matches := []Match{}
	scopes := map[string]bool{}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, nil, 0, err
		}
		match := Match{Scope: id, Title: fmt.Sprintf("%s: %s", rule.Name, name), Context: map[string]any{"entity_id": id, "entity_name": name, "metric": rule.Metric}}
		matches = append(matches, match)
		scopes[Fingerprint(rule.ID, id)] = true
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, nil, 0, err
	}
	rows.Close()

	currentRows, err := tx.Query(ctx, currentSQL)
	if err != nil {
		return nil, nil, 0, err
	}
	defer currentRows.Close()
	for currentRows.Next() {
		var id string
		if err := currentRows.Scan(&id); err != nil {
			return nil, nil, 0, err
		}
		scopes[Fingerprint(rule.ID, id)] = true
	}
	if err := currentRows.Err(); err != nil {
		return nil, nil, 0, err
	}
	return matches, scopes, float64(len(matches)), nil
}

func slaMatches(ctx context.Context, tx pgx.Tx, rule Rule) ([]Match, map[string]bool, float64, error) {
	rows, err := tx.Query(ctx, `
		SELECT id::text, title, affected_asset, severity::text,
		       EXTRACT(EPOCH FROM (NOW() - first_seen)) / 86400
		FROM findings
		WHERE status IN ('open','in_progress') AND (
			(severity = 'critical' AND first_seen < NOW() - INTERVAL '7 days') OR
			(severity = 'high' AND first_seen < NOW() - INTERVAL '30 days') OR
			(severity = 'medium' AND first_seen < NOW() - INTERVAL '90 days')
		)`)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()
	matches := []Match{}
	scopes := map[string]bool{}
	for rows.Next() {
		var id, title, asset, severity string
		var age float64
		if err := rows.Scan(&id, &title, &asset, &severity, &age); err != nil {
			return nil, nil, 0, err
		}
		match := Match{
			Scope:   id,
			Title:   fmt.Sprintf("%s: %s", rule.Name, title),
			Context: map[string]any{"finding_id": id, "finding_title": title, "asset": asset, "severity": severity, "age_days": int(math.Round(age))},
		}
		matches = append(matches, match)
		scopes[Fingerprint(rule.ID, id)] = true
	}
	return matches, scopes, float64(len(matches)), rows.Err()
}

func metricValue(ctx context.Context, tx pgx.Tx, metric string) (float64, error) {
	var value float64
	switch metric {
	case "open_critical_count":
		err := tx.QueryRow(ctx, `SELECT COUNT(*)::float8 FROM findings WHERE status IN ('open','in_progress') AND severity = 'critical'`).Scan(&value)
		return value, err
	case "mfa_coverage_pct":
		err := tx.QueryRow(ctx, `SELECT COALESCE(100.0 * COUNT(*) FILTER (WHERE mfa_enabled) / NULLIF(COUNT(*), 0), 100)::float8 FROM identity_users`).Scan(&value)
		return value, err
	case "public_exposure_count":
		err := tx.QueryRow(ctx, `
			SELECT COUNT(*)::float8 FROM findings
			 WHERE status IN ('open','in_progress')
			   AND (source_vendor ILIKE '%aws%' OR source_vendor ILIKE '%azure%' OR source_vendor ILIKE '%gcp%' OR COALESCE(raw_payload->>'cloud_provider', '') <> '')
			   AND (title ILIKE '%public%' OR COALESCE(raw_payload->>'public', '') ILIKE 'true')`).Scan(&value)
		return value, err
	case "posture_score":
		scores, err := scoring.NewFromQuerier(tx).CurrentScores(ctx)
		return float64(scores.Overall), err
	default:
		return 0, fmt.Errorf("unsupported metric %q", metric)
	}
}

func loadRules(ctx context.Context, tx pgx.Tx, trigger Trigger) ([]Rule, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, name, category, condition_type, metric, operator, threshold_value,
		       delta_direction, delta_amount, severity::text, evaluation_trigger, cooldown_minutes
		  FROM alert_rules
		 WHERE enabled
		   AND ($1 = 'manual' OR evaluation_trigger = $1 OR evaluation_trigger = 'both')
		 ORDER BY created_at, name`, trigger)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	rules := []Rule{}
	for rows.Next() {
		var rule Rule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Category, &rule.ConditionType, &rule.Metric, &rule.Operator, &rule.ThresholdValue, &rule.DeltaDirection, &rule.DeltaAmount, &rule.Severity, &rule.EvaluationTrigger, &rule.CooldownMinutes); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func loadState(ctx context.Context, tx pgx.Tx, ruleID uuid.UUID) (float64, time.Time, bool, error) {
	var value *float64
	var evaluated time.Time
	err := tx.QueryRow(ctx, `SELECT last_metric_value, last_evaluated_at FROM alert_rule_state WHERE rule_id = $1`, ruleID).Scan(&value, &evaluated)
	if err == pgx.ErrNoRows {
		return 0, time.Time{}, false, nil
	}
	if err != nil {
		return 0, time.Time{}, false, err
	}
	if value == nil {
		return 0, evaluated, true, nil
	}
	return *value, evaluated, true, nil
}

func loadActiveFingerprints(ctx context.Context, tx pgx.Tx, ruleID uuid.UUID) (map[string]bool, error) {
	rows, err := tx.Query(ctx, `SELECT fingerprint FROM alerts WHERE rule_id = $1 AND status IN ('firing','acknowledged')`, ruleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	active := map[string]bool{}
	for rows.Next() {
		var fingerprint string
		if err := rows.Scan(&fingerprint); err != nil {
			return nil, err
		}
		active[fingerprint] = true
	}
	return active, rows.Err()
}

func Fingerprint(ruleID uuid.UUID, scope string) string {
	sum := sha256.Sum256([]byte(ruleID.String() + ":" + scope))
	return hex.EncodeToString(sum[:])
}

func Compare(value float64, operator string, threshold float64) bool {
	switch operator {
	case "gt":
		return value > threshold
	case "lt":
		return value < threshold
	case "gte":
		return value >= threshold
	case "lte":
		return value <= threshold
	case "eq":
		return value == threshold
	default:
		return false
	}
}

func operatorText(operator string) string {
	return map[string]string{"gt": ">", "lt": "<", "gte": "≥", "lte": "≤", "eq": "="}[operator]
}
