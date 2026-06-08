import type { Metadata } from "next";
import { Providers } from "@/components/Providers";
import { NavRail } from "@/components/nav/NavRail";
import { IngestPipelineStrip } from "@/components/IngestPipelineStrip";
import "@mantine/core/styles.css";
import "@mantine/notifications/styles.css";
import "./globals.css";

export const metadata: Metadata = {
  title: "SecComply",
  description: "Unified Security Operations Dashboard",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en" data-mantine-color-scheme="dark">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&display=swap"
          rel="stylesheet"
        />
      </head>
      <body>
        <Providers>
          <div style={{ display: "flex", minHeight: "100vh" }}>
            <NavRail />
            <main
              style={{
                flex: 1,
                minWidth: 0,
                minHeight: "100vh",
                display: "flex",
                flexDirection: "column",
              }}
            >
              <IngestPipelineStrip />
              {children}
            </main>
          </div>
        </Providers>
      </body>
    </html>
  );
}
