"use client";

import { useQuery } from "@tanstack/react-query";
import { apiClient, Finding } from "@/lib/api";
import { SEVERITY_COLORS, SEVERITY_BG, Severity } from "@/lib/theme";
import { SeverityBadge } from "./SeverityBadge";

function FindingRow({ finding }: { finding: Finding }) {
  const color = SEVERITY_COLORS[finding.severity as Severity] ?? "#6b7280";
  const bg = SEVERITY_BG[finding.severity as Severity] ?? "transparent";

  return (
    <div
      className="fade-in"
      style={{
        borderLeft: `2px solid ${color}`,
        background: bg,
        padding: "10px 14px",
        borderBottom: "1px solid #141414",
        display: "grid",
        gridTemplateColumns: "1fr auto",
        gap: 8,
        cursor: "default",
        transition: "background 120ms",
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLDivElement).style.background = `rgba(255,255,255,0.02)`;
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLDivElement).style.background = bg;
      }}
    >
      <div style={{ minWidth: 0 }}>
        <div
          style={{
            fontFamily: "'Newsreader', Georgia, serif",
            fontSize: 13,
            color: "#d8d8d8",
            fontWeight: 500,
            lineHeight: 1.35,
            marginBottom: 4,
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
          }}
        >
          {finding.title}
        </div>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 10,
            flexWrap: "wrap",
          }}
        >
          <SeverityBadge severity={finding.severity} />
          <span
            style={{
              fontFamily: "'JetBrains Mono', monospace",
              fontSize: 10,
              color: "#555",
            }}
          >
            {finding.source_vendor}
          </span>
          {finding.affected_asset && (
            <span
              style={{
                fontFamily: "'JetBrains Mono', monospace",
                fontSize: 10,
                color: "#555",
              }}
            >
              {finding.affected_asset}
            </span>
          )}
        </div>
      </div>
      <div
        style={{
          fontFamily: "'JetBrains Mono', monospace",
          fontSize: 10,
          color: "#3a3a3a",
          whiteSpace: "nowrap",
          textAlign: "right",
          paddingTop: 2,
        }}
      >
        {new Date(finding.ingested_at).toLocaleTimeString([], {
          hour: "2-digit",
          minute: "2-digit",
        })}
        <br />
        <span style={{ color: "#2a2a2a" }}>
          {new Date(finding.ingested_at).toLocaleDateString([], {
            month: "short",
            day: "numeric",
          })}
        </span>
      </div>
    </div>
  );
}

export function FindingsStream() {
  const { data, isLoading } = useQuery({
    queryKey: ["findings", "stream"],
    queryFn: () => apiClient.getFindings({ limit: "40", status: "open" }),
    refetchInterval: 30_000,
  });

  const findings = data?.findings ?? [];

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <div
        style={{
          padding: "12px 14px 10px",
          borderBottom: "1px solid #1f1f1f",
          display: "flex",
          alignItems: "baseline",
          justifyContent: "space-between",
          gap: 8,
          flexShrink: 0,
        }}
      >
        <h2
          className="serif"
          style={{
            fontFamily: "'Newsreader', Georgia, serif",
            fontSize: 16,
            fontWeight: 500,
            color: "#c8c8c8",
            margin: 0,
            letterSpacing: "-0.01em",
          }}
        >
          Live Findings
        </h2>
        <span
          style={{
            fontFamily: "'JetBrains Mono', monospace",
            fontSize: 10,
            color: "#444",
          }}
        >
          {isLoading ? "—" : `${data?.total ?? 0} open`}
        </span>
      </div>

      <div style={{ flex: 1, overflow: "auto" }}>
        {isLoading &&
          Array.from({ length: 8 }).map((_, i) => (
            <div
              key={i}
              style={{
                height: 56,
                borderBottom: "1px solid #141414",
                borderLeft: "2px solid #1f1f1f",
                background: "#0e0e0e",
                margin: 0,
              }}
            />
          ))}
        {!isLoading && findings.length === 0 && (
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              alignItems: "center",
              justifyContent: "center",
              height: 200,
              gap: 8,
            }}
          >
            <span
              style={{
                fontFamily: "'JetBrains Mono', monospace",
                fontSize: 11,
                color: "#333",
              }}
            >
              NO FINDINGS
            </span>
            <span style={{ fontSize: 10, color: "#2a2a2a" }}>
              Drop a CSV or XLSX into the watch directory
            </span>
          </div>
        )}
        {findings.map((f) => (
          <FindingRow key={f.id} finding={f} />
        ))}
      </div>
    </div>
  );
}
