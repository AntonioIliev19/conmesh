# Skill and agent packs: how they work

This page explains how ConMesh installs and publishes the packs from its
catalog. For the list of packs and the everyday commands, see
[Skill and agent packs](README.md#skill-and-agent-packs) in the README.

## Where a package is installed

In a terminal, `conmesh pack add`, `conmesh skill add` and `conmesh agent add`
ask whether to install for **this project** or **all your projects** (global),
unless you pass `--project` or `--global`. Scripts, redirected input and
`--no-input` do not ask and use the project.

- **Project scope** stores the package in the project's `.conmesh` folder. It
  needs a ConMesh project (`conmesh init`).
- **Global scope** stores it in your user ConMesh folder. A global package
  reaches a project the next time you run `conmesh sync` in that project.

When the same package ID is installed in more than one place, a project
override wins over a project package, which wins over a global one.

## What gets installed together

A single skill or agent also installs the predefined packages it calls by name.
For example, `grill-with-docs` also installs `grilling` and `domain-modeling`.

A pack installs in one step. If one of its packages cannot be installed, for
example because you already have your own package with that ID, nothing is
installed and the store is left as it was. Installing a pack again reports the
packages that are already up to date as unchanged.

## How packages reach your AI tools

Installing a package copies it into ConMesh's package store. `conmesh sync
--trust` publishes it to the tools configured for the project: skills and
agents for Claude Code (`.claude/`) and Codex (`.agents/`, `.codex/`).
`conmesh sync --dry-run` shows the changes first.

Published skills and agents use ConMesh names, such as `/conmesh-tdd` or the
`conmesh-code-reviewer` agent. ConMesh only ever writes files with its own
prefix, so it never overwrites your files. References between the skills of a
pack use the same names. Matt Pocock's `research` skill is published as
`conmesh-source-research`, because ConMesh already publishes a
`conmesh-research` skill of its own.

Each skill keeps its description and, where set, `argument-hint` and
`disable-model-invocation`. Each agent keeps its description, model and tool
list. Anything else in the frontmatter, such as tool grants or hooks, is not
published.

## Code-aware publishing

When the project has indexing enabled, every published skill and agent also
gets a short section telling it to find code through the ConMesh index
(`conmesh_find_symbol`, `conmesh_find_reference_candidates`,
`conmesh_search_codebase`) before reading files broadly. Agents that limit
their tools are given the ConMesh index tools as well, so they can call them.

With indexing on, ConMesh also publishes its own research agent and skill. An
installed package that would be published under the same name, such as the
predefined repository research agent, is covered by them, and sync lists it as
skipped.

## Third-party content

The Caveman, Matt Pocock and ECC packs are copies of exact upstream commits.
Each package folder keeps the upstream licence files and a `CONMESH-CHANGES.md`
that names the source commit and lists what ConMesh changed. Their authors
maintain the content and are responsible for it; see
[Third-party agents and skills](README.md#third-party-agents-and-skills).

Maintainers refresh them with `go run ./tools/catalog` after reviewing the new
upstream commit; see [catalog/README.md](catalog/README.md).
