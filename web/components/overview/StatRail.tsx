"use client";

import { useQuery } from "@tanstack/react-query";
import { apiClient, DashboardSummary } from "@/lib/api";
import { SEVERITY_COLORS, SOURCE_TOOL_LABELS } from "@/lib/theme";

function StatBlock({
  label,
  value,
  color,
  mono = true,
}: {
  label: string;
  value: string | number;
  color?: string;
  mono?: boolean;
}) {
  return (
    <div
      style={{
        padding: "12px 16px",
        borderBottom: "1px solid #1a1a1a",
      }}
    >
      <div
        style={{
          fontFamily: "'JetBrains Mono', monospace",
          fontSize: 9,
          color: "#444",
          letterSpacing: "0.1em",
          textTransform: "uppercase",
          marginBottom: 4,
        }}
      >
        {label}
      </div>
      <div
        style={{
          fontFamily: mono ? "'JetBrains Mono', monospace" : "Inter, sans-serif",
          fontSize: 22,
          fontWeight: 600,
          color: color ?? "#d0d0d0",
          lineHeight: 1,
        }}
      >
        {typeof value === "number" ? value.toLocaleString() : value}
      </div>
    </div>
  );
}

function SeverityRow({
  label,
  count,
  color,
  total,
}: {
  label: string;
  count: number;
  color: string;
  total: number;
}) {
  const pct = total > 0 ? (count / total) * 100 : 0;
  return (
    <div style={{ padding: "6px 16px" }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "baseline",
          marginBottom: 3,
        }}
      >
        <span
          style={{
            fontFamily: "'JetBrains Mono', monospace",
            fontSize: 10,
            color,
            textTransform: "uppercase",
            letterSpacing: "0.06em",
          }}
        >
          {label}
        </span>
        <span
          style={{
            fontFamily: "'JetBrains Mono', monospace",
            fontSize: 11,
            color: count > 0 ? "#aaa" : "#333",
            fontWeight: 500,
          }}
        >
          {count.toLocaleString()}
        </span>
      </div>
      <div style={{ height: 2, background: "#1a1a1a", borderRadius: 1 }}>
        <div
          style={{
            height: "100%",
            width: `${pct}%`,
            background: color,
            borderRadius: 1,
            opacity: 0.6,
            transition: "width 300ms ease",
          }}
        />
      </div>
    </div>
  );
}

export function StatRail({ data }: { data?: DashboardSummary }) {
  const total = data?.total_findings ?? 0;

  const severityRows = [
    { key: "critical_count", label: "Critical", color: SEVERITY_COLORS.critical },
    { key: "high_count", label: "High", color: SEVERITY_COLORS.high },
    { key: "medium_count", label: "Medium", color: SEVERITY_COLORS.medium },
    { key: "low_count", label: "Low", color: SEVERITY_COLORS.low },
    { key: "info_count", label: "Info", color: SEVERITY_COLORS.info },
  ] as const;

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        borderLeft: "1px solid #1a1a1a",
      }}
    >
      <div
        style={{
          padding: "12px 16px 10px",
          borderBottom: "1px solid #1f1f1f",
          flexShrink: 0,
        }}
      >
        <h3
          style={{
            fontFamily: "'Newsreader', Georgia, serif",
            fontSize: 14,
            fontWeight: 500,
            color: "#888",
            margin: 0,
          }}
        >
          Posture
        </h3>
      </div>

      <StatBlock
        label="Open Findings"
        value={data?.open_findings ?? 0}
        color={data && data.open_findings > 0 ? "#e8e8e8" : "#333"}
      />
      <StatBlock
        label="Total Indexed"
        value={total}
        color="#555"
      />
      <StatBlock
        label="Files Today"
        value={data?.files_processed_today ?? 0}
        color="#94d2bd"
      />

      <div style={{ padding: "12px 16px 6px", borderTop: "1px solid #1a1a1a" }}>
        <div
          style={{
            fontFamily: "'JetBrains Mono', monospace",
            fontSize: 9,
            color: "#444",
            letterSpacing: "0.1em",
            textTransform: "uppercase",
            marginBottom: 8,
          }}
        >
          By Severity
        </div>
        {severityRows.map(({ key, label, color }) => (
          <SeverityRow
            key={key}
            label={label}
            count={data?.[key] ?? 0}
            color={color}
            total={total}
          />
        ))}
      </div>

      <div
        style={{
          padding: "12px 16px 6px",
          borderTop: "1px solid #1a1a1a",
          marginTop: "auto",
        }}
      >
        <div
          style={{
            fontFamily: "'JetBrains Mono', monospace",
            fontSize: 9,
            color: "#444",
            letterSpacing: "0.1em",
            textTransform: "uppercase",
            marginBottom: 8,
          }}
        >
          By Source
        </div>
        {data?.by_source_tool &&
          Object.entries(data.by_source_tool).map(([tool, count]) => (
            <div
              key={tool}
              style={{
                display: "flex",
                justifyContent: "space-between",
                padding: "3px 0",
              }}
            >
              <span
                style={{
                  fontFamily: "'JetBrains Mono', monospace",
                  fontSize: 10,
                  color: "#555",
                }}
              >
                {SOURCE_TOOL_LABELS[tool] ?? tool}
              </span>
              <span
                style={{
                  fontFamily: "'JetBrains Mono', monospace",
                  fontSize: 10,
                  color: "#888",
                  fontWeight: 500,
                }}
              >
                {(count as number).toLocaleString()}
              </span>
            </div>
          ))}
      </div>
    </div>
  );
}
