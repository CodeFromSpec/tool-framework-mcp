---
depends_on:
  - ROOT/external/goccy-go-yaml
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: frontmatter
    path: internal/frontmatter/frontmatter.go
---

# ROOT/tech_design/internal/frontmatter/implementation

Generates the frontmatter package implementation.

# Agent

## Implementation

The frontmatter is the YAML block between the first `---` and the
second `---` at the top of the file. Everything after the second
`---` is ignored.

Fields extracted:

| Field | Type | Description |
|---|---|---|
| `version` | int | Node version. Required. |
| `parent_version` | *int | Parent version. Nil if absent. |
| `subject_version` | *int | Subject version (test nodes). Nil if absent. |
| `depends_on` | []DependsOn | Cross-tree dependencies. |
| `implements` | []string | Output files. |

Unknown fields are ignored.

Each `depends_on` entry has:

| YAML key | Type | Required | Description |
|---|---|---|---|
| `path` | string | yes | Logical name of the dependency. |
| `version` | int | yes | Known version of the dependency. |

The parser reads line by line, extracts the frontmatter block, and
stops as soon as the closing `---` is found. The file body is never
read.
