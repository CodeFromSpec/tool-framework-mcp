---
depends_on:
  - SPEC/functional/logic/os/file
  - SPEC/functional/logic/os/path_utils(interface)
output: code-from-spec/functional/logic/parsing/frontmatter/output.md
---

# SPEC/functional/logic/parsing/frontmatter

Parses structured metadata from the top of spec node files.

# Public

## Namespace

    namespace: frontmatter

## Interface

```
record Frontmatter
  depends_on: list of strings
  input: string
  output: string

function FrontmatterParse(file_path: pathutils.PathCfs) -> Frontmatter
  errors:
    - FileUnreadable: the file cannot be opened or read.
    - MalformedYAML: the content between --- delimiters
      is not valid YAML.
    - (FileReader.*): propagated from FileOpen.
```

All fields default to empty (empty list, empty string) when
absent from the YAML.

# Agent

## Behavior

Open the file with `FileOpen`. Read line by line using
`FileReadLine`. Close the reader with `FileClose` when done.

The frontmatter is an optional YAML block delimited by `---` at
the top of a file. If present, it contains metadata fields that
the framework uses for dependency resolution and artifact
tracking.

The first line of the file must be exactly `---` (three
hyphens, nothing else — no leading or trailing whitespace).
The block ends at the next line that is exactly `---`.
Everything between is YAML. If the opening `---` is found
but the closing `---` is not, report "malformed YAML".

If the first line is not `---`, return an empty Frontmatter
record. This is not an error.

### YAML format

```yaml
---
depends_on:
  - SPEC/integrations/payments-api/create-transfer
  - ARTIFACT/extraction/email-templates
  - EXTERNAL/proto/payments/v1/transfers.proto
input: ARTIFACT/functional/transfers
output: internal/transfers/handler.go
---
```

YAML keys map directly to record fields. Keys not listed
in the Frontmatter record are silently ignored.

## Contracts

- The parser reads only the frontmatter block. It never reads
  the file body.
- Unknown YAML fields are silently ignored.
- All recognized fields are optional. An empty frontmatter
  block (`---\n---`) produces an empty record.
