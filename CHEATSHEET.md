# 🚀 Empirical SDD + DDD Starter — Cheat Sheet

A practical walkthrough using a realistic example: building a **“team standup notes”** feature inside an existing application.

# ⚡ TL;DR — Only 6 Commands


| Command              | Purpose                                                         |
| -------------------- | --------------------------------------------------------------- |
| `/sdd-init`          | First-time project setup (`ai/context/`)                        |
| `/sdd-spec-socratic` | Turn a vague idea into a structured spec (BA-style flow)        |
| `/sdd-spec`          | Fast path for already-clear requirements                        |
| `/sdd-orchestrate`   | Run the full PM → Architect → Dev → Test → Review workflow      |
| `/sdd-status`        | Show current orchestration state                                |
| `/sdd-handoff`       | Generate a structured handoff document (rarely needed manually) |


# 🪜 The 5 Steps

## Step 1 — Adopt It (~30 sec)

Inside your repository:

```bash
cp -R ./empirical-sdd-ddd-starter/ai ./ai
cp -R ./empirical-sdd-ddd-starter/.claude ./.claude
```

✅ Done.

No npm.

No installs.

No hooks.

Just two folders.

💡 Tip: if `.claude/settings.local.json` already exists, merge instead of overwriting.

## Step 2 — Initialize Context (~15 min, once per project)

Open Claude Code and run:

```bash
/sdd-init
```

Claude will guide you through populating:

- `project_vision.md`
- `personas.md`
- `tech_stack.md`
- `architecture_principles.md`
- `domain_glossary.md`
- `current_milestone.md`

💡 If you are unsure about something, use `TBD`.

The framework is designed to evolve incrementally.

## Step 3 — Draft a Spec

### 🅐 Vague Idea → Socratic / BA Flow

```bash
/sdd-spec-socratic
```

Claude runs a structured multi-pass interview covering:

1. Problem & users
2. WHO
3. WHAT
4. WHY
5. Acceptance criteria
6. Risks & assumptions

➡️ Output:

```
ai/specs/001-your-spec/spec.md
```

### 🅑 Clear Ask → Developer Fast Path

```bash
/sdd-spec
```

One request in → draft spec out.

Best for:

- clearly defined requirements
- implementation tasks
- technical enhancements
- bug fixes
- migrations

## Step 4 — Orchestrate

```bash
/sdd-orchestrate
```

Workflow:

```
PM → Architect → Developer → Tester → Reviewer → DONE
```

Each phase has its own lightweight validation gate.

Default mode = **HITL (Human-In-The-Loop)**

Claude pauses at every gate for approval before continuing.

### Autonomous Mode

```
orchestrate autonomously
```

Claude auto-approves gates and continues execution.

⚠️ Safety Rule:

If the same gate fails twice consecutively, orchestration pauses automatically for human review.

## Step 5 — Check Status

```bash
/sdd-status
```

Displays:

- Current spec
- Current role
- Current phase
- Execution mode
- Recent decisions
- Active send-backs

# 📖 Full Example — “Team Standup Notes”

## 🎬 Adoption

```bash
cd ./team-portal

cp -R ./empirical-sdd-ddd-starter/ai ./ai
cp -R ./empirical-sdd-ddd-starter/.claude ./.claude
```

Then run:

```bash
/sdd-init
```

## 🎬 Spec Drafting (Socratic Flow)

```bash
/sdd-spec-socratic
```

Example conversation:

**Claude:**

“What problem are you solving, and for whom?”

**You:**

“I want a Slack-based standup notes tool.”

**Claude:**

“That sounds like a solution. What is the underlying pain?”

**You:**

“Remote engineers forget to share updates, so leads lose visibility.”

Acceptance criteria become:

- Engineers can submit updates through a Slack command
- Leads receive a daily digest before 10:30 AM
- Notes are searchable
- Missing updates trigger reminders by 10:00 AM

➡️ Claude generates:

```
ai/specs/001-standup-notes/spec.md
```

## 🎬 Orchestration

```bash
/sdd-orchestrate
```

### PM Phase

- Refines requirements
- Clarifies edge cases
- Defines scope boundaries

### Architect Phase

Defines architecture & tradeoffs, for example:

- Slack Bolt app
- Postgres storage
- Scheduled digest job

### Developer Phase

- Implements the feature
- May use sub-agents/tools internally

### Tester Phase

- Verifies acceptance criteria
- Validates edge cases

### Reviewer Phase

- Reviews maintainability
- Reviews architecture alignment
- Reviews implementation quality

✅ DONE

# 🧭 Mode Cheat Sheet


| Input                                   | Behavior                       |
| --------------------------------------- | ------------------------------ |
| *(default)*                             | HITL at every gate             |
| `orchestrate autonomously`              | Fully autonomous orchestration |
| `pause`                                 | Return to manual/HITL mode     |
| `orchestrate dev and test autonomously` | Hybrid mode (recommended)      |


# 🩹 Common Situations


| Situation                     | Recommended Action                     |
| ----------------------------- | -------------------------------------- |
| Spec is too large             | Split before Architect phase           |
| Acceptance criteria are vague | Use `/sdd-spec-socratic`               |
| Lost or confused              | Run `/sdd-status`                      |
| Repeated gate failures        | Review `STATE.md` and recent decisions |
| Want to remove the framework  | `rm -rf ai .claude`                    |


# 🧠 Mental Model

```
Stakeholder Ask
        ↓
Analyst (optional Socratic interview)
        ↓
PM
        ↓
Architect
        ↓
Developer
        ↓
Tester
        ↓
Reviewer
        ↓
DONE
```

💡 Send-backs always return to the shallowest role capable of fixing the issue.

# 🎯 Three Core Rules

1. One spec ≈ one screen
2. Gates should stay lightweight, not bureaucratic
3. HITL should remain the default unless the task is truly small

🚀 Enjoy building.