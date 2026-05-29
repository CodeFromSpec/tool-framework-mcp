<!-- code-from-spec: ROOT/functional/tests/parsing/node_parsing@1vTAlPU-OMpcR78nxzZuWrezNT8 -->

# NodeParse Test Cases

---

## Happy Path

### Minimal node — name section only

**Setup:**
Create a node file for `ROOT/x` with the following content:
```
# ROOT/x

A simple node.
```

**Action:**
Call `NodeParse` with `"ROOT/x"`.

**Expected outcome:**
- `name_section.heading` = `"root/x"`
- `name_section.raw_heading` = `"# ROOT/x"`
- `name_section.content` = `["A simple node."]`
- `name_section.subsections` is empty
- `public` is absent
- `agent` is absent
- `private` is empty

---

### Full node — name, public, agent, private

**Setup:**
Create a node file for `ROOT/payments/fees` with frontmatter, then:
- A name section `# ROOT/payments/fees` with some content
- A public section `# Public` with subsections `## Interface` and `## Constraints`
- An agent section `# Agent` with some content
- Two private sections: `# Decisions` and `# Rationale`, each with content

**Action:**
Call `NodeParse` with `"ROOT/payments/fees"`.

**Expected outcome:**
- `name_section.heading` = `"root/payments/fees"`
- `public` is present with two subsections in order: headings `"interface"` and `"constraints"`
- `agent` is present with content
- `private` has two sections in file order: headings `"decisions"` and `"rationale"`

---

### Node with no public section

**Setup:**
Create a node file for `ROOT/decisions` with:
- A name section `# ROOT/decisions` with some content
- A private section `# Rationale` with some content

**Action:**
Call `NodeParse` with `"ROOT/decisions"`.

**Expected outcome:**
- `public` is absent
- `agent` is absent
- `private` has one section with heading `"rationale"`

---

### Public section with content before first subsection

**Setup:**
Create a node file for `ROOT/a` with:
- A name section `# ROOT/a`
- A public section `# Public` with two or more lines of direct content, followed by `## Interface` with its own content

**Action:**
Call `NodeParse` with `"ROOT/a"`.

**Expected outcome:**
- `public.content` = the list of lines appearing before the `## Interface` subsection (leading and trailing blank lines trimmed)
- `public.subsections` has one entry with heading `"interface"` and `raw_heading` = `"## Interface"`

---

### Public section with no content or subsections

**Setup:**
Create a node file where `# Public` is immediately followed by `# Agent` (no lines between them).

**Action:**
Call `NodeParse`.

**Expected outcome:**
- `public` is present
- `public.content` is an empty list
- `public.subsections` is an empty list

---

### Agent section with ## subsections

**Setup:**
Create a node file with an `# Agent` section containing:
- Some preamble text (multiple lines)
- `## Implementation guidance` subsection with content
- `## Contracts` subsection with content

**Action:**
Call `NodeParse`.

**Expected outcome:**
- `agent.content` = list of preamble lines (trimmed)
- `agent.raw_heading` = `"# Agent"`
- `agent.subsections` has two entries in order:
  - heading `"implementation guidance"`, with its content as list of lines
  - heading `"contracts"`, with its content as list of lines

---

### Private sections preserve file order

**Setup:**
Create a node file with three private sections in this order:
- `# TODO` with content
- `# Decisions` with content
- `# Rationale` with content

**Action:**
Call `NodeParse`.

**Expected outcome:**
- `private` has three sections in order: headings `"todo"`, `"decisions"`, `"rationale"`

---

### Content is raw markdown

**Setup:**
Create a node file with a public subsection whose content contains:
- A level-3 heading: `### Details`
- Bold text: `**important**`
- A fenced code block opened with ` ``` ` and closed with ` ``` `

**Action:**
Call `NodeParse`.

**Expected outcome:**
- The subsection `content` is a list of raw lines, including the `### Details` line, the `**important**` line, and all lines of the fenced code block verbatim

---

## Heading Normalization

### Case insensitive public detection

**Setup:**
Create a node file with `# PUBLIC` as the public heading.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- `public` is present
- `public.heading` = `"public"`

---

### Public with mixed case and extra whitespace

**Setup:**
Create a node file with `#   PuBLiC` as the public heading.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- `public` is present
- `public.heading` = `"public"`

---

### Node name with varied whitespace

**Setup:**
Create a node file with `#    ROOT/e` as the name heading.

**Action:**
Call `NodeParse` with `"ROOT/e"`.

**Expected outcome:**
- `name_section.heading` = `"root/e"`

---

### Subsection headings are normalized

**Setup:**
Create a node file with a public section containing subsections:
- `##   Interface`
- `## CONSTRAINTS`

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Subsection headings = `"interface"` and `"constraints"` respectively

---

### Closing hashes are stripped

**Setup:**
Create a node file with a subsection heading `## Interface ##`.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Subsection `heading` = `"interface"`
- Subsection `raw_heading` = `"## Interface ##"`

---

## Raw Heading Preservation

### Raw heading preserves original line

**Setup:**
Create a node file with `# Public` and `## Interface`.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- `public.raw_heading` = `"# Public"`
- The `## Interface` subsection has `raw_heading` = `"## Interface"`

---

### Raw heading preserves case

**Setup:**
Create a node file with `# PUBLIC` as the public heading.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- `public.heading` = `"public"` (normalized)
- `public.raw_heading` = `"# PUBLIC"` (original)

---

### Raw heading preserves closing hashes

**Setup:**
Create a node file with a subsection heading `## Foo ##`.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Subsection `heading` = `"foo"`
- Subsection `raw_heading` = `"## Foo ##"`

---

### Raw heading preserves extra whitespace

**Setup:**
Create a node file with `#   Public` as the public heading.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- `public.heading` = `"public"`
- `public.raw_heading` = `"#   Public"`

---

## Content Boundaries

### Level-3 and deeper headings are content

**Setup:**
Create a node file with a public subsection containing:
- `### Details` with text
- `#### Sub-details` with text

**Action:**
Call `NodeParse`.

**Expected outcome:**
- The `###` and `####` lines and all their associated text are included as raw lines within the subsection `content` list

---

### Fenced code blocks with heading-like content

**Setup:**
Create a node file with a fenced code block (opened with ` ``` `) inside a public subsection. The code block contains lines starting with `#` and `##`.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- The `#` and `##` lines inside the fenced code block are treated as content, not as structural section or subsection headings

---

### Fenced code block with tilde fence

**Setup:**
Create a node file with a code block opened by `~~~` inside a subsection. The code block contains `# Heading`.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- `# Heading` inside the tilde fence is treated as content, not as a structural heading

---

### Fenced code block with language tag

**Setup:**
Create a node file with a code block opened by ` ```yaml ` inside a subsection. The code block contains `# Heading`.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- `# Heading` inside the code block is treated as content, not as a structural heading

---

### Leading and trailing blank lines are trimmed

**Setup:**
Create a node file with blank lines before and after content in sections and subsections:
- Blank lines before the first content line in a section
- Blank lines after the last content line in a section
- Same pattern inside subsections

**Action:**
Call `NodeParse`.

**Expected outcome:**
- All `content` lists have no empty strings at the start or end (leading and trailing blank lines are trimmed)

---

## Frontmatter Handling

### Frontmatter is skipped

**Setup:**
Create a node file with frontmatter delimited by `---` on the first line and a second `---` line, followed by a body with a valid name heading and content.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- No error
- Frontmatter is ignored; the body is parsed correctly

---

### No frontmatter delimiters

**Setup:**
Create a node file with no `---` lines at all — only a valid body starting with a level-1 name heading.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- No error
- Body is parsed correctly

---

### Unclosed frontmatter

**Setup:**
Create a node file that begins with `---` but has no closing `---` line. The rest of the file contains body content.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Error: `"unexpected content before first heading"`

---

## Failure Cases

### ARTIFACT reference rejected

**Setup:**
None.

**Action:**
Call `NodeParse` with `"ARTIFACT/x(y)"`.

**Expected outcome:**
- Error: `"not a ROOT reference"`

---

### Qualifier rejected

**Setup:**
None.

**Action:**
Call `NodeParse` with `"ROOT/x(interface)"`.

**Expected outcome:**
- Error: `"has qualifier"`

---

### File does not exist

**Setup:**
None. Use a logical name whose resolved path does not correspond to any existing file.

**Action:**
Call `NodeParse` with the logical name.

**Expected outcome:**
- Error: `"file unreadable"`

---

### Propagates path errors

**Setup:**
None. Use a logical name that, after path resolution, produces a path error (e.g., a traversal attempt).

**Action:**
Call `NodeParse` with the invalid logical name.

**Expected outcome:**
- The path error from the resolution step is propagated to the caller

---

### Content before first heading

**Setup:**
Create a node file where non-blank text appears before any `#` heading.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Error: `"unexpected content before first heading"`

---

### Level-2 heading before any level-1 heading

**Setup:**
Create a node file that begins with a `##` heading before any `#` heading.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Error: `"unexpected content before first heading"`

---

### Empty body

**Setup:**
Create a node file with no body content (either completely empty, or containing only frontmatter with no body lines after the closing `---`).

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Error: `"unexpected content before first heading"`

---

### Node name does not match logical name

**Setup:**
Create a node file where the first heading is `# ROOT/other`.

**Action:**
Call `NodeParse` with `"ROOT/x"`.

**Expected outcome:**
- Error: `"node name does not match"`

---

### Node name case mismatch is not an error

**Setup:**
Create a node file with heading `# root/x`.

**Action:**
Call `NodeParse` with `"ROOT/x"`.

**Expected outcome:**
- No error — normalization makes the heading and the logical name equal

---

### Duplicate public section — same case

**Setup:**
Create a node file with two `# Public` sections.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Error: `"duplicate public section"`

---

### Duplicate public section — different case

**Setup:**
Create a node file with one `# Public` section and one `# PUBLIC` section.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Error: `"duplicate public section"`

---

### Duplicate agent section

**Setup:**
Create a node file with two `# Agent` sections.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Error: `"duplicate agent section"`

---

### Duplicate subsection in public — same case

**Setup:**
Create a node file with a public section containing two `## Interface` subsections.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Error: `"duplicate subsection"`

---

### Duplicate subsection in public — different case

**Setup:**
Create a node file with a public section containing `## Interface` and `## INTERFACE`.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Error: `"duplicate subsection"`

---

### Duplicate subsection in public — whitespace variation

**Setup:**
Create a node file with a public section containing `## Interface` and `##   Interface`.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Error: `"duplicate subsection"`

---

### Duplicate subsection in agent

**Setup:**
Create a node file with an agent section containing two `## Details` headings.

**Action:**
Call `NodeParse`.

**Expected outcome:**
- Error: `"duplicate subsection"`
