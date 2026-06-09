"use client";

import { ArcElement, Chart as ChartJS, ChartOptions, Tooltip } from "chart.js";
import { Doughnut } from "react-chartjs-2";
import { CHART, SEV } from "@/lib/colors";

ChartJS.register(ArcElement, Tooltip);

type SeverityKey = "critical" | "high" | "medium" | "low" | "info";

const SEVERITIES: Array<{ key: SeverityKey; label: string; color: string }> = [
  { key: "critical", label: "CRIT", color: SEV.CRITICAL },
  { key: "high", label: "HIGH", color: SEV.HIGH },
  { key: "medium", label: "MED", color: SEV.MEDIUM },
  { key: "low", label: "LOW", color: SEV.LOW },
  { key: "info", label: "INFO", color: SEV.INFO },
];

export type SeverityCounts = Record<SeverityKey, number>;

export default function SeverityDonut({
  counts,
  openTotal,
}: {
  counts: SeverityCounts;
  openTotal: number;
}) {
  const values = SEVERITIES.map((severity) => counts[severity.key] ?? 0);
  const total = values.reduce((sum, count) => sum + count, 0);

  const options: ChartOptions<"doughnut"> = {
    responsive: true,
    maintainAspectRatio: false,
    cutout: "65%",
    plugins: {
      legend: { display: false },
      tooltip: {
        ...CHART.tooltip,
      },
    },
  };

  return (
    <>
      <div style={{ width: 110, height: 110, margin: "0 auto", position: "relative" }}>
        <Doughnut
          data={{
            labels: SEVERITIES.map((severity) => severity.label),
            datasets: [
              {
                data: total > 0 ? values : [1],
                backgroundColor: total > 0 ? SEVERITIES.map((severity) => severity.color) : [CHART.tick],
                borderColor: CHART.border,
                borderWidth: 2,
                hoverOffset: 0,
              },
            ],
          }}
          options={options}
        />
        <div className="chart-center-label">
          <span style={{ fontSize: "var(--fs-metric-sm)", fontWeight: 600 }}>{openTotal}</span>
          <span className="chart-center-kicker">OPEN</span>
        </div>
      </div>
      <div className="severity-bars">
        {SEVERITIES.map((severity) => {
          const count = counts[severity.key] ?? 0;
          const width = total > 0 ? `${(count / total) * 100}%` : "0%";

          return (
            <div className="severity-bar-row" key={severity.key}>
              <span style={{ width: 30, color: severity.color }}>{severity.label}</span>
              <span className="severity-bar-track">
                <span
                  className="severity-bar-fill"
                  style={{ width, background: severity.color }}
                />
              </span>
              <span style={{ width: 20, textAlign: "right" }}>{count}</span>
            </div>
          );
        })}
      </div>
    </>
  );
}
