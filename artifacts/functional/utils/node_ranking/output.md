<!-- code-from-spec: ROOT/functional/utils/node_ranking@PENDING -->

## Data structures

```
record RankedEntry
  logical_name: string
  rank: integer
```

## Functions

### DetectCycles(nodes) -> (ranked_entries, cycle_participants)

**Step 1 -- Discovery**

1. Create an empty entries list and an empty dependency map.

2. For each node in nodes:
   a. Add the node's logical name to entries.
   b. Build the node's dependency list:
      - The parent node (derived by removing the last segment of
        the logical name; for "ROOT" there is no parent).
      - Each entry in the node's frontmatter depends_on list.
      - The node's frontmatter input artifact, if present.
   c. For each dependency target, verify it exists in the set of
      known entries or nodes.
      If it cannot be resolved, raise error "unresolvable reference".
   d. Store the dependency list in the dependency map keyed by
      logical name.

3. For each node that has an outputs field in its frontmatter:
   a. For each output, add the output's artifact path to entries.
   b. The artifact's dependency list contains only the node that
      generates it.
   c. Store the dependency list in the dependency map.

**Step 2 -- Initialization**

4. Set the rank of every entry to 0.

**Step 3 -- Iteration**

5. For each entry in the entries list:
   a. Compute new_rank as 1 + the maximum rank among all entries
      in its dependency list.
      If the dependency list is empty, new_rank stays at 0.
   b. If new_rank is greater than the entry's current rank,
      update the rank to new_rank and mark that a change occurred.

**Step 4 -- Convergence**

6. Repeat step 5 until a full pass completes with no rank changes.

**Step 5 -- Cycle detection**

7. Let N be the total number of entries.
   If convergence has not been reached within N full passes:
   a. Perform one more pass.
   b. Collect every entry whose rank changed during this final pass.
      These are the cycle participants.

8. Build a list of RankedEntry records from all entries and their
   final ranks.

9. Return the list of ranked entries and the list of cycle
   participant logical names (empty if no cycles were detected).
