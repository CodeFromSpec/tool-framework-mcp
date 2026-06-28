---
depends_on:
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/parsing/frontmatter
output: internal/logicalnames/logicalnames.go
---

# SPEC/golang/implementation/utils/logical_names

Parses logical names into a structured representation
with all resolved information: type, unqualified name,
qualifier, file path, and parent.

# Public

## Package

`package logicalnames`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/logicalnames"`

## Interface

```go
type NodeType int

const (
    NodeTypeSpec     NodeType = iota
    NodeTypeArtifact
    NodeTypeExternal
)

type LogicalName struct {
    Type      NodeType
    Name      string
    Qualifier *string
    Path      string
    Parent    *string
}

func LogicalNameParse(logicalName string) (*LogicalName, error)
func LogicalNameFromPath(cfsPath pathutils.PathCfs) (*LogicalName, error)
```

### LogicalName fields

- **Type** — `NodeTypeSpec`, `NodeTypeArtifact`, or
  `NodeTypeExternal`.
- **Name** — the unqualified logical name including the
  prefix. For `SPEC/x/y(z)`, Name is `SPEC/x/y`. For
  `ARTIFACT/x`, Name is `ARTIFACT/x`. For
  `EXTERNAL/f.go`, Name is `EXTERNAL/f.go`.
- **Qualifier** — nil if absent. For `SPEC/x/y(z)`,
  Qualifier points to `"z"`.
- **Path** — resolved file path as a PathCfs value:
  - SPEC: the `_node.md` path
    (e.g. `code-from-spec/x/y/_node.md`).
  - EXTERNAL: the file path relative to project root
    (e.g. `README.md`).
  - ARTIFACT: the value of `output` from the generator
    node's frontmatter (e.g. `internal/foo/foo.go`).
- **Parent** — nil for root SPEC nodes (direct children
  of `code-from-spec/`, e.g. `SPEC/golang`) and
  EXTERNAL references. For non-root SPEC nodes, the
  parent's logical name (e.g. `SPEC/x` for
  `SPEC/x/y`). For ARTIFACT references, the generator
  node's logical name (e.g. `SPEC/x/y` for
  `ARTIFACT/x/y`).

### LogicalNameParse

Parses a logical name string into a fully resolved
`LogicalName` struct.

Errors:
- `ErrUnrecognizedPrefix`: the string does not start
  with `SPEC/`, `ARTIFACT/`, or `EXTERNAL/`. Bare
  `SPEC` (without a trailing slash) is not valid.
- `ErrInvalidName`: the path portion is empty or
  invalid after stripping the prefix.
- `ErrNoOutput`: an ARTIFACT reference's generator node
  has no `output` field in its frontmatter.
- Propagated errors from `frontmatter.FrontmatterParse`.

### LogicalNameFromPath

Reverse resolution: takes a PathCfs value like
`code-from-spec/x/y/_node.md` and returns a
`LogicalName` with Type = `NodeTypeSpec`, fully
resolved.

Errors:
- `ErrInvalidPath`: the path does not match the
  expected `code-from-spec/.../_node.md` pattern.

# Agent

Implement the logical names component as a Go package.

## Logic

### LogicalNameParse(logicalName: string) -> *LogicalName

1. **Extract qualifier.** Find the first `(` in
   logicalName. If found, find the matching `)` after
   it. Extract the text between them as the qualifier
   string. Let `stripped` be the portion before `(`.
   If no `(`, qualifier is nil and `stripped` is
   logicalName unchanged.

2. **Classify by prefix.** Examine `stripped`:

   a. If `stripped` starts with `"SPEC/"`:
      Let `relative` = stripped with "SPEC/" removed.
      If `relative` is empty, raise ErrInvalidName.
      Let path = "code-from-spec/" + relative +
      "/_node.md".
      Compute parent: find the last "/" in relative.
      If no "/" found, parent is nil (this is a root
      node — a direct child of code-from-spec/).
      If "/" found, parent = "SPEC/" + substring
      before the last "/".
      Return LogicalName with Type = NodeTypeSpec,
      Name = stripped, Qualifier = qualifier,
      Path = path, Parent = parent (nil or pointer).

   c. If `stripped` starts with `"ARTIFACT/"`:
      Let `relative` = stripped with "ARTIFACT/" removed.
      If `relative` is empty, raise ErrInvalidName.
      Let generatorName = "SPEC/" + relative.
      Let generatorPath = "code-from-spec/" + relative
      + "/_node.md".
      Call FrontmatterParse(PathCfs{Value:
      generatorPath}). If it fails, propagate the error.
      If frontmatter.Output is empty, raise ErrNoOutput.
      Return LogicalName with Type = NodeTypeArtifact,
      Name = stripped, Qualifier = nil,
      Path = frontmatter.Output,
      Parent = pointer to generatorName.

   d. If `stripped` starts with `"EXTERNAL/"`:
      Let `relative` = stripped with "EXTERNAL/" removed.
      If `relative` is empty, raise ErrInvalidName.
      Return LogicalName with Type = NodeTypeExternal,
      Name = stripped, Qualifier = nil,
      Path = relative, Parent = nil.

   e. Otherwise: raise ErrUnrecognizedPrefix.

### LogicalNameFromPath(cfsPath: PathCfs) -> *LogicalName

1. Let `value` = cfsPath.Value.

2. If `value` does not start with "code-from-spec/",
   raise ErrInvalidPath.

3. If `value` does not end with "/_node.md",
   raise ErrInvalidPath.

4. Remove "code-from-spec/" prefix and "/_node.md"
   suffix. Let `relative` = the remainder.
   If `relative` is empty, raise ErrInvalidPath.

5. Let name = "SPEC/" + relative.
   Compute parent: find the last "/" in relative.
   If no "/" found, parent is nil (root node).
   If "/" found, parent = "SPEC/" + substring before
   the last "/".

6. Return LogicalName with Type = NodeTypeSpec,
   Name = name, Qualifier = nil,
   Path = value, Parent = parent (nil or pointer).

## Go-specific guidance

- Use the `pathutils` package for `PathCfs`.
- Use the `frontmatter` package for `FrontmatterParse`.
- Define sentinel errors with `errors.New`:
  `ErrUnrecognizedPrefix`, `ErrInvalidName`,
  `ErrNoOutput`, `ErrInvalidPath`.
- Wrap propagated errors with `fmt.Errorf` + `%w`.
- For the qualifier pointer, use a helper like
  `func stringPtr(s string) *string { return &s }`.
- The package name should be `logicalnames`.

# Private

## Decisions

### Struct-based interface replacing 12 functions

The original interface had 12 individual functions
(LogicalNameToPath, LogicalNameIsSpec,
LogicalNameGetParent, etc.). Replaced with a single
`LogicalNameParse` returning a `LogicalName` struct.
Eliminates repeated parsing of the same string and
simplifies consumer code (one call instead of 3-4).
`LogicalNameFromPath` kept as a separate constructor
for reverse resolution.

### ARTIFACT Path resolves via frontmatter I/O

`LogicalNameParse("ARTIFACT/x")` reads the generator
node's frontmatter to populate `Path` with the actual
output file path. This means Parse does I/O for
ARTIFACT types (not for SPEC or EXTERNAL). The
trade-off is that consumers no longer need to manually
resolve artifact paths — `ln.Path` is always usable.
Consequence: callers that only need type classification
(node_ranking, validate) should use string prefix
checks for ARTIFACT/EXTERNAL instead of Parse to avoid
unnecessary I/O.

### Bare SPEC no longer valid (v5 multi-root)

In v4, `SPEC` (without slash) referred to the single
root node at `code-from-spec/_node.md`. In v5,
`code-from-spec/` is not a node — root nodes are
direct children (e.g. `SPEC/golang`). Bare `SPEC`
now returns `ErrUnrecognizedPrefix`.
