"use client";

import { useMemo } from "react";
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
import { CHART, SEV } from "@/lib/colors";

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Filler, Tooltip);

export interface TrendPoint {
  date: string;
  critical: number;
  high: number;
  medium: number;
}

function formatTrendDate(value: string) {
  const date = new Date(`${value}T00:00:00`);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleDateString("en-US", { day: "numeric", month: "short" });
}

export default function TrendChart({ trend }: { trend: TrendPoint[] }) {
  const gridColor = CHART.grid;
  const tickColor = CHART.tick;

  const data = useMemo(
    () => ({
      labels: trend.map((point) => formatTrendDate(point.date)),
      datasets: [
        {
          label: "Critical",
          data: trend.map((point) => point.critical),
          borderColor: SEV.CRITICAL,
          backgroundColor: "rgba(255,59,59,0.08)",
          fill: true,
          spanGaps: false,
          tension: 0.4,
          borderWidth: 1.5,
          pointRadius: 0,
          pointHoverRadius: 3,
        },
        {
          label: "High",
          data: trend.map((point) => point.high),
          borderColor: SEV.HIGH,
          backgroundColor: "rgba(255,140,0,0.06)",
          fill: true,
          spanGaps: false,
          tension: 0.4,
          borderWidth: 1.5,
          pointRadius: 0,
          pointHoverRadius: 3,
        },
        {
          label: "Medium",
          data: trend.map((point) => point.medium),
          borderColor: SEV.MEDIUM,
          backgroundColor: "rgba(189,52,254,0.05)",
          fill: true,
          spanGaps: false,
          tension: 0.4,
          borderWidth: 1.5,
          pointRadius: 0,
          pointHoverRadius: 3,
        },
      ],
    }),
    [trend]
  );

  const options: ChartOptions<"line"> = {
    responsive: true,
    maintainAspectRatio: false,
    interaction: { mode: "index", intersect: false },
    plugins: {
      legend: { display: false },
      tooltip: {
        ...CHART.tooltip,
        displayColors: true,
      },
    },
    scales: {
      x: {
        grid: { color: gridColor },
        ticks: {
          color: tickColor,
          font: CHART.font,
          maxRotation: 0,
          autoSkip: true,
          maxTicksLimit: 7,
        },
      },
      y: {
        grid: { color: gridColor },
        ticks: { color: tickColor, font: CHART.font },
        beginAtZero: true,
        suggestedMax: 10,
      },
    },
  };

  return <Line data={data} options={options} />;
}
