package parsers

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/seccomply/seccomply/internal/models"
)

// CrowdStrikeParser handles CrowdStrike Falcon CSV exports.
// Typical headers: Detection ID, Status, Severity, Tactic, Technique,
//                  Hostname, Username, Filename, Timestamp, Description
type CrowdStrikeParser struct{}

func (p *CrowdStrikeParser) VendorName() string      { return "CrowdStrike" }
func (p *CrowdStrikeParser) SourceTool() models.SourceTool { return models.SourceToolEDR }

func (p *CrowdStrikeParser) Match(filename string, headers []string) bool {
	lower := strings.ToLower(filename)
	if strings.Contains(lower, "crowdstrike") || strings.Contains(lower, "falcon") {
		return true
	}
	idx := headerSet(headers)
	_, hasDetID := idx["detection id"]
	_, hasTactic := idx["tactic"]
	_, hasTechnique := idx["technique"]
	return hasDetID && (hasTactic || hasTechnique)
}

func (p *CrowdStrikeParser) Parse(headers []string, rows [][]string) ([]models.Finding, error) {
	idx := headerSet(headers)
	findings := make([]models.Finding, 0, len(rows))

	for _, row := range rows {
		if len(row) == 0 || allEmpty(row) {
			continue
		}

		severity := mapCSSeverity(colVal(row, idx, "severity"))
		ts := parseTimestamp(colVal(row, idx, "timestamp"))

		title := colVal(row, idx, "description")
		if title == "" {
			tactic := colVal(row, idx, "tactic")
			technique := colVal(row, idx, "technique")
			title = fmt.Sprintf("%s — %s", tactic, technique)
		}

		raw := rowToMap(headers, row)

		findings = append(findings, models.Finding{
			ID:            uuid.New(),
			SourceTool:    p.SourceTool(),
			SourceVendor:  p.VendorName(),
			ExternalID:    colVal(row, idx, "detection id"),
			Severity:      severity,
			Title:         title,
			Description:   colVal(row, idx, "description"),
			AffectedAsset: colVal(row, idx, "hostname"),
			FirstSeen:     ts,
			LastSeen:      ts,
			Status:        mapCSStatus(colVal(row, idx, "status")),
			RawPayload:    raw,
			IngestedAt:    time.Now(),
		})
	}
	return findings, nil
}

func mapCSSeverity(s string) models.Severity {
	switch strings.ToLower(s) {
	case "critical":
		return models.SeverityCritical
	case "high":
		return models.SeverityHigh
	case "medium", "moderate":
		return models.SeverityMedium
	case "low":
		return models.SeverityLow
	default:
		return models.SeverityInfo
	}
}

func mapCSStatus(s string) models.Status {
	switch strings.ToLower(s) {
	case "closed", "resolved":
		return models.StatusResolved
	case "suppressed", "ignored":
		return models.StatusSuppressed
	default:
		return models.StatusOpen
	}
}
