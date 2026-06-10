const BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

type Params = Record<string, string | number | undefined | null>;

async function request<T>(path: string, params?: Params): Promise<T> {
  const url = new URL(`${BASE}${path}`);
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== "") {
        url.searchParams.set(key, String(value));
      }
    });
  }

  const res = await fetch(url.toString(), { cache: "no-store" });
  if (!res.ok) {
    const msg = await res.text().catch(() => res.statusText);
    throw new Error(`API ${path} -> ${res.status}: ${msg}`);
  }
  return res.json();
}

async function mutate<T>(path: string, method: "PATCH" | "POST", body: unknown): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    const msg = await res.text().catch(() => res.statusText);
    throw new Error(`API ${path} -> ${res.status}: ${msg}`);
  }
  return res.json();
}

async function requestWithFallback<T>(paths: string[], params?: Params): Promise<T> {
  let lastError: unknown;
  for (const path of paths) {
    try {
      return await request<T>(path, params);
    } catch (error) {
      lastError = error;
      if (!(error instanceof Error) || !error.message.includes("-> 404:")) {
        throw error;
      }
    }
  }
  throw lastError;
}

export interface DashboardSummary {
  total_findings: number;
  open_findings: number;
  critical_count: number;
  high_count: number;
  medium_count: number;
  low_count: number;
  info_count: number;
  by_source_tool: Record<string, number>;
  files_processed_today: number;
  last_ingested_at: string | null;
}

export interface FindingsSummary {
  open_total: number;
  critical_count: number;
  resolved_today: number;
  avg_open_age_days: number;
  total_indexed: number;
  delta_since_yesterday: number | null;
  by_severity: {
    critical: number;
    high: number;
    medium: number;
    low: number;
    info: number;
  };
  by_source: Record<string, number>;
  top_assets: Array<{ asset: string; count: number }>;
  trend: Array<{ date: string; critical: number; high: number; medium: number }>;
  total?: number;
  sources_active?: number;
  open_findings?: number;
  files_today?: number;
  last_ingested_at?: string | null;
}

export interface IdentityUser {
  id: string;
  user_email: string;
  display_name: string;
  mfa_enabled: boolean;
  account_status: "active" | "suspended" | "dormant";
  last_login: string;
  is_privileged: boolean;
  groups: string;
  sso_apps_count: number;
  created_at: string;
  source_file_id: string | null;
}

export interface IdentitySummary {
  total_users: number;
  mfa_coverage: number;
  dormant_accounts: number;
  privileged_accounts: number;
  accounts_without_mfa: number;
  mfa_enabled: number;
  mfa_disabled: number;
  by_status: {
    active: number;
    suspended: number;
    dormant: number;
  };
  login_trend: Array<{ date: string; count: number }>;
  risky_users: IdentityUser[];
}

export interface PostureCategory {
  name: string;
  score: number;
  delta: number;
  issue_count: number;
  point_drag: number;
}

export interface PostureSnapshot {
  date: string;
  overall: number;
  vulnerability: number;
  identity: number;
  endpoint: number;
  cloud: number;
  compliance: number;
}

export interface PostureSummary {
  overall: number;
  grade: string;
  categories: PostureCategory[];
  trend: PostureSnapshot[];
  biggest_drag: PostureCategory;
  insight: string;
}

export interface AssetReference {
  entity_id: string;
  source_tool: string;
  source_vendor: string;
  severity: string;
  status: string;
  title: string;
  occurred_at: string;
}

export interface AssetRisk {
  id: string;
  canonical_key: string;
  display_name: string;
  hostname: string;
  ip_address: string;
  asset_risk_score: number;
  finding_count: number;
  finding_count_by_source: Record<string, number>;
  critical_count: number;
  source_count: number;
  is_public: boolean;
  has_edr: boolean;
  is_encrypted: boolean;
  owner: string;
  owner_is_privileged: boolean;
  owner_mfa_enabled: boolean;
  first_seen: string;
  last_seen: string;
  references: AssetReference[];
  summary: string;
}

export interface AssetInventory {
  assets: AssetRisk[];
  total: number;
  insight: string;
}

export interface PrioritizedAction {
  id: string;
  title: string;
  description: string;
  score_impact: number;
  effort: "low" | "medium" | "high";
  effort_weight: number;
  priority: number;
  affected_count: number;
  source: string;
  category: string;
  href: string;
}

export interface ActionQueue {
  actions: PrioritizedAction[];
  current_score: number;
  projected_score: number;
  top_five_gain: number;
}

export interface ExecutiveSummary {
  generated_at: string;
  period_days: number;
  period_label: string;
  posture: {
    score: number;
    grade: string;
    start_score: number;
    delta: number;
    trend: PostureSnapshot[];
  };
  changes: {
    new_findings: number;
    resolved: number;
    new_criticals: number;
    score_delta: number;
    highlights: string[];
  };
  top_risks: Array<{
    rank: number;
    title: string;
    summary: string;
    risk_score: number;
    critical_count: number;
    source_count: number;
  }>;
  frameworks: Array<{
    name: string;
    score: number;
    state: string;
  }>;
  executive_message: string;
}

export interface ComplianceSummary {
  overall_readiness: number;
  passing_controls: number;
  total_controls: number;
  frameworks: Array<{
    name: string;
    readiness: number;
    passing_controls: number;
    total_controls: number;
    controls_away: number;
    audit_ready: boolean;
    controls: Array<{
      code: string;
      title: string;
      summary: string;
      status: "passing" | "failing";
      blocking_issues: Array<{
        gap_type: string;
        label: string;
        count: number;
        href: string;
      }>;
    }>;
  }>;
}

export interface ActivityFeed {
  total: number;
  events: Array<{
    id: number;
    event_type: string;
    title: string;
    detail: string;
    severity: "info" | "success" | "warning" | "critical";
    entity_type: string;
    entity_id: string;
    metadata: Record<string, unknown>;
    created_at: string;
  }>;
  since_last_ingest: {
    available: boolean;
    filename: string;
    ingested_at: string | null;
    new_findings: number;
    resolved: number;
    new_criticals: number;
    score_delta: number;
    high_risk_assets: number;
  };
  comparison: {
    current: PeriodMetrics;
    previous: PeriodMetrics;
    delta: PeriodMetrics;
  };
}

export interface PeriodMetrics {
  days: number;
  new_findings: number;
  resolved_findings: number;
  new_criticals: number;
  files_ingested: number;
  score_delta: number;
}

export interface Finding {
  id: string;
  source_tool: string;
  source_vendor: string;
  external_id: string;
  severity: string;
  title: string;
  description: string;
  affected_asset: string;
  first_seen: string;
  last_seen: string;
  status: string;
  assignee: string;
  due_date: string | null;
  note: string;
  sla_status: "on_track" | "due_soon" | "breached" | "closed";
  sla_due_at: string;
  events?: FindingEvent[];
  raw_payload: Record<string, unknown>;
  ingested_at: string;
  source_file_id: string | null;
}

export interface FindingEvent {
  id: number;
  event_type: string;
  actor: string;
  from_value: string;
  to_value: string;
  note: string;
  created_at: string;
}

export interface FindingsResponse {
  findings: Finding[];
  total: number;
  limit: number;
  offset?: number;
  page?: number;
}

export interface SourceFile {
  id: string;
  filename: string;
  sha256: string;
  row_count: number;
  parse_status: string;
  error_log: string;
  vendor_matched: string;
  ingested_at: string;
}

export interface SourcesResponse {
  sources: SourceFile[];
  total: number;
  limit: number;
  offset: number;
}

export interface PipelineStatus {
  landed: number;
  parsed: number;
  normalized: number;
  indexed: number;
  files_landed: number;
  files_parsed: number;
  files_normalized: number;
  files_indexed: number;
  files_today: number;
  last_poll_at: string | null;
}

interface LegacyPipelineStatus {
  files_landed: number;
  files_parsed: number;
  files_normalized: number;
  files_indexed: number;
  files_today: number;
  last_poll_at: string | null;
}

export interface DeploymentInfo {
  deployment_id: string;
  version: string;
  uptime_seconds?: number;
  uptime?: string;
  watch_dir?: string;
  last_poll_at?: string | null;
  heartbeat_status?: string;
}

export type SystemInfo = DeploymentInfo;

function normalizeSummary(data: FindingsSummary | DashboardSummary): FindingsSummary {
  if ("open_total" in data) return data;

  return {
    total: data.total_findings,
    open_total: data.open_findings,
    critical_count: data.critical_count,
    resolved_today: 0,
    avg_open_age_days: 0,
    total_indexed: data.files_processed_today,
    delta_since_yesterday: null,
    open_findings: data.open_findings,
    sources_active: Object.keys(data.by_source_tool ?? {}).length,
    files_today: data.files_processed_today,
    last_ingested_at: data.last_ingested_at,
    by_source: data.by_source_tool,
    top_assets: [],
    trend: [],
    by_severity: {
      critical: data.critical_count,
      high: data.high_count,
      medium: data.medium_count,
      low: data.low_count,
      info: data.info_count,
    },
  };
}

function normalizePipeline(data: PipelineStatus | LegacyPipelineStatus): PipelineStatus {
  if ("landed" in data) {
    return {
      ...data,
      files_landed: data.files_landed ?? data.landed,
      files_parsed: data.files_parsed ?? data.parsed,
      files_normalized: data.files_normalized ?? data.normalized,
      files_indexed: data.files_indexed ?? data.indexed,
    };
  }

  return {
    landed: data.files_landed,
    parsed: data.files_parsed,
    normalized: data.files_normalized,
    indexed: data.files_indexed,
    files_landed: data.files_landed,
    files_parsed: data.files_parsed,
    files_normalized: data.files_normalized,
    files_indexed: data.files_indexed,
    files_today: data.files_today,
    last_poll_at: data.last_poll_at,
  };
}

function pageToOffset(params?: Params): Params | undefined {
  if (!params?.page || !params.limit) return params;
  const page = Number(params.page);
  const limit = Number(params.limit);
  return { ...params, offset: Math.max(0, page - 1) * limit };
}

export const apiClient = {
  getDashboard: () => request<DashboardSummary>("/api/dashboard"),
  getFindingsSummary: async () =>
    normalizeSummary(
      await requestWithFallback<FindingsSummary | DashboardSummary>([
        "/api/findings/summary",
        "/api/dashboard",
      ])
    ),
  getIdentitySummary: () => request<IdentitySummary>("/api/identity/summary"),
  getPostureSummary: () => request<PostureSummary>("/api/posture/summary"),
  getAssets: () => request<AssetInventory>("/api/assets"),
  getActions: () => request<ActionQueue>("/api/actions"),
  getExecutiveSummary: (days = 30) => request<ExecutiveSummary>("/api/executive/summary", { days }),
  getExecutiveReportUrl: (days = 30) => `${BASE}/api/executive/report.pdf?days=${days}`,
  getComplianceSummary: () => request<ComplianceSummary>("/api/compliance/summary"),
  getActivity: (days = 7, limit = 50) => request<ActivityFeed>("/api/activity", { days, limit }),
  getPipeline: async () =>
    normalizePipeline(
      await requestWithFallback<PipelineStatus | LegacyPipelineStatus>([
        "/api/ingest/status",
        "/api/pipeline",
      ])
    ),
  getDeployment: () =>
    requestWithFallback<DeploymentInfo>(["/api/deployment", "/api/system"]),
  getFindings: (params?: Params) =>
    request<FindingsResponse>("/api/findings", pageToOffset(params)),
  getFinding: (id: string) => request<Finding>(`/api/findings/${id}`),
  updateFindingWorkflow: (id: string, update: { actor?: string; assignee?: string; due_date?: string; status?: string; note?: string }) =>
    mutate<Finding>(`/api/findings/${id}/workflow`, "PATCH", update),
  getSources: (params?: Params) => request<SourcesResponse>("/api/sources", params),
  getSystem: () => request<DeploymentInfo>("/api/system"),
};
