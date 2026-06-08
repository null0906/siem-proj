# SecComply

Self-hosted security operations dashboard for mid-sized startups. Aggregates findings from CrowdStrike, Fortinet, Tenable (and any CSV/XLSX tool) into a single normalized view.

Ships as a Docker Compose stack. No cloud dependency. No source shipped.

---

## Quick Start

```bash
cp .env.example .env
# Edit .env — at minimum change POSTGRES_PASSWORD

docker compose up -d
```

Dashboard: http://localhost:3000  
API: http://localhost:8080/api/health

To populate with seed data:
```bash
docker compose run --rm seed
```

Or against a local Postgres:
```bash
DATABASE_URL="postgres://seccomply:seccomply@localhost:5432/seccomply?sslmode=disable" \
  go run ./scripts/seed
```

---

## Ingest a file

Drop any `.csv`, `.xlsx`, or `.xls` into the watched folder. The ingestor polls every 30 seconds.

```bash
# Map the inbox volume to a host path
# In docker-compose.yml, change:
#   - inbox_data:/data
# to:
#   - /your/host/path:/data

# Then drop files:
cp crowdstrike_export.csv /your/host/path/inbox/
```

Processed files are moved to `/data/processed/YYYY-MM-DD/` automatically.

---

## Adding a new parser

Create a single file in `internal/parsers/`:

```go
package parsers

import "github.com/seccomply/seccomply/internal/models"

type MyVendorParser struct{}

func (p *MyVendorParser) VendorName() string           { return "MyVendor" }
func (p *MyVendorParser) SourceTool() models.SourceTool { return models.SourceToolEDR }

func (p *MyVendorParser) Match(filename string, headers []string) bool {
    // Return true if this file belongs to your vendor.
    // Check filename patterns or header columns.
    return strings.Contains(strings.ToLower(filename), "myvendor")
}

func (p *MyVendorParser) Parse(headers []string, rows [][]string) ([]models.Finding, error) {
    idx := headerSet(headers)
    findings := make([]models.Finding, 0, len(rows))
    for _, row := range rows {
        findings = append(findings, models.Finding{
            ID:           uuid.New(),
            SourceTool:   p.SourceTool(),
            SourceVendor: p.VendorName(),
            Severity:     mapMySeverity(colVal(row, idx, "severity")),
            Title:        colVal(row, idx, "title"),
            // ... map remaining fields
        })
    }
    return findings, nil
}
```

Then register it in `registry.go` before the fallback parser:

```go
func init() {
    Register(&CrowdStrikeParser{})
    Register(&FortinetParser{})
    Register(&TenableParser{})
    Register(&MyVendorParser{}) // ← add before FallbackParser
    Register(&FallbackParser{})
}
```

That's it. Rebuild the ingestor binary and redeploy.

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DATABASE_URL` | — | Postgres connection string (required) |
| `POSTGRES_DB` | `seccomply` | Database name |
| `POSTGRES_USER` | `seccomply` | Database user |
| `POSTGRES_PASSWORD` | — | Database password (required, change this) |
| `PORT` | `8080` | API server port |
| `WATCH_DIR` | `/data/inbox` | Directory the ingestor polls |
| `HEARTBEAT_URL` | *(empty)* | Anonymous telemetry endpoint; leave blank to disable |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | Browser-visible API URL |
| `API_PORT` | `8080` | Host port for the API container |
| `WEB_PORT` | `3000` | Host port for the web container |

---

## Architecture

```
┌─────────────┐     polls /data/inbox      ┌──────────────┐
│  ingestor   │ ──────────────────────────▶│   Postgres   │
│  (Go bin)   │  inserts findings/events   │   (pg 16)    │
└─────────────┘                            └──────┬───────┘
                                                  │
┌─────────────┐     REST /api/*            ┌──────┴───────┐
│     web     │ ◀──────────────────────────│     api      │
│  (Next.js)  │                            │  (Go bin)    │
└─────────────┘                            └──────────────┘
```

### Parser precedence

The ingestor resolves parsers in registration order. The first `Match()` that returns `true` wins:

1. CrowdStrike — filename contains "crowdstrike"/"falcon" OR headers contain "Detection ID" + "Tactic"
2. Fortinet — filename contains "fortinet"/"fortigate" OR headers contain "devname" + "logid" + "srcip"
3. Tenable — filename contains "tenable"/"nessus" OR headers contain "Plugin ID" + "Risk"/"CVE"
4. Fallback — catches everything else via best-effort header heuristics

### Database schema

- `findings` — normalized finding records (UUID PK, severity enum, JSONB raw payload)
- `source_files` — one row per ingested file with parse status and error log
- `pipeline_events` — lightweight event log powering the ingestion pipeline strip
- `deployment` — single-row table storing the deployment UUID

---

## Building from source

```bash
# API
go build -o bin/api ./cmd/api

# Ingestor
go build -o bin/ingestor ./cmd/ingestor

# Frontend
cd web && npm ci && npm run build
```

---

## Production notes

- Set a strong `POSTGRES_PASSWORD` in `.env` and do not commit `.env` to version control.
- The `HEARTBEAT_URL` is optional and non-enforcing — it sends `deployment_id`, `version`, `install_age_days`, and `findings_count` once every 24 hours. Disable by leaving the variable unset.
- The `api` service runs database migrations on startup via `golang-migrate`. Migrations are embedded in the binary.
- Processed files are moved to `/data/processed/YYYY-MM-DD/` and are not deleted automatically. Back up or prune this directory as needed.
