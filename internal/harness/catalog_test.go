package harness

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AntonioIliev19/conmesh/internal/config"
	"github.com/AntonioIliev19/conmesh/internal/packages"
)

// Every package in the catalog must install and publish for Claude Code and
// Codex, with and without the code index, so a catalog refresh cannot ship a
// package that sync rejects.
func TestEveryCatalogPackagePublishes(t *testing.T) {
	packs, err := packages.Packs()
	if err != nil {
		t.Fatal(err)
	}
	var requests []packages.Request
	for _, p := range packs {
		for _, e := range p.Members {
			requests = append(requests, packages.Request{Kind: e.Type, Source: e.Source})
		}
	}
	if len(requests) < 60 {
		t.Fatalf("only %d catalog packages", len(requests))
	}
	for _, indexed := range []bool{false, true} {
		project, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		cfg := config.Default()
		cfg.Harnesses = []string{"claude", "codex"}
		cfg.Indexing.Enabled = indexed
		if _, _, err := config.Init(project, cfg); err != nil {
			t.Fatal(err)
		}
		m := &Manager{Project: project, Global: filepath.Join(project, "global"), Executable: filepath.Join(project, "conmesh.exe")}
		results, err := packages.New(m.Project, m.Global).AddAll(context.Background(), requests, "project", true)
		if err != nil || len(results) != len(requests) {
			t.Fatalf("install all: %d results, %v", len(results), err)
		}
		plan, err := m.Sync(context.Background(), false, true)
		if err != nil {
			t.Fatalf("sync indexed=%v: %v", indexed, err)
		}
		read := func(name string) string {
			data, err := os.ReadFile(filepath.Join(m.Project, filepath.FromSlash(name)))
			if err != nil {
				t.Fatal(err)
			}
			return string(data)
		}
		skill := read(".claude/skills/conmesh-caveman/SKILL.md")
		if !strings.Contains(skill, "name: conmesh-caveman") || !strings.Contains(skill, "`conmesh-ultracave`") {
			t.Fatalf("caveman skill:\n%s", skill)
		}
		if !strings.Contains(read(".claude/skills/conmesh-grill-me/SKILL.md"), "disable-model-invocation: true") {
			t.Fatal("user-invoked skill lost disable-model-invocation")
		}
		agent := read(".claude/agents/conmesh-code-reviewer.md")
		if !strings.Contains(agent, "model: sonnet") || !strings.Contains(agent, "description: Expert code review specialist") || strings.Count(agent, "\n---\n") != 1 {
			t.Fatalf("code-reviewer agent:\n%.600s", agent)
		}
		codex := read(".codex/agents/conmesh-cavecrew-investigator.toml")
		hasIndex := strings.Contains(skill, "## ConMesh code index") && strings.Contains(codex, "ConMesh code index")
		hasTools := strings.Contains(agent, "mcp__conmesh__conmesh_find_symbol")
		if hasIndex != indexed || hasTools != indexed {
			t.Fatalf("indexed=%v but index section=%v, agent index tools=%v", indexed, hasIndex, hasTools)
		}
		read(".agents/skills/conmesh-tdd/SKILL.md")
		read(".claude/skills/conmesh-source-research/SKILL.md")
		// With indexing, ConMesh's own research agent replaces the installed copy.
		research := read(".claude/agents/conmesh-repository-research.md")
		if indexed != strings.Contains(research, "execute the ConMesh research workflow") || indexed != strings.Contains(strings.Join(plan.Skipped, "\n"), "agent repository-research") {
			t.Fatalf("indexed=%v research agent:\n%.300s\nskipped: %v", indexed, research, plan.Skipped)
		}
	}
}
