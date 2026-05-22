---
outputs:
  - id: logical_names
    path: code-from-spec/functional/logical_names/output.md
---

# ROOT/functional/logical_names

Maps logical names to file paths and provides utilities for
navigating the spec tree hierarchy.

# Public

## Logical name format

Logical names use two prefixes:

- `ROOT/` — references a spec node.
- `ARTIFACT/` — references a generated artifact by node and id.

An optional parenthetical qualifier targets a specific part:
- `ROOT/x/y(z)` — the `## z` subsection of `# Public`.
- `ARTIFACT/x/y(id)` — the artifact with the given id.

## Path resolution

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

## Parent navigation

Every `ROOT/` node except `ROOT` itself has a parent:

| Logical name | Parent |
|---|---|
| `ROOT/x` | `ROOT` |
| `ROOT/x/y` | `ROOT/x` |
| `ROOT/x/y(z)` | `ROOT/x` |

`ARTIFACT/` names do not participate in parent navigation.

## Qualifier extraction

| Logical name | Has qualifier | Qualifier |
|---|---|---|
| `ROOT/x(y)` | yes | `y` |
| `ARTIFACT/x(y)` | yes | `y` |
| `ROOT/x` | no | — |

## Contracts

- All returned file paths use forward slashes as separators,
  regardless of the operating system.
- All functions are pure — no I/O, no errors.
- Unrecognized prefixes return failure (false).
