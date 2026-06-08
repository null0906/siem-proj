"use client";

import { ArcElement, Chart as ChartJS, ChartOptions, Tooltip } from "chart.js";
import { Doughnut } from "react-chartjs-2";
import { CHART, srcColor } from "@/lib/colors";

ChartJS.register(ArcElement, Tooltip);

export default function SourceDonut({
  bySource,
  totalIndexed,
}: {
  bySource: Record<string, number>;
  totalIndexed: number;
}) {
  const entries = Object.entries(bySource).slice(0, 5);
  const values = entries.map(([, count]) => count);
  const total = values.reduce((sum, count) => sum + count, 0);
  const colors = entries.map(([source]) => srcColor(source));

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
            labels: entries.map(([source]) => source),
            datasets: [
              {
                data: total > 0 ? values : [1],
                backgroundColor: total > 0 ? colors : [CHART.tick],
                borderColor: CHART.border,
                borderWidth: 2,
                hoverOffset: 0,
              },
            ],
          }}
          options={options}
        />
        <div className="chart-center-label">
          <span style={{ fontSize: 18, fontWeight: 500 }}>{totalIndexed}</span>
          <span className="chart-center-kicker">indexed</span>
        </div>
      </div>
      <div className="chart-legend-list">
        {(entries.length > 0 ? entries : [["No source", 0] as [string, number]]).map(
          ([source, count]) => (
            <div className="chart-legend-row" key={source}>
              <span className="chart-legend-name">
                <span
                  className="chart-legend-square"
                  style={{ background: srcColor(source) }}
                />
                {source}
              </span>
              <span>{count}</span>
            </div>
          )
        )}
      </div>
    </>
  );
}
