// App chrome
export const CYBER = {
  bgApp: "#070c12",
  bgPanel: "#0b1626",
  bgStrip: "#050a10",
  border: "rgba(0, 196, 255, 0.10)",
  borderHov: "rgba(0, 196, 255, 0.28)",
  accent: "#00c4ff",
  textPri: "#ccd6f6",
  textSec: "#4a6a8a",
  textMut: "#2d4560",
} as const;

// Severity - encodes risk (warm palette)
export const SEV = {
  CRITICAL: "#ff3b3b",
  HIGH: "#ff8c00",
  MEDIUM: "#bd34fe",
  LOW: "#00c4ff",
  INFO: "#2d4560",
} as const;

// Sources - encodes origin (cool/distinct palette, never overlaps with SEV)
export const SRC: Record<string, string> = {
  CrowdStrike: "#00ff88",
  Fortinet: "#f7c948",
  Tenable: "#ff6b9d",
  SentinelOne: "#38bdf8",
  default: "#2d4560",
};

export const srcColor = (v: string) => SRC[v] ?? SRC.default;
export const sevColor = (s: string) => SEV[s as keyof typeof SEV] ?? SEV.INFO;

// Chart.js constants (hardcoded - canvas cannot resolve CSS vars)
export const CHART = {
  grid: "rgba(0, 196, 255, 0.06)",
  tick: "#2d4560",
  font: { family: "'JetBrains Mono', monospace", size: 9 },
  border: "#080e18",
  tooltip: {
    backgroundColor: "#0b1626",
    borderColor: "rgba(0, 196, 255, 0.25)",
    borderWidth: 1,
    titleColor: "#4a6a8a",
    bodyColor: "#8892b0",
    padding: 8,
    titleFont: { family: "'JetBrains Mono', monospace", size: 10 },
    bodyFont: { family: "'JetBrains Mono', monospace", size: 10 },
  },
} as const;
