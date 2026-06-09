import { SEVERITY_COLORS, Severity } from "@/lib/theme";

interface Props {
  severity: string;
  size?: "sm" | "md";
}

export function SeverityBadge({ severity, size = "sm" }: Props) {
  const color = SEVERITY_COLORS[severity as Severity] ?? "#6b7280";
  const isLg = size === "md";

  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 4,
        fontFamily: "var(--font-mono)",
        fontSize: isLg ? "var(--fs-sm)" : "var(--fs-xs)",
        fontWeight: 600,
        letterSpacing: "0.08em",
        color,
        textTransform: "uppercase",
        whiteSpace: "nowrap",
      }}
    >
      <span
        style={{
          width: isLg ? 6 : 5,
          height: isLg ? 6 : 5,
          borderRadius: "50%",
          background: color,
          flexShrink: 0,
          opacity: 0.9,
        }}
      />
      {severity}
    </span>
  );
}
