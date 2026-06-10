package scoring

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Metrics struct {
	OpenCritical              int
	OpenHigh                  int
	SLABreached               int
	ExploitableUnpatched      int
	TotalAccounts             int
	AccountsWithoutMFA        int
	PrivilegedWithoutMFA      int
	DormantPrivileged         int
	TotalDevices              int
	UnencryptedDevices        int
	DevicesWithoutEDR         int
	NonCompliantDevices       int
	PublicExposures           int
	CriticalCloudFindings     int
	IAMOverPermissions        int
	ComplianceFrameworkScores []float64
	ComplianceMappings        []ControlMapping
}

type ControlMapping struct {
	Framework   string
	ControlCode string
	GapType     string
}

type Category struct {
	Name       string `json:"name"`
	Score      int    `json:"score"`
	Delta      int    `json:"delta"`
	IssueCount int    `json:"issue_count"`
	PointDrag  int    `json:"point_drag"`
}

type Snapshot struct {
	Date          string `json:"date"`
	Overall       int    `json:"overall"`
	Vulnerability int    `json:"vulnerability"`
	Identity      int    `json:"identity"`
	Endpoint      int    `json:"endpoint"`
	Cloud         int    `json:"cloud"`
	Compliance    int    `json:"compliance"`
}

type Summary struct {
	Overall     int        `json:"overall"`
	Grade       string     `json:"grade"`
	Categories  []Category `json:"categories"`
	Trend       []Snapshot `json:"trend"`
	BiggestDrag Category   `json:"biggest_drag"`
	Insight     string     `json:"insight"`
}

type Scores struct {
	Overall       int
	Vulnerability int
	Identity      int
	Endpoint      int
	Cloud         int
	Compliance    int
}

type Service struct {
	db Querier
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

type Querier interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func NewFromQuerier(db Querier) *Service {
	return &Service{db: db}
}

func (s *Service) CurrentScores(ctx context.Context) (Scores, error) {
	metrics, err := s.CurrentMetrics(ctx)
	if err != nil {
		return Scores{}, err
	}
	return Calculate(metrics), nil
}

func (s *Service) CurrentMetrics(ctx context.Context) (Metrics, error) {
	return s.loadMetrics(ctx)
}

func Calculate(metrics Metrics) Scores {
	vulnerability := clamp(100 -
		2*metrics.OpenCritical -
		metrics.OpenHigh -
		3*metrics.SLABreached -
		2*metrics.ExploitableUnpatched)

	withoutMFAPercent := 0
	if metrics.TotalAccounts > 0 {
		withoutMFAPercent = int(math.Round(float64(metrics.AccountsWithoutMFA) / float64(metrics.TotalAccounts) * 100))
	}
	identity := clamp(100 - withoutMFAPercent - 3*metrics.PrivilegedWithoutMFA - 2*metrics.DormantPrivileged)

	unencryptedPercent := percent(metrics.UnencryptedDevices, metrics.TotalDevices)
	withoutEDRPercent := percent(metrics.DevicesWithoutEDR, metrics.TotalDevices)
	endpoint := clamp(int(math.Round(100 -
		unencryptedPercent*0.5 -
		withoutEDRPercent*0.7 -
		float64(2*metrics.NonCompliantDevices))))

	cloud := clamp(100 -
		3*metrics.PublicExposures -
		2*metrics.CriticalCloudFindings -
		metrics.IAMOverPermissions)

	complianceScores := metrics.ComplianceFrameworkScores
	if len(metrics.ComplianceMappings) > 0 {
		complianceScores = FrameworkScores(metrics, metrics.ComplianceMappings)
	}
	compliance := 100
	if len(complianceScores) > 0 {
		var total float64
		for _, score := range complianceScores {
			total += score
		}
		compliance = clamp(int(math.Round(total / float64(len(complianceScores)))))
	}

	overall := int(math.Round(
		float64(vulnerability)*0.25 +
			float64(identity)*0.20 +
			float64(endpoint)*0.15 +
			float64(cloud)*0.20 +
			float64(compliance)*0.20,
	))

	return Scores{
		Overall:       overall,
		Vulnerability: vulnerability,
		Identity:      identity,
		Endpoint:      endpoint,
		Cloud:         cloud,
		Compliance:    compliance,
	}
}

func Grade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

func (s *Service) GetSummary(ctx context.Context) (Summary, error) {
	metrics, err := s.loadMetrics(ctx)
	if err != nil {
		return Summary{}, err
	}
	scores := Calculate(metrics)
	if err := s.storeSnapshot(ctx, time.Now(), scores); err != nil {
		return Summary{}, err
	}
	trend, err := s.loadTrend(ctx, 30)
	if err != nil {
		return Summary{}, err
	}

	weekAgo := scores
	for _, point := range trend {
		date, parseErr := time.Parse("2006-01-02", point.Date)
		if parseErr == nil && !date.After(time.Now().AddDate(0, 0, -7)) {
			weekAgo = Scores{
				Overall:       point.Overall,
				Vulnerability: point.Vulnerability,
				Identity:      point.Identity,
				Endpoint:      point.Endpoint,
				Cloud:         point.Cloud,
				Compliance:    point.Compliance,
			}
		}
	}

	categories := []Category{
		{Name: "Vulnerability", Score: scores.Vulnerability, Delta: scores.Vulnerability - weekAgo.Vulnerability, IssueCount: metrics.OpenCritical + metrics.OpenHigh + metrics.SLABreached + metrics.ExploitableUnpatched, PointDrag: 100 - scores.Vulnerability},
		{Name: "Identity", Score: scores.Identity, Delta: scores.Identity - weekAgo.Identity, IssueCount: metrics.AccountsWithoutMFA + metrics.PrivilegedWithoutMFA + metrics.DormantPrivileged, PointDrag: 100 - scores.Identity},
		{Name: "Endpoint", Score: scores.Endpoint, Delta: scores.Endpoint - weekAgo.Endpoint, IssueCount: metrics.UnencryptedDevices + metrics.DevicesWithoutEDR + metrics.NonCompliantDevices, PointDrag: 100 - scores.Endpoint},
		{Name: "Cloud", Score: scores.Cloud, Delta: scores.Cloud - weekAgo.Cloud, IssueCount: metrics.PublicExposures + metrics.CriticalCloudFindings + metrics.IAMOverPermissions, PointDrag: 100 - scores.Cloud},
		{Name: "Compliance", Score: scores.Compliance, Delta: scores.Compliance - weekAgo.Compliance, IssueCount: FailingControlCount(metrics, metrics.ComplianceMappings), PointDrag: 100 - scores.Compliance},
	}
	biggest := categories[0]
	for _, category := range categories[1:] {
		if category.Score < biggest.Score {
			biggest = category
		}
	}

	return Summary{
		Overall:     scores.Overall,
		Grade:       Grade(scores.Overall),
		Categories:  categories,
		Trend:       trend,
		BiggestDrag: biggest,
		Insight:     fmt.Sprintf("Your biggest drag is %s — %d issues pulling it down %d points.", biggest.Name, biggest.IssueCount, biggest.PointDrag),
	}, nil
}

func (s *Service) loadMetrics(ctx context.Context) (Metrics, error) {
	var metrics Metrics
	err := s.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE source_tool = 'scanner' AND status IN ('open','in_progress') AND severity = 'critical'),
			COUNT(*) FILTER (WHERE source_tool = 'scanner' AND status IN ('open','in_progress') AND severity = 'high'),
			COUNT(*) FILTER (WHERE source_tool = 'scanner' AND status IN ('open','in_progress') AND severity IN ('critical', 'high') AND first_seen < NOW() - INTERVAL '30 days'),
			COUNT(*) FILTER (WHERE source_tool = 'scanner' AND status IN ('open','in_progress') AND (
				COALESCE(raw_payload->>'Exploitable', '') ILIKE 'true'
				OR title ILIKE '%remote code execution%'
				OR title ILIKE '%authentication bypass%'
			))
		FROM findings`).Scan(
		&metrics.OpenCritical,
		&metrics.OpenHigh,
		&metrics.SLABreached,
		&metrics.ExploitableUnpatched,
	)
	if err != nil {
		return Metrics{}, err
	}

	err = s.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE NOT mfa_enabled),
			COUNT(*) FILTER (WHERE is_privileged AND NOT mfa_enabled),
			COUNT(*) FILTER (WHERE is_privileged AND (account_status = 'dormant' OR last_login < NOW() - INTERVAL '90 days'))
		FROM identity_users`).Scan(
		&metrics.TotalAccounts,
		&metrics.AccountsWithoutMFA,
		&metrics.PrivilegedWithoutMFA,
		&metrics.DormantPrivileged,
	)
	if err != nil {
		return Metrics{}, err
	}

	err = s.db.QueryRow(ctx, `
		SELECT
			COUNT(DISTINCT NULLIF(affected_asset, '')) FILTER (WHERE source_tool = 'edr'),
			COUNT(DISTINCT NULLIF(affected_asset, '')) FILTER (WHERE source_tool = 'edr' AND COALESCE(raw_payload->>'encrypted', '') ILIKE 'false'),
			COUNT(DISTINCT NULLIF(affected_asset, '')) FILTER (WHERE source_tool = 'edr' AND COALESCE(raw_payload->>'has_edr', '') ILIKE 'false'),
			COUNT(DISTINCT NULLIF(affected_asset, '')) FILTER (WHERE source_tool = 'edr' AND status IN ('open','in_progress') AND severity IN ('critical', 'high'))
		FROM findings`).Scan(
		&metrics.TotalDevices,
		&metrics.UnencryptedDevices,
		&metrics.DevicesWithoutEDR,
		&metrics.NonCompliantDevices,
	)
	if err != nil {
		return Metrics{}, err
	}

	err = s.db.QueryRow(ctx, `
		WITH cloud_findings AS (
			SELECT *
			FROM findings
			WHERE source_vendor ILIKE '%aws%'
			   OR source_vendor ILIKE '%azure%'
			   OR source_vendor ILIKE '%gcp%'
			   OR COALESCE(raw_payload->>'cloud_provider', '') <> ''
		)
		SELECT
			COUNT(*) FILTER (WHERE status IN ('open','in_progress') AND (title ILIKE '%public%' OR COALESCE(raw_payload->>'public', '') ILIKE 'true')),
			COUNT(*) FILTER (WHERE status IN ('open','in_progress') AND severity = 'critical'),
			COUNT(*) FILTER (WHERE status IN ('open','in_progress') AND (title ILIKE '%over-permission%' OR title ILIKE '%over permission%'))
		FROM cloud_findings`).Scan(
		&metrics.PublicExposures,
		&metrics.CriticalCloudFindings,
		&metrics.IAMOverPermissions,
	)
	if err != nil {
		return Metrics{}, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT framework, control_code, gap_type
		FROM control_mappings
		ORDER BY framework, control_code, gap_type`)
	if err != nil {
		return Metrics{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var mapping ControlMapping
		if err := rows.Scan(&mapping.Framework, &mapping.ControlCode, &mapping.GapType); err != nil {
			return Metrics{}, err
		}
		metrics.ComplianceMappings = append(metrics.ComplianceMappings, mapping)
	}
	metrics.ComplianceFrameworkScores = FrameworkScores(metrics, metrics.ComplianceMappings)
	return metrics, rows.Err()
}

func GapCount(metrics Metrics, gapType string) int {
	switch gapType {
	case "accounts_without_mfa":
		return metrics.AccountsWithoutMFA
	case "privileged_without_mfa":
		return metrics.PrivilegedWithoutMFA
	case "dormant_privileged":
		return metrics.DormantPrivileged
	case "unencrypted_devices":
		return metrics.UnencryptedDevices
	case "devices_without_edr":
		return metrics.DevicesWithoutEDR
	case "noncompliant_devices":
		return metrics.NonCompliantDevices
	case "sla_breached_criticals":
		return metrics.SLABreached
	case "open_critical_vulnerabilities":
		return metrics.OpenCritical
	case "exploitable_unpatched":
		return metrics.ExploitableUnpatched
	case "public_cloud_exposure":
		return metrics.PublicExposures
	case "critical_cloud_findings":
		return metrics.CriticalCloudFindings
	case "iam_overpermission":
		return metrics.IAMOverPermissions
	default:
		return 0
	}
}

func FrameworkScores(metrics Metrics, mappings []ControlMapping) []float64 {
	type controlState struct {
		failing bool
	}
	frameworks := map[string]map[string]*controlState{}
	for _, mapping := range mappings {
		controls := frameworks[mapping.Framework]
		if controls == nil {
			controls = map[string]*controlState{}
			frameworks[mapping.Framework] = controls
		}
		state := controls[mapping.ControlCode]
		if state == nil {
			state = &controlState{}
			controls[mapping.ControlCode] = state
		}
		if GapCount(metrics, mapping.GapType) > 0 {
			state.failing = true
		}
	}

	scores := make([]float64, 0, len(frameworks))
	for _, controls := range frameworks {
		passing := 0
		for _, state := range controls {
			if !state.failing {
				passing++
			}
		}
		if len(controls) > 0 {
			scores = append(scores, float64(passing)/float64(len(controls))*100)
		}
	}
	return scores
}

func FailingControlCount(metrics Metrics, mappings []ControlMapping) int {
	failing := map[string]bool{}
	for _, mapping := range mappings {
		if GapCount(metrics, mapping.GapType) > 0 {
			failing[mapping.Framework+"|"+mapping.ControlCode] = true
		}
	}
	return len(failing)
}

func (s *Service) storeSnapshot(ctx context.Context, date time.Time, scores Scores) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO posture_snapshots
			(snapshot_date, overall_score, vulnerability_score, identity_score, endpoint_score, cloud_score, compliance_score)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (snapshot_date) DO UPDATE SET
			overall_score = EXCLUDED.overall_score,
			vulnerability_score = EXCLUDED.vulnerability_score,
			identity_score = EXCLUDED.identity_score,
			endpoint_score = EXCLUDED.endpoint_score,
			cloud_score = EXCLUDED.cloud_score,
			compliance_score = EXCLUDED.compliance_score`,
		date.Format("2006-01-02"), scores.Overall, scores.Vulnerability, scores.Identity,
		scores.Endpoint, scores.Cloud, scores.Compliance,
	)
	return err
}

func (s *Service) loadTrend(ctx context.Context, days int) ([]Snapshot, error) {
	rows, err := s.db.Query(ctx, `
		SELECT snapshot_date::text, overall_score, vulnerability_score, identity_score,
		       endpoint_score, cloud_score, compliance_score
		FROM posture_snapshots
		WHERE snapshot_date >= CURRENT_DATE - ($1::int - 1)
		ORDER BY snapshot_date`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trend := []Snapshot{}
	for rows.Next() {
		var snapshot Snapshot
		if err := rows.Scan(&snapshot.Date, &snapshot.Overall, &snapshot.Vulnerability, &snapshot.Identity, &snapshot.Endpoint, &snapshot.Cloud, &snapshot.Compliance); err != nil {
			return nil, err
		}
		trend = append(trend, snapshot)
	}
	return trend, rows.Err()
}

func percent(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func clamp(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}
