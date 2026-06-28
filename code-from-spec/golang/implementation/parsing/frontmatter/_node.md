---
depends_on:
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/dependencies/goccy-go-yaml
  - SPEC/golang/implementation/os/path_utils
output: internal/frontmatter/frontmatter.go
---

# SPEC/golang/implementation/parsing/frontmatter

Parses structured metadata from the top of spec node
files.

# Public

## Package

`package frontmatter`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/frontmatter"`

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

Implement the frontmatter parser as a Go package.

## Logic

1. Call FileOpen(file_path, "read", 30000).
   If FileOpen raises any error, re-raise as
   ErrFileUnreadable (wrapping the original).

2. Call FileReadLine to read the first line.
   If EndOfFile is raised, call FileClose and return an
   empty Frontmatter record with
   depends_on = [], input = "", output = "".
   If the first line is not exactly "---", call
   FileClose and return an empty Frontmatter record.

3. Collect YAML lines:
   Initialize yaml_lines as an empty list of strings.
   Repeat:
     Call FileReadLine.
     If EndOfFile is raised:
       Call FileClose.
       Raise ErrMalformedYAML.
     If the line is exactly "---":
       Stop collecting.
     Else:
       Append the line to yaml_lines.

4. Call FileClose.

5. If yaml_lines is empty, return an empty Frontmatter
   record with depends_on = [], input = "", output = "".

6. Join yaml_lines into a single string, each line
   separated by a newline character. Parse the joined
   string as YAML. If parsing fails, raise ErrMalformedYAML.

7. From the parsed YAML, extract the following fields,
   ignoring all other keys:
   - depends_on: list of strings. If absent or null,
     use [].
   - input: string. If absent or null, use "".
   - output: string. If absent or null, use "".

8. Return a Frontmatter record with the extracted field
   values.

## Go-specific guidance

- Use `github.com/goccy/go-yaml` for YAML unmarshalling.
  Define an unexported struct with `yaml` tags to map YAML
  keys to Go fields, then convert to the exported types.
- Use the `file` package for file I/O: `FileOpen`,
  `FileReadLine`, `FileClose`.
- Error wrapping: wrap all errors with `fmt.Errorf` using
  `%w` so callers can match with `errors.Is()`.
- The package name should be `frontmatter`.
