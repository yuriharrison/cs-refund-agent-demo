# empirical-sdd-ddd-starter

A lightweight, drop-in starter that turns any repository into an **AI-native Spec-Driven + Document-Driven** project.

Markdown only. No install. Works out of the box in Claude Code via slash commands; works anywhere else by reading the playbooks directly.

---

## What you get

- **Six slash commands** (`/sdd-init`, `/sdd-spec`, `/sdd-spec-socratic`, `/sdd-orchestrate`, `/sdd-handoff`, `/sdd-status`)
- **Six roles** (Analyst, PM, Architect, Developer, Tester, Reviewer) — each with concrete gate criteria and explicit send-back rules
- **A Socratic spec-authoring skill** for Business Analysts to turn vague ideas into testable specs
- **An orchestration playbook** that drives a spec through the role loop with validation gates and feedback loops
- **Human-in-the-loop by default**, with an opt-in autonomous mode that hard-stops after 2 failed iterations
- **One state file** (`ai/STATE.md`) — not six, not twelve

---

## Quick start

### 1. Adopt it (copy two folders into your repo)

```bash
cp -R ai/        /path/to/your/repo/ai/
cp -R .claude/   /path/to/your/repo/.claude/
```

See `INSTALL.md` for details.

### 2. Fill context (~15 min)

In Claude Code:

```
/sdd-init
```

Or edit `ai/context/*` directly.

### 3. Draft your first spec

**Business Analyst** (vague idea):
```
/sdd-spec-socratic
```

**Developer** (clear ask):
```
/sdd-spec
```

### 4. Orchestrate

```
/sdd-orchestrate
```

Drives PM → Architect → Developer → Tester → Reviewer with explicit gates. Pauses for you at every step (HITL); say "orchestrate autonomously" to flip the mode.

---

## The flow

```
Stakeholder ask  →  Analyst (optional, Socratic)  →  PM  →  Architect  →  Developer  →  Tester  →  Reviewer  →  DONE

Every arrow has a gate. Every gate has a possible send-back to the shallowest fixing role.
```

See `ai/orchestration/workflow.md` for the full picture.

---

## Who this is for

- **Business Analysts** who want to draft good specs by being interviewed Socratically
- **Developers** who want orchestrated AI assistance without a heavy framework
- **Small / medium teams** who want structure without BMAD's surface area

---

## Design tenets

- **KISS.** Markdown only. No CLI. No npm. No hooks. Nothing to install.
- **Token-cheap at runtime.** Load only the role file for the active phase.
- **Pragmatic.** 3–8 specs per milestone. One screen per spec. One state file.
- **Validation loops by default.** Gates catch problems early; send-backs route to the shallowest fixing role.
- **Human-in-the-loop by default.** Autonomous is opt-in and has guard rails.

---

## What this is not

- Not a framework — there's nothing to install, nothing to maintain
- Not as extensive as **BMAD** (which has 12+ personas, 34+ workflows, Party Mode, three tracks)
- Not as heavy as **GSD** (which has six state files and an npm CLI)

It borrows the lessons of both and intentionally stays smaller. Built for repos that want structure without ceremony.

---

## Where to go next

- `INSTALL.md` — drop-in adoption
- `ai/README.md` — index of the scaffold
- `ai/orchestration/workflow.md` — the flow in one page
- `ai/skills/orchestrate.md` — the orchestration playbook
- `ai/skills/write_spec_socratic.md` — the Socratic interview flow
- `ai/skills/validate_handoff.md` — gates and send-back tree

---

## License

(Choose your license when you adopt this — the scaffold itself imposes none.)
