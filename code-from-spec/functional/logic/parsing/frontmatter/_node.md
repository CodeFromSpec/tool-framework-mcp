---
depends_on:
  - ROOT/functional/logic/os/file_reader
  - ROOT/functional/logic/os/path_utils(interface)
outputs:
  - id: frontmatter
    path: code-from-spec/functional/logic/parsing/frontmatter/output.md
---

# ROOT/functional/logic/parsing/frontmatter

Parses structured metadata from the top of spec node files.

Review status: pending

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

function ParseFrontmatter(file_path: PathCfs) -> Frontmatter
  errors:
    - file unreadable: the file cannot be opened or read.
    - malformed YAML: the content between --- delimiters is not valid YAML.
```

All fields default to empty (empty list, empty string) when
absent from the YAML.

# Agent

## Behavior

Open the file with `file_reader`. Read line by line using
`ReadLine`. Close the reader when done.

The frontmatter is an optional YAML block delimited by `---` at
the top of a file. If present, it contains metadata fields that
the framework uses for dependency resolution, artifact tracking,
and external file references.

If the first line is not `---`, return an empty Frontmatter
record. This is not an error.

## Contracts

- The parser reads only the frontmatter block. It never reads
  the file body.
- Unknown YAML fields are silently ignored.
- All recognized fields are optional. An empty frontmatter
  block (`---\n---`) produces an empty record.
