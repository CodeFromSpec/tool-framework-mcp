---
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

### Format validation

For each `_node.md` file in the tree:
- Frontmatter is parseable (if present).
- The first heading matches the logical name derived from the
  node's filesystem path (name verification).
- `depends_on` entries resolve to existing nodes.
- `depends_on` entries do not point to ancestors or descendants.
- `external` file paths exist and fragments match (hash check).
- `outputs` paths pass path validation.
- `# Agent` section only on leaf nodes.
- Frontmatter fields (`depends_on`, `external`, `input`, `outputs`)
  only on leaf nodes.

### Cycle detection

Detect cycles across `depends_on`, `input`, `external`, and
inheritance (parent -> child). Any cycle is reported with the
full path of the cycle.

### Staleness detection

For each node with `outputs`, compute the current chain hash
and compare it with the hash in each artifact's artifact tag
(`code-from-spec: <name>@<hash>`). Report:
- `stale` — hash mismatch.
- `missing` — artifact file does not exist.
- `current` — hashes match (not included in report).

## Contracts

- Reports all errors found — does not stop at the first.
- Staleness check only runs for nodes that have `outputs`.
