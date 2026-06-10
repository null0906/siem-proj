"use client";

import Link from "next/link";
import type { ActionQueue, PrioritizedAction } from "@/lib/api";

export function ActionQueueCard({ queue, compact = false }: { queue: ActionQueue; compact?: boolean }) {
  const actions = compact ? queue.actions.slice(0, 3) : queue.actions;
  return (
    <section className={compact ? "action-queue-card action-queue-compact" : "action-queue-card"}>
      <div className="action-queue-header">
        <div>
          <div className="action-queue-title">{compact ? "Top 3 actions" : "Prioritized action queue"}</div>
          <div className="action-queue-caption">Ranked by posture-point recovery per effort</div>
        </div>
        {compact ? <Link href="/actions">View queue</Link> : null}
      </div>
      <div className="action-projection">
        <div className="action-projection-copy">
          Complete the top five: <strong>{queue.current_score}</strong> → <strong>{queue.projected_score}</strong>
          <span>+{queue.top_five_gain} posture points</span>
        </div>
        <div className="action-projection-track">
          <span className="action-projection-current" style={{ width: `${queue.current_score}%` }} />
          <span className="action-projection-gain" style={{ left: `${queue.current_score}%`, width: `${queue.top_five_gain}%` }} />
        </div>
      </div>
      <div className="action-list">
        {actions.map((action, index) => <ActionRow key={action.id} action={action} rank={index + 1} compact={compact} />)}
      </div>
      {actions.length === 0 ? <div className="action-empty">No posture-improving actions are currently available.</div> : null}
    </section>
  );
}

function ActionRow({ action, rank, compact }: { action: PrioritizedAction; rank: number; compact: boolean }) {
  return (
    <Link className="action-row" href={action.href}>
      <span className="action-rank">{rank}</span>
      <span className="action-main">
        <strong>{action.title}</strong>
        {!compact ? <span className="action-description">{action.description}</span> : null}
        <span className="action-meta">{action.source} · {action.category} · {action.affected_count} affected</span>
      </span>
      <span className={`action-effort action-effort-${action.effort}`}>{action.effort}</span>
      <span className="action-impact">+{action.score_impact}<small>points</small></span>
    </Link>
  );
}
