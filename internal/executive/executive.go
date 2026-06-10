package executive

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seccomply/seccomply/internal/assets"
	"github.com/seccomply/seccomply/internal/compliance"
	"github.com/seccomply/seccomply/internal/scoring"
)

type Posture struct {
	Score      int                `json:"score"`
	Grade      string             `json:"grade"`
	StartScore int                `json:"start_score"`
	Delta      int                `json:"delta"`
	Trend      []scoring.Snapshot `json:"trend"`
}

type Changes struct {
	NewFindings  int      `json:"new_findings"`
	Resolved     int      `json:"resolved"`
	NewCriticals int      `json:"new_criticals"`
	ScoreDelta   int      `json:"score_delta"`
	Highlights   []string `json:"highlights"`
}

type Risk struct {
	Rank          int    `json:"rank"`
	Title         string `json:"title"`
	Summary       string `json:"summary"`
	RiskScore     int    `json:"risk_score"`
	CriticalCount int    `json:"critical_count"`
	SourceCount   int    `json:"source_count"`
}

type Framework struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
	State string `json:"state"`
}

type Summary struct {
	GeneratedAt      time.Time   `json:"generated_at"`
	PeriodDays       int         `json:"period_days"`
	PeriodLabel      string      `json:"period_label"`
	Posture          Posture     `json:"posture"`
	Changes          Changes     `json:"changes"`
	TopRisks         []Risk      `json:"top_risks"`
	Frameworks       []Framework `json:"frameworks"`
	ExecutiveMessage string      `json:"executive_message"`
}

type Service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) GetSummary(ctx context.Context, days int) (Summary, error) {
	days = NormalizeDays(days)
	postureSummary, err := scoring.New(s.db).GetSummary(ctx)
	if err != nil {
		return Summary{}, err
	}
	trend, err := s.loadTrend(ctx, days)
	if err != nil {
		return Summary{}, err
	}

	startScore := postureSummary.Overall
	if len(trend) > 0 {
		startScore = trend[0].Overall
	}
	scoreDelta := postureSummary.Overall - startScore

	changes, err := s.loadChanges(ctx, days)
	if err != nil {
		return Summary{}, err
	}
	changes.ScoreDelta = scoreDelta
	changes.Highlights = changeHighlights(changes, days)

	inventory, err := assets.New(s.db).Inventory(ctx)
	if err != nil {
		return Summary{}, err
	}
	risks := businessRisks(inventory.Assets, 5)
	complianceSummary, err := compliance.New(s.db).GetSummary(ctx)
	if err != nil {
		return Summary{}, err
	}
	frameworks := make([]Framework, 0, len(complianceSummary.Frameworks))
	for _, item := range complianceSummary.Frameworks {
		state := "Building readiness"
		if item.Readiness >= 85 {
			state = "Near audit-ready"
		} else if item.Readiness >= 70 {
			state = "Material progress"
		}
		frameworks = append(frameworks, Framework{Name: item.Name, Score: item.Readiness, State: state})
	}

	return Summary{
		GeneratedAt: time.Now().UTC(),
		PeriodDays:  days,
		PeriodLabel: PeriodLabel(days),
		Posture: Posture{
			Score:      postureSummary.Overall,
			Grade:      postureSummary.Grade,
			StartScore: startScore,
			Delta:      scoreDelta,
			Trend:      trend,
		},
		Changes:          changes,
		TopRisks:         risks,
		Frameworks:       frameworks,
		ExecutiveMessage: executiveMessage(postureSummary.Overall, scoreDelta, len(risks)),
	}, nil
}

func NormalizeDays(days int) int {
	switch days {
	case 7, 30, 60:
		return days
	default:
		return 30
	}
}

func PeriodLabel(days int) string {
	switch NormalizeDays(days) {
	case 7:
		return "last 7 days"
	case 60:
		return "last 60 days"
	default:
		return "last 30 days"
	}
}

func (s *Service) loadChanges(ctx context.Context, days int) (Changes, error) {
	var changes Changes
	err := s.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE ingested_at >= NOW() - ($1::int * INTERVAL '1 day')),
			COUNT(*) FILTER (WHERE status = 'resolved' AND last_seen >= NOW() - ($1::int * INTERVAL '1 day')),
			COUNT(*) FILTER (WHERE severity = 'critical' AND first_seen >= NOW() - ($1::int * INTERVAL '1 day'))
		FROM findings`, days).Scan(&changes.NewFindings, &changes.Resolved, &changes.NewCriticals)
	return changes, err
}

func (s *Service) loadTrend(ctx context.Context, days int) ([]scoring.Snapshot, error) {
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

	trend := []scoring.Snapshot{}
	for rows.Next() {
		var point scoring.Snapshot
		if err := rows.Scan(&point.Date, &point.Overall, &point.Vulnerability, &point.Identity, &point.Endpoint, &point.Cloud, &point.Compliance); err != nil {
			return nil, err
		}
		trend = append(trend, point)
	}
	return trend, rows.Err()
}

func businessRisks(items []assets.Asset, limit int) []Risk {
	if limit > len(items) {
		limit = len(items)
	}
	risks := make([]Risk, 0, limit)
	for index, asset := range items[:limit] {
		title := "Business system needs attention"
		switch {
		case asset.IsPublic:
			title = "Internet-exposed business system"
		case !asset.HasEDR && !asset.IsEncrypted:
			title = "Business system lacks core protection"
		case !asset.HasEDR:
			title = "Business system lacks active monitoring"
		case asset.CriticalCount > 0:
			title = "Critical weaknesses on a business system"
		}

		parts := []string{
			fmt.Sprintf("Risk score %d/100", asset.RiskScore),
			fmt.Sprintf("Observed %d security signals across %d sources", asset.FindingCount, asset.SourceCount),
		}
		if asset.CriticalCount > 0 {
			if asset.CriticalCount == 1 {
				parts = append(parts, "1 issue requires immediate attention")
			} else {
				parts = append(parts, fmt.Sprintf("%d issues require immediate attention", asset.CriticalCount))
			}
		}
		if asset.IsPublic {
			parts = append(parts, "Internet exposure increases business risk")
		}
		if !asset.HasEDR {
			parts = append(parts, "Active monitoring is not in place")
		}
		if !asset.IsEncrypted {
			parts = append(parts, "Data protection is incomplete")
		}

		risks = append(risks, Risk{
			Rank:          index + 1,
			Title:         title,
			Summary:       strings.Join(parts, ". ") + ".",
			RiskScore:     asset.RiskScore,
			CriticalCount: asset.CriticalCount,
			SourceCount:   asset.SourceCount,
		})
	}
	return risks
}

func readiness(scores scoring.Scores) []Framework {
	return []Framework{
		framework("SOC 2", weighted(scores.Identity, 30, scores.Endpoint, 25, scores.Cloud, 25, scores.Vulnerability, 20)),
		framework("ISO 27001", weighted(scores.Identity, 25, scores.Endpoint, 25, scores.Cloud, 20, scores.Vulnerability, 20, scores.Compliance, 10)),
		framework("DPDPA", weighted(scores.Identity, 30, scores.Endpoint, 35, scores.Cloud, 25, scores.Vulnerability, 10)),
	}
}

func weighted(values ...int) int {
	var total, weight int
	for index := 0; index+1 < len(values); index += 2 {
		total += values[index] * values[index+1]
		weight += values[index+1]
	}
	if weight == 0 {
		return 0
	}
	return int(math.Round(float64(total) / float64(weight)))
}

func framework(name string, score int) Framework {
	state := "Building readiness"
	if score >= 85 {
		state = "Near audit-ready"
	} else if score >= 70 {
		state = "Material progress"
	}
	return Framework{Name: name, Score: score, State: state}
}

func changeHighlights(changes Changes, days int) []string {
	period := PeriodLabel(days)
	highlights := []string{
		fmt.Sprintf("%d new security findings were identified in the %s.", changes.NewFindings, period),
		fmt.Sprintf("%d findings were resolved or closed.", changes.Resolved),
	}
	if changes.NewCriticals > 0 {
		highlights = append(highlights, fmt.Sprintf("%d newly identified issues require immediate attention.", changes.NewCriticals))
	} else {
		highlights = append(highlights, "No newly identified issues require immediate attention.")
	}
	switch {
	case changes.ScoreDelta > 0:
		highlights = append(highlights, fmt.Sprintf("Overall security posture improved by %d points.", changes.ScoreDelta))
	case changes.ScoreDelta < 0:
		highlights = append(highlights, fmt.Sprintf("Overall security posture declined by %d points.", -changes.ScoreDelta))
	default:
		highlights = append(highlights, "Overall security posture remained stable.")
	}
	return highlights
}

func executiveMessage(score, delta, riskCount int) string {
	direction := "remained stable"
	if delta > 0 {
		direction = fmt.Sprintf("improved by %d points", delta)
	} else if delta < 0 {
		direction = fmt.Sprintf("declined by %d points", -delta)
	}
	return fmt.Sprintf("Security posture is %d/100 and has %s. Leadership attention should remain focused on the %d highest-priority business risks.", score, direction, riskCount)
}
