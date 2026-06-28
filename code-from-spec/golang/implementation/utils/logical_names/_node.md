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

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/logicalnames"`

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
- **Parent** — nil for root SPEC nodes and EXTERNAL
  references. For non-root SPEC nodes, the parent's
  logical name (e.g. `SPEC/x` for `SPEC/x/y`). For
  ARTIFACT references, the generator node's logical
  name (e.g. `SPEC/x/y` for `ARTIFACT/x/y`).

### LogicalNameParse

Parses a logical name string into a fully resolved
`LogicalName` struct.

Errors:
- `ErrUnrecognizedPrefix`: the string does not start
  with `SPEC/`, `SPEC` (exact), `ARTIFACT/`, or
  `EXTERNAL/`.
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

   a. If `stripped` is exactly `"SPEC"`:
      Return LogicalName with Type = NodeTypeSpec,
      Name = "SPEC", Qualifier = qualifier,
      Path = "code-from-spec/_node.md",
      Parent = nil.

   b. If `stripped` starts with `"SPEC/"`:
      Let `relative` = stripped with "SPEC/" removed.
      If `relative` is empty, raise ErrInvalidName.
      Let path = "code-from-spec/" + relative +
      "/_node.md".
      Compute parent: find the last "/" in relative.
      If no "/" found, parent = "SPEC".
      If "/" found, parent = "SPEC/" + substring
      before the last "/".
      Return LogicalName with Type = NodeTypeSpec,
      Name = stripped, Qualifier = qualifier,
      Path = path, Parent = pointer to parent.

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

3. If `value` is exactly "code-from-spec/_node.md":
   Return LogicalName with Type = NodeTypeSpec,
   Name = "SPEC", Qualifier = nil,
   Path = "code-from-spec/_node.md",
   Parent = nil.

4. If `value` does not end with "/_node.md",
   raise ErrInvalidPath.

5. Remove "code-from-spec/" prefix and "/_node.md"
   suffix. Let `relative` = the remainder.
   If `relative` is empty, raise ErrInvalidPath.

6. Let name = "SPEC/" + relative.
   Compute parent: find the last "/" in relative.
   If no "/" found, parent = "SPEC".
   If "/" found, parent = "SPEC/" + substring before
   the last "/".

7. Return LogicalName with Type = NodeTypeSpec,
   Name = name, Qualifier = nil,
   Path = value, Parent = pointer to parent.

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
