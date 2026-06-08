package parsers

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/seccomply/seccomply/internal/models"
)

// FallbackParser handles any CSV/XLSX that doesn't match a specific vendor.
// It performs best-effort column detection using common header patterns.
type FallbackParser struct{}

func (p *FallbackParser) VendorName() string      { return "Unknown" }
func (p *FallbackParser) SourceTool() models.SourceTool { return models.SourceToolOther }

// Match always returns true — it's the catch-all of last resort.
func (p *FallbackParser) Match(_ string, _ []string) bool { return true }

func (p *FallbackParser) Parse(headers []string, rows [][]string) ([]models.Finding, error) {
	idx := headerSet(headers)
	findings := make([]models.Finding, 0, len(rows))

	for _, row := range rows {
		if len(row) == 0 || allEmpty(row) {
			continue
		}

		title := firstNonEmpty(row, idx,
			"title", "name", "finding", "alert", "event", "message", "msg", "description", "summary")
		if title == "" {
			title = "(untitled finding)"
		}

		asset := firstNonEmpty(row, idx,
			"asset", "host", "hostname", "ip", "ip address", "device", "source ip", "srcip", "target")

		severity := mapFallbackSeverity(firstNonEmpty(row, idx,
			"severity", "risk", "level", "priority", "criticality"))

		tsStr := firstNonEmpty(row, idx,
			"timestamp", "date", "time", "detected at", "created at", "event time", "log time")
		ts := parseTimestamp(tsStr)

		raw := rowToMap(headers, row)

		findings = append(findings, models.Finding{
			ID:            uuid.New(),
			SourceTool:    p.SourceTool(),
			SourceVendor:  p.VendorName(),
			ExternalID:    firstNonEmpty(row, idx, "id", "event id", "alert id", "finding id"),
			Severity:      severity,
			Title:         title,
			Description:   firstNonEmpty(row, idx, "description", "details", "summary", "message"),
			AffectedAsset: asset,
			FirstSeen:     ts,
			LastSeen:      ts,
			Status:        models.StatusOpen,
			RawPayload:    raw,
			IngestedAt:    time.Now(),
		})
	}
	return findings, nil
}

func mapFallbackSeverity(s string) models.Severity {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "critical", "crit", "p1", "1":
		return models.SeverityCritical
	case "high", "p2", "2":
		return models.SeverityHigh
	case "medium", "med", "moderate", "p3", "3":
		return models.SeverityMedium
	case "low", "p4", "4":
		return models.SeverityLow
	default:
		return models.SeverityInfo
	}
}

func firstNonEmpty(row []string, idx map[string]int, keys ...string) string {
	for _, k := range keys {
		if v := colVal(row, idx, k); v != "" {
			return v
		}
	}
	return ""
}
