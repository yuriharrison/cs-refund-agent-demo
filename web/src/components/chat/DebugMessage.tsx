import { useState } from "react";
import type { DebugMessage as DebugMessageType } from "../../hooks/useChat";

interface DebugMessageProps {
  message: DebugMessageType;
}

function ChevronIcon({ open }: { open: boolean }) {
  return (
    <svg
      className={`h-4 w-4 shrink-0 text-slate-500 transition-transform ${open ? "rotate-90" : ""}`}
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth={2}
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 5l7 7-7 7" />
    </svg>
  );
}

function ToolIcon() {
  return (
    <svg className="h-3.5 w-3.5 text-indigo-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
      <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
    </svg>
  );
}

function CheckCircleIcon() {
  return (
    <svg className="h-3.5 w-3.5 text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
    </svg>
  );
}

const TOOL_LABELS: Record<string, string> = {
  lookup_customer_orders: "Looking up orders",
  get_refund_policy: "Checking refund policy",
  issue_refund: "Processing refund",
  escalate_to_human: "Escalating to human agent",
};

function friendlyToolName(tool: string): string {
  return TOOL_LABELS[tool] ?? tool.replace(/_/g, " ");
}

function summarizeArgs(tool: string, args: Record<string, unknown>): string {
  const parts: string[] = [];
  if (tool === "lookup_customer_orders" && args.limit) {
    parts.push(`last ${args.limit} orders`);
  }
  if (tool === "get_refund_policy") {
    if (args.product_type) parts.push(String(args.product_type));
    if (args.condition) parts.push(String(args.condition).replace(/_/g, " "));
  }
  if (tool === "issue_refund") {
    if (args.refund_type) parts.push(String(args.refund_type));
    if (args.amount) parts.push(`$${Number(args.amount).toFixed(2)}`);
  }
  if (tool === "escalate_to_human" && args.reason) {
    const reason = String(args.reason);
    parts.push(reason.length > 50 ? reason.slice(0, 50) + "…" : reason);
  }
  return parts.length > 0 ? parts.join(" · ") : "";
}

function summarizeResult(tool: string, result: unknown): string {
  if (result == null) return "done";
  if (typeof result === "string") {
    try {
      const parsed = JSON.parse(result);
      return summarizeResult(tool, parsed);
    } catch {
      return result.length > 60 ? result.slice(0, 60) + "…" : result;
    }
  }
  if (typeof result !== "object") return String(result);

  const obj = result as Record<string, unknown>;

  if (obj.error) return `error: ${obj.error}`;
  if (tool === "get_refund_policy" && obj.action) {
    const action = String(obj.action).replace(/_/g, " ");
    if (obj.partial_percent) return `${action} (${obj.partial_percent}%)`;
    return action;
  }
  if (tool === "issue_refund" && obj.refund_id) {
    return `refund #${obj.refund_id} issued`;
  }
  if (tool === "escalate_to_human") return "escalated";
  if (Array.isArray(result)) return `${result.length} items`;

  return "done";
}

export function DebugMessage({ message }: DebugMessageProps) {
  const [open, setOpen] = useState(false);

  const isStart = message.variant === "start";
  const toolLabel = friendlyToolName(message.tool);

  const summary = isStart
    ? summarizeArgs(message.tool, message.arguments)
    : summarizeResult(message.tool, message.result);

  const body = isStart
    ? JSON.stringify(message.arguments, null, 2)
    : JSON.stringify(
        { result: message.result, duration_ms: message.durationMs },
        null,
        2,
      );

  return (
    <div className="animate-fade-in-up px-4" data-testid={`debug-${message.variant}`}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        data-testid="debug-toggle"
        className="flex w-full items-center gap-2 rounded-lg bg-slate-800/50 px-3 py-2 text-left transition-colors hover:bg-slate-800/70"
      >
        <ChevronIcon open={open} />
        {isStart ? <ToolIcon /> : <CheckCircleIcon />}
        <span className="text-xs text-slate-300">{toolLabel}</span>
        {summary && (
          <span className="truncate text-xs text-slate-500">{summary}</span>
        )}
        {!isStart && (
          <span className="ml-auto shrink-0 font-mono text-xs text-slate-500">
            {message.durationMs}ms
          </span>
        )}
      </button>
      {open && (
        <pre className="mt-1 overflow-x-auto rounded-lg bg-slate-900/60 p-3 font-mono text-xs text-slate-400">
          {body}
        </pre>
      )}
    </div>
  );
}
