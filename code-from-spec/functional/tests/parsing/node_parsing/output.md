<!-- code-from-spec: ROOT/functional/tests/parsing/node_parsing@X8NhxiEXQ0Kaqaq970yKnN2rQaI -->

## Test Cases for NodeParse

---

### Happy Path

---

#### TC-HP-01: Minimal node — name section only

Setup:
  Create a node file for logical name `"SPEC/x"`.
  File body:
    ```
    # SPEC/x
    A simple node.
    ```

Action:
  Call `NodeParse("SPEC/x")`.

Expected outcome:
  Returns a Node record where:
  - `name_section.heading` = `"spec/x"`
  - `name_section.raw_heading` = `"# SPEC/x"`
  - `name_section.content` = `["A simple node."]`
  - `name_section.subsections` = empty list
  - `public` = absent
  - `agent` = absent
  - `private` = absent

---

#### TC-HP-02: Full node — all section types

Setup:
  Create a node file for logical name `"SPEC/payments/fees"`.
  File body (with frontmatter):
    ```
    ---
    output: some/path
    ---
    # SPEC/payments/fees
    Fee description line.
    # Public
    ## Interface
    Interface content line.
    ## Constraints
    Constraints content line.
    # Agent
    Agent content line.
    # Private
    ## Decisions
    Decisions content line.
    ## Rationale
    Rationale content line.
    ```

Action:
  Call `NodeParse("SPEC/payments/fees")`.

Expected outcome:
  Returns a Node record where:
  - `name_section.heading` = `"spec/payments/fees"`
  - `name_section.content` = `["Fee description line."]`
  - `public` is present
    - `public.content` = empty list
    - `public.subsections` has two entries in order:
      - entry 1: `heading` = `"interface"`, `content` = `["Interface content line."]`
      - entry 2: `heading` = `"constraints"`, `content` = `["Constraints content line."]`
  - `agent` is present
    - `agent.content` = `["Agent content line."]`
  - `private` is present
    - `private.subsections` has two entries in order:
      - entry 1: `heading` = `"decisions"`, `content` = `["Decisions content line."]`
      - entry 2: `heading` = `"rationale"`, `content` = `["Rationale content line."]`

---

#### TC-HP-03: Node with no public section

Setup:
  Create a node file for logical name `"SPEC/decisions"`.
  File body:
    ```
    # SPEC/decisions
    Content line.
    # Private
    ## Rationale
    Rationale content.
    ```

Action:
  Call `NodeParse("SPEC/decisions")`.

Expected outcome:
  Returns a Node record where:
  - `public` = absent
  - `agent` = absent
  - `private` is present
    - `private.subsections` has one entry:
      - `heading` = `"rationale"`, `content` = `["Rationale content."]`

---

#### TC-HP-04: Public section with content before first subsection

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    Preamble line one.
    Preamble line two.
    ## Interface
    Interface content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  Returns a Node record where:
  - `public.content` = `["Preamble line one.", "Preamble line two."]`
  - `public.subsections` has one entry:
    - `heading` = `"interface"`, `content` = `["Interface content."]`

---

#### TC-HP-05: Public section with no content or subsections

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    # Agent
    Agent content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  Returns a Node record where:
  - `public` is present
    - `public.content` = empty list
    - `public.subsections` = empty list
  - `agent` is present with `agent.content` = `["Agent content."]`

---

#### TC-HP-06: Agent section with subsections

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Agent
    Agent preamble.
    ## Implementation guidance
    Implementation content.
    ## Contracts
    Contracts content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  Returns a Node record where:
  - `agent.content` = `["Agent preamble."]`
  - `agent.raw_heading` = `"# Agent"`
  - `agent.subsections` has two entries in order:
    - entry 1: `heading` = `"implementation guidance"`, `content` = `["Implementation content."]`
    - entry 2: `heading` = `"contracts"`, `content` = `["Contracts content."]`

---

#### TC-HP-07: Private section with subsections

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Private
    ## TODO
    Todo content.
    ## Decisions
    Decisions content.
    ## Rationale
    Rationale content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  Returns a Node record where:
  - `private` is present
    - `private.subsections` has three entries in order:
      - entry 1: `heading` = `"todo"`
      - entry 2: `heading` = `"decisions"`
      - entry 3: `heading` = `"rationale"`

---

#### TC-HP-08: Content is raw markdown

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    ## Summary
    ### A level-3 heading
    **Bold text here**
    ```python
    x = 1
    ```
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  Returns a Node record where:
  - `public.subsections` has one entry with `heading` = `"summary"`
  - That subsection's `content` = `["### A level-3 heading", "**Bold text here**", "```python", "x = 1", "```"]`
    (all lines as raw strings, in order)

---

### Heading Normalization

---

#### TC-HN-01: Case insensitive public detection

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # PUBLIC
    Public content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - `public` is present
  - `public.heading` = `"public"`

---

#### TC-HN-02: Public with mixed case and extra whitespace

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    #   PuBLiC
    Public content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - `public` is present
  - `public.heading` = `"public"`

---

#### TC-HN-03: Node name with varied whitespace

Setup:
  Create a node file for logical name `"SPEC/e"`.
  File body:
    ```
    #   SPEC/e
    Name content.
    ```

Action:
  Call `NodeParse("SPEC/e")`.

Expected outcome:
  - No error
  - `name_section.heading` = `"spec/e"`

---

#### TC-HN-04: Node name with ROOT/ heading does not match SPEC/

Setup:
  Create a node file with heading `# ROOT/x`.
  File body:
    ```
    # ROOT/x
    Content.
    ```

Action:
  Call `NodeParse("SPEC/x")`.

Expected outcome:
  - Error `NodeNameDoesNotMatch` — normalized heading `"root/x"` does not match `"spec/x"`

---

#### TC-HN-05: Subsection headings are normalized

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    ##   Interface
    Interface content.
    ## CONSTRAINTS
    Constraints content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - `public.subsections` has two entries:
    - entry 1: `heading` = `"interface"`
    - entry 2: `heading` = `"constraints"`

---

#### TC-HN-06: Closing hashes are stripped

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    ## Interface ##
    Interface content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Subsection `heading` = `"interface"`
  - Subsection `raw_heading` = `"## Interface ##"`

---

### Raw Heading Preservation

---

#### TC-RH-01: Raw heading preserves original line

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    ## Interface
    Interface content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - `public.raw_heading` = `"# Public"`
  - The Interface subsection's `raw_heading` = `"## Interface"`

---

#### TC-RH-02: Raw heading preserves case

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # PUBLIC
    Public content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - `public.heading` = `"public"` (normalized)
  - `public.raw_heading` = `"# PUBLIC"` (original)

---

#### TC-RH-03: Raw heading preserves closing hashes

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    ## Foo ##
    Foo content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Subsection `heading` = `"foo"`
  - Subsection `raw_heading` = `"## Foo ##"`

---

#### TC-RH-04: Raw heading preserves extra whitespace

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    #   Public
    Public content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - `public.heading` = `"public"`
  - `public.raw_heading` = `"#   Public"`

---

### Content Boundaries

---

#### TC-CB-01: Level-3 and deeper headings are content

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    ## Summary
    ### A deeper heading
    #### Even deeper
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - `public.subsections` has one entry with `heading` = `"summary"`
  - That subsection's `content` = `["### A deeper heading", "#### Even deeper"]`

---

#### TC-CB-02: Fenced code blocks with heading-like content (backtick)

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    ## Interface
    ` `` `
    # looks like heading
    ## also looks like heading
    ` `` `
    ```
  (where the code fence is three backticks with no spaces)

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - `public.subsections` has one entry with `heading` = `"interface"`
  - That subsection's `content` includes the lines `"# looks like heading"` and `"## also looks like heading"` as raw content (not parsed as structural headings)

---

#### TC-CB-03: Fenced code block with tilde fence

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    ## Interface
    ~~~
    # This looks like a heading
    ~~~
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - `public.subsections` has one entry with `heading` = `"interface"`
  - That subsection's `content` includes `"# This looks like a heading"` as a raw content line, not treated as a structural heading

---

#### TC-CB-04: Fenced code block with language tag

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    ## Interface
    ```python
    # This looks like a heading
    ` `` `
    ```
  (where the closing fence is three backticks with no spaces)

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - `public.subsections` has one entry with `heading` = `"interface"`
  - That subsection's `content` includes `"# This looks like a heading"` as a raw content line, not treated as a structural heading

---

#### TC-CB-05: Blank lines between heading and content are preserved

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public

    Public content line.
    ```
  (there is exactly one blank line between `# Public` and `Public content line.`)

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - `public.content` = `["", "Public content line."]`
    (the blank line is the empty string `""`, followed by the content line)

---

### Frontmatter Handling

---

#### TC-FM-01: Frontmatter is skipped

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File content:
    ```
    ---
    output: some/path
    depends_on: []
    ---
    # SPEC/a
    Name content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - No error
  - `name_section.heading` = `"spec/a"`
  - `name_section.content` = `["Name content."]`

---

#### TC-FM-02: No frontmatter delimiters

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File content (no `---` delimiters):
    ```
    # SPEC/a
    Name content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - No error
  - Body parsed correctly; `name_section.heading` = `"spec/a"`

---

#### TC-FM-03: Unclosed frontmatter

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File content:
    ```
    ---
    output: some/path
    # SPEC/a
    Name content.
    ```
  (opening `---` present, no closing `---` before body)

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `UnexpectedContentBeforeFirstHeading`

---

### Failure Cases

---

#### TC-FC-01: ARTIFACT reference rejected

Action:
  Call `NodeParse("ARTIFACT/x")`.

Expected outcome:
  - Error `NotASpecReference`

---

#### TC-FC-02: EXTERNAL reference rejected

Action:
  Call `NodeParse("EXTERNAL/x")`.

Expected outcome:
  - Error `NotASpecReference`

---

#### TC-FC-03: Qualifier rejected

Action:
  Call `NodeParse("SPEC/x(interface)")`.

Expected outcome:
  - Error `HasQualifier`

---

#### TC-FC-04: File does not exist

Setup:
  Use a logical name whose corresponding file does not exist on disk.

Action:
  Call `NodeParse` with that logical name.

Expected outcome:
  - Error `FileUnreadable`

---

#### TC-FC-05: Propagates path errors

Setup:
  Use a logical name that resolves to a path with traversal or other path-level error.

Action:
  Call `NodeParse` with that logical name.

Expected outcome:
  - The path error from the path resolution layer is propagated (a `FileReader.*` error)

---

#### TC-FC-06: Content before first heading

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    This is non-blank content.
    # SPEC/a
    Name content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `UnexpectedContentBeforeFirstHeading`

---

#### TC-FC-07: Level-2 heading before any level-1 heading

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    ## A subsection
    # SPEC/a
    Name content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `UnexpectedContentBeforeFirstHeading`

---

#### TC-FC-08: Empty body

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File content is either empty or contains only frontmatter with no body after the closing `---`.

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `UnexpectedContentBeforeFirstHeading`

---

#### TC-FC-09: Node name does not match logical name

Setup:
  Create a node file with heading `# SPEC/other`.
  File body:
    ```
    # SPEC/other
    Name content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `NodeNameDoesNotMatch`

---

#### TC-FC-10: Node name case mismatch is not an error

Setup:
  Create a node file for logical name `"SPEC/foo"`.
  File body:
    ```
    # spec/foo
    Name content.
    ```
  (heading text is lowercase, logical name passed in is uppercase)

Action:
  Call `NodeParse("SPEC/FOO")`.

Expected outcome:
  - No error — normalization makes both `"spec/foo"`, they match

---

#### TC-FC-11: Duplicate public section — same case

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    First public content.
    # Public
    Second public content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `DuplicatePublicSection`

---

#### TC-FC-12: Duplicate public section — different case

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    First public content.
    # PUBLIC
    Second public content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `DuplicatePublicSection`

---

#### TC-FC-13: Duplicate agent section

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Agent
    First agent content.
    # Agent
    Second agent content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `DuplicateAgentSection`

---

#### TC-FC-14: Duplicate private section

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Private
    First private content.
    # Private
    Second private content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `DuplicatePrivateSection`

---

#### TC-FC-15: Unrecognized section heading

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Decisions
    Some content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `UnrecognizedSection`

---

#### TC-FC-16: Unrecognized section — Rationale

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Rationale
    Rationale content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `UnrecognizedSection`

---

#### TC-FC-17: Unrecognized section — TODO

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # TODO
    Todo content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `UnrecognizedSection`

---

#### TC-FC-18: Duplicate subsection in public — same case

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    ## Interface
    First interface content.
    ## Interface
    Second interface content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `DuplicateSubsection`

---

#### TC-FC-19: Duplicate subsection in public — different case

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    ## Interface
    First interface content.
    ## INTERFACE
    Second interface content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `DuplicateSubsection`

---

#### TC-FC-20: Duplicate subsection in public — whitespace variation

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Public
    ## Interface
    First interface content.
    ##   Interface
    Second interface content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `DuplicateSubsection`

---

#### TC-FC-21: Duplicate subsection in agent

Setup:
  Create a node file for logical name `"SPEC/a"`.
  File body:
    ```
    # SPEC/a
    Name content.
    # Agent
    ## Guidance
    First guidance content.
    ## Guidance
    Second guidance content.
    ```

Action:
  Call `NodeParse("SPEC/a")`.

Expected outcome:
  - Error `DuplicateSubsection`
