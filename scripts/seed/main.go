package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seccomply/seccomply/internal/models"
)

func main() {
	dbURL := envOr("DATABASE_URL", "postgres://seccomply:seccomply@localhost:5433/seccomply?sslmode=disable")
	ctx := context.Background()

	cfg, _ := pgxpool.ParseConfig(dbURL)
	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		slog.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		slog.Error("db ping", "err", err)
		os.Exit(1)
	}

	rng := rand.New(rand.NewSource(42))

	sfID := insertSourceFile(ctx, db, "seed_crowdstrike.csv", "CrowdStrike", 80)
	insertFindings(ctx, db, sfID, makeCrowdStrikeFindings(rng, 80), rng)

	sfID = insertSourceFile(ctx, db, "seed_fortinet.xlsx", "Fortinet", 70)
	insertFindings(ctx, db, sfID, makeFortinetFindings(rng, 70), rng)

	sfID = insertSourceFile(ctx, db, "seed_tenable.csv", "Tenable", 50)
	insertFindings(ctx, db, sfID, makeTenableFindings(rng, 50), rng)

	fmt.Println("seed complete: 200 findings inserted")
}

func insertSourceFile(ctx context.Context, db *pgxpool.Pool, filename, vendor string, rows int) uuid.UUID {
	var id uuid.UUID
	db.QueryRow(ctx,
		`INSERT INTO source_files (filename, sha256, row_count, parse_status, vendor_matched)
		 VALUES ($1, $2, $3, 'success', $4)
		 ON CONFLICT (sha256) DO UPDATE SET row_count = EXCLUDED.row_count
		 RETURNING id`,
		filename,
		fmt.Sprintf("seed-%s-%d", filename, rows),
		rows,
		vendor,
	).Scan(&id)
	return id
}

func insertFindings(ctx context.Context, db *pgxpool.Pool, sfID uuid.UUID, findings []models.Finding, rng *rand.Rand) {
	for _, f := range findings {
		payload, _ := json.Marshal(f.RawPayload)
		db.Exec(ctx,
			`INSERT INTO findings
			 (id, source_tool, source_vendor, external_id, severity, title,
			  description, affected_asset, first_seen, last_seen, status,
			  raw_payload, ingested_at, source_file_id)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			 ON CONFLICT DO NOTHING`,
			f.ID, string(f.SourceTool), f.SourceVendor, f.ExternalID,
			string(f.Severity), f.Title, f.Description, f.AffectedAsset,
			f.FirstSeen, f.LastSeen, string(f.Status),
			payload, f.IngestedAt, sfID,
		)
	}
}

var severities = []models.Severity{
	models.SeverityCritical, models.SeverityCritical,
	models.SeverityHigh, models.SeverityHigh, models.SeverityHigh,
	models.SeverityMedium, models.SeverityMedium, models.SeverityMedium, models.SeverityMedium,
	models.SeverityLow, models.SeverityLow,
	models.SeverityInfo,
}

var statuses = []models.Status{
	models.StatusOpen, models.StatusOpen, models.StatusOpen,
	models.StatusOpen, models.StatusOpen,
	models.StatusResolved,
	models.StatusSuppressed,
}

var hosts = []string{
	"WKSTN-101", "WKSTN-204", "SRV-DB01", "SRV-WEB02", "SRV-AD01",
	"LAPTOP-CEO", "LAPTOP-CFO", "SRV-MAIL", "SRV-VPN", "WKSTN-305",
	"SRV-FILE01", "WKSTN-007", "SRV-K8S-01", "SRV-K8S-02", "WKSTN-DEVOPS",
}

func randTime(rng *rand.Rand, daysBack int) time.Time {
	return time.Now().Add(-time.Duration(rng.Intn(daysBack*24)) * time.Hour)
}

func makeCrowdStrikeFindings(rng *rand.Rand, n int) []models.Finding {
	tactics := []string{
		"Credential Access", "Lateral Movement", "Persistence",
		"Defense Evasion", "Collection", "Exfiltration", "Execution",
		"Privilege Escalation", "Discovery", "Command and Control",
	}
	techniques := []string{
		"T1003 - OS Credential Dumping",
		"T1059 - Command and Scripting Interpreter",
		"T1078 - Valid Accounts",
		"T1055 - Process Injection",
		"T1547 - Boot Autostart Execution",
		"T1021 - Remote Services",
		"T1486 - Data Encrypted for Impact",
		"T1190 - Exploit Public-Facing Application",
		"T1110 - Brute Force",
		"T1071 - Application Layer Protocol",
	}

	findings := make([]models.Finding, n)
	for i := range findings {
		tactic := tactics[rng.Intn(len(tactics))]
		technique := techniques[rng.Intn(len(techniques))]
		sev := severities[rng.Intn(len(severities))]
		ts := randTime(rng, 30)

		findings[i] = models.Finding{
			ID:            uuid.New(),
			SourceTool:    models.SourceToolEDR,
			SourceVendor:  "CrowdStrike",
			ExternalID:    fmt.Sprintf("ldt:%s:1", uuid.New().String()[:8]),
			Severity:      sev,
			Title:         fmt.Sprintf("%s — %s", tactic, technique),
			Description:   fmt.Sprintf("CrowdStrike Falcon detected %s activity via %s on the endpoint.", tactic, technique),
			AffectedAsset: hosts[rng.Intn(len(hosts))],
			FirstSeen:     ts,
			LastSeen:      ts.Add(time.Duration(rng.Intn(60)) * time.Minute),
			Status:        statuses[rng.Intn(len(statuses))],
			RawPayload: map[string]any{
				"Tactic":     tactic,
				"Technique":  technique,
				"Severity":   string(sev),
				"Sensor":     "falcon-sensor-6.52",
			},
			IngestedAt: time.Now().Add(-time.Duration(rng.Intn(72)) * time.Hour),
		}
	}
	return findings
}

func makeFortinetFindings(rng *rand.Rand, n int) []models.Finding {
	actions := []string{"block", "accept", "reset", "close"}
	subtypes := []string{
		"intrusion", "virus", "webfilter", "app-ctrl",
		"anomaly", "dlp", "ips", "voip",
	}
	srcIPs := []string{
		"203.0.113.42", "198.51.100.7", "192.0.2.99",
		"10.10.10.50", "172.16.5.23", "185.220.101.45",
		"45.33.32.156", "104.21.53.201",
	}
	dstIPs := []string{
		"10.0.1.15", "10.0.2.30", "192.168.1.100",
		"10.0.5.200", "172.16.0.1",
	}

	findings := make([]models.Finding, n)
	for i := range findings {
		srcIP := srcIPs[rng.Intn(len(srcIPs))]
		dstIP := dstIPs[rng.Intn(len(dstIPs))]
		action := actions[rng.Intn(len(actions))]
		subtype := subtypes[rng.Intn(len(subtypes))]
		sev := severities[rng.Intn(len(severities))]
		ts := randTime(rng, 30)
		logID := fmt.Sprintf("%013d", rng.Int63())

		msg := fmt.Sprintf("%s traffic detected from %s to %s", subtype, srcIP, dstIP)

		findings[i] = models.Finding{
			ID:            uuid.New(),
			SourceTool:    models.SourceToolFirewall,
			SourceVendor:  "Fortinet",
			ExternalID:    logID,
			Severity:      sev,
			Title:         fmt.Sprintf("%s → %s [%s]", srcIP, dstIP, action),
			Description:   msg,
			AffectedAsset: hosts[rng.Intn(len(hosts))],
			FirstSeen:     ts,
			LastSeen:      ts,
			Status:        statuses[rng.Intn(len(statuses))],
			RawPayload: map[string]any{
				"logid":   logID,
				"subtype": subtype,
				"srcip":   srcIP,
				"dstip":   dstIP,
				"action":  action,
				"level":   string(sev),
			},
			IngestedAt: time.Now().Add(-time.Duration(rng.Intn(72)) * time.Hour),
		}
	}
	return findings
}

func makeTenableFindings(rng *rand.Rand, n int) []models.Finding {
	vulns := []struct {
		name     string
		cve      string
		synopsis string
	}{
		{"OpenSSH Authentication Bypass", "CVE-2023-38408", "Remote authentication bypass in OpenSSH"},
		{"Apache Log4j Remote Code Execution", "CVE-2021-44228", "JNDI injection allows unauthenticated RCE"},
		{"SSL/TLS Certificate Expired", "", "The SSL certificate on this service has expired"},
		{"SMBv1 Protocol Enabled", "CVE-2017-0144", "EternalBlue-vulnerable SMBv1 protocol is enabled"},
		{"SSH Weak Algorithms", "", "The SSH service accepts weak MAC/cipher algorithms"},
		{"Microsoft Exchange ProxyLogon", "CVE-2021-26855", "SSRF vulnerability in Exchange allows bypass"},
		{"Unpatched Windows Print Spooler", "CVE-2021-34527", "PrintNightmare allows privilege escalation"},
		{"Jenkins Remote Code Execution", "CVE-2019-1003000", "Sandbox bypass in Groovy script engine"},
		{"Default Credentials Detected", "", "Service is accessible with vendor default credentials"},
		{"Outdated OpenSSL", "CVE-2022-0778", "Infinite loop in BN_mod_sqrt() on invalid certs"},
	}

	findings := make([]models.Finding, n)
	for i := range findings {
		v := vulns[rng.Intn(len(vulns))]
		sev := severities[rng.Intn(len(severities))]
		ts := randTime(rng, 30)
		host := hosts[rng.Intn(len(hosts))]
		port := []string{"22", "80", "443", "445", "8080", "8443", "3389"}[rng.Intn(7)]
		pluginID := fmt.Sprintf("%d", 10000+rng.Intn(90000))

		extID := pluginID
		if v.cve != "" {
			extID = v.cve
		}

		findings[i] = models.Finding{
			ID:            uuid.New(),
			SourceTool:    models.SourceToolScanner,
			SourceVendor:  "Tenable",
			ExternalID:    extID,
			Severity:      sev,
			Title:         v.name,
			Description:   v.synopsis,
			AffectedAsset: fmt.Sprintf("%s:%s", host, port),
			FirstSeen:     ts,
			LastSeen:      ts,
			Status:        statuses[rng.Intn(len(statuses))],
			RawPayload: map[string]any{
				"Plugin ID": pluginID,
				"CVE":       v.cve,
				"Risk":      string(sev),
				"Host":      host,
				"Port":      port,
				"Synopsis":  v.synopsis,
			},
			IngestedAt: time.Now().Add(-time.Duration(rng.Intn(72)) * time.Hour),
		}
	}
	return findings
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
