---
outputs:
  - id: logical_names
    path: code-from-spec/functional/utils/logical_names/output.md
---

# ROOT/functional/utils/logical_names

Maps logical names to file paths and provides utilities for
navigating the spec tree hierarchy.

# Public

## Interface

```
record ArtifactReference
  node_path: string
  artifact_id: string

function ResolvePath(logical_name) -> string
  errors:
    - unrecognized prefix: the logical name does not start with ROOT/ or ARTIFACT/.

function ResolveArtifactReference(logical_name) -> ArtifactReference
  errors:
    - unrecognized prefix: the logical name does not start with ARTIFACT/.
    - missing qualifier: the logical name has no parenthetical qualifier.

function GetParent(logical_name) -> string
  errors:
    - no parent: the logical name is ROOT itself.
    - not a ROOT reference: the logical name is an ARTIFACT/ reference.

function ReverseResolve(file_path) -> string
  errors:
    - invalid path: the path is not a _node.md file under code-from-spec/.

function ExtractQualifier(logical_name) -> optional string
```

### Logical name format

Logical names use two prefixes:

- `ROOT/` — references a spec node.
- `ARTIFACT/` — references a generated artifact by node and id.

An optional parenthetical qualifier targets a specific part:
- `ROOT/x/y(z)` — the `## z` subsection of `# Public`.
- `ARTIFACT/x/y(id)` — the artifact with the given id.

### Path resolution

`ROOT/` names resolve to `_node.md` files:

| Logical name | File path |
|---|---|
| `ROOT` | `code-from-spec/_node.md` |
| `ROOT/x/y` | `code-from-spec/x/y/_node.md` |
| `ROOT/x/y(z)` | `code-from-spec/x/y/_node.md` |

Qualifiers are stripped before resolving the path.

`ARTIFACT/` names cannot be resolved to a static path — they
require reading the target node's frontmatter to find the
output with the matching id. The resolution function returns
the node path and artifact id as separate values.

### Parent navigation

Every `ROOT/` node except `ROOT` itself has a parent:

| Logical name | Parent |
|---|---|
| `ROOT/x` | `ROOT` |
| `ROOT/x/y` | `ROOT/x` |
| `ROOT/x/y(z)` | `ROOT/x` |

`ARTIFACT/` names do not participate in parent navigation.

### Reverse resolution

Given a file path relative to the project root, derives the
logical name:

| File path | Logical name |
|---|---|
| `code-from-spec/_node.md` | `ROOT` |
| `code-from-spec/x/_node.md` | `ROOT/x` |
| `code-from-spec/x/y/_node.md` | `ROOT/x/y` |

Only handles `_node.md` files under `code-from-spec/`.

### Qualifier extraction

| Logical name | Has qualifier | Qualifier |
|---|---|---|
| `ROOT/x(y)` | yes | `y` |
| `ARTIFACT/x(y)` | yes | `y` |
| `ROOT/x` | no | — |

# Agent

## Behavior

All functions are pure — no I/O. They perform string
manipulation on logical names and file paths.

## Contracts

- All returned file paths use forward slashes as separators,
  regardless of the operating system.
- All functions are pure — no I/O, no errors (except where
  noted in the interface).
- Unrecognized prefixes return failure (false).
