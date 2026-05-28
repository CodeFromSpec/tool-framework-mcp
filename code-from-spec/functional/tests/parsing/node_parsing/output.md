<!-- code-from-spec: ROOT/functional/tests/parsing/node_parsing@mspCLGQQDJrB-Vsmn79l0t9sbkg -->

# Node Parsing Tests

---

## Happy Path

---

### Minimal node — name section only

**Setup**
Create a node file for `ROOT/x` with:
- First heading: `# ROOT/x`
- Content: `"A simple node."`

**Action**
Call `NodeParse("ROOT/x")`.

**Expected outcome**
- `name_section.heading` = `"root/x"`
- `name_section.content` = `"A simple node."`
- `name_section.subsections` is empty
- `public` is absent
- `agent` is absent
- `private` is empty

---

### Full node — name, public, agent, private

**Setup**
Create a node file for `ROOT/payments/fees` with:
- Frontmatter block (between `---` delimiters)
- Name heading: `# ROOT/payments/fees`, with some content
- `# Public` section containing:
  - `## Interface` subsection with content
  - `## Constraints` subsection with content
- `# Agent` section with content
- `# Decisions` section with content
- `# Rationale` section with content

**Action**
Call `NodeParse("ROOT/payments/fees")`.

**Expected outcome**
- `name_section.heading` = `"root/payments/fees"`
- `public` is present, with two subsections:
  - `subsections[0].heading` = `"interface"`
  - `subsections[1].heading` = `"constraints"`
- `agent` is present with content
- `private` has two sections in file order:
  - `private[0].heading` = `"decisions"`
  - `private[1].heading` = `"rationale"`

---

### Node with no public section

**Setup**
Create a node file for `ROOT/decisions` with:
- Name heading: `# ROOT/decisions`, with content
- `# Rationale` section with content

**Action**
Call `NodeParse("ROOT/decisions")`.

**Expected outcome**
- `public` is absent
- `agent` is absent
- `private` has one section with `heading` = `"rationale"`

---

### Public section with content before first subsection

**Setup**
Create a node file for `ROOT/a` with:
- Name heading: `# ROOT/a`
- `# Public` section containing:
  - Direct content text (e.g., `"Some introductory text."`)
  - `## Interface` subsection with content

**Action**
Call `NodeParse("ROOT/a")`.

**Expected outcome**
- `public.content` = `"Some introductory text."`
- `public.subsections` has one entry with `heading` = `"interface"`

---

### Public section with no content or subsections

**Setup**
Create a node file with:
- Name heading
- `# Public` heading immediately followed by `# Agent` heading (no content between them)
- `# Agent` section with content

**Action**
Call `NodeParse`.

**Expected outcome**
- `public` is present
- `public.content` is empty
- `public.subsections` is empty

---

### Agent section with ## subsections

**Setup**
Create a node file with:
- Name heading
- `# Agent` section containing:
  - Preamble text (e.g., `"Agent preamble."`)
  - `## Implementation guidance` subsection with content
  - `## Contracts` subsection with content

**Action**
Call `NodeParse`.

**Expected outcome**
- `agent.content` = `"Agent preamble."`
- `agent.subsections` has two entries:
  - `agent.subsections[0].heading` = `"implementation guidance"`, with its content
  - `agent.subsections[1].heading` = `"contracts"`, with its content

---

### Private sections preserve file order

**Setup**
Create a node file with:
- Name heading
- `# TODO` section with content
- `# Decisions` section with content
- `# Rationale` section with content

**Action**
Call `NodeParse`.

**Expected outcome**
- `private` has three sections in file order:
  - `private[0].heading` = `"todo"`
  - `private[1].heading` = `"decisions"`
  - `private[2].heading` = `"rationale"`

---

### Content is raw markdown

**Setup**
Create a node file with a public subsection whose content includes:
- A level-3 heading: `### Details`
- Bold text: `**important**`
- A fenced code block (using backtick fence)

**Action**
Call `NodeParse`.

**Expected outcome**
- The subsection content is the raw markdown text, unchanged
- `###` heading lines, `**bold**`, and fenced code blocks are present as-is in the content string

---

## Heading Normalization

---

### Case insensitive public detection

**Setup**
Create a node file with `# PUBLIC` as the public heading.

**Action**
Call `NodeParse`.

**Expected outcome**
- `public` is present
- `public.heading` = `"public"`

---

### Public with mixed case and extra whitespace

**Setup**
Create a node file with `#   PuBLiC` as the public heading.

**Action**
Call `NodeParse`.

**Expected outcome**
- `public` is present
- `public.heading` = `"public"`

---

### Node name with varied whitespace

**Setup**
Create a node file for `ROOT/e` with name heading `#    ROOT/e`.

**Action**
Call `NodeParse("ROOT/e")`.

**Expected outcome**
- `name_section.heading` = `"root/e"`
- No error

---

### Subsection headings are normalized

**Setup**
Create a node file with a public section containing:
- `##   Interface`
- `## CONSTRAINTS`

**Action**
Call `NodeParse`.

**Expected outcome**
- `public.subsections[0].heading` = `"interface"`
- `public.subsections[1].heading` = `"constraints"`

---

### Closing hashes are stripped

**Setup**
Create a node file with a subsection heading `## Interface ##`.

**Action**
Call `NodeParse`.

**Expected outcome**
- The subsection heading = `"interface"`

---

## Content Boundaries

---

### Level-3 and deeper headings are content

**Setup**
Create a node file with a public subsection that contains:
- `### Details` heading with text
- `#### Sub-details` heading with text

**Action**
Call `NodeParse`.

**Expected outcome**
- The `###` and `####` lines and their text are included as raw content within the subsection
- They do not create new sections or subsections in the parsed result

---

### Fenced code blocks with heading-like content (backtick fence)

**Setup**
Create a node file with a fenced code block (using triple backticks) inside a public subsection.
The code block body contains lines starting with `#` and `##`.

**Action**
Call `NodeParse`.

**Expected outcome**
- The `#` and `##` lines inside the code block are treated as content, not as structural headings
- The section structure is unaffected by the heading-like lines inside the code block

---

### Fenced code block with tilde fence

**Setup**
Create a node file with a code block opened by `~~~` inside a subsection.
The code block body contains `# Heading`.

**Action**
Call `NodeParse`.

**Expected outcome**
- The `# Heading` line inside the tilde fence is treated as content, not a structural heading

---

### Fenced code block with language tag

**Setup**
Create a node file with a code block opened by ` ```yaml ` (backtick fence with language tag) inside a subsection.
The code block body contains `# Heading`.

**Action**
Call `NodeParse`.

**Expected outcome**
- The `# Heading` line inside the code block is treated as content, not a structural heading

---

### Leading and trailing blank lines are trimmed

**Setup**
Create a node file where content fields in sections and subsections are surrounded by blank lines.

**Action**
Call `NodeParse`.

**Expected outcome**
- All `content` fields have their leading and trailing blank lines removed
- Internal blank lines are preserved as-is

---

## Frontmatter Handling

---

### Frontmatter is skipped

**Setup**
Create a node file with a frontmatter block between `---` delimiters, followed by a body with a name heading and content.

**Action**
Call `NodeParse`.

**Expected outcome**
- Frontmatter is skipped
- Body is parsed correctly with no error

---

### No frontmatter delimiters

**Setup**
Create a node file with no `---` lines at all — body only, starting with a name heading.

**Action**
Call `NodeParse`.

**Expected outcome**
- No error
- Body is parsed correctly

---

### Unclosed frontmatter

**Setup**
Create a node file that starts with `---` but has no closing `---` — the rest of the file is body content.

**Action**
Call `NodeParse`.

**Expected outcome**
- Error: `"unexpected content before first heading"`

---

## Failure Cases

---

### ARTIFACT reference rejected

**Setup**
No file setup needed.

**Action**
Call `NodeParse("ARTIFACT/x(y)")`.

**Expected outcome**
- Error: `"not a ROOT reference"`

---

### Qualifier rejected

**Setup**
No file setup needed.

**Action**
Call `NodeParse("ROOT/x(interface)")`.

**Expected outcome**
- Error: `"has qualifier"`

---

### File does not exist

**Setup**
No file setup needed. Use a logical name whose resolved file path does not exist.

**Action**
Call `NodeParse` with the non-existent logical name.

**Expected outcome**
- Error: `"file unreadable"`

---

### Propagates path errors

**Setup**
No file setup needed.

**Action**
Call `NodeParse` with a logical name that resolves to an invalid file path (e.g., a path containing traversal components that the path resolver rejects).

**Expected outcome**
- The path error from `FileOpen` is propagated as-is

---

### Content before first heading

**Setup**
Create a node file where non-blank text appears before any `#` heading.

**Action**
Call `NodeParse`.

**Expected outcome**
- Error: `"unexpected content before first heading"`

---

### Level-2 heading before any level-1 heading

**Setup**
Create a node file that begins with a `##` heading before any `#` heading.

**Action**
Call `NodeParse`.

**Expected outcome**
- Error: `"unexpected content before first heading"`

---

### Empty body

**Setup**
Create a node file with no body content (either empty file, or only a frontmatter block with no body).

**Action**
Call `NodeParse`.

**Expected outcome**
- Error: `"unexpected content before first heading"`

---

### Node name does not match logical name

**Setup**
Create a node file where the first heading is `# ROOT/other`.

**Action**
Call `NodeParse("ROOT/x")`.

**Expected outcome**
- Error: `"node name does not match"`

---

### Node name case mismatch is not an error

**Setup**
Create a node file with heading `# root/x` (all lowercase).

**Action**
Call `NodeParse("ROOT/x")`.

**Expected outcome**
- No error
- Normalization causes `name_section.heading` = `"root/x"`, matching the normalized logical name

---

### Duplicate public section — same case

**Setup**
Create a node file with two `# Public` sections.

**Action**
Call `NodeParse`.

**Expected outcome**
- Error: `"duplicate public section"`

---

### Duplicate public section — different case

**Setup**
Create a node file with one `# Public` section and one `# PUBLIC` section.

**Action**
Call `NodeParse`.

**Expected outcome**
- Error: `"duplicate public section"`

---

### Duplicate agent section

**Setup**
Create a node file with two `# Agent` sections.

**Action**
Call `NodeParse`.

**Expected outcome**
- Error: `"duplicate agent section"`

---

### Duplicate subsection in public — same case

**Setup**
Create a node file with a `# Public` section containing two `## Interface` subsections.

**Action**
Call `NodeParse`.

**Expected outcome**
- Error: `"duplicate subsection"`

---

### Duplicate subsection in public — different case

**Setup**
Create a node file with a `# Public` section containing one `## Interface` subsection and one `## INTERFACE` subsection.

**Action**
Call `NodeParse`.

**Expected outcome**
- Error: `"duplicate subsection"`

---

### Duplicate subsection in public — whitespace variation

**Setup**
Create a node file with a `# Public` section containing one `## Interface` subsection and one `##   Interface` subsection.

**Action**
Call `NodeParse`.

**Expected outcome**
- Error: `"duplicate subsection"`

---

### Duplicate subsection in agent

**Setup**
Create a node file with a `# Agent` section containing two `## Details` subsections.

**Action**
Call `NodeParse`.

**Expected outcome**
- Error: `"duplicate subsection"`
