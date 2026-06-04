export type SSEEventType =
  | "agent_thinking"
  | "tool_call_start"
  | "tool_call_result"
  | "agent_message"
  | "system_confirmation"
  | "system_escalation"
  | "human_message"
  | "token_update"
  | "error";

export interface SSEEvent<T = unknown> {
  type: SSEEventType;
  data: T;
}

export type SSEEventHandler = (event: SSEEvent) => void;

export interface SSEConnection {
  close: () => void;
}

export function connectSSE(
  sessionId: string,
  onEvent: SSEEventHandler,
  onError?: (error: Event) => void,
): SSEConnection {
  const url = `/api/chat/stream?session_id=${encodeURIComponent(sessionId)}`;
  const source = new EventSource(url);

  const eventTypes: SSEEventType[] = [
    "agent_thinking",
    "tool_call_start",
    "tool_call_result",
    "agent_message",
    "system_confirmation",
    "system_escalation",
    "human_message",
    "token_update",
    "error",
  ];

  for (const type of eventTypes) {
    source.addEventListener(type, (e: MessageEvent) => {
      try {
        const data = JSON.parse(e.data as string);
        onEvent({ type, data });
      } catch {
        onEvent({ type, data: e.data });
      }
    });
  }

  source.onerror = (err) => {
    onError?.(err);
  };

  return {
    close: () => {
      source.close();
    },
  };
}
