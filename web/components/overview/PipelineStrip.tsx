"use client";

import { useQuery } from "@tanstack/react-query";
import { apiClient, PipelineStatus } from "@/lib/api";

const STAGES = [
  { key: "files_landed", label: "LANDED", icon: "⬇" },
  { key: "files_parsed", label: "PARSED", icon: "⚙" },
  { key: "files_normalized", label: "NORMALIZED", icon: "⊕" },
  { key: "files_indexed", label: "INDEXED", icon: "◈" },
] as const;

function Arrow() {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        gap: 4,
        color: "#2a2a2a",
        flexShrink: 0,
      }}
    >
      <div style={{ height: 1, width: 20, background: "#2a2a2a" }} />
      <svg width="6" height="8" viewBox="0 0 6 8" fill="none">
        <path d="M1 1l4 3-4 3" stroke="#2a2a2a" strokeWidth="1.2" strokeLinecap="round" strokeLinejoin="round" />
      </svg>
    </div>
  );
}

export function PipelineStrip() {
  const { data, isLoading } = useQuery<PipelineStatus>({
    queryKey: ["pipeline"],
    queryFn: apiClient.getPipeline,
    refetchInterval: 15_000,
  });

  return (
    <div
      style={{
        height: 44,
        borderBottom: "1px solid #1f1f1f",
        background: "#0d0d0d",
        display: "flex",
        alignItems: "center",
        paddingLeft: 20,
        paddingRight: 20,
        gap: 0,
        overflow: "hidden",
        flexShrink: 0,
      }}
    >
      {/* Label */}
      <div
        style={{
          fontFamily: "'JetBrains Mono', monospace",
          fontSize: 10,
          color: "#94d2bd",
          letterSpacing: "0.12em",
          textTransform: "uppercase",
          marginRight: 20,
          display: "flex",
          alignItems: "center",
          gap: 6,
          flexShrink: 0,
        }}
      >
        <span className="cursor-blink" style={{ color: "#94d2bd" }}>▋</span>
        INGEST PIPELINE
      </div>

      {/* Stages */}
      <div style={{ display: "flex", alignItems: "center", gap: 0, flex: 1 }}>
        {STAGES.map((stage, i) => {
          const count = data ? (data[stage.key] as number) : 0;
          return (
            <div key={stage.key} style={{ display: "flex", alignItems: "center" }}>
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 6,
                  padding: "0 12px",
                  height: 44,
                  borderRight: i < STAGES.length - 1 ? "none" : undefined,
                }}
              >
                <span style={{ color: "#333", fontSize: 11 }}>{stage.icon}</span>
                <div>
                  <div
                    style={{
                      fontFamily: "'JetBrains Mono', monospace",
                      fontSize: 9,
                      color: "#444",
                      letterSpacing: "0.1em",
                    }}
                  >
                    {stage.label}
                  </div>
                  <div
                    style={{
                      fontFamily: "'JetBrains Mono', monospace",
                      fontSize: 13,
                      color: isLoading ? "#333" : count > 0 ? "#94d2bd" : "#3a3a3a",
                      fontWeight: 500,
                    }}
                  >
                    {isLoading ? "—" : count.toLocaleString()}
                  </div>
                </div>
              </div>
              {i < STAGES.length - 1 && <Arrow />}
            </div>
          );
        })}
      </div>

      {/* Right: files today counter */}
      <div
        style={{
          marginLeft: "auto",
          flexShrink: 0,
          display: "flex",
          alignItems: "center",
          gap: 10,
          paddingLeft: 20,
          borderLeft: "1px solid #1f1f1f",
        }}
      >
        <div>
          <div
            style={{
              fontFamily: "'JetBrains Mono', monospace",
              fontSize: 9,
              color: "#444",
              letterSpacing: "0.1em",
            }}
          >
            FILES TODAY
          </div>
          <div
            style={{
              fontFamily: "'JetBrains Mono', monospace",
              fontSize: 18,
              color: data && data.files_today > 0 ? "#e8e8e8" : "#333",
              fontWeight: 600,
              lineHeight: 1,
            }}
          >
            {isLoading ? "—" : (data?.files_today ?? 0).toString().padStart(3, "0")}
          </div>
        </div>

        {data?.last_poll_at && (
          <div>
            <div
              style={{
                fontFamily: "'JetBrains Mono', monospace",
                fontSize: 9,
                color: "#444",
                letterSpacing: "0.1em",
              }}
            >
              LAST POLL
            </div>
            <div
              style={{
                fontFamily: "'JetBrains Mono', monospace",
                fontSize: 11,
                color: "#555",
              }}
            >
              {new Date(data.last_poll_at).toLocaleTimeString([], {
                hour: "2-digit",
                minute: "2-digit",
                second: "2-digit",
              })}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
