package prioritization

import (
	"context"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seccomply/seccomply/internal/scoring"
)

type Action struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	ScoreImpact   int     `json:"score_impact"`
	Effort        string  `json:"effort"`
	EffortWeight  int     `json:"effort_weight"`
	Priority      float64 `json:"priority"`
	AffectedCount int     `json:"affected_count"`
	Source        string  `json:"source"`
	Category      string  `json:"category"`
	Href          string  `json:"href"`
}

type Queue struct {
	Actions        []Action `json:"actions"`
	CurrentScore   int      `json:"current_score"`
	ProjectedScore int      `json:"projected_score"`
	TopFiveGain    int      `json:"top_five_gain"`
}

type candidate struct {
	Action
	apply func(*scoring.Metrics)
}

type Service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Queue(ctx context.Context) (Queue, error) {
	metrics, err := scoring.New(s.db).CurrentMetrics(ctx)
	if err != nil {
		return Queue{}, err
	}
	current := scoring.Calculate(metrics).Overall
	candidates := buildCandidates(metrics)
	actions := make([]Action, 0, len(candidates))
	for i := range candidates {
		after := metrics
		candidates[i].apply(&after)
		candidates[i].ScoreImpact = scoring.Calculate(after).Overall - current
		if candidates[i].ScoreImpact <= 0 || candidates[i].AffectedCount <= 0 {
			continue
		}
		candidates[i].Priority = float64(candidates[i].ScoreImpact) / float64(candidates[i].EffortWeight)
		actions = append(actions, candidates[i].Action)
	}
	sort.SliceStable(actions, func(i, j int) bool {
		if actions[i].Priority == actions[j].Priority {
			if actions[i].ScoreImpact == actions[j].ScoreImpact {
				return actions[i].AffectedCount > actions[j].AffectedCount
			}
			return actions[i].ScoreImpact > actions[j].ScoreImpact
		}
		return actions[i].Priority > actions[j].Priority
	})
	for i := range actions {
		actions[i].Priority = roundOne(actions[i].Priority)
	}

	projectedMetrics := metrics
	for i := 0; i < len(actions) && i < 5; i++ {
		for _, item := range candidates {
			if item.ID == actions[i].ID {
				item.apply(&projectedMetrics)
				break
			}
		}
	}
	projected := scoring.Calculate(projectedMetrics).Overall
	return Queue{
		Actions:        actions,
		CurrentScore:   current,
		ProjectedScore: projected,
		TopFiveGain:    projected - current,
	}, nil
}

func buildCandidates(metrics scoring.Metrics) []candidate {
	return []candidate{
		makeCandidate("enable-mfa", "Enable MFA for accounts", "Enroll every account currently operating without MFA.", "low", 1, metrics.AccountsWithoutMFA, "Identity providers", "Identity", "/identity", func(m *scoring.Metrics) {
			m.AccountsWithoutMFA = 0
			m.PrivilegedWithoutMFA = 0
		}),
		makeCandidate("secure-privileged-accounts", "Secure privileged accounts without MFA", "Prioritize MFA enrollment for privileged identities.", "low", 1, metrics.PrivilegedWithoutMFA, "Identity providers", "Identity", "/identity", func(m *scoring.Metrics) {
			m.PrivilegedWithoutMFA = 0
			m.AccountsWithoutMFA = max(0, m.AccountsWithoutMFA-metrics.PrivilegedWithoutMFA)
		}),
		makeCandidate("review-dormant-privileged", "Review dormant privileged accounts", "Disable or validate privileged accounts inactive for more than 90 days.", "low", 1, metrics.DormantPrivileged, "Identity providers", "Identity", "/identity", func(m *scoring.Metrics) {
			m.DormantPrivileged = 0
		}),
		makeCandidate("remediate-critical-vulnerabilities", "Remediate open critical vulnerabilities", "Patch or mitigate the open critical vulnerability backlog.", "high", 3, metrics.OpenCritical, "Tenable", "Vulnerability", "/findings?severity=critical&source_tool=scanner", func(m *scoring.Metrics) {
			m.OpenCritical = 0
		}),
		makeCandidate("remediate-high-vulnerabilities", "Remediate open high vulnerabilities", "Reduce the open high-severity vulnerability backlog.", "high", 3, metrics.OpenHigh, "Tenable", "Vulnerability", "/findings?severity=high&source_tool=scanner", func(m *scoring.Metrics) {
			m.OpenHigh = 0
		}),
		makeCandidate("resolve-sla-breaches", "Resolve SLA-breached vulnerabilities", "Clear critical and high findings that have exceeded remediation SLA.", "medium", 2, metrics.SLABreached, "Tenable", "Vulnerability", "/findings?source_tool=scanner&status=open", func(m *scoring.Metrics) {
			m.SLABreached = 0
		}),
		makeCandidate("patch-exploitable", "Patch exploitable unpatched findings", "Remediate vulnerabilities with known exploitation paths.", "medium", 2, metrics.ExploitableUnpatched, "Tenable", "Vulnerability", "/findings?source_tool=scanner&status=open", func(m *scoring.Metrics) {
			m.ExploitableUnpatched = 0
		}),
		makeCandidate("deploy-edr", "Deploy EDR to uncovered devices", "Install and verify endpoint detection coverage.", "medium", 2, metrics.DevicesWithoutEDR, "Endpoint data", "Endpoint", "/assets", func(m *scoring.Metrics) {
			m.DevicesWithoutEDR = 0
		}),
		makeCandidate("encrypt-devices", "Encrypt unencrypted devices", "Enable full-disk encryption on uncovered endpoints.", "medium", 2, metrics.UnencryptedDevices, "Endpoint data", "Endpoint", "/assets", func(m *scoring.Metrics) {
			m.UnencryptedDevices = 0
		}),
		makeCandidate("remediate-noncompliant-devices", "Remediate non-compliant devices", "Resolve high-risk endpoint control failures.", "high", 3, metrics.NonCompliantDevices, "Endpoint data", "Endpoint", "/assets", func(m *scoring.Metrics) {
			m.NonCompliantDevices = 0
		}),
		makeCandidate("close-public-exposures", "Close public cloud exposures", "Remove unnecessary internet-facing cloud access.", "medium", 2, metrics.PublicExposures, "Cloud findings", "Cloud", "/findings?status=open", func(m *scoring.Metrics) {
			m.PublicExposures = 0
		}),
		makeCandidate("resolve-critical-cloud", "Resolve critical cloud findings", "Remediate critical cloud configuration risks.", "high", 3, metrics.CriticalCloudFindings, "Cloud findings", "Cloud", "/findings?severity=critical&status=open", func(m *scoring.Metrics) {
			m.CriticalCloudFindings = 0
		}),
		makeCandidate("reduce-iam-overpermission", "Reduce IAM over-permission", "Remove unnecessary cloud identity privileges.", "medium", 2, metrics.IAMOverPermissions, "Cloud findings", "Cloud", "/findings?status=open", func(m *scoring.Metrics) {
			m.IAMOverPermissions = 0
		}),
	}
}

func makeCandidate(id, title, description, effort string, effortWeight, count int, source, category, href string, apply func(*scoring.Metrics)) candidate {
	return candidate{
		Action: Action{
			ID: id, Title: title, Description: description, Effort: effort,
			EffortWeight: effortWeight, AffectedCount: count, Source: source,
			Category: category, Href: href,
		},
		apply: apply,
	}
}

func roundOne(value float64) float64 {
	return float64(int(value*10+0.5)) / 10
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
