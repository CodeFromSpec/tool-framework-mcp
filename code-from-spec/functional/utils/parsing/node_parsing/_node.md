---
depends_on:
  - ROOT/functional/utils/logical_names
  - ROOT/functional/utils/file_reader
  - ROOT/functional/utils/parsing/name_normalization
outputs:
  - id: node_parsing
    path: artifacts/functional/utils/parsing/node_parsing/output.md
---

# ROOT/functional/utils/parsing/node_parsing

Parses the body of a spec node file into a structured
representation of its sections and subsections.

# Public

## Interface

```
record Subsection
  heading: string
  content: string

record Section
  heading: string
  content: string
  subsections: list of Subsection

record ParsedNode
  name_section: Section
  public: optional Section
  agent: optional Section
  private: list of Section

function ParseNode(logical_name) -> ParsedNode
  errors:
    - unexpected content before first heading: file body has content before the first level-1 heading.
    - node name does not match: the first heading does not match the logical name after normalization.
    - duplicate public section: more than one `# Public` section exists.
    - duplicate subsection: two `##` headings within `# Public` normalize to the same text.
```

# Agent

## Behavior

Given a logical name, resolves the file path, skips the
frontmatter, and parses the remaining body into sections.

### Heading normalization

Headings are normalized before comparison: trim whitespace,
collapse internal whitespace to a single space, apply Unicode
simple case folding. See `ROOT/functional/utils/parsing/name_normalization`.

## Contracts

- Only level-1 (`#`) and level-2 (`##`) headings are structural.
  Level-3 and deeper are treated as content.
- Headings inside fenced code blocks are not structural.
- Leading and trailing blank lines in section content are trimmed.
