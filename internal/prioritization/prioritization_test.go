package prioritization

import (
	"testing"

	"github.com/seccomply/seccomply/internal/scoring"
)

func TestCandidateImpactUsesScoringDelta(t *testing.T) {
	metrics := scoring.Metrics{
		TotalAccounts:        100,
		AccountsWithoutMFA:   20,
		PrivilegedWithoutMFA: 4,
	}
	current := scoring.Calculate(metrics).Overall
	candidates := buildCandidates(metrics)
	var found candidate
	for _, item := range candidates {
		if item.ID == "enable-mfa" {
			found = item
			break
		}
	}
	after := metrics
	found.apply(&after)
	impact := scoring.Calculate(after).Overall - current
	if impact <= 0 {
		t.Fatalf("expected positive scoring impact, got %d", impact)
	}
	if after.AccountsWithoutMFA != 0 || after.PrivilegedWithoutMFA != 0 {
		t.Fatal("enable MFA action did not clear MFA gaps")
	}
}

func TestRoundOne(t *testing.T) {
	if got := roundOne(2.666); got != 2.7 {
		t.Fatalf("roundOne = %v, want 2.7", got)
	}
}
