"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Tooltip } from "@mantine/core";
import {
  IconActivity,
  IconTimelineEvent,
  IconChecklist,
  IconAlertTriangle,
  IconDatabaseImport,
  IconKey,
  IconPresentation,
  IconShieldCheck,
  IconServer,
  IconSettings,
} from "@tabler/icons-react";

const navItems = [
  { href: "/", label: "Overview", short: "OVE", Icon: IconActivity },
  { href: "/findings", label: "Findings", short: "FIN", Icon: IconAlertTriangle },
  { href: "/identity", label: "Identity", short: "IAM", Icon: IconKey },
  { href: "/assets", label: "Assets", short: "AST", Icon: IconServer },
  { href: "/actions", label: "Actions", short: "ACT", Icon: IconChecklist },
  { href: "/executive", label: "Executive", short: "EXE", Icon: IconPresentation },
  { href: "/compliance", label: "Compliance", short: "CMP", Icon: IconShieldCheck },
  { href: "/activity", label: "Activity", short: "LOG", Icon: IconTimelineEvent },
  { href: "/sources", label: "Sources", short: "SOU", Icon: IconDatabaseImport },
  { href: "/system", label: "System", short: "SYS", Icon: IconSettings },
];

export function NavRail() {
  const pathname = usePathname();

  return (
    <nav
      style={{
        width: 40,
        minHeight: "100vh",
        background: "#050a10",
        borderRight: "1px solid rgba(0, 196, 255, 0.10)",
        display: "flex",
        flexDirection: "column",
        alignItems: "stretch",
        paddingTop: 8,
        flexShrink: 0,
        position: "sticky",
        top: 0,
        zIndex: 30,
      }}
    >
      {navItems.map(({ href, label, short, Icon }) => {
        const isActive = href === "/" ? pathname === "/" : pathname.startsWith(href);

        return (
          <Tooltip key={href} label={label} position="right" withArrow>
            <Link
              href={href}
              aria-label={label}
              style={{
                width: 40,
                height: 46,
                display: "flex",
                flexDirection: "column",
                alignItems: "center",
                justifyContent: "center",
                gap: 2,
                borderLeft: isActive ? "2px solid #00c4ff" : "2px solid transparent",
                color: isActive ? "#00c4ff" : "#2d4560",
                textDecoration: "none",
              }}
              onMouseEnter={(event) => {
                if (!isActive) event.currentTarget.style.color = "#4a6a8a";
              }}
              onMouseLeave={(event) => {
                if (!isActive) event.currentTarget.style.color = "#2d4560";
              }}
            >
              <Icon size={15} stroke={1.6} />
              <span
                style={{
                  fontFamily: "var(--font-sans)",
                  fontSize: "var(--fs-caption)",
                  fontWeight: 500,
                  letterSpacing: "0.04em",
                  lineHeight: 1,
                }}
              >
                {short}
              </span>
            </Link>
          </Tooltip>
        );
      })}
    </nav>
  );
}
