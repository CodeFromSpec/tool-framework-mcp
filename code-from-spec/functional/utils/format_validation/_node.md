---
depends_on:
  - ROOT/functional/utils/logical_names
  - ROOT/functional/utils/frontmatter
  - ROOT/functional/utils/node_parsing
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: format_validation
    path: code-from-spec/functional/utils/format_validation/output.md
---

# ROOT/functional/utils/format_validation

Validates that spec nodes conform to the structural rules
defined by the framework.

# Public

## Behavior

### Input

A list of discovered nodes with their parsed frontmatter
and parsed body.

### Output

A list of format errors. Each error has:
- `node` — the logical name of the offending node.
- `rule` — which rule was violated.
- `detail` — human-readable description.

## Validation rules

### Name verification

The first heading in the file (`# <name>`) must match the
logical name derived from the node's filesystem path.
Comparison uses normalized names (trim, collapse whitespace,
case fold).

### Frontmatter field restrictions

The fields `depends_on`, `external`, `input`, and `outputs`
are only permitted on leaf nodes. If an intermediate or
root node has any of these fields, it is a format error.

### Agent section restrictions

Only leaf nodes may have a `# Agent` section. If a root or
intermediate node has `# Agent`, it is a format error.

### Dependency targets

Each `depends_on` entry must:
- Resolve to an existing `_node.md` file.
- Not point to an ancestor of the current node (redundant —
  ancestor content is already inherited).
- Not point to a descendant of the current node (would
  create a circular dependency).

### External file existence

Each `external` entry's `path` must point to an existing
file. If `fragments` are declared, each fragment's `hash`
must match the hash computed from the content at the
declared `lines` range.

### Output path validation

Each `outputs` entry's `path` must pass path validation
(no traversal, no absolute paths, within project root).

### Duplicate public subsections

Within a `# Public` section, all `##` subsection headings
must be unique after normalization.

## Error conditions

| Condition | Description |
|---|---|
| Name mismatch | Heading does not match filesystem-derived logical name. |
| Frontmatter on non-leaf | Leaf-only fields present on intermediate or root node. |
| Agent on non-leaf | `# Agent` section present on non-leaf node. |
| Invalid dependency target | `depends_on` points to ancestor, descendant, or non-existent node. |
| Missing external file | `external` path does not exist. |
| Fragment hash mismatch | Fragment content hash does not match declared hash. |
| Invalid output path | Output path fails path validation. |
| Duplicate subsection | Two `##` headings in `# Public` normalize to the same text. |
