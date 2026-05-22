# ROOT/golang/internal/parsenode

Parses the body of a spec node file, returning a structured
representation of all sections.

# Public

## Package

`package parsenode`

## Dependencies

- `github.com/yuin/goldmark` — CommonMark parsing of the body.
  The body is parsed into an AST; only level-1 and level-2
  headings are used as structural delimiters.

## Interface

```go
type Subsection struct {
	Heading string
	Content string
}

type Section struct {
	Heading     string
	Content     string
	Subsections []Subsection
}

type NodeBody struct {
	NameSection Section
	Public      *Section
	Private     []Section
}

var (
	ErrRead                 = errors.New("error reading file")
	ErrFrontmatterMissing   = errors.New("frontmatter not found")
	ErrUnexpectedContent    = errors.New("unexpected content before first heading")
	ErrInvalidNodeName      = errors.New("node name section does not match logical name")
	ErrDuplicatePublic      = errors.New("duplicate public section")
	ErrDuplicateSubsection  = errors.New("duplicate subsection in public")
)

func ParseNode(logicalName string) (*NodeBody, error)
```

`Public` is nil when no `# Public` section exists in the file.

Errors returned by `ParseNode` wrap the sentinel with context
(file path, underlying error) using `fmt.Errorf`, so callers
can match with `errors.Is()`.

### Error sentinels

| Sentinel | Returned when |
|---|---|
| `ErrRead` | The file cannot be read. |
| `ErrFrontmatterMissing` | No `---` delimiters found at the top of the file. |
| `ErrUnexpectedContent` | Non-heading content appears before the first level-1 heading. |
| `ErrInvalidNodeName` | The first level-1 heading does not match the logical name. |
| `ErrDuplicatePublic` | More than one level-1 heading normalizes to `public`. |
| `ErrDuplicateSubsection` | Two or more level-2 headings within `# Public` have the same normalized text. |
