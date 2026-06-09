"use client";

import { ArcElement, Chart as ChartJS, ChartOptions, Tooltip } from "chart.js";
import { Doughnut } from "react-chartjs-2";
import { CHART, SEV } from "@/lib/colors";

ChartJS.register(ArcElement, Tooltip);

export default function IdentityMfaDonut({
  enabled,
  disabled,
  coverage,
}: {
  enabled: number;
  disabled: number;
  coverage: number;
}) {
  const options: ChartOptions<"doughnut"> = {
    responsive: true,
    maintainAspectRatio: false,
    cutout: "68%",
    plugins: {
      legend: { display: false },
      tooltip: CHART.tooltip,
    },
  };

  return (
    <>
      <div style={{ width: 118, height: 118, margin: "0 auto", position: "relative" }}>
        <Doughnut
          data={{
            labels: ["MFA enabled", "MFA disabled"],
            datasets: [
              {
                data: enabled + disabled > 0 ? [enabled, disabled] : [1],
                backgroundColor: enabled + disabled > 0 ? ["#00ff88", SEV.CRITICAL] : [CHART.tick],
                borderColor: CHART.border,
                borderWidth: 2,
                hoverOffset: 0,
              },
            ],
          }}
          options={options}
        />
        <div className="chart-center-label">
          <span style={{ fontSize: "var(--fs-metric-sm)", fontWeight: 600 }}>{coverage.toFixed(1)}%</span>
          <span className="chart-center-kicker">covered</span>
        </div>
      </div>
      <div className="chart-legend-list">
        <LegendRow color="#00ff88" label="Enabled" count={enabled} />
        <LegendRow color={SEV.CRITICAL} label="Disabled" count={disabled} />
      </div>
    </>
  );
}

function LegendRow({ color, label, count }: { color: string; label: string; count: number }) {
  return (
    <div className="chart-legend-row">
      <span className="chart-legend-name">
        <span className="chart-legend-square" style={{ background: color }} />
        {label}
      </span>
      <span>{count}</span>
    </div>
  );
}
