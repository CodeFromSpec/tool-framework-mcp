---
depends_on:
  - ROOT/functional/logic/utils/logical_names(interface)
  - ROOT/functional/logic/os/file_reader(interface)
  - ROOT/functional/logic/utils/text_normalization(interface)
outputs:
  - id: node_parsing
    path: code-from-spec/functional/logic/parsing/node_parsing/output.md
---

# ROOT/functional/logic/parsing/node_parsing

Parses the body of a spec node file into a structured
representation of its sections and subsections.

# Public

## Interface

```
record NodeSubsection
  heading: string
  content: string

record NodeSection
  heading: string
  content: string
  subsections: list of NodeSubsection

record Node
  name_section: NodeSection
  public: optional NodeSection
  agent: optional NodeSection
  private: list of NodeSection

function NodeParse(logical_name: string) -> Node
  errors:
    - has qualifier: the logical name contains a
      parenthetical qualifier.
    - (path errors): propagated from FileOpen.
    - file unreadable: the file cannot be opened or read.
    - unexpected content before first heading: file body
      has content before the first level-1 heading, or
      has no level-1 heading at all.
    - node name does not match: the first heading does not
      match the logical name after normalization.
    - duplicate public section: more than one `# Public`
      section exists.
    - duplicate agent section: more than one `# Agent`
      section exists.
    - duplicate subsection: two `##` headings within
      `# Public` normalize to the same text.
```

`NodeSubsection` and `NodeSection` headings are stored in
normalized form (after `NormalizeText`).

Subsections (`##`) are only structural within `# Public` —
they support selective import via `depends_on` qualifiers.
In `# Agent` and private sections, `##` headings are
treated as content.

# Agent

## Behavior

Given a logical name:

1. If `LogicalNameHasQualifier` returns true, raise
   "has qualifier" — this function parses the full node,
   not a subsection.
2. Resolve the file path using `LogicalNameToPath`.
3. Open the file with `FileOpen`.
4. Skip the frontmatter: if the first line is exactly
   `---`, read and discard lines until the next `---`.
   If end of file is reached without finding the closing
   `---`, raise "unexpected content before first heading".
5. Parse the remaining body into sections.
6. Close the reader with `FileClose` when done.

### ATX heading recognition

Only ATX headings are recognized (CommonMark). An ATX
heading line starts with one or more `#` characters
followed by at least one space. The heading text is
everything after the `# ` prefix (hash(es) + space),
trimmed of leading and trailing whitespace. Lines
like `#Foo` (no space after `#`) are not headings —
they are content.

The heading level is determined by the number of `#`
characters: `#` = level 1, `##` = level 2, etc.

### Heading normalization

After extracting the heading text, normalize it using
`NormalizeText` before comparison or storage: trim
whitespace, collapse internal whitespace to a single
space, apply Unicode simple case folding.

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
- In the public section: content between the `#` heading
  and the first `##` heading (or the next `#` heading or
  end of file) is the section's `content` field.
  Level-2 (`##`) headings start subsections. Content
  between a `##` heading and the next `##` or `#` heading
  (or end of file) is the subsection's `content` field.
  If two `##` headings normalize to the same text, raise
  "duplicate subsection". The `subsections` list contains
  all `##` subsections in order.
- In all other sections (name, agent, private): the
  section's `content` field is everything between the
  `#` heading and the next `#` heading (or end of file).
  `##` headings are not structural — they are part of the
  content. The `subsections` list is always empty.
- Level-3 and deeper headings are always content.
- Headings inside fenced code blocks are not structural —
  they are treated as content. A fenced code block opens
  with a line of three or more consecutive backtick
  characters or three or more consecutive tilde characters,
  optionally followed by a language tag. It closes with a
  line of at least as many of the same character as the
  opening line. All lines between are content, regardless
  of whether they look like headings.
- Leading and trailing blank lines in section and
  subsection content are trimmed.

## Contracts

- Only level-1 (`#`) and level-2 (`##`) headings are
  structural. Level-3 and deeper are content.
- Headings inside fenced code blocks are not structural.
- Leading and trailing blank lines in content are trimmed.
- Subsection duplicate detection only applies within
  `# Public`.
