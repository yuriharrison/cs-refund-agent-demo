import type { TokenStats } from "../../hooks/useChat";

interface TokenCounterProps {
  stats: TokenStats;
}

function formatCost(stats: TokenStats): string {
  const cost =
    (stats.promptTokens / 1_000_000) * 0.14 +
    (stats.completionTokens / 1_000_000) * 0.28;
  return cost < 0.0001 ? "$0.0000" : `$${cost.toFixed(4)}`;
}

export function TokenCounter({ stats }: TokenCounterProps) {
  return (
    <p className="text-right text-xs text-slate-500" data-testid="token-counter">
      Tokens: {stats.total.toLocaleString()} · {formatCost(stats)}
    </p>
  );
}
