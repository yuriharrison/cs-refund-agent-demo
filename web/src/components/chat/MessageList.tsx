import { useEffect, useRef } from "react";
import type { ChatMessage } from "../../hooks/useChat";
import { DebugMessage } from "./DebugMessage";
import { HumanMessage } from "./HumanMessage";
import { MessageBubble } from "./MessageBubble";
import { SystemMessage } from "./SystemMessage";

interface MessageListProps {
  messages: ChatMessage[];
  isTyping: boolean;
}

function TypingIndicator() {
  return (
    <div className="flex animate-fade-in-up justify-start" data-testid="typing-indicator">
      <div className="glass flex items-center gap-1.5 rounded-2xl px-4 py-3">
        <span
          className="h-2 w-2 rounded-full bg-slate-400 animate-pulse-dot"
          style={{ animationDelay: "0ms" }}
        />
        <span
          className="h-2 w-2 rounded-full bg-slate-400 animate-pulse-dot"
          style={{ animationDelay: "200ms" }}
        />
        <span
          className="h-2 w-2 rounded-full bg-slate-400 animate-pulse-dot"
          style={{ animationDelay: "400ms" }}
        />
      </div>
    </div>
  );
}

function renderMessage(message: ChatMessage) {
  switch (message.type) {
    case "customer":
      return (
        <MessageBubble key={message.id} role="customer" content={message.content} />
      );
    case "agent":
      return (
        <MessageBubble
          key={message.id}
          role="agent"
          content={message.content}
          streaming={message.streaming}
        />
      );
    case "human":
      return (
        <HumanMessage
          key={message.id}
          content={message.content}
          agentName={message.agentName}
        />
      );
    case "system":
      return <SystemMessage key={message.id} message={message} />;
    case "debug":
      return <DebugMessage key={message.id} message={message} />;
  }
}

export function MessageList({ messages, isTyping }: MessageListProps) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages, isTyping]);

  return (
    <div className="flex-1 overflow-y-auto px-4 py-6" data-testid="message-list">
      <div className="mx-auto flex max-w-3xl flex-col gap-4">
        {messages.length === 0 && !isTyping && (
          <div className="flex flex-col items-center justify-center py-24 text-center" data-testid="empty-state">
            <p className="text-lg font-medium text-slate-300">ShopEase Support</p>
            <p className="mt-2 max-w-sm text-sm text-slate-500">
              Hi! I&apos;m your AI support agent. Ask me about orders, refunds, or
              anything else — I&apos;m here to help.
            </p>
          </div>
        )}
        {messages.map(renderMessage)}
        {isTyping && <TypingIndicator />}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
