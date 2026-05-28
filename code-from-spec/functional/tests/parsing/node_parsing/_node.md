---
depends_on:
  - ROOT/functional/logic/parsing/node_parsing(interface)
outputs:
  - id: node_parsing_tests
    path: code-from-spec/functional/tests/parsing/node_parsing/output.md
---

# ROOT/functional/tests/parsing/node_parsing

Test cases for the node parsing component.

# Public

## Test cases

### Happy path

#### Minimal node — name section only

Create a node file for `ROOT/x` with only a name heading
and a description. Call `NodeParse` with `"ROOT/x"`.

Expect `name_section.heading` = `"root/x"`,
`name_section.content` = `"A simple node."`,
`name_section.subsections` empty,
`public` absent, `agent` absent, `private` empty.

#### Full node — name, public, agent, private

Create a node file for `ROOT/payments/fees` with
frontmatter, a name section, a public section with
`## Interface` and `## Constraints` subsections, an
agent section, and two private sections (`# Decisions`,
`# Rationale`). Call `NodeParse`.

Expect `name_section.heading` = `"root/payments/fees"`,
`public` present with two subsections (`"interface"`,
`"constraints"`), `agent` present with content,
`private` has two sections in file order.

#### Node with no public section

Create a node file for `ROOT/decisions` with a name
section and a private section `# Rationale`. Call
`NodeParse`.

Expect `public` absent, `agent` absent,
`private` has one section with heading `"rationale"`.

#### Public section with content before first subsection

Create a node file for `ROOT/a` with a public section
that has direct content before a `## Interface`
subsection. Call `NodeParse`.

Expect `public.content` = the text before the subsection,
`public.subsections` has one entry with heading
`"interface"`.

#### Public section with no content or subsections

Create a node file where `# Public` is immediately
followed by `# Agent`. Call `NodeParse`.

Expect `public` present with empty `content` and empty
`subsections` list.

#### Agent section with ## subsections

Create a node file with an agent section containing
some preamble text, then `## Implementation guidance`
and `## Contracts` subsections. Call `NodeParse`.

Expect `agent.content` = the preamble text,
`agent.subsections` has two entries with headings
`"implementation guidance"` and `"contracts"`, each
with their own content.

#### Private sections preserve file order

Create a node file with three private sections:
`# TODO`, `# Decisions`, `# Rationale`. Call `NodeParse`.

Expect `private` has three sections in order:
`"todo"`, `"decisions"`, `"rationale"`.

#### Content is raw markdown

Create a node file with a subsection containing level-3
headings, bold text, and code blocks. Call `NodeParse`.

Expect the subsection content is the raw markdown text,
including `###` headings, `**bold**`, and fenced code
blocks.

### Heading normalization

#### Case insensitive public detection

Create a node with `# PUBLIC` as the public heading.
Call `NodeParse`. Expect `public` present, heading =
`"public"`.

#### Public with mixed case and extra whitespace

Create a node with `#   PuBLiC` as the public heading.
Call `NodeParse`. Expect `public` present, heading =
`"public"`.

#### Node name with varied whitespace

Create a node with `#    ROOT/e` as the name heading.
Call `NodeParse` with `"ROOT/e"`. Expect
`name_section.heading` = `"root/e"`.

#### Subsection headings are normalized

Create a node with subsections `##   Interface` and
`## CONSTRAINTS`. Call `NodeParse`. Expect subsection
headings = `"interface"` and `"constraints"`.

#### Closing hashes are stripped

Create a node with heading `## Interface ##`. Call
`NodeParse`. Expect subsection heading = `"interface"`.

### Content boundaries

#### Level-3 and deeper headings are content

Create a node with a public subsection containing
`### Details` and `#### Sub-details`. Call `NodeParse`.

Expect the `###` and `####` lines and their text are
included as raw content within the subsection.

#### Fenced code blocks with heading-like content

Create a node with a fenced code block inside a public
subsection that contains lines starting with `#` and
`##`. Call `NodeParse`.

Expect the heading-like lines inside the code block are
treated as content, not as structural headings.

#### Fenced code block with tilde fence

Create a node with a code block opened by `~~~` inside
a subsection, containing `# Heading`. Call `NodeParse`.

Expect the `# Heading` inside the tilde fence is content.

#### Fenced code block with language tag

Create a node with a code block opened by ` ```yaml `
inside a subsection, containing `# Heading`. Call
`NodeParse`.

Expect the `# Heading` inside the code block is content.

#### Leading and trailing blank lines are trimmed

Create a node with blank lines surrounding content in
sections and subsections. Call `NodeParse`.

Expect leading and trailing blank lines are trimmed from
all `content` fields.

### Frontmatter handling

#### Frontmatter is skipped

Create a node file with frontmatter (between `---`
delimiters) and a body. Call `NodeParse`.

Expect frontmatter is skipped. Body parsed correctly.

#### No frontmatter delimiters

Create a node file with no `---` at all — body only.
Call `NodeParse`. Expect no error. Body parsed correctly.

#### Unclosed frontmatter

Create a node file that starts with `---` but has no
closing `---`. Call `NodeParse`.

Expect "unexpected content before first heading".

### Failure cases

#### ARTIFACT reference rejected

Call `NodeParse` with `"ARTIFACT/x(y)"`.
Expect "not a ROOT reference".

#### Qualifier rejected

Call `NodeParse` with `"ROOT/x(interface)"`.
Expect "has qualifier".

#### File does not exist

Call `NodeParse` with a logical name whose file does
not exist. Expect "file unreadable".

#### Propagates path errors

Call `NodeParse` with an invalid logical name that
causes a path error (e.g., after resolving to a path
with traversal). Expect the path error is propagated.

#### Content before first heading

Create a node file with text before any heading. Call
`NodeParse`. Expect "unexpected content before first
heading".

#### Level-2 heading before any level-1 heading

Create a node file with a `##` heading before any `#`
heading. Call `NodeParse`. Expect "unexpected content
before first heading".

#### Empty body

Create a node file with no content (or only frontmatter,
no body). Call `NodeParse`. Expect "unexpected content
before first heading".

#### Node name does not match logical name

Create a node file where the first heading is
`# ROOT/other` but call `NodeParse` with `"ROOT/x"`.
Expect "node name does not match".

#### Node name case mismatch is not an error

Create a node file with heading `# root/x` and call
`NodeParse` with `"ROOT/x"`. Expect no error —
normalization makes them equal.

#### Duplicate public section — same case

Create a node with two `# Public` sections. Call
`NodeParse`. Expect "duplicate public section".

#### Duplicate public section — different case

Create a node with `# Public` and `# PUBLIC`. Call
`NodeParse`. Expect "duplicate public section".

#### Duplicate agent section

Create a node with two `# Agent` sections. Call
`NodeParse`. Expect "duplicate agent section".

#### Duplicate subsection in public — same case

Create a node with two `## Interface` subsections
under public. Call `NodeParse`. Expect "duplicate
subsection".

#### Duplicate subsection in public — different case

Create a node with `## Interface` and `## INTERFACE`
under public. Call `NodeParse`. Expect "duplicate
subsection".

#### Duplicate subsection in public — whitespace variation

Create a node with `## Interface` and `##   Interface`
under public. Call `NodeParse`. Expect "duplicate
subsection".

#### Duplicate subsection in agent

Create a node with two `## Details` headings inside
`# Agent`. Call `NodeParse`. Expect "duplicate
subsection".

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface: `NodeParse`.
