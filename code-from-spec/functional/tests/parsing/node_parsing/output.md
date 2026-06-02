<!-- code-from-spec: ROOT/functional/tests/parsing/node_parsing@Xjm3_jzZjCBah8XVPIUpWvuXpeA -->

# Test Specification: NodeParse

## Happy Path

### Minimal node — name section only

Setup: Create a node file for `ROOT/x` containing:
```
# ROOT/x
A simple node.
```

Action: Call `NodeParse` with `"ROOT/x"`.

Expected outcome:
- `name_section.heading` = `"root/x"`
- `name_section.raw_heading` = `"# ROOT/x"`
- `name_section.content` = `["A simple node."]`
- `name_section.subsections` = empty list
- `public` absent
- `agent` absent
- `private` = empty list

No error.

---

### Full node — all section types

Setup: Create a node file for `ROOT/payments/fees` containing:
```
---
output: some/output.md
---
# ROOT/payments/fees
Description of this node.
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

Action: Call `NodeParse` with `"ROOT/payments/fees"`.

Expected outcome:
- `name_section.heading` = `"root/payments/fees"`, content = `["Description of this node."]`
- `public` present with `content` = empty list (no lines before first subsection), and two subsections:
  - `heading` = `"interface"`, content = `["Interface content line."]`
  - `heading` = `"constraints"`, content = `["Constraints content line."]`
- `agent` present with content = `["Agent content line."]`
- `private` has two sections in order:
  - `heading` = `"decisions"`, content = `["Decisions content line."]`
  - `heading` = `"rationale"`, content = `["Rationale content line."]`

No error.

---

### Node with no public section

Setup: Create a node file for `ROOT/decisions` containing:
```
# ROOT/decisions
Description.
# Rationale
Rationale content.
```

Action: Call `NodeParse` with `"ROOT/decisions"`.

Expected outcome:
- `public` absent
- `agent` absent
- `private` has one section with `heading` = `"rationale"` and content = `["Rationale content."]`

No error.

---

### Public section with content before first subsection

Setup: Create a node file for `ROOT/a` containing:
```
# ROOT/a
Node description.
# Public
Preamble line one.
Preamble line two.
## Interface
Interface content.
```

Action: Call `NodeParse` with `"ROOT/a"`.

Expected outcome:
- `public.content` = `["Preamble line one.", "Preamble line two."]`
- `public.subsections` has one entry with `heading` = `"interface"` and content = `["Interface content."]`

No error.

---

### Public section with no content or subsections

Setup: Create a node file containing:
```
# ROOT/b
Node description.
# Public
# Agent
Agent content.
```

Action: Call `NodeParse` with `"ROOT/b"`.

Expected outcome:
- `public` present with `content` = empty list and `subsections` = empty list
- `agent` present with content = `["Agent content."]`

No error.

---

### Agent section with subsections

Setup: Create a node file containing:
```
# ROOT/c
Node description.
# Agent
Agent preamble line.
## Implementation guidance
Implementation content.
## Contracts
Contracts content.
```

Action: Call `NodeParse` with `"ROOT/c"`.

Expected outcome:
- `agent.content` = `["Agent preamble line."]`
- `agent.raw_heading` = `"# Agent"`
- `agent.subsections` has two entries:
  - `heading` = `"implementation guidance"`, content = `["Implementation content."]`
  - `heading` = `"contracts"`, content = `["Contracts content."]`

No error.

---

### Private sections preserve file order

Setup: Create a node file containing:
```
# ROOT/d
Node description.
# TODO
Todo content.
# Decisions
Decisions content.
# Rationale
Rationale content.
```

Action: Call `NodeParse` with `"ROOT/d"`.

Expected outcome:
- `private` has three sections in order: `"todo"`, `"decisions"`, `"rationale"`

No error.

---

### Content is raw markdown

Setup: Create a node file containing:
```
# ROOT/e
Node description.
# Public
## Interface
### Sub-heading
**bold text**
` `` `go
some code
` `` `
```
(Note: the fenced code block uses standard backtick fences.)

Action: Call `NodeParse` with `"ROOT/e"`.

Expected outcome:
- The Interface subsection content is a list containing the `### Sub-heading` line, the `**bold text**` line, the opening fence line, the `some code` line, and the closing fence line — all as raw strings.

No error.

---

## Heading Normalization

### Case insensitive public detection

Setup: Create a node file containing:
```
# ROOT/f
Description.
# PUBLIC
Public content.
```

Action: Call `NodeParse` with `"ROOT/f"`.

Expected outcome: `public` present with `heading` = `"public"`. No error.

---

### Public with mixed case and extra whitespace

Setup: Create a node file containing:
```
# ROOT/g
Description.
#   PuBLiC
Public content.
```

Action: Call `NodeParse` with `"ROOT/g"`.

Expected outcome: `public` present with `heading` = `"public"`. No error.

---

### Node name with varied whitespace

Setup: Create a node file containing:
```
#   ROOT/e
Description.
```

Action: Call `NodeParse` with `"ROOT/e"`.

Expected outcome: `name_section.heading` = `"root/e"`. No error.

---

### Subsection headings are normalized

Setup: Create a node file containing:
```
# ROOT/h
Description.
# Public
##   Interface
Interface content.
## CONSTRAINTS
Constraints content.
```

Action: Call `NodeParse` with `"ROOT/h"`.

Expected outcome: Subsection headings = `"interface"` and `"constraints"`. No error.

---

### Closing hashes are stripped

Setup: Create a node file containing:
```
# ROOT/i
Description.
# Public
## Interface ##
Interface content.
```

Action: Call `NodeParse` with `"ROOT/i"`.

Expected outcome:
- Subsection `heading` = `"interface"`
- Subsection `raw_heading` = `"## Interface ##"`

No error.

---

## Raw Heading Preservation

### Raw heading preserves original line

Setup: Create a node file containing:
```
# ROOT/j
Description.
# Public
## Interface
Interface content.
```

Action: Call `NodeParse` with `"ROOT/j"`.

Expected outcome:
- `public.raw_heading` = `"# Public"`
- The Interface subsection `raw_heading` = `"## Interface"`

No error.

---

### Raw heading preserves case

Setup: Create a node file containing:
```
# ROOT/k
Description.
# PUBLIC
Public content.
```

Action: Call `NodeParse` with `"ROOT/k"`.

Expected outcome:
- `public.heading` = `"public"` (normalized)
- `public.raw_heading` = `"# PUBLIC"` (original)

No error.

---

### Raw heading preserves closing hashes

Setup: Create a node file containing:
```
# ROOT/l
Description.
# Public
## Foo ##
Foo content.
```

Action: Call `NodeParse` with `"ROOT/l"`.

Expected outcome:
- Subsection `heading` = `"foo"`
- Subsection `raw_heading` = `"## Foo ##"`

No error.

---

### Raw heading preserves extra whitespace

Setup: Create a node file containing:
```
# ROOT/m
Description.
#   Public
Public content.
```

Action: Call `NodeParse` with `"ROOT/m"`.

Expected outcome:
- `public.heading` = `"public"`
- `public.raw_heading` = `"#   Public"`

No error.

---

## Content Boundaries

### Level-3 and deeper headings are content

Setup: Create a node file containing:
```
# ROOT/n
Description.
# Public
## Interface
### Sub-heading
#### Deep heading
Content line.
```

Action: Call `NodeParse` with `"ROOT/n"`.

Expected outcome: The Interface subsection content list includes the `### Sub-heading` line and the `#### Deep heading` line as raw strings, not treated as structural headings. No error.

---

### Fenced code blocks with heading-like content

Setup: Create a node file containing:
```
# ROOT/o
Description.
# Public
## Interface
` `` `
# looks like heading
## also looks like heading
` `` `
```
(Standard backtick fences.)

Action: Call `NodeParse` with `"ROOT/o"`.

Expected outcome: The `#` and `##` lines inside the code block are in the Interface subsection content list, not treated as structural headings. No error.

---

### Fenced code block with tilde fence

Setup: Create a node file containing:
```
# ROOT/p
Description.
# Public
## Interface
~~~
# looks like level-1 heading
~~~
```

Action: Call `NodeParse` with `"ROOT/p"`.

Expected outcome: The `# looks like level-1 heading` line inside the tilde-fenced block is in the Interface subsection content, not treated as a structural heading. No error.

---

### Fenced code block with language tag

Setup: Create a node file containing:
```
# ROOT/q
Description.
# Public
## Interface
` `` `go
# looks like level-1 heading
` `` `
```
(Standard backtick fence with `go` language tag.)

Action: Call `NodeParse` with `"ROOT/q"`.

Expected outcome: The `# looks like level-1 heading` line inside the code block is in the Interface subsection content, not treated as a structural heading. No error.

---

### Blank lines between heading and content are preserved

Setup: Create a node file containing:
```
# ROOT/r
Description.
# Public

Content line.
```
(One blank line between the Public heading and the content line.)

Action: Call `NodeParse` with `"ROOT/r"`.

Expected outcome: `public.content` = `["", "Content line."]` — the blank line is preserved as an empty string at index 0. No error.

---

## Frontmatter Handling

### Frontmatter is skipped

Setup: Create a node file containing:
```
---
output: some/path.md
---
# ROOT/s
Description.
```

Action: Call `NodeParse` with `"ROOT/s"`.

Expected outcome: Frontmatter is skipped. `name_section.heading` = `"root/s"`. No error.

---

### No frontmatter delimiters

Setup: Create a node file containing:
```
# ROOT/t
Description.
```
(No `---` delimiters at all.)

Action: Call `NodeParse` with `"ROOT/t"`.

Expected outcome: No error. Body parsed correctly. `name_section.heading` = `"root/t"`.

---

### Unclosed frontmatter

Setup: Create a node file starting with `---` but no closing `---`:
```
---
output: some/path.md
# ROOT/u
Description.
```

Action: Call `NodeParse` with `"ROOT/u"`.

Expected outcome: Error `UnexpectedContentBeforeFirstHeading`.

---

## Failure Cases

### ARTIFACT reference rejected

Setup: None.

Action: Call `NodeParse` with `"ARTIFACT/x"`.

Expected outcome: Error `NotARootReference`.

---

### Qualifier rejected

Setup: None.

Action: Call `NodeParse` with `"ROOT/x(interface)"`.

Expected outcome: Error `HasQualifier`.

---

### File does not exist

Setup: None.

Action: Call `NodeParse` with a logical name whose resolved file does not exist on disk.

Expected outcome: Error `FileUnreadable`.

---

### Propagates path errors

Setup: None.

Action: Call `NodeParse` with a logical name that resolves to a path with directory traversal.

Expected outcome: The path error (e.g., `DirectoryTraversal`) is propagated from `FileReader`/`PathUtils`.

---

### Content before first heading

Setup: Create a node file containing:
```
This line appears before any heading.
# ROOT/v
Description.
```

Action: Call `NodeParse` with `"ROOT/v"`.

Expected outcome: Error `UnexpectedContentBeforeFirstHeading`.

---

### Level-2 heading before any level-1 heading

Setup: Create a node file containing:
```
## Some subsection
Description.
```

Action: Call `NodeParse` with `"ROOT/w"`.

Expected outcome: Error `UnexpectedContentBeforeFirstHeading`.

---

### Empty body

Setup: Create a node file with no content (or only frontmatter, no body headings).

Action: Call `NodeParse` with `"ROOT/empty"`.

Expected outcome: Error `UnexpectedContentBeforeFirstHeading`.

---

### Node name does not match logical name

Setup: Create a node file containing:
```
# ROOT/actual/name
Description.
```

Action: Call `NodeParse` with `"ROOT/different/name"`.

Expected outcome: Error `NodeNameDoesNotMatch`.

---

### Node name case mismatch is not an error

Setup: Create a node file containing:
```
# root/x
Description.
```

Action: Call `NodeParse` with `"ROOT/X"`.

Expected outcome: No error — normalization makes `"root/x"` and `"root/x"` equal.

---

### Duplicate public section — same case

Setup: Create a node file containing:
```
# ROOT/dup
Description.
# Public
First public content.
# Public
Second public content.
```

Action: Call `NodeParse` with `"ROOT/dup"`.

Expected outcome: Error `DuplicatePublicSection`.

---

### Duplicate public section — different case

Setup: Create a node file containing:
```
# ROOT/dup2
Description.
# Public
First public content.
# PUBLIC
Second public content.
```

Action: Call `NodeParse` with `"ROOT/dup2"`.

Expected outcome: Error `DuplicatePublicSection`.

---

### Duplicate agent section

Setup: Create a node file containing:
```
# ROOT/dup3
Description.
# Agent
First agent content.
# Agent
Second agent content.
```

Action: Call `NodeParse` with `"ROOT/dup3"`.

Expected outcome: Error `DuplicateAgentSection`.

---

### Duplicate subsection in public — same case

Setup: Create a node file containing:
```
# ROOT/dup4
Description.
# Public
## Interface
Content.
## Interface
More content.
```

Action: Call `NodeParse` with `"ROOT/dup4"`.

Expected outcome: Error `DuplicateSubsection`.

---

### Duplicate subsection in public — different case

Setup: Create a node file containing:
```
# ROOT/dup5
Description.
# Public
## Interface
Content.
## INTERFACE
More content.
```

Action: Call `NodeParse` with `"ROOT/dup5"`.

Expected outcome: Error `DuplicateSubsection`.

---

### Duplicate subsection in public — whitespace variation

Setup: Create a node file containing:
```
# ROOT/dup6
Description.
# Public
## Interface
Content.
##   Interface
More content.
```

Action: Call `NodeParse` with `"ROOT/dup6"`.

Expected outcome: Error `DuplicateSubsection`.

---

### Duplicate subsection in agent

Setup: Create a node file containing:
```
# ROOT/dup7
Description.
# Agent
## Rules
Content.
## Rules
More content.
```

Action: Call `NodeParse` with `"ROOT/dup7"`.

Expected outcome: Error `DuplicateSubsection`.
