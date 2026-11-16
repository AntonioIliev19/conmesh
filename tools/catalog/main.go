// Command catalog vendors the third-party skills and agents listed in
// catalog/catalog.json into catalog/packages, so ConMesh can install them
// without network access. Each pack names an upstream repository and an exact
// commit; this tool fetches that commit and writes one ConMesh package per
// member with the upstream licence files and a note of what was changed.
//
//	go run ./tools/catalog            # regenerate catalog/packages
//	go run ./tools/catalog -check     # fail if catalog/packages is out of date
//
// Packs without a "url" are ConMesh's own packages in agents/ and are skipped.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type member struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Path string `json:"path"`
	// Title replaces the display title made from the ID, for names like "TDD".
	Title string `json:"title"`
	// Upstream is the skill or agent's own name when the ConMesh ID differs,
	// to avoid a clash with a name ConMesh publishes itself.
	Upstream string `json:"upstream_name"`
	// Requires replaces the detected dependencies when a member only mentions
	// siblings it can work without, such as a setup skill listing what it configures.
	Requires *[]string `json:"requires,omitempty"`
}

type pack struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	URL         string   `json:"url"`
	License     string   `json:"license"`
	Commit      string   `json:"commit"`
	Version     string   `json:"version"`
	Notices     []string `json:"notices"`
	// Distinct lists one-word member names that are never ordinary words, so
	// every quoted mention of them refers to the skill or agent.
	Distinct []string `json:"distinct_names"`
	Members  []member `json:"members"`
}

type catalog struct {
	SchemaVersion int `json:"schema_version"`
	Categories    []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Packs []pack `json:"packs"`
	} `json:"categories"`
}

// Files that only matter to the upstream repository, never to an installed skill.
var skipped = map[string]bool{"README.md": true, "CLAUDE.md": true, "package.json": true, "agents/openai.yaml": true}

// Frontmatter kept per kind; everything else (tool grants, hooks, model
// choices for skills, metadata) is dropped and listed in CONMESH-CHANGES.md.
var keep = map[string][]string{
	"skill": {"name", "description", "argument-hint", "disable-model-invocation"},
	"agent": {"name", "description", "model", "tools"},
}

func main() {
	root := flag.String("root", ".", "repository root")
	check := flag.Bool("check", false, "compare instead of writing")
	flag.Parse()
	if err := run(*root, *check); err != nil {
		fmt.Fprintln(os.Stderr, "catalog:", err)
		os.Exit(1)
	}
}

func run(root string, check bool) error {
	raw, err := os.ReadFile(filepath.Join(root, "catalog", "catalog.json"))
	if err != nil {
		return err
	}
	var c catalog
	if err := json.Unmarshal(raw, &c); err != nil {
		return err
	}
	out := map[string][]byte{} // package-relative path -> content
	for _, cat := range c.Categories {
		for _, p := range cat.Packs {
			if p.URL == "" {
				continue
			}
			files, err := vendor(p)
			if err != nil {
				return fmt.Errorf("%s: %w", p.ID, err)
			}
			for name, data := range files {
				if _, dup := out[name]; dup {
					return fmt.Errorf("%s: duplicate package file %s", p.ID, name)
				}
				out[name] = data
			}
		}
	}
	dir := filepath.Join(root, "catalog", "packages")
	if check {
		return compare(dir, out)
	}
	// Empty the folder rather than removing it, so a shell or editor that has
	// it open does not block regeneration on Windows.
	old, err := os.ReadDir(dir)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	for _, e := range old {
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	for name, data := range out {
		target := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
	}
	fmt.Printf("wrote %d files to %s\n", len(out), dir)
	return nil
}

func compare(dir string, want map[string][]byte) error {
	have := map[string]bool{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		have[filepath.ToSlash(rel)] = true
		return nil
	})
	if err != nil {
		return err
	}
	var diffs []string
	for name, data := range want {
		got, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
		if err != nil || !bytes.Equal(got, data) {
			diffs = append(diffs, name)
		}
		delete(have, name)
	}
	for name := range have {
		diffs = append(diffs, name+" (not generated)")
	}
	if len(diffs) > 0 {
		sort.Strings(diffs)
		return fmt.Errorf("catalog/packages is out of date: %s", strings.Join(diffs, ", "))
	}
	fmt.Println("catalog/packages is up to date")
	return nil
}

// vendor fetches one pack's commit and returns its packages' files.
func vendor(p pack) (map[string][]byte, error) {
	if !regexp.MustCompile(`^[0-9a-f]{40}$`).MatchString(p.Commit) {
		return nil, errors.New("commit must be a full 40-character SHA")
	}
	tmp, err := os.MkdirTemp("", "conmesh-catalog-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)
	paths := append([]string{}, p.Notices...)
	for _, m := range p.Members {
		paths = append(paths, m.Path)
	}
	sparse := make([]string, 0, len(paths))
	for _, s := range paths {
		sparse = append(sparse, "/"+s)
	}
	steps := [][]string{
		{"init", "-q"},
		{"remote", "add", "origin", p.URL},
		{"fetch", "-q", "--depth=1", "--filter=blob:none", "origin", p.Commit},
		append([]string{"sparse-checkout", "set", "--no-cone"}, sparse...),
		{"-c", "advice.detachedHead=false", "checkout", "-q", "FETCH_HEAD"},
	}
	for _, args := range steps {
		cmd := exec.Command("git", args...)
		cmd.Dir = tmp
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("git %s: %v: %s", args[0], err, output)
		}
	}
	ids, distinct := map[string]string{}, map[string]bool{}
	for _, m := range p.Members {
		ids[m.upstream()] = m.ID
	}
	for _, id := range p.Distinct {
		distinct[id] = true
	}
	out := map[string][]byte{}
	for _, m := range p.Members {
		files, dropped, err := readMember(tmp, m)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", m.ID, err)
		}
		requires := map[string]bool{}
		for name, data := range files {
			if strings.HasSuffix(name, ".md") {
				files[name] = rename(data, ids, distinct, m.ID, requires)
			}
		}
		for _, notice := range p.Notices {
			data, err := os.ReadFile(filepath.Join(tmp, filepath.FromSlash(notice)))
			if err != nil {
				return nil, fmt.Errorf("notice %s: %w", notice, err)
			}
			files[notice] = data
		}
		entry := "SKILL.md"
		if m.Type == "agent" {
			entry = "AGENT.md"
		}
		title, description := describe(files[entry], m.ID)
		if m.Title != "" {
			title = m.Title
		}
		files["conmesh-package.json"] = jsonBytes(map[string]any{"schema_version": 1, "id": m.ID, "type": m.Type, "version": p.Version, "license": p.License})
		needs := []string{}
		for id := range requires {
			needs = append(needs, id)
		}
		if m.Requires != nil {
			needs = append([]string{}, *m.Requires...)
		}
		sort.Strings(needs)
		files[m.ID+".meta.json"] = jsonBytes(map[string]any{
			"index":    map[string]any{"title": title, "description": description, "type": m.Type},
			"pack":     p.ID,
			"requires": needs,
			"upstream": map[string]any{"repository": p.URL, "commit": p.Commit, "path": m.Path, "author": p.Author, "license": p.License},
		})
		files["CONMESH-CHANGES.md"] = changes(p, m, dropped)
		for name, data := range files {
			out[m.ID+"/"+name] = data
		}
	}
	return out, nil
}

func (m member) upstream() string {
	if m.Upstream != "" {
		return m.Upstream
	}
	return m.ID
}

// readMember returns a member's files with its entrypoint normalised.
func readMember(checkout string, m member) (map[string][]byte, []string, error) {
	src := filepath.Join(checkout, filepath.FromSlash(m.Path))
	files := map[string][]byte{}
	if m.Type == "agent" {
		data, err := os.ReadFile(src)
		if err != nil {
			return nil, nil, err
		}
		files["AGENT.md"] = data
	} else {
		err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(src, p)
			rel = filepath.ToSlash(rel)
			if skipped[rel] || strings.HasPrefix(rel, "tests/") {
				return nil
			}
			data, err := os.ReadFile(p)
			files[rel] = data
			return err
		})
		if err != nil {
			return nil, nil, err
		}
	}
	entry := "SKILL.md"
	if m.Type == "agent" {
		entry = "AGENT.md"
	}
	data, ok := files[entry]
	if !ok {
		return nil, nil, fmt.Errorf("%s not found", entry)
	}
	normalised, dropped, err := frontmatter(data, m)
	if err != nil {
		return nil, nil, err
	}
	files[entry] = normalised
	return files, dropped, nil
}

// frontmatter keeps only the portable keys, puts them in a fixed order and
// folds the description onto one line.
func frontmatter(data []byte, m member) ([]byte, []string, error) {
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return nil, nil, errors.New("entrypoint has no frontmatter")
	}
	front, body, ok := strings.Cut(text[4:], "\n---\n")
	if !ok {
		return nil, nil, errors.New("unterminated frontmatter")
	}
	var meta map[string]any
	if err := yaml.Unmarshal([]byte(front), &meta); err != nil {
		return nil, nil, err
	}
	if meta["name"] != m.upstream() {
		return nil, nil, fmt.Errorf("frontmatter name %v does not match id", meta["name"])
	}
	meta["name"] = m.ID
	if d, ok := meta["description"].(string); ok {
		meta["description"] = strings.Join(strings.Fields(d), " ")
	}
	var b strings.Builder
	b.WriteString("---\n")
	for _, k := range keep[m.Type] {
		v, ok := meta[k]
		if !ok {
			continue
		}
		line, err := yaml.Marshal(map[string]any{k: v})
		if err != nil {
			return nil, nil, err
		}
		b.Write(line)
		delete(meta, k)
	}
	b.WriteString("---\n")
	var dropped []string
	for k := range meta {
		dropped = append(dropped, k)
	}
	sort.Strings(dropped)
	return []byte(b.String() + strings.TrimLeft(body, "\n")), dropped, nil
}

// rename points references to sibling skills and agents at the names ConMesh
// publishes them under (conmesh-<id>) and records which siblings are needed.
func rename(data []byte, ids map[string]string, distinct map[string]bool, self string, requires map[string]bool) []byte {
	// The frontmatter is included, since a description can name sibling
	// commands. Its unquoted "name:" value is never matched; ConMesh renames it
	// when publishing.
	text := string(data)
	names := make([]string, 0, len(ids))
	for id := range ids {
		names = append(names, regexp.QuoteMeta(id))
	}
	sort.Slice(names, func(i, j int) bool { return len(names[i]) > len(names[j]) })
	alt := strings.Join(names, "|")
	quoted := regexp.MustCompile("([`\"])(/?)(" + alt + ")([`\"])")
	slash := regexp.MustCompile(`(^|[\s(\[` + "`" + `])/(` + alt + `)\b([^/\w-]|$)`)
	// A one-word name such as "pr" or "research" is also an ordinary word or a
	// label, so it is renamed only where it plainly names a skill or agent:
	// after "Skill tool" in the same sentence, or followed by "skill"/"agent".
	toolBefore := regexp.MustCompile(`Skill tool[^.]*$`)
	kindAfter := regexp.MustCompile(`^\s+(skill|agent|subagent)s?\b`)
	note := func(id string) {
		if id != self {
			requires[id] = true
		}
	}
	lines := strings.SplitAfter(text, "\n")
	for i, line := range lines {
		var b strings.Builder
		last := 0
		for _, at := range quoted.FindAllStringSubmatchIndex(line, -1) {
			open, slashed, id, closing := line[at[2]:at[3]], line[at[4]:at[5]], line[at[6]:at[7]], line[at[8]:at[9]]
			plain := slashed == "" && !strings.Contains(id, "-") && !distinct[id]
			if open != closing || plain && !toolBefore.MatchString(line[:at[0]]) && !kindAfter.MatchString(line[at[1]:]) {
				continue
			}
			note(ids[id])
			b.WriteString(line[last:at[0]] + open + slashed + "conmesh-" + ids[id] + closing)
			last = at[1]
		}
		line = b.String() + line[last:]
		lines[i] = slash.ReplaceAllStringFunc(line, func(s string) string {
			g := slash.FindStringSubmatch(s)
			note(ids[g[2]])
			return g[1] + "/conmesh-" + ids[g[2]] + g[3]
		})
	}
	return []byte(strings.Join(lines, ""))
}

func describe(entry []byte, id string) (string, string) {
	var meta struct {
		Description string `yaml:"description"`
	}
	text := strings.ReplaceAll(string(entry), "\r\n", "\n")
	if front, _, ok := strings.Cut(strings.TrimPrefix(text, "---\n"), "\n---\n"); ok {
		_ = yaml.Unmarshal([]byte(front), &meta)
	}
	title := strings.ReplaceAll(id, "-", " ")
	title = strings.ToUpper(title[:1]) + title[1:]
	description := meta.Description
	if r := []rune(description); len(r) > 200 {
		description = string(r[:199]) + "…"
	}
	return title, description
}

func changes(p pack, m member, dropped []string) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "# Source and changes\n\n")
	fmt.Fprintf(&b, "This package is a copy of `%s` from %s at commit `%s`, by %s, licensed under %s (see the licence files in this folder).\n\n", m.Path, p.URL, p.Commit, p.Author, p.License)
	fmt.Fprintf(&b, "ConMesh changed it as follows:\n\n")
	fmt.Fprintf(&b, "- Added `conmesh-package.json`, `%s.meta.json` and this file.\n", m.ID)
	fmt.Fprintf(&b, "- References to other %s skills and agents use the names ConMesh publishes them under, which start with `conmesh-`.\n", p.Title)
	if m.Upstream != "" {
		fmt.Fprintf(&b, "- Renamed from `%s` to `%s`, because ConMesh already publishes a `conmesh-%s` of its own.\n", m.Upstream, m.ID, m.Upstream)
	}
	if len(dropped) > 0 {
		fmt.Fprintf(&b, "- Removed frontmatter keys ConMesh does not publish: %s.\n", strings.Join(dropped, ", "))
	}
	fmt.Fprintf(&b, "- Left out files used only by the upstream repository (README, tests, package metadata, Codex plugin metadata).\n")
	fmt.Fprintf(&b, "\nThe upstream authors maintain the content. Report problems with it to %s.\n", p.URL)
	return []byte(b.String())
}

func jsonBytes(v any) []byte {
	data, _ := json.MarshalIndent(v, "", "  ")
	return append(data, '\n')
}
