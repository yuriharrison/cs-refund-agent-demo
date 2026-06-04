# Project Vision

## Project Name
customer-support-demo

---

## Problem Statement
Demonstrating an AI agent handling customer refund support in a chat interface, showcasing the agentic flow and decision-making process.

---

## Users
- Developers: Operating the demo, viewing token usage, and observing debug messages.
- Mock Customers: Interacting with the agent to request refunds.

---

## Core Features
- Single page, full-screen chat interface.
- Agentic flow for handling refund requests (gather data, check policy, authorize/deny/escalate).
- Debug system messages illustrating AI function calls and actions.
- Testcase selection box to automatically emulate user flows.
- Token count visibility.

---

## Constraints
- Time sensitive task, strictly limited to defined scope.
- No authentication or login.
- Seed mocked data on a local SQLite database.
- Must use OpenAI API.

---

## Success Criteria
- Agent correctly handles various refund test cases (full, partial, deny, escalate).
- UI clearly distinguishes between Customer, Agent, System, and Human messages.
- Automated testing snapshot pattern successfully replays API responses.
