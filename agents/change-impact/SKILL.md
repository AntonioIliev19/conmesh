---
name: change-impact
description: Before changing, renaming or deleting a function, type, file or config key, list everything that depends on it through the ConMesh code index, so the change is planned with its full blast radius. Use before refactors, signature changes and removals.
---
# Change impact with the ConMesh index

Know what breaks before you edit.

1. Find the definition with `conmesh_find_symbol`.
2. List uses with `conmesh_find_reference_candidates`. For config keys, routes and strings, also run `conmesh_search_codebase` with the literal value.
3. Open each candidate only around the hit and sort it: caller, test, documentation, or false match.
4. Report before editing:
   - `path:line` of each real dependent and what it needs (signature, behaviour, name).
   - Tests that cover the change, and dependents with no test.
   - The order of edits that keeps the build working.
5. After the change, search again for the old name to confirm nothing still uses it.

Reference results are lexical candidates with no type resolution: dynamic calls, reflection and generated code can be missed, so say which areas you could not confirm. If the CONMESH tools are unavailable, use `conmesh references <name>` from the project root or normal search.
