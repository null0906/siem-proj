import type { Metadata } from "next";
import { Providers } from "@/components/Providers";
import { NavRail } from "@/components/nav/NavRail";
import { IngestPipelineStrip } from "@/components/IngestPipelineStrip";
import "@fontsource-variable/inter";
import "@fontsource/ibm-plex-mono/400.css";
import "@fontsource/ibm-plex-mono/500.css";
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
