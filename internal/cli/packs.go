package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AntonioIliev19/conmesh/internal/packages"
)

// packCommand runs `conmesh pack list|show|add`. A pack installs a group of
// predefined skills and agents that work together, in one transaction.
func packCommand(ctx context.Context, m *packages.Manager, args []string, project string, report **stepReport) (string, error) {
	if len(args) < 2 {
		return "", errUsage
	}
	switch args[1] {
	case "list":
		opts, err := parseOptions(args[2:], map[string]bool{"json": false})
		if err != nil || len(opts.positionals) != 0 {
			return "", errUsage
		}
		packs, err := packages.Packs()
		if err != nil {
			return "", err
		}
		if opts.has("json") {
			data, _ := json.MarshalIndent(packs, "", "  ")
			return string(data) + "\n", nil
		}
		return packListing(packs), nil
	case "show":
		opts, err := parseOptions(args[2:], map[string]bool{"json": false})
		if err != nil || len(opts.positionals) != 1 {
			return "", errUsage
		}
		p, err := packages.PackByID(opts.positionals[0])
		if err != nil {
			return "", err
		}
		if opts.has("json") {
			data, _ := json.MarshalIndent(p, "", "  ")
			return string(data) + "\n", nil
		}
		return packDetail(p), nil
	case "add":
		opts, err := parseOptions(args[2:], map[string]bool{"global": false, "project": false, "trust": false})
		if err != nil || len(opts.positionals) != 1 {
			return "", errUsage
		}
		scope, err := opts.scope()
		if err != nil {
			return "", err
		}
		p, err := packages.PackByID(opts.positionals[0])
		if err != nil {
			return "", err
		}
		ids := make([]string, 0, len(p.Members))
		for _, e := range p.Members {
			ids = append(ids, e.ID)
		}
		return installCatalog(ctx, m, ids, scope, opts.has("trust"), project, report)
	}
	return "", errUsage
}

// installCatalog installs predefined packages and everything they require in
// one transaction.
func installCatalog(ctx context.Context, m *packages.Manager, ids []string, scope string, trusted bool, project string, report **stepReport) (string, error) {
	entries, err := packages.WithRequires(ids)
	if err != nil {
		return "", err
	}
	requests := make([]packages.Request, 0, len(entries))
	for _, e := range entries {
		requests = append(requests, packages.Request{Kind: e.Type, Source: e.Source})
	}
	results, err := m.AddAll(ctx, requests, scope, trusted)
	if err != nil {
		return "", err
	}
	requested := map[string]bool{}
	for _, id := range ids {
		requested[id] = true
	}
	var out strings.Builder
	r := &stepReport{}
	for _, result := range results {
		e := result.Entry
		fmt.Fprint(&out, formatResult(result))
		label := "Installed"
		if result.Unchanged {
			label = "Unchanged"
		} else {
			r.changed = append(r.changed, fmt.Sprintf("+ %s %s %s (%s)", e.Type, e.ID, e.Version, e.Scope))
		}
		detail := fmt.Sprintf("version %s, %s scope", e.Version, e.Scope)
		if !requested[e.ID] {
			detail += ", required by another package"
		}
		r.steps = append(r.steps, reportStep{label, e.Type + " " + e.ID, detail})
	}
	if len(r.changed) > 0 {
		if project != "" {
			r.next = syncNext()
		} else {
			r.next = []nextCommand{{"conmesh sync --trust", "run in each project to publish them to its AI tools"}}
		}
	}
	*report = r
	return out.String(), nil
}

func packCounts(p packages.Pack) string {
	skills, agents := 0, 0
	for _, e := range p.Members {
		if e.Type == "skill" {
			skills++
		} else {
			agents++
		}
	}
	var parts []string
	if skills > 0 {
		parts = append(parts, plural(skills, "skill"))
	}
	if agents > 0 {
		parts = append(parts, plural(agents, "agent"))
	}
	return strings.Join(parts, ", ")
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}

// packOrigin names who maintains a pack's content.
func packOrigin(p packages.Pack) string {
	if p.ThirdParty {
		return fmt.Sprintf("by %s, %s, third party", p.Author, p.License)
	}
	return "by ConMesh"
}

func packListing(packs []packages.Pack) string {
	var b strings.Builder
	category := ""
	for _, p := range packs {
		if p.CategoryTitle != category {
			if category != "" {
				b.WriteString("\n")
			}
			category = p.CategoryTitle
			fmt.Fprintln(&b, category)
		}
		mark := ""
		if p.Recommended {
			mark = "  [recommended]"
		}
		fmt.Fprintf(&b, "  %-24s %s (%s; %s)%s\n", p.ID, p.Title, packCounts(p), packOrigin(p), mark)
		fmt.Fprintf(&b, "  %-24s %s\n", "", p.Description)
	}
	b.WriteString("\nInstall a pack:  conmesh pack add <pack> --trust\n")
	b.WriteString("See what is in it:  conmesh pack show <pack>\n")
	return b.String()
}

func packDetail(p packages.Pack) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s (%s)\n%s\n", p.Title, packOrigin(p), p.Description)
	if p.URL != "" {
		fmt.Fprintf(&b, "Source: %s\n", p.URL)
	}
	b.WriteString("\n")
	width := 0
	for _, e := range p.Members {
		width = max(width, len(e.ID))
	}
	for _, e := range p.Members {
		fmt.Fprintf(&b, "  %-5s  %-*s  %s\n", e.Type, width, e.ID, e.Description)
	}
	fmt.Fprintf(&b, "\nInstall all:  conmesh pack add %s --trust\n", p.ID)
	fmt.Fprintf(&b, "Install one:  conmesh skill add builtin:<id> --trust  (or agent add)\n")
	if p.ThirdParty {
		b.WriteString("Third-party content is maintained by its authors, who are responsible for it. Review it before use.\n")
	}
	return b.String()
}
