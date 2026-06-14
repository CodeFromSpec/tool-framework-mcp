---
depends_on:
  - ROOT/functional/logic/spec_tree/scan
  - ROOT/functional/logic/spec_tree/validate
  - ROOT/functional/logic/utils/node_ranking
  - ROOT/functional/logic/chain/resolver
  - ROOT/functional/logic/chain/hash
  - ROOT/functional/logic/parsing/artifact_tag
  - ROOT/functional/logic/parsing/frontmatter
  - ROOT/functional/logic/parsing/node_parsing
  - ROOT/functional/logic/os/list_files
  - ROOT/functional/logic/os/path_utils(interface)
output: code-from-spec/functional/logic/mcp_tools/validate_specs/output.md
---

# ROOT/functional/logic/mcp_tools/validate_specs

Validates the spec tree for format errors, circular
references, and artifact staleness.

# Public

## Namespace

    namespace: mcpvalidatespecs

## Interface

```
record StalenessEntry
  node: string
  artifact_path: string
  status: string
  detail: string
  rank: integer

record ValidationReport
  format_errors: list of spectreevalidate.FormatError
  cycles: list of string
  staleness: list of StalenessEntry

function MCPValidateSpecs() -> ValidationReport
```

No parameters. Scans the entire spec tree starting from
`code-from-spec/`. Always returns a report — never
raises an error. Problems are collected in the report.

`StalenessEntry.status` is one of:
- `"missing"` — file does not exist.
- `"stale"` — hash mismatch.
- `"malformed tag"` — file exists but has no artifact
  tag or the tag cannot be parsed.

Entries where the hash matches are not included.

`StalenessEntry.rank` is the rank from `NodeRankCompute`.
Entries with equal rank have no dependency between them
and can be processed in parallel.

`cycles` is a flat list of logical names involved in
non-convergence during ranking (as returned by
`NodeRankCompute`).

# Agent

## Behavior

### Step 1 — Discover nodes

Call `SpecTreeScan()` to find all `_node.md` files. If
it fails, return a report with a single format error
(node = "", rule = "scan", detail = error message) and
empty cycles/staleness.

Also discover all subdirectories under `code-from-spec/`
using `ListFiles` (or equivalent). This list is needed
by `SpecTreeValidate` to detect subdirectories without
`_node.md`.

### Step 2 — Parse all nodes

For each discovered node:
- Call `FrontmatterParse` with the node's file path.
- Call `NodeParse` with the node's logical name.

If parsing fails for a node, record a FormatError
(node = logical name, rule = "parse", detail = error
message) and exclude the node from subsequent steps.

Cache the results — each node is parsed once and reused
by subsequent steps.

### Step 3 — Format validation

Build a list of `SpecTreeValidateInput` from the
successfully parsed nodes (logical name + frontmatter +
node). Call `SpecTreeValidate(entries, all_dirs)` with
the discovered subdirectories. Collect all returned
`FormatError` entries.

### Step 4 — Ranking and cycle detection

Skip this step if Step 2 or Step 3 produced any format
errors — ranking depends on valid frontmatter, so the
results would be unreliable.

Build a list of `NodeRankInput` from the successfully
parsed nodes (logical name + frontmatter). Call
`NodeRankCompute(entries)`.

If `NodeRankCompute` returns an UnresolvableReference
error, record it as a FormatError (node = "", rule =
"ranking", detail = error message). Staleness entries
fall back to alphabetical order by node logical name.

Otherwise, store the ranked entries and cycle
participants.

### Step 5 — Staleness detection

For each node that has `output` in its frontmatter,
in rank order (lowest rank first; alphabetical if no
ranking available):

1. Call `ChainResolve(logical_name)` to get the
   resolved Chain. If it fails, record a StalenessEntry
   with status = "missing" and detail = error message,
   and continue.

2. Call `ChainHashCompute(chain)` to get the current
   chain hash. If it fails, record a StalenessEntry
   with status = "missing" and detail = error message,
   and continue.

3. Construct a `PathCfs` from `frontmatter.output`.
   Call `ArtifactTagExtract` with the path.
   - If FileUnreadable: record StalenessEntry with
     status = "missing", detail describing the reason.
   - If NoTagFound or MalformedTag: record
     StalenessEntry with status = "malformed tag",
     detail describing the reason.
   - If the tag's hash does not match the chain hash:
     record StalenessEntry with status = "stale",
     detail showing file hash vs expected hash.
   - If the hash matches: skip (not included).

   Set `artifact_path` from `frontmatter.output`.
   Set `rank` from the node's rank (from Step 4, or 0
   if no ranking available).

### Step 6 — Assemble report

Return `ValidationReport` with:
- `format_errors`: all FormatErrors from Steps 2, 3, 4
- `cycles`: cycle participant logical names from Step 4
  (empty list if no cycles or ranking skipped)
- `staleness`: all StalenessEntries from Step 5, ordered
  by rank ascending then node logical name ascending

## Contracts

- Always returns a report — never raises an error.
- Reports all problems found — does not stop at the
  first.
- Staleness check runs for all nodes with output,
  even if format errors exist for other nodes.
- Staleness entries include rank for parallel processing.
- Each node is parsed at most once.
