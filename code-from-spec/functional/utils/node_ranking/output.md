<!-- code-from-spec: ROOT/functional/utils/node_ranking@3TsXr-vKrqXTDJ1LE42S4Fw1N3M -->

# Node Ranking

## Records

```
record RankedEntry
  logical_name: string
  rank: integer
```

## Functions

---

### DetectCycles(nodes) -> (ranked_entries, cycle_participants)

**Parameters**

- `nodes`: list of records, each containing:
  - `logical_name`: string (a ROOT/ logical name)
  - `frontmatter`: a Frontmatter record with fields:
    - `depends_on`: list of strings (logical names)
    - `input`: string (logical name of an ARTIFACT/ reference, or empty)
    - `outputs`: list of Output records, each with `id` and `path`

**Returns**

- `ranked_entries`: list of RankedEntry records (one per spec node and one per artifact)
- `cycle_participants`: list of strings (logical names involved in cycles; empty if no cycles)

**Errors**

- `"unresolvable reference"`: a `depends_on` or `input` target cannot be resolved to a known entry.

---

#### Step 1 — Discovery: collect all entries and build dependency lists

1. Initialize an empty index called `entry_map`.
   Each key is a logical name string; each value is a record:
   - `logical_name`: string
   - `rank`: integer (starts at 0)
   - `dependencies`: list of logical name strings

2. For each node in `nodes`:
   a. Add a spec-node entry to `entry_map` keyed by the node's `logical_name`.
      Set `rank` to 0. Set `dependencies` to an empty list for now.

   b. For each output in the node's `frontmatter.outputs`:
      - Construct the artifact logical name as:
        `"ARTIFACT/" + <node path without ROOT/ prefix> + "(" + output.id + ")"`
        For example, node `ROOT/functional/utils/frontmatter` with output id
        `frontmatter` produces key `ARTIFACT/functional/utils/frontmatter(frontmatter)`.
      - Add an artifact entry to `entry_map` keyed by that artifact logical name.
        Set `rank` to 0.
        Set `dependencies` to a list containing only the generating node's logical name.

3. For each node in `nodes`, populate the spec-node entry's dependency list:
   a. Start with a list containing the node's parent logical name.
      - If the node's logical name is `ROOT`, it has no parent. Skip this step.
      - Otherwise, compute the parent by stripping the last path segment from the
        node's logical name (e.g., parent of `ROOT/x/y` is `ROOT/x`).

   b. For each entry in the node's `frontmatter.depends_on`:
      - If the entry starts with `ARTIFACT/`, look it up in `entry_map`.
        If not found, raise error `"unresolvable reference"`.
        Add the entry's logical name to the dependency list.
      - If the entry starts with `ROOT/`, verify it exists in `entry_map`.
        If not found, raise error `"unresolvable reference"`.
        Add the entry's logical name to the dependency list.

   c. If the node's `frontmatter.input` is not empty:
      - Look up the input value in `entry_map`.
        If not found, raise error `"unresolvable reference"`.
      - Add the input value to the dependency list.

   d. Store the completed dependency list in the spec-node entry in `entry_map`.

---

#### Step 2 — Initialization

4. For every entry in `entry_map`, set `rank` to 0.
   (All entries start at rank 0 regardless of position in the graph.)

---

#### Step 3 — Iteration (single pass)

Define a sub-procedure `RunOnePass(entry_map)` -> `changed` (boolean):

1. Set `changed` to false.

2. For each entry in `entry_map`:
   a. If the entry's `dependencies` list is empty, skip it.
      (ROOT has no dependencies; its rank stays 0.)

   b. For each logical name in the entry's `dependencies`:
      - Look up that logical name in `entry_map` to get its current `rank`.

   c. Compute `max_dep_rank` as the maximum `rank` among all dependency entries.

   d. Compute `new_rank` = `max_dep_rank` + 1.

   e. If `new_rank` is greater than the entry's current `rank`:
      - Update the entry's `rank` to `new_rank`.
      - Set `changed` to true.

3. Return `changed`.

---

#### Step 4 — Convergence: repeat until stable

5. Set `N` to the total number of entries in `entry_map`.

6. Set `pass_count` to 0.

7. Repeat:
   a. Call `RunOnePass(entry_map)`. Store the result as `changed`.
   b. Increment `pass_count` by 1.
   c. If `changed` is false, stop repeating. Convergence reached.
   d. If `pass_count` equals `N`, stop repeating.
      (N passes have been completed — proceed to cycle detection.)

---

#### Step 5 — Cycle detection

8. If `pass_count` is less than `N`, there are no cycles.
   Set `cycle_participants` to an empty list.

9. Else (`pass_count` equals `N`):
   a. Run one additional pass by calling `RunOnePass(entry_map)`.
      Collect the logical names of every entry whose rank changed during this pass.
   b. Set `cycle_participants` to that collected list of logical names.

---

#### Step 6 — Build result

10. Initialize `ranked_entries` as an empty list.

11. For each entry in `entry_map`:
    - Create a RankedEntry record with:
      - `logical_name`: the entry's logical name
      - `rank`: the entry's final rank
    - Append it to `ranked_entries`.

12. Return (`ranked_entries`, `cycle_participants`).

---

## Contracts and invariants

- Every entry in `entry_map` appears in `ranked_entries` exactly once.
- `cycle_participants` contains all entries that participated in a cycle,
  not merely one representative entry.
- Cycle detection is a by-product of the ranking iteration; no separate
  graph traversal is performed.
- Entries with equal rank have no dependency between them — neither
  directly nor transitively.
- The root node `ROOT` always has rank 0 (it has no dependencies).
- An artifact always has a rank strictly greater than the node that generates it.
