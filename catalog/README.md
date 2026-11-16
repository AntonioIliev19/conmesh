# Skill and agent catalog

[catalog.json](catalog.json) groups every package ConMesh ships into categories
and packs. `conmesh pack list`, `pack show`, `pack add` and the menu's
**Install** screen read it from the binary.

- **Third-party packs** have a `url` and an exact `commit`. Their packages live
  in [packages/](packages), copied from that commit by `tools/catalog`.
- **ConMesh's own packs** have no `url`. Their packages live in
  [../agents](../agents).

Every predefined package must belong to exactly one pack; a test checks this.

## Refresh or add third-party content

1. Read the upstream change you are about to take: the skill and agent text,
   the licence, and anything the files tell an agent to run.
2. Update the pack's `commit` (a full 40-character SHA) and `version`, or add a
   pack or member in `catalog.json`:
   - `path` is the skill folder or agent file in the upstream repository.
   - `notices` lists the licence and notice files to copy into each package.
   - `distinct_names` lists one-word member names that are never ordinary
     words, so every quoted mention of them is renamed.
   - A member's `requires` replaces the detected dependencies, and
     `upstream_name` gives a member a different ConMesh ID than its upstream
     name.
3. Regenerate the packages and review the diff:

   ```text
   go run ./tools/catalog
   git diff -- catalog/packages
   go test ./internal/packages ./internal/harness ./internal/cli
   ```

`go run ./tools/catalog -check` fails when `packages/` no longer matches what
the pinned commits produce.

## What the tool changes

For each member, the tool:

- keeps only the frontmatter keys ConMesh publishes: for skills `name`,
  `description`, `argument-hint` and `disable-model-invocation`; for agents
  `name`, `description`, `model` and `tools`;
- renames references to sibling skills and agents (`"tdd"`, `` `cavecrew-builder` ``,
  `/grill-me`) to the names ConMesh publishes (`conmesh-tdd` and so on), and
  records them as `requires` so installing one package also installs the ones
  it calls;
- leaves out upstream-only files (README, tests, package metadata, Codex
  plugin metadata);
- adds `conmesh-package.json`, `<id>.meta.json`, the licence files, and a
  `CONMESH-CHANGES.md` that names the source commit and lists these changes.

The upstream authors maintain the content and are responsible for it.
