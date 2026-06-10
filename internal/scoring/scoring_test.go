package scoring

import "testing"

func TestCalculate(t *testing.T) {
	scores := Calculate(Metrics{
		OpenCritical:              2,
		OpenHigh:                  4,
		SLABreached:               1,
		ExploitableUnpatched:      1,
		TotalAccounts:             100,
		AccountsWithoutMFA:        10,
		PrivilegedWithoutMFA:      2,
		DormantPrivileged:         1,
		TotalDevices:              100,
		UnencryptedDevices:        10,
		DevicesWithoutEDR:         20,
		NonCompliantDevices:       2,
		PublicExposures:           1,
		CriticalCloudFindings:     2,
		IAMOverPermissions:        3,
		ComplianceFrameworkScores: []float64{80, 90, 100},
	})

	if scores.Vulnerability != 87 {
		t.Fatalf("vulnerability = %d, want 87", scores.Vulnerability)
	}
	if scores.Identity != 82 {
		t.Fatalf("identity = %d, want 82", scores.Identity)
	}
	if scores.Endpoint != 77 {
		t.Fatalf("endpoint = %d, want 77", scores.Endpoint)
	}
	if scores.Cloud != 90 {
		t.Fatalf("cloud = %d, want 90", scores.Cloud)
	}
	if scores.Compliance != 90 {
		t.Fatalf("compliance = %d, want 90", scores.Compliance)
	}
	if scores.Overall != 86 {
		t.Fatalf("overall = %d, want 86", scores.Overall)
	}
}

func TestCalculateFloorsAndGrade(t *testing.T) {
	scores := Calculate(Metrics{OpenCritical: 100})
	if scores.Vulnerability != 0 {
		t.Fatalf("vulnerability = %d, want floor 0", scores.Vulnerability)
	}

	tests := map[int]string{95: "A", 85: "B", 75: "C", 65: "D", 55: "F"}
	for score, want := range tests {
		if got := Grade(score); got != want {
			t.Fatalf("Grade(%d) = %q, want %q", score, got, want)
		}
	}
}

func TestFrameworkScoresAndFailingControlCount(t *testing.T) {
	metrics := Metrics{AccountsWithoutMFA: 10}
	mappings := []ControlMapping{
		{Framework: "SOC 2", ControlCode: "CC6.1", GapType: "accounts_without_mfa"},
		{Framework: "SOC 2", ControlCode: "CC7.2", GapType: "devices_without_edr"},
		{Framework: "ISO 27001", ControlCode: "A.5.15", GapType: "accounts_without_mfa"},
	}
	scores := FrameworkScores(metrics, mappings)
	if len(scores) != 2 {
		t.Fatalf("len(scores) = %d, want 2", len(scores))
	}
	if got := FailingControlCount(metrics, mappings); got != 2 {
		t.Fatalf("FailingControlCount() = %d, want 2", got)
	}
}
