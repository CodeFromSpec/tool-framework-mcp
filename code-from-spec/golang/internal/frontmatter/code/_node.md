---
depends_on:
  - ROOT/dependencies/goccy-go-yaml
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: frontmatter
    path: internal/frontmatter/frontmatter.go
---

# ROOT/golang/internal/frontmatter/code

Generates the frontmatter package implementation.

# Agent

## Implementation

The frontmatter is the YAML block between the first `---` and the
second `---` at the top of the file. Everything after the second
`---` is ignored.

If no `---` delimiters are found, return an empty `Frontmatter`
struct (not an error).

Fields extracted:

| Field | Type | Description |
|---|---|---|
| `depends_on` | []string | Logical names of dependencies. |
| `external` | []External | External file references. |
| `input` | string | Single ARTIFACT/ logical name. |
| `outputs` | []Output | Output file mappings. |

Unknown fields are ignored.

Each `external` entry has:

| YAML key | Type | Required | Description |
|---|---|---|---|
| `path` | string | yes | Path to the external file. |
| `fragments` | []ExternalFragment | no | List of referenced fragments. |

Each fragment has:

| YAML key | Type | Required | Description |
|---|---|---|---|
| `description` | string | no | Description of the fragment. |
| `lines` | string | yes | Line range reference. |
| `hash` | string | yes | Content hash for staleness detection. |

Each `outputs` entry has:

| YAML key | Type | Description |
|---|---|---|
| `id` | string | Identifier for the output. |
| `path` | string | File path of the output. |

The parser reads line by line, extracts the frontmatter block, and
stops as soon as the closing `---` is found. The file body is never
read.
