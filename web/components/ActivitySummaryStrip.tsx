"use client";

import Link from "next/link";
import type { ActivityFeed } from "@/lib/api";

export function ActivitySummaryStrip({ feed }: { feed: ActivityFeed }) {
  const summary = feed.since_last_ingest;
  if (!summary.available) return null;
  const score = summary.score_delta;

  return (
    <Link href="/activity" className="activity-summary-strip">
      <span className="activity-summary-label">Since last ingest</span>
      <span><b className={summary.new_findings > 0 ? "negative" : ""}>+{summary.new_findings}</b> findings</span>
      <span><b className="positive">−{summary.resolved}</b> resolved</span>
      <span><b className={summary.new_criticals > 0 ? "critical" : ""}>+{summary.new_criticals}</b> critical</span>
      <span><b className={score >= 0 ? "positive" : "negative"}>{score > 0 ? "+" : ""}{score}</b> score</span>
      <span className="activity-summary-time">
        {summary.ingested_at ? new Date(summary.ingested_at).toLocaleString([], { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" }) : ""}
      </span>
    </Link>
  );
}
