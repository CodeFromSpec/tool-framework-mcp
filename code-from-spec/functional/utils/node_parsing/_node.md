---
outputs:
  - id: node_parsing
    path: code-from-spec/functional/utils/node_parsing/output.md
---

# ROOT/functional/utils/node_parsing

Parses the body of a spec node file into a structured
representation of its sections and subsections.

# Public

## Behavior

Given a logical name, resolves the file path, skips the
frontmatter, and parses the remaining body into sections.

### Input

A logical name (e.g., `ROOT/golang/server`).

### Output

A structured record with:

| Field | Description |
|---|---|
| `name_section` | The first section — heading matches the logical name. |
| `public` | The `# Public` section, if present. May be absent. |
| `agent` | The `# Agent` section, if present. May be absent. |
| `private` | All other sections. |

Each **section** has:
- `heading` — the normalized heading text.
- `content` — raw markdown between this heading and the next.
- `subsections` — list of level-2 sections within it.

Each **subsection** has:
- `heading` — the normalized heading text.
- `content` — raw markdown between this heading and the next.

## Heading normalization

Headings are normalized before comparison: trim whitespace,
collapse internal whitespace to a single space, apply Unicode
simple case folding. See `ROOT/functional/utils/name_normalization`.

## Validation rules

| Rule | Error condition |
|---|---|
| First element after frontmatter must be a level-1 heading | Unexpected content before first heading |
| First heading must match the logical name (after normalization) | Node name does not match |
| At most one `# Public` section | Duplicate public section |
| All `##` headings within `# Public` must be unique (after normalization) | Duplicate subsection |

## Contracts

- Only level-1 (`#`) and level-2 (`##`) headings are structural.
  Level-3 and deeper are treated as content.
- Headings inside fenced code blocks are not structural.
- Leading and trailing blank lines in section content are trimmed.
