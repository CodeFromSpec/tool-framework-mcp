# File Format

Detailed file format rules for Code from Spec specification
files. This level of detail is primarily relevant for tool
implementors. Spec authors and AI agents can rely on the
summary in CODE_FROM_SPEC.md.

This document assumes familiarity with
[CODE_FROM_SPEC.md](../CODE_FROM_SPEC.md).

---

## Encoding

Specification files are UTF-8 encoded, without BOM.

---

## Markdown

Specification files use [CommonMark](https://commonmark.org/)
for Markdown formatting.

---

## YAML frontmatter

Frontmatter is not part of CommonMark — it is an extension
adopted by this framework.

The frontmatter block starts with a line containing exactly `---`
(three hyphens, nothing else) as the first line of the file, and
ends with the next line containing exactly `---`. The content
between the two delimiters is parsed as YAML.

---

## Heading levels

Only ATX headings (`#` prefix) are recognized by the framework.
Setext headings are not supported.

Only two heading levels are structural for the framework:

- **Level 1 (`#`)** — delimits top-level sections (node
  name section, `# Public`, `# Agent`, `# Private`).
- **Level 2 (`##`)** — delimits subsections within a
  top-level section (e.g. `## Interface` within `# Public`).

Headings of level 3 and deeper (`###`, `####`, ...) are content
within the section or subsection that contains them. They have no
structural meaning for the framework.

---

## Heading content normalization

Heading content is normalized before comparison using these rules,
applied in order:

1. **Trim** — leading and trailing whitespace is removed.
2. **Collapse** — each sequence of one or more whitespace
   characters within the heading content is replaced by a single
   `U+0020` (space).
3. **Case fold** — the result is case-folded using Unicode simple
   case folding.

The whitespace characters recognized by the framework are space
(`U+0020`) and horizontal tab (`U+0009`). Any other Unicode
whitespace (e.g. `U+00A0` non-breaking space) is not recognized —
it is treated as part of the heading text.

These normalization rules apply equally to headings in
specification files and to the parenthetical qualifier in logical
names. For example, all of the following are equivalent:

- `## Testes de aceitação` (heading in a file)
- `##   TESTES   DE   ACEITAÇÃO  ` (heading in a file)
- `SPEC/x/y(Testes de aceitação)` (logical name)
- `SPEC/x/y(Testes    de  aceitação  )` (logical name)
- `SPEC/x/y(testes de ACEITAÇÃO)` (logical name)

---

## Block extraction

How section and subsection content is extracted from a
`_node.md` for chain assembly and hashing. The same extracted
form is used for both: what is hashed is exactly what is
delivered in the chain.

A **block** is a top-level section (`#`) or a subsection
(`##`). A block's raw content is everything between its
heading line and the next structural heading (`#` or `##`)
or the end of the file.

Extraction normalizes only the block's boundaries:

1. Leading blank lines (immediately after the heading line)
   are removed.
2. Trailing blank lines are removed.
3. The content ends with exactly one LF.

A blank line is a line that is empty or contains only
whitespace (`U+0020` and `U+0009`).

Everything between the first and the last non-blank line is
preserved byte for byte. Internal blank-line runs,
indentation, alignment, and trailing whitespace are content —
they may carry meaning (code examples, output formats) and
are never normalized. Boundary blank lines, by contrast, are
document layout: they separate blocks and belong to no block.

This makes a block's extracted content independent of its
surroundings. Adding `# Private` after `# Public`, or
appending a new `##` subsection, does not change the
extracted content of neighboring blocks — and therefore does
not change their hashes.

### Concatenation

When multiple blocks are combined — such as the `##`
subsections of `# Public`, in document order — each block is
rendered as its heading line (with trailing whitespace
removed) followed by its extracted content, and consecutive
blocks are separated by exactly one blank line.
