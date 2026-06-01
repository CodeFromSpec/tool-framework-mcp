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

Create a node file for `ROOT/x` containing a name
heading followed immediately (no blank line) by a
single line of text `A simple node.`, then no other
headings. Call `NodeParse` with `"ROOT/x"`.

Expect `name_section.heading` = `"root/x"`,
`name_section.raw_heading` = the original heading line,
`name_section.content` = `["A simple node."]`,
`name_section.subsections` empty,
`public` absent, `agent` absent, `private` empty.

#### Full node — all section types

Create a node file for `ROOT/payments/fees` with
frontmatter, a name heading followed immediately by
one line of description, then a Public section with
two subsections (Interface and Constraints — each with
one line of content, no blank lines between heading
and content), then an Agent section with one line of
content, then two private sections (Decisions and
Rationale) each with one line of content. No blank
lines between any heading and its content.

Call `NodeParse` with `"ROOT/payments/fees"`.

Expect:
- `name_section.heading` = `"root/payments/fees"`,
  content = one-element list with the description line
- `public` present with two subsections `"interface"`
  and `"constraints"`, each with one-element content
- `public.content` = empty list (no lines before first
  subsection)
- `agent` present with one-element content
- `private` has two sections in order: `"decisions"`,
  `"rationale"`

#### Node with no public section

Create a node file for `ROOT/decisions` with a name
heading, one line of content, then a private section
`Rationale` with content. No Public or Agent sections.

Call `NodeParse`. Expect `public` absent, `agent`
absent, `private` has one section with heading
`"rationale"`.

#### Public section with content before first subsection

Create a node file for `ROOT/a` with a name heading
and content, then a Public section with two lines of
preamble text (no blank line after the Public heading),
then an Interface subsection with one line. No blank
lines between any heading and its content.

Call `NodeParse`. Expect `public.content` = the two
preamble lines as a two-element list.
`public.subsections` has one entry with heading
`"interface"`.

#### Public section with no content or subsections

Create a node file where the Public heading is
immediately followed by an Agent heading (no lines
between them). Call `NodeParse`.

Expect `public` present with empty `content` (empty
list) and empty `subsections` list.

#### Agent section with subsections

Create a node file with a name heading, then an Agent
section with one line of preamble (no blank line after
heading), then two subsections (Implementation guidance
and Contracts) each with one line of content (no blank
lines between headings and content).

Call `NodeParse`. Expect `agent.content` = one-element
list with the preamble line. `agent.raw_heading` = the
original Agent heading line. `agent.subsections` has
two entries with headings `"implementation guidance"`
and `"contracts"`.

#### Private sections preserve file order

Create a node file with three private sections in this
order: TODO, Decisions, Rationale.

Call `NodeParse`. Expect `private` has three sections
in order: `"todo"`, `"decisions"`, `"rationale"`.

#### Content is raw markdown

Create a node file with a public subsection containing
a level-3 heading line, a line with bold text, and a
fenced code block (backtick fence with content inside).
No blank lines between the subsection heading and the
content lines.

Call `NodeParse`. Expect the subsection content is a
list containing the level-3 heading line, the bold text
line, and all code block lines (opening fence, content,
closing fence) as raw strings.

### Heading normalization

#### Case insensitive public detection

Create a node with `PUBLIC` (all caps) as the public
heading text. Call `NodeParse`. Expect `public` present,
heading = `"public"`.

#### Public with mixed case and extra whitespace

Create a node with extra spaces between `#` and
`PuBLiC` in the heading. Call `NodeParse`. Expect
`public` present, heading = `"public"`.

#### Node name with varied whitespace

Create a node with extra spaces between `#` and
`ROOT/e` in the name heading. Call `NodeParse` with
`"ROOT/e"`. Expect `name_section.heading` = `"root/e"`.

#### Subsection headings are normalized

Create a node with subsections with extra whitespace
(`##   Interface`) and all caps (`## CONSTRAINTS`).
Call `NodeParse`. Expect subsection headings =
`"interface"` and `"constraints"`.

#### Closing hashes are stripped

Create a node with heading `## Interface ##` (with
closing hashes). Call `NodeParse`. Expect subsection
heading = `"interface"`, raw_heading = `"## Interface ##"`.

### Raw heading preservation

#### Raw heading preserves original line

Create a node with standard headings `# Public` and
`## Interface`. Call `NodeParse`. Expect
`public.raw_heading` = `"# Public"`, subsection
`raw_heading` = `"## Interface"`.

#### Raw heading preserves case

Create a node with `# PUBLIC` heading. Call `NodeParse`.
Expect `public.heading` = `"public"` (normalized),
`public.raw_heading` = `"# PUBLIC"` (original).

#### Raw heading preserves closing hashes

Create a node with `## Foo ##`. Call `NodeParse`.
Expect subsection `heading` = `"foo"`,
`raw_heading` = `"## Foo ##"`.

#### Raw heading preserves extra whitespace

Create a node with `#   Public` (extra spaces) heading.
Call `NodeParse`. Expect `public.heading` = `"public"`,
`public.raw_heading` = `"#   Public"`.

### Content boundaries

#### Level-3 and deeper headings are content

Create a node with a public subsection containing lines
that start with `###` and `####`. Call `NodeParse`.

Expect those lines are included in the subsection
content list, not treated as structural headings.

#### Fenced code blocks with heading-like content

Create a node with a fenced code block (backtick)
inside a public subsection, where the code block
contains lines starting with `#` and `##`. Call
`NodeParse`.

Expect the heading-like lines inside the code block are
in the subsection content, not treated as structural
headings.

#### Fenced code block with tilde fence

Create a node with a code block opened by `~~~` inside
a subsection, containing a line that looks like a
level-1 heading. Call `NodeParse`.

Expect that line is content, not a structural heading.

#### Fenced code block with language tag

Create a node with a code block opened by backtick
fence with a language tag, containing a line that looks
like a level-1 heading. Call `NodeParse`.

Expect that line is content, not a structural heading.

#### Blank lines between heading and content are preserved

Create a node file where there is one blank line
between the Public heading and the first content line.
Call `NodeParse`.

Expect `public.content` starts with an empty string
(the blank line) followed by the content line.

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

Expect error UnexpectedContentBeforeFirstHeading.

### Failure cases

#### ARTIFACT reference rejected

Call `NodeParse` with `"ARTIFACT/x(y)"`.
Expect error NotARootReference.

#### Qualifier rejected

Call `NodeParse` with `"ROOT/x(interface)"`.
Expect error HasQualifier.

#### File does not exist

Call `NodeParse` with a logical name whose file does
not exist. Expect error FileUnreadable.

#### Propagates path errors

Call `NodeParse` with an invalid logical name that
causes a path error (e.g., after resolving to a path
with traversal). Expect the path error is propagated.

#### Content before first heading

Create a node file with text before any heading. Call
`NodeParse`. Expect error
UnexpectedContentBeforeFirstHeading.

#### Level-2 heading before any level-1 heading

Create a node file with a level-2 heading before any
level-1 heading. Call `NodeParse`. Expect error
UnexpectedContentBeforeFirstHeading.

#### Empty body

Create a node file with no content (or only frontmatter,
no body). Call `NodeParse`. Expect error
UnexpectedContentBeforeFirstHeading.

#### Node name does not match logical name

Create a node file where the first heading text does
not match the logical name. Call `NodeParse` with a
different logical name. Expect error
NodeNameDoesNotMatch.

#### Node name case mismatch is not an error

Create a node file with lowercase heading text and call
`NodeParse` with the uppercase logical name. Expect no
error — normalization makes them equal.

#### Duplicate public section — same case

Create a node with two Public sections (same case).
Call `NodeParse`. Expect error DuplicatePublicSection.

#### Duplicate public section — different case

Create a node with Public sections in different cases.
Call `NodeParse`. Expect error DuplicatePublicSection.

#### Duplicate agent section

Create a node with two Agent sections. Call `NodeParse`.
Expect error DuplicateAgentSection.

#### Duplicate subsection in public — same case

Create a node with two identical subsection headings
under public. Call `NodeParse`. Expect error DuplicateSubsection.

#### Duplicate subsection in public — different case

Create a node with two subsection headings that differ
only in case under public. Call `NodeParse`. Expect
"duplicate subsection".

#### Duplicate subsection in public — whitespace variation

Create a node with two subsection headings that differ
only in whitespace under public. Call `NodeParse`.
Expect error DuplicateSubsection.

#### Duplicate subsection in agent

Create a node with two identical subsection headings
inside Agent. Call `NodeParse`. Expect error DuplicateSubsection.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface: `NodeParse`.
- When a test case specifies expected content values,
  construct the test file so that the expected content
  is the correct result of parsing. Pay attention to
  blank lines — they are preserved in content.
