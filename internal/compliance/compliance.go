package compliance

import (
	"context"
	"math"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seccomply/seccomply/internal/scoring"
)

type BlockingIssue struct {
	GapType string `json:"gap_type"`
	Label   string `json:"label"`
	Count   int    `json:"count"`
	Href    string `json:"href"`
}

type Control struct {
	Code           string          `json:"code"`
	Title          string          `json:"title"`
	Summary        string          `json:"summary"`
	Status         string          `json:"status"`
	BlockingIssues []BlockingIssue `json:"blocking_issues"`
}

type Framework struct {
	Name            string    `json:"name"`
	Readiness       int       `json:"readiness"`
	PassingControls int       `json:"passing_controls"`
	TotalControls   int       `json:"total_controls"`
	ControlsAway    int       `json:"controls_away"`
	AuditReady      bool      `json:"audit_ready"`
	Controls        []Control `json:"controls"`
}

type Summary struct {
	Frameworks       []Framework `json:"frameworks"`
	OverallReadiness int         `json:"overall_readiness"`
	PassingControls  int         `json:"passing_controls"`
	TotalControls    int         `json:"total_controls"`
}

type mapping struct {
	Framework       string
	ControlCode     string
	ControlTitle    string
	ControlSummary  string
	GapType         string
	RemediationText string
	RemediationHref string
}

type Service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) GetSummary(ctx context.Context) (Summary, error) {
	metrics, err := scoring.New(s.db).CurrentMetrics(ctx)
	if err != nil {
		return Summary{}, err
	}
	rows, err := s.db.Query(ctx, `
		SELECT framework, control_code, control_title, control_summary,
		       gap_type, remediation_label, remediation_href
		FROM control_mappings
		ORDER BY CASE framework WHEN 'SOC 2' THEN 1 WHEN 'ISO 27001' THEN 2 ELSE 3 END,
		         control_code, gap_type`)
	if err != nil {
		return Summary{}, err
	}
	defer rows.Close()

	mappings := []mapping{}
	for rows.Next() {
		var item mapping
		if err := rows.Scan(&item.Framework, &item.ControlCode, &item.ControlTitle, &item.ControlSummary, &item.GapType, &item.RemediationText, &item.RemediationHref); err != nil {
			return Summary{}, err
		}
		mappings = append(mappings, item)
	}
	if err := rows.Err(); err != nil {
		return Summary{}, err
	}
	return evaluate(metrics, mappings), nil
}

func evaluate(metrics scoring.Metrics, mappings []mapping) Summary {
	type frameworkBuilder struct {
		name     string
		controls map[string]*Control
	}
	order := []string{"SOC 2", "ISO 27001", "DPDPA"}
	builders := map[string]*frameworkBuilder{}
	for _, name := range order {
		builders[name] = &frameworkBuilder{name: name, controls: map[string]*Control{}}
	}

	for _, item := range mappings {
		builder := builders[item.Framework]
		if builder == nil {
			builder = &frameworkBuilder{name: item.Framework, controls: map[string]*Control{}}
			builders[item.Framework] = builder
			order = append(order, item.Framework)
		}
		control := builder.controls[item.ControlCode]
		if control == nil {
			control = &Control{
				Code:           item.ControlCode,
				Title:          item.ControlTitle,
				Summary:        item.ControlSummary,
				Status:         "passing",
				BlockingIssues: []BlockingIssue{},
			}
			builder.controls[item.ControlCode] = control
		}
		if count := scoring.GapCount(metrics, item.GapType); count > 0 {
			control.Status = "failing"
			control.BlockingIssues = append(control.BlockingIssues, BlockingIssue{
				GapType: item.GapType,
				Label:   item.RemediationText,
				Count:   count,
				Href:    item.RemediationHref,
			})
		}
	}

	result := Summary{Frameworks: []Framework{}}
	for _, name := range order {
		builder := builders[name]
		if builder == nil || len(builder.controls) == 0 {
			continue
		}
		codes := make([]string, 0, len(builder.controls))
		for code := range builder.controls {
			codes = append(codes, code)
		}
		sort.Strings(codes)

		framework := Framework{Name: name, TotalControls: len(codes), Controls: []Control{}}
		for _, code := range codes {
			control := *builder.controls[code]
			if control.Status == "passing" {
				framework.PassingControls++
			}
			framework.Controls = append(framework.Controls, control)
		}
		framework.ControlsAway = framework.TotalControls - framework.PassingControls
		framework.Readiness = int(math.Round(float64(framework.PassingControls) / float64(framework.TotalControls) * 100))
		framework.AuditReady = framework.ControlsAway == 0
		result.PassingControls += framework.PassingControls
		result.TotalControls += framework.TotalControls
		result.Frameworks = append(result.Frameworks, framework)
	}
	if result.TotalControls > 0 {
		result.OverallReadiness = int(math.Round(float64(result.PassingControls) / float64(result.TotalControls) * 100))
	}
	return result
}
