"use client";

import { BarElement, CategoryScale, Chart as ChartJS, ChartOptions, LinearScale, Tooltip } from "chart.js";
import { Bar } from "react-chartjs-2";
import { CHART } from "@/lib/colors";

ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip);

export interface TopAsset {
  asset: string;
  count: number;
}

const BAR_COLORS = [
  "rgba(255,59,59,0.85)",
  "rgba(255,59,59,0.68)",
  "rgba(255,59,59,0.54)",
  "rgba(255,59,59,0.40)",
  "rgba(255,59,59,0.28)",
];

export default function AssetsBar({ topAssets }: { topAssets: TopAsset[] }) {
  const rows = topAssets.length > 0 ? topAssets : [{ asset: "No affected assets", count: 0 }];
  const height = topAssets.length * 40 + 80;

  const options: ChartOptions<"bar"> = {
    indexAxis: "y",
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: {
        ...CHART.tooltip,
      },
    },
    scales: {
      x: {
        beginAtZero: true,
        grid: { color: CHART.grid },
        ticks: { color: CHART.tick, font: CHART.font },
      },
      y: {
        grid: { display: false },
        ticks: { color: "#4a6a8a", font: CHART.font },
      },
    },
  };

  return (
    <div style={{ height }}>
      <Bar
        data={{
          labels: rows.map((asset) => asset.asset),
          datasets: [
            {
              data: rows.map((asset) => asset.count),
              backgroundColor: BAR_COLORS,
              borderRadius: 0,
              borderWidth: 0,
            },
          ],
        }}
        options={options}
      />
    </div>
  );
}
