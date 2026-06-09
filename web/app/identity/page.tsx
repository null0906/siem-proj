"use client";

import type React from "react";
import dynamic from "next/dynamic";
import { Skeleton } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import { apiClient, IdentitySummary, IdentityUser } from "@/lib/api";
import { CYBER, SEV } from "@/lib/colors";
import { Chip } from "@/components/Chip";

const LoginTrend = dynamic(() => import("@/components/charts/IdentityLoginTrend"), { ssr: false });
const MfaDonut = dynamic(() => import("@/components/charts/IdentityMfaDonut"), { ssr: false });
const StatusBar = dynamic(() => import("@/components/charts/IdentityStatusBar"), { ssr: false });

const emptySummary: IdentitySummary = {
  total_users: 0,
  mfa_coverage: 0,
  dormant_accounts: 0,
  privileged_accounts: 0,
  accounts_without_mfa: 0,
  mfa_enabled: 0,
  mfa_disabled: 0,
  by_status: { active: 0, suspended: 0, dormant: 0 },
  login_trend: [],
  risky_users: [],
};

function coverageColor(value: number) {
  if (value > 95) return "#00ff88";
  if (value < 90) return SEV.CRITICAL;
  return "#f7c948";
}

export default function IdentityPage() {
  const { data, isLoading, isFetching } = useQuery<IdentitySummary>({
    queryKey: ["identity-summary"],
    queryFn: apiClient.getIdentitySummary,
    staleTime: 30_000,
    refetchInterval: 60_000,
  });
  const summary = data ?? emptySummary;
  const loading = isLoading && !data;

  return (
    <div style={{ padding: "18px 20px 24px", flex: 1, overflow: "auto" }}>
      <header style={{ borderLeft: `2px solid ${CYBER.accent}`, paddingLeft: 12, marginBottom: 16 }}>
        <h1 style={headingStyle}>Identity & access</h1>
        <p style={subtitleStyle}>
          MFA posture · account lifecycle · privileged access &nbsp;·&nbsp;{" "}
          <span style={{ color: CYBER.accent }}>auto-refresh 60s</span>
          {isFetching && !loading ? <span> · refreshing</span> : null}
        </p>
      </header>

      <section className="overview-metrics-grid" style={{ marginBottom: 12 }}>
        {loading
          ? Array.from({ length: 4 }).map((_, index) => <Skeleton key={index} height={92} radius={0} />)
          : (
            <>
              <MetricCard
                label="MFA coverage"
                value={`${summary.mfa_coverage.toFixed(1)}%`}
                valueColor={coverageColor(summary.mfa_coverage)}
                topAccentColor={coverageColor(summary.mfa_coverage)}
                delta={`${summary.mfa_enabled} of ${summary.total_users} accounts`}
              />
              <MetricCard
                label="Dormant accounts"
                value={summary.dormant_accounts}
                valueColor={SEV.CRITICAL}
                topAccentColor={SEV.CRITICAL}
                delta="no login in 90+ days"
              />
              <MetricCard
                label="Privileged accounts"
                value={summary.privileged_accounts}
                valueColor={CYBER.accent}
                topAccentColor="rgba(0,196,255,0.40)"
                delta="admin or elevated access"
              />
              <MetricCard
                label="Accounts without MFA"
                value={summary.accounts_without_mfa}
                valueColor={SEV.CRITICAL}
                topAccentColor="rgba(255,59,59,0.45)"
                delta="requires enrollment"
              />
            </>
          )}
      </section>

      <Panel style={{ marginBottom: 12, minHeight: 250 }}>
        <PanelHeader title="Login activity trend" right="Last 30 days" />
        {loading ? (
          <Skeleton height={190} radius={0} />
        ) : (
          <div style={{ height: 190 }}>
            <LoginTrend trend={summary.login_trend} />
          </div>
        )}
      </Panel>

      <section className="overview-panels-grid">
        <Panel style={{ minHeight: 290 }}>
          <PanelHeader title="MFA coverage" />
          {loading ? (
            <Skeleton height={210} radius={0} />
          ) : (
            <MfaDonut
              enabled={summary.mfa_enabled}
              disabled={summary.mfa_disabled}
              coverage={summary.mfa_coverage}
            />
          )}
        </Panel>

        <Panel style={{ minHeight: 290 }}>
          <PanelHeader title="Accounts by status" />
          {loading ? <Skeleton height={210} radius={0} /> : <StatusBar status={summary.by_status} />}
        </Panel>

        <Panel style={{ minHeight: 290 }}>
          <PanelHeader title="Privileged accounts without MFA" />
          {loading ? <Skeleton height={220} radius={0} /> : <RiskyUsersTable users={summary.risky_users} />}
        </Panel>
      </section>
    </div>
  );
}

function RiskyUsersTable({ users }: { users: IdentityUser[] }) {
  if (users.length === 0) {
    return <div style={emptyStyle}>No privileged MFA gaps</div>;
  }

  return (
    <div style={{ overflowX: "auto" }}>
      <table style={{ width: "100%", borderCollapse: "collapse", tableLayout: "fixed" }}>
        <thead>
          <tr>
            {["User", "Last login", "Groups"].map((label) => (
              <th key={label} style={tableHeaderStyle}>{label}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {users.map((user) => (
            <tr key={user.id} style={{ borderBottom: "1px solid rgba(0,196,255,0.07)" }}>
              <td style={{ ...tableCellStyle, width: "46%" }}>
                <span style={ellipsisStyle}>{user.user_email}</span>
              </td>
              <td style={{ ...tableCellStyle, width: "24%", color: SEV.CRITICAL }}>
                {new Date(user.last_login).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
              </td>
              <td style={{ ...tableCellStyle, width: "30%" }}>
                <Chip>{user.groups.split(",")[0] || "Admin"}</Chip>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function MetricCard({
  label,
  value,
  delta,
  valueColor,
  topAccentColor,
}: {
  label: string;
  value: string | number;
  delta: string;
  valueColor: string;
  topAccentColor: string;
}) {
  return (
    <div style={{
      background: "var(--cy-panel)",
      borderTop: `1px solid ${topAccentColor}`,
      borderLeft: `1px solid ${CYBER.border}`,
      borderRight: `1px solid ${CYBER.border}`,
      borderBottom: `1px solid ${CYBER.border}`,
      padding: "11px 13px",
      minHeight: 92,
    }}>
      <div style={metricLabelStyle}>{label}</div>
      <div style={{ ...metricValueStyle, color: valueColor }}>{value}</div>
      <div style={metricDeltaStyle}>{delta}</div>
    </div>
  );
}

function Panel({ children, style }: { children: React.ReactNode; style?: React.CSSProperties }) {
  return (
    <section style={{ background: CYBER.bgPanel, border: `1px solid ${CYBER.border}`, padding: "12px 14px", ...style }}>
      {children}
    </section>
  );
}

function PanelHeader({ title, right }: { title: string; right?: string }) {
  return (
    <div style={{ display: "flex", justifyContent: "space-between", gap: 12, marginBottom: 12, fontFamily: "var(--font-sans)", fontSize: "var(--fs-title)", fontWeight: 600, color: "var(--text-title)" }}>
      <span>{title}</span>
      {right ? <span style={{ color: "var(--text-secondary)", fontSize: "var(--fs-caption)", fontWeight: 400 }}>{right}</span> : null}
    </div>
  );
}

const headingStyle: React.CSSProperties = { fontFamily: "var(--font-sans)", fontSize: "var(--fs-h1)", fontWeight: 600, color: "var(--text-title)", margin: "0 0 2px", lineHeight: 1.2 };
const subtitleStyle: React.CSSProperties = { margin: 0, fontFamily: "var(--font-sans)", fontSize: "var(--fs-caption)", color: "var(--text-secondary)", letterSpacing: "0.04em" };
const metricLabelStyle: React.CSSProperties = { fontFamily: "var(--font-sans)", fontSize: "var(--fs-label)", fontWeight: 500, letterSpacing: "0.04em", color: "var(--text-secondary)", textTransform: "uppercase", marginBottom: 5 };
const metricValueStyle: React.CSSProperties = { fontFamily: "var(--font-sans)", fontSize: "var(--fs-metric)", fontWeight: 600, lineHeight: 1, marginBottom: 3, fontVariantNumeric: "tabular-nums", fontFeatureSettings: "'tnum' 1, 'cv05' 1" };
const metricDeltaStyle: React.CSSProperties = { fontFamily: "var(--font-sans)", fontSize: "var(--fs-caption)", color: "var(--text-secondary)" };
const tableHeaderStyle: React.CSSProperties = { padding: "6px 5px", textAlign: "left", fontFamily: "var(--font-sans)", fontSize: "var(--fs-label)", fontWeight: 500, color: "var(--text-secondary)", borderBottom: `1px solid ${CYBER.border}` };
const tableCellStyle: React.CSSProperties = { padding: "8px 5px", fontFamily: "var(--font-mono)", fontSize: "var(--fs-sm)", fontVariantNumeric: "tabular-nums", color: "#8892b0", overflow: "hidden" };
const ellipsisStyle: React.CSSProperties = { display: "block", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" };
const emptyStyle: React.CSSProperties = { padding: "56px 0", textAlign: "center", fontFamily: "var(--font-sans)", fontSize: "var(--fs-body)", color: "var(--text-secondary)" };
