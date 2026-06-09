"use client";

import {
  CategoryScale,
  Chart as ChartJS,
  ChartOptions,
  Filler,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip,
} from "chart.js";
import { Line } from "react-chartjs-2";
import { CHART, CYBER } from "@/lib/colors";

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Filler, Tooltip);

export default function IdentityLoginTrend({
  trend,
}: {
  trend: Array<{ date: string; count: number }>;
}) {
  const options: ChartOptions<"line"> = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { mode: "index", intersect: false },
    plugins: {
      legend: { display: false },
      tooltip: { ...CHART.tooltip, displayColors: false },
    },
    scales: {
      x: {
        grid: { color: CHART.grid },
        ticks: { color: CHART.tick, font: CHART.font, maxRotation: 0, maxTicksLimit: 8 },
      },
      y: {
        beginAtZero: true,
        suggestedMax: 10,
        grid: { color: CHART.grid },
        ticks: { color: CHART.tick, font: CHART.font },
      },
    },
  };

  return (
    <Line
      data={{
        labels: trend.map((point) =>
          new Date(`${point.date}T00:00:00`).toLocaleDateString("en-US", {
            day: "numeric",
            month: "short",
          })
        ),
        datasets: [
          {
            label: "Users logged in",
            data: trend.map((point) => point.count),
            borderColor: CYBER.accent,
            backgroundColor: "rgba(0,196,255,0.08)",
            fill: true,
            tension: 0.35,
            borderWidth: 1.5,
            pointRadius: 0,
            pointHoverRadius: 3,
          },
        ],
      }}
      options={options}
    />
  );
}
