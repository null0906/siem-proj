"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { apiClient, SourceFile } from "@/lib/api";

const PAGE_SIZE = 50;

function ParseStatusBadge({ status }: { status: string }) {
  const map: Record<string, { color: string; label: string }> = {
    success: { color: "#4ade80", label: "SUCCESS" },
    failed: { color: "#e5484d", label: "FAILED" },
    partial: { color: "#f5d75e", label: "PARTIAL" },
    pending: { color: "#555", label: "PENDING" },
  };
  const { color, label } = map[status] ?? { color: "#555", label: status.toUpperCase() };

  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 4,
        fontFamily: "'JetBrains Mono', monospace",
        fontSize: 10,
        color,
        textTransform: "uppercase",
        letterSpacing: "0.06em",
      }}
    >
      <span
        style={{
          width: 5,
          height: 5,
          borderRadius: "50%",
          background: color,
          opacity: 0.8,
        }}
      />
      {label}
    </span>
  );
}

export default function SourcesPage() {
  const [offset, setOffset] = useState(0);

  const { data, isLoading } = useQuery({
    queryKey: ["sources", offset],
    queryFn: () => apiClient.getSources({ limit: String(PAGE_SIZE), offset: String(offset) }),
    placeholderData: (prev) => prev,
  });

  const sources = data?.sources ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / PAGE_SIZE);
  const currentPage = Math.floor(offset / PAGE_SIZE) + 1;

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100vh" }}>
      {/* Header */}
      <div
        style={{
          padding: "14px 24px 12px",
          borderBottom: "1px solid #1f1f1f",
          flexShrink: 0,
          display: "flex",
          alignItems: "baseline",
          justifyContent: "space-between",
        }}
      >
        <h1
          style={{
            fontFamily: "'Newsreader', Georgia, serif",
            fontSize: 20,
            fontWeight: 500,
            color: "#c8c8c8",
            margin: 0,
          }}
        >
          Ingested Sources
        </h1>
        <span
          style={{
            fontFamily: "'JetBrains Mono', monospace",
            fontSize: 11,
            color: "#444",
          }}
        >
          {isLoading ? "…" : `${total.toLocaleString()} files`}
        </span>
      </div>

      {/* Table */}
      <div style={{ flex: 1, overflow: "auto" }}>
        <table
          style={{
            width: "100%",
            borderCollapse: "collapse",
            tableLayout: "fixed",
          }}
        >
          <thead>
            <tr style={{ borderBottom: "1px solid #1f1f1f" }}>
              {[
                { label: "FILENAME", width: "30%" },
                { label: "VENDOR", width: "12%" },
                { label: "ROWS", width: "8%" },
                { label: "STATUS", width: "10%" },
                { label: "INGESTED", width: "14%" },
                { label: "SHA-256", width: "14%" },
                { label: "ERROR", width: "12%" },
              ].map(({ label, width }) => (
                <th
                  key={label}
                  style={{
                    width,
                    padding: "8px 12px",
                    textAlign: "left",
                    fontFamily: "'JetBrains Mono', monospace",
                    fontSize: 9,
                    fontWeight: 600,
                    letterSpacing: "0.1em",
                    color: "#444",
                    background: "#0d0d0d",
                    position: "sticky",
                    top: 0,
                    borderBottom: "1px solid #1f1f1f",
                  }}
                >
                  {label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {isLoading &&
              Array.from({ length: 10 }).map((_, i) => (
                <tr key={i} style={{ borderBottom: "1px solid #111" }}>
                  {Array.from({ length: 7 }).map((_, j) => (
                    <td key={j} style={{ padding: "9px 12px" }}>
                      <div style={{ height: 12, background: "#141414", borderRadius: 2 }} />
                    </td>
                  ))}
                </tr>
              ))}
            {!isLoading && sources.length === 0 && (
              <tr>
                <td
                  colSpan={7}
                  style={{
                    padding: "60px 0",
                    textAlign: "center",
                    fontFamily: "'JetBrains Mono', monospace",
                    fontSize: 11,
                    color: "#333",
                  }}
                >
                  NO FILES INGESTED YET
                </td>
              </tr>
            )}
            {!isLoading &&
              sources.map((s: SourceFile) => (
                <tr
                  key={s.id}
                  style={{
                    borderBottom: "1px solid #111",
                    transition: "background 80ms",
                  }}
                  onMouseEnter={(e) => {
                    (e.currentTarget as HTMLTableRowElement).style.background = "#0f0f0f";
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLTableRowElement).style.background = "transparent";
                  }}
                >
                  <td style={{ padding: "9px 12px", overflow: "hidden" }}>
                    <span
                      style={{
                        fontFamily: "'JetBrains Mono', monospace",
                        fontSize: 12,
                        color: "#c0c0c0",
                        display: "block",
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        whiteSpace: "nowrap",
                      }}
                    >
                      {s.filename}
                    </span>
                  </td>
                  <td style={{ padding: "9px 12px" }}>
                    <span
                      style={{
                        fontFamily: "'JetBrains Mono', monospace",
                        fontSize: 11,
                        color: "#666",
                      }}
                    >
                      {s.vendor_matched || "—"}
                    </span>
                  </td>
                  <td style={{ padding: "9px 12px" }}>
                    <span
                      style={{
                        fontFamily: "'JetBrains Mono', monospace",
                        fontSize: 12,
                        color: "#888",
                      }}
                    >
                      {s.row_count.toLocaleString()}
                    </span>
                  </td>
                  <td style={{ padding: "9px 12px" }}>
                    <ParseStatusBadge status={s.parse_status} />
                  </td>
                  <td style={{ padding: "9px 12px" }}>
                    <span
                      style={{
                        fontFamily: "'JetBrains Mono', monospace",
                        fontSize: 10,
                        color: "#444",
                      }}
                    >
                      {new Date(s.ingested_at).toLocaleString([], {
                        month: "short",
                        day: "numeric",
                        hour: "2-digit",
                        minute: "2-digit",
                      })}
                    </span>
                  </td>
                  <td style={{ padding: "9px 12px", overflow: "hidden" }}>
                    <span
                      style={{
                        fontFamily: "'JetBrains Mono', monospace",
                        fontSize: 9,
                        color: "#333",
                        display: "block",
                        overflow: "hidden",
                        textOverflow: "ellipsis",
                        whiteSpace: "nowrap",
                      }}
                      title={s.sha256}
                    >
                      {s.sha256.slice(0, 16)}…
                    </span>
                  </td>
                  <td style={{ padding: "9px 12px", overflow: "hidden" }}>
                    {s.error_log ? (
                      <span
                        style={{
                          fontFamily: "'JetBrains Mono', monospace",
                          fontSize: 10,
                          color: "#e5484d",
                          display: "block",
                          overflow: "hidden",
                          textOverflow: "ellipsis",
                          whiteSpace: "nowrap",
                        }}
                        title={s.error_log}
                      >
                        {s.error_log}
                      </span>
                    ) : (
                      <span style={{ color: "#2a2a2a", fontSize: 10 }}>—</span>
                    )}
                  </td>
                </tr>
              ))}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {total > PAGE_SIZE && (
        <div
          style={{
            padding: "8px 24px",
            borderTop: "1px solid #1f1f1f",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            flexShrink: 0,
          }}
        >
          <span
            style={{
              fontFamily: "'JetBrains Mono', monospace",
              color: "#444",
              fontSize: 11,
            }}
          >
            Page {currentPage} of {totalPages}
          </span>
          <div style={{ display: "flex", gap: 6 }}>
            <button
              onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
              disabled={offset === 0}
              style={{
                background: "transparent",
                border: "1px solid #1f1f1f",
                color: offset === 0 ? "#2a2a2a" : "#888",
                padding: "4px 12px",
                cursor: offset === 0 ? "not-allowed" : "pointer",
                fontFamily: "'JetBrains Mono', monospace",
                fontSize: 11,
              }}
            >
              ← prev
            </button>
            <button
              onClick={() => setOffset(offset + PAGE_SIZE)}
              disabled={offset + PAGE_SIZE >= total}
              style={{
                background: "transparent",
                border: "1px solid #1f1f1f",
                color: offset + PAGE_SIZE >= total ? "#2a2a2a" : "#888",
                padding: "4px 12px",
                cursor: offset + PAGE_SIZE >= total ? "not-allowed" : "pointer",
                fontFamily: "'JetBrains Mono', monospace",
                fontSize: 11,
              }}
            >
              next →
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
