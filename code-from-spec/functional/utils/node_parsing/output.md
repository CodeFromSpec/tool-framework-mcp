<!-- code-from-spec: ROOT/functional/utils/node_parsing@qQHziiQt_5xfFrDtzuSYaHPUkeE -->

# node_parsing

## Data Structures

```
record Subsection
  heading: string       -- normalized heading text
  content: string       -- trimmed body content under this ## heading

record Section
  heading: string       -- normalized heading text
  content: string       -- trimmed body content directly under this # heading
  subsections: list of Subsection

record ParsedNode
  name_section: Section         -- the first # heading section (the node name)
  public:       optional Section -- the # Public section, if present
  agent:        optional Section -- the # Agent section, if present
  private:      list of Section  -- all other # sections (excluding name, public, agent)
```

## Functions

---

### ParseNode(logical_name) -> ParsedNode

**Parameters**
- `logical_name` — a ROOT/ logical name identifying the spec node to parse

**Returns**
- a `ParsedNode` record describing the structured content of the node file

**Errors**
- `"unexpected content before first heading"` — the file body contains non-blank text before the first level-1 heading
- `"node name does not match"` — the first level-1 heading does not match the logical name after normalization
- `"duplicate public section"` — more than one `# Public` section is present in the file
- `"duplicate subsection"` — two `##` headings within the same `# Public` section normalize to the same text

---

**Steps**

1. Resolve the file path from `logical_name` using `ResolvePath`.
   If resolution raises an error, propagate it.

2. Open the file using `OpenFileReader(file_path)`.
   If the file cannot be opened, raise error `"file unreadable"`.

3. Skip the frontmatter block:
   a. Read lines until a line containing exactly `---` is found.
      That first `---` is the start of the frontmatter.
   b. Continue reading lines until a second line containing exactly `---` is found.
      That second `---` closes the frontmatter.
   c. If either `---` delimiter is not found before end of file,
      treat the entire file as having no frontmatter (reset to beginning of body).

4. Parse the remaining body lines into raw sections.
   Maintain state:
   - `current_section`: the active level-1 section being accumulated (or none)
   - `current_subsection`: the active level-2 subsection being accumulated (or none)
   - `in_fenced_block`: boolean, initially false
   - `pre_heading_lines`: lines collected before the first level-1 heading

   For each line read until end of file:

   a. If the line starts with ` ``` ` (three backticks), toggle `in_fenced_block`.
      Append the line to the current accumulator and continue.

   b. If `in_fenced_block` is true, append the line to the current accumulator
      and continue. (Headings inside fenced blocks are not structural.)

   c. If the line starts with exactly `## ` (two hashes and a space),
      and `current_section` is set:
      - If `current_subsection` is set, finalize it:
        trim leading and trailing blank lines from its accumulated content,
        append it to `current_section.subsections`.
      - Start a new `current_subsection` with:
        - `heading` = NormalizeName(text after `## `)
        - `content` = empty accumulator
      Continue.

   d. If the line starts with exactly `# ` (one hash and a space):
      - If `current_subsection` is set, finalize it (trim, append to section).
        Clear `current_subsection`.
      - If `current_section` is set, finalize it:
        trim leading and trailing blank lines from its accumulated content,
        append it to the sections list.
      - Start a new `current_section` with:
        - `heading` = NormalizeName(text after `# `)
        - `content` = empty accumulator
        - `subsections` = empty list
      Continue.

   e. Otherwise (a content line):
      - If `current_subsection` is set, append the line to `current_subsection.content`.
      - Else if `current_section` is set, append the line to `current_section.content`.
      - Else, append the line to `pre_heading_lines`.

5. After all lines are read:
   - If `current_subsection` is set, finalize it (trim, append to section).
   - If `current_section` is set, finalize it (trim, append to sections list).

6. Close the reader with `Close`.

7. Validate pre-heading content:
   If `pre_heading_lines` contains any non-blank line,
   raise error `"unexpected content before first heading"`.

8. Validate that at least one section was found.
   If the sections list is empty, raise error `"unexpected content before first heading"`.
   (A file with no headings has all content before the first heading.)

9. Extract the name section:
   - Take the first entry from the sections list as `name_section`.
   - Derive the expected name: apply `NormalizeName` to the last path segment
     of `logical_name` (the part after the last `/`, with any parenthetical qualifier stripped).
   - Normalize `name_section.heading` using `NormalizeName`.
   - If the two normalized values are not equal,
     raise error `"node name does not match"`.

10. Assign remaining sections:
    Initialize `public` = none, `agent` = none, `private` = empty list.

    For each section after the first (i.e., all sections except `name_section`):

    a. Normalize the section heading using `NormalizeName`.

    b. If the normalized heading equals `"public"`:
       - If `public` is already set, raise error `"duplicate public section"`.
       - Check for duplicate subsections within this section:
         For each pair of subsections, compare their normalized headings.
         If any two are equal, raise error `"duplicate subsection"`.
       - Set `public` = this section.

    c. Else if the normalized heading equals `"agent"`:
       Set `agent` = this section.

    d. Else:
       Append this section to `private`.

11. Return a `ParsedNode` record with:
    - `name_section` = the name section extracted in step 9
    - `public`       = the public section (or none)
    - `agent`        = the agent section (or none)
    - `private`      = the list of all other sections
