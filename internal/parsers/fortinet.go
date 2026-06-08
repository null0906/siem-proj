package parsers

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/seccomply/seccomply/internal/models"
)

// FortinetParser handles Fortinet FortiGate firewall log XLSX exports.
// Typical headers: date, time, devname, devid, logid, type, subtype,
//                  level, srcip, dstip, srcport, dstport, action, msg
type FortinetParser struct{}

func (p *FortinetParser) VendorName() string      { return "Fortinet" }
func (p *FortinetParser) SourceTool() models.SourceTool { return models.SourceToolFirewall }

func (p *FortinetParser) Match(filename string, headers []string) bool {
	lower := strings.ToLower(filename)
	if strings.Contains(lower, "fortinet") || strings.Contains(lower, "fortigate") ||
		strings.Contains(lower, "forti") {
		return true
	}
	idx := headerSet(headers)
	_, hasDevname := idx["devname"]
	_, hasLogid := idx["logid"]
	_, hasSrcIP := idx["srcip"]
	return hasDevname && hasLogid && hasSrcIP
}

func (p *FortinetParser) Parse(headers []string, rows [][]string) ([]models.Finding, error) {
	idx := headerSet(headers)
	findings := make([]models.Finding, 0, len(rows))

	for _, row := range rows {
		if len(row) == 0 || allEmpty(row) {
			continue
		}

		dateStr := colVal(row, idx, "date")
		timeStr := colVal(row, idx, "time")
		ts := parseTimestamp(dateStr + " " + timeStr)

		srcIP := colVal(row, idx, "srcip")
		dstIP := colVal(row, idx, "dstip")
		msg := colVal(row, idx, "msg")
		action := colVal(row, idx, "action")

		title := msg
		if title == "" {
			title = fmt.Sprintf("%s → %s [%s]",
				srcIP, dstIP, strings.ToUpper(action))
		}

		raw := rowToMap(headers, row)

		findings = append(findings, models.Finding{
			ID:            uuid.New(),
			SourceTool:    p.SourceTool(),
			SourceVendor:  p.VendorName(),
			ExternalID:    colVal(row, idx, "logid"),
			Severity:      mapFortSeverity(colVal(row, idx, "level")),
			Title:         title,
			Description:   msg,
			AffectedAsset: colVal(row, idx, "devname"),
			FirstSeen:     ts,
			LastSeen:      ts,
			Status:        mapFortAction(action),
			RawPayload:    raw,
			IngestedAt:    time.Now(),
		})
	}
	return findings, nil
}

func mapFortSeverity(s string) models.Severity {
	switch strings.ToLower(s) {
	case "critical", "emergency", "alert":
		return models.SeverityCritical
	case "high", "error":
		return models.SeverityHigh
	case "medium", "warning":
		return models.SeverityMedium
	case "low", "notice":
		return models.SeverityLow
	default:
		return models.SeverityInfo
	}
}

func mapFortAction(s string) models.Status {
	switch strings.ToLower(s) {
	case "accept", "allow":
		return models.StatusOpen
	case "close", "reset", "block":
		return models.StatusResolved
	default:
		return models.StatusOpen
	}
}
