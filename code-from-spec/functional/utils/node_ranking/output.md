<!-- code-from-spec: ROOT/functional/utils/node_ranking@cJD8AoL5qmtYHKAbGrMZVFk3oGc -->

# Node Ranking

Computes an integer rank for every spec node and artifact in the
discovered set. Rank determines processing order: lower-ranked
entries must be processed before higher-ranked ones. Also
identifies entries that participate in dependency cycles.

---

## Data Structures

```
record ExternalFragment
  description: optional string
  lines: string
  hash: string

record External
  path: string
  fragments: optional list of ExternalFragment

record Output
  id: string
  path: string

record Frontmatter
  depends_on: list of strings
  external: list of External
  input: string
  outputs: list of Output

record NodeInfo
  logical_name: string        -- e.g. "ROOT/x/y"
  frontmatter: Frontmatter

record Entry
  logical_name: string        -- key in the entry map (bare, no qualifier)
  kind: "node" or "artifact"
  generating_node: string     -- only for kind = "artifact": the node logical name
  dependencies: list of strings  -- logical names this entry directly depends on
  rank: integer

record RankedEntry
  logical_name: string
  rank: integer
```

---

## Functions

### BuildEntryMap

```
function BuildEntryMap(node_infos) -> entry_map
  -- node_infos: list of NodeInfo records (all discovered spec nodes)
  -- entry_map: map from logical_name (string) -> Entry

  1. Initialize entry_map as an empty map.

  2. For each node_info in node_infos:

     a. Create a node entry:
          entry.logical_name    = node_info.logical_name
          entry.kind            = "node"
          entry.generating_node = ""
          entry.dependencies    = empty list
          entry.rank            = 0
        Insert into entry_map keyed by node_info.logical_name.

     b. For each output in node_info.frontmatter.outputs:
          -- Construct the ARTIFACT/ logical name for this output.
          -- The key format is: ARTIFACT/<path-suffix>(<id>)
          -- where <path-suffix> is the part of node_info.logical_name
          -- after "ROOT/" (i.e., strip the "ROOT/" prefix).
          --
          -- Example:
          --   node logical_name = "ROOT/functional/utils/frontmatter"
          --   output id         = "frontmatter"
          --   artifact key      = "ARTIFACT/functional/utils/frontmatter(frontmatter)"

          path_suffix   = StripRootPrefix(node_info.logical_name)
          artifact_key  = "ARTIFACT/" + path_suffix + "(" + output.id + ")"

          Create an artifact entry:
            entry.logical_name    = artifact_key
            entry.kind            = "artifact"
            entry.generating_node = node_info.logical_name
            entry.dependencies    = empty list
            entry.rank            = 0
          Insert into entry_map keyed by artifact_key.

  3. Return entry_map.
```

---

### BuildDependencies

```
function BuildDependencies(entry_map, node_infos) -> (entry_map, unresolvable_refs)
  -- Populates the dependencies list for every entry.
  -- Returns the updated entry_map and a list of unresolvable reference strings.

  unresolvable_refs = empty list

  1. For each node_info in node_infos:
       node_entry = entry_map[node_info.logical_name]

       a. Add the parent dependency (all nodes except ROOT):
            parent = GetParent(node_info.logical_name)
            -- GetParent returns empty string when the node IS ROOT.
            if parent is not empty:
              if parent is not in entry_map:
                append parent to unresolvable_refs
              else:
                append parent to node_entry.dependencies

       b. For each ref in node_info.frontmatter.depends_on:
            -- Strip qualifier from ROOT/ references before lookup.
            -- ARTIFACT/ references are used as-is (qualifier is part of the key).
            lookup_key = NormalizeRef(ref)
            if lookup_key is not in entry_map:
              append ref to unresolvable_refs
            else:
              append lookup_key to node_entry.dependencies

       c. If node_info.frontmatter.input is not empty:
            input_ref = node_info.frontmatter.input
            -- input always refers to an artifact; use as-is.
            if input_ref is not in entry_map:
              append input_ref to unresolvable_refs
            else:
              append input_ref to node_entry.dependencies

       d. Write node_entry back into entry_map.

  2. For each entry in entry_map where entry.kind = "artifact":
       -- Artifact depends only on the node that generates it.
       artifact_entry = entry
       gen_node = artifact_entry.generating_node
       if gen_node is not in entry_map:
         append gen_node to unresolvable_refs
       else:
         append gen_node to artifact_entry.dependencies
       Write artifact_entry back into entry_map.

  3. Return (entry_map, unresolvable_refs).
```

---

### NormalizeRef

```
function NormalizeRef(ref) -> string
  -- Strips the parenthetical qualifier from ROOT/ references.
  -- ARTIFACT/ references are returned unchanged.

  1. If ref starts with "ARTIFACT/":
       return ref

  2. If ref starts with "ROOT/":
       stripped = StripQualifier(ref)
       -- StripQualifier removes "(anything)" from the end.
       -- e.g. "ROOT/x/y(z)" -> "ROOT/x/y"
       return stripped

  3. return ref   -- unrecognized prefix; caller will report unresolvable
```

---

### DetectCycles

```
function DetectCycles(nodes) -> (ranked_entries, cycle_participants)
  -- nodes: list of NodeInfo records for all discovered spec nodes.
  -- ranked_entries: list of RankedEntry (one per entry, nodes + artifacts).
  -- cycle_participants: list of logical_name strings involved in cycles
  --                     (empty list if no cycles detected).

  errors:
    - "unresolvable reference: <ref>" when a depends_on or input target
      cannot be found in the entry map.

  -- ── Step 1: Discovery ──────────────────────────────────────────────────

  1. Call BuildEntryMap(nodes) -> entry_map.

  -- ── Step 2: Build dependency edges ─────────────────────────────────────

  2. Call BuildDependencies(entry_map, nodes) -> (entry_map, unresolvable_refs).

     If unresolvable_refs is not empty:
       raise error "unresolvable reference: " + join(unresolvable_refs, ", ")

  -- ── Step 3: Initialization ──────────────────────────────────────────────

  3. All entries already have rank = 0 (set during BuildEntryMap).
     No additional initialization needed.

  -- ── Step 4: Iterative rank propagation ─────────────────────────────────

  4. total_entries = count of entries in entry_map.
     pass_number = 0

     Repeat:
       changed = false
       pass_number = pass_number + 1

       For each entry in entry_map:
         if entry.dependencies is empty:
           computed_rank = 0
         else:
           max_dep_rank = 0
           for each dep_name in entry.dependencies:
             dep_entry = entry_map[dep_name]
             if dep_entry.rank > max_dep_rank:
               max_dep_rank = dep_entry.rank
           computed_rank = 1 + max_dep_rank

         if computed_rank > entry.rank:
           entry.rank = computed_rank
           changed = true
           Write entry back into entry_map.

       if changed is false:
         -- Ranks have converged; no cycles.
         break out of Repeat loop.

       if pass_number >= total_entries:
         -- After N passes ranks are still changing; a cycle exists.
         -- Collect participants: entries whose rank changed in this last pass.
         cycle_participants = empty list
         for each entry in entry_map:
           -- Re-run the rank computation one more time to find still-changing entries.
           if entry.dependencies is not empty:
             max_dep_rank = 0
             for each dep_name in entry.dependencies:
               dep_entry = entry_map[dep_name]
               if dep_entry.rank > max_dep_rank:
                 max_dep_rank = dep_entry.rank
             computed_rank = 1 + max_dep_rank
             if computed_rank > entry.rank:
               append entry.logical_name to cycle_participants
         break out of Repeat loop.

  -- ── Step 5: Assemble results ────────────────────────────────────────────

  5. ranked_entries = empty list
     For each entry in entry_map:
       Create RankedEntry:
         logical_name = entry.logical_name
         rank         = entry.rank
       Append to ranked_entries.

  6. If cycle_participants is not defined (normal convergence path):
       cycle_participants = empty list

  7. Return (ranked_entries, cycle_participants).
```

---

## Helper: StripRootPrefix

```
function StripRootPrefix(logical_name) -> string
  -- Removes the leading "ROOT/" from a logical name.
  -- e.g. "ROOT/x/y" -> "x/y"
  -- e.g. "ROOT"     -> ""  (edge case for the root node itself)

  1. If logical_name equals "ROOT":
       return ""

  2. If logical_name starts with "ROOT/":
       return substring of logical_name after the first "ROOT/" prefix.

  3. return logical_name   -- unchanged if not a ROOT/ name
```

---

## Contracts and Invariants

- Every entry in the returned `ranked_entries` has a non-negative rank.
- `ROOT` (the root spec node) always receives rank 0 because it has no
  parent and no `depends_on` or `input` references.
- A node with rank R can safely be processed after all entries with
  rank < R have been processed.
- Entries with the same rank have no dependency relationship with each
  other and may be processed in any order (including in parallel).
- `cycle_participants` contains ALL entries that are part of a cycle,
  not just a representative. This is a consequence of collecting every
  entry whose rank was still increasing after N passes.
- Cycle detection requires no separate graph traversal; it is a natural
  side effect of the iterative rank propagation failing to converge.
- If `cycle_participants` is non-empty, the ranks of those entries are
  unreliable and should not be used for ordering decisions.
```
