# Evaluation: with and without ConMesh

This folder holds the harness for checking whether ConMesh makes a coding agent better on a real
project. It supports the coursework experiment: the same tasks are run with no AI steering ("bare")
and with ConMesh set up ("conmesh"), and the results are compared.

The target project, task prompts, hidden tests and run results are kept **outside this public
repository**, because they contain code from a private project. Only the generic tooling lives here.

## Conditions

| Condition | What the agent gets |
|---|---|
| `bare` | A fresh clone of the project with all hand-written AI steering removed (`CLAUDE.md`, `AGENTS.md`, `.agents/`, `.claude/`, `.codex/`). No MCP servers. |
| `conmesh` | The same clone, then `conmesh init --harness claude --index`, `conmesh index`, `conmesh pack add conmesh-code-aware`, `conmesh sync --trust`. The agent gets the ConMesh MCP server, the code-search and change-impact skills and the research agent. |
| `conmesh-directed` | As `conmesh`, plus a "Code navigation (required)" section in `CLAUDE.md` telling the agent to search the ConMesh index before Grep, Glob or Read. This separates "ConMesh is not useful" from "ConMesh is not used". |

Both conditions use the same model, prompt and allowed tools, and run headless with `claude -p`.
User-level settings are excluded (`--setting-sources project,local`), and only the MCP servers named
for that condition are loaded (`--strict-mcp-config`).

## Task folder layout

```
tasks/<id>/
  task.json      id, type (feature|defect|refactor), seeds, hiddenTests, testFilter, grepChecks
  prompt.md      what the agent is told (a symptom or request, no file names)
  hidden/*.cs    acceptance tests the agent never sees; added after the run
```

`seeds` are exact text replacements applied before the run (to plant a defect). `grepChecks`
count pattern matches after the run (to check that a refactor is complete).

## Running

```powershell
.\run-eval.ps1 -Mirror D:\eval\project.git -TaskDir D:\eval\tasks\T1-defect -WorkRoot D:\eval -Condition bare -Rep 1 -Solution App.sln -TestProject tests/App.Tests
.\run-eval.ps1 -Mirror D:\eval\project.git -TaskDir D:\eval\tasks\T1-defect -WorkRoot D:\eval -Condition conmesh -Rep 1 -Solution App.sln -TestProject tests/App.Tests
.\summarize.ps1 -Csv D:\eval\results\runs.csv
```

To test a local ConMesh build instead of the installed one, pass `-ConmeshExe <path to conmesh.exe>`
and a separate `-Rep` range so the rows can be told apart (`conmesh_version` also records the build).

Each run writes `results/<run-id>/` (the agent's JSON output, diff and test logs) and adds one row
to `results/runs.csv`.

## What is measured

| Column | Meaning |
|---|---|
| `accepted` | Hidden tests pass, the full suite passes, the grep checks pass and the run did not time out. This stands in for "accepted without human correction". |
| `wall_seconds` | Time from starting the agent to its final answer |
| `num_turns`, `*_tokens`, `cost_usd` | Agent effort and AI cost, as reported by Claude Code |
| `files_changed`, `lines_added`, `lines_deleted` | Size of the agent's change |
| `setup_seconds` | One-off ConMesh setup time (zero for `bare`) |

## Threats to validity

- Model output is non-deterministic. Repeat each task and condition (`-Rep`) and compare medians.
- Running tasks in sequence on the same machine shares nothing between runs (fresh clones, no
  session persistence), but the model and service can change over time. Record the date and the
  versions; alternate the order of conditions.
- Three tasks on one project is a small sample. Report it as a case study, not as proof.
- The `conmesh` condition does not yet include approved research knowledge (`.conmesh/knowledge`).
  That is a separate, stronger condition worth adding once the research output has been reviewed.
