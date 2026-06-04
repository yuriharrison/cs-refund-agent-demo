import { useCallback, useEffect, useRef, useState } from "react";
import { connectSSE, type SSEConnection, type SSEEvent } from "../lib/sse";

export interface TokenStats {
  promptTokens: number;
  completionTokens: number;
  total: number;
}

export interface CustomerMessage {
  id: string;
  type: "customer";
  content: string;
}

export interface AgentMessage {
  id: string;
  type: "agent";
  content: string;
  streaming: boolean;
}

export interface HumanMessage {
  id: string;
  type: "human";
  content: string;
  agentName: string;
}

export interface SystemRefundMessage {
  id: string;
  type: "system";
  variant: "refund";
  refundId: string;
  amount: number;
  refundType: string;
  orderItemId: string;
}

export interface SystemEscalationMessage {
  id: string;
  type: "system";
  variant: "escalation";
  reason: string;
  humanAgentName: string;
}

export interface SystemErrorMessage {
  id: string;
  type: "system";
  variant: "error";
  message: string;
}

export type SystemMessage =
  | SystemRefundMessage
  | SystemEscalationMessage
  | SystemErrorMessage;

export interface DebugStartMessage {
  id: string;
  type: "debug";
  variant: "start";
  tool: string;
  arguments: Record<string, unknown>;
}

export interface DebugResultMessage {
  id: string;
  type: "debug";
  variant: "result";
  tool: string;
  result: unknown;
  durationMs: number;
}

export type DebugMessage = DebugStartMessage | DebugResultMessage;

export type ChatMessage =
  | CustomerMessage
  | AgentMessage
  | HumanMessage
  | SystemMessage
  | DebugMessage;

const STREAM_FINALIZE_MS = 500;

async function postMessage(
  content: string,
  sessionId: string | null,
): Promise<{ sessionId: string; messageId: string }> {
  const body: { content: string; session_id?: string } = { content };
  if (sessionId) {
    body.session_id = sessionId;
  }

  const res = await fetch("/api/chat/message", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });

  if (!res.ok) {
    throw new Error(`Failed to send message (${res.status})`);
  }

  const data = (await res.json()) as { session_id: string; message_id: string };
  return { sessionId: data.session_id, messageId: data.message_id };
}

async function resetSession(sessionId: string): Promise<string> {
  const res = await fetch("/api/chat/reset", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ session_id: sessionId }),
  });

  if (!res.ok) {
    throw new Error(`Failed to reset session (${res.status})`);
  }

  const data = (await res.json()) as { session_id: string };
  return data.session_id;
}

export function useChat() {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const [isTyping, setIsTyping] = useState(false);
  const [isConcluded, setIsConcluded] = useState(false);
  const [tokenStats, setTokenStats] = useState<TokenStats>({
    promptTokens: 0,
    completionTokens: 0,
    total: 0,
  });

  const sseRef = useRef<SSEConnection | null>(null);
  const streamingAgentIdRef = useRef<string | null>(null);
  const streamTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const sessionIdRef = useRef<string | null>(null);

  useEffect(() => {
    sessionIdRef.current = sessionId;
  }, [sessionId]);

  const clearStreamTimeout = useCallback(() => {
    if (streamTimeoutRef.current) {
      clearTimeout(streamTimeoutRef.current);
      streamTimeoutRef.current = null;
    }
  }, []);

  const finalizeAgentStream = useCallback(() => {
    clearStreamTimeout();
    const agentId = streamingAgentIdRef.current;
    if (!agentId) return;

    streamingAgentIdRef.current = null;
    setMessages((prev) =>
      prev.map((msg) =>
        msg.type === "agent" && msg.id === agentId
          ? { ...msg, streaming: false }
          : msg,
      ),
    );
    setIsProcessing(false);
    setIsTyping(false);
  }, [clearStreamTimeout]);

  const scheduleStreamFinalize = useCallback(() => {
    clearStreamTimeout();
    streamTimeoutRef.current = setTimeout(() => {
      finalizeAgentStream();
    }, STREAM_FINALIZE_MS);
  }, [clearStreamTimeout, finalizeAgentStream]);

  const disconnectSSE = useCallback(() => {
    sseRef.current?.close();
    sseRef.current = null;
  }, []);

  const handleSSEEvent = useCallback(
    (event: SSEEvent) => {
      switch (event.type) {
        case "agent_thinking": {
          setIsTyping(true);
          break;
        }

        case "agent_message": {
          setIsTyping(false);
          const { content } = event.data as { content: string };

          if (!streamingAgentIdRef.current) {
            const id = crypto.randomUUID();
            streamingAgentIdRef.current = id;
            setMessages((prev) => [
              ...prev,
              { id, type: "agent", content, streaming: true },
            ]);
          } else {
            const agentId = streamingAgentIdRef.current;
            setMessages((prev) =>
              prev.map((msg) =>
                msg.type === "agent" && msg.id === agentId
                  ? { ...msg, content: msg.content + content }
                  : msg,
              ),
            );
          }
          scheduleStreamFinalize();
          break;
        }

        case "tool_call_start": {
          if (streamingAgentIdRef.current) {
            finalizeAgentStream();
          }
          const { tool, arguments: args } = event.data as {
            tool: string;
            arguments: Record<string, unknown>;
          };
          setMessages((prev) => [
            ...prev,
            {
              id: crypto.randomUUID(),
              type: "debug",
              variant: "start",
              tool,
              arguments: args ?? {},
            },
          ]);
          break;
        }

        case "tool_call_result": {
          const { tool, result, duration_ms } = event.data as {
            tool: string;
            result: unknown;
            duration_ms: number;
          };
          setMessages((prev) => [
            ...prev,
            {
              id: crypto.randomUUID(),
              type: "debug",
              variant: "result",
              tool,
              result,
              durationMs: duration_ms ?? 0,
            },
          ]);
          break;
        }

        case "system_confirmation": {
          if (streamingAgentIdRef.current) {
            finalizeAgentStream();
          }
          const { action, details } = event.data as {
            action: string;
            details: {
              refund_id: string;
              amount: number;
              type: string;
              order_item_id: string;
            };
          };
          if (action === "refund_issued") {
            setMessages((prev) => [
              ...prev,
              {
                id: crypto.randomUUID(),
                type: "system",
                variant: "refund",
                refundId: details.refund_id,
                amount: details.amount,
                refundType: details.type,
                orderItemId: details.order_item_id,
              },
            ]);
          }
          setIsProcessing(false);
          setIsTyping(false);
          setIsConcluded(true);
          break;
        }

        case "system_escalation": {
          if (streamingAgentIdRef.current) {
            finalizeAgentStream();
          }
          const { reason, human_agent_name } = event.data as {
            reason: string;
            human_agent_name: string;
          };
          setMessages((prev) => [
            ...prev,
            {
              id: crypto.randomUUID(),
              type: "system",
              variant: "escalation",
              reason,
              humanAgentName: human_agent_name,
            },
          ]);
          setIsProcessing(false);
          setIsTyping(false);
          setIsConcluded(true);
          break;
        }

        case "human_message": {
          if (streamingAgentIdRef.current) {
            finalizeAgentStream();
          }
          const { content, agent_name } = event.data as {
            content: string;
            agent_name: string;
          };
          setMessages((prev) => [
            ...prev,
            {
              id: crypto.randomUUID(),
              type: "human",
              content,
              agentName: agent_name,
            },
          ]);
          setIsProcessing(false);
          setIsTyping(false);
          setIsConcluded(true);
          break;
        }

        case "token_update": {
          const { prompt_tokens, completion_tokens, total } = event.data as {
            prompt_tokens: number;
            completion_tokens: number;
            total: number;
          };
          setTokenStats((prev) => ({
            promptTokens: prev.promptTokens + prompt_tokens,
            completionTokens: prev.completionTokens + completion_tokens,
            total: prev.total + total,
          }));
          break;
        }

        case "error": {
          if (streamingAgentIdRef.current) {
            finalizeAgentStream();
          }
          const { message } = event.data as { message: string };
          setMessages((prev) => [
            ...prev,
            {
              id: crypto.randomUUID(),
              type: "system",
              variant: "error",
              message,
            },
          ]);
          setIsProcessing(false);
          setIsTyping(false);
          break;
        }
      }
    },
    [finalizeAgentStream, scheduleStreamFinalize],
  );

  const connectToSession = useCallback(
    (id: string) => {
      disconnectSSE();
      sseRef.current = connectSSE(id, handleSSEEvent);
    },
    [disconnectSSE, handleSSEEvent],
  );

  const sendMessage = useCallback(
    async (content: string) => {
      const trimmed = content.trim();
      if (!trimmed || isProcessing) return;

      const customerMsg: CustomerMessage = {
        id: crypto.randomUUID(),
        type: "customer",
        content: trimmed,
      };
      setMessages((prev) => [...prev, customerMsg]);
      setIsProcessing(true);
      setIsTyping(true);

      try {
        const result = await postMessage(trimmed, sessionIdRef.current);
        if (!sessionIdRef.current) {
          setSessionId(result.sessionId);
          connectToSession(result.sessionId);
        }
      } catch (err) {
        setIsProcessing(false);
        setMessages((prev) => [
          ...prev,
          {
            id: crypto.randomUUID(),
            type: "system",
            variant: "error",
            message:
              err instanceof Error ? err.message : "Failed to send message",
          },
        ]);
      }
    },
    [isProcessing, connectToSession],
  );

  const reset = useCallback(async () => {
    clearStreamTimeout();
    streamingAgentIdRef.current = null;
    disconnectSSE();

    let newSessionId: string | null = null;

    if (sessionIdRef.current) {
      try {
        newSessionId = await resetSession(sessionIdRef.current);
      } catch (err) {
        setMessages([
          {
            id: crypto.randomUUID(),
            type: "system",
            variant: "error",
            message:
              err instanceof Error ? err.message : "Failed to reset session",
          },
        ]);
        setIsProcessing(false);
        setIsTyping(false);
        return;
      }
    }

    setMessages([]);
    setTokenStats({ promptTokens: 0, completionTokens: 0, total: 0 });
    setIsProcessing(false);
    setIsTyping(false);
    setIsConcluded(false);

    if (newSessionId) {
      setSessionId(newSessionId);
      connectToSession(newSessionId);
    } else {
      setSessionId(null);
    }
  }, [clearStreamTimeout, disconnectSSE, connectToSession]);

  useEffect(() => {
    return () => {
      clearStreamTimeout();
      disconnectSSE();
    };
  }, [clearStreamTimeout, disconnectSSE]);

  return {
    messages,
    sessionId,
    isProcessing,
    isTyping,
    isConcluded,
    tokenStats,
    sendMessage,
    reset,
  };
}
