package cli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackListShowAndAdd(t *testing.T) {
	dir := physicalTemp(t)
	global := filepath.Join(dir, "global")
	var stdout, stderr bytes.Buffer
	call := func(args ...string) int {
		stdout.Reset()
		stderr.Reset()
		return run(args, &stdout, &stderr, "test", dir, global)
	}
	if code := call("pack", "list"); code != 0 {
		t.Fatal(code, stderr.String())
	}
	for _, text := range []string{"Token saving", "caveman", "mattpocock-skills", "Engineering workflow", "third party", "conmesh pack add <pack> --trust"} {
		if !strings.Contains(stdout.String(), text) {
			t.Fatalf("pack list missing %q:\n%s", text, stdout.String())
		}
	}
	if code := call("pack", "show", "caveman"); code != 0 || !strings.Contains(stdout.String(), "cavecrew-investigator") || !strings.Contains(stdout.String(), "responsible") {
		t.Fatal("pack show", code, stdout.String(), stderr.String())
	}
	if code := call("pack", "show", "no-such-pack"); code != 1 || !strings.Contains(stderr.String(), "unknown pack") {
		t.Fatal("unknown pack", code, stderr.String())
	}
	if code := call("pack", "add", "caveman"); code != 1 || !strings.Contains(stderr.String(), "--trust") {
		t.Fatal("pack add without --trust must refuse", code, stderr.String())
	}
	if code := call("pack", "add", "caveman", "--global", "--trust"); code != 0 {
		t.Fatal("pack add", code, stderr.String())
	}
	if strings.Count(stdout.String(), "Installed") != 15 {
		t.Fatalf("pack add output:\n%s", stdout.String())
	}
	if code := call("pack", "add", "caveman", "--global", "--trust"); code != 0 || strings.Count(stdout.String(), "Unchanged") != 15 {
		t.Fatal("second pack add should change nothing", code, stdout.String(), stderr.String())
	}
	// One skill brings the skills it calls by name.
	if code := call("skill", "add", "builtin:grill-with-docs", "--global", "--trust"); code != 0 {
		t.Fatal("skill add", code, stderr.String())
	}
	for _, id := range []string{"grill-with-docs", "grilling", "domain-modeling"} {
		if !strings.Contains(stdout.String(), "Installed skill "+id+" ") {
			t.Fatalf("missing %s:\n%s", id, stdout.String())
		}
	}
	if code := call("agent", "add", "builtin:grilling", "--global", "--trust"); code != 1 || !strings.Contains(stderr.String(), "use conmesh skill add") {
		t.Fatal("wrong kind must be explained", code, stderr.String())
	}
	if code := call("skill", "list", "--global", "--json"); code != 0 {
		t.Fatal(code, stderr.String())
	}
	var items []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &items); err != nil || len(items) != 15 {
		t.Fatalf("installed skills: %d, %v", len(items), err)
	}
}
