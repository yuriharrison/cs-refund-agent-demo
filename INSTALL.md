# Install / Drop In

This is a starter scaffold, not a framework you install. To use it in **your** repository:

---

## Three-step adoption

### 1. Copy two folders into your repo

```bash
# from this starter, copy the two folders to your target repo
cp -R ai/        /path/to/your/repo/ai/
cp -R .claude/   /path/to/your/repo/.claude/
```

That's it — no install, no npm, no hooks. Pure markdown.

### 2. Fill in context (one-time, ~15 minutes)

Open and edit, in order:

- `ai/context/project_vision.md` — what you're building and why
- `ai/context/personas.md` — who uses it
- `ai/context/tech_stack.md` — what's in the box
- `ai/context/architecture_principles.md` — your constraints
- `ai/context/domain_glossary.md` — shared terminology
- `ai/context/current_milestone.md` — where you are now

You can use the `/sdd-init` slash command in Claude Code to be walked through this interactively.

### 3. Start working

Two entry points depending on who you are:

**Business Analyst** — vague idea to concrete spec:
```
/sdd-spec-socratic
```
Runs a five-pass Socratic interview; produces a draft `spec.md`.

**Developer** — known spec to shipped code:
```
/sdd-orchestrate
```
Runs the gated PM → Architect → Developer → Tester → Reviewer loop.

See `README.md` for the full flow.

---

## What you do NOT need

- Node.js
- npm / pnpm / yarn
- A specific editor
- Any installed CLI

The whole scaffold is markdown. The slash commands work in Claude Code out of the box. If you don't use Claude Code, you can still read and follow the skills directly — they're written as natural-language playbooks.

---

## Folder map (what each thing is)

```
your-repo/
├── .claude/commands/   ← slash commands (Claude Code picks these up automatically)
└── ai/
    ├── STATE.md        ← single state file — the only always-loaded file at runtime
    ├── context/        ← your project's context (you fill this in)
    ├── roles/          ← role definitions: Analyst, PM, Architect, Developer, Tester, Reviewer
    ├── skills/         ← reusable playbooks (write_spec, orchestrate, validate_handoff, …)
    ├── orchestration/  ← the workflow, handoff rules, HITL policy
    ├── templates/      ← prompt templates for common tasks
    ├── contracts/      ← handoff/question/feedback contracts
    └── specs/          ← one folder per spec (use _template/ as the starting point)
```

---

## Removing it later

Delete the two folders. Nothing else touches your repo.
