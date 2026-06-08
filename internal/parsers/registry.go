package parsers

import (
	"strings"

	"github.com/seccomply/seccomply/internal/models"
)

// Parser is implemented by every vendor-specific file parser.
type Parser interface {
	// Match returns true if this parser should handle the given file.
	// filename is the base name; headers are the first row of the sheet/CSV.
	Match(filename string, headers []string) bool
	// Parse converts raw rows (excluding the header row) into findings.
	Parse(headers []string, rows [][]string) ([]models.Finding, error)
	// VendorName is the canonical display name for this parser's source.
	VendorName() string
	// SourceTool is the normalized tool category.
	SourceTool() models.SourceTool
}

var registry []Parser

func init() {
	// Registration order matters: more specific parsers must come before fallback.
	Register(&CrowdStrikeParser{})
	Register(&FortinetParser{})
	Register(&TenableParser{})
	Register(&FallbackParser{})
}

// Register adds a parser to the global registry.
// Call from each parser's init() or directly — order determines priority.
func Register(p Parser) {
	registry = append(registry, p)
}

// Resolve returns the first parser that matches the given filename and headers.
// Returns nil if no parser matches (the fallback parser always matches, so this
// should only be nil if the registry was modified externally).
func Resolve(filename string, headers []string) Parser {
	for _, p := range registry {
		if p.Match(filename, headers) {
			return p
		}
	}
	return nil
}

// normalizeHeader returns a lowercase, trimmed header value for loose matching.
func normalizeHeader(h string) string {
	return strings.ToLower(strings.TrimSpace(h))
}

// headerSet builds a set from a header slice for O(1) presence checks.
func headerSet(headers []string) map[string]int {
	m := make(map[string]int, len(headers))
	for i, h := range headers {
		m[normalizeHeader(h)] = i
	}
	return m
}

// colVal safely returns a cell value by header name from a row.
func colVal(row []string, idx map[string]int, key string) string {
	i, ok := idx[normalizeHeader(key)]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}
