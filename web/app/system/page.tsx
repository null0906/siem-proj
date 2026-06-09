"use client";

import { useQuery } from "@tanstack/react-query";
import { apiClient, SystemInfo } from "@/lib/api";

function InfoRow({
  label,
  value,
  accent = false,
  mono = true,
}: {
  label: string;
  value: React.ReactNode;
  accent?: boolean;
  mono?: boolean;
}) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "baseline",
        padding: "10px 0",
        borderBottom: "1px solid #141414",
        gap: 16,
      }}
    >
      <span
        style={{
          fontFamily: "var(--font-sans)",
          fontSize: "var(--fs-label)",
          color: "var(--text-secondary)",
          flexShrink: 0,
          width: 160,
        }}
      >
        {label}
      </span>
      <span
        style={{
          fontFamily: mono ? "var(--font-mono)" : "var(--font-sans)",
          fontSize: "var(--fs-sm)",
          fontVariantNumeric: mono ? "tabular-nums" : undefined,
          color: accent ? "#94d2bd" : "#aaa",
          wordBreak: "break-all",
        }}
      >
        {value}
      </span>
    </div>
  );
}

export default function SystemPage() {
  const { data, isLoading, error, dataUpdatedAt } = useQuery<SystemInfo>({
    queryKey: ["system"],
    queryFn: apiClient.getSystem,
    refetchInterval: 15_000,
  });

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100vh" }}>
      <div
        style={{
          padding: "14px 24px 12px",
          borderBottom: "1px solid #1f1f1f",
          flexShrink: 0,
        }}
      >
        <h1
          style={{
            fontFamily: "var(--font-sans)",
            fontSize: "var(--fs-h1)",
            fontWeight: 600,
            color: "var(--text-title)",
            margin: 0,
          }}
        >
          System
        </h1>
      </div>

      <div style={{ flex: 1, overflow: "auto", padding: "0 24px" }}>
        {error && (
          <div
            style={{
              margin: "20px 0",
              padding: "12px 16px",
              background: "rgba(229,72,77,0.06)",
              border: "1px solid rgba(229,72,77,0.2)",
              fontFamily: "var(--font-sans)",
              fontSize: "var(--fs-sm)",
              color: "#e5484d",
            }}
          >
            API UNREACHABLE — {(error as Error).message}
          </div>
        )}

        <div style={{ maxWidth: 680, paddingTop: 4 }}>
          {/* Deployment */}
          <div
            style={{
              marginBottom: 4,
              paddingTop: 16,
              borderBottom: "1px solid #1f1f1f",
              paddingBottom: 4,
            }}
          >
            <span
              style={{
                fontFamily: "var(--font-sans)",
                fontSize: "var(--fs-title)",
                fontWeight: 600,
                color: "var(--text-title)",
              }}
            >
              Deployment
            </span>
          </div>

          <InfoRow
            label="Deployment ID"
            value={
              isLoading ? "…" : data?.deployment_id || "—"
            }
            accent
          />
          <InfoRow
            label="Version"
            value={isLoading ? "…" : data?.version ?? "—"}
          />
          <InfoRow
            label="Uptime"
            value={isLoading ? "…" : data?.uptime ?? "—"}
          />

          {/* Ingestor */}
          <div
            style={{
              marginBottom: 4,
              paddingTop: 20,
              borderBottom: "1px solid #1f1f1f",
              paddingBottom: 4,
            }}
          >
            <span
              style={{
                fontFamily: "var(--font-sans)",
                fontSize: "var(--fs-title)",
                fontWeight: 600,
                color: "var(--text-title)",
              }}
            >
              Ingestor
            </span>
          </div>

          <InfoRow
            label="Watch Directory"
            value={isLoading ? "…" : data?.watch_dir ?? "—"}
          />
          <InfoRow
            label="Last Poll At"
            value={
              isLoading
                ? "…"
                : data?.last_poll_at
                ? new Date(data.last_poll_at).toLocaleString()
                : "no events yet"
            }
          />

          {/* Telemetry */}
          <div
            style={{
              marginBottom: 4,
              paddingTop: 20,
              borderBottom: "1px solid #1f1f1f",
              paddingBottom: 4,
            }}
          >
            <span
              style={{
                fontFamily: "var(--font-sans)",
                fontSize: "var(--fs-title)",
                fontWeight: 600,
                color: "var(--text-title)",
              }}
            >
              Telemetry
            </span>
          </div>

          <InfoRow
            label="Heartbeat"
            value={
              isLoading ? "…" : (
                <span
                  style={{
                    color:
                      data?.heartbeat_status === "enabled" ? "#4ade80" : "#333",
                  }}
                >
                  {data?.heartbeat_status?.toUpperCase() ?? "—"}
                </span>
              )
            }
          />

          {/* API status */}
          <div
            style={{
              marginBottom: 4,
              paddingTop: 20,
              borderBottom: "1px solid #1f1f1f",
              paddingBottom: 4,
            }}
          >
            <span
              style={{
                fontFamily: "var(--font-sans)",
                fontSize: "var(--fs-title)",
                fontWeight: 600,
                color: "var(--text-title)",
              }}
            >
              API
            </span>
          </div>

          <InfoRow
            label="Status"
            value={
              error ? (
                <span style={{ color: "#e5484d" }}>UNREACHABLE</span>
              ) : (
                <span style={{ color: "#4ade80" }}>CONNECTED</span>
              )
            }
          />
          <InfoRow
            label="Last Fetched"
            value={
              dataUpdatedAt
                ? new Date(dataUpdatedAt).toLocaleTimeString()
                : "—"
            }
          />
        </div>
      </div>
    </div>
  );
}
