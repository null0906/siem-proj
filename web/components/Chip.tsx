import type { ReactNode } from "react";

export function Chip({ children }: { children: ReactNode }) {
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        maxWidth: "100%",
        minHeight: 16,
        border: "0.5px solid var(--border)",
        borderRadius: 3,
        padding: "1px 5px",
        background: "light-dark(var(--mantine-color-gray-1), var(--mantine-color-dark-6))",
        color: "var(--text-secondary)",
        fontFamily: "var(--font-mono)",
        fontSize: "var(--fs-sm)",
        fontVariantNumeric: "tabular-nums",
        lineHeight: 1.3,
        overflow: "hidden",
        textOverflow: "ellipsis",
        whiteSpace: "nowrap",
      }}
    >
      {children}
    </span>
  );
}
