---
depends_on:
  - ROOT/golang/dependencies/goccy-go-yaml
input: ARTIFACT/functional/logic/parsing/frontmatter(frontmatter)
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: frontmatter
    path: internal/frontmatter/frontmatter.go
---

# ROOT/golang/implementation/internal/frontmatter/code

Generates the frontmatter package implementation in Go.

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `github.com/goccy/go-yaml` for YAML unmarshalling.
  Define an unexported struct with `yaml` tags to map YAML
  keys to Go fields, then convert to the exported types.
- Error wrapping: wrap all errors with `fmt.Errorf` using
  `%w` so callers can match with `errors.Is()`.
- Use `os.ReadFile` to read the file.
- Parse line by line using `bufio.Scanner` or
  `strings.Split` — do not parse the body, stop after the
  closing `---`.
