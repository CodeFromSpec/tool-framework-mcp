---
outputs:
  - id: frontmatter
    path: code-from-spec/functional/utils/frontmatter/output.md
---

# ROOT/functional/utils/frontmatter

Parses structured metadata from the top of spec node files.

# Public

## Interface

```
record ExternalFragment
  description: optional string
  lines: string
  hash: string

record External
  path: string
  fragments: optional list of ExternalFragment

record Output
  id: string
  path: string

record Frontmatter
  depends_on: list of strings
  external: list of External
  input: string
  outputs: list of Output

function ParseFrontmatter(file_path) -> frontmatter
  errors:
    - file unreadable: the file cannot be opened or read.
    - malformed YAML: the content between --- delimiters is not valid YAML.
```

All fields default to empty (empty list, empty string) when
absent from the YAML.

# Agent

## Behavior

The frontmatter is an optional YAML block delimited by `---` at
the top of a file. If present, it contains metadata fields that
the framework uses for dependency resolution, artifact tracking,
and external file references.

If the file has no `---` delimiters, return an empty Frontmatter
record. This is not an error.

## Contracts

- The parser reads only the frontmatter block. It never reads
  the file body.
- Unknown YAML fields are silently ignored.
- All recognized fields are optional. An empty frontmatter
  block (`---\n---`) produces an empty record.
