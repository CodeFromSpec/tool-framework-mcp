<!-- code-from-spec: ROOT/functional/tests/parsing/node_parsing@hVpxjVNPGPq9wkjk5Q1nxJ91jGo -->

# Node Parsing Tests

Each test case below describes the setup (file contents and function call),
the action (`NodeParse`), and the expected outcome.

---

## Happy Path

### TC-HP-01: Minimal node — name section only

**Setup:** Create a node file for `ROOT/x`. File body:
```
# ROOT/x
A simple node.
```

**Action:** Call `NodeParse("ROOT/x")`.

**Expected:**
- `name_section.heading` = `"root/x"`
- `name_section.raw_heading` = `"# ROOT/x"`
- `name_section.content` = `["A simple node."]`
- `name_section.subsections` = empty list
- `public` = absent
- `agent` = absent
- `private` = empty list

---

### TC-HP-02: Full node — all section types

**Setup:** Create a node file for `ROOT/payments/fees`. File body:
```
---
(frontmatter)
---
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

**Action:** Call `NodeParse("ROOT/payments/fees")`.

**Expected:**
- `name_section.heading` = `"root/payments/fees"`
- `name_section.content` = `["Description line."]`
- `public` present
  - `public.content` = `[]` (empty — no lines before first subsection)
  - `public.subsections` has two entries:
    - subsection 0: `heading` = `"interface"`, `content` = `["Interface content line."]`
    - subsection 1: `heading` = `"constraints"`, `content` = `["Constraints content line."]`
- `agent` present
  - `agent.content` = `["Agent content line."]`
- `private` has two sections in order:
  - section 0: `heading` = `"decisions"`, `content` = `["Decisions content line."]`
  - section 1: `heading` = `"rationale"`, `content` = `["Rationale content line."]`

---

### TC-HP-03: Node with no public section

**Setup:** Create a node file for `ROOT/decisions`. File body:
```
# ROOT/decisions
Description line.
# Rationale
Rationale content line.
```

**Action:** Call `NodeParse("ROOT/decisions")`.

**Expected:**
- `public` = absent
- `agent` = absent
- `private` has one section: `heading` = `"rationale"`, `content` = `["Rationale content line."]`

---

### TC-HP-04: Public section with content before first subsection

**Setup:** Create a node file for `ROOT/a`. File body:
```
# ROOT/a
Name content line.
# Public
Preamble line one.
Preamble line two.
## Interface
Interface content line.
```

**Action:** Call `NodeParse("ROOT/a")`.

**Expected:**
- `public.content` = `["Preamble line one.", "Preamble line two."]`
- `public.subsections` has one entry: `heading` = `"interface"`, `content` = `["Interface content line."]`

---

### TC-HP-05: Public section with no content or subsections

**Setup:** Create a node file. File body:
```
# ROOT/b
Name content.
# Public
# Agent
Agent content line.
```

**Action:** Call `NodeParse("ROOT/b")`.

**Expected:**
- `public` present
  - `public.content` = `[]` (empty list)
  - `public.subsections` = `[]` (empty list)

---

### TC-HP-06: Agent section with subsections

**Setup:** Create a node file for `ROOT/c`. File body:
```
# ROOT/c
Name content.
# Agent
Preamble line.
## Implementation guidance
Guidance content line.
## Contracts
Contracts content line.
```

**Action:** Call `NodeParse("ROOT/c")`.

**Expected:**
- `agent.content` = `["Preamble line."]`
- `agent.raw_heading` = `"# Agent"`
- `agent.subsections` has two entries:
  - subsection 0: `heading` = `"implementation guidance"`, `content` = `["Guidance content line."]`
  - subsection 1: `heading` = `"contracts"`, `content` = `["Contracts content line."]`

---

### TC-HP-07: Private sections preserve file order

**Setup:** Create a node file. File body:
```
# ROOT/d
Name content.
# TODO
TODO content.
# Decisions
Decisions content.
# Rationale
Rationale content.
```

**Action:** Call `NodeParse("ROOT/d")`.

**Expected:**
- `private` has three sections in order:
  - section 0: `heading` = `"todo"`
  - section 1: `heading` = `"decisions"`
  - section 2: `heading` = `"rationale"`

---

### TC-HP-08: Content is raw markdown

**Setup:** Create a node file. File body:
```
# ROOT/f
Name content.
# Public
## Overview
### A level-3 heading
**Bold text**
` `` `go
fmt.Println("hello")
` `` `
```
(The backtick fence is a real three-backtick fence. No blank lines between the
subsection heading and the content lines.)

**Action:** Call `NodeParse("ROOT/f")`.

**Expected:**
- `public.subsections` has one entry with `heading` = `"overview"`
- That subsection's `content` = `["### A level-3 heading", "**Bold text**", "` ``` `go", "fmt.Println(\"hello\")", "` ``` `"]`
  (each line as a raw string, including fence lines)

---

## Heading Normalization

### TC-HN-01: Case insensitive public detection

**Setup:** Create a node file. File body:
```
# ROOT/g
Name content.
# PUBLIC
Public content.
```

**Action:** Call `NodeParse("ROOT/g")`.

**Expected:**
- `public` present, `public.heading` = `"public"`

---

### TC-HN-02: Public with mixed case and extra whitespace

**Setup:** Create a node file. File body:
```
# ROOT/h
Name content.
#   PuBLiC
Public content.
```

**Action:** Call `NodeParse("ROOT/h")`.

**Expected:**
- `public` present, `public.heading` = `"public"`

---

### TC-HN-03: Node name with varied whitespace

**Setup:** Create a node file for `ROOT/e`. File body:
```
#   ROOT/e
Name content.
```

**Action:** Call `NodeParse("ROOT/e")`.

**Expected:**
- `name_section.heading` = `"root/e"`

---

### TC-HN-04: Subsection headings are normalized

**Setup:** Create a node file. File body:
```
# ROOT/i
Name content.
# Public
##   Interface
Interface content.
## CONSTRAINTS
Constraints content.
```

**Action:** Call `NodeParse("ROOT/i")`.

**Expected:**
- `public.subsections` has two entries:
  - subsection 0: `heading` = `"interface"`
  - subsection 1: `heading` = `"constraints"`

---

### TC-HN-05: Closing hashes are stripped

**Setup:** Create a node file. File body:
```
# ROOT/j
Name content.
# Public
## Interface ##
Interface content.
```

**Action:** Call `NodeParse("ROOT/j")`.

**Expected:**
- Subsection `heading` = `"interface"`
- Subsection `raw_heading` = `"## Interface ##"`

---

## Raw Heading Preservation

### TC-RH-01: Raw heading preserves original line

**Setup:** Create a node file. File body:
```
# ROOT/k
Name content.
# Public
## Interface
Interface content.
```

**Action:** Call `NodeParse("ROOT/k")`.

**Expected:**
- `public.raw_heading` = `"# Public"`
- First subsection `raw_heading` = `"## Interface"`

---

### TC-RH-02: Raw heading preserves case

**Setup:** Create a node file. File body:
```
# ROOT/l
Name content.
# PUBLIC
Public content.
```

**Action:** Call `NodeParse("ROOT/l")`.

**Expected:**
- `public.heading` = `"public"` (normalized)
- `public.raw_heading` = `"# PUBLIC"` (original)

---

### TC-RH-03: Raw heading preserves closing hashes

**Setup:** Create a node file. File body:
```
# ROOT/m
Name content.
# Public
## Foo ##
Foo content.
```

**Action:** Call `NodeParse("ROOT/m")`.

**Expected:**
- Subsection `heading` = `"foo"`
- Subsection `raw_heading` = `"## Foo ##"`

---

### TC-RH-04: Raw heading preserves extra whitespace

**Setup:** Create a node file. File body:
```
# ROOT/n
Name content.
#   Public
Public content.
```

**Action:** Call `NodeParse("ROOT/n")`.

**Expected:**
- `public.heading` = `"public"`
- `public.raw_heading` = `"#   Public"`

---

## Content Boundaries

### TC-CB-01: Level-3 and deeper headings are content

**Setup:** Create a node file. File body:
```
# ROOT/o
Name content.
# Public
## Overview
### Sub-sub heading
#### Even deeper
```

**Action:** Call `NodeParse("ROOT/o")`.

**Expected:**
- `public.subsections` has one entry: `heading` = `"overview"`
- That subsection's `content` = `["### Sub-sub heading", "#### Even deeper"]`
  (those lines are content, not structural headings)

---

### TC-CB-02: Fenced code blocks with heading-like content (backtick fence)

**Setup:** Create a node file. File body:
```
# ROOT/p
Name content.
# Public
## Overview
` ``` `
# this looks like a heading
## also looks like a heading
` ``` `
```
(Real three-backtick fences.)

**Action:** Call `NodeParse("ROOT/p")`.

**Expected:**
- `public.subsections` has one entry: `heading` = `"overview"`
- That subsection's `content` includes the fence lines and the lines starting
  with `#` and `##` as raw strings — they are NOT treated as structural headings.

---

### TC-CB-03: Fenced code block with tilde fence

**Setup:** Create a node file. File body:
```
# ROOT/q
Name content.
# Public
## Overview
~~~
# looks like a level-1 heading
~~~
```

**Action:** Call `NodeParse("ROOT/q")`.

**Expected:**
- The line `"# looks like a level-1 heading"` is in the subsection content,
  not treated as a structural heading.

---

### TC-CB-04: Fenced code block with language tag

**Setup:** Create a node file. File body:
```
# ROOT/r
Name content.
# Public
## Overview
` ``` `python
# looks like a level-1 heading
` ``` `
```
(Real three-backtick fences with `python` language tag.)

**Action:** Call `NodeParse("ROOT/r")`.

**Expected:**
- The line `"# looks like a level-1 heading"` is in the subsection content,
  not treated as a structural heading.

---

### TC-CB-05: Blank lines between heading and content are preserved

**Setup:** Create a node file. File body:
```
# ROOT/s
Name content.
# Public

Content line.
```
(One blank line between `# Public` and `Content line.`)

**Action:** Call `NodeParse("ROOT/s")`.

**Expected:**
- `public.content` = `["", "Content line."]`
  (the blank line appears as an empty string at index 0)

---

## Frontmatter Handling

### TC-FM-01: Frontmatter is skipped

**Setup:** Create a node file. File:
```
---
depends_on: []
---
# ROOT/t
Name content.
```

**Action:** Call `NodeParse("ROOT/t")`.

**Expected:**
- No error.
- Frontmatter is ignored; body is parsed correctly.
- `name_section.heading` = `"root/t"`, `name_section.content` = `["Name content."]`

---

### TC-FM-02: No frontmatter delimiters

**Setup:** Create a node file with no `---` delimiters at all. File body:
```
# ROOT/u
Name content.
```

**Action:** Call `NodeParse("ROOT/u")`.

**Expected:**
- No error.
- Body parsed correctly: `name_section.heading` = `"root/u"`,
  `name_section.content` = `["Name content."]`

---

### TC-FM-03: Unclosed frontmatter

**Setup:** Create a node file that starts with `---` but has no closing `---`. File:
```
---
depends_on: []
# ROOT/v
Name content.
```

**Action:** Call `NodeParse("ROOT/v")`.

**Expected:**
- Error: `"unexpected content before first heading"`

---

## Failure Cases

### TC-FC-01: ARTIFACT reference rejected

**Setup:** No file setup needed.

**Action:** Call `NodeParse("ARTIFACT/x(y)")`.

**Expected:**
- Error: `"not a ROOT reference"`

---

### TC-FC-02: Qualifier rejected

**Setup:** No file setup needed.

**Action:** Call `NodeParse("ROOT/x(interface)")`.

**Expected:**
- Error: `"has qualifier"`

---

### TC-FC-03: File does not exist

**Setup:** Use a logical name whose corresponding file does not exist on disk.

**Action:** Call `NodeParse` with that logical name.

**Expected:**
- Error: `"file unreadable"`

---

### TC-FC-04: Propagates path errors

**Setup:** Use a logical name that resolves to a path containing traversal
(e.g., a name that after path resolution produces an invalid or out-of-bounds path).

**Action:** Call `NodeParse` with that logical name.

**Expected:**
- The path error from `FileOpen` is propagated as-is (not wrapped as
  `"file unreadable"`).

---

### TC-FC-05: Content before first heading

**Setup:** Create a node file. File body:
```
Some text before any heading.
# ROOT/w
Name content.
```

**Action:** Call `NodeParse("ROOT/w")`.

**Expected:**
- Error: `"unexpected content before first heading"`

---

### TC-FC-06: Level-2 heading before any level-1 heading

**Setup:** Create a node file. File body:
```
## Subsection before level-1
# ROOT/aa
Name content.
```

**Action:** Call `NodeParse("ROOT/aa")`.

**Expected:**
- Error: `"unexpected content before first heading"`

---

### TC-FC-07: Empty body

**Setup:** Create a node file with no body content (empty file, or only
frontmatter with an empty body after the closing `---`).

**Action:** Call `NodeParse` with the corresponding logical name.

**Expected:**
- Error: `"unexpected content before first heading"`

---

### TC-FC-08: Node name does not match logical name

**Setup:** Create a node file. File body:
```
# ROOT/something-else
Name content.
```

**Action:** Call `NodeParse("ROOT/different")`.

**Expected:**
- Error: `"node name does not match"`

---

### TC-FC-09: Node name case mismatch is not an error

**Setup:** Create a node file. File body:
```
# root/ab
Name content.
```

**Action:** Call `NodeParse("ROOT/ab")`.

**Expected:**
- No error.
- `name_section.heading` = `"root/ab"` (normalization makes both equal).

---

### TC-FC-10: Duplicate public section — same case

**Setup:** Create a node file. File body:
```
# ROOT/ac
Name content.
# Public
Public content 1.
# Public
Public content 2.
```

**Action:** Call `NodeParse("ROOT/ac")`.

**Expected:**
- Error: `"duplicate public section"`

---

### TC-FC-11: Duplicate public section — different case

**Setup:** Create a node file. File body:
```
# ROOT/ad
Name content.
# Public
Public content 1.
# PUBLIC
Public content 2.
```

**Action:** Call `NodeParse("ROOT/ad")`.

**Expected:**
- Error: `"duplicate public section"`

---

### TC-FC-12: Duplicate agent section

**Setup:** Create a node file. File body:
```
# ROOT/ae
Name content.
# Agent
Agent content 1.
# Agent
Agent content 2.
```

**Action:** Call `NodeParse("ROOT/ae")`.

**Expected:**
- Error: `"duplicate agent section"`

---

### TC-FC-13: Duplicate subsection in public — same case

**Setup:** Create a node file. File body:
```
# ROOT/af
Name content.
# Public
## Interface
Interface content 1.
## Interface
Interface content 2.
```

**Action:** Call `NodeParse("ROOT/af")`.

**Expected:**
- Error: `"duplicate subsection"`

---

### TC-FC-14: Duplicate subsection in public — different case

**Setup:** Create a node file. File body:
```
# ROOT/ag
Name content.
# Public
## Interface
Interface content 1.
## INTERFACE
Interface content 2.
```

**Action:** Call `NodeParse("ROOT/ag")`.

**Expected:**
- Error: `"duplicate subsection"`

---

### TC-FC-15: Duplicate subsection in public — whitespace variation

**Setup:** Create a node file. File body:
```
# ROOT/ah
Name content.
# Public
## Interface
Interface content 1.
##   Interface
Interface content 2.
```

**Action:** Call `NodeParse("ROOT/ah")`.

**Expected:**
- Error: `"duplicate subsection"`

---

### TC-FC-16: Duplicate subsection in agent

**Setup:** Create a node file. File body:
```
# ROOT/ai
Name content.
# Agent
## Guidance
Guidance content 1.
## Guidance
Guidance content 2.
```

**Action:** Call `NodeParse("ROOT/ai")`.

**Expected:**
- Error: `"duplicate subsection"`
