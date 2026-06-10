package ingestor

import (
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seccomply/seccomply/internal/activity"
	"github.com/seccomply/seccomply/internal/alerting"
	"github.com/seccomply/seccomply/internal/assets"
	"github.com/seccomply/seccomply/internal/models"
	"github.com/seccomply/seccomply/internal/parsers"
	"github.com/xuri/excelize/v2"
)

const (
	pollInterval = 30 * time.Second
)

type Watcher struct {
	db           *pgxpool.Pool
	watchDir     string
	processedDir string
	log          *slog.Logger
}

func New(db *pgxpool.Pool, watchDir string) *Watcher {
	processedDir := filepath.Join(filepath.Dir(watchDir), "processed")
	return &Watcher{
		db:           db,
		watchDir:     watchDir,
		processedDir: processedDir,
		log:          slog.Default(),
	}
}

func (w *Watcher) Run(ctx context.Context) {
	w.log.Info("ingestor started", "watch_dir", w.watchDir, "interval", pollInterval)
	if err := os.MkdirAll(w.watchDir, 0755); err != nil {
		w.log.Error("cannot create watch dir", "err", err)
	}
	w.poll(ctx)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.poll(ctx)
		}
	}
}

func (w *Watcher) poll(ctx context.Context) {
	entries, err := os.ReadDir(w.watchDir)
	if err != nil {
		w.log.Warn("poll readdir failed", "err", err)
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".csv" && ext != ".xlsx" && ext != ".xls" {
			continue
		}
		fullPath := filepath.Join(w.watchDir, name)
		if err := w.processFile(ctx, fullPath, name); err != nil {
			w.log.Error("process file failed", "file", name, "err", err)
		}
	}
}

func (w *Watcher) processFile(ctx context.Context, fullPath, filename string) error {
	before, err := activity.New(w.db).CaptureState(ctx)
	if err != nil {
		w.log.Warn("capture pre-ingest activity state failed", "err", err)
		before = activity.State{HighRiskAssets: map[string]bool{}}
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	hash := sha256.Sum256(data)
	sha := hex.EncodeToString(hash[:])

	// Skip already-processed files (idempotent via sha256 unique constraint).
	var exists bool
	err = w.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM source_files WHERE sha256 = $1)`, sha).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check sha: %w", err)
	}
	if exists {
		w.log.Debug("file already processed, skipping", "file", filename)
		w.moveFile(fullPath, filename)
		w.refreshAssets(ctx)
		w.evaluateAlerts(ctx)
		return nil
	}

	w.emitPipelineEvent(ctx, "landed", filename)

	var sfID uuid.UUID
	err = w.db.QueryRow(ctx,
		`INSERT INTO source_files (filename, sha256, parse_status)
		 VALUES ($1, $2, 'pending') RETURNING id`,
		filename, sha,
	).Scan(&sfID)
	if err != nil {
		return fmt.Errorf("insert source_file: %w", err)
	}

	headers, rows, parseErr := readFile(fullPath, data)
	if parseErr != nil {
		w.updateSourceFile(ctx, sfID, models.ParseStatusFailed, 0, "", parseErr.Error())
		return nil
	}

	w.emitPipelineEvent(ctx, "parsed", filename)

	if identityParser := parsers.ResolveIdentity(filename, headers); identityParser != nil {
		users, parseErr := identityParser.ParseIdentity(headers, rows)
		if parseErr != nil {
			w.updateSourceFile(ctx, sfID, models.ParseStatusFailed, len(rows), identityParser.VendorName(), parseErr.Error())
			return nil
		}

		w.emitPipelineEvent(ctx, "normalized", filename)
		if err := w.insertIdentityUsers(ctx, sfID, users); err != nil {
			w.updateSourceFile(ctx, sfID, models.ParseStatusPartial, len(rows), identityParser.VendorName(), err.Error())
			return nil
		}

		w.updateSourceFile(ctx, sfID, models.ParseStatusSuccess, len(rows), identityParser.VendorName(), "")
		w.emitPipelineEvent(ctx, "indexed", filename)
		w.log.Info("identity file ingested",
			"file", filename,
			"vendor", identityParser.VendorName(),
			"users", len(users),
		)
		w.recordActivity(ctx, before, activity.IngestResult{
			SourceFileID: sfID, Filename: filename, Vendor: identityParser.VendorName(), Rows: len(rows),
		})
		w.moveFile(fullPath, filename)
		return nil
	}

	parser := parsers.Resolve(filename, headers)
	if parser == nil {
		w.updateSourceFile(ctx, sfID, models.ParseStatusFailed, 0, "", "no parser matched")
		return nil
	}

	findings, parseErr := parser.Parse(headers, rows)
	if parseErr != nil {
		w.updateSourceFile(ctx, sfID, models.ParseStatusFailed, len(rows), parser.VendorName(), parseErr.Error())
		return nil
	}

	w.emitPipelineEvent(ctx, "normalized", filename)

	if err := w.insertFindings(ctx, sfID, findings); err != nil {
		w.updateSourceFile(ctx, sfID, models.ParseStatusPartial, len(rows), parser.VendorName(), err.Error())
		return nil
	}

	w.updateSourceFile(ctx, sfID, models.ParseStatusSuccess, len(rows), parser.VendorName(), "")
	w.emitPipelineEvent(ctx, "indexed", filename)

	w.log.Info("file ingested",
		"file", filename,
		"vendor", parser.VendorName(),
		"findings", len(findings),
	)

	w.moveFile(fullPath, filename)
	w.refreshAssets(ctx)
	newCriticals, resolved := findingCounts(findings)
	w.recordActivity(ctx, before, activity.IngestResult{
		SourceFileID: sfID, Filename: filename, Vendor: parser.VendorName(), Rows: len(rows),
		NewFindings: len(findings), NewCriticals: newCriticals, Resolved: resolved,
	})
	w.evaluateAlerts(ctx)
	return nil
}

func findingCounts(findings []models.Finding) (criticals, resolved int) {
	for _, finding := range findings {
		if finding.Severity == models.SeverityCritical {
			criticals++
		}
		if finding.Status == models.StatusResolved {
			resolved++
		}
	}
	return criticals, resolved
}

func (w *Watcher) recordActivity(ctx context.Context, before activity.State, result activity.IngestResult) {
	if err := activity.New(w.db).RecordIngest(ctx, before, result); err != nil {
		w.log.Warn("post-ingest activity recording failed", "err", err)
	}
}

func (w *Watcher) refreshAssets(ctx context.Context) {
	if err := assets.New(w.db).Refresh(ctx); err != nil {
		w.log.Warn("post-ingest asset refresh failed", "err", err)
		return
	}
	w.log.Info("post-ingest asset refresh complete")
}

func (w *Watcher) evaluateAlerts(ctx context.Context) {
	result, err := alerting.New(w.db).Evaluate(ctx, alerting.TriggerOnIngest)
	if err != nil {
		w.log.Warn("post-ingest alert evaluation failed", "err", err)
		return
	}
	w.log.Info("post-ingest alert evaluation complete",
		"fired", result.Fired,
		"resolved", result.AutoResolved,
		"suppressed", result.Suppressed,
	)
}

func readFile(path string, data []byte) (headers []string, rows [][]string, err error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".csv":
		return readCSV(data)
	case ".xlsx", ".xls":
		return readXLSX(path)
	default:
		return nil, nil, fmt.Errorf("unsupported extension: %s", ext)
	}
}

func readCSV(data []byte) ([]string, [][]string, error) {
	r := csv.NewReader(strings.NewReader(string(data)))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	all, err := r.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(all) < 1 {
		return nil, nil, fmt.Errorf("empty file")
	}
	return all[0], all[1:], nil
}

func readXLSX(path string) ([]string, [][]string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, fmt.Errorf("no sheets in workbook")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, nil, fmt.Errorf("get rows: %w", err)
	}
	if len(rows) < 1 {
		return nil, nil, fmt.Errorf("empty sheet")
	}
	return rows[0], rows[1:], nil
}

func (w *Watcher) insertFindings(ctx context.Context, sfID uuid.UUID, findings []models.Finding) error {
	batch := &pgx.Batch{}
	for _, f := range findings {
		payload, _ := json.Marshal(f.RawPayload)
		batch.Queue(
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
	br := w.db.SendBatch(ctx, batch)
	defer br.Close()
	for range findings {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert: %w", err)
		}
	}
	return nil
}

func (w *Watcher) insertIdentityUsers(ctx context.Context, sfID uuid.UUID, users []models.IdentityUser) error {
	batch := &pgx.Batch{}
	for _, user := range users {
		batch.Queue(
			`INSERT INTO identity_users
			 (id, user_email, display_name, mfa_enabled, account_status, last_login,
			  is_privileged, groups, sso_apps_count, created_at, source_file_id)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			 ON CONFLICT (user_email, source_file_id) DO UPDATE SET
			  display_name = EXCLUDED.display_name,
			  mfa_enabled = EXCLUDED.mfa_enabled,
			  account_status = EXCLUDED.account_status,
			  last_login = EXCLUDED.last_login,
			  is_privileged = EXCLUDED.is_privileged,
			  groups = EXCLUDED.groups,
			  sso_apps_count = EXCLUDED.sso_apps_count`,
			user.ID, user.UserEmail, user.DisplayName, user.MFAEnabled, user.AccountStatus,
			user.LastLogin, user.IsPrivileged, user.Groups, user.SSOAppsCount, user.CreatedAt, sfID,
		)
	}
	br := w.db.SendBatch(ctx, batch)
	defer br.Close()
	for range users {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("identity batch insert: %w", err)
		}
	}
	return nil
}

func (w *Watcher) updateSourceFile(ctx context.Context, id uuid.UUID, status models.ParseStatus, rows int, vendor, errLog string) {
	_, err := w.db.Exec(ctx,
		`UPDATE source_files SET parse_status=$1, row_count=$2, vendor_matched=$3, error_log=$4 WHERE id=$5`,
		string(status), rows, vendor, errLog, id,
	)
	if err != nil {
		w.log.Warn("update source_file failed", "id", id, "err", err)
	}
}

func (w *Watcher) emitPipelineEvent(ctx context.Context, eventType, filename string) {
	_, _ = w.db.Exec(ctx,
		`INSERT INTO pipeline_events (event_type, filename) VALUES ($1, $2)`,
		eventType, filename,
	)
}

func (w *Watcher) moveFile(src, filename string) {
	dateDir := filepath.Join(w.processedDir, time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		w.log.Warn("cannot create processed dir", "err", err)
		return
	}
	dst := filepath.Join(dateDir, filename)
	if err := os.Rename(src, dst); err != nil {
		// Copy+delete as fallback (cross-device rename).
		if copyErr := copyFile(src, dst); copyErr == nil {
			os.Remove(src)
		} else {
			w.log.Warn("move file failed", "src", src, "dst", dst, "err", err)
		}
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
