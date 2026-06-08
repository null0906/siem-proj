"use client";

import { useMemo, useState } from "react";
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { useQuery } from "@tanstack/react-query";
import { Drawer, Pagination, ScrollArea, Select, TextInput } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { apiClient, Finding, FindingsResponse } from "@/lib/api";
import { SeverityPill } from "@/components/SeverityPill";

const PAGE_SIZE = 50;
const columnHelper = createColumnHelper<Finding>();

const sourceOptions = [
  { value: "firewall", label: "Firewall" },
  { value: "edr", label: "EDR" },
  { value: "scanner", label: "Scanner" },
  { value: "dlp", label: "DLP" },
  { value: "other", label: "Other" },
];

function formatDate(value: string) {
  return new Date(value).toLocaleDateString("en-GB", {
    day: "2-digit",
    month: "short",
    year: "2-digit",
  });
}

function statusColor(status: string) {
  if (status === "open") return { bg: "rgba(220, 38, 38, 0.12)", text: "#fca5a5" };
  if (status === "resolved") return { bg: "rgba(37, 99, 235, 0.12)", text: "#93c5fd" };
  return { bg: "rgba(107, 114, 128, 0.14)", text: "#d1d5db" };
}

function StatusBadge({ status }: { status: string }) {
  const color = statusColor(status);
  return (
    <span
      style={{
        display: "inline-flex",
        alignItems: "center",
        borderRadius: 2,
        padding: "2px 5px",
        background: color.bg,
        color: color.text,
        fontFamily: "var(--font-mono)",
        fontSize: 9,
        letterSpacing: "0.06em",
        textTransform: "uppercase",
      }}
    >
      {status}
    </span>
  );
}

export default function FindingsPage() {
  const [severity, setSeverity] = useState<string | null>(null);
  const [source, setSource] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<Finding | null>(null);
  const [opened, { open, close }] = useDisclosure(false);

  const params = useMemo(
    () => ({
      severity: severity ?? undefined,
      source_tool: source ?? undefined,
      status: status ?? undefined,
      search,
      page,
      limit: PAGE_SIZE,
    }),
    [severity, source, status, search, page]
  );

  const { data, isLoading } = useQuery<FindingsResponse>({
    queryKey: ["findings", params],
    queryFn: () => apiClient.getFindings(params),
    placeholderData: (previous) => previous,
  });

  const findings = data?.findings ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const columns = useMemo(
    () => [
      columnHelper.accessor("severity", {
        header: "Severity",
        size: 90,
        cell: (info) => <SeverityPill severity={info.getValue()} />,
      }),
      columnHelper.accessor("title", {
        header: "Finding",
        size: 420,
        cell: (info) => (
          <span
            style={{
              display: "block",
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
              fontSize: 13,
              fontWeight: 500,
              color: "var(--text-primary)",
            }}
          >
            {info.getValue()}
          </span>
        ),
      }),
      columnHelper.accessor("source_vendor", {
        header: "Vendor",
        size: 110,
        cell: (info) => (
          <span style={{ color: "var(--text-secondary)", fontSize: 12 }}>
            {info.getValue() || "unknown"}
          </span>
        ),
      }),
      columnHelper.accessor("affected_asset", {
        header: "Asset",
        size: 140,
        cell: (info) => (
          <span
            style={{
              display: "block",
              overflow: "hidden",
              textOverflow: "ellipsis",
              whiteSpace: "nowrap",
              fontFamily: "var(--font-mono)",
              fontSize: 11,
              color: "var(--text-secondary)",
            }}
          >
            {info.getValue() || "unknown"}
          </span>
        ),
      }),
      columnHelper.accessor("status", {
        header: "Status",
        size: 80,
        cell: (info) => <StatusBadge status={info.getValue()} />,
      }),
      columnHelper.accessor("first_seen", {
        header: "First seen",
        size: 110,
        cell: (info) => (
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: 10,
              color: "var(--text-tertiary)",
            }}
          >
            {formatDate(info.getValue())}
          </span>
        ),
      }),
    ],
    []
  );

  const table = useReactTable({
    data: findings,
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
  });

  const resetPage = (fn: () => void) => {
    fn();
    setPage(1);
  };

  return (
    <div style={{ flex: 1, display: "flex", flexDirection: "column", minHeight: 0 }}>
      <header
        style={{
          padding: "18px 20px 12px",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          gap: 16,
          borderBottom: "0.5px solid var(--border)",
        }}
      >
        <h1 style={{ margin: 0, fontSize: 22, fontWeight: 500, lineHeight: 1.2 }}>
          Findings
        </h1>
        <span
          style={{
            fontFamily: "var(--font-mono)",
            fontSize: 12,
            color: "var(--text-tertiary)",
          }}
        >
          {isLoading ? "..." : `${total} total`}
        </span>
      </header>

      <div
        style={{
          display: "flex",
          gap: 8,
          alignItems: "center",
          padding: "10px 20px",
          borderBottom: "0.5px solid var(--border)",
          background: "var(--surface)",
        }}
      >
        <Select
          placeholder="Severity"
          value={severity}
          onChange={(value) => resetPage(() => setSeverity(value))}
          data={["critical", "high", "medium", "low", "info"]}
          size="xs"
          clearable
          style={{ width: 126 }}
        />
        <Select
          placeholder="Source"
          value={source}
          onChange={(value) => resetPage(() => setSource(value))}
          data={sourceOptions}
          size="xs"
          clearable
          style={{ width: 126 }}
        />
        <Select
          placeholder="Status"
          value={status}
          onChange={(value) => resetPage(() => setStatus(value))}
          data={["open", "resolved", "suppressed"]}
          size="xs"
          clearable
          style={{ width: 132 }}
        />
        <TextInput
          placeholder="Search findings..."
          value={search}
          onChange={(event) => resetPage(() => setSearch(event.currentTarget.value))}
          size="xs"
          style={{ width: 260 }}
        />
      </div>

      <div style={{ flex: 1, minHeight: 0, overflow: "auto" }}>
        <table style={{ width: "100%", minWidth: 860, borderCollapse: "collapse", tableLayout: "fixed" }}>
          <thead>
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <th
                    key={header.id}
                    style={{
                      width: header.getSize(),
                      height: 32,
                      padding: "0 12px",
                      textAlign: "left",
                      borderBottom: "1px solid var(--border)",
                      background: "var(--surface)",
                      color: "var(--text-tertiary)",
                      fontFamily: "var(--font-mono)",
                      fontSize: 9,
                      fontWeight: 500,
                      letterSpacing: "0.1em",
                      textTransform: "uppercase",
                    }}
                  >
                    {flexRender(header.column.columnDef.header, header.getContext())}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {isLoading &&
              Array.from({ length: 10 }).map((_, index) => (
                <tr key={index} style={{ height: 44, borderBottom: "0.5px solid var(--border)" }}>
                  <td colSpan={columns.length} style={{ padding: "0 12px" }}>
                    <div style={{ width: "100%", height: 12, background: "var(--surface-raised)" }} />
                  </td>
                </tr>
              ))}

            {!isLoading &&
              table.getRowModel().rows.map((row) => (
                <tr
                  key={row.id}
                  onClick={() => {
                    setSelected(row.original);
                    open();
                  }}
                  style={{
                    height: 44,
                    borderBottom: "0.5px solid var(--border)",
                    cursor: "pointer",
                  }}
                  onMouseEnter={(event) => {
                    event.currentTarget.style.background = "var(--surface)";
                  }}
                  onMouseLeave={(event) => {
                    event.currentTarget.style.background = "transparent";
                  }}
                >
                  {row.getVisibleCells().map((cell) => (
                    <td
                      key={cell.id}
                      style={{
                        width: cell.column.getSize(),
                        padding: "0 12px",
                        overflow: "hidden",
                      }}
                    >
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))}
          </tbody>
        </table>
      </div>

      <footer
        style={{
          minHeight: 44,
          padding: "7px 20px",
          borderTop: "0.5px solid var(--border)",
          display: "flex",
          alignItems: "center",
          justifyContent: "flex-end",
        }}
      >
        <Pagination
          total={totalPages}
          value={page}
          onChange={setPage}
          size="xs"
          withEdges
        />
      </footer>

      <Drawer
        opened={opened}
        onClose={close}
        position="right"
        size={480}
        title={selected?.title}
        styles={{
          content: { background: "var(--surface)", color: "var(--text-primary)" },
          header: { background: "var(--surface)", borderBottom: "0.5px solid var(--border)" },
          title: { fontSize: 18, fontWeight: 500, lineHeight: 1.3 },
          body: { padding: 0 },
        }}
      >
        {selected && <FindingDetail finding={selected} />}
      </Drawer>
    </div>
  );
}

function FindingDetail({ finding }: { finding: Finding }) {
  const metadata = [
    ["Vendor", finding.source_vendor || "unknown"],
    ["Asset", finding.affected_asset || "unknown"],
    ["First seen", new Date(finding.first_seen).toLocaleString()],
    ["Last seen", new Date(finding.last_seen).toLocaleString()],
    ["Source file", finding.source_file_id ?? "unknown"],
  ];

  return (
    <ScrollArea h="calc(100vh - 64px)">
      <div style={{ padding: 18 }}>
        <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 16 }}>
          <SeverityPill severity={finding.severity} />
          <StatusBadge status={finding.status} />
        </div>

        <div
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr",
            border: "0.5px solid var(--border)",
            marginBottom: 18,
          }}
        >
          {metadata.map(([label, value]) => (
            <div
              key={label}
              style={{
                minWidth: 0,
                padding: "9px 10px",
                borderRight: "0.5px solid var(--border)",
                borderBottom: "0.5px solid var(--border)",
              }}
            >
              <div
                style={{
                  marginBottom: 4,
                  color: "var(--text-tertiary)",
                  fontFamily: "var(--font-mono)",
                  fontSize: 9,
                  letterSpacing: "0.08em",
                  textTransform: "uppercase",
                }}
              >
                {label}
              </div>
              <div
                style={{
                  overflowWrap: "anywhere",
                  color: "var(--text-secondary)",
                  fontFamily: "var(--font-mono)",
                  fontSize: 11,
                }}
              >
                {value}
              </div>
            </div>
          ))}
        </div>

        <div
          style={{
            marginBottom: 8,
            color: "var(--text-tertiary)",
            fontFamily: "var(--font-mono)",
            fontSize: 9,
            letterSpacing: "0.1em",
            textTransform: "uppercase",
          }}
        >
          Raw payload
        </div>
        <pre
          style={{
            margin: 0,
            padding: 12,
            overflow: "auto",
            background: "var(--bg)",
            border: "0.5px solid var(--border)",
            color: "var(--text-secondary)",
            fontFamily: "var(--font-mono)",
            fontSize: 11,
            lineHeight: 1.55,
            whiteSpace: "pre-wrap",
          }}
        >
          {JSON.stringify(finding.raw_payload ?? {}, null, 2)}
        </pre>
      </div>
    </ScrollArea>
  );
}
