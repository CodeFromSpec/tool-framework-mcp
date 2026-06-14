---
depends_on:
  - SPEC/golang/dependencies/goccy-go-yaml
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/os/file_reader
  - ARTIFACT/golang/interfaces/os/path_utils
input: ARTIFACT/functional/logic/parsing/frontmatter
output: internal/frontmatter/frontmatter.go
---

# SPEC/golang/implementation/parsing/frontmatter

# Agent

Implement the pseudocode from the input as a Go package.

## Go-specific guidance

- Use `github.com/goccy/go-yaml` for YAML unmarshalling.
  Define an unexported struct with `yaml` tags to map YAML
  keys to Go fields, then convert to the exported types.
- Use the `filereader` package for file I/O: `FileOpen`,
  `FileReadLine`, `FileClose`.
- Error wrapping: wrap all errors with `fmt.Errorf` using
  `%w` so callers can match with `errors.Is()`.
- The package name should be `frontmatter`.
