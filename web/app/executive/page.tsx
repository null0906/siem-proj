"use client";

import { useState } from "react";
import { Skeleton } from "@mantine/core";
import { IconDownload, IconTrendingDown, IconTrendingUp } from "@tabler/icons-react";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";
import { SEV } from "@/lib/colors";

const PERIODS = [
  { days: 7, label: "7 days" },
  { days: 30, label: "30 days" },
  { days: 60, label: "60 days" },
];

function scoreColor(score: number) {
  if (score >= 90) return "#00ff88";
  if (score >= 80) return "#7ee787";
  if (score >= 70) return "#f7c948";
  if (score >= 60) return SEV.HIGH;
  return SEV.CRITICAL;
}

export default function ExecutivePage() {
  const [days, setDays] = useState(30);
  const { data, isLoading, isFetching } = useQuery({
    queryKey: ["executive", days],
    queryFn: () => apiClient.getExecutiveSummary(days),
    staleTime: 30_000,
  });

  return (
    <div className="executive-page">
      <header className="executive-header">
        <div>
          <div className="executive-eyebrow">Leadership security brief</div>
          <h1>Executive summary</h1>
          <p>A clear view of security progress, business risk, and readiness.</p>
        </div>
        <div className="executive-header-actions">
          <div className="executive-period-control" aria-label="Reporting period">
            {PERIODS.map((period) => (
              <button
                type="button"
                key={period.days}
                className={days === period.days ? "active" : ""}
                onClick={() => setDays(period.days)}
              >
                {period.label}
              </button>
            ))}
          </div>
          <a className="executive-export" href={apiClient.getExecutiveReportUrl(days)}>
            <IconDownload size={16} />
            Export board report
          </a>
        </div>
      </header>

      {isLoading || !data ? (
        <ExecutiveSkeleton />
      ) : (
        <>
          <section className="executive-hero">
            <div className="executive-score">
              <div className="executive-score-value" style={{ color: scoreColor(data.posture.score) }}>
                {data.posture.score}
              </div>
              <div>
                <div className="executive-grade">Grade {data.posture.grade}</div>
                <div className="executive-score-label">Security posture out of 100</div>
              </div>
            </div>
            <div className="executive-message">
              <span className={data.posture.delta >= 0 ? "positive" : "negative"}>
                {data.posture.delta >= 0 ? <IconTrendingUp size={17} /> : <IconTrendingDown size={17} />}
                {data.posture.delta > 0 ? "+" : ""}{data.posture.delta} points
              </span>
              <p>{data.executive_message}</p>
              <div className="executive-trend" aria-label={`${data.period_label} posture trend`}>
                {data.posture.trend.map((point) => (
                  <span
                    key={point.date}
                    title={`${point.date}: ${point.overall}`}
                    style={{ height: `${Math.max(12, point.overall)}%`, background: scoreColor(point.overall) }}
                  />
                ))}
              </div>
              <small>{data.period_label}{isFetching ? " · refreshing" : ""}</small>
            </div>
          </section>

          <section className="executive-change-section">
            <div className="executive-section-heading">
              <div>
                <div className="executive-eyebrow">Period review</div>
                <h2>What changed</h2>
              </div>
            </div>
            <div className="executive-change-grid">
              <ChangeMetric value={data.changes.new_findings} label="New findings identified" />
              <ChangeMetric value={data.changes.resolved} label="Findings resolved" positive />
              <ChangeMetric value={data.changes.new_criticals} label="New urgent issues" danger={data.changes.new_criticals > 0} />
              <ChangeMetric value={`${data.changes.score_delta > 0 ? "+" : ""}${data.changes.score_delta}`} label="Posture point change" positive={data.changes.score_delta > 0} danger={data.changes.score_delta < 0} />
            </div>
            <div className="executive-highlights">
              {data.changes.highlights.map((highlight) => <p key={highlight}>{highlight}</p>)}
            </div>
          </section>

          <div className="executive-lower-grid">
            <section className="executive-section">
              <div className="executive-section-heading">
                <div>
                  <div className="executive-eyebrow">Leadership attention</div>
                  <h2>Top business risks</h2>
                </div>
              </div>
              <div className="executive-risk-list">
                {data.top_risks.map((risk) => (
                  <div className="executive-risk" key={risk.rank}>
                    <div className="executive-risk-rank">{risk.rank}</div>
                    <div>
                      <h3>{risk.title}</h3>
                      <p>{risk.summary}</p>
                    </div>
                    <div className="executive-risk-score" style={{ color: scoreColor(100 - risk.risk_score) }}>
                      {risk.risk_score}
                      <span>risk</span>
                    </div>
                  </div>
                ))}
              </div>
            </section>

            <section className="executive-section">
              <div className="executive-section-heading">
                <div>
                  <div className="executive-eyebrow">Enterprise readiness</div>
                  <h2>Compliance outlook</h2>
                </div>
              </div>
              <div className="executive-framework-list">
                {data.frameworks.map((framework) => (
                  <div className="executive-framework" key={framework.name}>
                    <div className="executive-framework-meta">
                      <div>
                        <strong>{framework.name}</strong>
                        <span>{framework.state}</span>
                      </div>
                      <b>{framework.score}%</b>
                    </div>
                    <div className="executive-framework-track">
                      <span style={{ width: `${framework.score}%`, background: scoreColor(framework.score) }} />
                    </div>
                  </div>
                ))}
              </div>
              <p className="executive-readiness-note">
                Readiness is estimated from current operational controls and is not an audit opinion.
              </p>
            </section>
          </div>
        </>
      )}
    </div>
  );
}

function ChangeMetric({ value, label, positive, danger }: { value: number | string; label: string; positive?: boolean; danger?: boolean }) {
  return (
    <div className="executive-change-metric">
      <strong style={{ color: danger ? SEV.CRITICAL : positive ? "#00ff88" : "var(--text-title)" }}>{value}</strong>
      <span>{label}</span>
    </div>
  );
}

function ExecutiveSkeleton() {
  return (
    <>
      <Skeleton height={190} radius={0} style={{ marginBottom: 14 }} />
      <Skeleton height={230} radius={0} style={{ marginBottom: 14 }} />
      <Skeleton height={320} radius={0} />
    </>
  );
}
