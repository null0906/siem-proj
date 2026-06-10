"use client";

import { useEffect, useMemo, useState } from "react";
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button, Drawer, Pagination, ScrollArea, Select, Textarea, TextInput } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { apiClient, Finding, FindingsResponse } from "@/lib/api";
import { SeverityPill } from "@/components/SeverityPill";

const PAGE_SIZE = 50;
const CURRENT_USER = "security@seccomply.demo";
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
        fontFamily: "var(--font-sans)",
        fontSize: "var(--fs-xs)",
      }}
    >
      {status}
    </span>
  );
}

function SLABadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    breached: "#ff3b3b",
    due_soon: "#f7c948",
    on_track: "#00ff88",
    closed: "#54657d",
  };
  return <span className={`finding-sla finding-sla-${status}`} style={{ color: colors[status] ?? colors.closed }}>{status.replace("_", " ")}</span>;
}

export default function FindingsPage() {
  const [severity, setSeverity] = useState<string | null>(null);
  const [source, setSource] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [myQueue, setMyQueue] = useState(false);
  const [overdue, setOverdue] = useState(false);
  const [page, setPage] = useState(1);
  const [selected, setSelected] = useState<Finding | null>(null);
  const [opened, { open, close }] = useDisclosure(false);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    setSeverity(params.get("severity"));
    setSource(params.get("source_tool"));
    setStatus(params.get("status"));
    setSearch(params.get("search") ?? "");
    setPage(1);
  }, []);

  const params = useMemo(
    () => ({
      severity: severity ?? undefined,
      source_tool: source ?? undefined,
      status: status ?? undefined,
      search,
      assignee: myQueue ? CURRENT_USER : undefined,
      overdue: overdue ? "true" : undefined,
      page,
      limit: PAGE_SIZE,
    }),
    [severity, source, status, search, myQueue, overdue, page]
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
              fontSize: "var(--fs-base)",
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
          <span style={{ color: "var(--text-secondary)", fontSize: "var(--fs-sm)", fontFamily: "var(--font-sans)" }}>
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
              fontSize: "var(--fs-sm)",
              fontVariantNumeric: "tabular-nums",
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
      columnHelper.accessor("sla_status", {
        header: "SLA",
        size: 100,
        cell: (info) => <SLABadge status={info.getValue()} />,
      }),
      columnHelper.accessor("assignee", {
        header: "Owner",
        size: 150,
        cell: (info) => <span className="finding-owner-cell">{info.getValue() || "Unassigned"}</span>,
      }),
      columnHelper.accessor("first_seen", {
        header: "First seen",
        size: 110,
        cell: (info) => (
          <span
            style={{
              fontFamily: "var(--font-mono)",
              fontSize: "var(--fs-xs)",
              fontVariantNumeric: "tabular-nums",
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
        <h1 style={{ margin: 0, fontFamily: "var(--font-sans)", fontSize: "var(--fs-h1)", fontWeight: 600, lineHeight: 1.2, color: "var(--text-title)" }}>
          Findings
        </h1>
        <span
          style={{
            fontFamily: "var(--font-sans)",
            fontSize: "var(--fs-body)",
            fontVariantNumeric: "tabular-nums",
            color: "var(--text-secondary)",
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
          data={["open", "in_progress", "resolved", "risk_accepted", "suppressed"]}
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
        <button type="button" className={myQueue ? "finding-filter-toggle active" : "finding-filter-toggle"} onClick={() => resetPage(() => setMyQueue(!myQueue))}>
          My queue
        </button>
        <button type="button" className={overdue ? "finding-filter-toggle active overdue" : "finding-filter-toggle"} onClick={() => resetPage(() => setOverdue(!overdue))}>
          Overdue
        </button>
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
                      color: "var(--text-secondary)",
                      fontFamily: "var(--font-sans)",
                      fontSize: "var(--fs-label)",
                      fontWeight: 500,
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
          title: { fontFamily: "var(--font-sans)", fontSize: "var(--fs-lg)", fontWeight: 600, lineHeight: 1.3, color: "var(--text-title)" },
          body: { padding: 0 },
        }}
      >
        {selected && <FindingDetail finding={selected} onUpdated={setSelected} />}
      </Drawer>
    </div>
  );
}

function FindingDetail({ finding, onUpdated }: { finding: Finding; onUpdated: (finding: Finding) => void }) {
  const queryClient = useQueryClient();
  const { data: detail } = useQuery({
    queryKey: ["finding", finding.id],
    queryFn: () => apiClient.getFinding(finding.id),
  });
  const current = detail ?? finding;
  const [assignee, setAssignee] = useState(current.assignee ?? "");
  const [status, setStatus] = useState(current.status);
  const [dueDate, setDueDate] = useState(current.due_date ? current.due_date.slice(0, 16) : "");
  const [note, setNote] = useState("");
  useEffect(() => {
    setAssignee(current.assignee ?? "");
    setStatus(current.status);
    setDueDate(current.due_date ? current.due_date.slice(0, 16) : "");
  }, [current.assignee, current.status, current.due_date]);
  const update = useMutation({
    mutationFn: () => apiClient.updateFindingWorkflow(current.id, {
      actor: CURRENT_USER,
      assignee,
      status,
      ...(dueDate ? { due_date: new Date(dueDate).toISOString() } : {}),
      ...(note.trim() ? { note: note.trim() } : {}),
    }),
    onSuccess: (updated) => {
      onUpdated(updated);
      setNote("");
      queryClient.setQueryData(["finding", finding.id], updated);
      queryClient.invalidateQueries({ queryKey: ["findings"] });
    },
  });
  const metadata = [
    { label: "Vendor", value: current.source_vendor || "unknown", mono: false },
    { label: "Asset", value: current.affected_asset || "unknown", mono: true },
    { label: "First seen", value: new Date(current.first_seen).toLocaleString(), mono: true },
    { label: "Last seen", value: new Date(current.last_seen).toLocaleString(), mono: true },
    { label: "SLA due", value: new Date(current.sla_due_at).toLocaleString(), mono: true },
    { label: "Source file", value: current.source_file_id ?? "unknown", mono: true },
  ];

  return (
    <ScrollArea h="calc(100vh - 64px)">
      <div style={{ padding: 18 }}>
        <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 16 }}>
          <SeverityPill severity={current.severity} />
          <StatusBadge status={current.status} />
          <SLABadge status={current.sla_status} />
        </div>

        <section className="finding-workflow">
          <div className="finding-detail-heading">Remediation workflow</div>
          <div className="finding-workflow-grid">
            <TextInput label="Owner" value={assignee} onChange={(event) => setAssignee(event.currentTarget.value)} placeholder="owner@company.com" size="xs" />
            <Select label="Status" value={status} onChange={(value) => setStatus(value ?? "open")} data={["open", "in_progress", "resolved", "risk_accepted"]} size="xs" />
            <TextInput label="Due date" type="datetime-local" value={dueDate} onChange={(event) => setDueDate(event.currentTarget.value)} size="xs" />
          </div>
          <Textarea label="Add note" value={note} onChange={(event) => setNote(event.currentTarget.value)} minRows={2} size="xs" />
          <Button className="finding-workflow-save" size="xs" loading={update.isPending} onClick={() => update.mutate()}>Save workflow</Button>
          {update.isError ? <div className="finding-workflow-error">Could not save workflow changes.</div> : null}
        </section>

        <section className="finding-audit">
          <div className="finding-detail-heading">Audit trail</div>
          {(current.events ?? []).map((event) => (
            <div className="finding-audit-event" key={event.id}>
              <span>{event.event_type.replaceAll("_", " ")}</span>
              <strong>{event.note || event.to_value || "Updated"}</strong>
              <small>{event.actor} · {new Date(event.created_at).toLocaleString()}</small>
            </div>
          ))}
          {(current.events ?? []).length === 0 ? <div className="finding-audit-empty">No workflow changes yet.</div> : null}
        </section>

        <div
          style={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr",
            border: "0.5px solid var(--border)",
            marginBottom: 18,
          }}
        >
          {metadata.map(({ label, value, mono }) => (
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
                  color: "var(--text-secondary)",
                  fontFamily: "var(--font-sans)",
                  fontSize: "var(--fs-label)",
                }}
              >
                {label}
              </div>
              <div
                style={{
                  overflowWrap: "anywhere",
                  color: "var(--text-secondary)",
                  fontFamily: mono ? "var(--font-mono)" : "var(--font-sans)",
                  fontSize: "var(--fs-sm)",
                  fontVariantNumeric: "tabular-nums",
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
            color: "var(--text-title)",
            fontFamily: "var(--font-sans)",
            fontSize: "var(--fs-title)",
            fontWeight: 600,
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
            fontSize: "var(--fs-sm)",
            lineHeight: 1.55,
            whiteSpace: "pre-wrap",
          }}
        >
          {JSON.stringify(current.raw_payload ?? {}, null, 2)}
        </pre>
      </div>
    </ScrollArea>
  );
}
