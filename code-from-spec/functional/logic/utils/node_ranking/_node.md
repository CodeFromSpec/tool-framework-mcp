---
depends_on:
  - ROOT/functional/logic/utils/logical_names
  - ROOT/functional/logic/parsing/frontmatter
outputs:
  - id: node_ranking
    path: code-from-spec/functional/logic/utils/node_ranking/output.md
---

# ROOT/functional/logic/utils/node_ranking

Detects circular references in the spec tree using
iterative ranking.

Review status: pending

# Public

## Interface

```
record RankedEntry
  logical_name: string
  rank: integer

function DetectCycles(nodes: list of DiscoveredNode) -> (ranked_entries, cycle_participants)
  errors:
    - unresolvable reference: a depends_on or input target cannot be resolved.
```

Takes the full set of discovered nodes with their parsed
frontmatter. Returns the ranked entries and a list of
logical names involved in cycles (empty if no cycles).

# Agent

## Behavior

### Rank definition

Every node and artifact receives an integer rank. Nodes
with lower rank must be processed before nodes with
higher rank.

- The root node (`ROOT`) has rank 0.
- For any other spec node: rank = 1 + max(rank of parent,
  rank of each `depends_on` entry, rank of the `input`
  artifact if present).
- For an artifact: rank = 1 + rank of the node that
  generates it.

### Algorithm

**Step 1 — Discovery**

Collect all entries: spec nodes and artifacts (from each
node's `outputs` field).

Each artifact is indexed by its `ARTIFACT/` logical name,
constructed from the generating node's logical name and
the output's `id`. For example, node `ROOT/functional/logic/parsing/frontmatter`
with output `id: frontmatter` produces an artifact entry
keyed as `ARTIFACT/functional/logic/parsing/frontmatter(frontmatter)`.

When resolving `depends_on` and `input` references to
entries in the entry map:
- `ARTIFACT/` references are used as-is — the qualifier
  is part of the key (e.g. `ARTIFACT/x(id)`).
- `ROOT/` references with a parenthetical qualifier must
  have the qualifier stripped before lookup, because the
  entry map is keyed by the bare node name (e.g.
  `ROOT/x(y)` looks up `ROOT/x`). The dependency edge
  points to the bare node entry.

For each entry, build a dependency list:
- Spec nodes depend on: parent, `depends_on` entries
  (after qualifier stripping for `ROOT/` refs),
  `input` artifact (if present).
- Artifacts depend on: the node that generates them.

**Step 2 — Initialization**

Assign rank 0 to every entry.

**Step 3 — Iteration**

For each entry, compute its rank as 1 + max rank of its
dependency list. If the computed rank is higher than the
current rank, update it.

**Step 4 — Convergence**

Repeat step 3 until no rank changes in a full pass.

**Step 5 — Cycle detection**

If after N full passes (where N is the total number of
entries) any rank still changes, a cycle exists. Entries
whose rank changed in the last pass are reported as
cycle participants.

## Contracts

- Returns all cycle participants — not just one.
- Cycle detection is a side effect of ranking — no
  separate graph traversal is needed.
- Entries with equal rank have no dependency between
  them.
