---
depends_on:
  - ROOT/functional/logic/utils/logical_names(interface)
  - ROOT/functional/logic/os/file_reader(interface)
  - ROOT/functional/logic/utils/text_normalization(interface)
output: code-from-spec/functional/logic/parsing/node_parsing/output.md
---

# ROOT/functional/logic/parsing/node_parsing

Parses the body of a spec node file into a structured
representation of its sections and subsections.

# Public

## Namespace

    namespace: parsenode

## Interface

```
record NodeSubsection
  heading: string
  raw_heading: string
  content: list of string

record NodeSection
  heading: string
  raw_heading: string
  content: list of string
  subsections: list of NodeSubsection

record Node
  name_section: NodeSection
  public: optional NodeSection
  agent: optional NodeSection
  private: list of NodeSection

function NodeParse(logical_name: string) -> Node
  errors:
    - NotARootReference: the logical name does not
      start with ROOT/.
    - HasQualifier: the logical name contains a
      parenthetical qualifier.
    - FileUnreadable: the file cannot be opened or read.
    - UnexpectedContentBeforeFirstHeading: file body
      has non-blank content before the first level-1
      heading, or has no level-1 heading at all. Blank
      lines before the first heading are not an error.
    - NodeNameDoesNotMatch: the first heading does not
      match the logical name after normalization.
    - DuplicatePublicSection: more than one Public
      section exists.
    - DuplicateAgentSection: more than one Agent
      section exists.
    - DuplicateSubsection: two level-2 headings within
      the same section normalize to the same text.
    - (FileReader.*): propagated from FileOpen.
```

`NodeSubsection` and `NodeSection` have two heading fields:
`heading` is the normalized form (after `NormalizeText`),
used for comparisons and lookups. `raw_heading` is the
original line as read from the file (e.g.
`# Public`, `## Interface ##`), preserved for hashing.
Content fields are lists of strings — each element is a
line as returned by `FileReadLine`, preserving the
original text exactly as read from the file.

Private sections preserve the order they appear in the
file.

A section that exists in the file but has no content
(e.g., `# Public` immediately followed by `# Agent`) is
present with an empty `content` and an empty `subsections`
list — it is not absent.

Subsections (`##`) are structural in all sections. Each
section can have a `content` field (text before the first
`##`) and a `subsections` list (one entry per `##`
heading). Selective import via `depends_on` qualifiers
only applies to `# Public` subsections, but the parsing
structure is uniform across all sections.

# Agent

## Behavior

Given a logical name:

1. If `LogicalNameIsArtifact` returns true, raise
   "not a ROOT reference".
2. If `LogicalNameHasQualifier` returns true, raise
   "has qualifier" — this function parses the full node,
   not a subsection.
3. Resolve the file path using `LogicalNameToPath`.
4. Open the file with `FileOpen`. If it fails, raise
   "file unreadable".
5. Skip the frontmatter: if the first line is exactly
   `---`, read and discard lines until the next `---`.
   If end of file is reached without finding the closing
   `---`, raise "unexpected content before first heading".
6. Parse the remaining body into sections.
7. Close the reader with `FileClose` when done — in all
   cases, whether parsing succeeds or fails.

### ATX heading recognition

Only ATX headings are recognized (CommonMark). An ATX
heading line starts with one or more `#` characters
followed by at least one space. The heading text is
everything after the `# ` prefix (hash(es) + space),
trimmed of leading and trailing whitespace. Lines
like `#Foo` (no space after `#`) are not headings —
they are content.

CommonMark allows optional closing `#` sequences:
`## Foo ##` has heading text `Foo` (the closing hashes
are stripped). If present, the closing sequence must be
preceded by at least one space.

The heading level is determined by the number of leading
`#` characters: `#` = level 1, `##` = level 2, etc.

### Heading storage

When a heading line is recognized, store two values:
- `raw_heading`: the original line as read from the
  file, unchanged (e.g. `# Public`, `## Interface ##`).
- `heading`: the extracted heading text, normalized with
  `NormalizeText` (see below).

### Heading normalization

After extracting the heading text (stripping `#` prefix,
optional closing `#` sequences, and whitespace), normalize
it using `NormalizeText` before comparison or storage in
the `heading` field: trim whitespace, collapse internal
whitespace to a single space, apply Unicode simple case
folding.

### Section classification

After normalizing a level-1 heading with `NormalizeText`:

- The **first** level-1 heading is always the node name
  section. Its normalized heading text must match the
  logical name (also normalized with `NormalizeText`).
  For example, heading `# ROOT/functional/logic/os`
  has text `ROOT/functional/logic/os`, which normalizes
  to `root/functional/logic/os`. If it does not match,
  raise "node name does not match".
- A heading that normalizes to `"public"` is the public
  section. If a second one appears, raise "duplicate
  public section".
- A heading that normalizes to `"agent"` is the agent
  section. If a second one appears, raise "duplicate
  agent section".
- Any other level-1 heading is a private section.

### Section parsing

- Level-1 (`#`) headings start a new section.
- Level-2 (`##`) headings start a new subsection within
  the current section. This applies uniformly to all
  sections (name, public, agent, private).
- Content between the `#` heading and the first `##`
  heading (or the next `#` heading or end of file) is
  the section's `content` field.
- Content between a `##` heading and the next `##` or
  `#` heading (or end of file) is the subsection's
  `content` field.
- If two `##` headings within the same section normalize
  to the same text, raise "duplicate subsection".
- The `subsections` list contains all `##` subsections
  in order of appearance.
- Level-3 and deeper headings are always content.
- Headings inside fenced code blocks are not structural —
  they are treated as content. A fenced code block opens
  with a line of three or more consecutive backtick
  characters or three or more consecutive tilde characters,
  optionally followed by a language tag. It closes with a
  line of at least as many of the same character as the
  opening line. All lines between are content, regardless
  of whether they look like headings.
- Content preserves all lines as read from the file,
  including leading and trailing blank lines.

## Contracts

- Level-1 (`#`) headings are structural in all sections.
- Level-2 (`##`) headings are structural in all sections.
- Level-3 and deeper headings are always content.
- Headings inside fenced code blocks are not structural.
- Content preserves all lines including leading and
  trailing blank lines.
- Duplicate subsection detection applies within each
  section independently.
- `FileClose` is called in all cases — success or error.
