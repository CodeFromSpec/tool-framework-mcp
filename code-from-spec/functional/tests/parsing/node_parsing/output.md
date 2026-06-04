<!-- code-from-spec: ROOT/functional/tests/parsing/node_parsing@BFDnn0rOkS1W9vUX82g5xjFK8Ok -->

## Happy path

### Minimal node — name section only

Setup: create a node file for `ROOT/x` with the following body:
```
# ROOT/x
A simple node.
```

Action: call `NodeParse("ROOT/x")`.

Expected:
- `name_section.heading` = `"root/x"`
- `name_section.raw_heading` = `"# ROOT/x"`
- `name_section.content` = `["A simple node."]`
- `name_section.subsections` = empty list
- `public` = absent
- `agent` = absent
- `private` = empty list

---

### Full node — all section types

Setup: create a node file for `ROOT/payments/fees` with frontmatter and the following body:
```
# ROOT/payments/fees
Description line.
# Public
## Interface
Interface content line.
## Constraints
Constraints content line.
# Agent
Agent content line.
# Decisions
Decisions content line.
# Rationale
Rationale content line.
```

Action: call `NodeParse("ROOT/payments/fees")`.

Expected:
- `name_section.heading` = `"root/payments/fees"`
- `name_section.content` = `["Description line."]`
- `public` is present
- `public.content` = empty list (no lines before first subsection)
- `public.subsections` has two entries:
  - first: `heading` = `"interface"`, content = `["Interface content line."]`
  - second: `heading` = `"constraints"`, content = `["Constraints content line."]`
- `agent` is present with `content` = `["Agent content line."]`
- `private` has two sections in order:
  - first: `heading` = `"decisions"`, content = `["Decisions content line."]`
  - second: `heading` = `"rationale"`, content = `["Rationale content line."]`

---

### Node with no public section

Setup: create a node file for `ROOT/decisions` with the following body:
```
# ROOT/decisions
Description line.
# Rationale
Rationale content.
```

Action: call `NodeParse("ROOT/decisions")`.

Expected:
- `public` = absent
- `agent` = absent
- `private` has one section with `heading` = `"rationale"`

---

### Public section with content before first subsection

Setup: create a node file for `ROOT/a` with the following body:
```
# ROOT/a
Name content.
# Public
Preamble line one.
Preamble line two.
## Interface
Interface content.
```

Action: call `NodeParse("ROOT/a")`.

Expected:
- `public.content` = `["Preamble line one.", "Preamble line two."]`
- `public.subsections` has one entry with `heading` = `"interface"`, content = `["Interface content."]`

---

### Public section with no content or subsections

Setup: create a node file for `ROOT/b` with the following body:
```
# ROOT/b
Name content.
# Public
# Agent
Agent content.
```

Action: call `NodeParse("ROOT/b")`.

Expected:
- `public` is present
- `public.content` = empty list
- `public.subsections` = empty list

---

### Agent section with subsections

Setup: create a node file for `ROOT/c` with the following body:
```
# ROOT/c
Name content.
# Agent
Preamble line.
## Implementation guidance
Guidance content.
## Contracts
Contracts content.
```

Action: call `NodeParse("ROOT/c")`.

Expected:
- `agent.content` = `["Preamble line."]`
- `agent.raw_heading` = `"# Agent"`
- `agent.subsections` has two entries:
  - first: `heading` = `"implementation guidance"`, content = `["Guidance content."]`
  - second: `heading` = `"contracts"`, content = `["Contracts content."]`

---

### Private sections preserve file order

Setup: create a node file for `ROOT/d` with the following body:
```
# ROOT/d
Name content.
# TODO
Todo content.
# Decisions
Decisions content.
# Rationale
Rationale content.
```

Action: call `NodeParse("ROOT/d")`.

Expected:
- `private` has three sections in order: `heading` values `"todo"`, `"decisions"`, `"rationale"`

---

### Content is raw markdown

Setup: create a node file for `ROOT/e` with the following body:
```
# ROOT/e
Name content.
# Public
## Details
### Sub-heading
**Bold text**
` `` `go
code content
` `` `
```
(The fenced code block uses standard triple-backtick fences and a language tag.)

Action: call `NodeParse("ROOT/e")`.

Expected:
- The subsection `"details"` content is a list containing exactly:
  - the level-3 heading line (`"### Sub-heading"`)
  - the bold text line (`"**Bold text**"`)
  - the opening fence line (e.g., `` ```go ``)
  - the code content line (`"code content"`)
  - the closing fence line (`` ``` ``)

---

## Heading normalization

### Case insensitive public detection

Setup: create a node file for `ROOT/f` with heading `# PUBLIC`.

Action: call `NodeParse("ROOT/f")`.

Expected: `public` is present, `public.heading` = `"public"`.

---

### Public with mixed case and extra whitespace

Setup: create a node file for `ROOT/g` with heading `#   PuBLiC` (extra spaces after `#`).

Action: call `NodeParse("ROOT/g")`.

Expected: `public` is present, `public.heading` = `"public"`.

---

### Node name with varied whitespace

Setup: create a node file for `ROOT/e` with heading `#   ROOT/e` (extra spaces after `#`).

Action: call `NodeParse("ROOT/e")`.

Expected: `name_section.heading` = `"root/e"`.

---

### Subsection headings are normalized

Setup: create a node file for `ROOT/h` with the following body:
```
# ROOT/h
Name content.
# Public
##   Interface
Interface content.
## CONSTRAINTS
Constraints content.
```

Action: call `NodeParse("ROOT/h")`.

Expected: subsection headings = `"interface"` and `"constraints"`.

---

### Closing hashes are stripped

Setup: create a node file for `ROOT/i` with a subsection heading `## Interface ##`.

Action: call `NodeParse("ROOT/i")`.

Expected:
- subsection `heading` = `"interface"`
- subsection `raw_heading` = `"## Interface ##"`

---

## Raw heading preservation

### Raw heading preserves original line

Setup: create a node file for `ROOT/j` with headings `# Public` and `## Interface`.

Action: call `NodeParse("ROOT/j")`.

Expected:
- `public.raw_heading` = `"# Public"`
- the Interface subsection's `raw_heading` = `"## Interface"`

---

### Raw heading preserves case

Setup: create a node file for `ROOT/k` with heading `# PUBLIC`.

Action: call `NodeParse("ROOT/k")`.

Expected:
- `public.heading` = `"public"`
- `public.raw_heading` = `"# PUBLIC"`

---

### Raw heading preserves closing hashes

Setup: create a node file for `ROOT/l` with a subsection heading `## Foo ##`.

Action: call `NodeParse("ROOT/l")`.

Expected:
- subsection `heading` = `"foo"`
- subsection `raw_heading` = `"## Foo ##"`

---

### Raw heading preserves extra whitespace

Setup: create a node file for `ROOT/m` with heading `#   Public` (extra spaces).

Action: call `NodeParse("ROOT/m")`.

Expected:
- `public.heading` = `"public"`
- `public.raw_heading` = `"#   Public"`

---

## Content boundaries

### Level-3 and deeper headings are content

Setup: create a node file for `ROOT/n` with the following body:
```
# ROOT/n
Name content.
# Public
## Details
### Third level heading
#### Fourth level heading
```

Action: call `NodeParse("ROOT/n")`.

Expected: subsection `"details"` content includes `"### Third level heading"` and `"#### Fourth level heading"` as content lines, not treated as structural headings.

---

### Fenced code blocks with heading-like content (backtick)

Setup: create a node file for `ROOT/o` with a public subsection containing a backtick fenced code block that has lines starting with `#` and `##` inside.

Action: call `NodeParse("ROOT/o")`.

Expected: the `#` and `##` lines inside the code block are in the subsection content, not treated as structural headings.

---

### Fenced code block with tilde fence

Setup: create a node file for `ROOT/p` with a subsection containing a `~~~` fenced code block that has a line starting with `#` inside.

Action: call `NodeParse("ROOT/p")`.

Expected: the heading-like line inside the code block is content, not a structural heading.

---

### Fenced code block with language tag

Setup: create a node file for `ROOT/q` with a subsection containing a backtick fence with a language tag (e.g., `` ```go ``), where the code block body has a line starting with `#`.

Action: call `NodeParse("ROOT/q")`.

Expected: the heading-like line inside the code block is content, not a structural heading.

---

### Blank lines between heading and content are preserved

Setup: create a node file for `ROOT/r` with the following body:
```
# ROOT/r
Name content.
# Public

First content line.
```
(There is one blank line between `# Public` and the first content line.)

Action: call `NodeParse("ROOT/r")`.

Expected: `public.content` = `["", "First content line."]` — the blank line is the first element, followed by the content line.

---

## Frontmatter handling

### Frontmatter is skipped

Setup: create a node file for `ROOT/s` with frontmatter delimiters (`---` ... `---`) and then a body.

Action: call `NodeParse("ROOT/s")`.

Expected: no error. Frontmatter is skipped. Body is parsed correctly.

---

### No frontmatter delimiters

Setup: create a node file for `ROOT/t` with no `---` delimiters — body only.

Action: call `NodeParse("ROOT/t")`.

Expected: no error. Body is parsed correctly.

---

### Unclosed frontmatter

Setup: create a node file for `ROOT/u` that starts with `---` but has no closing `---`.

Action: call `NodeParse("ROOT/u")`.

Expected: error `UnexpectedContentBeforeFirstHeading`.

---

## Failure cases

### ARTIFACT reference rejected

Action: call `NodeParse("ARTIFACT/x")`.

Expected: error `NotARootReference`.

---

### Qualifier rejected

Action: call `NodeParse("ROOT/x(interface)")`.

Expected: error `HasQualifier`.

---

### File does not exist

Action: call `NodeParse` with a logical name whose corresponding file does not exist.

Expected: error `FileUnreadable`.

---

### Propagates path errors

Action: call `NodeParse` with a logical name that resolves to an invalid path (e.g., containing path traversal segments).

Expected: the path error from `FileOpen` is propagated.

---

### Content before first heading

Setup: create a node file for `ROOT/v` with text before any heading:
```
This is content before any heading.
# ROOT/v
```

Action: call `NodeParse("ROOT/v")`.

Expected: error `UnexpectedContentBeforeFirstHeading`.

---

### Level-2 heading before any level-1 heading

Setup: create a node file for `ROOT/w` that begins with a level-2 heading before any level-1 heading:
```
## Some subsection
# ROOT/w
```

Action: call `NodeParse("ROOT/w")`.

Expected: error `UnexpectedContentBeforeFirstHeading`.

---

### Empty body

Setup: create a node file for `ROOT/x2` with no body content (or only frontmatter, no body lines).

Action: call `NodeParse("ROOT/x2")`.

Expected: error `UnexpectedContentBeforeFirstHeading`.

---

### Node name does not match logical name

Setup: create a node file with first heading `# ROOT/actual` but call `NodeParse("ROOT/other")`.

Action: call `NodeParse("ROOT/other")`.

Expected: error `NodeNameDoesNotMatch`.

---

### Node name case mismatch is not an error

Setup: create a node file with first heading `# root/casematch` (lowercase).

Action: call `NodeParse("ROOT/CASEMATCH")` (uppercase logical name).

Expected: no error — normalization makes the heading and logical name equal.

---

### Duplicate public section — same case

Setup: create a node file for `ROOT/y` with two `# Public` sections.

Action: call `NodeParse("ROOT/y")`.

Expected: error `DuplicatePublicSection`.

---

### Duplicate public section — different case

Setup: create a node file for `ROOT/z` with `# Public` and `# PUBLIC` sections.

Action: call `NodeParse("ROOT/z")`.

Expected: error `DuplicatePublicSection`.

---

### Duplicate agent section

Setup: create a node file for `ROOT/aa` with two `# Agent` sections.

Action: call `NodeParse("ROOT/aa")`.

Expected: error `DuplicateAgentSection`.

---

### Duplicate subsection in public — same case

Setup: create a node file for `ROOT/bb` with two `## Interface` subsections under `# Public`.

Action: call `NodeParse("ROOT/bb")`.

Expected: error `DuplicateSubsection`.

---

### Duplicate subsection in public — different case

Setup: create a node file for `ROOT/cc` with subsections `## Interface` and `## INTERFACE` under `# Public`.

Action: call `NodeParse("ROOT/cc")`.

Expected: error `DuplicateSubsection`.

---

### Duplicate subsection in public — whitespace variation

Setup: create a node file for `ROOT/dd` with subsections `## Interface` and `##   Interface` (extra spaces) under `# Public`.

Action: call `NodeParse("ROOT/dd")`.

Expected: error `DuplicateSubsection`.

---

### Duplicate subsection in agent

Setup: create a node file for `ROOT/ee` with two identical subsection headings under `# Agent`.

Action: call `NodeParse("ROOT/ee")`.

Expected: error `DuplicateSubsection`.
