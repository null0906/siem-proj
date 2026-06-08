package parsers

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/seccomply/seccomply/internal/models"
)

// TenableParser handles Tenable Nessus CSV exports.
// Typical headers: Plugin ID, CVE, CVSS v2.0 Base Score, Risk,
//                  Host, Protocol, Port, Name, Synopsis, Description,
//                  Solution, See Also, Plugin Output
type TenableParser struct{}

func (p *TenableParser) VendorName() string      { return "Tenable" }
func (p *TenableParser) SourceTool() models.SourceTool { return models.SourceToolScanner }

func (p *TenableParser) Match(filename string, headers []string) bool {
	lower := strings.ToLower(filename)
	if strings.Contains(lower, "tenable") || strings.Contains(lower, "nessus") {
		return true
	}
	idx := headerSet(headers)
	_, hasPlugin := idx["plugin id"]
	_, hasRisk := idx["risk"]
	_, hasCVE := idx["cve"]
	return hasPlugin && (hasRisk || hasCVE)
}

func (p *TenableParser) Parse(headers []string, rows [][]string) ([]models.Finding, error) {
	idx := headerSet(headers)
	findings := make([]models.Finding, 0, len(rows))

	for _, row := range rows {
		if len(row) == 0 || allEmpty(row) {
			continue
		}

		pluginID := colVal(row, idx, "plugin id")
		cve := colVal(row, idx, "cve")
		name := colVal(row, idx, "name")
		host := colVal(row, idx, "host")
		port := colVal(row, idx, "port")
		protocol := colVal(row, idx, "protocol")

		extID := pluginID
		if cve != "" {
			extID = cve
		}

		title := name
		if title == "" {
			title = fmt.Sprintf("Plugin %s on %s:%s/%s", pluginID, host, port, protocol)
		}

		desc := colVal(row, idx, "synopsis")
		if desc == "" {
			desc = colVal(row, idx, "description")
		}

		raw := rowToMap(headers, row)

		findings = append(findings, models.Finding{
			ID:            uuid.New(),
			SourceTool:    p.SourceTool(),
			SourceVendor:  p.VendorName(),
			ExternalID:    extID,
			Severity:      mapTenableSeverity(colVal(row, idx, "risk")),
			Title:         title,
			Description:   desc,
			AffectedAsset: fmt.Sprintf("%s:%s", host, port),
			FirstSeen:     time.Now(),
			LastSeen:      time.Now(),
			Status:        models.StatusOpen,
			RawPayload:    raw,
			IngestedAt:    time.Now(),
		})
	}
	return findings, nil
}

func mapTenableSeverity(s string) models.Severity {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "critical":
		return models.SeverityCritical
	case "high":
		return models.SeverityHigh
	case "medium":
		return models.SeverityMedium
	case "low":
		return models.SeverityLow
	case "none", "info", "informational":
		return models.SeverityInfo
	default:
		return models.SeverityInfo
	}
}
