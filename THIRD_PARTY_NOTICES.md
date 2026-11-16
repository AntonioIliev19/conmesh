# Third-party notices

ConMesh is licensed under [Apache-2.0](LICENSE). Third-party components retain their upstream licenses and notices.

The pinned Go module graph and complete collected license texts are available in:

- [Dependency inventory](results/dependencies-phase5/dependencies.json)
- [CycloneDX SBOM](results/dependencies-phase5/sbom.cdx.json)
- [Full module dependency graph](results/dependencies-phase5/go-module-graph.txt)
- [Preserved license and notice texts](results/dependencies-phase5/THIRD_PARTY_LICENSE_TEXTS.md)

The inventory includes the entire module graph, including build/test-only modules;
it is broader than the set linked into an individual binary. Per-artifact build
information records the linked modules separately. Module `h1` values are Go module
content checksums, not hashes of native release artifacts.

Native components include Tree-sitter and TypeScript/C# grammars, modernc SQLite
and its libc support. Nested native notices are collected along with top-level
module licenses. The preserved text also includes Go toolchain LICENSE/PATENTS
and its vendored library notices. Review every artifact's inventory after a
dependency or toolchain change; generated collection assists that review and does
not settle license compatibility on its own.

Regenerate in a **new** directory with:

```text
go run ./tools/release --inventory-only --out path-to-new-private-inventory
```

## Skill and agent catalog

The ConMesh binary embeds a catalog of third-party skills and agents, copied
from pinned upstream commits into [catalog/packages](catalog/packages) by
`go run ./tools/catalog`. Each package folder keeps the upstream licence files
and a `CONMESH-CHANGES.md` that names the source commit and lists ConMesh's
changes. Release notices include each pack's licence texts.

| Pack | Author | Licence | Source |
|---|---|---|---|
| Caveman | Julius Brussee | Apache-2.0 (with NOTICE and the MIT text for earlier contributions) | [juliusbrussee/caveman](https://github.com/juliusbrussee/caveman) |
| Matt Pocock skills | Matt Pocock | MIT | [mattpocock/skills](https://github.com/mattpocock/skills) |
| ECC agents and skills | Affaan Mustafa | MIT | [affaan-m/ecc](https://github.com/affaan-m/ecc) |

Their authors maintain this content and are responsible for it; see
[Third-party agents and skills](README.md#third-party-agents-and-skills).

No embedding model weights are bundled. A user-supplied model or package has
its own separate license and provenance requirements.
