---
depends_on:
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing/node_parsing
  - SPEC/golang/implementation/utils/logical_names
output: internal/parsenode/parsenode_test.go
---

# SPEC/golang/tests/parsing/node_parsing

# Agent

## Test setup guidance

Tests create `_node.md` files on disk, then call
`NodeParse` with logical names. Use `testChdir` and
create `code-from-spec/.../_node.md` files.

Node files must follow the spec format: optional
frontmatter between `---` delimiters, then body with
ATX headings.

## Test cases

### Happy path

#### Minimal node — name section only

Setup:
- Create `code-from-spec/x/_node.md` with:
  `# SPEC/x`
  `A simple node.`

Actions:
1. Call `NodeParse("SPEC/x")`.

Expected:
- name_section.Heading = "spec/x"
- name_section.RawHeading = "# SPEC/x"
- name_section.Content = ["A simple node."]
- name_section.Subsections = empty
- Public, Agent, Private all nil

#### Full node — all section types

Setup:
- Create `code-from-spec/payments/fees/_node.md` with
  frontmatter, name heading + description, Public with
  Interface and Constraints subsections, Agent section,
  Private with Decisions and Rationale subsections.

Actions:
1. Call `NodeParse("SPEC/payments/fees")`.

Expected:
- name_section.Heading = "spec/payments/fees"
- Public present with two subsections "interface",
  "constraints"
- Public.Content = empty (no lines before first ##)
- Agent present with content
- Private present with two subsections "decisions",
  "rationale"

#### Node with no public section

Setup:
- Node with name heading, content, Private with
  Rationale subsection. No Public or Agent.

Expected: Public nil, Agent nil, Private present.

#### Public section with content before first subsection

Setup:
- Node with Public having two preamble lines before
  `## Interface`.

Expected:
- Public.Content = two-element list with preamble
- Public.Subsections has one entry "interface"

#### Public section with no content or subsections

Setup:
- Public heading immediately followed by Agent heading.

Expected:
- Public present with empty Content and empty
  Subsections.

#### Agent section with subsections

Setup:
- Agent with preamble line, then `## Implementation
  guidance` and `## Contracts` subsections.

Expected:
- Agent.Content = one-element list
- Agent.Subsections has two entries

#### Private section with subsections

Setup:
- Private with TODO, Decisions, Rationale subsections.

Expected:
- Private present with three subsections in order.

#### Content is raw markdown

Setup:
- Public subsection containing a level-3 heading,
  bold text, and a fenced code block.

Expected:
- All lines in subsection Content as raw strings.

### Heading normalization

#### Case insensitive public detection

Setup: Node with `# PUBLIC` heading.

Expected: Public present, Heading = "public".

#### Public with mixed case and extra whitespace

Setup: Node with `#   PuBLiC` heading.

Expected: Public present, Heading = "public".

#### Node name with varied whitespace

Setup: Node with `#   SPEC/e` heading.

Expected: name_section.Heading = "spec/e".

#### Node name with ROOT/ heading does not match SPEC/

Setup: Node with `# ROOT/x` heading.

Actions: Call `NodeParse("SPEC/x")`.

Expected: Error ErrNodeNameDoesNotMatch.

#### Subsection headings are normalized

Setup: Node with `##   Interface` and `## CONSTRAINTS`.

Expected: Subsection headings = "interface",
"constraints".

#### Closing hashes are stripped

Setup: Node with `## Interface ##`.

Expected: Heading = "interface",
RawHeading = "## Interface ##".

### Raw heading preservation

#### Raw heading preserves original line

Setup: Node with `# Public` and `## Interface`.

Expected: Public.RawHeading = "# Public",
subsection RawHeading = "## Interface".

#### Raw heading preserves case

Setup: Node with `# PUBLIC`.

Expected: Heading = "public",
RawHeading = "# PUBLIC".

#### Raw heading preserves closing hashes

Setup: Node with `## Foo ##`.

Expected: Heading = "foo",
RawHeading = "## Foo ##".

#### Raw heading preserves extra whitespace

Setup: Node with `#   Public`.

Expected: Heading = "public",
RawHeading = "#   Public".

### Content boundaries

#### Level-3 and deeper headings are content

Setup: Public subsection with `###` and `####` lines.

Expected: Those lines in subsection Content.

#### Fenced code blocks with heading-like content

Setup: Backtick fence inside subsection with `#` and
`##` lines.

Expected: Heading-like lines are content, not
structural.

#### Fenced code block with tilde fence

Setup: `~~~` fence with `# heading` inside.

Expected: Content, not structural.

#### Fenced code block with language tag

Setup: ` ```python ` fence with `# comment` inside.

Expected: Content, not structural.

#### Blank lines between heading and content are preserved

Setup: One blank line between Public heading and
content.

Expected: Public.Content starts with "" (empty string)
then content line.

### Frontmatter handling

#### Frontmatter is skipped

Setup: Node with frontmatter between `---` delimiters.

Expected: No error, body parsed correctly.

#### No frontmatter delimiters

Setup: Node with no `---` at all.

Expected: No error, body parsed correctly.

#### Unclosed frontmatter

Setup: Node starts with `---` but no closing `---`.

Expected: Error ErrUnexpectedContentBeforeFirstHeading.

### Failure cases

#### ARTIFACT reference rejected

Actions: Call `NodeParse("ARTIFACT/x")`.

Expected: Error ErrNotASpecReference.

#### EXTERNAL reference rejected

Actions: Call `NodeParse("EXTERNAL/x")`.

Expected: Error ErrNotASpecReference.

#### Qualifier rejected

Actions: Call `NodeParse("SPEC/x(interface)")`.

Expected: Error ErrHasQualifier.

#### File does not exist

Actions: Call `NodeParse` with non-existent file.

Expected: Error ErrFileUnreadable.

#### Propagates path errors

Actions: Call `NodeParse("SPEC/tra\\versal")`.

Expected: Error PathContainsBackslash propagated,
not FileUnreadable.

#### Content before first heading

Setup: Text before any heading.

Expected: Error ErrUnexpectedContentBeforeFirstHeading.

#### Level-2 heading before any level-1 heading

Setup: `## Interface` before any `#` heading.

Expected: Error ErrUnexpectedContentBeforeFirstHeading.

#### Empty body

Setup: Empty file or only frontmatter.

Expected: Error ErrUnexpectedContentBeforeFirstHeading.

#### Node name does not match logical name

Setup: Heading text doesn't match logical name.

Expected: Error ErrNodeNameDoesNotMatch.

#### Node name case mismatch is not an error

Setup: Lowercase heading, uppercase logical name.

Expected: No error — normalization matches.

#### Duplicate public section — same case

Setup: Two `# Public` headings.

Expected: Error ErrDuplicatePublicSection.

#### Duplicate public section — different case

Setup: `# Public` then `# PUBLIC`.

Expected: Error ErrDuplicatePublicSection.

#### Duplicate agent section

Setup: Two `# Agent` headings.

Expected: Error ErrDuplicateAgentSection.

#### Duplicate private section

Setup: Two `# Private` headings.

Expected: Error ErrDuplicatePrivateSection.

#### Unrecognized section heading

Setup: `# Decisions` as top-level heading.

Expected: Error ErrUnrecognizedSection.

#### Unrecognized section — Rationale

Setup: `# Rationale` as top-level heading.

Expected: Error ErrUnrecognizedSection.

#### Unrecognized section — TODO

Setup: `# TODO` as top-level heading.

Expected: Error ErrUnrecognizedSection.

#### Duplicate subsection in public — same case

Setup: Two `## Interface` headings under Public.

Expected: Error ErrDuplicateSubsection.

#### Duplicate subsection in public — different case

Setup: `## Interface` then `## INTERFACE` under Public.

Expected: Error ErrDuplicateSubsection.

#### Duplicate subsection in public — whitespace variation

Setup: `## Interface` then `##   Interface` under
Public.

Expected: Error ErrDuplicateSubsection.

#### Duplicate subsection in agent

Setup: Two `## Guidance` headings inside Agent.

Expected: Error ErrDuplicateSubsection.

## Go-specific guidance

- The package name is `parsenode_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
- When a test case specifies expected content values,
  construct the test file so that the expected content
  is the correct result of parsing. Pay attention to
  blank lines — they are preserved in content.
