"use client";

import { BarElement, CategoryScale, Chart as ChartJS, ChartOptions, LinearScale, Tooltip } from "chart.js";
import { Bar } from "react-chartjs-2";
import { CHART } from "@/lib/colors";

ChartJS.register(CategoryScale, LinearScale, BarElement, Tooltip);

export default function IdentityStatusBar({
  status,
}: {
  status: { active: number; suspended: number; dormant: number };
}) {
  const options: ChartOptions<"bar"> = {
    indexAxis: "y",
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: { display: false },
      tooltip: CHART.tooltip,
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
    <div style={{ height: 190 }}>
      <Bar
        data={{
          labels: ["Active", "Suspended", "Dormant"],
          datasets: [
            {
              data: [status.active, status.suspended, status.dormant],
              backgroundColor: ["#00ff88", "#f7c948", "#ff6b9d"],
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
