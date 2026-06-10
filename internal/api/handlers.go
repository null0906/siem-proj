package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seccomply/seccomply/internal/activity"
	"github.com/seccomply/seccomply/internal/alerting"
	"github.com/seccomply/seccomply/internal/assets"
	"github.com/seccomply/seccomply/internal/compliance"
	"github.com/seccomply/seccomply/internal/executive"
	"github.com/seccomply/seccomply/internal/models"
	"github.com/seccomply/seccomply/internal/prioritization"
	"github.com/seccomply/seccomply/internal/remediation"
	"github.com/seccomply/seccomply/internal/scoring"
)

const version = "0.1.0"

var startTime = time.Now()

type Handlers struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func NewHandlers(db *pgxpool.Pool) *Handlers {
	return &Handlers{db: db, log: slog.Default()}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"version": version,
		"uptime":  time.Since(startTime).String(),
	})
}

func (h *Handlers) GetDeployment(w http.ResponseWriter, r *http.Request) {
	var d models.Deployment
	err := h.db.QueryRow(r.Context(),
		`SELECT id, deployment_id, version, created_at FROM deployment LIMIT 1`,
	).Scan(&d.ID, &d.DeploymentID, &d.Version, &d.CreatedAt)
	if err == pgx.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no deployment record"})
		return
	}
	if err != nil {
		h.serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (h *Handlers) GetDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var summary models.DashboardSummary
	summary.BySourceTool = map[string]int64{}

	err := h.db.QueryRow(ctx,
		`SELECT
			COUNT(*) FILTER (WHERE status IN ('open','in_progress')),
			COUNT(*) FILTER (WHERE severity = 'critical'),
			COUNT(*) FILTER (WHERE severity = 'high'),
			COUNT(*) FILTER (WHERE severity = 'medium'),
			COUNT(*) FILTER (WHERE severity = 'low'),
			COUNT(*) FILTER (WHERE severity = 'info'),
			COUNT(*)
		 FROM findings`,
	).Scan(
		&summary.OpenFindings,
		&summary.CriticalCount,
		&summary.HighCount,
		&summary.MediumCount,
		&summary.LowCount,
		&summary.InfoCount,
		&summary.TotalFindings,
	)
	if err != nil {
		h.serverErr(w, err)
		return
	}

	rows, err := h.db.Query(ctx,
		`SELECT source_tool, COUNT(*) FROM findings GROUP BY source_tool`)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var tool string
		var cnt int64
		if err := rows.Scan(&tool, &cnt); err == nil {
			summary.BySourceTool[tool] = cnt
		}
	}

	err = h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM source_files WHERE ingested_at >= CURRENT_DATE`,
	).Scan(&summary.FilesProcessed)
	if err != nil {
		h.serverErr(w, err)
		return
	}

	var lastIngested time.Time
	err = h.db.QueryRow(ctx,
		`SELECT MAX(ingested_at) FROM findings`,
	).Scan(&lastIngested)
	if err == nil && !lastIngested.IsZero() {
		summary.LastIngestedAt = &lastIngested
	}

	writeJSON(w, http.StatusOK, summary)
}

func (h *Handlers) GetFindingsSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	summary := models.FindingsOverviewSummary{
		BySource:  map[string]int64{},
		TopAssets: []models.TopAsset{},
		Trend:     []models.FindingsTrendPoint{},
	}

	err := h.db.QueryRow(ctx,
		`SELECT
			COUNT(*) FILTER (WHERE status IN ('open','in_progress')),
			COUNT(*) FILTER (WHERE status IN ('open','in_progress') AND severity = 'critical'),
			COUNT(*) FILTER (WHERE status = 'resolved' AND last_seen >= CURRENT_DATE),
			COALESCE(AVG(EXTRACT(EPOCH FROM (NOW() - first_seen)) / 86400)
				FILTER (WHERE status IN ('open','in_progress')), 0),
			COUNT(*) FILTER (WHERE status IN ('open','in_progress') AND severity = 'critical'),
			COUNT(*) FILTER (WHERE status IN ('open','in_progress') AND severity = 'high'),
			COUNT(*) FILTER (WHERE status IN ('open','in_progress') AND severity = 'medium'),
			COUNT(*) FILTER (WHERE status IN ('open','in_progress') AND severity = 'low'),
			COUNT(*) FILTER (WHERE status IN ('open','in_progress') AND severity = 'info')
		 FROM findings`,
	).Scan(
		&summary.OpenTotal,
		&summary.CriticalCount,
		&summary.ResolvedToday,
		&summary.AvgOpenAgeDays,
		&summary.BySeverity.Critical,
		&summary.BySeverity.High,
		&summary.BySeverity.Medium,
		&summary.BySeverity.Low,
		&summary.BySeverity.Info,
	)
	if err != nil {
		h.serverErr(w, err)
		return
	}

	if err := h.db.QueryRow(ctx, `SELECT COUNT(*) FROM source_files`).Scan(&summary.TotalIndexed); err != nil {
		h.serverErr(w, err)
		return
	}

	rows, err := h.db.Query(ctx,
		`SELECT COALESCE(NULLIF(source_vendor, ''), source_tool::text) AS source, COUNT(*)
		 FROM findings
		 WHERE status IN ('open','in_progress')
		 GROUP BY source
		 ORDER BY COUNT(*) DESC, source ASC
		 LIMIT 5`)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var source string
		var count int64
		if err := rows.Scan(&source, &count); err == nil {
			summary.BySource[source] = count
		}
	}

	rows, err = h.db.Query(ctx,
		`SELECT COALESCE(NULLIF(affected_asset, ''), 'Unknown asset') AS asset, COUNT(*)
		 FROM findings
		 WHERE status IN ('open','in_progress')
		 GROUP BY asset
		 ORDER BY COUNT(*) DESC, asset ASC
		 LIMIT 5`)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var asset models.TopAsset
		if err := rows.Scan(&asset.Asset, &asset.Count); err == nil {
			summary.TopAssets = append(summary.TopAssets, asset)
		}
	}

	trendByDate := map[string]*models.FindingsTrendPoint{}
	for i := 13; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		point := models.FindingsTrendPoint{Date: day}
		trendByDate[day] = &point
		summary.Trend = append(summary.Trend, point)
	}

	rows, err = h.db.Query(ctx,
		`SELECT ingested_at::date::text AS day,
			COUNT(*) FILTER (WHERE severity = 'critical'),
			COUNT(*) FILTER (WHERE severity = 'high'),
			COUNT(*) FILTER (WHERE severity = 'medium')
		 FROM findings
		 WHERE ingested_at >= CURRENT_DATE - INTERVAL '13 days'
		 GROUP BY day
		 ORDER BY day`)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var day string
		var critical, high, medium int64
		if err := rows.Scan(&day, &critical, &high, &medium); err == nil {
			trendByDate[day] = &models.FindingsTrendPoint{
				Date:     day,
				Critical: critical,
				High:     high,
				Medium:   medium,
			}
		}
	}
	for i, point := range summary.Trend {
		if updated := trendByDate[point.Date]; updated != nil {
			summary.Trend[i] = *updated
		}
	}

	writeJSON(w, http.StatusOK, summary)
}

func (h *Handlers) GetIdentitySummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	summary := models.IdentitySummary{
		LoginTrend: []models.IdentityTrendPoint{},
		RiskyUsers: []models.IdentityUser{},
	}

	err := h.db.QueryRow(ctx,
		`SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE mfa_enabled),
			COUNT(*) FILTER (WHERE NOT mfa_enabled),
			COUNT(*) FILTER (WHERE account_status = 'dormant' OR last_login < NOW() - INTERVAL '90 days'),
			COUNT(*) FILTER (WHERE is_privileged),
			COUNT(*) FILTER (WHERE account_status = 'active'),
			COUNT(*) FILTER (WHERE account_status = 'suspended'),
			COUNT(*) FILTER (WHERE account_status = 'dormant')
		 FROM identity_users`,
	).Scan(
		&summary.TotalUsers,
		&summary.MFAEnabled,
		&summary.MFADisabled,
		&summary.DormantAccounts,
		&summary.PrivilegedAccounts,
		&summary.ByStatus.Active,
		&summary.ByStatus.Suspended,
		&summary.ByStatus.Dormant,
	)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	summary.AccountsWithoutMFA = summary.MFADisabled
	if summary.TotalUsers > 0 {
		summary.MFACoverage = float64(summary.MFAEnabled) / float64(summary.TotalUsers) * 100
	}

	trendByDate := map[string]int64{}
	for i := 29; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		trendByDate[day] = 0
		summary.LoginTrend = append(summary.LoginTrend, models.IdentityTrendPoint{Date: day})
	}
	rows, err := h.db.Query(ctx,
		`SELECT last_login::date::text, COUNT(*)
		 FROM identity_users
		 WHERE last_login >= CURRENT_DATE - INTERVAL '29 days'
		 GROUP BY last_login::date
		 ORDER BY last_login::date`)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	for rows.Next() {
		var day string
		var count int64
		if err := rows.Scan(&day, &count); err == nil {
			trendByDate[day] = count
		}
	}
	rows.Close()
	for i := range summary.LoginTrend {
		summary.LoginTrend[i].Count = trendByDate[summary.LoginTrend[i].Date]
	}

	rows, err = h.db.Query(ctx,
		`SELECT id, user_email, display_name, mfa_enabled, account_status, last_login,
		        is_privileged, groups, sso_apps_count, created_at, source_file_id
		 FROM identity_users
		 WHERE is_privileged AND NOT mfa_enabled
		 ORDER BY last_login ASC
		 LIMIT 10`)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var user models.IdentityUser
		if err := rows.Scan(
			&user.ID, &user.UserEmail, &user.DisplayName, &user.MFAEnabled,
			&user.AccountStatus, &user.LastLogin, &user.IsPrivileged, &user.Groups,
			&user.SSOAppsCount, &user.CreatedAt, &user.SourceFileID,
		); err == nil {
			summary.RiskyUsers = append(summary.RiskyUsers, user)
		}
	}

	writeJSON(w, http.StatusOK, summary)
}

func (h *Handlers) GetPostureSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := scoring.New(h.db).GetSummary(r.Context())
	if err != nil {
		h.serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handlers) EvaluateAlerts(w http.ResponseWriter, r *http.Request) {
	result, err := alerting.New(h.db).Evaluate(r.Context(), alerting.TriggerManual)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handlers) GetAssets(w http.ResponseWriter, r *http.Request) {
	inventory, err := assets.New(h.db).Inventory(r.Context())
	if err != nil {
		h.serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, inventory)
}

func (h *Handlers) GetActions(w http.ResponseWriter, r *http.Request) {
	queue, err := prioritization.New(h.db).Queue(r.Context())
	if err != nil {
		h.serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, queue)
}

func (h *Handlers) GetExecutiveSummary(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	summary, err := executive.New(h.db).GetSummary(r.Context(), days)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handlers) GetExecutiveReportPDF(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	summary, err := executive.New(h.db).GetSummary(r.Context(), days)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	report, err := executive.RenderPDF(summary)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="seccomply-board-report-%s.pdf"`, time.Now().Format("2006-01-02")))
	w.Header().Set("Content-Length", strconv.Itoa(len(report)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(report)
}

func (h *Handlers) GetComplianceSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := compliance.New(h.db).GetSummary(r.Context())
	if err != nil {
		h.serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *Handlers) GetActivityFeed(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	feed, err := activity.New(h.db).Feed(r.Context(), limit, days)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, feed)
}

func (h *Handlers) ListFindings(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	offset, _ := strconv.Atoi(q.Get("offset"))

	var conditions []string
	var args []any
	argN := 1

	if sv := q.Get("severity"); sv != "" {
		sevs := strings.Split(sv, ",")
		placeholders := make([]string, len(sevs))
		for i, s := range sevs {
			args = append(args, strings.TrimSpace(s))
			placeholders[i] = fmt.Sprintf("$%d", argN)
			argN++
		}
		conditions = append(conditions, fmt.Sprintf("severity IN (%s)", strings.Join(placeholders, ",")))
	}

	if st := q.Get("source_tool"); st != "" {
		args = append(args, st)
		conditions = append(conditions, fmt.Sprintf("source_tool = $%d", argN))
		argN++
	}

	if s := q.Get("status"); s != "" {
		args = append(args, s)
		conditions = append(conditions, fmt.Sprintf("status = $%d", argN))
		argN++
	}
	if assignee := q.Get("assignee"); assignee != "" {
		args = append(args, assignee)
		conditions = append(conditions, fmt.Sprintf("assignee = $%d", argN))
		argN++
	}
	if q.Get("overdue") == "true" {
		conditions = append(conditions, `(status NOT IN ('resolved','risk_accepted') AND COALESCE(
			due_date,
			first_seen + CASE severity
				WHEN 'critical' THEN INTERVAL '7 days'
				WHEN 'high' THEN INTERVAL '30 days'
				WHEN 'medium' THEN INTERVAL '60 days'
				ELSE INTERVAL '90 days'
			END
		) < NOW())`)
	}

	if search := q.Get("search"); search != "" {
		args = append(args, "%"+search+"%")
		conditions = append(conditions, fmt.Sprintf(
			"(title ILIKE $%d OR description ILIKE $%d OR affected_asset ILIKE $%d)",
			argN, argN, argN,
		))
		argN++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM findings %s`, where)
	var total int64
	if err := h.db.QueryRow(r.Context(), countQuery, args...).Scan(&total); err != nil {
		h.serverErr(w, err)
		return
	}

	args = append(args, limit, offset)
	dataQuery := fmt.Sprintf(`
		SELECT id, source_tool, source_vendor, external_id, severity, title,
		       description, affected_asset, first_seen, last_seen, status,
		       assignee, due_date, note, raw_payload, ingested_at, source_file_id
		FROM findings %s
		ORDER BY ingested_at DESC, severity DESC
		LIMIT $%d OFFSET $%d`, where, argN, argN+1)

	rows, err := h.db.Query(r.Context(), dataQuery, args...)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	defer rows.Close()

	findings := make([]models.Finding, 0)
	for rows.Next() {
		var f models.Finding
		var rawJSON []byte
		err := rows.Scan(
			&f.ID, &f.SourceTool, &f.SourceVendor, &f.ExternalID,
			&f.Severity, &f.Title, &f.Description, &f.AffectedAsset,
			&f.FirstSeen, &f.LastSeen, &f.Status,
			&f.Assignee, &f.DueDate, &f.Note, &rawJSON, &f.IngestedAt, &f.SourceFileID,
		)
		if err != nil {
			continue
		}
		if rawJSON != nil {
			json.Unmarshal(rawJSON, &f.RawPayload)
		}
		remediation.ApplySLA(&f, time.Now())
		findings = append(findings, f)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"findings": findings,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *Handlers) GetFinding(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}

	f, err := remediation.New(h.db).Get(r.Context(), id)
	if err == pgx.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if err != nil {
		h.serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *Handlers) UpdateFindingWorkflow(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var update remediation.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	finding, err := remediation.New(h.db).Update(r.Context(), id, update)
	if err == pgx.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}
	if err != nil {
		h.serverErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, finding)
}

func (h *Handlers) ListSources(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	var total int64
	if err := h.db.QueryRow(r.Context(), `SELECT COUNT(*) FROM source_files`).Scan(&total); err != nil {
		h.serverErr(w, err)
		return
	}

	rows, err := h.db.Query(r.Context(),
		`SELECT id, filename, sha256, row_count, parse_status, error_log, vendor_matched, ingested_at
		 FROM source_files ORDER BY ingested_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	defer rows.Close()

	sources := make([]models.SourceFile, 0)
	for rows.Next() {
		var s models.SourceFile
		if err := rows.Scan(&s.ID, &s.Filename, &s.SHA256, &s.RowCount,
			&s.ParseStatus, &s.ErrorLog, &s.VendorMatched, &s.IngestedAt); err == nil {
			sources = append(sources, s)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"sources": sources,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *Handlers) GetSystem(w http.ResponseWriter, r *http.Request) {
	watchDir := os.Getenv("WATCH_DIR")
	if watchDir == "" {
		watchDir = "/data/inbox"
	}

	var lastPoll *time.Time
	var lastPollTime time.Time
	if err := h.db.QueryRow(r.Context(),
		`SELECT MAX(created_at) FROM pipeline_events WHERE event_type = 'landed'`,
	).Scan(&lastPollTime); err == nil && !lastPollTime.IsZero() {
		lastPoll = &lastPollTime
	}

	var heartbeatURL = os.Getenv("HEARTBEAT_URL")
	heartbeatStatus := "disabled"
	if heartbeatURL != "" {
		heartbeatStatus = "enabled"
	}

	var deploymentID string
	h.db.QueryRow(r.Context(),
		`SELECT deployment_id FROM deployment LIMIT 1`,
	).Scan(&deploymentID)

	writeJSON(w, http.StatusOK, map[string]any{
		"deployment_id":    deploymentID,
		"version":          version,
		"uptime":           time.Since(startTime).String(),
		"watch_dir":        watchDir,
		"last_poll_at":     lastPoll,
		"heartbeat_status": heartbeatStatus,
	})
}

func (h *Handlers) GetPipeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	today := time.Now().Truncate(24 * time.Hour)

	counts := map[string]int64{}
	rows, err := h.db.Query(ctx,
		`SELECT event_type, COUNT(*) FROM pipeline_events
		 WHERE created_at >= $1 GROUP BY event_type`, today)
	if err != nil {
		h.serverErr(w, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var t string
		var n int64
		if rows.Scan(&t, &n) == nil {
			counts[t] = n
		}
	}

	var lastPollAt *time.Time
	var lp time.Time
	if err := h.db.QueryRow(ctx,
		`SELECT MAX(created_at) FROM pipeline_events`,
	).Scan(&lp); err == nil && !lp.IsZero() {
		lastPollAt = &lp
	}

	writeJSON(w, http.StatusOK, models.PipelineStatus{
		Landed:          counts["landed"],
		Parsed:          counts["parsed"],
		Normalized:      counts["normalized"],
		Indexed:         counts["indexed"],
		FilesLanded:     counts["landed"],
		FilesParsed:     counts["parsed"],
		FilesNormalized: counts["normalized"],
		FilesIndexed:    counts["indexed"],
		FilesToday:      counts["indexed"],
		LastPollAt:      lastPollAt,
	})
}

func (h *Handlers) serverErr(w http.ResponseWriter, err error) {
	h.log.Error("handler error", "err", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	w.Write(buf.Bytes())
}

// heartbeat sends anonymous telemetry every 24h.
func Heartbeat(ctx context.Context, db *pgxpool.Pool) {
	url := os.Getenv("HEARTBEAT_URL")
	if url == "" {
		return
	}

	send := func() {
		var deploymentID string
		var findingsCount int64
		db.QueryRow(ctx, `SELECT deployment_id FROM deployment LIMIT 1`).Scan(&deploymentID)
		db.QueryRow(ctx, `SELECT COUNT(*) FROM findings`).Scan(&findingsCount)

		var installAge int
		var createdAt time.Time
		if err := db.QueryRow(ctx, `SELECT created_at FROM deployment LIMIT 1`).Scan(&createdAt); err == nil {
			installAge = int(time.Since(createdAt).Hours() / 24)
		}

		payload, _ := json.Marshal(map[string]any{
			"deployment_id":    deploymentID,
			"version":          version,
			"install_age_days": installAge,
			"findings_count":   findingsCount,
		})
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
		if err == nil {
			resp.Body.Close()
		}
	}

	send()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			send()
		}
	}
}
