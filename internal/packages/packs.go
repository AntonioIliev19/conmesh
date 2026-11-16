package packages

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/AntonioIliev19/conmesh/internal/security"
)

// Pack is a group of predefined packages that work together, such as one
// upstream author's skills. Packs belong to a category for display.
type Pack struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Author        string `json:"author"`
	URL           string `json:"url,omitempty"`
	License       string `json:"license"`
	Category      string `json:"category"`
	CategoryTitle string `json:"category_title"`
	ThirdParty    bool   `json:"third_party"`
	// Recommended packs are listed first and marked as recommended.
	Recommended bool           `json:"recommended,omitempty"`
	Members     []CatalogEntry `json:"members"`
}

// Packs lists the catalog's packs in catalog order, each with its members.
// Every predefined package belongs to exactly one pack.
func Packs() ([]Pack, error) {
	raw, err := catalogIndex()
	if err != nil {
		return nil, errors.New("predefined package catalog unavailable")
	}
	var index struct {
		SchemaVersion int `json:"schema_version"`
		Categories    []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Packs []struct {
				ID          string `json:"id"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Author      string `json:"author"`
				URL         string `json:"url"`
				License     string `json:"license"`
				Recommended bool   `json:"recommended"`
				Members     []struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"members"`
			} `json:"packs"`
		} `json:"categories"`
	}
	if json.Unmarshal(raw, &index) != nil || index.SchemaVersion != 1 {
		return nil, errors.New("invalid predefined package catalog")
	}
	all, err := Catalog("")
	if err != nil {
		return nil, err
	}
	byID := map[string]CatalogEntry{}
	for _, e := range all {
		byID[e.ID] = e
	}
	seen := map[string]bool{}
	packIDs := map[string]bool{}
	var out []Pack
	for _, c := range index.Categories {
		for _, p := range c.Packs {
			if security.ValidateID(p.ID) != nil || packIDs[p.ID] || len(p.Members) == 0 {
				return nil, errors.New("invalid predefined package catalog")
			}
			packIDs[p.ID] = true
			pack := Pack{ID: p.ID, Title: displayText(p.Title, 64), Description: displayText(p.Description, 200), Author: displayText(p.Author, 64),
				URL: p.URL, License: displayText(p.License, 64), Category: c.ID, CategoryTitle: displayText(c.Title, 64), ThirdParty: p.URL != "", Recommended: p.Recommended}
			for _, m := range p.Members {
				e, ok := byID[m.ID]
				if !ok || e.Type != m.Type || seen[m.ID] {
					return nil, fmt.Errorf("predefined package catalog lists %q incorrectly", m.ID)
				}
				seen[m.ID] = true
				e.Pack = p.ID
				pack.Members = append(pack.Members, e)
			}
			out = append(out, pack)
		}
	}
	return out, nil
}

// PackByID returns one pack.
func PackByID(id string) (Pack, error) {
	packs, err := Packs()
	if err != nil {
		return Pack{}, err
	}
	for _, p := range packs {
		if p.ID == id {
			return p, nil
		}
	}
	return Pack{}, errors.New("unknown pack; run conmesh pack list")
}

// WithRequires returns the predefined packages named by ids followed by every
// predefined package they require, directly or indirectly, each once.
func WithRequires(ids []string) ([]CatalogEntry, error) {
	all, err := Catalog("")
	if err != nil {
		return nil, err
	}
	byID := map[string]CatalogEntry{}
	for _, e := range all {
		byID[e.ID] = e
	}
	var out []CatalogEntry
	added := map[string]bool{}
	queue := append([]string{}, ids...)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if added[id] {
			continue
		}
		e, ok := byID[id]
		if !ok {
			return nil, errors.New("unknown predefined package")
		}
		added[id] = true
		out = append(out, e)
		queue = append(queue, e.Requires...)
	}
	return out, nil
}
