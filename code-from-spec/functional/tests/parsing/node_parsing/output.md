<!-- code-from-spec: ROOT/functional/tests/parsing/node_parsing@EH7BlPmP0KKtO0dxFX0Q3BsIY7Q -->

# Node Parsing Tests

Each test case lists: **Setup**, **Action**, and **Expected Outcome**.

---

## Happy Path

---

### 1. Minimal node — name section only

**Setup**
Create a node file for `ROOT/x` with the following body:

```
# ROOT/x

A simple node.
```

**Action**
Call `NodeParse("ROOT/x")`.

**Expected Outcome**
- Returns a Node record with no error.
- `name_section.heading` = `"root/x"`
- `name_section.content` = `"A simple node."`
- `name_section.subsections` is empty.
- `public` is absent.
- `agent` is absent.
- `private` is empty.

---

### 2. Full node — name, public, agent, private

**Setup**
Create a node file for `ROOT/payments/fees` with frontmatter and the following body:

```
# ROOT/payments/fees

Overview of fees.

# Public

General public description.

## Interface

The interface details.

## Constraints

The constraints.

# Agent

Agent guidance here.

# Decisions

Decision log.

# Rationale

Why these decisions.
```

**Action**
Call `NodeParse("ROOT/payments/fees")`.

**Expected Outcome**
- Returns a Node record with no error.
- `name_section.heading` = `"root/payments/fees"`
- `name_section.content` = `"Overview of fees."`
- `public` is present.
  - `public.heading` = `"public"`
  - `public.subsections` has two entries:
    - first: heading = `"interface"`, content = `"The interface details."`
    - second: heading = `"constraints"`, content = `"The constraints."`
- `agent` is present.
  - `agent.heading` = `"agent"`
  - `agent.content` = `"Agent guidance here."`
- `private` has two sections in file order:
  - first: heading = `"decisions"`, content = `"Decision log."`
  - second: heading = `"rationale"`, content = `"Why these decisions."`

---

### 3. Node with no public section

**Setup**
Create a node file for `ROOT/decisions` with the following body:

```
# ROOT/decisions

Node overview.

# Rationale

Why this exists.
```

**Action**
Call `NodeParse("ROOT/decisions")`.

**Expected Outcome**
- Returns a Node record with no error.
- `public` is absent.
- `agent` is absent.
- `private` has one section:
  - heading = `"rationale"`, content = `"Why this exists."`

---

### 4. Public section with content before first subsection

**Setup**
Create a node file for `ROOT/a` with the following body:

```
# ROOT/a

Name content.

# Public

Introductory public text.

## Interface

Interface details.
```

**Action**
Call `NodeParse("ROOT/a")`.

**Expected Outcome**
- Returns a Node record with no error.
- `public.content` = `"Introductory public text."`
- `public.subsections` has one entry:
  - heading = `"interface"`, content = `"Interface details."`

---

### 5. Public section with no content or subsections

**Setup**
Create a node file where `# Public` is immediately followed by `# Agent`:

```
# ROOT/b

Name content.

# Public

# Agent

Agent content.
```

**Action**
Call `NodeParse("ROOT/b")`.

**Expected Outcome**
- Returns a Node record with no error.
- `public` is present.
  - `public.content` is empty.
  - `public.subsections` is empty.
- `agent` is present.
  - `agent.content` = `"Agent content."`

---

### 6. Agent section with ## headings treated as content

**Setup**
Create a node file with the following body:

```
# ROOT/c

Name content.

# Agent

Some preamble.

## Implementation guidance

Details here.

## Contracts

Contract details.
```

**Action**
Call `NodeParse("ROOT/c")`.

**Expected Outcome**
- Returns a Node record with no error.
- `agent.content` includes the raw `## Implementation guidance` and
  `## Contracts` lines and their text as part of the content string.
- `agent.subsections` is empty (the record has no subsections field
  populated — `##` headings are content inside `# Agent`).

---

### 7. Private sections preserve file order

**Setup**
Create a node file with the following body:

```
# ROOT/d

Name content.

# TODO

Todo items.

# Decisions

Decision log.

# Rationale

Rationale text.
```

**Action**
Call `NodeParse("ROOT/d")`.

**Expected Outcome**
- Returns a Node record with no error.
- `private` has three sections in order:
  - first: heading = `"todo"`
  - second: heading = `"decisions"`
  - third: heading = `"rationale"`

---

### 8. Content is raw markdown

**Setup**
Create a node file where a public subsection contains level-3 headings,
bold text, and a fenced code block:

```
# ROOT/f

Name content.

# Public

## Interface

### Details

Some **bold** text.

```go
func example() {}
```

#### Sub-details

More text.
```

**Action**
Call `NodeParse("ROOT/f")`.

**Expected Outcome**
- Returns a Node record with no error.
- `public.subsections` has one entry with heading `"interface"`.
- The subsection `content` is raw markdown text that includes:
  - The `### Details` line.
  - The `**bold**` text.
  - The fenced code block (opening fence, body, closing fence).
  - The `#### Sub-details` line.
  - No structural interpretation of those lines.

---

## Heading Normalization

---

### 9. Case insensitive public detection

**Setup**
Create a node file with `# PUBLIC` as the public heading:

```
# ROOT/g

Name content.

# PUBLIC

Public content.
```

**Action**
Call `NodeParse("ROOT/g")`.

**Expected Outcome**
- Returns a Node record with no error.
- `public` is present.
- `public.heading` = `"public"`.

---

### 10. Public with mixed case and extra whitespace

**Setup**
Create a node file with `#   PuBLiC` as the public heading:

```
# ROOT/h

Name content.

#   PuBLiC

Public content.
```

**Action**
Call `NodeParse("ROOT/h")`.

**Expected Outcome**
- Returns a Node record with no error.
- `public` is present.
- `public.heading` = `"public"`.

---

### 11. Node name with varied whitespace

**Setup**
Create a node file with `#    ROOT/e` as the name heading:

```
#    ROOT/e

Name content.
```

**Action**
Call `NodeParse("ROOT/e")`.

**Expected Outcome**
- Returns a Node record with no error.
- `name_section.heading` = `"root/e"`.

---

### 12. Subsection headings are normalized

**Setup**
Create a node file with subsections using varied case:

```
# ROOT/i

Name content.

# Public

## Interface

Interface content.

##   CONSTRAINTS

Constraints content.
```

**Action**
Call `NodeParse("ROOT/i")`.

**Expected Outcome**
- Returns a Node record with no error.
- `public.subsections` has two entries:
  - first: heading = `"interface"`
  - second: heading = `"constraints"`

---

### 13. Closing hashes are stripped

**Setup**
Create a node file with a subsection heading that has trailing hashes:

```
# ROOT/j

Name content.

# Public

## Interface ##

Interface content.
```

**Action**
Call `NodeParse("ROOT/j")`.

**Expected Outcome**
- Returns a Node record with no error.
- `public.subsections` has one entry.
- Subsection heading = `"interface"`.

---

## Content Boundaries

---

### 14. Level-3 and deeper headings are content

**Setup**
Create a node file where a public subsection contains `###` and `####` headings:

```
# ROOT/k

Name content.

# Public

## Overview

### Details

Detail text.

#### Sub-details

Sub-detail text.
```

**Action**
Call `NodeParse("ROOT/k")`.

**Expected Outcome**
- Returns a Node record with no error.
- `public.subsections` has one entry with heading `"overview"`.
- The subsection content includes the raw `### Details` line,
  `"Detail text."`, the raw `#### Sub-details` line, and `"Sub-detail text."`.
- No additional structural sections are created for `###` or `####`.

---

### 15. Fenced code blocks with heading-like content (backtick fence)

**Setup**
Create a node file with a public subsection containing a fenced code block
that has lines starting with `#` and `##`:

```
# ROOT/l

Name content.

# Public

## Notes

Some text.

```
# This looks like a heading
## This too
```

More text.
```

**Action**
Call `NodeParse("ROOT/l")`.

**Expected Outcome**
- Returns a Node record with no error.
- `public.subsections` has one entry with heading `"notes"`.
- The lines `# This looks like a heading` and `## This too`
  are included as raw content within the subsection — they are
  not treated as structural headings.

---

### 16. Fenced code block with tilde fence

**Setup**
Create a node file with a public subsection containing a tilde-fenced
code block with a `# Heading` line inside:

```
# ROOT/m

Name content.

# Public

## Notes

~~~
# Heading
~~~

More text.
```

**Action**
Call `NodeParse("ROOT/m")`.

**Expected Outcome**
- Returns a Node record with no error.
- The `# Heading` line inside the tilde fence is treated as content,
  not as a structural heading.
- `public.subsections` has one entry with heading `"notes"`.

---

### 17. Fenced code block with language tag

**Setup**
Create a node file with a public subsection containing a fenced code block
opened with ` ```yaml ` that contains a `# Heading` line:

```
# ROOT/n

Name content.

# Public

## Notes

```yaml
# Heading
```

More text.
```

**Action**
Call `NodeParse("ROOT/n")`.

**Expected Outcome**
- Returns a Node record with no error.
- The `# Heading` line inside the `yaml` fenced block is treated as
  content, not as a structural heading.
- `public.subsections` has one entry with heading `"notes"`.

---

### 18. Leading and trailing blank lines are trimmed

**Setup**
Create a node file with blank lines surrounding content in sections
and subsections:

```
# ROOT/o


Name content with leading and trailing blanks.


# Public


Public content with blanks.


## Interface


Interface content with blanks.


```

**Action**
Call `NodeParse("ROOT/o")`.

**Expected Outcome**
- Returns a Node record with no error.
- `name_section.content` has no leading or trailing blank lines.
- `public.content` has no leading or trailing blank lines.
- `public.subsections[0].content` has no leading or trailing blank lines.

---

## Frontmatter Handling

---

### 19. Frontmatter is skipped

**Setup**
Create a node file with frontmatter delimiters followed by a body:

```
---
depends_on:
  - ROOT/other
---

# ROOT/p

Name content.

# Public

Public content.
```

**Action**
Call `NodeParse("ROOT/p")`.

**Expected Outcome**
- Returns a Node record with no error.
- Frontmatter is skipped entirely.
- `name_section.heading` = `"root/p"`.
- `public` is present with content `"Public content."`.

---

### 20. No frontmatter delimiters

**Setup**
Create a node file with no `---` lines — body only:

```
# ROOT/q

Name content.
```

**Action**
Call `NodeParse("ROOT/q")`.

**Expected Outcome**
- Returns a Node record with no error.
- `name_section.heading` = `"root/q"`.
- `name_section.content` = `"Name content."`.

---

### 21. Unclosed frontmatter

**Setup**
Create a node file that starts with `---` but has no closing `---`:

```
---
depends_on:
  - ROOT/other

# ROOT/r

Name content.
```

**Action**
Call `NodeParse("ROOT/r")`.

**Expected Outcome**
- Raises error `"unexpected content before first heading"`.

---

## Failure Cases

---

### 22. ARTIFACT reference rejected

**Setup**
No file setup needed.

**Action**
Call `NodeParse("ARTIFACT/x(y)")`.

**Expected Outcome**
- Raises error `"not a ROOT reference"`.

---

### 23. Qualifier rejected

**Setup**
No file setup needed.

**Action**
Call `NodeParse("ROOT/x(interface)")`.

**Expected Outcome**
- Raises error `"has qualifier"`.

---

### 24. File does not exist

**Setup**
No file setup needed. Use a logical name whose resolved file path
does not exist on disk.

**Action**
Call `NodeParse` with that logical name.

**Expected Outcome**
- Raises error `"file unreadable"`.

---

### 25. Propagates path errors

**Setup**
No file setup needed. Use a logical name that, after path resolution,
produces a path error (e.g., a traversal attempt).

**Action**
Call `NodeParse` with that logical name.

**Expected Outcome**
- The path error from `FileOpen` is propagated as-is (not wrapped
  or replaced with a different error message).

---

### 26. Content before first heading

**Setup**
Create a node file with non-blank text before any heading:

```
This is unexpected text.

# ROOT/s

Name content.
```

**Action**
Call `NodeParse("ROOT/s")`.

**Expected Outcome**
- Raises error `"unexpected content before first heading"`.

---

### 27. Level-2 heading before any level-1 heading

**Setup**
Create a node file where a `##` heading appears before any `#` heading:

```
## Subsection before any section

# ROOT/t

Name content.
```

**Action**
Call `NodeParse("ROOT/t")`.

**Expected Outcome**
- Raises error `"unexpected content before first heading"`.

---

### 28. Empty body

**Setup**
Create a node file with no body content (either completely empty,
or containing only a frontmatter block with no body following it).

**Action**
Call `NodeParse` with the corresponding logical name.

**Expected Outcome**
- Raises error `"unexpected content before first heading"`.

---

### 29. Node name does not match logical name

**Setup**
Create a node file where the first heading is `# ROOT/other`:

```
# ROOT/other

Name content.
```

**Action**
Call `NodeParse("ROOT/x")`.

**Expected Outcome**
- Raises error `"node name does not match"`.

---

### 30. Node name case mismatch is not an error

**Setup**
Create a node file with heading `# root/x` (lowercase):

```
# root/x

Name content.
```

**Action**
Call `NodeParse("ROOT/x")`.

**Expected Outcome**
- Returns a Node record with no error.
- `name_section.heading` = `"root/x"`.
- Normalization makes the heading and the logical name equal.

---

### 31. Duplicate public section — same case

**Setup**
Create a node file with two `# Public` sections:

```
# ROOT/u

Name content.

# Public

First public.

# Public

Second public.
```

**Action**
Call `NodeParse("ROOT/u")`.

**Expected Outcome**
- Raises error `"duplicate public section"`.

---

### 32. Duplicate public section — different case

**Setup**
Create a node file with `# Public` and `# PUBLIC`:

```
# ROOT/v

Name content.

# Public

First.

# PUBLIC

Second.
```

**Action**
Call `NodeParse("ROOT/v")`.

**Expected Outcome**
- Raises error `"duplicate public section"`.

---

### 33. Duplicate agent section

**Setup**
Create a node file with two `# Agent` sections:

```
# ROOT/w

Name content.

# Agent

First agent.

# Agent

Second agent.
```

**Action**
Call `NodeParse("ROOT/w")`.

**Expected Outcome**
- Raises error `"duplicate agent section"`.

---

### 34. Duplicate subsection in public — same case

**Setup**
Create a node file with two `## Interface` subsections under `# Public`:

```
# ROOT/aa

Name content.

# Public

## Interface

First interface.

## Interface

Second interface.
```

**Action**
Call `NodeParse("ROOT/aa")`.

**Expected Outcome**
- Raises error `"duplicate subsection"`.

---

### 35. Duplicate subsection in public — different case

**Setup**
Create a node file with `## Interface` and `## INTERFACE` under `# Public`:

```
# ROOT/bb

Name content.

# Public

## Interface

First.

## INTERFACE

Second.
```

**Action**
Call `NodeParse("ROOT/bb")`.

**Expected Outcome**
- Raises error `"duplicate subsection"`.

---

### 36. Duplicate subsection in public — whitespace variation

**Setup**
Create a node file with `## Interface` and `##   Interface` under `# Public`:

```
# ROOT/cc

Name content.

# Public

## Interface

First.

##   Interface

Second.
```

**Action**
Call `NodeParse("ROOT/cc")`.

**Expected Outcome**
- Raises error `"duplicate subsection"`.

---

### 37. Duplicate ## in non-public is not an error

**Setup**
Create a node file with two `## Details` headings inside `# Agent`:

```
# ROOT/dd

Name content.

# Agent

## Details

First details.

## Details

Second details.
```

**Action**
Call `NodeParse("ROOT/dd")`.

**Expected Outcome**
- Returns a Node record with no error.
- `agent.content` includes both `## Details` headings and their
  text as raw content.
- `agent.subsections` is empty.
