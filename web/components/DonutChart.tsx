"use client";

import { ArcElement, Chart as ChartJS, ChartOptions, Tooltip } from "chart.js";
import { Doughnut } from "react-chartjs-2";
import { CHART, SEV } from "@/lib/colors";

ChartJS.register(ArcElement, Tooltip);

type Counts = Record<"critical" | "high" | "medium" | "low" | "info", number>;

const SEVERITY_LEVELS: Array<keyof typeof SEV> = ["CRITICAL", "HIGH", "MEDIUM", "LOW", "INFO"];

export function DonutChart({ counts }: { counts: Counts }) {
  const values = SEVERITY_LEVELS.map((level) => counts[level.toLowerCase() as keyof Counts] ?? 0);
  const total = values.reduce((sum, count) => sum + count, 0);

  const options: ChartOptions<"doughnut"> = {
    responsive: true,
    maintainAspectRatio: false,
    cutout: "66%",
    animation: {
      duration: 900,
      easing: "easeInOutQuart",
    },
    plugins: {
      legend: { display: false },
      tooltip: { enabled: false },
    },
  };

  return (
    <div style={{ width: 116, height: 116, position: "relative", margin: "0 auto" }}>
      <Doughnut
        data={{
          labels: SEVERITY_LEVELS,
          datasets: [
            {
              data: total > 0 ? values : [1],
              backgroundColor:
                total > 0
                  ? SEVERITY_LEVELS.map((level) => SEV[level])
                  : [CHART.tick],
              borderColor: CHART.border,
              borderWidth: 2,
              hoverOffset: 0,
            },
          ],
        }}
        options={options}
      />
      <div
        style={{
          position: "absolute",
          inset: 0,
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          pointerEvents: "none",
        }}
      >
        <span style={{ fontFamily: "var(--font-mono)", fontSize: 20, fontWeight: 500 }}>
          {total}
        </span>
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 8,
            letterSpacing: "0.12em",
            color: "var(--text-tertiary)",
          }}
        >
          TOTAL
        </span>
      </div>
    </div>
  );
}
