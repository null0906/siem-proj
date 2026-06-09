"use client";

import type React from "react";
import dynamic from "next/dynamic";
import { Skeleton } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import { apiClient, FindingsSummary } from "@/lib/api";
import { CYBER, SEV } from "@/lib/colors";

const TrendChart = dynamic(() => import("@/components/charts/TrendChart"), { ssr: false });
const SourceDonut = dynamic(() => import("@/components/charts/SourceDonut"), { ssr: false });
const AssetsBar = dynamic(() => import("@/components/charts/AssetsBar"), { ssr: false });
const SeverityDonut = dynamic(() => import("@/components/charts/SeverityDonut"), { ssr: false });

const emptySummary: FindingsSummary = {
  open_total: 0,
  critical_count: 0,
  resolved_today: 0,
  avg_open_age_days: 0,
  total_indexed: 0,
  delta_since_yesterday: null,
  by_severity: { critical: 0, high: 0, medium: 0, low: 0, info: 0 },
  by_source: {},
  top_assets: [],
  trend: [],
};

function ageDirection(summary: FindingsSummary) {
  const current = summary.trend.at(-1);
  const previous = summary.trend.at(-2);
  if (!current || !previous) return "steady trend";
  const currentTotal = current.critical + current.high + current.medium;
  const previousTotal = previous.critical + previous.high + previous.medium;
  if (currentTotal > previousTotal) return "↑ trend";
  if (currentTotal < previousTotal) return "↓ trend";
  return "steady trend";
}

function sourceCount(summary: FindingsSummary) {
  return Object.keys(summary.by_source ?? {}).length;
}

export default function OverviewPage() {
  const { data, isLoading, isFetching } = useQuery<FindingsSummary>({
    queryKey: ["findings-summary"],
    queryFn: apiClient.getFindingsSummary,
    staleTime: 30_000,
    refetchInterval: 60_000,
  });

  const summary = data ?? emptySummary;
  const loading = isLoading && !data;

  return (
    <div style={{ padding: "18px 20px 24px", flex: 1, overflow: "auto" }}>
      <header style={{ borderLeft: `2px solid ${CYBER.accent}`, paddingLeft: 12, marginBottom: 16 }}>
        <h1
          style={{
            fontFamily: "var(--font-sans)",
            fontSize: "var(--fs-h1)",
            fontWeight: 600,
            color: "var(--text-title)",
            margin: 0,
            marginBottom: 2,
            lineHeight: 1.2,
          }}
        >
          Security overview
        </h1>
        <p
          style={{
            margin: 0,
            fontFamily: "var(--font-sans)",
            fontSize: "var(--fs-caption)",
            color: "var(--text-secondary)",
            letterSpacing: "0.04em",
          }}
        >
          Posture · source distribution · asset impact &nbsp;·&nbsp;{" "}
          <span style={{ color: CYBER.accent }}>auto-refresh 60s</span>
          {isFetching && !loading ? <span> · refreshing</span> : null}
        </p>
      </header>

      <section className="overview-metrics-grid" style={{ marginBottom: 12 }}>
        {loading ? (
          Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={index} height={92} radius={0} />
          ))
        ) : (
          <>
            <MetricCard
              label="Open findings"
              value={summary.open_total}
              valueColor={SEV.CRITICAL}
              topAccentColor={SEV.CRITICAL}
              delta={
                summary.delta_since_yesterday === null
                  ? null
                  : `${summary.delta_since_yesterday >= 0 ? "↑" : "↓"} ${Math.abs(
                      summary.delta_since_yesterday
                    )} since yesterday`
              }
            />
            <MetricCard
              label="Critical"
              value={summary.critical_count}
              valueColor={SEV.CRITICAL}
              topAccentColor="rgba(255, 59, 59, 0.45)"
              delta={`↓ ${summary.resolved_today} resolved today`}
            />
            <MetricCard
              label="Avg open age"
              value={`${summary.avg_open_age_days.toFixed(1)}d`}
              topAccentColor="rgba(0, 196, 255, 0.40)"
              delta={ageDirection(summary)}
            />
            <MetricCard
              label="Resolved today"
              value={summary.resolved_today}
              valueColor="#00ff88"
              topAccentColor="rgba(0, 255, 136, 0.45)"
              delta={`across ${sourceCount(summary)} sources`}
            />
          </>
        )}
      </section>

      <Panel style={{ marginBottom: 12, minHeight: 222 }}>
        <PanelHeader
          title="Findings trend"
          right={
            <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
              <LegendItem color={SEV.CRITICAL} label="Critical" />
              <LegendItem color={SEV.HIGH} label="High" />
              <LegendItem color={SEV.MEDIUM} label="Medium" />
              <span style={{ color: "var(--text-secondary)" }}>Last 14 days</span>
            </div>
          }
        />
        {loading ? (
          <Skeleton height={160} radius={0} />
        ) : (
          <div style={{ height: 160 }}>
            <TrendChart trend={summary.trend} />
          </div>
        )}
      </Panel>

      <section className="overview-panels-grid">
        <Panel style={{ minHeight: 260 }}>
          <PanelHeader title="Source distribution" />
          {loading ? (
            <Skeleton height={190} radius={0} />
          ) : (
            <SourceDonut bySource={summary.by_source} totalIndexed={summary.total_indexed} />
          )}
        </Panel>

        <Panel style={{ minHeight: 260 }}>
          <PanelHeader title="Top affected assets" />
          {loading ? (
            <Skeleton height={summary.top_assets.length * 40 + 80 || 280} radius={0} />
          ) : (
            <AssetsBar topAssets={summary.top_assets} />
          )}
        </Panel>

        <Panel style={{ minHeight: 260 }}>
          <PanelHeader
            title={
              <span style={{ display: "inline-flex", alignItems: "center", gap: 8 }}>
                <span
                  style={{
                    width: 6,
                    height: 6,
                    borderRadius: "50%",
                    background: SEV.CRITICAL,
                    animation: "pulse-red 2s ease-in-out infinite",
                  }}
                />
                By severity
              </span>
            }
          />
          {loading ? (
            <Skeleton height={214} radius={0} />
          ) : (
            <SeverityDonut counts={summary.by_severity} openTotal={summary.open_total} />
          )}
        </Panel>
      </section>
    </div>
  );
}

function MetricCard({
  label,
  value,
  delta,
  valueColor = "var(--text-primary)",
  topAccentColor,
}: {
  label: string;
  value: string | number;
  delta?: string | null;
  valueColor?: string;
  topAccentColor: string;
}) {
  return (
    <div
      style={{
        background: "var(--cy-panel)",
        borderTop: `1px solid ${topAccentColor}`,
        borderLeft: "1px solid rgba(0,196,255,0.10)",
        borderRight: "1px solid rgba(0,196,255,0.10)",
        borderBottom: "1px solid rgba(0,196,255,0.10)",
        borderRadius: 0,
        padding: "11px 13px",
        minHeight: 92,
      }}
    >
      <div
        style={{
          fontFamily: "var(--font-sans)",
          fontSize: "var(--fs-label)",
          letterSpacing: "0.04em",
          color: "var(--text-secondary)",
          textTransform: "uppercase",
          marginBottom: 5,
        }}
      >
        {label}
      </div>
      <div
        style={{
          fontFamily: "var(--font-sans)",
          fontSize: "var(--fs-metric)",
          fontWeight: 600,
          fontVariantNumeric: "tabular-nums",
          fontFeatureSettings: "'tnum' 1, 'cv05' 1",
          lineHeight: 1,
          marginBottom: 3,
          color: valueColor,
        }}
      >
        {value}
      </div>
      {delta && (
        <div
          style={{
            fontFamily: "var(--font-sans)",
            fontSize: "var(--fs-caption)",
            color: "var(--text-secondary)",
          }}
        >
          {delta}
        </div>
      )}
    </div>
  );
}

function Panel({
  children,
  style,
}: {
  children: React.ReactNode;
  style?: React.CSSProperties;
}) {
  return (
    <section
      style={{
        background: "var(--cy-panel)",
        border: "1px solid rgba(0,196,255,0.10)",
        borderRadius: 0,
        padding: "12px 14px",
        ...style,
      }}
    >
      {children}
    </section>
  );
}

function PanelHeader({
  title,
  right,
}: {
  title: React.ReactNode;
  right?: React.ReactNode;
}) {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        gap: 12,
        marginBottom: 12,
        fontFamily: "var(--font-sans)",
        fontSize: "var(--fs-title)",
        fontWeight: 600,
        color: "var(--text-title)",
      }}
    >
      <span>{title}</span>
      {right && <span style={{ display: "inline-flex", alignItems: "center", gap: 10 }}>{right}</span>}
    </div>
  );
}

function LegendItem({ color, label }: { color: string; label: string }) {
  return (
    <span style={{ display: "inline-flex", alignItems: "center", gap: 6 }}>
      <span style={{ width: 18, height: 1.5, background: color, borderRadius: 1 }} />
      <span>{label}</span>
    </span>
  );
}
