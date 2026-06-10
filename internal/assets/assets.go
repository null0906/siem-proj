package assets

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Reference struct {
	EntityID     uuid.UUID
	SourceTool   string
	SourceVendor string
	Severity     string
	Status       string
	Title        string
	OccurredAt   time.Time
}

type Asset struct {
	ID                   uuid.UUID      `json:"id"`
	CanonicalKey         string         `json:"canonical_key"`
	DisplayName          string         `json:"display_name"`
	Hostname             string         `json:"hostname"`
	IPAddress            string         `json:"ip_address"`
	RiskScore            int            `json:"asset_risk_score"`
	FindingCount         int            `json:"finding_count"`
	FindingCountBySource map[string]int `json:"finding_count_by_source"`
	CriticalCount        int            `json:"critical_count"`
	SourceCount          int            `json:"source_count"`
	IsPublic             bool           `json:"is_public"`
	HasEDR               bool           `json:"has_edr"`
	IsEncrypted          bool           `json:"is_encrypted"`
	Owner                string         `json:"owner"`
	OwnerIsPrivileged    bool           `json:"owner_is_privileged"`
	OwnerMFAEnabled      bool           `json:"owner_mfa_enabled"`
	FirstSeen            time.Time      `json:"first_seen"`
	LastSeen             time.Time      `json:"last_seen"`
	References           []ReferenceDTO `json:"references"`
	Summary              string         `json:"summary"`
}

type ReferenceDTO struct {
	EntityID     uuid.UUID `json:"entity_id"`
	SourceTool   string    `json:"source_tool"`
	SourceVendor string    `json:"source_vendor"`
	Severity     string    `json:"severity"`
	Status       string    `json:"status"`
	Title        string    `json:"title"`
	OccurredAt   time.Time `json:"occurred_at"`
}

type Inventory struct {
	Assets  []Asset `json:"assets"`
	Total   int     `json:"total"`
	Insight string  `json:"insight"`
}

type aggregate struct {
	Asset
	References     []Reference
	sources        map[string]bool
	severity       map[string]int
	ownerEntityID  *uuid.UUID
	ownerCreatedAt time.Time
}

type Service struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func Normalize(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(value); err == nil {
		value = host
	} else if strings.Count(value, ":") == 1 {
		parts := strings.SplitN(value, ":", 2)
		value = parts[0]
	}
	value = strings.TrimSuffix(value, ".")
	if net.ParseIP(value) == nil {
		value = strings.SplitN(value, ".", 2)[0]
	}
	return strings.TrimSpace(value)
}

func (s *Service) Refresh(ctx context.Context) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(7392042)`); err != nil {
		return err
	}

	rows, err := tx.Query(ctx, `
		SELECT id, affected_asset, source_tool::text, source_vendor, severity::text,
		       status::text, title, first_seen, last_seen, raw_payload
		FROM findings
		WHERE NULLIF(TRIM(affected_asset), '') IS NOT NULL`)
	if err != nil {
		return err
	}

	aggregates := map[string]*aggregate{}
	for rows.Next() {
		var ref Reference
		var assetName string
		var firstSeen time.Time
		var raw []byte
		if err := rows.Scan(&ref.EntityID, &assetName, &ref.SourceTool, &ref.SourceVendor, &ref.Severity, &ref.Status, &ref.Title, &firstSeen, &ref.OccurredAt, &raw); err != nil {
			rows.Close()
			return err
		}
		key := Normalize(assetName)
		if key == "" {
			continue
		}
		item := aggregates[key]
		if item == nil {
			display := strings.ToUpper(key)
			item = &aggregate{
				Asset: Asset{
					ID:              uuid.New(),
					CanonicalKey:    key,
					DisplayName:     display,
					Hostname:        display,
					IsEncrypted:     true,
					OwnerMFAEnabled: true,
					FirstSeen:       firstSeen,
					LastSeen:        ref.OccurredAt,
				},
				sources:  map[string]bool{},
				severity: map[string]int{},
			}
			if net.ParseIP(key) != nil {
				item.IPAddress = key
				item.Hostname = ""
				item.DisplayName = key
			}
			aggregates[key] = item
		}
		item.References = append(item.References, ref)
		item.sources[ref.SourceVendor] = true
		item.severity[ref.Severity]++
		item.FindingCount++
		if ref.Severity == "critical" && ref.Status == "open" {
			item.CriticalCount++
		}
		if ref.SourceTool == "edr" {
			item.HasEDR = true
		}
		if firstSeen.Before(item.FirstSeen) {
			item.FirstSeen = firstSeen
		}
		if ref.OccurredAt.After(item.LastSeen) {
			item.LastSeen = ref.OccurredAt
		}
		var payload map[string]any
		_ = json.Unmarshal(raw, &payload)
		item.applyPayload(payload)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `TRUNCATE asset_references, assets`); err != nil {
		return err
	}
	keys := make([]string, 0, len(aggregates))
	for key := range aggregates {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		item := aggregates[key]
		item.SourceCount = len(item.sources)
		item.FindingCountBySource = map[string]int{}
		for _, ref := range item.References {
			item.FindingCountBySource[ref.SourceVendor]++
		}
		if item.Owner != "" {
			var ownerID uuid.UUID
			var privileged, mfaEnabled bool
			var createdAt time.Time
			err := tx.QueryRow(ctx, `
				SELECT id, is_privileged, mfa_enabled, created_at
				FROM identity_users
				WHERE LOWER(user_email) = LOWER($1)
				ORDER BY created_at
				LIMIT 1`, item.Owner).Scan(&ownerID, &privileged, &mfaEnabled, &createdAt)
			if err == nil {
				item.OwnerIsPrivileged = privileged
				item.OwnerMFAEnabled = mfaEnabled
				item.ownerEntityID = &ownerID
				item.ownerCreatedAt = createdAt
			} else if err != pgx.ErrNoRows {
				return err
			}
		}
		item.RiskScore = calculateRisk(item)
		counts, _ := json.Marshal(item.FindingCountBySource)
		if _, err := tx.Exec(ctx, `
			INSERT INTO assets
				(id, canonical_key, display_name, hostname, ip_address, asset_risk_score,
				 finding_count, finding_count_by_source, critical_count, source_count,
				 is_public, has_edr, is_encrypted, owner, owner_is_privileged,
				 owner_mfa_enabled, first_seen, last_seen, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,NOW())`,
			item.ID, item.CanonicalKey, item.DisplayName, item.Hostname, item.IPAddress,
			item.RiskScore, item.FindingCount, counts, item.CriticalCount, item.SourceCount,
			item.IsPublic, item.HasEDR, item.IsEncrypted, item.Owner, item.OwnerIsPrivileged,
			item.OwnerMFAEnabled, item.FirstSeen, item.LastSeen,
		); err != nil {
			return err
		}
		for _, ref := range item.References {
			if _, err := tx.Exec(ctx, `
				INSERT INTO asset_references
					(asset_id, entity_type, entity_id, source_tool, source_vendor, severity, status, title, occurred_at)
				VALUES ($1,'finding',$2,$3,$4,$5,$6,$7,$8)`,
				item.ID, ref.EntityID, ref.SourceTool, ref.SourceVendor, ref.Severity, ref.Status, ref.Title, ref.OccurredAt,
			); err != nil {
				return err
			}
		}
		if item.ownerEntityID != nil {
			if _, err := tx.Exec(ctx, `
				INSERT INTO asset_references
					(asset_id, entity_type, entity_id, source_tool, source_vendor, severity, status, title, occurred_at)
				VALUES ($1,'identity_user',$2,'identity','Identity','info','active',$3,$4)`,
				item.ID, *item.ownerEntityID, "Owner: "+item.Owner, item.ownerCreatedAt,
			); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (a *aggregate) applyPayload(payload map[string]any) {
	for rawKey, rawValue := range payload {
		key := strings.ToLower(strings.ReplaceAll(rawKey, " ", "_"))
		value := strings.TrimSpace(fmt.Sprint(rawValue))
		switch key {
		case "public", "is_public", "publicly_exposed":
			a.IsPublic = strings.EqualFold(value, "true") || value == "1" || strings.EqualFold(value, "yes")
		case "encrypted", "is_encrypted":
			if strings.EqualFold(value, "false") || value == "0" || strings.EqualFold(value, "no") {
				a.IsEncrypted = false
			}
		case "has_edr", "edr_installed":
			if strings.EqualFold(value, "true") || value == "1" || strings.EqualFold(value, "yes") {
				a.HasEDR = true
			}
		case "owner", "owner_email", "user_email", "username":
			if a.Owner == "" && strings.Contains(value, "@") {
				a.Owner = strings.ToLower(value)
			}
		case "owner_is_privileged", "privileged_owner":
			a.OwnerIsPrivileged = strings.EqualFold(value, "true") || value == "1" || strings.EqualFold(value, "yes")
		case "owner_mfa_enabled":
			a.OwnerMFAEnabled = !(strings.EqualFold(value, "false") || value == "0" || strings.EqualFold(value, "no"))
		}
	}
}

func calculateRisk(a *aggregate) int {
	risk := a.severity["critical"]*15 + a.severity["high"]*8 + a.severity["medium"]*4 + a.severity["low"]*2 + a.severity["info"]
	if risk > 65 {
		risk = 65
	}
	if a.SourceCount > 1 {
		risk += (a.SourceCount - 1) * 5
	}
	if !a.HasEDR {
		risk += 15
	}
	if !a.IsEncrypted {
		risk += 10
	}
	if a.IsPublic {
		risk += 20
	}
	if a.OwnerIsPrivileged {
		risk += 10
	}
	if risk > 100 {
		return 100
	}
	return risk
}

func (s *Service) Inventory(ctx context.Context) (Inventory, error) {
	if err := s.Refresh(ctx); err != nil {
		return Inventory{}, err
	}
	rows, err := s.db.Query(ctx, `
		SELECT id, canonical_key, display_name, hostname, ip_address, asset_risk_score,
		       finding_count, finding_count_by_source, critical_count, source_count,
		       is_public, has_edr, is_encrypted, owner, owner_is_privileged,
		       owner_mfa_enabled, first_seen, last_seen
		FROM assets
		ORDER BY asset_risk_score DESC, critical_count DESC, display_name`)
	if err != nil {
		return Inventory{}, err
	}
	defer rows.Close()
	inventory := Inventory{Assets: []Asset{}}
	for rows.Next() {
		var asset Asset
		var counts []byte
		if err := rows.Scan(&asset.ID, &asset.CanonicalKey, &asset.DisplayName, &asset.Hostname, &asset.IPAddress, &asset.RiskScore, &asset.FindingCount, &counts, &asset.CriticalCount, &asset.SourceCount, &asset.IsPublic, &asset.HasEDR, &asset.IsEncrypted, &asset.Owner, &asset.OwnerIsPrivileged, &asset.OwnerMFAEnabled, &asset.FirstSeen, &asset.LastSeen); err != nil {
			return Inventory{}, err
		}
		_ = json.Unmarshal(counts, &asset.FindingCountBySource)
		asset.References, err = s.references(ctx, asset.ID)
		if err != nil {
			return Inventory{}, err
		}
		asset.Summary = summary(asset)
		inventory.Assets = append(inventory.Assets, asset)
	}
	inventory.Total = len(inventory.Assets)
	if inventory.Total > 0 {
		top := inventory.Assets[0]
		inventory.Insight = fmt.Sprintf("Your highest-risk asset is %s — appears in %d tools with %d critical issues.", top.DisplayName, top.SourceCount, top.CriticalCount)
	}
	return inventory, rows.Err()
}

func (s *Service) references(ctx context.Context, assetID uuid.UUID) ([]ReferenceDTO, error) {
	rows, err := s.db.Query(ctx, `
		SELECT entity_id, source_tool, source_vendor, severity, status, title, occurred_at
		FROM asset_references WHERE asset_id = $1
		ORDER BY occurred_at DESC, severity
		LIMIT 40`, assetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	refs := []ReferenceDTO{}
	for rows.Next() {
		var ref ReferenceDTO
		if err := rows.Scan(&ref.EntityID, &ref.SourceTool, &ref.SourceVendor, &ref.Severity, &ref.Status, &ref.Title, &ref.OccurredAt); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}

func summary(asset Asset) string {
	parts := []string{}
	if asset.CriticalCount > 0 {
		parts = append(parts, countLabel(asset.CriticalCount, "critical issue", "critical issues"))
	}
	sources := make([]string, 0, len(asset.FindingCountBySource))
	for source := range asset.FindingCountBySource {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	for _, source := range sources {
		count := asset.FindingCountBySource[source]
		parts = append(parts, fmt.Sprintf("%d %s %s", count, source, plural(count, "finding", "findings")))
	}
	if !asset.HasEDR {
		parts = append(parts, "no EDR agent")
	}
	if !asset.IsEncrypted {
		parts = append(parts, "unencrypted")
	}
	if asset.IsPublic {
		parts = append(parts, "publicly exposed")
	}
	if asset.Owner != "" {
		owner := "owned by " + asset.Owner
		if asset.OwnerIsPrivileged {
			owner += " (privileged"
			if !asset.OwnerMFAEnabled {
				owner += ", no MFA"
			}
			owner += ")"
		}
		parts = append(parts, owner)
	}
	return strings.Join(parts, " · ")
}

func countLabel(count int, singular, pluralLabel string) string {
	return fmt.Sprintf("%d %s", count, plural(count, singular, pluralLabel))
}

func plural(count int, singular, pluralLabel string) string {
	if count == 1 {
		return singular
	}
	return pluralLabel
}
