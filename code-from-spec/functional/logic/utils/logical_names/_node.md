---
depends_on:
  - SPEC/functional/logic/os/path_utils(interface)
output: code-from-spec/functional/logic/utils/logical_names/output.md
---

# SPEC/functional/logic/utils/logical_names

Maps logical names to file paths and provides utilities for
navigating the spec tree hierarchy.

# Public

## Namespace

    namespace: logicalnames

## Interface

```
function LogicalNameToPath(logical_name: string) -> pathutils.PathCfs
  errors:
    - UnsupportedReference: the logical name is not a
      SPEC/ reference (neither SPEC nor SPEC/...).
```

Converts a `SPEC/` logical name to the `PathCfs` of the
corresponding `_node.md` file. Strips any qualifier before
resolving. Only accepts `SPEC/` references (including
`SPEC` itself).

```
function LogicalNameFromPath(cfs_path: pathutils.PathCfs) -> string
  errors:
    - InvalidPath: the path is not a _node.md file
      under code-from-spec/.
```

Derives the `SPEC/` logical name from a `_node.md` file
path. The inverse of `LogicalNameToPath`. Always returns
a `SPEC/` reference.

```
function LogicalNameGetParent(logical_name: string) -> string
  errors:
    - NoParent: the logical name is SPEC itself.
    - NotASpecReference: the logical name is not a
      SPEC/ reference (neither SPEC nor SPEC/...).
```

Returns the logical name of the parent node. Strips any
qualifier before computing the parent. Only accepts
`SPEC/` references (including `SPEC` itself, which
returns NoParent). Always returns a `SPEC/` reference.

```
function LogicalNameGetQualifier(logical_name: string) -> optional string
```

Extracts the parenthetical qualifier from a logical name.
Returns absent if no qualifier is present. Works with
`SPEC/`, `ARTIFACT/`, and `EXTERNAL/` references, though
`ARTIFACT/` and `EXTERNAL/` references do not use
qualifiers and always return absent.
For example, `SPEC/x/y(z)` → `z`; `SPEC/x/y` → absent.

```
function LogicalNameStripQualifier(logical_name: string) -> string
```

Returns the logical name without the parenthetical
qualifier. If no qualifier is present, returns the input
unchanged. Works with `SPEC/`, `ARTIFACT/`, and
`EXTERNAL/` references.
For example, `SPEC/x/y(z)` → `SPEC/x/y`;
`SPEC/x/y` → `SPEC/x/y`.

```
function LogicalNameHasParent(logical_name: string) -> boolean
```

Returns true if the logical name is a `SPEC/` reference
other than `SPEC` itself. Returns false for `SPEC`,
`ARTIFACT/`, `EXTERNAL/`, and unrecognized prefixes.

```
function LogicalNameHasQualifier(logical_name: string) -> boolean
```

Returns true if the logical name contains a parenthetical
qualifier. Works with `SPEC/`, `ARTIFACT/`, and
`EXTERNAL/` references.

```
function LogicalNameIsArtifact(logical_name: string) -> boolean
```

Returns true if the logical name starts with `ARTIFACT/`.

```
function LogicalNameIsSpec(logical_name: string) -> boolean
```

Returns true if the logical name is exactly `SPEC` or
starts with `SPEC/`.

```
function LogicalNameIsExternal(logical_name: string) -> boolean
```

Returns true if the logical name starts with `EXTERNAL/`.

```
function LogicalNameGetArtifactGenerator(logical_name: string) -> string
  errors:
    - NotAnArtifactReference: the logical name does not
      start with ARTIFACT/.
```

Returns the `SPEC/` logical name of the node that generates
the referenced artifact. Strips the `ARTIFACT/` prefix and
prepends `SPEC/`.
For example, `ARTIFACT/x/y` → `SPEC/x/y`.

```
function LogicalNameExternalToPath(logical_name: string) -> pathutils.PathCfs
  errors:
    - NotAnExternalReference: the logical name does not
      start with EXTERNAL/.
```

Converts an `EXTERNAL/` logical name to a `PathCfs`.
Strips the `EXTERNAL/` prefix and returns the remainder
as a `PathCfs` (relative to the project root).
For example, `EXTERNAL/proto/v1/api.proto` →
`proto/v1/api.proto`.

### Logical name format

Logical names use three prefixes:

- `SPEC/` — references a spec node. The bare string `SPEC`
  (without a trailing slash) is a valid `SPEC/` reference —
  it refers to the root node.
- `ARTIFACT/` — references a generated artifact by node path.
- `EXTERNAL/` — references a plain project file by path
  relative to the project root.

An optional parenthetical qualifier targets a specific part:
- `SPEC/x/y(z)` — the `## z` subsection of `# Public`.

`ARTIFACT/` and `EXTERNAL/` references do not use qualifiers.

### Path resolution

`SPEC/` names resolve to `_node.md` files as `PathCfs` values:

| Logical name | PathCfs |
|---|---|
| `SPEC` | `code-from-spec/_node.md` |
| `SPEC/x/y` | `code-from-spec/x/y/_node.md` |
| `SPEC/x/y(z)` | `code-from-spec/x/y/_node.md` |

Qualifiers are stripped before resolving the path.

`EXTERNAL/` names resolve to project-root-relative paths:

| Logical name | PathCfs |
|---|---|
| `EXTERNAL/proto/v1/api.proto` | `proto/v1/api.proto` |
| `EXTERNAL/docker-compose.yaml` | `docker-compose.yaml` |

`ARTIFACT/` names cannot be fully resolved by this module —
the final artifact path lives in the generating node's
frontmatter, which requires I/O. To resolve an artifact
reference, the caller:

1. Calls `LogicalNameGetArtifactGenerator` to get the
   generating node's logical name
   (e.g. `ARTIFACT/x/y` → `SPEC/x/y`).
2. Calls `LogicalNameToPath` to get the generating node's
   `PathCfs`.
3. Reads the node's frontmatter and uses its output `path`
   to locate the artifact file.

### Parent navigation

Every `SPEC/` node except `SPEC` itself has a parent:

| Logical name | Parent |
|---|---|
| `SPEC/x` | `SPEC` |
| `SPEC/x/y` | `SPEC/x` |
| `SPEC/x/y(z)` | `SPEC/x` |

`ARTIFACT/` and `EXTERNAL/` names do not participate in
parent navigation.

### Reverse resolution

Given a `PathCfs`, derives the `SPEC/` logical name.
Only handles `_node.md` files under `code-from-spec/` —
always returns a `SPEC/` reference:

| PathCfs | Logical name |
|---|---|
| `code-from-spec/_node.md` | `SPEC` |
| `code-from-spec/x/_node.md` | `SPEC/x` |
| `code-from-spec/x/y/_node.md` | `SPEC/x/y` |

### Qualifier extraction and stripping

| Logical name | Has qualifier | Qualifier | Stripped |
|---|---|---|---|
| `SPEC/x(y)` | yes | `y` | `SPEC/x` |
| `SPEC/x` | no | — | `SPEC/x` |
| `ARTIFACT/x` | no | — | `ARTIFACT/x` |
| `EXTERNAL/x` | no | — | `EXTERNAL/x` |

# Agent

## Behavior

All functions are pure — no I/O. They perform string
manipulation on logical names and file paths.

## Contracts

- All returned file paths use forward slashes as separators,
  regardless of the operating system.
- All functions are pure — no I/O, no errors (except where
  noted in the interface).
- Unrecognized prefixes (including `ROOT/`) raise an error
  (for functions that declare errors) or return false (for
  boolean checks).
- Functions that return logical names always use the `SPEC/`
  prefix (canonical form).
