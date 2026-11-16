package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// catalogNotices adds the licence and notice texts of the third-party skill
// and agent packs embedded from catalog/packages, once per pack.
func catalogNotices(notices *strings.Builder) error {
	raw, err := os.ReadFile(filepath.Join("catalog", "catalog.json"))
	if err != nil {
		return fmt.Errorf("read catalog: %w", err)
	}
	var c struct {
		Categories []struct {
			Packs []struct {
				Title   string   `json:"title"`
				Author  string   `json:"author"`
				URL     string   `json:"url"`
				License string   `json:"license"`
				Commit  string   `json:"commit"`
				Notices []string `json:"notices"`
				Members []struct {
					ID string `json:"id"`
				} `json:"members"`
			} `json:"packs"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(raw, &c); err != nil {
		return fmt.Errorf("parse catalog: %w", err)
	}
	for _, cat := range c.Categories {
		for _, p := range cat.Packs {
			if p.URL == "" || len(p.Members) == 0 {
				continue
			}
			for _, notice := range p.Notices {
				// Every member carries identical copies; the first one stands for the pack.
				data, err := os.ReadFile(filepath.Join("catalog", "packages", p.Members[0].ID, filepath.FromSlash(notice)))
				if err != nil {
					return fmt.Errorf("catalog notice for %s: %w", p.Title, err)
				}
				fmt.Fprintf(notices, "\n## Catalog pack %s by %s (%s, %s at %s) — %s\n\n```text\n%s\n```\n", p.Title, p.Author, p.License, p.URL, p.Commit, notice, data)
			}
		}
	}
	return nil
}
