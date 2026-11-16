// Package catalog embeds ConMesh's catalog of installable skills and agents:
// catalog.json groups packages into categories and packs, and packages/ holds
// third-party packages copied from pinned upstream commits by tools/catalog.
// ConMesh's own packages live in the agents directory.
package catalog

import "embed"

// FS holds catalog.json and one directory per vendored package under packages/.
//
//go:embed catalog.json packages
var FS embed.FS
