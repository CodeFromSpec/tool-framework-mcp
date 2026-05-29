<!-- code-from-spec: ROOT/functional/tests/parsing/node_parsing@u261o14-reEiuBacFqmRGFTeiSI -->

# Test Specification: NodeParse

---

## Happy Path

---

### Test: Minimal node — name section only

**Setup**
Create a node file for `ROOT/x` with the following content:
```
# ROOT/x

A simple node.
```

**Action**
Call `NodeParse` with `"ROOT/x"`.

**Expected outcome**
- Returns a Node record with no error.
- `name_section.heading` = `"root/x"`
- `name_section.raw_heading` = `"# ROOT/x"`
- `name_section.content` = `["A simple node."]`
- `name_section.subsections` is empty
- `public` is absent
- `agent` is absent
- `private` is empty

---

### Test: Full node — name, public, agent, private

**Setup**
Create a node file for `ROOT/payments/fees` with frontmatter, a name section, a public section containing `## Interface` and `## Constraints` subsections, an agent section with content, and two private sections (`# Decisions`, `# Rationale`).

**Action**
Call `NodeParse` with `"ROOT/payments/fees"`.

**Expected outcome**
- Returns a Node record with no error.
- `name_section.heading` = `"root/payments/fees"`
- `public` is present with two subsections having headings `"interface"` and `"constraints"`, in that order
- `agent` is present with content
- `private` has two sections in file order: one with heading `"decisions"`, one with heading `"rationale"`

---

### Test: Node with no public section

**Setup**
Create a node file for `ROOT/decisions` with a name section and a single private section `# Rationale`.

**Action**
Call `NodeParse` with `"ROOT/decisions"`.

**Expected outcome**
- Returns a Node record with no error.
- `public` is absent
- `agent` is absent
- `private` has one section with heading `"rationale"`

---

### Test: Public section with content before first subsection

**Setup**
Create a node file for `ROOT/a` with a public section that has direct content lines before a `## Interface` subsection.

**Action**
Call `NodeParse` with `"ROOT/a"`.

**Expected outcome**
- Returns a Node record with no error.
- `public.content` equals the list of lines that appeared before the `## Interface` heading
- `public.subsections` has one entry with heading `"interface"` and raw_heading `"## Interface"`

---

### Test: Public section with no content or subsections

**Setup**
Create a node file where `# Public` is immediately followed by `# Agent`, with no lines between them.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- `public` is present
- `public.content` is an empty list
- `public.subsections` is an empty list

---

### Test: Agent section with subsections

**Setup**
Create a node file with an agent section containing some preamble lines, then a `## Implementation guidance` subsection with content, then a `## Contracts` subsection with content.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- `agent.content` equals the list of preamble lines
- `agent.raw_heading` = `"# Agent"`
- `agent.subsections` has two entries in order:
  - heading `"implementation guidance"`, with content as list of lines
  - heading `"contracts"`, with content as list of lines

---

### Test: Private sections preserve file order

**Setup**
Create a node file with three private sections in order: `# TODO`, `# Decisions`, `# Rationale`.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- `private` has three sections in file order with headings `"todo"`, `"decisions"`, `"rationale"`

---

### Test: Content is raw markdown

**Setup**
Create a node file with a public subsection whose content includes level-3 headings (`### Details`), bold text (`**bold**`), and fenced code blocks.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- The subsection `content` is a list of raw lines exactly as they appear in the file, including `### Details`, `**bold**`, and the fenced code block delimiters and body

---

## Heading Normalization

---

### Test: Case insensitive public detection

**Setup**
Create a node file where the public section heading is `# PUBLIC`.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- `public` is present
- `public.heading` = `"public"`

---

### Test: Public with mixed case and extra whitespace

**Setup**
Create a node file where the public section heading is `#   PuBLiC`.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- `public` is present
- `public.heading` = `"public"`

---

### Test: Node name with varied whitespace

**Setup**
Create a node file for `ROOT/e` where the name heading is `#    ROOT/e` (extra leading whitespace after `#`).

**Action**
Call `NodeParse` with `"ROOT/e"`.

**Expected outcome**
- Returns a Node record with no error.
- `name_section.heading` = `"root/e"`

---

### Test: Subsection headings are normalized

**Setup**
Create a node file with a public section containing subsections `##   Interface` and `## CONSTRAINTS`.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- The two subsection headings are `"interface"` and `"constraints"`

---

### Test: Closing hashes are stripped

**Setup**
Create a node file with a subsection heading `## Interface ##`.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- The subsection `heading` = `"interface"`
- The subsection `raw_heading` = `"## Interface ##"`

---

## Raw Heading Preservation

---

### Test: Raw heading preserves original line

**Setup**
Create a node file with `# Public` and `## Interface` headings.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- `public.raw_heading` = `"# Public"`
- The subsection `raw_heading` = `"## Interface"`

---

### Test: Raw heading preserves case

**Setup**
Create a node file where the public heading is `# PUBLIC`.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- `public.heading` = `"public"` (normalized)
- `public.raw_heading` = `"# PUBLIC"` (original)

---

### Test: Raw heading preserves closing hashes

**Setup**
Create a node file with a subsection heading `## Foo ##`.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- Subsection `heading` = `"foo"`
- Subsection `raw_heading` = `"## Foo ##"`

---

### Test: Raw heading preserves extra whitespace

**Setup**
Create a node file with the heading `#   Public` (extra whitespace after `#`).

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- `public.heading` = `"public"`
- `public.raw_heading` = `"#   Public"`

---

## Content Boundaries

---

### Test: Level-3 and deeper headings are content

**Setup**
Create a node file with a public subsection that contains `### Details` and `#### Sub-details` headings with text beneath them.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- The `###` and `####` lines and their following text are included as raw content lines within the subsection
- No new subsections are created for them

---

### Test: Fenced code blocks with heading-like content (backtick fence)

**Setup**
Create a node file with a public subsection that contains a fenced code block opened by triple backticks. Inside the code block there are lines starting with `#` and `##`.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- The heading-like lines inside the fenced code block are treated as content, not as structural headings
- No new sections or subsections are created for them

---

### Test: Fenced code block with tilde fence

**Setup**
Create a node file with a public subsection containing a code block opened by `~~~`. Inside the fence there is a line `# Heading`.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- The `# Heading` line inside the tilde fence is treated as content, not as a structural heading

---

### Test: Fenced code block with language tag

**Setup**
Create a node file with a public subsection containing a code block opened by a triple-backtick fence with a language tag (e.g., ` ```yaml `). Inside the fence there is a line `# Heading`.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- The `# Heading` line inside the code block is treated as content, not as a structural heading

---

### Test: Leading and trailing blank lines are preserved

**Setup**
Create a node file where sections and subsections have blank lines at the start and end of their content.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- Leading and trailing blank lines are preserved in all `content` lists exactly as they appear in the file

---

## Frontmatter Handling

---

### Test: Frontmatter is skipped

**Setup**
Create a node file that begins with a frontmatter block (content between two `---` delimiter lines) followed by a body with a name heading and content.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- Frontmatter is skipped entirely
- The body is parsed correctly

---

### Test: No frontmatter delimiters

**Setup**
Create a node file with no `---` delimiters anywhere — body only, starting with the name heading.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns a Node record with no error.
- Body is parsed correctly

---

### Test: Unclosed frontmatter

**Setup**
Create a node file that starts with `---` but has no closing `---` delimiter. The rest of the file contains a name heading and content.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns error `"unexpected content before first heading"`

---

## Failure Cases

---

### Test: ARTIFACT reference rejected

**Setup**
No file setup needed.

**Action**
Call `NodeParse` with `"ARTIFACT/x(y)"`.

**Expected outcome**
- Returns error `"not a ROOT reference"`

---

### Test: Qualifier rejected

**Setup**
No file setup needed.

**Action**
Call `NodeParse` with `"ROOT/x(interface)"`.

**Expected outcome**
- Returns error `"has qualifier"`

---

### Test: File does not exist

**Setup**
No file setup needed. Use a logical name whose resolved file path does not exist on disk.

**Action**
Call `NodeParse` with the logical name.

**Expected outcome**
- Returns error `"file unreadable"`

---

### Test: Propagates path errors

**Setup**
No file setup needed. Use a logical name that resolves to a path containing a traversal component or is otherwise invalid at the path level.

**Action**
Call `NodeParse` with the invalid logical name.

**Expected outcome**
- The path error from `FileOpen` is propagated unchanged

---

### Test: Content before first heading

**Setup**
Create a node file where non-blank text appears before any `#` heading.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns error `"unexpected content before first heading"`

---

### Test: Level-2 heading before any level-1 heading

**Setup**
Create a node file where a `##` heading appears before any `#` heading.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns error `"unexpected content before first heading"`

---

### Test: Empty body

**Setup**
Create a node file with no content, or with only a frontmatter block and no body lines.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns error `"unexpected content before first heading"`

---

### Test: Node name does not match logical name

**Setup**
Create a node file where the first heading is `# ROOT/other`.

**Action**
Call `NodeParse` with `"ROOT/x"`.

**Expected outcome**
- Returns error `"node name does not match"`

---

### Test: Node name case mismatch is not an error

**Setup**
Create a node file with the first heading `# root/x` (lowercase).

**Action**
Call `NodeParse` with `"ROOT/x"`.

**Expected outcome**
- Returns a Node record with no error — normalization makes the names equal

---

### Test: Duplicate public section — same case

**Setup**
Create a node file containing two `# Public` sections.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns error `"duplicate public section"`

---

### Test: Duplicate public section — different case

**Setup**
Create a node file containing `# Public` and `# PUBLIC` sections.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns error `"duplicate public section"`

---

### Test: Duplicate agent section

**Setup**
Create a node file containing two `# Agent` sections.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns error `"duplicate agent section"`

---

### Test: Duplicate subsection in public — same case

**Setup**
Create a node file with a public section containing two `## Interface` subsections.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns error `"duplicate subsection"`

---

### Test: Duplicate subsection in public — different case

**Setup**
Create a node file with a public section containing `## Interface` and `## INTERFACE` subsections.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns error `"duplicate subsection"`

---

### Test: Duplicate subsection in public — whitespace variation

**Setup**
Create a node file with a public section containing `## Interface` and `##   Interface` subsections.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns error `"duplicate subsection"`

---

### Test: Duplicate subsection in agent

**Setup**
Create a node file with an agent section containing two `## Details` headings.

**Action**
Call `NodeParse`.

**Expected outcome**
- Returns error `"duplicate subsection"`
