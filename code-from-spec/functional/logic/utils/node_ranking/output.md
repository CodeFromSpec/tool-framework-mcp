<!-- code-from-spec: ROOT/functional/logic/utils/node_ranking@SRPv1KJ8gMJsLrYWDKfAijz_SF8 -->

## Records

```
record NodeRankInput
  logical_name: string
  frontmatter: frontmatter.Frontmatter

record NodeRankEntry
  logical_name: string
  rank: integer
```

## Internal Records

```
record EntryState
  logical_name: string
  dependencies: list of strings
  rank: integer
```

## Functions

```
function NodeRankCompute(entries: list of NodeRankInput) -> (ranked: list of NodeRankEntry, cycles: list of string)
  errors:
    - UnresolvableReference: a depends_on or input
      target cannot be resolved.
```

### Step 1 — Build entry map

  1. Create an empty entry map (keyed by logical name, values are EntryState records).

  2. For each NodeRankInput in entries:
       a. Add a spec node EntryState to the map:
            - logical_name = NodeRankInput.logical_name
            - dependencies = [] (filled in Step 2)
            - rank = 0

       b. For each output in NodeRankInput.frontmatter.outputs:
            - Derive the artifact logical name:
                i.   Strip the "ROOT/" prefix from NodeRankInput.logical_name
                     to get the node path segment.
                     If NodeRankInput.logical_name is exactly "ROOT", the
                     path segment is empty string — artifact key becomes
                     "ARTIFACT/(id)".
                ii.  Prepend "ARTIFACT/" to that segment.
                iii. Append "(" + output.id + ")".
                     Example: node "ROOT/a/b" with output id "foo" →
                     "ARTIFACT/a/b(foo)".
            - Add an artifact EntryState to the map:
                - logical_name = artifact logical name from above
                - dependencies = [] (filled in Step 2)
                - rank = 0

### Step 2 — Build dependency edges

  3. For each spec node entry in the entry map:
       a. If the logical name is "ROOT", skip (no dependencies, handled in Step 3).

       b. Derive parent:
            - Call LogicalNameGetParent(logical_name) to get parent logical name.
            - If parent is not in the entry map,
              raise error "UnresolvableReference" for that parent.
            - Add parent logical name to this entry's dependencies.

       c. For each item in this node's frontmatter.depends_on:
            - If item starts with "ARTIFACT/":
                - Use item as-is as the lookup key.
            - Else (it is a "ROOT/" reference):
                - Call LogicalNameStripQualifier(item) to get the bare
                  logical name for lookup.
            - If the lookup key is not in the entry map,
              raise error "UnresolvableReference" for that key.
            - Add the lookup key to this entry's dependencies.

       d. If frontmatter.input is non-empty:
            - Use frontmatter.input as-is as the lookup key
              (it is an "ARTIFACT/" reference).
            - If the lookup key is not in the entry map,
              raise error "UnresolvableReference" for that key.
            - Add the lookup key to this entry's dependencies.

  4. For each artifact entry in the entry map:
       - The generating node's logical name is derived by calling
         LogicalNameGetArtifactGenerator(artifact_logical_name),
         which strips "ARTIFACT/" prefix and qualifier to return a "ROOT/" name.
       - If the generating node logical name is not in the entry map,
         raise error "UnresolvableReference" for that node.
       - Add the generating node logical name to this artifact entry's dependencies.

### Step 3 — Initialize ranks

  5. Set the rank of the "ROOT" entry to 0 (fixed, will not be updated).
     Set the rank of every other entry to 0 as an initial value.

### Step 4 — Iterate and detect cycles

  6. Let N = total number of entries in the entry map.

  7. Let changed_in_last_pass = empty list.

  8. Repeat N times (passes indexed 1 through N):
       a. Set changed_this_pass = empty list.
       b. For each entry in the entry map (excluding "ROOT"):
            i.  Compute new_rank = 1 + max(rank of each entry in this entry's dependencies).
                If this entry has no dependencies, new_rank = 1 (though only
                ROOT has no dependencies and it is excluded).
            ii. If new_rank > entry.rank:
                  - Update entry.rank = new_rank.
                  - Add entry.logical_name to changed_this_pass.
       c. Set changed_in_last_pass = changed_this_pass.
       d. If changed_this_pass is empty, stop iterating (graph has converged,
          no cycles detected).

  9. If iteration completed all N passes without converging (the loop ran to
     completion without an early stop in step 8d):
       - The entries in changed_in_last_pass are the cycle participants to report.
       - Do not raise an error — proceed to Step 5 and include cycle information
         in the return value.

     If the loop stopped early (converged), cycles = empty list.

### Step 5 — Output

  10. Collect all EntryState records from the entry map as NodeRankEntry records
      (logical_name + rank).

  11. Sort the list:
        - Primary sort: rank ascending.
        - Secondary sort: logical_name ascending (lexicographic).

  12. Return:
        - ranked = the sorted list of NodeRankEntry records.
        - cycles = the list of logical names from changed_in_last_pass
          (empty if no cycles were detected).
```
