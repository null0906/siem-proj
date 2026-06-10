"use client";

import { useState } from "react";
import Link from "next/link";
import { Skeleton } from "@mantine/core";
import { IconArrowRight, IconCheck, IconX } from "@tabler/icons-react";
import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api";
import { SEV } from "@/lib/colors";

function readinessColor(score: number) {
  if (score >= 85) return "#00ff88";
  if (score >= 70) return "#f7c948";
  return SEV.CRITICAL;
}

export default function CompliancePage() {
  const [selected, setSelected] = useState("SOC 2");
  const { data, isLoading, isFetching } = useQuery({
    queryKey: ["compliance"],
    queryFn: apiClient.getComplianceSummary,
    staleTime: 30_000,
    refetchInterval: 60_000,
  });
  const framework = data?.frameworks.find((item) => item.name === selected) ?? data?.frameworks[0];

  return (
    <div className="compliance-page">
      <header className="compliance-header">
        <div>
          <h1>Compliance readiness</h1>
          <p>
            Live controls mapped to current security evidence
            {isFetching && !isLoading ? " · refreshing" : ""}
          </p>
        </div>
        {!isLoading && data ? (
          <div className="compliance-overall">
            <strong>{data.overall_readiness}%</strong>
            <span>{data.passing_controls} of {data.total_controls} controls passing</span>
          </div>
        ) : null}
      </header>

      {isLoading || !data || !framework ? (
        <ComplianceSkeleton />
      ) : (
        <>
          <section className="compliance-framework-grid">
            {data.frameworks.map((item) => {
              const color = readinessColor(item.readiness);
              return (
                <button
                  type="button"
                  key={item.name}
                  className={selected === item.name ? "compliance-framework-card active" : "compliance-framework-card"}
                  onClick={() => setSelected(item.name)}
                >
                  <div className="compliance-framework-top">
                    <div>
                      <strong>{item.name}</strong>
                      <span>{item.controls_away === 0 ? "Audit-ready" : `${item.controls_away} controls away from audit-ready`}</span>
                    </div>
                    <b style={{ color }}>{item.readiness}%</b>
                  </div>
                  <div className="compliance-framework-bar"><span style={{ width: `${item.readiness}%`, background: color }} /></div>
                </button>
              );
            })}
          </section>

          <section className="compliance-controls">
            <div className="compliance-controls-heading">
              <div>
                <h2>{framework.name} controls</h2>
                <p>{framework.passing_controls} passing · {framework.controls_away} blocked</p>
              </div>
              <div className="compliance-readiness-badge" style={{ color: readinessColor(framework.readiness) }}>
                {framework.readiness}% ready
              </div>
            </div>

            <div className="compliance-control-list">
              {framework.controls.map((control) => (
                <article className={`compliance-control compliance-control-${control.status}`} key={control.code}>
                  <div className="compliance-control-status">
                    {control.status === "passing" ? <IconCheck size={16} /> : <IconX size={16} />}
                  </div>
                  <div className="compliance-control-main">
                    <div className="compliance-control-title">
                      <span>{control.code}</span>
                      <h3>{control.title}</h3>
                    </div>
                    <p>{control.summary}</p>
                    {control.blocking_issues.length ? (
                      <div className="compliance-blockers">
                        <strong>This control fails because:</strong>
                        {control.blocking_issues.map((issue) => (
                          <Link href={issue.href} key={issue.gap_type}>
                            <span>{issue.count} {issue.label}</span>
                            <IconArrowRight size={14} />
                          </Link>
                        ))}
                      </div>
                    ) : (
                      <div className="compliance-passing-copy">No mapped open issues are blocking this control.</div>
                    )}
                  </div>
                </article>
              ))}
            </div>
          </section>
        </>
      )}
    </div>
  );
}

function ComplianceSkeleton() {
  return (
    <>
      <div className="compliance-framework-grid">
        {Array.from({ length: 3 }).map((_, index) => <Skeleton key={index} height={112} radius={0} />)}
      </div>
      <Skeleton height={480} radius={0} />
    </>
  );
}
