package remediation

import (
	"testing"
	"time"

	"github.com/seccomply/seccomply/internal/models"
)

func TestApplySLA(t *testing.T) {
	now := time.Now()
	finding := models.Finding{Severity: models.SeverityCritical, Status: models.StatusOpen, FirstSeen: now.Add(-8 * 24 * time.Hour)}
	ApplySLA(&finding, now)
	if finding.SLAStatus != "breached" {
		t.Fatalf("SLAStatus = %q, want breached", finding.SLAStatus)
	}
	due := now.Add(48 * time.Hour)
	finding.DueDate = &due
	ApplySLA(&finding, now)
	if finding.SLAStatus != "due_soon" {
		t.Fatalf("SLAStatus = %q, want due_soon", finding.SLAStatus)
	}
}
