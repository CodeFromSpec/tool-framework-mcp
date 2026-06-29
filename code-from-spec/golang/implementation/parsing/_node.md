# SPEC/golang/implementation/parsing

Parsing components for spec node files: frontmatter
extraction, body structure parsing, text normalization,
and logical name resolution.

# Public

## Package

`package parsing`

## Interface

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsing"`

### Text normalization

```go
func NormalizeText(rawString string) string
```

Pure function. Trims leading/trailing whitespace
(U+0020 and U+0009 only), collapses internal
whitespace runs to a single space, applies Unicode
simple case folding.

| Input | Output |
|---|---|
| `"  Interface  "` | `"interface"` |
| `"PUBLIC"` | `"public"` |
| `"Straße"` | `"strasse"` |
| `""` | `""` |

### Node parsing

```go
type NodeFrontmatter struct {
    DependsOn []string
    Input     *string
    Output    *string
}

type NodeSubsection struct {
    Heading    string
    RawHeading string
    Content    []string
}

type NodeSection struct {
    Heading     string
    RawHeading  string
    Content     []string
    Subsections []*NodeSubsection
}

type Node struct {
    Reference   CfsReference     // mandatory
    Frontmatter *NodeFrontmatter // nil if absent
    NameSection NodeSection      // mandatory
    Public      *NodeSection     // nil if absent
    Agent       *NodeSection     // nil if absent
    Private     *NodeSection     // nil if absent
}

func ParseNode(logicalName string) (*Node, error)
```

`NodeFrontmatter` fields are nil when absent from the
YAML. `DependsOn` defaults to nil (not empty slice)
when absent.

`Heading` is the normalized form (after `NormalizeText`),
used for comparisons and lookups. `RawHeading` is the
original heading line as it appears in the file (including
`#` prefix and closing `##` if present), preserved for
hashing.

`Content` is a list of lines between the heading and
the next structural heading (or end of file). Lines do
not include line terminators.

#### ParseNode

Opens a spec node file, extracts frontmatter and body,
and returns a structured representation. The file is
opened and read once.

Errors:
- `ErrNotASpecReference`
- `ErrHasQualifier`
- `ErrFileUnreadable`
- `ErrMalformedYAML`
- `ErrUnexpectedContentBeforeFirstHeading`
- `ErrNodeNameDoesNotMatch`
- `ErrDuplicatePublicSection`
- `ErrDuplicateAgentSection`
- `ErrDuplicatePrivateSection`
- `ErrUnrecognizedSection`
- `ErrDuplicateSubsection`

### CFS references

```go
type CfsNodeType int

const (
    CfsNodeTypeSpec     CfsNodeType = iota
    CfsNodeTypeArtifact
    CfsNodeTypeExternal
)

type CfsReference struct {
    NodeType    CfsNodeType
    LogicalName string
    Qualifier   *string
    Path        string
    ParentName  *string
}

func CfsReferenceFromName(logicalName string) (*CfsReference, error)
func CfsReferenceFromPath(cfsPath oslayer.CfsPath) (*CfsReference, error)
```

#### CfsReference fields

- **NodeType** — `CfsNodeTypeSpec`, `CfsNodeTypeArtifact`,
  or `CfsNodeTypeExternal`.
- **LogicalName** — the unqualified logical name
  including the prefix. For `SPEC/x/y(z)`, LogicalName
  is `SPEC/x/y`. For `ARTIFACT/x`, LogicalName is
  `ARTIFACT/x`. For `EXTERNAL/f.go`, LogicalName is
  `EXTERNAL/f.go`.
- **Qualifier** — nil if absent. For `SPEC/x/y(z)`,
  Qualifier points to `"z"`.
- **Path** — resolved file path as a CfsPath value:
  - SPEC: the `_node.md` path
    (e.g. `code-from-spec/x/y/_node.md`).
  - EXTERNAL: the file path relative to project root
    (e.g. `README.md`).
  - ARTIFACT: the value of `output` from the generator
    node's frontmatter (e.g. `internal/foo/foo.go`).
- **ParentName** — nil for root SPEC nodes (direct
  children of `code-from-spec/`, e.g. `SPEC/golang`)
  and EXTERNAL references. For non-root SPEC nodes,
  the parent's logical name (e.g. `SPEC/x` for
  `SPEC/x/y`). For ARTIFACT references, the generator
  node's logical name (e.g. `SPEC/x/y` for
  `ARTIFACT/x/y`).

#### CfsReferenceFromName

Parses a logical name string into a fully resolved
`CfsReference` struct. For ARTIFACT/ references, reads
the generator node's frontmatter via `ParseNode` to
resolve the output path.

Errors:
- `ErrUnrecognizedPrefix`: the string does not start
  with `SPEC/`, `ARTIFACT/`, or `EXTERNAL/`. Bare
  `SPEC` (without a trailing slash) is not valid.
- `ErrInvalidName`: the path portion is empty or
  invalid after stripping the prefix.
- `ErrNoOutput`: an ARTIFACT reference's generator node
  has no `output` field in its frontmatter.
- Propagated errors from `ParseNode`.

#### CfsReferenceFromPath

Reverse resolution: takes a CfsPath value like
`code-from-spec/x/y/_node.md` and returns a
`CfsReference` with NodeType = `CfsNodeTypeSpec`,
fully resolved.

Errors:
- `ErrInvalidPath`: the path does not match the
  expected `code-from-spec/.../_node.md` pattern.
