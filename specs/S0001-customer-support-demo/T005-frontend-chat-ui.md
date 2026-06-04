---
id: T005
name: Frontend chat UI — React + SSE streaming + all message types
status: DONE
deps: [T002]
---

# T005 — Frontend Chat UI

## Description

Build the complete React frontend: a full-screen dark-mode chat interface with SSE streaming, all message type renderings (customer, agent, system, debug, human), typing indicator, token counter, and session management. After this task, the frontend connects to the backend, displays real-time streamed events, and provides a polished chat experience.

The frontend uses Vite + React + TypeScript + Tailwind CSS + shadcn/ui. The API client is hand-written initially (Orval codegen is wired in T006 with the usecase endpoints).

## Scope

- `web/` — full Vite + React + TypeScript project scaffold
  - `package.json`, `vite.config.ts`, `tailwind.config.ts`, `tsconfig.json`, `index.html`
  - `components.json` — shadcn/ui config
- `web/src/main.tsx` — React root
- `web/src/App.tsx` — app shell, layout
- `web/src/lib/sse.ts` — SSE client helper (connect, parse events, reconnect)
- `web/src/hooks/useChat.ts` — chat state management + SSE subscription
- `web/src/components/chat/ChatContainer.tsx` — full-screen chat layout (header + messages + input)
- `web/src/components/chat/MessageList.tsx` — scrollable message area with auto-scroll
- `web/src/components/chat/MessageBubble.tsx` — customer (right, accent) and agent (left, glass) bubbles
- `web/src/components/chat/SystemMessage.tsx` — centered action cards (refund confirmed, escalation, error)
- `web/src/components/chat/DebugMessage.tsx` — collapsible tool call cards (monospace, muted)
- `web/src/components/chat/HumanMessage.tsx` — warm-toned human agent bubble with badge
- `web/src/components/chat/ChatInput.tsx` — message input + send button, disabled during processing
- `web/src/components/chat/TokenCounter.tsx` — token count + cost label below input
- `web/src/styles/index.css` — global styles, dark theme tokens, glassmorphism utilities
- Update `Makefile` — `dev-fe`, `dev` (concurrent), `setup` targets

## Visual Design

- **Dark mode** default with deep slate/gray palette
- **Glassmorphism** on agent message bubbles and system cards (backdrop-blur, subtle border)
- **Customer messages**: right-aligned, solid indigo-500, white text
- **Agent messages**: left-aligned, glass-morphic with subtle border
- **Human messages**: left-aligned, warm amber/orange hue, "Support Agent" badge
- **System cards**: centered, full-width, colored accent (green/amber/red) based on type
- **Debug cards**: collapsible, monospace, muted gray background
- **Typing indicator**: three pulsing dots with agent icon during processing
- **Smooth animations**: message entry (fade-in + slide-up), typing dots pulse

## Acceptance Criteria

- [ ] `npm run dev` starts Vite dev server proxying to backend at `:8080`
- [ ] Chat connects to SSE stream on session start
- [ ] Customer messages appear right-aligned with accent color
- [ ] Agent messages stream in real-time (word by word from SSE chunks)
- [ ] Tool call events render as collapsible debug cards
- [ ] System confirmation cards show refund details with green accent
- [ ] System escalation cards show reason with amber accent
- [ ] Human agent messages appear with distinct warm styling
- [ ] Typing indicator shows while agent is processing
- [ ] Token counter updates in real-time below input
- [ ] Chat input disabled while agent is processing
- [ ] Session reset button clears chat and starts new session
- [ ] Auto-scroll to bottom on new messages
- [ ] Dark mode looks polished and modern

## Test Cases

### Unit Tests

- `MessageBubble.test.tsx` — renders correct alignment and style per role
- `SystemMessage.test.tsx` — renders correct icon and color per event type
- `DebugMessage.test.tsx` — collapsed by default, expands on click, shows tool details
- `TokenCounter.test.tsx` — formats token count and cost correctly
- `useChat.test.ts` — state transitions: idle → processing → received, message accumulation
- `sse.test.ts` — parses SSE event format correctly, handles reconnection
