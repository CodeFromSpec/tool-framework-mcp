---
depends_on:
  - ROOT/dependencies/yuin-goldmark
  - ROOT/golang/internal/logical_names
  - ROOT/golang/internal/normalizename
external:
  - path: CODE_FROM_SPEC.md
outputs:
  - id: parsenode
    path: internal/parsenode/parsenode.go
---

# ROOT/golang/internal/parsenode/code

Generates the parsenode package implementation.

# Agent

## Implementation

#### Step 1 — Resolve logical name

Resolve the logical name to a file path using
`logicalnames.PathFromLogicalName`.

#### Step 2 — Skip frontmatter

The frontmatter is delimited by the first `---` and the
second `---` at the top of the file. Find the closing `---`
and discard everything up to and including it. If no
frontmatter delimiters are found, it is an error.

#### Step 3 — Parse body as CommonMark

The body is everything after the closing `---` of the
frontmatter. Parse it with goldmark to produce an AST.

#### Step 4 — Validate: first child is a level-1 heading

The first direct child of the document root node must be a
level-1 heading. If it is not, it is an error.

#### Step 5 — Validate: node name section

Extract the inline text content of the first level-1 heading
(see "Extracting heading text" in `ROOT/dependencies/yuin-goldmark`).
Apply `normalizename.NormalizeName` to it and to the logical name received
as argument. If the results do not match, it is an error.

#### Step 6 — Validate: no duplicate public section

For each level-1 heading, extract its inline text content
(see "Extracting heading text" in `ROOT/dependencies/yuin-goldmark`)
and apply `normalizename.NormalizeName`. If more than one result equals
`public`, it is an error.

#### Step 7 — Validate: no duplicate public subsections

For each level-2 heading within the public section, extract
its inline text content (see "Extracting heading text" in
`ROOT/dependencies/yuin-goldmark`) and apply `normalizename.NormalizeName`. If any
two results are equal, it is an error.

#### Step 8 — Extract sections

Iterate the direct children of the document root. Each level-1
heading starts a new section. All AST nodes between one level-1
heading and the next (or end of document) are the content of
that section.

For each section, extract:
- **Heading** — extract the inline text content of the level-1
  heading (see "Extracting heading text" in
  `ROOT/dependencies/yuin-goldmark`) and apply `normalizename.NormalizeName`.
- **Content** — the raw source bytes between the end of the
  level-1 heading and the start of the line of the first
  level-2 heading within the section (or the start of the line
  of the next level-1 heading / end of document if there are
  no level-2 headings). Use "Extracting raw source between
  headings" in `ROOT/dependencies/yuin-goldmark` to determine heading
  line boundaries.
- **Subsections** — each level-2 heading within the section
  starts a subsection. A subsection's heading is obtained by
  extracting the inline text content of the level-2 heading
  (see "Extracting heading text" in `ROOT/dependencies/yuin-goldmark`)
  and applying `normalizename.NormalizeName`. A
  subsection's content is the raw source bytes between the end
  of the level-2 heading and the start of the line of the next
  level-2 heading, the next level-1 heading, or the end of
  document.

Leading and trailing blank lines in content are trimmed.

#### Step 9 — Classify sections

1. The first section is the node name section.
2. A section whose `normalizename.NormalizeName` result equals `public`
   is the public section.
3. All other sections are private.

### Invariants

- The first element after the frontmatter is always a level-1
  heading. If it is not, the file is invalid
  (`ErrUnexpectedContent`).
- Every subsection (`##`) is contained within a section (`#`).
  There are no orphan subsections.
- There is exactly one node name section (the first `#`),
  and its normalized heading matches the logical name received
  as argument.
- There is at most one public section (`# Public`). Duplicates
  are rejected.
- All `##` subsections within `# Public` have unique normalized
  headings. Duplicates are rejected.
- Headings of level 3 and deeper are content, not structural
  delimiters. They appear inside `Section.Content` or
  `Subsection.Content` as raw markdown.
