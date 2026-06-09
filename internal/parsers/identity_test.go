package parsers

import "testing"

func TestResolveAndParseGenericIdentityExport(t *testing.T) {
	headers := []string{
		"Email", "Display Name", "MFA Enabled", "Account Status",
		"Last Login", "Is Privileged", "Groups", "SSO Apps Count",
	}
	rows := [][]string{
		{
			"admin@example.com", "Demo Admin", "yes", "active",
			"2026-06-08", "true", "Security, Administrators", "12",
		},
	}

	parser := ResolveIdentity("users-export.csv", headers)
	if parser == nil {
		t.Fatal("expected identity parser to match generic user export")
	}

	users, err := parser.ParseIdentity(headers, rows)
	if err != nil {
		t.Fatalf("parse identity rows: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 identity user, got %d", len(users))
	}

	user := users[0]
	if user.UserEmail != "admin@example.com" {
		t.Fatalf("unexpected email: %s", user.UserEmail)
	}
	if !user.MFAEnabled || !user.IsPrivileged {
		t.Fatal("expected MFA-enabled privileged user")
	}
	if user.SSOAppsCount != 12 {
		t.Fatalf("expected 12 SSO apps, got %d", user.SSOAppsCount)
	}
}

func TestResolveOktaIdentityExport(t *testing.T) {
	headers := []string{"login", "status", "multifactor", "last login"}
	parser := ResolveIdentity("okta-directory.csv", headers)
	if parser == nil || parser.VendorName() != "Okta" {
		t.Fatalf("expected Okta parser, got %#v", parser)
	}
}
