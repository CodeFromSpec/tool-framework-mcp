# ROOT/golang/internal/frontmatter

Reads and parses the YAML frontmatter from spec node files.

# Public

## Package

`package frontmatter`

## Dependencies

Uses `github.com/goccy/go-yaml` for YAML parsing.

## Interface

```go
type Output struct {
    ID   string
    Path string
}

type ExternalFragment struct {
    Description string
    Lines       string
    Hash        string
}

type External struct {
    Path      string
    Fragments []ExternalFragment
}

type Frontmatter struct {
    DependsOn []string
    External  []External
    Input     string
    Outputs   []Output
}

var (
    ErrRead             = errors.New("error reading file")
    ErrFrontmatterParse = errors.New("error parsing frontmatter")
)

func ParseFrontmatter(filePath string) (*Frontmatter, error)
```

`ParseFrontmatter` reads the file, extracts the frontmatter block,
and returns the parsed result. If the file has no frontmatter
delimiters, it returns an empty `Frontmatter` (not an error).

Errors returned by `ParseFrontmatter` wrap the sentinel with
context (file path, underlying error) using `fmt.Errorf`, so
callers can match with `errors.Is()`.

### Error handling

All errors wrap a sentinel so callers can use `errors.Is()`:

| Sentinel | Returned when |
|---|---|
| `ErrRead` | The file cannot be read. |
| `ErrFrontmatterParse` | The YAML frontmatter is malformed. |
