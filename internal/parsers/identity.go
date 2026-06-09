package parsers

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/seccomply/seccomply/internal/models"
)

type IdentityParser interface {
	Match(filename string, headers []string) bool
	ParseIdentity(headers []string, rows [][]string) ([]models.IdentityUser, error)
	VendorName() string
}

type OktaIdentityParser struct{}
type EntraIdentityParser struct{}
type GoogleIdentityParser struct{}
type GenericIdentityParser struct{}

var identityRegistry = []IdentityParser{
	&OktaIdentityParser{},
	&EntraIdentityParser{},
	&GoogleIdentityParser{},
	&GenericIdentityParser{},
}

func ResolveIdentity(filename string, headers []string) IdentityParser {
	for _, parser := range identityRegistry {
		if parser.Match(filename, headers) {
			return parser
		}
	}
	return nil
}

func (p *OktaIdentityParser) VendorName() string    { return "Okta" }
func (p *EntraIdentityParser) VendorName() string   { return "Microsoft Entra ID" }
func (p *GoogleIdentityParser) VendorName() string  { return "Google Workspace" }
func (p *GenericIdentityParser) VendorName() string { return "Identity" }

func (p *OktaIdentityParser) Match(filename string, headers []string) bool {
	name := strings.ToLower(filename)
	return strings.Contains(name, "okta") ||
		hasIdentityHeaders(headers, []string{"login", "status", "multifactor"})
}

func (p *EntraIdentityParser) Match(filename string, headers []string) bool {
	name := strings.ToLower(filename)
	return strings.Contains(name, "entra") || strings.Contains(name, "azuread") ||
		hasIdentityHeaders(headers, []string{"user principal name", "account enabled", "strong authentication"})
}

func (p *GoogleIdentityParser) Match(filename string, headers []string) bool {
	name := strings.ToLower(filename)
	return strings.Contains(name, "google") || strings.Contains(name, "workspace") ||
		hasIdentityHeaders(headers, []string{"email address", "2-step verification enrolled", "last login time"})
}

func (p *GenericIdentityParser) Match(_ string, headers []string) bool {
	normalized := make(map[string]bool, len(headers))
	for _, header := range headers {
		normalized[normalizeHeader(header)] = true
	}
	hasEmail := normalized["email"] || normalized["email address"] || normalized["user email"] ||
		normalized["login"] || normalized["user principal name"]
	hasIdentitySignal := normalized["mfa"] || normalized["mfa enabled"] ||
		normalized["last login"] || normalized["last login time"] ||
		normalized["account status"] || normalized["2-step verification enrolled"]
	return hasEmail && hasIdentitySignal
}

func hasIdentityHeaders(headers []string, required []string) bool {
	set := headerSet(headers)
	for _, header := range required {
		if _, ok := set[normalizeHeader(header)]; !ok {
			return false
		}
	}
	return true
}

func (p *OktaIdentityParser) ParseIdentity(headers []string, rows [][]string) ([]models.IdentityUser, error) {
	return parseIdentityRows(headers, rows)
}

func (p *EntraIdentityParser) ParseIdentity(headers []string, rows [][]string) ([]models.IdentityUser, error) {
	return parseIdentityRows(headers, rows)
}

func (p *GoogleIdentityParser) ParseIdentity(headers []string, rows [][]string) ([]models.IdentityUser, error) {
	return parseIdentityRows(headers, rows)
}

func (p *GenericIdentityParser) ParseIdentity(headers []string, rows [][]string) ([]models.IdentityUser, error) {
	return parseIdentityRows(headers, rows)
}

func parseIdentityRows(headers []string, rows [][]string) ([]models.IdentityUser, error) {
	idx := headerSet(headers)
	users := make([]models.IdentityUser, 0, len(rows))

	for _, row := range rows {
		if len(row) == 0 || allEmpty(row) {
			continue
		}

		email := firstNonEmpty(row, idx,
			"user_email", "user email", "email", "email address", "primary email", "login", "user principal name", "upn")
		if email == "" {
			continue
		}

		lastLogin := parseTimestamp(firstNonEmpty(row, idx,
			"last_login", "last login", "last login time", "last sign-in", "last signin", "last activity"))
		status := normalizeIdentityStatus(firstNonEmpty(row, idx,
			"account_status", "account status", "status", "user status", "account enabled"), lastLogin)

		users = append(users, models.IdentityUser{
			ID:          uuid.New(),
			UserEmail:   strings.ToLower(email),
			DisplayName: firstNonEmpty(row, idx, "display_name", "display name", "name", "full name"),
			MFAEnabled: parseIdentityBool(firstNonEmpty(row, idx,
				"mfa_enabled", "mfa enabled", "mfa", "multifactor", "strong authentication", "2-step verification enrolled", "2sv enrolled")),
			AccountStatus: status,
			LastLogin:     lastLogin,
			IsPrivileged: parseIdentityBool(firstNonEmpty(row, idx,
				"is_privileged", "is privileged", "privileged", "admin", "is admin", "super admin")),
			Groups: firstNonEmpty(row, idx, "groups", "group membership", "member of", "roles"),
			SSOAppsCount: parseIdentityInt(firstNonEmpty(row, idx,
				"sso_apps_count", "sso apps count", "applications", "assigned apps", "app count")),
			CreatedAt: parseTimestamp(firstNonEmpty(row, idx,
				"created_at", "created at", "created", "creation time", "user creation time")),
		})
	}

	return users, nil
}

func parseIdentityBool(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "yes", "y", "1", "enabled", "enrolled", "active":
		return true
	default:
		return false
	}
}

func parseIdentityInt(value string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(value))
	return n
}

func normalizeIdentityStatus(value string, lastLogin time.Time) string {
	status := strings.ToLower(strings.TrimSpace(value))
	switch status {
	case "false", "disabled", "deactivated", "suspended", "locked":
		return "suspended"
	case "dormant", "inactive":
		return "dormant"
	}
	if lastLogin.Before(time.Now().AddDate(0, 0, -90)) {
		return "dormant"
	}
	return "active"
}
