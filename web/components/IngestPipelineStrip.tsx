"use client";

import { useQuery } from "@tanstack/react-query";
import { apiClient, PipelineStatus } from "@/lib/api";

const stages = [
  ["landed", "LANDED"],
  ["parsed", "PARSED"],
  ["normalized", "NORMALIZED"],
  ["indexed", "INDEXED"],
] as const;

export function IngestPipelineStrip() {
  const { data } = useQuery<PipelineStatus>({
    queryKey: ["ingest-status"],
    queryFn: apiClient.getPipeline,
    staleTime: 30_000,
    refetchInterval: 60_000,
  });
  const noFilesToday =
    (data?.landed ?? 0) === 0 &&
    (data?.parsed ?? 0) === 0 &&
    (data?.normalized ?? 0) === 0 &&
    (data?.indexed ?? 0) === 0;

  return (
    <div
      style={{
        height: 30,
        width: "100%",
        display: "flex",
        alignItems: "center",
        gap: 10,
        padding: "0 16px",
        borderBottom: "1px solid rgba(0, 196, 255, 0.14)",
        background: "#050a10",
        overflow: "hidden",
        flexShrink: 0,
      }}
    >
      <span
        style={{
          fontFamily: "var(--font-sans)",
          fontSize: "var(--fs-2xs)",
          letterSpacing: "0.04em",
          color: "#2d4560",
          whiteSpace: "nowrap",
        }}
      >
        INGEST PIPELINE
      </span>

      {stages.map(([key, label], index) => (
        <div key={key} style={{ display: "flex", alignItems: "center", gap: 8 }}>
          <span
            style={{
              fontFamily: "var(--font-sans)",
              fontSize: "var(--fs-2xs)",
              letterSpacing: "0.04em",
              color: "#2d4560",
            }}
          >
            {label}
          </span>
          <span
            style={{
              fontFamily: "var(--font-sans)",
              fontSize: "var(--fs-body)",
              fontWeight: 500,
              fontVariantNumeric: "tabular-nums",
              color: noFilesToday ? "#2d4560" : "#00c4ff",
              minWidth: 12,
            }}
          >
            {data?.[key] ?? 0}
          </span>
          {index < stages.length - 1 && (
            <span style={{ color: "#2d4560", fontSize: "var(--fs-sm)" }}>→</span>
          )}
        </div>
      ))}

      <span
        style={{
          marginLeft: "auto",
          fontFamily: "var(--font-sans)",
          fontSize: "var(--fs-2xs)",
          letterSpacing: "0.04em",
          color: "#2d4560",
          whiteSpace: "nowrap",
        }}
      >
        FILES TODAY
      </span>
      <span
        style={{
          fontFamily: "var(--font-sans)",
          fontSize: noFilesToday ? "var(--fs-caption)" : "var(--fs-body)",
          fontWeight: 500,
          fontVariantNumeric: "tabular-nums",
          letterSpacing: noFilesToday ? "0.08em" : "0.18em",
          color: noFilesToday ? "#2d4560" : "#00c4ff",
          whiteSpace: "nowrap",
        }}
      >
        {noFilesToday ? "(no files today)" : (data?.files_today ?? 0).toString().padStart(3, "0")}
      </span>
    </div>
  );
}
