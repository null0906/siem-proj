package alerting

import (
	"testing"

	"github.com/google/uuid"
)

func TestCompare(t *testing.T) {
	tests := []struct {
		value, threshold float64
		operator         string
		want             bool
	}{
		{11, 10, "gt", true},
		{10, 10, "gt", false},
		{9, 10, "lt", true},
		{10, 10, "gte", true},
		{10, 10, "lte", true},
		{10, 10, "eq", true},
		{10, 10, "unknown", false},
	}
	for _, test := range tests {
		if got := Compare(test.value, test.operator, test.threshold); got != test.want {
			t.Fatalf("Compare(%v, %q, %v) = %v, want %v", test.value, test.operator, test.threshold, got, test.want)
		}
	}
}

func TestFingerprintIsStableAndScoped(t *testing.T) {
	ruleID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	first := Fingerprint(ruleID, "global")
	if first != Fingerprint(ruleID, "global") {
		t.Fatal("fingerprint is not stable")
	}
	if first == Fingerprint(ruleID, "asset-1") {
		t.Fatal("fingerprint does not distinguish scopes")
	}
	if first == Fingerprint(uuid.New(), "global") {
		t.Fatal("fingerprint does not distinguish rules")
	}
}
