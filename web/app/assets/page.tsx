"use client";

import { useState } from "react";
import { Skeleton } from "@mantine/core";
import { IconChevronDown, IconChevronRight } from "@tabler/icons-react";
import { useQuery } from "@tanstack/react-query";
import { apiClient, AssetRisk } from "@/lib/api";
import { CYBER, SEV, sevColor, srcColor } from "@/lib/colors";

function riskColor(score: number) {
  if (score >= 80) return SEV.CRITICAL;
  if (score >= 60) return SEV.HIGH;
  if (score >= 40) return SEV.MEDIUM;
  return CYBER.accent;
}

export default function AssetsPage() {
  const [expanded, setExpanded] = useState<string | null>(null);
  const { data, isLoading, isFetching } = useQuery({
    queryKey: ["assets"],
    queryFn: apiClient.getAssets,
    staleTime: 30_000,
    refetchInterval: 60_000,
  });
  const assets = data?.assets ?? [];

  return (
    <div style={{ padding: "18px 20px 24px", flex: 1, overflow: "auto" }}>
      <header style={{ borderLeft: `2px solid ${CYBER.accent}`, paddingLeft: 12, marginBottom: 16 }}>
        <h1 style={{ margin: "0 0 2px", fontSize: "var(--fs-h1)", fontWeight: 600, lineHeight: 1.2, color: "var(--text-title)" }}>
          Asset risk inventory
        </h1>
        <p style={{ margin: 0, color: "var(--text-secondary)", fontSize: "var(--fs-caption)" }}>
          Cross-tool identity resolution · unified risk context · {data?.total ?? 0} canonical assets
          {isFetching && !isLoading ? " · refreshing" : ""}
        </p>
      </header>

      {isLoading ? (
        <Skeleton height={62} radius={0} style={{ marginBottom: 12 }} />
      ) : (
        <div className="asset-insight-banner">
          <span />
          <div>
            <strong>Highest-risk correlation</strong>
            <div>{data?.insight || "No correlated assets yet."}</div>
          </div>
        </div>
      )}

      <section className="asset-inventory-panel">
        <div className="asset-table-wrap">
          <table className="asset-table">
            <thead>
              <tr>
                <th aria-label="Expand" />
                <th>Asset</th>
                <th>Risk score</th>
                <th>Sources</th>
                <th>Critical</th>
                <th>Signals</th>
                <th>Owner</th>
              </tr>
            </thead>
            <tbody>
              {isLoading
                ? Array.from({ length: 8 }).map((_, index) => (
                    <tr key={index}><td colSpan={7}><Skeleton height={34} radius={0} /></td></tr>
                  ))
                : assets.map((asset) => {
                    const open = expanded === asset.id;
                    return (
                      <AssetRow
                        key={asset.id}
                        asset={asset}
                        open={open}
                        onToggle={() => setExpanded(open ? null : asset.id)}
                      />
                    );
                  })}
            </tbody>
          </table>
        </div>
        {!isLoading && assets.length === 0 ? <div className="asset-empty">No assets found in normalized findings.</div> : null}
      </section>
    </div>
  );
}

function AssetRow({ asset, open, onToggle }: { asset: AssetRisk; open: boolean; onToggle: () => void }) {
  const color = riskColor(asset.asset_risk_score);
  const signals = [
    asset.is_public ? { label: "Public", color: SEV.CRITICAL } : null,
    !asset.has_edr ? { label: "No EDR", color: SEV.HIGH } : null,
    !asset.is_encrypted ? { label: "Unencrypted", color: SEV.MEDIUM } : null,
  ].filter(Boolean) as Array<{ label: string; color: string }>;

  return (
    <>
      <tr className="asset-row" onClick={onToggle}>
        <td>{open ? <IconChevronDown size={15} /> : <IconChevronRight size={15} />}</td>
        <td>
          <div className="asset-name">{asset.display_name}</div>
          <div className="asset-summary">{asset.finding_count} findings across {asset.source_count} tools</div>
        </td>
        <td>
          <div className="asset-risk-value" style={{ color }}>{asset.asset_risk_score}</div>
          <div className="asset-risk-track"><span style={{ width: `${asset.asset_risk_score}%`, background: color }} /></div>
        </td>
        <td>
          <div className="asset-source-list">
            {Object.entries(asset.finding_count_by_source).map(([source, count]) => (
              <span key={source} style={{ borderColor: `${srcColor(source)}66`, color: srcColor(source) }}>
                {source} {count}
              </span>
            ))}
          </div>
        </td>
        <td className="asset-number" style={{ color: asset.critical_count > 0 ? SEV.CRITICAL : "var(--text-muted)" }}>
          {asset.critical_count}
        </td>
        <td>
          <div className="asset-signal-list">
            {signals.length ? signals.map((signal) => <span key={signal.label} style={{ color: signal.color }}>{signal.label}</span>) : <span>Healthy controls</span>}
          </div>
        </td>
        <td>
          <div className="asset-owner">{asset.owner || "Unassigned"}</div>
          {asset.owner_is_privileged ? <div className="asset-summary">Privileged{asset.owner_mfa_enabled ? "" : " · no MFA"}</div> : null}
        </td>
      </tr>
      {open ? (
        <tr className="asset-detail-row">
          <td />
          <td colSpan={6}>
            <div className="asset-correlation-summary">{asset.summary}</div>
            <div className="asset-timeline">
              {asset.references.map((ref) => (
                <div className="asset-timeline-item" key={ref.entity_id}>
                  <span className="asset-timeline-dot" style={{ background: sevColor(ref.severity.toUpperCase()) }} />
                  <span className="asset-timeline-time">
                    {new Date(ref.occurred_at).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
                  </span>
                  <span className="asset-source-badge" style={{ color: srcColor(ref.source_vendor) }}>{ref.source_vendor}</span>
                  <span className="asset-timeline-title">{ref.title}</span>
                  <span className="asset-timeline-status">{ref.status}</span>
                </div>
              ))}
            </div>
          </td>
        </tr>
      ) : null}
    </>
  );
}
