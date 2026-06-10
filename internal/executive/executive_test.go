package executive

import (
	"strings"
	"testing"

	"github.com/seccomply/seccomply/internal/assets"
	"github.com/seccomply/seccomply/internal/scoring"
)

func TestBusinessRisksDoNotExposeAssetIdentifiers(t *testing.T) {
	risks := businessRisks([]assets.Asset{{
		DisplayName:   "SRV-WEB02",
		IPAddress:     "10.0.0.2",
		RiskScore:     92,
		FindingCount:  12,
		SourceCount:   3,
		CriticalCount: 2,
		IsPublic:      true,
	}}, 5)
	if len(risks) != 1 {
		t.Fatalf("len(risks) = %d, want 1", len(risks))
	}
	if strings.Contains(risks[0].Title, "SRV-WEB02") ||
		strings.Contains(risks[0].Summary, "SRV-WEB02") ||
		strings.Contains(risks[0].Summary, "10.0.0.2") {
		t.Fatal("business risk exposed a raw asset identifier")
	}
}

func TestReadinessUsesLiveScores(t *testing.T) {
	frameworks := readiness(scoring.Scores{Vulnerability: 60, Identity: 80, Endpoint: 70, Cloud: 90, Compliance: 100})
	if len(frameworks) != 3 {
		t.Fatalf("len(frameworks) = %d, want 3", len(frameworks))
	}
	if frameworks[0].Score != 76 {
		t.Fatalf("SOC 2 score = %d, want 76", frameworks[0].Score)
	}
}

func TestRenderPDF(t *testing.T) {
	data, err := RenderPDF(Summary{
		PeriodLabel:      "last 30 days",
		Posture:          Posture{Score: 82, Grade: "B", Delta: 4},
		Changes:          Changes{Highlights: []string{"Security posture improved."}},
		TopRisks:         []Risk{{Rank: 1, Title: "Business system needs attention", Summary: "Risk score 88/100."}},
		Frameworks:       []Framework{{Name: "SOC 2", Score: 76, State: "Material progress"}},
		ExecutiveMessage: "Security posture is improving.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 4 || string(data[:4]) != "%PDF" {
		t.Fatal("rendered report is not a PDF")
	}
}
