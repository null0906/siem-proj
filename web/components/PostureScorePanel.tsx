"use client";

import dynamic from "next/dynamic";
import { Skeleton } from "@mantine/core";
import type { PostureSummary } from "@/lib/api";
import { CYBER, SEV } from "@/lib/colors";

const PostureTrend = dynamic(() => import("@/components/charts/PostureTrend"), { ssr: false });

function scoreColor(score: number) {
  if (score >= 90) return "#00ff88";
  if (score >= 80) return "#7ee787";
  if (score >= 70) return "#f7c948";
  if (score >= 60) return SEV.HIGH;
  return SEV.CRITICAL;
}

export function PostureScorePanel({ data, loading }: { data?: PostureSummary; loading: boolean }) {
  if (loading || !data) {
    return <Skeleton height={210} radius={0} style={{ marginBottom: 12 }} />;
  }

  const color = scoreColor(data.overall);
  const scoreDegrees = Math.max(0, Math.min(100, data.overall)) * 3.6;

  return (
    <section className="posture-score-panel">
      <div className="posture-score-primary">
        <div
          className="posture-score-gauge"
          style={{ background: `conic-gradient(${color} ${scoreDegrees}deg, rgba(84,101,125,0.18) ${scoreDegrees}deg)` }}
        >
          <div className="posture-score-gauge-inner">
            <span className="posture-score-value" style={{ color }}>{data.overall}</span>
            <span className="posture-score-grade">Grade {data.grade}</span>
          </div>
        </div>
        <div className="posture-score-trend">
          <div className="posture-score-heading">Security posture</div>
          <div className="posture-score-caption">30-day score trend</div>
          <div style={{ height: 86, marginTop: 7 }}>
            <PostureTrend trend={data.trend} color={color} />
          </div>
        </div>
      </div>

      <div className="posture-category-list">
        {data.categories.map((category) => {
          const categoryColor = scoreColor(category.score);
          return (
            <div className="posture-category-row" key={category.name}>
              <div className="posture-category-meta">
                <span>{category.name}</span>
                <span className="posture-category-score">{category.score}</span>
                <span className="posture-category-delta" style={{ color: category.delta >= 0 ? "#00ff88" : SEV.CRITICAL }}>
                  {category.delta > 0 ? "▲" : category.delta < 0 ? "▼" : "•"} {Math.abs(category.delta)}
                </span>
              </div>
              <div className="posture-category-track">
                <span style={{ width: `${category.score}%`, background: categoryColor }} />
              </div>
            </div>
          );
        })}
      </div>

      <div className="posture-insight">
        <span style={{ background: color }} />
        {data.insight}
      </div>
    </section>
  );
}
