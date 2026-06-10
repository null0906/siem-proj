"use client";

import { Skeleton } from "@mantine/core";
import { useQuery } from "@tanstack/react-query";
import { ActionQueueCard } from "@/components/ActionQueueCard";
import { apiClient } from "@/lib/api";
import { CYBER } from "@/lib/colors";

export default function ActionsPage() {
  const { data, isLoading, isFetching } = useQuery({
    queryKey: ["actions"],
    queryFn: apiClient.getActions,
    staleTime: 30_000,
    refetchInterval: 60_000,
  });

  return (
    <div style={{ padding: "18px 20px 24px", flex: 1, overflow: "auto" }}>
      <header style={{ borderLeft: `2px solid ${CYBER.accent}`, paddingLeft: 12, marginBottom: 16 }}>
        <h1 style={{ margin: "0 0 2px", fontSize: "var(--fs-h1)", fontWeight: 600, lineHeight: 1.2, color: "var(--text-title)" }}>
          Fix this first
        </h1>
        <p style={{ margin: 0, color: "var(--text-secondary)", fontSize: "var(--fs-caption)" }}>
          Highest-impact remediation · effort-aware ranking · live posture projection
          {isFetching && !isLoading ? " · refreshing" : ""}
        </p>
      </header>
      {isLoading || !data ? <Skeleton height={480} radius={0} /> : <ActionQueueCard queue={data} />}
    </div>
  );
}
