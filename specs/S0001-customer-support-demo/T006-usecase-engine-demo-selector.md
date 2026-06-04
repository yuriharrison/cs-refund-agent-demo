---
id: T006
name: Usecase engine — backend registry/runner + frontend command palette + codegen pipeline
status: PENDING
deps: [T003, T005]
---

# T006 — Usecase Engine & Demo Selector

## Description

Implement the usecase system: a backend registry of all 12 scripted demo scenarios, a runner that sends customer messages at timed intervals, and a frontend command palette for selecting and executing usecases. Also wire the OpenAPI codegen pipeline (swaggo + Orval) so the frontend uses a generated TypeScript client.

After this task, a user can open the demo palette, select a scenario (e.g., "Refund — Product Specified"), watch it play out automatically in the chat with proper timing, and see the "Demo in progress..." indicator.

## Scope

### Backend

- `internal/usecase/registry.go` — all 12 usecase definitions from spec §7 (ID, name, description, ordered customer messages)
- `internal/usecase/runner.go` — step-by-step execution: send customer message, wait for agent response, 2s delay, send next message
- `internal/api/usecase_handler.go` — `GET /api/usecases` (list), `POST /api/usecases/{id}/run` (start execution)
- Swagger annotations on all handler functions (swaggo/swag)
- `docs/swagger.json` — generated OpenAPI spec

### Frontend

- `web/src/components/demo/UsecaseSelector.tsx` — floating command palette overlay (searchable list, triggered by header button)
- `web/src/hooks/useUsecases.ts` — fetch usecase list, trigger execution
- Update `web/src/components/chat/ChatContainer.tsx` — header "Demo" button, "Demo in progress..." indicator, disable input during demo
- `web/orval.config.ts` — Orval codegen configuration pointing to `docs/swagger.json`
- `web/src/api/` — generated TypeScript API client (Orval output)

### Codegen Pipeline

- Update `Makefile` — `swagger`, `codegen` targets
- `make codegen` generates OpenAPI spec then runs Orval

## Usecase Definitions

All 12 usecases from spec §7:
1. `refund_product_specified` — 2 messages
2. `refund_no_product` — 3 messages
3. `feedback_only` — 3 messages
4. `complaint_then_refund` — 3 messages
5. `refund_denied` — 3 messages
6. `full_refund_auto` — 2 messages
7. `partial_refund_declined` — 3 messages
8. `partial_refund_accepted` — 3 messages
9. `escalate_by_policy` — 2 messages
10. `escalate_by_error` — 2 messages (+ force-error header)
11. `subscription_trial_refund` — 2 messages
12. `subscription_late_cancel` — 3 messages

## Acceptance Criteria

- [ ] `GET /api/usecases` returns all 12 usecases with correct metadata
- [ ] `POST /api/usecases/{id}/run` starts executing the scenario with timed messages
- [ ] Customer messages appear in chat at 2s intervals after agent responds
- [ ] Frontend command palette opens from header button, shows searchable list
- [ ] Selecting a usecase resets chat and begins execution
- [ ] "Demo in progress..." indicator visible during execution
- [ ] Chat input disabled during demo execution
- [ ] Demo ends naturally when all messages are sent and agent finishes
- [ ] `make swagger` generates `docs/swagger.json`
- [ ] `make codegen` generates TypeScript client in `web/src/api/`
- [ ] Usecase `escalate_by_error` sends the `X-Demo-Force-Error` header

## Test Cases

### Unit Tests

- `TestUsecaseRegistry_AllPresent` — verify all 12 usecases registered with correct IDs
- `TestUsecaseRegistry_MessageCounts` — verify each usecase has expected number of steps
- `TestUsecaseRunner_StepExecution` — mock chat service, verify messages sent in order with delays
- `TestUsecaseRunner_ErrorUsecase` — verify force-error header is passed for `escalate_by_error`
- `UsecaseSelector.test.tsx` — renders all usecases, filters on search, triggers run on select

### Integration Tests

- `TestUsecaseAPI_List` — GET endpoint returns 12 usecases with correct fields
- `TestUsecaseAPI_Run` — start a usecase, verify messages appear in session history
