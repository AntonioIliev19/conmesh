package packages

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every predefined package belongs to exactly one pack, so the menu and
// `pack list` show everything ConMesh ships.
func TestEveryPredefinedPackageIsInOnePack(t *testing.T) {
	packs, err := Packs()
	if err != nil {
		t.Fatal(err)
	}
	all, err := Catalog("")
	if err != nil {
		t.Fatal(err)
	}
	inPack := map[string]int{}
	for _, p := range packs {
		if p.CategoryTitle == "" || p.Title == "" || p.Description == "" || p.ThirdParty && (p.URL == "" || p.Author == "") {
			t.Fatalf("pack %s is missing display details: %+v", p.ID, p)
		}
		for _, e := range p.Members {
			inPack[e.ID]++
			if e.Pack != p.ID {
				t.Fatalf("%s reports pack %q, want %q", e.ID, e.Pack, p.ID)
			}
		}
	}
	for _, e := range all {
		if inPack[e.ID] != 1 {
			t.Errorf("predefined %s %s is in %d packs", e.Type, e.ID, inPack[e.ID])
		}
		for _, id := range e.Requires {
			if _, err := WithRequires([]string{id}); err != nil {
				t.Errorf("%s requires unknown %s", e.ID, id)
			}
		}
	}
	if len(all) < 60 {
		t.Fatalf("only %d predefined packages", len(all))
	}
}

func TestWithRequiresFollowsChainsOnce(t *testing.T) {
	entries, err := WithRequires([]string{"grill-with-docs", "grilling"})
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, e := range entries {
		ids = append(ids, e.ID)
	}
	if strings.Join(ids, ",") != "grill-with-docs,grilling,domain-modeling" {
		t.Fatalf("resolved %v", ids)
	}
	// Caveman and its variants name each other; the cycle must end.
	if entries, err := WithRequires([]string{"caveman"}); err != nil || len(entries) != 3 {
		t.Fatalf("caveman cycle: %d entries, %v", len(entries), err)
	}
	if _, err := WithRequires([]string{"no-such-package"}); err == nil {
		t.Fatal("unknown package accepted")
	}
}

// A pack is installed in one transaction: one conflicting member leaves the
// store exactly as it was.
func TestAddAllIsAllOrNothing(t *testing.T) {
	m := newManager(t)
	ctx := context.Background()
	// A user's own "tdd" skill from a folder occupies the ID.
	dir := t.TempDir()
	files := map[string]string{
		SourceManifest: `{"schema_version":1,"id":"tdd","type":"skill","version":"1","license":"MIT"}`,
		"LICENSE":      "MIT",
		"SKILL.md":     "---\nname: tdd\ndescription: Mine.\n---\nMy TDD.\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.Add(ctx, "skill", dir, "", "project", true); err != nil {
		t.Fatal(err)
	}
	requests := []Request{{Kind: "skill", Source: BuiltinPrefix + "grilling"}, {Kind: "skill", Source: BuiltinPrefix + "tdd"}}
	if _, err := m.AddAll(ctx, requests, "project", true); err == nil || !strings.Contains(err.Error(), "skill tdd is already installed from another source") {
		t.Fatalf("conflict not reported: %v", err)
	}
	items, err := m.List(ctx, "skill", "project")
	if err != nil || len(items) != 1 || items[0].ID != "tdd" {
		t.Fatalf("store changed after a failed pack install: %+v %v", items, err)
	}
	if _, err := m.AddAll(ctx, requests[:1], "project", false); err == nil {
		t.Fatal("untrusted batch accepted")
	}
	results, err := m.AddAll(ctx, requests[:1], "project", true)
	if err != nil || len(results) != 1 || results[0].Unchanged {
		t.Fatalf("batch add: %+v %v", results, err)
	}
}
