<!-- code-from-spec: ROOT/functional/tests/parsing/node_parsing@UoRNLjTYyMLUKrycB5gsHVCN1Rw -->

# NodeParse Test Specification

Each test case is described with: setup (the file content and call arguments),
actions (which function to call), and expected outcome.

---

## Happy Path

---

### HP-01: Minimal node — name section only

**Setup**

Create a node file for logical name `ROOT/x`.
File body (no frontmatter):

```
# ROOT/x
A simple node.
```

**Action**

Call `NodeParse("ROOT/x")`.

**Expected outcome**

No error.
- `name_section.heading` = `"root/x"`
- `name_section.raw_heading` = `"# ROOT/x"`
- `name_section.content` = `["A simple node."]`
- `name_section.subsections` = empty list
- `public` absent
- `agent` absent
- `private` = empty list

---

### HP-02: Full node — all section types

**Setup**

Create a node file for logical name `ROOT/payments/fees`.
File body (with frontmatter block at top, any valid YAML between `---` delimiters):

```
---
(any valid frontmatter)
---
# ROOT/payments/fees
Fees description.
# Public
## Interface
Interface content.
## Constraints
Constraints content.
# Agent
Agent content.
# Decisions
Decisions content.
# Rationale
Rationale content.
```

**Action**

Call `NodeParse("ROOT/payments/fees")`.

**Expected outcome**

No error.
- `name_section.heading` = `"root/payments/fees"`
- `name_section.content` = `["Fees description."]`
- `public` present
  - `public.content` = empty list (no lines before first `##`)
  - `public.subsections` has two entries in order:
    - entry 0: `heading` = `"interface"`, `content` = `["Interface content."]`
    - entry 1: `heading` = `"constraints"`, `content` = `["Constraints content."]`
- `agent` present
  - `agent.content` = `["Agent content."]`
- `private` has two sections in order:
  - section 0: `heading` = `"decisions"`, `content` = `["Decisions content."]`
  - section 1: `heading` = `"rationale"`, `content` = `["Rationale content."]`

---

### HP-03: Node with no public section

**Setup**

Create a node file for logical name `ROOT/decisions`.
File body:

```
# ROOT/decisions
Some decision content.
# Rationale
Rationale content.
```

**Action**

Call `NodeParse("ROOT/decisions")`.

**Expected outcome**

No error.
- `public` absent
- `agent` absent
- `private` has one section:
  - `heading` = `"rationale"`, `content` = `["Rationale content."]`

---

### HP-04: Public section with content before first subsection

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
Preamble line one.
Preamble line two.
## Interface
Interface content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public.content` = `["Preamble line one.", "Preamble line two."]`
- `public.subsections` has one entry:
  - `heading` = `"interface"`, `content` = `["Interface content."]`

---

### HP-05: Public section with no content or subsections

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
# Agent
Agent content.
```

(No lines between `# Public` and `# Agent`.)

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public` present
- `public.content` = empty list
- `public.subsections` = empty list

---

### HP-06: Agent section with subsections

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Agent
Agent preamble.
## Implementation guidance
Implementation guidance content.
## Contracts
Contracts content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `agent.content` = `["Agent preamble."]`
- `agent.raw_heading` = `"# Agent"`
- `agent.subsections` has two entries in order:
  - entry 0: `heading` = `"implementation guidance"`, `content` = `["Implementation guidance content."]`
  - entry 1: `heading` = `"contracts"`, `content` = `["Contracts content."]`

---

### HP-07: Private sections preserve file order

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# TODO
TODO content.
# Decisions
Decisions content.
# Rationale
Rationale content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `private` has three sections in order:
  - section 0: `heading` = `"todo"`
  - section 1: `heading` = `"decisions"`
  - section 2: `heading` = `"rationale"`

---

### HP-08: Content is raw markdown

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
## Interface
### Level three heading
**bold text**
` `` `go
some code
` `` `
```

(Use actual backtick fences without spaces — shown above with spaces only for
clarity in this document.)

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public.subsections` has one entry with `heading` = `"interface"`
- `public.subsections[0].content` = `["### Level three heading", "**bold text**", "` `` `go", "some code", "` `` `"]`
  (each line is the raw string as read from the file)

---

## Heading Normalization

---

### HN-01: Case insensitive public detection

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# PUBLIC
Public content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public` present
- `public.heading` = `"public"`

---

### HN-02: Public with mixed case and extra whitespace

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
#   PuBLiC
Public content.
```

(Extra spaces between `#` and `PuBLiC`.)

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public` present
- `public.heading` = `"public"`

---

### HN-03: Node name with varied whitespace

**Setup**

Create a node file for logical name `ROOT/e`.
File body:

```
#   ROOT/e
Name content.
```

(Extra spaces between `#` and `ROOT/e`.)

**Action**

Call `NodeParse("ROOT/e")`.

**Expected outcome**

No error.
- `name_section.heading` = `"root/e"`

---

### HN-04: Subsection headings are normalized

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
##   Interface
Interface content.
## CONSTRAINTS
Constraints content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public.subsections` has two entries:
  - entry 0: `heading` = `"interface"`
  - entry 1: `heading` = `"constraints"`

---

### HN-05: Closing hashes are stripped

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
## Interface ##
Interface content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public.subsections[0].heading` = `"interface"`
- `public.subsections[0].raw_heading` = `"## Interface ##"`

---

## Raw Heading Preservation

---

### RH-01: Raw heading preserves original line

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
## Interface
Interface content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public.raw_heading` = `"# Public"`
- `public.subsections[0].raw_heading` = `"## Interface"`

---

### RH-02: Raw heading preserves case

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# PUBLIC
Public content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public.heading` = `"public"`
- `public.raw_heading` = `"# PUBLIC"`

---

### RH-03: Raw heading preserves closing hashes

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
## Foo ##
Foo content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public.subsections[0].heading` = `"foo"`
- `public.subsections[0].raw_heading` = `"## Foo ##"`

---

### RH-04: Raw heading preserves extra whitespace

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
#   Public
Public content.
```

(Extra spaces between `#` and `Public`.)

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public.heading` = `"public"`
- `public.raw_heading` = `"#   Public"`

---

## Content Boundaries

---

### CB-01: Level-3 and deeper headings are content

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
## Interface
### Sub-sub heading
#### Even deeper
Interface content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public.subsections[0].heading` = `"interface"`
- `public.subsections[0].content` contains `"### Sub-sub heading"` and
  `"#### Even deeper"` as regular content lines (not treated as structural
  headings).

---

### CB-02: Fenced code blocks with heading-like content (backtick fence)

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
## Interface
` `` `
# Looks like a heading
## Also looks like a heading
` `` `
Real content.
```

(Use actual backtick fences.)

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public.subsections` has exactly one entry with `heading` = `"interface"`.
- The lines `"# Looks like a heading"` and `"## Also looks like a heading"`
  appear inside `public.subsections[0].content`, not as new sections or
  subsections.

---

### CB-03: Fenced code block with tilde fence

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
## Interface
~~~
# Inside tilde fence
~~~
Real content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `"# Inside tilde fence"` appears in `public.subsections[0].content` as a
  raw content line, not treated as a structural heading.

---

### CB-04: Fenced code block with language tag

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
## Interface
` `` `python
# Inside fenced block
` `` `
Real content.
```

(Use actual backtick fences with `python` language tag.)

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `"# Inside fenced block"` appears in `public.subsections[0].content` as a
  raw content line, not treated as a structural heading.

---

### CB-05: Blank lines between heading and content are preserved

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public

Public content line.
```

(There is one blank line between `# Public` and `Public content line.`)

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `public.content` = `["", "Public content line."]`
  (the blank line is preserved as an empty string at index 0)

---

## Frontmatter Handling

---

### FM-01: Frontmatter is skipped

**Setup**

Create a node file for logical name `ROOT/a`.
File content:

```
---
depends_on: []
---
# ROOT/a
Name content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `name_section.heading` = `"root/a"`
- `name_section.content` = `["Name content."]`
- Frontmatter content does not appear in any section content.

---

### FM-02: No frontmatter delimiters

**Setup**

Create a node file for logical name `ROOT/a` with no `---` delimiters at all.
File body:

```
# ROOT/a
Name content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

No error.
- `name_section.heading` = `"root/a"`
- `name_section.content` = `["Name content."]`

---

### FM-03: Unclosed frontmatter

**Setup**

Create a node file for logical name `ROOT/a`.
File content:

```
---
depends_on: []
# ROOT/a
Name content.
```

(Opening `---` present but no closing `---`.)

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

Error `UnexpectedContentBeforeFirstHeading`.

---

## Failure Cases

---

### FC-01: ARTIFACT reference rejected

**Setup**

No file needed.

**Action**

Call `NodeParse("ARTIFACT/x(y)")`.

**Expected outcome**

Error `NotARootReference`.

---

### FC-02: Qualifier rejected

**Setup**

No file needed.

**Action**

Call `NodeParse("ROOT/x(interface)")`.

**Expected outcome**

Error `HasQualifier`.

---

### FC-03: File does not exist

**Setup**

Use a logical name whose corresponding file does not exist on disk.

**Action**

Call `NodeParse` with that logical name (e.g., `"ROOT/does/not/exist"`).

**Expected outcome**

Error `FileUnreadable`.

---

### FC-04: Propagates path errors

**Setup**

Use a logical name that, when resolved to a file path, produces a path
error (for example, a name that results in path traversal after resolution).

**Action**

Call `NodeParse` with that logical name.

**Expected outcome**

The path error from the path resolution layer is propagated as-is.

---

### FC-05: Content before first heading

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
This line appears before any heading.
# ROOT/a
Name content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

Error `UnexpectedContentBeforeFirstHeading`.

---

### FC-06: Level-2 heading before any level-1 heading

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
## Early subsection
# ROOT/a
Name content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

Error `UnexpectedContentBeforeFirstHeading`.

---

### FC-07: Empty body

**Setup**

Create a node file for logical name `ROOT/a` with an empty body (no content,
or only a frontmatter block and nothing after it).

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

Error `UnexpectedContentBeforeFirstHeading`.

---

### FC-08: Node name does not match logical name

**Setup**

Create a node file whose first heading text is `ROOT/other`.
Call `NodeParse` with `"ROOT/a"`.

File body:

```
# ROOT/other
Some content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

Error `NodeNameDoesNotMatch`.

---

### FC-09: Node name case mismatch is not an error

**Setup**

Create a node file for logical name `ROOT/A` (uppercase in the call),
but the heading uses lowercase `root/a`.

File body:

```
# root/a
Name content.
```

**Action**

Call `NodeParse("ROOT/A")`.

**Expected outcome**

No error. Normalization makes both sides equal (`"root/a"`).

---

### FC-10: Duplicate public section — same case

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
First public content.
# Public
Second public content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

Error `DuplicatePublicSection`.

---

### FC-11: Duplicate public section — different case

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
First public content.
# PUBLIC
Second public content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

Error `DuplicatePublicSection`.

---

### FC-12: Duplicate agent section

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Agent
First agent content.
# Agent
Second agent content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

Error `DuplicateAgentSection`.

---

### FC-13: Duplicate subsection in public — same case

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
## Interface
First interface content.
## Interface
Second interface content.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

Error `DuplicateSubsection`.

---

### FC-14: Duplicate subsection in public — different case

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
## Interface
First.
## INTERFACE
Second.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

Error `DuplicateSubsection`.

---

### FC-15: Duplicate subsection in public — whitespace variation

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Public
## Interface
First.
##   Interface
Second.
```

(Extra spaces in the second heading.)

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

Error `DuplicateSubsection`.

---

### FC-16: Duplicate subsection in agent

**Setup**

Create a node file for logical name `ROOT/a`.
File body:

```
# ROOT/a
Name content.
# Agent
## Guidance
First guidance.
## Guidance
Second guidance.
```

**Action**

Call `NodeParse("ROOT/a")`.

**Expected outcome**

Error `DuplicateSubsection`.
