# ROOT/golang/implementation/internal/logical_names

Centralizes conversion between logical names and file paths.

# Public

## Package

`package logicalnames`

## Interface

```go
func PathFromLogicalName(logicalName string) (string, bool)
func HasParent(logicalName string) (hasParent, ok bool)
func ParentLogicalName(logicalName string) (string, bool)
func HasQualifier(logicalName string) (hasQualifier, ok bool)
func QualifierName(logicalName string) (string, bool)
func IsArtifactRef(logicalName string) bool
func ArtifactRefParts(logicalName string) (nodePath string, artifactID string, ok bool)
```

### PathFromLogicalName

Resolves a logical name to a file path relative to the
project root. Returned paths always use forward slashes as
separators, regardless of the operating system. Use
`filepath.ToSlash` on the result before returning.

If the logical name has a parenthetical qualifier, it is
stripped before resolving the path. `ROOT/x(y)` resolves
to the same path as `ROOT/x`.

Only handles `ROOT/` references. Returns `("", false)` for
`ARTIFACT/` references — use `IsArtifactRef` and
`ArtifactRefParts` instead.

| Logical name | File path |
|---|---|
| `ROOT` | `code-from-spec/_node.md` |
| `ROOT/x/y` | `code-from-spec/x/y/_node.md` |
| `ROOT/x/y(z)` | `code-from-spec/x/y/_node.md` |

Rules:
- `ROOT` → `code-from-spec/_node.md`
- `ROOT/<path>` → `code-from-spec/<path>/_node.md`
- `ROOT/<path>(<qualifier>)` → `code-from-spec/<path>/_node.md`

### HasParent

Determines whether a logical name has a parent node.
Returns `(hasParent, ok)` where `ok` indicates whether
the input is a valid logical name.

| Logical name | hasParent | ok |
|---|---|---|
| `ROOT` | `false` | `true` |
| `ROOT/x` | `true` | `true` |
| `ROOT/x(y)` | `true` | `true` |
| `""` | `false` | `false` |

Rules:
- `ROOT` → no parent
- `ROOT/<path>` and `ROOT/<path>(<qualifier>)` → has parent
- `ARTIFACT/` references → not valid for this function (`false`, `false`)
- Anything else → not a valid logical name

### ParentLogicalName

Derives the parent's logical name from a node's logical
name. Returns `(parent, true)` on success, `("", false)`
if the node has no parent or input is invalid.

The qualifier is stripped before deriving the parent.

| Logical name | Parent |
|---|---|
| `ROOT/x` | `ROOT` |
| `ROOT/x/y` | `ROOT/x` |
| `ROOT/x/y(z)` | `ROOT/x` |

Rules:
- `ROOT/<path>` → strip last segment. If only one
  segment remains, parent is `ROOT`.
- `ROOT/<path>(<qualifier>)` → strip qualifier, then
  strip last segment.

### HasQualifier

Determines whether a logical name has a parenthetical
qualifier. Returns `(hasQualifier, ok)` where `ok`
indicates whether the input is a valid logical name.

| Logical name | hasQualifier | ok |
|---|---|---|
| `ROOT` | `false` | `true` |
| `ROOT/x` | `false` | `true` |
| `ROOT/x(y)` | `true` | `true` |
| `ROOT/x/y(z)` | `true` | `true` |
| `ARTIFACT/x(y)` | `true` | `true` |
| `""` | `false` | `false` |

### QualifierName

Extracts the qualifier from a logical name. Returns
`(qualifier, true)` on success, `("", false)` if there
is no qualifier.

| Logical name | Qualifier |
|---|---|
| `ROOT/x(y)` | `"y"` |
| `ROOT/x/y(z)` | `"z"` |
| `ARTIFACT/x(y)` | `"y"` |
| `ROOT/x` | `""`, `false` |
| `ROOT` | `""`, `false` |

### IsArtifactRef

Returns true if the logical name starts with `ARTIFACT/`.

### ArtifactRefParts

Parses an `ARTIFACT/` reference into its node path and
artifact ID. `ARTIFACT/x/y(id)` returns
(`code-from-spec/x/y/_node.md`, `"id"`, true).

The qualifier (artifact ID) is required for ARTIFACT/
references. Returns `("", "", false)` if the input is not
an ARTIFACT/ reference or has no qualifier.

| Logical name | nodePath | artifactID |
|---|---|---|
| `ARTIFACT/x(y)` | `code-from-spec/x/_node.md` | `"y"` |
| `ARTIFACT/x/y(z)` | `code-from-spec/x/y/_node.md` | `"z"` |
| `ARTIFACT/x` | `""` | `""` (false — no qualifier) |
| `ROOT/x(y)` | `""` | `""` (false — not ARTIFACT/) |

### LogicalNameFromPath

Derives the logical name from a file path relative to the
project root. This is the reverse of `PathFromLogicalName`.

```go
func LogicalNameFromPath(filePath string) (string, bool)
```

Only handles `_node.md` files under `code-from-spec/`.
Returns `("", false)` for paths that do not match.

| File path | Logical name |
|---|---|
| `code-from-spec/_node.md` | `ROOT` |
| `code-from-spec/x/_node.md` | `ROOT/x` |
| `code-from-spec/x/y/_node.md` | `ROOT/x/y` |

### Error handling

These are pure functions operating on strings. They do
not perform I/O or return errors.
`PathFromLogicalName` returns `(result, true)` on success
and `("", false)` if the input does not match any known
pattern.
