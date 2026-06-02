---
depends_on:
  - ROOT/functional/logic/utils/logical_names
  - ROOT/functional/logic/parsing/frontmatter(interface)
output: code-from-spec/functional/logic/utils/node_ranking/output.md
---

# ROOT/functional/logic/utils/node_ranking

Iterative ranking of spec tree nodes and artifacts,
with cycle detection as a side effect.

# Public

## Interface

```
record NodeRankInput
  logical_name: string
  frontmatter: frontmatter.Frontmatter

record NodeRankEntry
  logical_name: string
  rank: integer

function NodeRankCompute(entries: list of NodeRankInput) -> (ranked: list of NodeRankEntry, cycles: list of string)
  errors:
    - UnresolvableReference: a depends_on or input
      target cannot be resolved.
```

Takes the full set of discovered nodes with their parsed
frontmatter. Returns ranked entries (nodes and artifacts)
and a list of logical names involved in cycles (empty if
no cycles).

## Description

Every node and artifact receives an integer rank. Nodes
with lower rank must be processed before nodes with
higher rank. Entries with equal rank have no dependency
between them and can be processed in parallel.

- The root node (`ROOT`) has rank 0 (fixed, special case).
- For any other spec node: rank = 1 + max(rank of parent,
  rank of each `depends_on` entry, rank of the `input`
  artifact if present).
- For an artifact: rank = 1 + rank of the node that
  generates it.

The algorithm is an iterative relaxation (Bellman-Ford
style): repeat rank updates until convergence or until
N passes are exhausted. If the loop does not converge
within N passes, a cycle exists.

# Agent

## Behavior

### Step 1 — Build entry map

From the input list, build an entry map keyed by logical
name. Each entry tracks its dependency list and current
rank.

For each `NodeRankInput`:
- Add a spec node entry keyed by `logical_name`.
- For each output in `frontmatter.outputs`, add an
  artifact entry keyed by its `ARTIFACT/` logical name.
  Construct the artifact logical name by stripping the
  `ROOT/` prefix from the node's logical name, prepending
  `ARTIFACT/`, and appending `(id)` where `id` is the
  output's id field. Example: node `ROOT/a/b` with
  output id `foo` → `ARTIFACT/a/b(foo)`.

### Step 2 — Build dependency edges

Every entry will have at least one dependency: spec
nodes (other than ROOT) always have a parent, and
artifact entries always depend on their generating node.
ROOT is the only entry with no dependencies — it is
handled as a special case in Step 3.

For each spec node entry:
- **Parent**: derive from logical name using
  `LogicalNameGetParent`. The root node has no parent
  (it is a special case — see Step 3).
- **depends_on**: for each entry in
  `frontmatter.depends_on`, determine the lookup key.
  For `ARTIFACT/` references, use as-is (the qualifier
  is part of the key). For `ROOT/` references, use
  `LogicalNameStripQualifier` to get the bare logical
  name for lookup. The dependency edge points to the
  bare node entry.
- **input**: if `frontmatter.input` is non-empty, add it
  as a dependency (it is an `ARTIFACT/` reference, used
  as-is).

For each artifact entry:
- Depends on the node that generates it (the node whose
  `outputs` produced this artifact).

If any dependency target is not found in the entry map,
return the "unresolvable reference" error.

### Step 3 — Initialize ranks

Assign rank 0 to the root node (`ROOT`). The root is a
special case: its rank is fixed at 0 and it is excluded
from the iteration loop.

Assign rank 0 to all other entries as an initial value.

### Step 4 — Iterate and detect cycles

Let N = total number of entries in the map.

Repeat up to N times:
- For each entry (excluding `ROOT`), compute:
  rank = 1 + max(rank of its dependencies).
  If the computed rank exceeds the current rank, update
  it and mark this pass as "changed".
- If a full pass produces no changes, stop (converged,
  no cycles).

In a cycle-free graph, convergence is guaranteed within
N-1 passes. If after N full passes any rank still
changes, a cycle exists. Report the entries whose rank changed in the last
pass — these are not necessarily all cycle participants,
but they are sufficient to guide diagnosis. The goal is
to surface the cycle, not to enumerate it completely.

### Step 5 — Output

Return all entries as `NodeRankEntry` (logical_name +
rank), sorted by rank ascending then logical name
ascending. Return cycle participants as a list of logical
names.

## Contracts

- Returns all entries (nodes and artifacts), not just
  nodes.
- Reports entries involved in non-convergence — enough
  to guide diagnosis, not necessarily the full cycle.
- Cycle detection is a side effect of ranking — no
  separate graph traversal is needed.
