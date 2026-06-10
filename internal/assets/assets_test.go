package assets

import "testing"

func TestNormalize(t *testing.T) {
	tests := map[string]string{
		"SRV-WEB02:443":             "srv-web02",
		"srv-web02.seccomply.local": "srv-web02",
		"  WKSTN-101  ":             "wkstn-101",
		"10.0.1.15:8443":            "10.0.1.15",
		"2001:db8::1":               "2001:db8::1",
		"":                          "",
	}
	for input, want := range tests {
		if got := Normalize(input); got != want {
			t.Fatalf("Normalize(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCalculateRiskCapsAtHundred(t *testing.T) {
	item := &aggregate{
		Asset:    Asset{SourceCount: 4, IsPublic: true, HasEDR: false, IsEncrypted: false, OwnerIsPrivileged: true},
		severity: map[string]int{"critical": 10, "high": 10},
	}
	if got := calculateRisk(item); got != 100 {
		t.Fatalf("calculateRisk() = %d, want 100", got)
	}
}
