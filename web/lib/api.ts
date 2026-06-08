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
  raw_payload: Record<string, unknown>;
  ingested_at: string;
  source_file_id: string | null;
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
  getSources: (params?: Params) => request<SourcesResponse>("/api/sources", params),
  getSystem: () => request<DeploymentInfo>("/api/system"),
};
