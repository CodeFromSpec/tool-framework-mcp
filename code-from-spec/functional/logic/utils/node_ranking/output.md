<!-- code-from-spec: ROOT/functional/logic/utils/node_ranking@6kqVyq52zQSlETyDNKaRPwedc-k -->

# NodeRankCompute

## Records

```
record NodeRankInput
  logical_name: string
  frontmatter: Frontmatter

record NodeRankEntry
  logical_name: string
  rank: integer

record InternalEntry
  logical_name: string
  dependencies: list of strings
  rank: integer
```

## Function

```
function NodeRankCompute(entries: list of NodeRankInput)
  -> (ranked: list of NodeRankEntry, cycles: list of string)
  errors:
    - "unresolvable reference": a depends_on or input target
      cannot be resolved to any entry in the entry map.
```

### Step 1 — Build entry map

1. Create an empty entry map keyed by logical name.
   Each value is an InternalEntry with an empty dependencies list and rank 0.

2. For each NodeRankInput in entries:
   a. Add an InternalEntry to the map keyed by `logical_name`.
      Set dependencies to empty list (populated in Step 2).
      Set rank to 0.

   b. For each output in `frontmatter.outputs`:
      - Construct the artifact logical name:
        - Strip the "ROOT/" prefix from `logical_name`.
        - Prepend "ARTIFACT/".
        - Append "(<id>)" where <id> is the output's id field.
        - Example: node "ROOT/a/b" with output id "foo" → "ARTIFACT/a/b(foo)".
      - Add an InternalEntry to the map keyed by this artifact logical name.
        Set dependencies to empty list (populated in Step 2).
        Set rank to 0.

### Step 2 — Build dependency edges

1. For each spec node InternalEntry (entries whose key starts with "ROOT/"):
   a. If the logical name is "ROOT", skip (no dependencies — handled in Step 3).
   b. Otherwise:
      - **Parent dependency**: call LogicalNameGetParent on the logical name.
        Add the parent logical name to this entry's dependencies list.
      - **depends_on dependencies**: for each item in `frontmatter.depends_on`:
        - If the item starts with "ARTIFACT/", use it as-is as the lookup key.
        - If the item starts with "ROOT/", call LogicalNameStripQualifier on it
          to get the bare logical name; use that as the lookup key.
        - If the lookup key is not found in the entry map,
          raise error "unresolvable reference".
        - Add the lookup key to this entry's dependencies list.
      - **input dependency**: if `frontmatter.input` is non-empty:
        - Use the input value as-is as the lookup key (it is an ARTIFACT/ reference).
        - If the lookup key is not found in the entry map,
          raise error "unresolvable reference".
        - Add the lookup key to this entry's dependencies list.

2. For each artifact InternalEntry (entries whose key starts with "ARTIFACT/"):
   a. Determine the generating node's logical name:
      - Strip the "ARTIFACT/" prefix from the artifact key.
      - Remove the trailing "(<id>)" qualifier.
      - Prepend "ROOT/".
      - This is the generating node's logical name.
   b. If the generating node logical name is not found in the entry map,
      raise error "unresolvable reference".
   c. Add the generating node logical name to this artifact entry's dependencies list.

### Step 3 — Initialize ranks

1. Set the rank of the entry keyed "ROOT" to 0.
   This rank is fixed and will not be updated during iteration.

2. All other entries already have rank 0 from Step 1 (no change needed).

### Step 4 — Iterate and detect cycles

1. Let N = total number of entries in the map.

2. Set changed_in_last_pass = empty list.

3. Repeat up to N times (loop index i from 1 to N):
   a. Set changed_this_pass = empty list.
   b. For each InternalEntry in the map (excluding "ROOT"):
      - For each dependency logical name in the entry's dependencies list:
        - Look up the dependency's current rank in the entry map.
      - Compute new_rank = 1 + max(rank of all dependencies).
        If the entry has no dependencies, new_rank = 1.
      - If new_rank > current rank of this entry:
        - Update the entry's rank to new_rank.
        - Add this entry's logical name to changed_this_pass.
   c. If changed_this_pass is empty, stop iterating (converged, no cycles).
   d. Set changed_in_last_pass = changed_this_pass.

4. If the loop completed all N iterations without converging
   (i.e., changed_this_pass was non-empty on the final pass):
   - The entries in changed_in_last_pass are cycle participants.
   - Proceed to Step 5 with these as the cycles list.
   Otherwise, set cycles list to empty.

### Step 5 — Output

1. Build the ranked list:
   For each InternalEntry in the entry map, create a NodeRankEntry
   with logical_name and rank.

2. Sort the ranked list:
   - Primary sort: rank ascending.
   - Secondary sort: logical_name ascending (lexicographic).

3. Return:
   - ranked: the sorted list of NodeRankEntry records.
   - cycles: the list of logical names from changed_in_last_pass
     (empty if no cycles were detected).
```

## Error conditions

| Condition | Description |
|---|---|
| "unresolvable reference" | A depends_on entry, input field, or artifact's generating node could not be found in the entry map. The reference target was not present among the discovered nodes. |
