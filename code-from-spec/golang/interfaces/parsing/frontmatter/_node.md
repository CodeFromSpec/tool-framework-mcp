---
depends_on:
  - ARTIFACT/golang/interfaces/os/path_utils
output: code-from-spec/golang/interfaces/parsing/frontmatter/output.md
---

# SPEC/golang/interfaces/parsing/frontmatter

Parses structured metadata from the top of spec node
files.

# Public

## Package

`package frontmatter`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/frontmatter"`

## Interface

```go
type Frontmatter struct {
	DependsOn []string
	Input     string
	Output    string
}

func FrontmatterParse(filePath pathutils.PathCfs) (*Frontmatter, error)
```

All fields default to empty (empty slice, empty string)
when absent from the YAML.

### Errors

- `ErrFileUnreadable`: the file cannot be opened or
  read.
- `ErrMalformedYAML`: the content between `---`
  delimiters is not valid YAML, or an opening `---` is
  found but no closing `---` follows.
- Propagated errors from `file` package.

# Agent

Generate an interface specification document listing
the package, import path, struct definition, and
function signatures.
