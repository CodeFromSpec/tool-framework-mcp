<!-- code-from-spec: ROOT/functional/logic/parsing/node_parsing@7bmo0NQXq41_DE3MVZqwzvfF13E -->

## Namespace

    namespace: parsenode

## Records

```
record NodeSubsection
  heading: string
  raw_heading: string
  content: list of string

record NodeSection
  heading: string
  raw_heading: string
  content: list of string
  subsections: list of NodeSubsection

record Node
  name_section: NodeSection
  public: optional NodeSection
  agent: optional NodeSection
  private: list of NodeSection
```

`heading` is the normalized form (after `NormalizeText`), used for comparisons
and lookups. `raw_heading` is the original line as read from the file, unchanged.
Content fields are lists of strings — each element is a line as returned by
`FileReadLine`, preserving original text exactly.

## Functions

```
function NodeParse(logical_name: string) -> Node
  errors:
    - NotARootReference: the logical name does not start with ROOT/.
    - HasQualifier: the logical name contains a parenthetical qualifier.
    - FileUnreadable: the file cannot be opened or read.
    - UnexpectedContentBeforeFirstHeading: file body has non-blank content
      before the first level-1 heading, or has no level-1 heading at all.
    - NodeNameDoesNotMatch: the first heading does not match the logical name
      after normalization.
    - DuplicatePublicSection: more than one Public section exists.
    - DuplicateAgentSection: more than one Agent section exists.
    - DuplicateSubsection: two level-2 headings within the same section
      normalize to the same text.
    - (FileReader.*): propagated from FileOpen.
```

### NodeParse

  1. If `LogicalNameIsArtifact(logical_name)` returns true,
     raise error `NotARootReference`.

  2. If `LogicalNameHasQualifier(logical_name)` returns true,
     raise error `HasQualifier`.

  3. Call `LogicalNameToPath(logical_name)` to get the file path.

  4. Call `FileOpen` with the file path.
     If `FileOpen` raises an error, propagate it.

  5. Skip the frontmatter:
     Read the first line using `FileReadLine`.
     If `EndOfFile` is raised, call `FileClose` and raise
     `UnexpectedContentBeforeFirstHeading`.
     If the first line is exactly `"---"`:
       Read lines and discard them until a line that is exactly `"---"` is found.
       If `EndOfFile` is reached before the closing `"---"`,
       call `FileClose` and raise `UnexpectedContentBeforeFirstHeading`.
     Else:
       The first line is the first body line — hold it for step 6 processing
       without consuming another line.

  6. Parse the file body into sections. Maintain:
     - a current section (starts as absent)
     - a current subsection (starts as absent)
     - a fenced code block state (open: boolean, fence character, fence length)
     - a result Node being built

     For each line (beginning with any held line from step 5, then continuing
     with `FileReadLine` until `EndOfFile`):

     a. **Fenced code block detection** (applied before heading detection):
        If not currently inside a fenced code block:
          If the line consists of 3 or more consecutive backtick characters
          optionally followed by a language tag (no other characters before
          the backticks), or 3 or more consecutive tilde characters optionally
          followed by a language tag, mark the fenced block as open, recording
          the fence character (`` ` `` or `~`) and fence length.
          Treat the line as content (not a heading).
        If currently inside a fenced code block:
          If the line consists of at least as many of the same fence character
          as the opening line (with nothing before them), mark the fenced block
          as closed.
          Treat the line as content (not a heading) regardless.

     b. **Heading detection** (only if not inside a fenced code block):
        A heading line starts with one or more `"#"` characters immediately
        followed by at least one space. The heading level equals the number of
        leading `"#"` characters. The heading text is everything after the
        `"# "` prefix, trimmed of leading and trailing whitespace.
        If the heading text ends with one or more `"#"` characters preceded by
        at least one space, strip that trailing sequence and trim again.
        Normalize the heading text using `NormalizeText`.
        Lines like `"#Foo"` (no space after `#`) are not headings — treat as content.

     c. **Level-1 heading** (not inside fenced code block, heading level is 1):
        Close the current subsection into the current section (if any).
        Close the current section into the Node (if any) — see "Section closing".
        Start a new section with:
          `raw_heading`: the original line
          `heading`: normalized heading text
          `content`: empty list
          `subsections`: empty list
        Classify the section — see "Section classification".
        Set current section to this new section, current subsection to absent.

     d. **Level-2 heading** (not inside fenced code block, heading level is 2):
        If current section is absent, treat the line as pre-heading content
        (handle as in step e).
        Else:
          Close the current subsection into the current section (if any).
          Check if any existing subsection in the current section has the same
          normalized heading. If so, call `FileClose` and raise
          `DuplicateSubsection`.
          Start a new subsection with:
            `raw_heading`: the original line
            `heading`: normalized heading text
            `content`: empty list
          Set current subsection to this new subsection.

     e. **Content line** (not a structural heading, or level 3+, or inside
        fenced code block, or heading with no space after `#`):
        If current section is absent:
          If the line is not blank, call `FileClose` and raise
          `UnexpectedContentBeforeFirstHeading`.
          Otherwise, discard the blank line.
        Else if current subsection is present:
          Append the line to the current subsection's `content`.
        Else:
          Append the line to the current section's `content`.

  7. After `EndOfFile`:
     Close the current subsection into the current section (if any).
     Close the current section into the Node (if any).
     Call `FileClose`.

  8. If no level-1 heading was ever found (name_section is absent),
     raise `UnexpectedContentBeforeFirstHeading`.

  9. Return the completed Node.

### Section classification

When a new level-1 section is started:
  If this is the first section (name_section is absent):
    Normalize the logical name using `NormalizeText`.
    If the section's normalized heading does not match the normalized logical name,
    call `FileClose` and raise `NodeNameDoesNotMatch`.
    Set `name_section` to this section.
  Else if the section's normalized heading equals `"public"`:
    If `public` is already set, call `FileClose` and raise `DuplicatePublicSection`.
    Set `public` to this section.
  Else if the section's normalized heading equals `"agent"`:
    If `agent` is already set, call `FileClose` and raise `DuplicateAgentSection`.
    Set `agent` to this section.
  Else:
    Append this section to `private`.

### Section closing

When closing a current subsection into the current section:
  Append the subsection to the section's `subsections` list.
  Set current subsection to absent.

When closing a current section into the Node:
  The section has already been classified and assigned to the Node field.
  Update the Node field in place with accumulated `content` and `subsections`.
  Set current section to absent.
