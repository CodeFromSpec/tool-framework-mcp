<!-- code-from-spec: SPEC/functional/tests/parsing/node_parsing@TV0Ik8lUgy5lfxoOkGqLdHVw4pA -->

## Happy path

### Minimal node — name section only

Setup: Create a node file for `SPEC/x` with the following lines:
  `# SPEC/x`
  `A simple node.`

Action: Call `NodeParse("SPEC/x")`.

Expected outcome:
- `name_section.heading` = `"spec/x"`
- `name_section.raw_heading` = `"# SPEC/x"`
- `name_section.content` = `["A simple node."]`
- `name_section.subsections` = empty list
- `public` absent
- `agent` absent
- `private` absent

---

### Full node — all section types

Setup: Create a node file for `SPEC/payments/fees` with:
  Frontmatter block (`---` ... `---`)
  `# SPEC/payments/fees`
  `Fees node description.`
  `# Public`
  `## Interface`
  `Interface line.`
  `## Constraints`
  `Constraints line.`
  `# Agent`
  `Agent guidance line.`
  `# Private`
  `## Decisions`
  `Decisions line.`
  `## Rationale`
  `Rationale line.`

Action: Call `NodeParse("SPEC/payments/fees")`.

Expected outcome:
- `name_section.heading` = `"spec/payments/fees"`
- `name_section.content` = `["Fees node description."]`
- `public` present
- `public.content` = empty list
- `public.subsections` has two entries in order:
  - heading `"interface"`, content `["Interface line."]`
  - heading `"constraints"`, content `["Constraints line."]`
- `agent` present
- `agent.content` = `["Agent guidance line."]`
- `private` present
- `private.subsections` has two entries in order:
  - heading `"decisions"`, content `["Decisions line."]`
  - heading `"rationale"`, content `["Rationale line."]`

---

### Node with no public section

Setup: Create a node file for `SPEC/decisions` with:
  `# SPEC/decisions`
  `Decision description.`
  `# Private`
  `## Rationale`
  `Rationale content.`

Action: Call `NodeParse("SPEC/decisions")`.

Expected outcome:
- `public` absent
- `agent` absent
- `private` present with one subsection:
  - heading `"rationale"`, content `["Rationale content."]`

---

### Public section with content before first subsection

Setup: Create a node file for `SPEC/a` with:
  `# SPEC/a`
  `Node content.`
  `# Public`
  `Preamble line one.`
  `Preamble line two.`
  `## Interface`
  `Interface line.`

Action: Call `NodeParse("SPEC/a")`.

Expected outcome:
- `public.content` = `["Preamble line one.", "Preamble line two."]`
- `public.subsections` has one entry:
  - heading `"interface"`, content `["Interface line."]`

---

### Public section with no content or subsections

Setup: Create a node file with:
  `# SPEC/b`
  `Node content.`
  `# Public`
  `# Agent`
  `Agent line.`

Action: Call `NodeParse("SPEC/b")`.

Expected outcome:
- `public` present
- `public.content` = empty list
- `public.subsections` = empty list

---

### Agent section with subsections

Setup: Create a node file with:
  `# SPEC/c`
  `Node content.`
  `# Agent`
  `Preamble line.`
  `## Implementation guidance`
  `Guidance content.`
  `## Contracts`
  `Contracts content.`

Action: Call `NodeParse("SPEC/c")`.

Expected outcome:
- `agent.content` = `["Preamble line."]`
- `agent.raw_heading` = `"# Agent"`
- `agent.subsections` has two entries:
  - heading `"implementation guidance"`, content `["Guidance content."]`
  - heading `"contracts"`, content `["Contracts content."]`

---

### Private section with subsections

Setup: Create a node file with:
  `# SPEC/d`
  `Node content.`
  `# Private`
  `## TODO`
  `Todo content.`
  `## Decisions`
  `Decisions content.`
  `## Rationale`
  `Rationale content.`

Action: Call `NodeParse("SPEC/d")`.

Expected outcome:
- `private` present with three subsections in order:
  - heading `"todo"`
  - heading `"decisions"`
  - heading `"rationale"`

---

### Content is raw markdown

Setup: Create a node file with:
  `# SPEC/f`
  `Node content.`
  `# Public`
  `## Interface`
  `### A level-3 heading`
  `**bold text**`
  ` ``` `
  `code here`
  ` ``` `

Action: Call `NodeParse("SPEC/f")`.

Expected outcome:
- `public.subsections[0].content` = `["### A level-3 heading", "**bold text**", "` ``` `", "code here", "` ``` `"]`
  (each line as a raw string, no structural interpretation)

---

## Heading normalization

### Case insensitive public detection

Setup: Create a node file with:
  `# SPEC/g`
  `Node content.`
  `# PUBLIC`

Action: Call `NodeParse("SPEC/g")`.

Expected outcome:
- `public` present
- `public.heading` = `"public"`

---

### Public with mixed case and extra whitespace

Setup: Create a node file with:
  `# SPEC/h`
  `Node content.`
  `#   PuBLiC`

Action: Call `NodeParse("SPEC/h")`.

Expected outcome:
- `public` present
- `public.heading` = `"public"`

---

### Node name with varied whitespace

Setup: Create a node file with:
  `#   SPEC/e`
  `Node content.`

Action: Call `NodeParse("SPEC/e")`.

Expected outcome:
- `name_section.heading` = `"spec/e"`

---

### Node name with ROOT/ heading does not match SPEC/

Setup: Create a node file with:
  `# ROOT/x`
  `Node content.`

Action: Call `NodeParse("SPEC/x")`.

Expected outcome:
- Error NodeNameDoesNotMatch (`"root/x"` does not match `"spec/x"`)

---

### Subsection headings are normalized

Setup: Create a node file with:
  `# SPEC/i`
  `Node content.`
  `# Public`
  `##   Interface`
  `Interface content.`
  `## CONSTRAINTS`
  `Constraints content.`

Action: Call `NodeParse("SPEC/i")`.

Expected outcome:
- `public.subsections[0].heading` = `"interface"`
- `public.subsections[1].heading` = `"constraints"`

---

### Closing hashes are stripped

Setup: Create a node file with:
  `# SPEC/j`
  `Node content.`
  `# Public`
  `## Interface ##`
  `Interface content.`

Action: Call `NodeParse("SPEC/j")`.

Expected outcome:
- `public.subsections[0].heading` = `"interface"`
- `public.subsections[0].raw_heading` = `"## Interface ##"`

---

## Raw heading preservation

### Raw heading preserves original line

Setup: Create a node file with:
  `# SPEC/k`
  `Node content.`
  `# Public`
  `## Interface`
  `Interface content.`

Action: Call `NodeParse("SPEC/k")`.

Expected outcome:
- `public.raw_heading` = `"# Public"`
- `public.subsections[0].raw_heading` = `"## Interface"`

---

### Raw heading preserves case

Setup: Create a node file with:
  `# SPEC/l`
  `Node content.`
  `# PUBLIC`

Action: Call `NodeParse("SPEC/l")`.

Expected outcome:
- `public.heading` = `"public"`
- `public.raw_heading` = `"# PUBLIC"`

---

### Raw heading preserves closing hashes

Setup: Create a node file with:
  `# SPEC/m`
  `Node content.`
  `# Public`
  `## Foo ##`
  `Foo content.`

Action: Call `NodeParse("SPEC/m")`.

Expected outcome:
- `public.subsections[0].heading` = `"foo"`
- `public.subsections[0].raw_heading` = `"## Foo ##"`

---

### Raw heading preserves extra whitespace

Setup: Create a node file with:
  `# SPEC/n`
  `Node content.`
  `#   Public`

Action: Call `NodeParse("SPEC/n")`.

Expected outcome:
- `public.heading` = `"public"`
- `public.raw_heading` = `"#   Public"`

---

## Content boundaries

### Level-3 and deeper headings are content

Setup: Create a node file with:
  `# SPEC/o`
  `Node content.`
  `# Public`
  `## Interface`
  `### A subsub heading`
  `#### Even deeper`
  `Interface content.`

Action: Call `NodeParse("SPEC/o")`.

Expected outcome:
- `public.subsections[0].content` contains `"### A subsub heading"` and `"#### Even deeper"` as content lines

---

### Fenced code blocks with heading-like content (backtick fence)

Setup: Create a node file with:
  `# SPEC/p`
  `Node content.`
  `# Public`
  `## Interface`
  ` ``` `
  `# looks like heading`
  `## also heading-like`
  ` ``` `
  `Normal content.`

Action: Call `NodeParse("SPEC/p")`.

Expected outcome:
- `public.subsections[0].content` includes the lines starting with `#` and `##` as raw content
- No additional sections or subsections are created for those lines

---

### Fenced code block with tilde fence

Setup: Create a node file with a subsection whose content contains:
  ` ~~~ `
  `# looks like heading`
  ` ~~~ `

Action: Call `NodeParse` on the file.

Expected outcome:
- The line `"# looks like heading"` is treated as content, not a structural heading

---

### Fenced code block with language tag

Setup: Create a node file with a subsection whose content contains:
  ` ```python `
  `# python comment that looks like heading`
  ` ``` `

Action: Call `NodeParse` on the file.

Expected outcome:
- The line `"# python comment that looks like heading"` is treated as content, not a structural heading

---

### Blank lines between heading and content are preserved

Setup: Create a node file with:
  `# SPEC/q`
  `Node content.`
  `# Public`
  (blank line)
  `Public content.`

Action: Call `NodeParse("SPEC/q")`.

Expected outcome:
- `public.content` = `["", "Public content."]`
  (first element is an empty string representing the blank line)

---

## Frontmatter handling

### Frontmatter is skipped

Setup: Create a node file with:
  `---`
  `depends_on: []`
  `---`
  `# SPEC/r`
  `Body content.`

Action: Call `NodeParse("SPEC/r")`.

Expected outcome:
- No error
- `name_section.content` = `["Body content."]`
- Frontmatter not present in any content lists

---

### No frontmatter delimiters

Setup: Create a node file with no `---` delimiters:
  `# SPEC/s`
  `Body content.`

Action: Call `NodeParse("SPEC/s")`.

Expected outcome:
- No error
- `name_section.content` = `["Body content."]`

---

### Unclosed frontmatter

Setup: Create a node file with:
  `---`
  `depends_on: []`
  (no closing `---`, rest is body-like content)

Action: Call `NodeParse("SPEC/s2")`.

Expected outcome:
- Error UnexpectedContentBeforeFirstHeading

---

## Failure cases

### ARTIFACT reference rejected

Action: Call `NodeParse("ARTIFACT/x")`.

Expected outcome:
- Error NotASpecReference

---

### EXTERNAL reference rejected

Action: Call `NodeParse("EXTERNAL/x")`.

Expected outcome:
- Error NotASpecReference

---

### Qualifier rejected

Action: Call `NodeParse("SPEC/x(interface)")`.

Expected outcome:
- Error HasQualifier

---

### File does not exist

Action: Call `NodeParse` with a logical name whose corresponding file does not exist on disk.

Expected outcome:
- Error FileUnreadable

---

### Propagates path errors

Action: Call `NodeParse` with a logical name that when resolved produces a path error (e.g., traversal components).

Expected outcome:
- The path error is propagated from the path resolution step

---

### Content before first heading

Setup: Create a node file with:
  `Some text before heading.`
  `# SPEC/t`
  `Content.`

Action: Call `NodeParse("SPEC/t")`.

Expected outcome:
- Error UnexpectedContentBeforeFirstHeading

---

### Level-2 heading before any level-1 heading

Setup: Create a node file with:
  `## Interface`
  `Content.`

Action: Call `NodeParse("SPEC/u")`.

Expected outcome:
- Error UnexpectedContentBeforeFirstHeading

---

### Empty body

Setup: Create a node file with only frontmatter and no body, or with an entirely empty file.

Action: Call `NodeParse("SPEC/v")`.

Expected outcome:
- Error UnexpectedContentBeforeFirstHeading

---

### Node name does not match logical name

Setup: Create a node file with heading `# SPEC/wrong-name`.

Action: Call `NodeParse("SPEC/correct-name")`.

Expected outcome:
- Error NodeNameDoesNotMatch

---

### Node name case mismatch is not an error

Setup: Create a node file with heading `# spec/mynode` (lowercase).

Action: Call `NodeParse("SPEC/MYNODE")`.

Expected outcome:
- No error — normalization makes both `"spec/mynode"` and they match

---

### Duplicate public section — same case

Setup: Create a node file with:
  `# SPEC/w`
  `Content.`
  `# Public`
  `Public content.`
  `# Public`
  `More public content.`

Action: Call `NodeParse("SPEC/w")`.

Expected outcome:
- Error DuplicatePublicSection

---

### Duplicate public section — different case

Setup: Create a node file with:
  `# SPEC/ww`
  `Content.`
  `# Public`
  `Public content.`
  `# PUBLIC`
  `More public content.`

Action: Call `NodeParse("SPEC/ww")`.

Expected outcome:
- Error DuplicatePublicSection

---

### Duplicate agent section

Setup: Create a node file with two `# Agent` headings.

Action: Call `NodeParse` on the file.

Expected outcome:
- Error DuplicateAgentSection

---

### Duplicate private section

Setup: Create a node file with two `# Private` headings.

Action: Call `NodeParse` on the file.

Expected outcome:
- Error DuplicatePrivateSection

---

### Unrecognized section heading

Setup: Create a node file with:
  `# SPEC/y`
  `Content.`
  `# Decisions`
  `Decisions content.`

Action: Call `NodeParse("SPEC/y")`.

Expected outcome:
- Error UnrecognizedSection

---

### Unrecognized section — Rationale

Setup: Create a node file with a `# Rationale` top-level heading (not inside Private).

Action: Call `NodeParse` on the file.

Expected outcome:
- Error UnrecognizedSection

---

### Unrecognized section — TODO

Setup: Create a node file with a `# TODO` top-level heading.

Action: Call `NodeParse` on the file.

Expected outcome:
- Error UnrecognizedSection

---

### Duplicate subsection in public — same case

Setup: Create a node file with:
  `# SPEC/z`
  `Content.`
  `# Public`
  `## Interface`
  `Content A.`
  `## Interface`
  `Content B.`

Action: Call `NodeParse("SPEC/z")`.

Expected outcome:
- Error DuplicateSubsection

---

### Duplicate subsection in public — different case

Setup: Create a node file with:
  `# SPEC/z2`
  `Content.`
  `# Public`
  `## Interface`
  `Content A.`
  `## INTERFACE`
  `Content B.`

Action: Call `NodeParse("SPEC/z2")`.

Expected outcome:
- Error DuplicateSubsection

---

### Duplicate subsection in public — whitespace variation

Setup: Create a node file with:
  `# SPEC/z3`
  `Content.`
  `# Public`
  `## Interface`
  `Content A.`
  `##   Interface`
  `Content B.`

Action: Call `NodeParse("SPEC/z3")`.

Expected outcome:
- Error DuplicateSubsection

---

### Duplicate subsection in agent

Setup: Create a node file with:
  `# SPEC/z4`
  `Content.`
  `# Agent`
  `## Guidance`
  `Guidance content.`
  `## Guidance`
  `Duplicate content.`

Action: Call `NodeParse("SPEC/z4")`.

Expected outcome:
- Error DuplicateSubsection
