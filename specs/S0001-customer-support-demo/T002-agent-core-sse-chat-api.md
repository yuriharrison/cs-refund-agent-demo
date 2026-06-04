---
id: T002
name: Agent core with SSE event bus and chat API — first vertical slice of the ReAct loop
status: DONE
deps: [T001]
---

# T002 — Agent Core with SSE Event Bus and Chat API

## Description

Implement the Eino ReAct agent with the `lookup_customer_orders` tool, the SSE event bus, and the chat API endpoints (`POST /api/chat/message`, `GET /api/chat/stream`, `GET /api/chat/history`, `POST /api/chat/reset`). After this task, a user can send a message, the agent processes it through the LLM with one tool available, and events stream in real time via SSE.

This is the first end-to-end validatable slice of the AI agent: message in → agent reasoning → tool call → streamed response out.

## Scope

- `internal/agent/agent.go` — Eino ReAct agent construction, message processing loop with callbacks for event emission
- `internal/agent/system_prompt.go` — system prompt template with customer context injection
- `internal/agent/tools.go` — tool registration framework
- `internal/agent/tools_lookup_orders.go` — `lookup_customer_orders` tool: queries customer orders with items, returns JSON
- `internal/chat/service.go` — chat session management, message persistence (in-memory), session creation/reset
- `internal/chat/events.go` — SSE event bus: `sync.Map` of session → subscribers, publish/subscribe/unsubscribe, buffered channels
- `internal/chat/models.go` — message types (roles: customer, agent, system, human), event type constants
- `internal/api/chat_handler.go` — `POST /api/chat/message` (accepts message, triggers async agent), `GET /api/chat/history`, `POST /api/chat/reset`
- `internal/api/sse_handler.go` — `GET /api/chat/stream` SSE endpoint with keep-alive

## Agent Event Emission

The agent emits these events during processing:
- `agent_thinking` — when agent starts reasoning
- `tool_call_start` — before tool execution (tool name + args)
- `tool_call_result` — after tool execution (result + duration)
- `agent_message` — final response text (streamed in chunks)

## Acceptance Criteria

- [ ] `POST /api/chat/message` returns 202 with `session_id` and `message_id`
- [ ] SSE stream at `/api/chat/stream?session_id=X` emits events as agent processes
- [ ] `lookup_customer_orders` tool correctly returns order data from DB
- [ ] Agent responds conversationally when asked about orders
- [ ] `GET /api/chat/history` returns all messages for a session in order
- [ ] `POST /api/chat/reset` clears session and returns new `session_id`
- [ ] Multiple SSE clients can subscribe to the same session (fan-out)
- [ ] Disconnected SSE clients are cleaned up (channel removed from map)

## Test Cases

### Unit Tests

- `TestEventBus_PublishSubscribe` — publish event, verify subscriber receives it
- `TestEventBus_MultipleSubscribers` — publish event, verify all subscribers receive
- `TestEventBus_Unsubscribe` — unsubscribe, verify no further events received
- `TestChatService_CreateSession` — create session, verify ID generated
- `TestChatService_AddMessage` — add messages, verify history order
- `TestChatService_Reset` — reset session, verify history cleared and new ID returned
- `TestLookupOrdersTool_ReturnsCorrectJSON` — mock DB, verify tool output format matches spec (orders with items, dates, product names)
- `TestLookupOrdersTool_LimitParameter` — verify limit parameter restricts results

### Integration Tests

- `TestChatAPI_SendMessage` — POST message, verify 202 response and session creation
- `TestChatAPI_SSEStream` — send message, connect to SSE, verify events arrive (at least `agent_thinking` and `agent_message`)
- `TestChatAPI_History` — send messages, GET history, verify all present
- `TestChatAPI_Reset` — reset session, verify new session_id and empty history
