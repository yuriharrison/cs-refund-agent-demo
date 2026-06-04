import type { SystemMessage as SystemMessageType } from "../../hooks/useChat";

interface SystemMessageProps {
  message: SystemMessageType;
}

function CheckIcon() {
  return (
    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
    </svg>
  );
}

function HandoffIcon() {
  return (
    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M7 11.5V14m0-2.5v-6a1.5 1.5 0 113 0m-3 6a1.5 1.5 0 00-3 0v2a7.5 7.5 0 0015 0v-5a1.5 1.5 0 00-3 0m-6-3V11m0-5.5v-1a1.5 1.5 0 013 0v1m0 0V11m0-5.5a1.5 1.5 0 013 0v3m0 0V11"
      />
    </svg>
  );
}

function WarningIcon() {
  return (
    <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
      />
    </svg>
  );
}

export function SystemMessage({ message }: SystemMessageProps) {
  if (message.variant === "refund") {
    return (
      <div className="animate-fade-in-up px-4" data-testid="system-refund">
        <div className="glass mx-auto flex max-w-md items-start gap-3 rounded-xl border-emerald-500/30 bg-emerald-500/10 p-4">
          <div className="rounded-full bg-emerald-500/20 p-2 text-emerald-400">
            <CheckIcon />
          </div>
          <div className="min-w-0 flex-1">
            <p className="font-medium text-emerald-300">Refund Confirmed</p>
            <p className="mt-1 text-sm text-emerald-200/80">
              ${message.amount.toFixed(2)} {message.refundType} refund issued
            </p>
            <p className="mt-1 font-mono text-xs text-emerald-400/60">
              {message.refundId} · {message.orderItemId}
            </p>
          </div>
        </div>
      </div>
    );
  }

  if (message.variant === "escalation") {
    return (
      <div className="animate-fade-in-up px-4" data-testid="system-escalation">
        <div className="glass mx-auto flex max-w-md items-start gap-3 rounded-xl border-amber-500/30 bg-amber-500/10 p-4">
          <div className="rounded-full bg-amber-500/20 p-2 text-amber-400">
            <HandoffIcon />
          </div>
          <div className="min-w-0 flex-1">
            <p className="font-medium text-amber-300">Escalated to Human Agent</p>
            <p className="mt-1 text-sm text-amber-200/80">{message.reason}</p>
            <p className="mt-1 text-xs text-amber-400/70">
              Connecting you with {message.humanAgentName}…
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="animate-fade-in-up px-4" data-testid="system-error">
      <div className="glass mx-auto flex max-w-md items-start gap-3 rounded-xl border-red-500/30 bg-red-500/10 p-4">
        <div className="rounded-full bg-red-500/20 p-2 text-red-400">
          <WarningIcon />
        </div>
        <div className="min-w-0 flex-1">
          <p className="font-medium text-red-300">Error</p>
          <p className="mt-1 text-sm text-red-200/80">{message.message}</p>
        </div>
      </div>
    </div>
  );
}
