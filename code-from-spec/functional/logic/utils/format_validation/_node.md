---
depends_on:
  - ROOT/functional/logic/os/file_reader
  - ROOT/functional/logic/utils/frontmatter
  - ROOT/functional/logic/utils/logical_names
  - ROOT/functional/logic/utils/name_normalization
  - ROOT/functional/logic/utils/node_parsing
  - ROOT/functional/logic/os/path_utils
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: format_validation
    path: code-from-spec/functional/logic/utils/format_validation/output.md
---

# ROOT/functional/logic/utils/format_validation

Linter for spec nodes. Reads every node in the spec tree,
parses its frontmatter and body, and checks structural rules
defined by the framework. Reports all violations found.

Review status: pending

# Public

## Interface

```
record FormatError
  node: string
  rule: string
  detail: string

function ValidateFormat(discovered_nodes) -> list of FormatError
  errors:
    - unreadable node: a node file cannot be read.
```

`discovered_nodes` is a list of discovered nodes (logical
name + file path), as returned by `node_discovery`.

For each node, the function reads and parses the file using
`frontmatter` and `node_parsing`, then checks all rules.
A node is classified as leaf or intermediate by checking
whether any other discovered node is a child of it.

Returns a list of format errors (empty if all nodes are
valid).

# Agent

## Behavior

For each discovered node:

1. Open the file with `file_reader`. Close the reader
   when done with each file.
2. Parse frontmatter using `frontmatter`.
3. Parse body using `node_parsing`.
4. Run all validation rules below. Collect all errors —
   do not stop at the first.

A node has children if any other discovered node's logical
name starts with its logical name followed by `/`.

### Validation rules

#### Name verification

The first heading in the file (`# <name>`) must match the
logical name derived from the node's filesystem path using
`logical_names` reverse resolution. Comparison uses
`name_normalization`.

#### Frontmatter field restrictions

The fields `depends_on`, `external`, `input`, and `outputs`
are only permitted on nodes without children. If a node
with children has any of these fields, it is a format error.

#### Agent section restrictions

Only nodes without children may have a `# Agent` section.
If a node with children has `# Agent`, it is a format error.

#### Dependency targets

Each `depends_on` entry must:
- Resolve to an existing `_node.md` file using
  `logical_names`.
- Not point to an ancestor of the current node (redundant —
  ancestor content is already inherited).
- Not point to a descendant of the current node (would
  create a circular dependency).

Ancestor/descendant is determined by comparing logical name
prefixes.

#### External file existence

Each `external` entry's `path` must point to an existing
file. If `fragments` are declared, read the file using
`file_reader`, extract the declared line range, compute
SHA-1 + base64url hash, and verify it matches the declared
`hash`.

#### Output path validation

Each `outputs` entry's `path` must pass `path_validation`
(no traversal, no absolute paths, within project root).

#### Duplicate public subsections

Within a `# Public` section, all `##` subsection headings
must be unique after `name_normalization`.

## Contracts

- All nodes are validated — not just leaf nodes.
- All errors are collected — validation does not stop at
  the first error.
