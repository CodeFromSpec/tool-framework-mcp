# File Format

Detailed file format rules for Code from Spec specification
files. This document assumes familiarity with
[CODE_FROM_SPEC.md](CODE_FROM_SPEC.md).

This level of detail is primarily relevant for tool implementors
(parsers, staleness checkers, chain assemblers). Spec authors
and AI agents can rely on the summary in CODE_FROM_SPEC.md.

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

- **Level 1 (`#`)** — delimits top-level sections (node name,
  `# Public`, `# Agent`, private sections).
- **Level 2 (`##`)** — delimits subsections within a top-level
  section (e.g. `## Interface` within `# Public`).

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
- `ROOT/x/y(Testes de aceitação)` (logical name)
- `ROOT/x/y(Testes    de  aceitação  )` (logical name)
- `ROOT/x/y(testes de ACEITAÇÃO)` (logical name)
