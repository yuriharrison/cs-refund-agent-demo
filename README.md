# cs-refund-agent-demo

A self-contained demo of an AI customer support agent that handles refund workflows — gathering order data, evaluating refund policies, making autonomous decisions (full/partial/deny), and escalating to human support when necessary — all visible through a real-time chat interface with full debug observability.

Built with **Go** (backend), **CloudWeGo Eino** (ReAct agent), and **React + Vite** (frontend). See the full [spec](specs/S0001-customer-support-demo/S0001-customer-support-demo.md).

---

## What it does

A customer chats with an AI support agent about a refund. The agent autonomously:

1. Looks up the customer's orders
2. Identifies the product and reason
3. Checks the refund policy for that product type + condition
4. Takes action: issues a full/partial refund, denies it with an explanation, or escalates to a human agent
5. Streams every step (tool calls, reasoning, results) to the frontend via SSE

The frontend renders the full conversation with distinct visual treatment for agent messages, system confirmations (refund cards), escalation notices, human agent handoffs, and collapsible debug panels showing every tool call.

A built-in **demo selector** (command palette) lets you run 11 predefined scenarios that exercise every branch of the refund workflow without typing anything.

---

## Architecture

```
┌─────────────────────────────────────────────────────┐
│  Frontend (Vite + React + Tailwind)                 │
│  Chat UI · SSE Stream · Demo Selector               │
└──────────────────────┬──────────────────────────────┘
                       │ HTTP + SSE
┌──────────────────────▼──────────────────────────────┐
│  Backend (Go + Chi)                                 │
│  ┌────────────────────────────────────────────────┐ │
│  │  Eino ReAct Agent (GPT-5.4 mini)              │ │
│  │  Tools: lookup_orders · get_policy ·           │ │
│  │         issue_refund · escalate_to_human       │ │
│  └────────────────────────────────────────────────┘ │
│  SSE Event Bus · Chat Service · Token Tracker       │
│  GORM + SQLite (seeded on startup)                  │
└─────────────────────────────────────────────────────┘
```

---

## Quick start

### Prerequisites

- Go 1.22+
- Node.js 18+
- An OpenRouter API key (or OpenAI-compatible endpoint) set as `OPEN_ROUTER_API_KEY`

### Setup

```bash
make setup
```

### Run

```bash
make dev
```

This starts the Go backend and Vite dev server concurrently. Open [http://localhost:5173](http://localhost:5173).

### Test

```bash
make test           # Run all Go tests (uses snapshots)
make test-e2e       # Run Playwright E2E tests
make test-refresh   # Re-record snapshots from live API
```

---

## Demo scenarios

Click the **Demo** button in the header to open the command palette. Each scenario is a scripted sequence of customer messages that the agent responds to naturally:

| Scenario | What happens |
|:---|:---|
| Full Refund — Defective Product | Customer reports defective headphones, agent issues full refund |
| Full Refund — Wrong Item | Wrong item shipped, immediate full refund |
| Refund — Product Not Specified | Customer doesn't say which product; agent looks up orders and asks |
| Partial Refund — Customer Accepts | Change-of-mind return, agent offers partial, customer accepts |
| Partial Refund — Customer Declines | Partial offered, customer refuses, agent escalates to human |
| Refund Denied — No Refund Policy | T-shirt change-of-mind, policy says no refund |
| Complaint Then Refund | Customer complains first, then requests refund |
| Feedback Only — No Refund | Customer gives feedback, no refund needed |
| Escalation — Software Policy | Software refunds always require human review |
| Escalation — System Error | Mocked infrastructure failure triggers escalation |
| Subscription — Trial Refund | Cancellation within trial window, full refund |

---

## Project structure

```
cmd/server/main.go              Server entrypoint
internal/
  agent/                        Eino ReAct agent, system prompt, 4 tools
  api/                          Chi router, chat/SSE/usecase handlers
  chat/                         Session service, event bus, models
  db/                           GORM + SQLite setup, seed data
  domain/                       Customer, Order, Product, Refund models
  token/                        Token usage tracking + cost calculation
  usecase/                      Demo scenario registry + step runner
  testutil/                     Snapshot record/replay, test fixtures
web/src/
  components/chat/              ChatContainer, MessageList, MessageBubble,
                                SystemMessage, DebugMessage, HumanMessage,
                                ChatInput, TokenCounter
  components/demo/              UsecaseSelector (command palette)
  hooks/                        useChat (state + SSE), useUsecases
  lib/                          SSE client helper
  styles/                       Global styles
specs/                          Feature spec + implementation tasks
```

---

## How testing works

Tests use a **snapshot system** that records OpenAI API responses on first run and replays them on subsequent runs — making tests fast, deterministic, and free after initial recording.

```bash
# Run with snapshots (default, no API key needed)
make test

# Re-record from live API
make test-refresh

# Run live API tests only
make test-live
```

---

## Makefile commands

| Command | Description |
|:---|:---|
| `make dev` | Start backend + frontend concurrently |
| `make dev-be` | Start Go backend only |
| `make dev-fe` | Start Vite frontend only |
| `make test` | Run all Go tests with snapshots |
| `make test-e2e` | Run Playwright E2E tests |
| `make test-refresh` | Re-record snapshots from live API |
| `make test-live` | Run live API tests against OpenRouter |
| `make setup` | Install Go + Node dependencies |
| `make swagger` | Generate OpenAPI spec |
| `make codegen` | Generate TypeScript API client |
