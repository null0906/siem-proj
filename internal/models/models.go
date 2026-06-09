package models

import (
	"time"

	"github.com/google/uuid"
)

type SourceTool string

const (
	SourceToolFirewall SourceTool = "firewall"
	SourceToolEDR      SourceTool = "edr"
	SourceToolScanner  SourceTool = "scanner"
	SourceToolDLP      SourceTool = "dlp"
	SourceToolOther    SourceTool = "other"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

type Status string

const (
	StatusOpen       Status = "open"
	StatusResolved   Status = "resolved"
	StatusSuppressed Status = "suppressed"
)

type ParseStatus string

const (
	ParseStatusPending ParseStatus = "pending"
	ParseStatusSuccess ParseStatus = "success"
	ParseStatusFailed  ParseStatus = "failed"
	ParseStatusPartial ParseStatus = "partial"
)

type Finding struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	SourceTool    SourceTool     `json:"source_tool" db:"source_tool"`
	SourceVendor  string         `json:"source_vendor" db:"source_vendor"`
	ExternalID    string         `json:"external_id" db:"external_id"`
	Severity      Severity       `json:"severity" db:"severity"`
	Title         string         `json:"title" db:"title"`
	Description   string         `json:"description" db:"description"`
	AffectedAsset string         `json:"affected_asset" db:"affected_asset"`
	FirstSeen     time.Time      `json:"first_seen" db:"first_seen"`
	LastSeen      time.Time      `json:"last_seen" db:"last_seen"`
	Status        Status         `json:"status" db:"status"`
	RawPayload    map[string]any `json:"raw_payload" db:"raw_payload"`
	IngestedAt    time.Time      `json:"ingested_at" db:"ingested_at"`
	SourceFileID  *uuid.UUID     `json:"source_file_id" db:"source_file_id"`
}

type SourceFile struct {
	ID            uuid.UUID   `json:"id" db:"id"`
	Filename      string      `json:"filename" db:"filename"`
	SHA256        string      `json:"sha256" db:"sha256"`
	RowCount      int         `json:"row_count" db:"row_count"`
	ParseStatus   ParseStatus `json:"parse_status" db:"parse_status"`
	ErrorLog      string      `json:"error_log" db:"error_log"`
	VendorMatched string      `json:"vendor_matched" db:"vendor_matched"`
	IngestedAt    time.Time   `json:"ingested_at" db:"ingested_at"`
}

type IdentityUser struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserEmail     string     `json:"user_email" db:"user_email"`
	DisplayName   string     `json:"display_name" db:"display_name"`
	MFAEnabled    bool       `json:"mfa_enabled" db:"mfa_enabled"`
	AccountStatus string     `json:"account_status" db:"account_status"`
	LastLogin     time.Time  `json:"last_login" db:"last_login"`
	IsPrivileged  bool       `json:"is_privileged" db:"is_privileged"`
	Groups        string     `json:"groups" db:"groups"`
	SSOAppsCount  int        `json:"sso_apps_count" db:"sso_apps_count"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	SourceFileID  *uuid.UUID `json:"source_file_id" db:"source_file_id"`
}

type IdentityTrendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type IdentityStatusCounts struct {
	Active    int64 `json:"active"`
	Suspended int64 `json:"suspended"`
	Dormant   int64 `json:"dormant"`
}

type IdentitySummary struct {
	TotalUsers         int64                `json:"total_users"`
	MFACoverage        float64              `json:"mfa_coverage"`
	DormantAccounts    int64                `json:"dormant_accounts"`
	PrivilegedAccounts int64                `json:"privileged_accounts"`
	AccountsWithoutMFA int64                `json:"accounts_without_mfa"`
	MFAEnabled         int64                `json:"mfa_enabled"`
	MFADisabled        int64                `json:"mfa_disabled"`
	ByStatus           IdentityStatusCounts `json:"by_status"`
	LoginTrend         []IdentityTrendPoint `json:"login_trend"`
	RiskyUsers         []IdentityUser       `json:"risky_users"`
}

type Deployment struct {
	ID           uuid.UUID `json:"id" db:"id"`
	DeploymentID uuid.UUID `json:"deployment_id" db:"deployment_id"`
	Version      string    `json:"version" db:"version"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type DashboardSummary struct {
	TotalFindings  int64            `json:"total_findings"`
	OpenFindings   int64            `json:"open_findings"`
	CriticalCount  int64            `json:"critical_count"`
	HighCount      int64            `json:"high_count"`
	MediumCount    int64            `json:"medium_count"`
	LowCount       int64            `json:"low_count"`
	InfoCount      int64            `json:"info_count"`
	BySourceTool   map[string]int64 `json:"by_source_tool"`
	FilesProcessed int64            `json:"files_processed_today"`
	LastIngestedAt *time.Time       `json:"last_ingested_at"`
}

type SeverityCounts struct {
	Critical int64 `json:"critical"`
	High     int64 `json:"high"`
	Medium   int64 `json:"medium"`
	Low      int64 `json:"low"`
	Info     int64 `json:"info"`
}

type TopAsset struct {
	Asset string `json:"asset"`
	Count int64  `json:"count"`
}

type FindingsTrendPoint struct {
	Date     string `json:"date"`
	Critical int64  `json:"critical"`
	High     int64  `json:"high"`
	Medium   int64  `json:"medium"`
}

type FindingsOverviewSummary struct {
	OpenTotal           int64                `json:"open_total"`
	CriticalCount       int64                `json:"critical_count"`
	ResolvedToday       int64                `json:"resolved_today"`
	AvgOpenAgeDays      float64              `json:"avg_open_age_days"`
	TotalIndexed        int64                `json:"total_indexed"`
	DeltaSinceYesterday *int64               `json:"delta_since_yesterday"`
	BySeverity          SeverityCounts       `json:"by_severity"`
	BySource            map[string]int64     `json:"by_source"`
	TopAssets           []TopAsset           `json:"top_assets"`
	Trend               []FindingsTrendPoint `json:"trend"`
}

type FindingsFilter struct {
	Severity   []Severity
	SourceTool []SourceTool
	Status     []Status
	Search     string
	Limit      int
	Offset     int
}

type PipelineStatus struct {
	Landed          int64      `json:"landed"`
	Parsed          int64      `json:"parsed"`
	Normalized      int64      `json:"normalized"`
	Indexed         int64      `json:"indexed"`
	FilesLanded     int64      `json:"files_landed"`
	FilesParsed     int64      `json:"files_parsed"`
	FilesNormalized int64      `json:"files_normalized"`
	FilesIndexed    int64      `json:"files_indexed"`
	FilesToday      int64      `json:"files_today"`
	LastPollAt      *time.Time `json:"last_poll_at"`
}
