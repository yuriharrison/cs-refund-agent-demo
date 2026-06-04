import { useChat } from "../../hooks/useChat";
import { ChatInput } from "./ChatInput";
import { MessageList } from "./MessageList";
import { TokenCounter } from "./TokenCounter";

function formatSessionId(id: string | null): string {
  if (!id) return "—";
  return id.length > 8 ? id.slice(0, 8) : id;
}

export function ChatContainer() {
  const { messages, sessionId, isProcessing, isTyping, isConcluded, tokenStats, sendMessage, reset } =
    useChat();

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
              Session #{formatSessionId(sessionId)}
            </p>
          </div>
        </div>
        <button
          type="button"
          onClick={() => void reset()}
          disabled={!sessionId && messages.length === 0}
          className="rounded-lg border border-white/10 px-3 py-1.5 text-xs text-slate-400 transition-colors hover:border-white/20 hover:bg-white/5 hover:text-slate-200 disabled:cursor-not-allowed disabled:opacity-40"
          data-testid="reset-button"
        >
          Reset
        </button>
      </header>

      <MessageList messages={messages} isTyping={isTyping} />

      <footer className="shrink-0 border-t border-white/10 bg-slate-900/80 px-6 py-4 backdrop-blur-sm" data-testid="chat-footer">
        <div className="mx-auto max-w-3xl">
          <div className="relative">
            <ChatInput onSend={sendMessage} disabled={isProcessing || isConcluded} />
            {isConcluded && (
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
    </div>
  );
}
