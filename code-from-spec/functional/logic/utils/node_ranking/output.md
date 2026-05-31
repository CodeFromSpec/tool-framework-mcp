<!-- code-from-spec: ROOT/functional/logic/utils/node_ranking@V62VETHY8wGS4QRnPX4zYjlQ1hQ -->

## Records

```
record NodeRankInput
  logical_name: string
  frontmatter: Frontmatter

record NodeRankEntry
  logical_name: string
  rank: integer
```

## Function

```
function NodeRankCompute(entries: list of NodeRankInput) -> (ranked: list of NodeRankEntry, cycles: list of string)
  errors:
    - UnresolvableReference: a depends_on or input
      target cannot be resolved.
```

### Step 1 — Build entry map

1. Create an empty entry map keyed by logical name.
   Each map value tracks:
   - logical_name: string
   - rank: integer (initial value 0)
   - dependencies: list of strings (logical names)

2. For each item in entries:
   a. Add a spec node entry to the map keyed by item.logical_name,
      with an empty dependencies list and rank 0.
   b. For each output in item.frontmatter.outputs:
      - Derive the artifact logical name:
        - Take item.logical_name, strip the "ROOT/" prefix.
        - Prepend "ARTIFACT/".
        - Append "(<id>)" where <id> is output.id.
        - Example: node "ROOT/a/b" with output id "foo" → "ARTIFACT/a/b(foo)".
      - Add an artifact entry to the map keyed by that artifact logical name,
        with an empty dependencies list and rank 0.

### Step 2 — Build dependency edges

3. For each spec node entry in the map:
   a. If the logical name is "ROOT", skip (no dependencies — special case).
   b. Otherwise, compute the parent using LogicalNameGetParent on the logical name.
      Add the parent as a dependency.
   c. For each ref in the original frontmatter.depends_on:
      - If ref starts with "ARTIFACT/", use it as-is as the dependency key.
      - If ref starts with "ROOT/", apply LogicalNameStripQualifier to get the
        bare logical name, and use that as the dependency key.
      - Add the dependency key to this entry's dependencies list.
   d. If frontmatter.input is non-empty, add it as a dependency (it is an
      "ARTIFACT/" reference, used as-is).

4. For each artifact entry in the map:
   a. Determine the generating node's logical name:
      - Apply LogicalNameGetArtifactGenerator to the artifact's logical name
        to get the "ROOT/" generator name.
   b. Add that generator logical name as a dependency.

5. For each entry in the map, for each dependency in its dependencies list:
   - If the dependency key is not present in the entry map,
     raise error "UnresolvableReference".

### Step 3 — Initialize ranks

6. Set rank of "ROOT" to 0 (fixed, never updated).
7. Set rank of all other entries to 0 as the starting value.

### Step 4 — Iterate and detect cycles

8. Let N = total number of entries in the map.

9. Repeat up to N times (loop index from 1 to N):
   a. Set changed = false.
   b. For each entry in the map (excluding "ROOT"):
      - Compute candidate_rank = 1 + max(rank of each entry in its dependencies list).
      - If candidate_rank > entry.rank:
        - Update entry.rank = candidate_rank.
        - Set changed = true.
        - Record this entry's logical name as "changed in this pass".
   c. If changed is false, stop iterating (converged, no cycles).

10. If the loop completed all N passes and the last pass still had changes:
    - The entries whose ranks changed in the final pass are the cycle participants.
    - Collect their logical names into the cycles list.
    - Proceed to output with whatever ranks were computed.
    If the loop stopped early (converged), cycles list is empty.

### Step 5 — Output

11. Collect all entries from the map as NodeRankEntry records (logical_name + rank).

12. Sort the list:
    - Primary: rank ascending.
    - Secondary: logical_name ascending (lexicographic).

13. Return:
    - ranked: the sorted list of NodeRankEntry.
    - cycles: list of logical names identified in step 10 (empty if no cycles).

### Contracts

- All entries — both spec nodes and artifacts — are included in the output.
- Cycle detection is a by-product of the iterative relaxation; no separate
  graph traversal is performed.
- The cycles list surfaces entries involved in non-convergence to guide
  diagnosis; it is not guaranteed to enumerate every member of every cycle.
```
