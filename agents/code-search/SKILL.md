---
name: code-search
description: Find where something is defined, used or implemented in this repository through the ConMesh code index instead of reading many files. Use for "where is X", "what calls Y", "how does Z work", or before editing unfamiliar code.
---
# Code search with the ConMesh index

Answer location questions from the index, then read only what it points to.

1. Pick the narrowest tool:
   - `conmesh_find_symbol` for a known name (function, type, class, constant).
   - `conmesh_find_reference_candidates` for callers and uses of a name.
   - `conmesh_search_codebase` for a concept, error text or behaviour.
2. Read only the returned files and line ranges. Widen the search only when the results do not answer the question.
3. Reply with `path:line` locations and one short note each. Do not paste whole files.

Reference results are lexical candidates, not a complete call graph: confirm a hit in the source before relying on it. If `conmesh_index_status` shows the index is missing or old, say so and fall back to normal file search. Without the CONMESH tools, run `conmesh search "<query>"` or `conmesh symbol <name>` from the project root.
