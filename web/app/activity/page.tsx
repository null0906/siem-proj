"use client";

import { useState } from "react";
import { Skeleton } from "@mantine/core";
import {
  IconAlertTriangle,
  IconArrowDown,
  IconArrowUp,
  IconCheck,
  IconFileImport,
  IconGauge,
  IconServer,
} from "@tabler/icons-react";
import { useQuery } from "@tanstack/react-query";
import { apiClient, PeriodMetrics } from "@/lib/api";
import { CYBER, SEV } from "@/lib/colors";

const PERIODS = [7, 30, 60];

const eventIcons = {
  file_ingested: IconFileImport,
  findings_added: IconAlertTriangle,
  findings_resolved: IconCheck,
  critical_detected: IconAlertTriangle,
  score_changed: IconGauge,
  asset_risk_threshold: IconServer,
} as const;

function severityColor(severity: string) {
  if (severity === "critical") return SEV.CRITICAL;
  if (severity === "warning") return "#f7c948";
  if (severity === "success") return "#00ff88";
  return CYBER.accent;
}

export default function ActivityPage() {
  const [days, setDays] = useState(7);
  const { data, isLoading, isFetching } = useQuery({
    queryKey: ["activity", days],
    queryFn: () => apiClient.getActivity(days),
    staleTime: 15_000,
    refetchInterval: 30_000,
  });

  return (
    <div className="activity-page">
      <header className="activity-header">
        <div>
          <h1>Security activity</h1>
          <p>Real operational changes across ingests, findings, posture, and assets{isFetching && !isLoading ? " · refreshing" : ""}</p>
        </div>
        <div className="activity-periods">
          {PERIODS.map((period) => (
            <button type="button" key={period} className={period === days ? "active" : ""} onClick={() => setDays(period)}>
              {period} days
            </button>
          ))}
        </div>
      </header>

      {isLoading || !data ? (
        <ActivitySkeleton />
      ) : (
        <>
          <section className="activity-comparison">
            <ComparisonMetric label="New findings" current={data.comparison.current.new_findings} delta={data.comparison.delta.new_findings} negative />
            <ComparisonMetric label="Resolved" current={data.comparison.current.resolved_findings} delta={data.comparison.delta.resolved_findings} />
            <ComparisonMetric label="New criticals" current={data.comparison.current.new_criticals} delta={data.comparison.delta.new_criticals} negative />
            <ComparisonMetric label="Files ingested" current={data.comparison.current.files_ingested} delta={data.comparison.delta.files_ingested} />
            <ComparisonMetric label="Score change" current={data.comparison.current.score_delta} delta={data.comparison.delta.score_delta} signed />
          </section>

          <section className="activity-feed-panel">
            <div className="activity-feed-heading">
              <div>
                <h2>Program timeline</h2>
                <p>{data.total} recorded events · newest first</p>
              </div>
              <span>Compared with previous {days} days</span>
            </div>
            <div className="activity-feed-list">
              {data.events.map((event) => {
                const Icon = eventIcons[event.event_type as keyof typeof eventIcons] ?? IconGauge;
                const color = severityColor(event.severity);
                return (
                  <article className="activity-event" key={event.id}>
                    <div className="activity-event-icon" style={{ color, borderColor: `${color}66` }}><Icon size={15} /></div>
                    <div className="activity-event-main">
                      <h3>{event.title}</h3>
                      <p>{event.detail}</p>
                    </div>
                    <time>{new Date(event.created_at).toLocaleString([], { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" })}</time>
                  </article>
                );
              })}
            </div>
          </section>
        </>
      )}
    </div>
  );
}

function ComparisonMetric({ label, current, delta, negative, signed }: { label: string; current: number; delta: number; negative?: boolean; signed?: boolean }) {
  const favorable = negative ? delta <= 0 : delta >= 0;
  return (
    <div className="activity-comparison-metric">
      <strong>{signed && current > 0 ? "+" : ""}{current}</strong>
      <span>{label}</span>
      <small className={favorable ? "positive" : "negative"}>
        {delta > 0 ? <IconArrowUp size={12} /> : delta < 0 ? <IconArrowDown size={12} /> : null}
        {Math.abs(delta)} vs previous period
      </small>
    </div>
  );
}

function ActivitySkeleton() {
  return (
    <>
      <Skeleton height={110} radius={0} style={{ marginBottom: 12 }} />
      <Skeleton height={520} radius={0} />
    </>
  );
}
