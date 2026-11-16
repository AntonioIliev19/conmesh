# Project setup guide

The detailed reference for setting up ConMesh in a project. The README has the
short version. Coding agents that set up ConMesh for a user follow this guide
together with [INSTALL.md](INSTALL.md), which says what to ask and confirm
first. The examples use PowerShell.

## 1. Initialize the project

```powershell
Set-Location 'C:\path\to\your-project'
conmesh init --harness codex --index
conmesh status
```

- **What it creates:** `.conmesh/config.yaml`, the project's ConMesh settings.
  The file uses JSON syntax despite its extension.
- **`--harness`:** the AI tools ConMesh integrates with, comma-separated:
  - `claude` and `codex` get the MCP server, the research skill and agent,
    installed packages and published knowledge.
  - `cursor` and `copilot` get published project knowledge as Cursor rules or
    GitHub Copilot instructions (see step 6). ConMesh does not configure MCP for
    them; connect it as described in [INSTALL.md](INSTALL.md#connect-the-mcp-server).
  - For OpenCode or another MCP host, run `conmesh init --index` without
    `--harness` and connect the MCP server in that tool.
- **`--index`:** enables indexing. Without it, indexing stays off; use plain
  `conmesh init` if you only want package management.
- **Already initialized:** `init` keeps the existing settings. To change them,
  edit `.conmesh/config.yaml`: set `indexing.enabled` to `true` or add tools to
  the `harnesses` array.
- **Check the runtime:** `conmesh doctor` reports which features this build
  supports, including native Tree-sitter parsing for indexing.

## 2. Build the local index

```powershell
conmesh index
conmesh search "request handler" --limit 10
conmesh symbol Handler
```

- **What gets indexed:** allowed files only. ConMesh applies its built-in
  exclusions, `.gitignore` and `.conmeshignore`, and skips files that look like
  secrets.
- **Keep it current:** run `conmesh index` again after source changes. Unchanged
  files are reused.
- **Search:** matches the parts of camelCase and snake_case names, so `handler`
  finds `requestHandler`. `conmesh symbol` looks up exact symbol names.
- **Files with syntax errors** or over the parser limits are still indexed as
  plain text, without symbols. The summary line reports how many, and why any
  file was skipped.
- **Upgrades:** a newer ConMesh upgrades the index in place the first time it
  writes to it and re-extracts all files once. An older ConMesh cannot read the
  upgraded index. After going back to an older version, for example with
  `conmesh update --rollback`, delete `.conmesh/state/index.sqlite` and its
  `-wal` and `-shm` files, then run `conmesh index` again.

## 3. Connect your AI tools

```powershell
conmesh sync --dry-run
conmesh sync --trust
```

- **Preview first:** `--dry-run` lists every file sync would write. Review it,
  and the instructions of any installed packages, before `--trust`.
- **Claude Code and Codex:** sync connects the ConMesh MCP server (`.mcp.json`,
  `.codex/config.toml`), installs the `conmesh-research` skill and a custom
  Research Agent, and publishes installed skills, agents and approved knowledge.
- **Other tools** (OpenCode, Cursor, VS Code with Copilot, Windsurf): set up the
  ConMesh MCP server once in the tool, as a stdio server running
  `conmesh mcp serve --project <project root>`. See
  [INSTALL.md](INSTALL.md#connect-the-mcp-server) for each tool's file.
- **Your files are safe:** sync only writes files ConMesh owns, keeps backups,
  and refuses to overwrite anything that was changed outside ConMesh.

## 4. Run research

Send this in a chat with your AI tool, opened in the project:

```text
Run ConMesh research
```

- **Skill picker:** you can also choose `conmesh-research` from your tool's skill
  list.
- **What happens:** your current agent delegates to one local subagent. It checks
  the index, creates `.conmesh/knowledge/manifest.yaml` if missing, registers its
  proposed Markdown documents and validates them, then reports back in the same
  chat.
- **Requirements:** your tool needs local subagent support with MCP and project
  write access.
- **Privacy:** your tool's model provider may receive selected source code during
  research.
- **Other installed skills:** `conmesh sync --trust` publishes them too. In Codex,
  type `$` or use `/skills` to pick one; see
  [Codex skill discovery](https://learn.chatgpt.com/docs/build-skills#how-chatgpt-and-codex-use-skills).

## 5. Review the proposed knowledge

```powershell
conmesh knowledge list
conmesh knowledge show <id>
conmesh knowledge approve <id> --reviewed-digest <digest>
```

Read the Markdown files and the manifest diff. For each entry you want to keep,
`knowledge show` prints the document, its citations and a review digest. Approve
the entry with that digest. Proposed entries are never published.

## 6. Publish the knowledge

```powershell
conmesh sync --dry-run
conmesh sync --trust
```

Each approved entry is published in the format of every tool listed in
`harnesses`, so the tool loads it for the matching code by itself:

| Tool | Published as |
| --- | --- |
| Claude Code | Skill in `.claude/skills/conmesh-kb-<id>/SKILL.md` |
| Codex and other tools that read `.agents/skills` | Skill in `.agents/skills/conmesh-kb-<id>/SKILL.md` |
| Cursor | Rule in `.cursor/rules/conmesh-kb-<id>.mdc`, attached for the entry's scopes |
| GitHub Copilot | Instructions in `.github/instructions/conmesh-kb-<id>.instructions.md`, applied to the entry's scopes |

- **Project agents:** research also proposes agents (entries with `kind: agent`),
  such as a feature developer for each major part of the code. Approved agents
  are published for Claude Code (`.claude/agents/conmesh-agent-<id>.md`) and
  Codex (`.codex/agents/conmesh-agent-<id>.toml`), each pointing to the guidance
  for its part of the code. Cursor and Copilot receive the guidance only.
- **Found by any session:** guidance and agents are listed in a marked "Project
  knowledge" section of `CLAUDE.md` and `AGENTS.md`, so sessions that know
  nothing about ConMesh, with any model, load the right guidance on their own.
  Commit these files to share them with your team.
- **Your instructions stay yours:** in a `CLAUDE.md` or `AGENTS.md` you wrote,
  ConMesh changes only the section between its `BEGIN`/`END CONMESH PROJECT
  KNOWLEDGE` markers, and refuses to overwrite that section if someone edited it.
- **Only current, reviewed guidance is published:** when a cited file changes,
  the entry is withdrawn at the next sync until you review and approve it again.
  `sync --dry-run` lists every entry it left out and why.
- **Edit the source, not the copy:** change `.conmesh/knowledge/`, approve it,
  and sync again. Edits to the published files are overwritten.

## Skill and agent packs

Install packs at any point after step 1; `conmesh sync --trust` publishes them.
See [Skill and agent packs](README.md#skill-and-agent-packs) and
[PACKS.md](PACKS.md).
