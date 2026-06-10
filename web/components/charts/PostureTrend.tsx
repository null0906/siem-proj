"use client";

import { CategoryScale, Chart as ChartJS, Filler, LinearScale, LineElement, PointElement, Tooltip } from "chart.js";
import { Line } from "react-chartjs-2";
import { CHART } from "@/lib/colors";
import type { PostureSnapshot } from "@/lib/api";

ChartJS.register(CategoryScale, Filler, LinearScale, LineElement, PointElement, Tooltip);

export default function PostureTrend({ trend, color }: { trend: PostureSnapshot[]; color: string }) {
  return (
    <Line
      data={{
        labels: trend.map((point) => point.date),
        datasets: [{
          label: "Posture score",
          data: trend.map((point) => point.overall),
          borderColor: color,
          backgroundColor: `${color}18`,
          borderWidth: 2,
          fill: true,
          tension: 0.35,
          pointRadius: 0,
          pointHoverRadius: 3,
        }],
      }}
      options={{
        responsive: true,
        maintainAspectRatio: false,
        interaction: { mode: "index", intersect: false },
        plugins: {
          legend: { display: false },
          tooltip: { ...CHART.tooltip, displayColors: false },
        },
        scales: {
          x: { display: false },
          y: { display: false, min: 0, max: 100 },
        },
      }}
    />
  );
}
