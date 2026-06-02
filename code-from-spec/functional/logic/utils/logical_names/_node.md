---
depends_on:
  - ROOT/functional/logic/os/path_utils(interface)
output: code-from-spec/functional/logic/utils/logical_names/output.md
---

# ROOT/functional/logic/utils/logical_names

Maps logical names to file paths and provides utilities for
navigating the spec tree hierarchy.

# Public

## Interface

```
function LogicalNameToPath(logical_name: string) -> pathutils.PathCfs
  errors:
    - UnsupportedReference: the logical name is not a
      ROOT/ reference (neither ROOT nor ROOT/...).
```

Converts a `ROOT/` logical name to the `PathCfs` of the
corresponding `_node.md` file. Strips any qualifier before
resolving. Only accepts `ROOT/` references (including
`ROOT` itself).

```
function LogicalNameFromPath(cfs_path: pathutils.PathCfs) -> string
  errors:
    - InvalidPath: the path is not a _node.md file
      under code-from-spec/.
```

Derives the `ROOT/` logical name from a `_node.md` file
path. The inverse of `LogicalNameToPath`. Always returns
a `ROOT/` reference.

```
function LogicalNameGetParent(logical_name: string) -> string
  errors:
    - NoParent: the logical name is ROOT itself.
    - NotARootReference: the logical name is not a
      ROOT/ reference (neither ROOT nor ROOT/...).
```

Returns the logical name of the parent node. Strips any
qualifier before computing the parent. Only accepts
`ROOT/` references (including `ROOT` itself, which
returns NoParent).

```
function LogicalNameGetQualifier(logical_name: string) -> optional string
```

Extracts the parenthetical qualifier from a logical name.
Returns absent if no qualifier is present. Works with both
`ROOT/` and `ARTIFACT/` references, though `ARTIFACT/`
references do not use qualifiers and always return absent.
For example, `ROOT/x/y(z)` → `z`; `ROOT/x/y` → absent.

```
function LogicalNameStripQualifier(logical_name: string) -> string
```

Returns the logical name without the parenthetical
qualifier. If no qualifier is present, returns the input
unchanged. Works with both `ROOT/` and `ARTIFACT/`
references, though `ARTIFACT/` references do not use
qualifiers. For example, `ROOT/x/y(z)` → `ROOT/x/y`;
`ROOT/x/y` → `ROOT/x/y`.

```
function LogicalNameHasParent(logical_name: string) -> boolean
```

Returns true if the logical name is a `ROOT/` reference
other than `ROOT` itself. Returns false for `ROOT`,
`ARTIFACT/` references, and unrecognized prefixes.

```
function LogicalNameHasQualifier(logical_name: string) -> boolean
```

Returns true if the logical name contains a parenthetical
qualifier. Works with both `ROOT/` and `ARTIFACT/`
references, though `ARTIFACT/` references do not use
qualifiers and always return false.

```
function LogicalNameIsArtifact(logical_name: string) -> boolean
```

Returns true if the logical name starts with `ARTIFACT/`.

```
function LogicalNameGetArtifactGenerator(logical_name: string) -> string
  errors:
    - NotAnArtifactReference: the logical name does not
      start with ARTIFACT/.
```

Returns the `ROOT/` logical name of the node that generates
the referenced artifact. Strips the `ARTIFACT/` prefix.
For example, `ARTIFACT/x/y` → `ROOT/x/y`.

### Logical name format

Logical names use two prefixes:

- `ROOT/` — references a spec node. The bare string `ROOT`
  (without a trailing slash) is a valid `ROOT/` reference —
  it refers to the root node.
- `ARTIFACT/` — references a generated artifact by node and id.

An optional parenthetical qualifier targets a specific part:
- `ROOT/x/y(z)` — the `## z` subsection of `# Public`.

`ARTIFACT/` references do not use qualifiers.

### Path resolution

`ROOT/` names resolve to `_node.md` files as `PathCfs` values:

| Logical name | PathCfs |
|---|---|
| `ROOT` | `code-from-spec/_node.md` |
| `ROOT/x/y` | `code-from-spec/x/y/_node.md` |
| `ROOT/x/y(z)` | `code-from-spec/x/y/_node.md` |

Qualifiers are stripped before resolving the path.

`ARTIFACT/` names cannot be fully resolved by this module —
the final artifact path lives in the generating node's
frontmatter, which requires I/O. To resolve an artifact
reference, the caller:

1. Calls `LogicalNameGetArtifactGenerator` to get the
   generating node's logical name
   (e.g. `ARTIFACT/x/y` → `ROOT/x/y`).
2. Calls `LogicalNameToPath` to get the generating node's
   `PathCfs`.
3. Reads the node's frontmatter and uses its output `path`
   to locate the artifact file.

### Parent navigation

Every `ROOT/` node except `ROOT` itself has a parent:

| Logical name | Parent |
|---|---|
| `ROOT/x` | `ROOT` |
| `ROOT/x/y` | `ROOT/x` |
| `ROOT/x/y(z)` | `ROOT/x` |

`ARTIFACT/` names do not participate in parent navigation.

### Reverse resolution

Given a `PathCfs`, derives the `ROOT/` logical name.
Only handles `_node.md` files under `code-from-spec/` —
always returns a `ROOT/` reference:

| PathCfs | Logical name |
|---|---|
| `code-from-spec/_node.md` | `ROOT` |
| `code-from-spec/x/_node.md` | `ROOT/x` |
| `code-from-spec/x/y/_node.md` | `ROOT/x/y` |

### Qualifier extraction and stripping

| Logical name | Has qualifier | Qualifier | Stripped |
|---|---|---|---|
| `ROOT/x(y)` | yes | `y` | `ROOT/x` |
| `ROOT/x` | no | — | `ROOT/x` |
| `ARTIFACT/x` | no | — | `ARTIFACT/x` |

# Agent

## Behavior

All functions are pure — no I/O. They perform string
manipulation on logical names and file paths.

## Contracts

- All returned file paths use forward slashes as separators,
  regardless of the operating system.
- All functions are pure — no I/O, no errors (except where
  noted in the interface).
- Unrecognized prefixes raise an error (for functions that
  declare errors) or return false (for boolean checks).
