import { SEV, sevColor } from "@/lib/colors";

const TEXT_COLORS: Record<keyof typeof SEV, string> = {
  CRITICAL: "#fff",
  HIGH: "#fff",
  MEDIUM: "#fff",
  LOW: "#070c12",
  INFO: "#8892b0",
};

export function SeverityPill({ severity }: { severity: string }) {
  const key = severity.toUpperCase() as keyof typeof SEV;
  const bg = sevColor(key);
  const text = TEXT_COLORS[key] ?? TEXT_COLORS.INFO;

  return (
    <span
      style={{
        display: "inline-block",
        minWidth: 64,
        borderRadius: 2,
        padding: "2px 6px",
        background: bg,
        color: text,
        textAlign: "center",
        fontFamily: "var(--font-mono)",
        fontSize: "var(--fs-xs)",
        fontWeight: 500,
        letterSpacing: "0.07em",
        fontVariantNumeric: "tabular-nums",
        lineHeight: 1.25,
        textTransform: "uppercase",
      }}
    >
      {severity}
    </span>
  );
}
