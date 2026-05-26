<!-- code-from-spec: ROOT/functional/utils/node_ranking@4QTKCsICzk8SumK980zOxy0opSI -->

# node_ranking

## Records

```
record RankedEntry
  logical_name: string
  rank: integer
```

## Functions

---

### DetectCycles(nodes) -> (ranked_entries, cycle_participants)

Parameters:
- nodes: list of records, each containing:
  - logical_name: string
  - frontmatter: Frontmatter record (with depends_on, input, outputs fields)

Returns:
- ranked_entries: list of RankedEntry
- cycle_participants: list of strings (logical names involved in cycles; empty if no cycles)

Errors:
- "unresolvable reference": a depends_on or input target cannot be resolved to a known entry.

---

**Step 1 — Discovery**

1. Initialize an empty map called `dependency_map` from logical_name -> list of logical_names.
   Initialize an empty map called `rank_map` from logical_name -> integer.

2. For each node in nodes:
   a. Add the node's logical_name to `dependency_map` with an empty dependency list.
   b. For each output in node's frontmatter.outputs:
      - Construct the artifact logical name as "ARTIFACT/<node_path>(<output.id>)".
      - Add the artifact logical name to `dependency_map` with a dependency list containing only the node's logical_name.

3. For each node in nodes:
   a. Build the node's dependency list:
      - If the node's logical_name is "ROOT":
        - Dependency list is empty.
      - Else:
        - Compute the parent logical_name using GetParent(node's logical_name).
        - Start dependency list with [parent logical_name].
        - For each entry in node's frontmatter.depends_on:
          - If the entry is not a key in `dependency_map`, raise error "unresolvable reference".
          - Append entry to the dependency list.
        - If node's frontmatter.input is not empty:
          - If the input is not a key in `dependency_map`, raise error "unresolvable reference".
          - Append the input to the dependency list.
   b. Store the dependency list in `dependency_map` for this node's logical_name.

**Step 2 — Initialization**

4. For each logical_name in `dependency_map`:
   - Set `rank_map[logical_name]` = 0.

**Step 3 & 4 — Iteration until convergence**

5. Set `total_entries` = number of keys in `dependency_map`.
   Set `pass_count` = 0.
   Set `changed` = true.

6. While `changed` is true:
   a. Set `changed` = false.
   b. Increment `pass_count` by 1.
   c. For each logical_name in `dependency_map`:
      - Get its dependency list from `dependency_map`.
      - If the dependency list is empty:
        - computed_rank = 0.
      - Else:
        - computed_rank = 1 + maximum of `rank_map[dep]` for each dep in the dependency list.
      - If computed_rank > `rank_map[logical_name]`:
        - Set `rank_map[logical_name]` = computed_rank.
        - Set `changed` = true.

**Step 5 — Cycle detection**

7. If `pass_count` > `total_entries` and `changed` is still true after the last pass:
   - A cycle exists.
   - Perform one additional pass to identify which entries still change:
     a. Initialize an empty list `cycle_participants`.
     b. For each logical_name in `dependency_map`:
        - Get its dependency list.
        - If the dependency list is empty:
          - computed_rank = 0.
        - Else:
          - computed_rank = 1 + maximum of `rank_map[dep]` for each dep in the dependency list.
        - If computed_rank > `rank_map[logical_name]`:
          - Append logical_name to `cycle_participants`.
   - Set cycle_participants to the collected list.
   Else:
   - Set `cycle_participants` to an empty list.

8. Build `ranked_entries` as a list of RankedEntry records:
   - For each logical_name in `rank_map`:
     - Create a RankedEntry with logical_name = logical_name, rank = `rank_map[logical_name]`.

9. Return (`ranked_entries`, `cycle_participants`).
```

## Contracts

- All entries in `dependency_map` (both spec nodes and artifacts) appear in the returned `ranked_entries`.
- `cycle_participants` contains all logical names still changing after N passes — not just one representative.
- Cycle detection is a natural consequence of the iterative ranking algorithm; no separate graph traversal is performed.
- Entries with equal rank have no dependency relationship with each other.
- ROOT always resolves to rank 0, as it has no dependencies.
- Artifacts always have rank = 1 + rank of their generating node.
```
