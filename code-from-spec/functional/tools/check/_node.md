---
outputs:
  - id: check
    path: code-from-spec/functional/tools/check/output.md
---

# ROOT/functional/tools/check

Validates the spec tree for format errors, circular references,
and artifact staleness.

# Public

## Behavior

### Input

No parameters. Scans the entire spec tree starting from
`code-from-spec/`.

### Output

A structured report with three categories:

| Category | Description |
|---|---|
| `format_errors` | Structural problems in spec nodes. |
| `circular_references` | Cycles in `depends_on`, `input`, `external`, or inheritance. |
| `staleness` | Artifacts whose chain hash differs from their artifact tag. |

### Format validation

For each `_node.md` file in the tree:
- Frontmatter is parseable (if present).
- The first heading matches the node's logical name.
- `depends_on` entries resolve to existing nodes.
- `depends_on` entries do not point to ancestors or descendants.
- `external` file paths exist and fragments match (hash check).
- `outputs` paths pass path validation.
- `# Agent` section only on leaf nodes.
- Frontmatter fields (`depends_on`, `external`, `input`, `outputs`)
  only on leaf nodes.

### Cycle detection

Detect cycles across `depends_on`, `input`, `external`, and
inheritance (parent → child). Any cycle is reported with the
full path of the cycle.

### Staleness detection

For each node with `outputs`, compute the current chain hash
and compare it with the hash in each artifact's artifact tag
(`code-from-spec: <name>@<hash>`). Report:
- `stale` — hash mismatch.
- `missing` — artifact file does not exist.
- `current` — hashes match (not included in report).

## Error conditions

| Condition | Description |
|---|---|
| Unreadable file | A spec node file cannot be read. |
| Parse failure | A spec node file has invalid structure. |
