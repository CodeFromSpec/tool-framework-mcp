---
depends_on:
  - ROOT/functional/utils/node_ranking
  - ROOT/functional/utils/format_validation
  - ROOT/functional/utils/logical_names
  - ROOT/functional/utils/node_discovery
  - ROOT/functional/utils/artifact_tag
  - ROOT/functional/utils/frontmatter
  - ROOT/functional/utils/name_normalization
  - ROOT/functional/utils/node_parsing
  - ROOT/functional/utils/path_validation
outputs:
  - id: validate_specs
    path: artifacts/functional/mcp_tools/validate_specs/output.md
---

# ROOT/functional/mcp_tools/validate_specs

Validates the spec tree for format errors, circular references,
and artifact staleness.

# Public

## Interface

```
record StalenessEntry
  node: string
  artifact_path: string
  status: string

record ValidationReport
  format_errors: list of FormatError
  circular_references: list of list of string
  staleness: list of StalenessEntry

function ValidateSpecs() -> ValidationReport
  errors:
    - unreadable file: a spec node file cannot be read.
    - parse failure: a spec node file has invalid structure.
```

No parameters. Scans the entire spec tree starting from
`code-from-spec/`.

# Agent

## Behavior

### Step 1 — Discover nodes

Use `node_discovery` to find all `_node.md` files in the
spec tree. Derive each node's logical name using
`logical_names` reverse resolution.

### Step 2 — Parse all nodes

For each discovered node:
- Use `frontmatter` to parse the YAML frontmatter.
- Use `node_parsing` to parse the body into sections.

Cache the results — each node is parsed once and reused
by subsequent steps.

### Step 3 — Format validation

Use `format_validation` to check each node against the
structural rules. This uses:
- `logical_names` to verify `depends_on` targets resolve.
- `name_normalization` to compare headings with logical
  names derived from filesystem paths.
- `path_validation` to verify `outputs` paths are safe.

Collect all `FormatError` entries.

### Step 4 — Ranking and cycle detection

Use `node_ranking` to rank all nodes and artifacts
and detect circular references. Pass the full set of
discovered nodes with their parsed frontmatter.

The ranking determines processing order for staleness
resolution: lower rank first. If cycles are detected,
report the cycle participants.

### Step 5 — Staleness detection

For each node with `outputs`, in rank order (lowest
rank first):
1. Compute the chain hash using the same algorithm as
   `load_chain` (SHA-1 of concatenated position hashes,
   base64url encoded).
2. For each output, use `artifact_tag` to extract the
   hash from the generated file.
3. Compare:
   - File does not exist → report `missing`.
   - File exists but has no artifact tag → report `missing`.
   - Hash mismatch → report `stale`.
   - Hash matches → skip (not included in report).

### Output

Assemble the `ValidationReport` with all collected
format errors, cycles, and staleness entries. Staleness
entries are ordered by rank (lowest first) so that the
caller can resolve them in dependency order.

## Contracts

- Reports all errors found — does not stop at the first.
- Staleness check only runs for nodes that have `outputs`.
- Nodes that fail format validation are still checked for
  staleness where possible.
