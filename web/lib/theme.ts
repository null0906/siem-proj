import { createTheme } from "@mantine/core";

export const theme = createTheme({
  primaryColor: "cyan",
  fontFamily: "var(--font-sans)",
  fontFamilyMonospace: "var(--font-mono)",
  defaultRadius: 0,
  black: "#070c12",
  colors: {
    dark: [
      "#ccd6f6",
      "#8892b0",
      "#4a6a8a",
      "#2d4560",
      "#1e3048",
      "#0f1e30",
      "#0b1626",
      "#080e18",
      "#050a10",
      "#070c12",
    ],
  },
  other: {
    cyber: {
      accent: "#00c4ff",
      criticalRed: "#ff3b3b",
      resolvedGreen: "#00ff88",
    },
  },

  components: {
    Button: {
      defaultProps: {
        radius: 0,
      },
    },
    Badge: {
      defaultProps: {
        radius: 0,
      },
    },
    Paper: {
      defaultProps: {
        radius: 0,
      },
    },
    Card: {
      defaultProps: {
        radius: 0,
      },
    },
    TextInput: {
      defaultProps: {
        radius: 0,
      },
    },
    Select: {
      defaultProps: {
        radius: 0,
      },
    },
    Drawer: {
      defaultProps: {
        radius: 0,
      },
    },
  },
});

export const SEVERITY_COLORS = {
  critical: "#ff3b3b",
  high: "#ff8c00",
  medium: "#bd34fe",
  low: "#00c4ff",
  info: "#2d4560",
} as const;

export const SEVERITY_BG = {
  critical: "rgba(229,72,77,0.08)",
  high: "rgba(236,148,85,0.08)",
  medium: "rgba(245,215,94,0.08)",
  low: "rgba(134,182,200,0.08)",
  info: "rgba(107,114,128,0.08)",
} as const;

export type Severity = keyof typeof SEVERITY_COLORS;

export const SOURCE_TOOL_LABELS: Record<string, string> = {
  firewall: "Firewall",
  edr: "EDR",
  scanner: "Scanner",
  dlp: "DLP",
  other: "Other",
};
