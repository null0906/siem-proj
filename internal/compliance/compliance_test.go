package compliance

import (
	"testing"

	"github.com/seccomply/seccomply/internal/scoring"
)

func TestEvaluateMapsLiveGapsToControls(t *testing.T) {
	summary := evaluate(scoring.Metrics{AccountsWithoutMFA: 14}, []mapping{
		{Framework: "SOC 2", ControlCode: "CC6.1", ControlTitle: "Access", GapType: "accounts_without_mfa", RemediationText: "accounts lack MFA", RemediationHref: "/identity"},
		{Framework: "SOC 2", ControlCode: "CC7.2", ControlTitle: "Monitoring", GapType: "devices_without_edr", RemediationText: "devices lack EDR", RemediationHref: "/assets"},
	})
	if summary.OverallReadiness != 50 {
		t.Fatalf("overall readiness = %d, want 50", summary.OverallReadiness)
	}
	framework := summary.Frameworks[0]
	if framework.ControlsAway != 1 || framework.Controls[0].BlockingIssues[0].Count != 14 {
		t.Fatalf("unexpected framework result: %+v", framework)
	}
}
