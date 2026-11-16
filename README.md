<div align="center">
  <img src="./conmesh.svg" alt="ConMesh" width="740">
  <p><strong>The Context Intelligence Layer for AI Agents</strong></p>
  <p>Local-first • Agent-agnostic • Codebase-aware</p>
</div>

**Give coding agents context about the codebase you already have.** ConMesh builds an opt-in local index for code search and symbol lookup. Its Research Agent uses repository evidence to draft scoped Markdown guidance about architecture, conventions, and examples for review. This is designed for established codebases, including enterprise projects where teams want control over what gets indexed and which guidance gets approved.

ConMesh also manages agents and skills and can expose its index over MCP to Claude Code, Codex, and other compatible hosts.

ConMesh ships a catalog of skill and agent packs, grouped by category, that install with one
command and use ConMesh's code index. Start with the recommended **ConMesh code-aware** pack, then add [Caveman](https://github.com/juliusbrussee/caveman)
for token saving, [Matt Pocock's skills](https://github.com/mattpocock/skills) for
engineering workflow, and [ECC](https://github.com/affaan-m/ecc) reviewers and planners. See [Skill and agent packs](#skill-and-agent-packs).

## Fast install with your coding agent

Paste this prompt into a coding-agent session with terminal access.

```text
Set up ConMesh using the steering guide at
https://raw.githubusercontent.com/AntonioIliev19/conmesh/main/INSTALL.md
Read it first. Ask me whether I want CLI-only installation or setup for one or more repositories,
and which AI coding tools I use so you can connect the ConMesh MCP server to each of them.
Show the plan and get my confirmation before installing or changing anything.
Ask separately whether I want to index any repository, and confirm each exact path first.
```

The [installation steering guide](INSTALL.md) also works as a manual checklist.

## Install ConMesh

### Windows (recommended)

- Download the Windows x64 ZIP from [GitHub Releases](https://github.com/AntonioIliev19/conmesh/releases/latest).
- Extract it and run `install.cmd`. It installs for your user account and adds
  ConMesh to your user PATH.
- Open a new terminal and check it works:

```powershell
conmesh --version
```

### From a checkout or your own build

Run the installer with the path to a trusted `conmesh.exe`:

```powershell
# From the ConMesh checkout, with conmesh.exe in its root
.\scripts\install.cmd .\conmesh.exe

# Or from any folder, with full paths
& 'C:\path\to\ConMesh\scripts\install.cmd' 'C:\path\to\ConMesh\conmesh.exe'
```

Then close and reopen your terminal (and an editor such as VS Code if it starts
your terminals) and run `conmesh --version`. To build the executable, see
[Build from source](#build-from-source).

### Use the menu

Run `conmesh` (or `conmesh menu`) in a terminal.

- **Packages:** Install (packs by category, a single skill or agent, or a local
  folder or HTTPS Git URL), Update, List, Uninstall.
- **Project:** Workspace, Doctor.
- **ConMesh:** Help, Self-update (installs a verified release), Quit.
- **"same as" line:** shows the equivalent command for the highlighted item.

| Key | Action |
|---|---|
| Up / Down | Move the selection |
| Enter | Select |
| Esc or Backspace | Go back (at the top level, leave the menu) |
| `q` | Quit the menu |
| Ctrl-C | Exit immediately |

### Scripting

- **No prompts:** `--no-input` never asks; without a terminal, `conmesh` prints
  the command reference instead of the menu.
- **JSON output:** `--json` for `doctor`, `agent list` and `skill list`;
  `status` always prints JSON.
- **No colour:** `--no-color`, `NO_COLOR` or `TERM=dumb`.
- **No network on open:** the menu checks for releases only with
  `CONMESH_AUTO_UPDATE_CHECK=1`, at most once a day.
- **Typos:** a mistyped command such as `conmesh serch` suggests the closest one.

### If `conmesh` is not found

```powershell
$conmeshExe = Join-Path ([Environment]::GetFolderPath('LocalApplicationData')) 'ConMesh\bin\conmesh.exe'
Test-Path $conmeshExe
& $conmeshExe --version
Get-Command conmesh -All
```

If `Test-Path` returns `False`, install ConMesh first. Otherwise restart the
terminal and any editor that started it, so they pick up the new PATH.

## Setup for your project

Three steps connect ConMesh to a project. Steps 4 to 6 turn your codebase into
reviewed guidance that every AI session loads. The [project setup guide](SETUP.md)
explains every option; coding agents setting up ConMesh follow it.

### 1. Initialize ConMesh

```powershell
Set-Location 'C:\path\to\your-project'
conmesh init --harness codex --index
```

Creates `.conmesh/config.yaml`. `--harness` names your AI tools (`claude`,
`codex`, `cursor`, `copilot`, comma-separated) and `--index` turns on code
indexing. Run plain `conmesh init` if you only want package management.

### 2. Build the local index

```powershell
conmesh index
conmesh search "request handler" --limit 10
```

Indexes the allowed files of your project, respecting `.gitignore` and
`.conmeshignore`. Run `conmesh index` again after source changes; unchanged
files are reused.

### 3. Connect your AI tools

```powershell
conmesh sync --dry-run
conmesh sync --trust
```

Preview the changes, then apply them. For Claude Code and Codex this connects
the ConMesh MCP server and adds the research skill and agent. For other tools,
connect the MCP server once as described in the [setup guide](SETUP.md#3-connect-your-ai-tools).

### 4. Run research

Send `Run ConMesh research` in your project's AI chat. A local subagent reads the
index and proposes Markdown guidance about your architecture and conventions
for you to review. Your tool's model provider may receive selected source code.

### 5. Review the proposed knowledge

Review the Markdown files and the manifest diff. For each entry you want to
approve, run `conmesh knowledge show <id>` with its manifest ID. If the document
and citations look right, use the returned review digest in
`conmesh knowledge approve <id> --reviewed-digest <digest>`.

### 6. Publish the knowledge to every session

Run `conmesh sync --dry-run`, then `conmesh sync --trust`.

- **Every tool, its own format:** approved guidance becomes skills for Claude
  Code and Codex, rules for Cursor and instructions for GitHub Copilot, loaded
  for the matching code automatically.
- **Project agents:** approved agents from research are published for Claude
  Code and Codex.
- **Any session finds it:** a marked section in `CLAUDE.md` and `AGENTS.md`
  lists the guidance, so sessions that know nothing about ConMesh use it too.
  Commit these files to share them with your team.
- **Your files stay yours:** ConMesh only changes its own marked section and
  files, and publishes only current, reviewed guidance.
- **Edit the source:** change `.conmesh/knowledge/`, approve, and sync again.

Where each file goes and how stale guidance is handled: [setup guide, step 6](SETUP.md#6-publish-the-knowledge).

## Skill and agent packs

Packs are groups of skills and agents that work together. They ship inside
ConMesh, so installing one needs no network access.

| Category | Pack | What you get | Install |
|---|---|---|---|
| Code-aware (recommended) | `conmesh-code-aware` | `code-search` and `change-impact` skills that answer from the ConMesh index, plus the repository research agent. Start here: it is what makes every other skill and agent work from your code index. | `conmesh pack add conmesh-code-aware --trust` |
| Token saving | `caveman` | Terse output modes (`/conmesh-caveman`, `/conmesh-ultracave`), compressed commit and review messages, small task workflows, and three subagents that return `path:line` tables instead of prose. 12 skills, 3 agents. | `conmesh pack add caveman --trust` |
| Engineering workflow | `mattpocock-skills` | Grilling sessions, specs and tickets, TDD, bug diagnosis, code review and codebase design. 27 skills. | `conmesh pack add mattpocock-skills --trust` |
| Code review and security | `ecc-review` | Code, security, silent-failure, database and performance reviewers, plus dead-code cleanup. 6 agents. | `conmesh pack add ecc-review --trust` |
| Planning and architecture | `ecc-planning` | Planner, architect, code explorer and doc updater. 4 agents. | `conmesh pack add ecc-planning --trust` |
| Build and test | `ecc-build-test` | Build-error resolver, TDD guide, E2E runner and a verification loop. | `conmesh pack add ecc-build-test --trust` |
| Languages and frameworks | `ecc-language-reviewers`, `conmesh-frameworks` | TypeScript, React, Python, Go, Rust, Java and C# reviewers; Angular and .NET agents. | `conmesh pack add <pack> --trust` |

```text
conmesh pack list                     # every pack, by category
conmesh pack show caveman             # what a pack installs
conmesh pack add caveman --trust      # install it
conmesh skill add builtin:tdd --trust # or a single skill (agent add for agents)
conmesh sync --trust                  # publish to Claude Code and Codex
```

- **Install a pack or a single item:** `conmesh pack add <pack> --trust` or
  `conmesh skill add builtin:<id> --trust` (`agent add` for agents). Items it
  needs come along automatically.
- **Choose where:** ConMesh asks whether to install for **this project** or
  **all your projects** (global). Pass `--project` or `--global` to skip the
  question.
- **Publish to your AI tools:** run `conmesh sync --trust` in the project
  (`--dry-run` previews it). Global installs reach each project on its next sync.
- **Find them by name:** published skills and agents start with `conmesh-`, for
  example `/conmesh-caveman` or the `conmesh-code-reviewer` agent.
- **Turn on indexing for the full benefit:** with `conmesh init --index`, every
  published skill and agent is told to find code through the ConMesh index.
- **Review third-party packs before use:** Caveman, Matt Pocock and ECC content
  is maintained by its authors.

[PACKS.md](PACKS.md) explains how installs, sync, naming and code-aware
publishing work in detail.

## Third-party agents and skills

ConMesh can install agents and skills from three kinds of source: the packs and
predefined packages shipped with ConMesh (`builtin:<id>`), a local folder, or any
HTTPS Git repository. ConMesh verifies what it copies (manifest, license, file
limits, no secrets, no executable hooks), but it does not review what a
third-party package tells your agent to do.

**Third-party packages are provided by their authors, not by the ConMesh
project.** This includes the Caveman, Matt Pocock and ECC packs in the catalog
as well as anything installed from a local folder or a Git URL. Their authors
are responsible for their content, license and behaviour. Before you pass
`--trust`, read the package's instructions and license yourself, and install
only packages you trust. The ConMesh project and its contributors accept no
responsibility for third-party packages or for anything an agent does while
following them.

## Implementation status

Implemented: `--version`, help, `doctor`, `init`, `status`, scoped agent/skill
add/list/remove, explicit package update, opt-in incremental SQLite/FTS5 indexing,
TypeScript/C# symbols, authorized search, read-only MCP stdio and owned project
sync for Codex/Claude, scoped knowledge with explicit approval, a portable
in-chat research skill, opt-in local semantic retrieval and audited
last-resort fallback.
GitHub release binaries include the trusted public key for signed updates. Source builds require a one-time install of a release binary to use self-update.

## Build from source

Requires Go 1.27.4. Native indexing builds need a C compiler; other CLI commands
and read-only index queries can use a CGO-disabled build. SQLite is embedded;
no separate database installation or model is required.

```text
go test ./...
go build -o conmesh.exe ./cmd/conmesh
go run ./cmd/conmesh --version
go run ./cmd/conmesh --help
go run ./cmd/conmesh doctor
```

Use `-o conmesh` on Linux/macOS. `doctor` reports implementation status without
scanning. Indexing requires explicit opt-in.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md).

## License

ConMesh is licensed under [Apache-2.0](LICENSE). See [third-party notices](THIRD_PARTY_NOTICES.md) for dependencies.

## External inspirations (not code dependencies)

- [AWS AI-DLC workflows](https://github.com/awslabs/aidlc-workflows) — harness-neutral methodologies.
- [Yuval Yeret Portfolio PDLC](https://github.com/yyeret/portfolio-pdlc) — evidence-based workflows.
- [Matt Pocock skills](https://github.com/mattpocock/skills) — composable skills, shipped as an optional catalog pack.
- [ECC](https://github.com/affaan-m/ecc) — a large agent and skill collection; the catalog's packs follow its idea, backed by ConMesh's code index.

Do not vendor or copy external repositories outside the catalog. Catalog content
is copied only from a pinned commit, after reviewing its source, licence and
compatibility, with its licence files and a record of ConMesh's changes.


## Release versions

Release tags use `vMAJOR.MINOR.PATCH`. The first stable release is `v1.0.0`.
Use `v1.0.1` for compatible fixes and `v1.1.0` for compatible features.
Signed release builds can check for updates from the menu. Set
`CONMESH_AUTO_UPDATE_CHECK=1` to opt into background checks.
