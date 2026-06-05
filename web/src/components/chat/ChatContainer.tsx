import { useState } from "react";
import { useChat } from "../../hooks/useChat";
import { useUsecases } from "../../hooks/useUsecases";
import { UsecaseSelector } from "../demo/UsecaseSelector";
import { ChatInput } from "./ChatInput";
import { MessageList } from "./MessageList";
import { TokenCounter } from "./TokenCounter";

function formatSessionId(id: string | null): string {
  if (!id) return "—";
  return id.length > 8 ? id.slice(0, 8) : id;
}

export function ChatContainer() {
  const {
    messages,
    sessionId,
    isProcessing,
    isTyping,
    isConcluded,
    isDemoRunning,
    tokenStats,
    sendMessage,
    reset,
    prepareForDemo,
  } = useChat();

  const { usecases, run: runUsecase } = useUsecases();
  const [selectorOpen, setSelectorOpen] = useState(false);

  const handleSelectUsecase = async (uc: { id: string }) => {
    setSelectorOpen(false);
    try {
      const sid = await prepareForDemo();
      await runUsecase(uc.id, sid);
    } catch {
      // errors surface through SSE events
    }
  };

  const inputDisabled = isProcessing || isConcluded || isDemoRunning;

  return (
    <div className="flex h-full flex-col" data-testid="chat-container">
      <header className="flex shrink-0 items-center justify-between border-b border-white/10 bg-slate-900/80 px-6 py-4 backdrop-blur-sm" data-testid="chat-header">
        <div className="flex items-center gap-3">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-500/20 text-indigo-400">
            <svg className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
              />
            </svg>
          </div>
          <div>
            <h1 className="text-sm font-semibold text-slate-100">ShopEase Support</h1>
            <p className="text-xs text-slate-500" data-testid="session-id">
              {isDemoRunning ? (
                <span className="inline-flex items-center gap-1.5 text-amber-400">
                  <span className="relative flex h-2 w-2">
                    <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-amber-400 opacity-75" />
                    <span className="relative inline-flex h-2 w-2 rounded-full bg-amber-400" />
                  </span>
                  Demo in progress…
                </span>
              ) : (
                <>Session #{formatSessionId(sessionId)}</>
              )}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setSelectorOpen(true)}
            disabled={isDemoRunning}
            className="flex items-center gap-1.5 rounded-lg border border-indigo-500/30 bg-indigo-500/10 px-3 py-1.5 text-xs font-medium text-indigo-300 transition-colors hover:border-indigo-500/50 hover:bg-indigo-500/20 disabled:cursor-not-allowed disabled:opacity-40"
            data-testid="demo-button"
          >
            <svg className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
              <path strokeLinecap="round" strokeLinejoin="round" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Demo
          </button>
          <button
            type="button"
            onClick={() => void reset()}
            disabled={(!sessionId && messages.length === 0) || isDemoRunning}
            className="rounded-lg border border-white/10 px-3 py-1.5 text-xs text-slate-400 transition-colors hover:border-white/20 hover:bg-white/5 hover:text-slate-200 disabled:cursor-not-allowed disabled:opacity-40"
            data-testid="reset-button"
          >
            Reset
          </button>
        </div>
      </header>

      <MessageList messages={messages} isTyping={isTyping} />

      <footer className="shrink-0 border-t border-white/10 bg-slate-900/80 px-6 py-4 backdrop-blur-sm" data-testid="chat-footer">
        <div className="mx-auto max-w-3xl">
          <div className="relative">
            <ChatInput onSend={sendMessage} disabled={inputDisabled} />
            {isConcluded && !isDemoRunning && (
              <div
                className="absolute inset-0 flex items-center justify-center rounded-xl bg-slate-900/70 backdrop-blur-sm"
                data-testid="concluded-overlay"
              >
                <span className="text-sm font-medium text-slate-300">
                  Interaction concluded
                </span>
              </div>
            )}
          </div>
          <div className="mt-2">
            <TokenCounter stats={tokenStats} />
          </div>
        </div>
      </footer>

      {selectorOpen && (
        <UsecaseSelector
          usecases={usecases}
          onSelect={handleSelectUsecase}
          onClose={() => setSelectorOpen(false)}
        />
      )}
    </div>
  );
}
