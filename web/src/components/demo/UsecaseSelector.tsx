import { useEffect, useRef, useState } from "react";
import type { Usecase } from "../../hooks/useUsecases";

interface UsecaseSelectorProps {
  usecases: Usecase[];
  onSelect: (usecase: Usecase) => void;
  onClose: () => void;
}

const CATEGORY_ORDER = [
  "Refund",
  "Escalation",
  "Subscription",
  "Other",
] as const;

function categorize(name: string): string {
  if (name.startsWith("Full Refund") || name.startsWith("Partial Refund") || name.startsWith("Refund"))
    return "Refund";
  if (name.startsWith("Escalation")) return "Escalation";
  if (name.startsWith("Subscription")) return "Subscription";
  return "Other";
}

export function UsecaseSelector({ usecases, onSelect, onClose }: UsecaseSelectorProps) {
  const [query, setQuery] = useState("");
  const [selectedIdx, setSelectedIdx] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  const filtered = usecases.filter((uc) => {
    const q = query.toLowerCase();
    return (
      uc.name.toLowerCase().includes(q) ||
      uc.description.toLowerCase().includes(q) ||
      uc.id.toLowerCase().includes(q)
    );
  });

  const grouped = CATEGORY_ORDER.map((cat) => ({
    category: cat,
    items: filtered.filter((uc) => categorize(uc.name) === cat),
  })).filter((g) => g.items.length > 0);

  const flatItems = grouped.flatMap((g) => g.items);

  useEffect(() => {
    setSelectedIdx(0);
  }, [query]);

  useEffect(() => {
    const el = listRef.current?.querySelector(`[data-idx="${selectedIdx}"]`);
    el?.scrollIntoView({ block: "nearest" });
  }, [selectedIdx]);

  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        onClose();
        return;
      }
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setSelectedIdx((i) => Math.min(i + 1, flatItems.length - 1));
        return;
      }
      if (e.key === "ArrowUp") {
        e.preventDefault();
        setSelectedIdx((i) => Math.max(i - 1, 0));
        return;
      }
      if (e.key === "Enter" && flatItems.length > 0) {
        e.preventDefault();
        onSelect(flatItems[selectedIdx]);
      }
    };

    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, [flatItems, selectedIdx, onSelect, onClose]);

  let flatIdx = 0;

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center bg-black/60 pt-[15vh] backdrop-blur-sm"
      onClick={onClose}
      data-testid="usecase-overlay"
    >
      <div
        className="w-full max-w-lg overflow-hidden rounded-xl border border-white/10 bg-slate-900 shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-3 border-b border-white/10 px-4 py-3">
          <svg className="h-4 w-4 shrink-0 text-slate-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search demo scenarios..."
            className="flex-1 bg-transparent text-sm text-slate-100 placeholder-slate-500 outline-none"
            data-testid="usecase-search"
          />
          <kbd className="rounded border border-white/10 bg-slate-800 px-1.5 py-0.5 text-[10px] text-slate-500">
            ESC
          </kbd>
        </div>

        <div
          ref={listRef}
          className="max-h-[50vh] overflow-y-auto overscroll-contain p-2"
          data-testid="usecase-list"
        >
          {flatItems.length === 0 && (
            <p className="px-3 py-6 text-center text-sm text-slate-500">
              No matching scenarios
            </p>
          )}

          {grouped.map((group) => (
            <div key={group.category}>
              <p className="px-3 pb-1 pt-3 text-[11px] font-semibold uppercase tracking-wider text-slate-500">
                {group.category}
              </p>
              {group.items.map((uc) => {
                const idx = flatIdx++;
                const isSelected = idx === selectedIdx;

                return (
                  <button
                    key={uc.id}
                    type="button"
                    data-idx={idx}
                    onClick={() => onSelect(uc)}
                    onMouseEnter={() => setSelectedIdx(idx)}
                    className={`flex w-full flex-col gap-0.5 rounded-lg px-3 py-2.5 text-left transition-colors ${
                      isSelected
                        ? "bg-indigo-500/15 text-slate-100"
                        : "text-slate-300 hover:bg-white/5"
                    }`}
                    data-testid={`usecase-item-${uc.id}`}
                  >
                    <span className="text-sm font-medium">{uc.name}</span>
                    <span className="text-xs text-slate-500">{uc.description}</span>
                  </button>
                );
              })}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
