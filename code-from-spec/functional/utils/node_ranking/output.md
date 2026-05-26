<!-- code-from-spec: ROOT/functional/utils/node_ranking@4QTKCsICzk8SumK980zOxy0opSI -->

# node_ranking

Assigns integer ranks to all spec nodes and artifacts so that
lower-ranked entries are processed before higher-ranked ones.
Cycle detection is a by-product of the ranking algorithm — no
separate graph traversal is needed.

---

## Records

```
record RankedEntry
  logical_name: string   -- identifies the spec node or artifact
  rank: integer          -- processing order; lower = earlier

record NodeInfo
  logical_name: string
  parent: optional string          -- logical name of parent node (absent for ROOT)
  depends_on: list of strings      -- logical names from frontmatter depends_on
  input: optional string           -- logical name of input artifact (if present)
  outputs: list of Output          -- artifacts produced by this node
```

---

## Functions

### DetectCycles

```
function DetectCycles(nodes) -> (ranked_entries, cycle_participants)

  Parameters:
    nodes — list of NodeInfo records representing the full set of
            discovered spec nodes, each carrying parsed frontmatter data.

  Returns:
    ranked_entries     — list of RankedEntry; one entry per spec node
                         and one entry per artifact declared in any node's
                         outputs field.
    cycle_participants — list of logical names involved in a dependency
                         cycle; empty if no cycles were found.

  Errors:
    - "unresolvable reference": a depends_on entry or input artifact
      target cannot be matched to any known entry in the working set.
```

**Step 1 — Build the working set**

  1. Create an empty map called `entries` keyed by logical_name.

  2. For each node in nodes:
     a. Add an entry for the spec node itself with the key being
        the node's logical_name.
     b. For each output declared in the node's outputs field:
        Add an entry for the artifact using the artifact's logical
        name (constructed as ARTIFACT/<node_path>(<output.id>)).

  3. For each entry in `entries`, build its dependency list:
     - If the entry is a spec node:
         - If the node is ROOT (logical_name equals "ROOT"):
             dependency list is empty.
         - Else:
             dependency list starts with the parent logical_name.
             Append each entry in depends_on.
             If input is present, append the input artifact logical_name.
     - If the entry is an artifact:
         dependency list contains exactly the logical_name of the
         spec node that declared this artifact in its outputs.

  4. Verify all dependency references:
     For each dependency logical_name in every entry's dependency list,
     check that it exists as a key in `entries`.
     If any reference is not found, raise error "unresolvable reference"
     identifying the missing logical_name.

**Step 2 — Initialize ranks**

  5. Set rank = 0 for every entry in `entries`.

**Step 3 — Iterative rank propagation**

  6. Set `changed` = true.
     Set `pass_count` = 0.
     Set `N` = total number of entries in `entries`.

  7. While `changed` is true:
     a. Set `changed` = false.
     b. Increment `pass_count` by 1.
     c. For each entry in `entries`:
        i.  Compute `max_dep_rank` = maximum rank among all entries
            in this entry's dependency list.
            If the dependency list is empty, `max_dep_rank` = -1
            (so that 1 + max_dep_rank = 0, preserving rank 0 for ROOT).
        ii. Set `candidate` = 1 + `max_dep_rank`.
        iii. If `candidate` > current rank of this entry:
               Update the entry's rank to `candidate`.
               Set `changed` = true.

**Step 4 — Convergence check (cycle detection)**

  8. If `pass_count` > `N` and `changed` is still true after the
     last pass (meaning ranks were still changing at pass N+1 or
     beyond):
     a. Perform one additional pass, recording all entries whose
        rank changed during that pass into `cycle_participants`.
     b. Stop iterating.
     Otherwise (converged within N passes):
     a. Set `cycle_participants` = empty list.

**Step 5 — Build output**

  9. For each entry in `entries`, construct a RankedEntry with:
       logical_name = entry's logical_name
       rank         = entry's final rank

  10. Return (list of all RankedEntry records, cycle_participants).
```

---

## Contracts

- Every spec node and every artifact appears in `ranked_entries`.
- ROOT always receives rank 0 (no dependencies).
- For any spec node other than ROOT:
    rank = 1 + max(rank of parent, rank of each depends_on entry,
                   rank of input artifact if present).
- For any artifact:
    rank = 1 + rank of the node that generates it.
- Entries with equal rank have no dependency relationship between
  them and may be processed in any order relative to each other.
- All cycle participants are reported — not just the first one found.
- If no cycles exist, cycle_participants is an empty list.
