<!-- code-from-spec: ROOT/functional/logic/utils/node_ranking@WMtnzK5Zvlo7qCKD8LbJdNc4aEU -->

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
  rank: integer
  dependencies: list of strings
```

## Functions

```
function NodeRankCompute(entries: list of NodeRankInput)
  -> (ranked: list of NodeRankEntry, cycles: list of string)

  errors:
    - "unresolvable reference": a depends_on or input target
      cannot be resolved in the entry map.
```

### Step 1 — Build entry map

  1. Create an empty entry map keyed by logical name,
     where each value is an InternalEntry.

  2. For each NodeRankInput in entries:

     a. Add a spec node entry to the map:
        - key: the NodeRankInput's logical_name
        - rank: 0
        - dependencies: (to be filled in Step 2)

     b. For each output in the NodeRankInput's
        frontmatter.outputs:
        - Construct the artifact logical name:
          Strip the "ROOT/" prefix from the node's
          logical_name. Prepend "ARTIFACT/". Append
          "(" + output.id + ")".
          Example: node "ROOT/a/b", output id "foo"
          → "ARTIFACT/a/b(foo)".
        - Add an artifact entry to the map:
          - key: the constructed artifact logical name
          - rank: 0
          - dependencies: (to be filled in Step 2)

### Step 2 — Build dependency edges

  3. For each spec node entry in the entry map:

     a. If the entry's logical_name is "ROOT":
        - Set its dependencies to an empty list.
        - Skip to the next entry.

     b. Otherwise, compute the parent logical name using
        LogicalNameGetParent(logical_name).
        Add the parent logical name to this entry's
        dependencies list.

     c. For each item in frontmatter.depends_on:
        - If the item starts with "ARTIFACT/":
          use the item as-is as the lookup key.
        - If the item starts with "ROOT/":
          strip any parenthetical qualifier — find the
          first "(" character and truncate the string
          there. Use the resulting bare name as the
          lookup key.
        - If the lookup key is not found in the entry
          map, raise error "unresolvable reference".
        - Add the lookup key to this entry's
          dependencies list.

     d. If frontmatter.input is non-empty:
        - Use frontmatter.input as-is as the lookup key.
        - If the lookup key is not found in the entry
          map, raise error "unresolvable reference".
        - Add the lookup key to this entry's
          dependencies list.

  4. For each artifact entry in the entry map:

     a. Determine the generating node's logical name:
        Use LogicalNameGetArtifactGenerator(artifact_key)
        to obtain the "ROOT/" logical name of the node
        that generates this artifact.

     b. If the generating node's logical name is not
        found in the entry map, raise error
        "unresolvable reference".

     c. Set the artifact entry's dependencies to a list
        containing only the generating node's logical name.

### Step 3 — Initialize ranks

  5. Set the rank of the "ROOT" entry to 0. Its rank is
     fixed and will not change during iteration.

  6. Set the rank of all other entries to 0 as the
     initial value.

### Step 4 — Iterate and detect cycles

  7. Let N = total number of entries in the entry map.

  8. Initialize changed_in_last_pass to an empty list.

  9. Repeat up to N times (pass index from 1 to N):

     a. Set changed_this_pass to an empty list.

     b. For each entry in the entry map, excluding "ROOT":

        i.  For each dependency logical name in the
            entry's dependencies list, look up its
            current rank in the entry map.

        ii. Compute candidate_rank =
            1 + max(rank of all dependencies).
            If the dependencies list is empty,
            candidate_rank = 1.

        iii. If candidate_rank > entry's current rank:
             - Update the entry's rank to candidate_rank.
             - Add the entry's logical name to
               changed_this_pass.

     c. If changed_this_pass is empty:
        - Convergence reached. No cycles. Stop iterating.

     d. Otherwise, set changed_in_last_pass to
        changed_this_pass.

  10. If the loop completed all N passes and the last
      pass still produced changes (changed_in_last_pass
      is non-empty), a cycle exists.
      - The cycle participants are the entries listed in
        changed_in_last_pass.

### Step 5 — Output

  11. Collect all entries from the entry map as
      NodeRankEntry records (logical_name + rank).

  12. Sort the collected entries:
      - Primary sort: rank ascending.
      - Secondary sort: logical_name ascending
        (lexicographic).

  13. Return:
      - ranked: the sorted list of NodeRankEntry records.
      - cycles: the list of logical names from
        changed_in_last_pass (empty if no cycle was
        detected).
```
