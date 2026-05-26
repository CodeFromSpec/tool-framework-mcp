<!-- code-from-spec: ROOT/functional/utils/node_parsing@R71DvYscAVs0Lp2XpuQu4Qh5QT0 -->

# node_parsing

## Records

```
record Subsection
  heading: string        -- original heading text (not normalized)
  content: string        -- trimmed content lines between this ## and the next heading

record Section
  heading: string        -- original heading text (not normalized)
  content: string        -- trimmed content lines between this # and the first ## (or next # / EOF)
  subsections: list of Subsection

record ParsedNode
  name_section: Section          -- the first # heading (the node name)
  public:       optional Section -- the # Public section, if present
  agent:        optional Section -- the # Agent section, if present
  private:      list of Section  -- all other # sections (order preserved)
```

---

## function ParseNode(logical_name) -> ParsedNode

Parameters:
- `logical_name` — string, a ROOT/ logical name identifying the spec node

Returns:
- `ParsedNode` record

Errors:
- `"unexpected content before first heading"` — file body has non-blank lines before the first level-1 heading
- `"node name does not match"` — the first `#` heading, after normalization, does not match the normalized logical name
- `"duplicate public section"` — more than one `# Public` section is found
- `"duplicate subsection"` — two `##` headings within `# Public` normalize to the same text

### Steps

1. Resolve the file path.
   Call `ResolvePath(logical_name)` from `logical_names`.
   This yields an absolute file path to the `_node.md` file.

2. Open the file for reading.
   Call `OpenFileReader(file_path)`.
   If the file cannot be opened, raise error "file unreadable".

3. Skip the frontmatter block.
   Read lines one by one.
   - If the very first line is exactly `"---"`, enter frontmatter mode:
     continue reading and discarding lines until a second line containing
     exactly `"---"` is found (this closes the frontmatter block).
   - If the first line is not `"---"`, no frontmatter is present;
     do not consume any lines (treat the file body as starting immediately).

4. Read remaining lines into a body list.
   Continue calling `ReadLine` until "end of file" is raised.
   Normalize CRLF to LF as provided by `file_reader` (already done by ReadLine).
   Store the original lines; do not strip them yet.

5. Track fenced code block state.
   Maintain a boolean `inside_fence`, initially false.
   A line that begins with exactly three or more backtick characters (`` ` ``)
   or three or more tilde characters (`~`) toggles `inside_fence`.
   - When `inside_fence` is true, any line that looks like a heading is treated
     as plain content, not a structural heading.

6. Collect raw sections from the body.
   Iterate over body lines, maintaining:
   - `current_level` — 1 or 2 (the depth of the most recently seen structural heading)
   - `current_heading` — string (the heading text, original, without the `#` prefix or surrounding whitespace)
   - `current_lines` — list of strings accumulating content

   For each line:

   a. Update `inside_fence` if the line starts a or closes a fence (step 5 rule).

   b. If `inside_fence` is false and the line matches a level-1 heading
      (starts with exactly `"# "` or is exactly `"#"`):
      - Flush the current accumulator (see step 7).
      - Set `current_level` = 1, `current_heading` = text after the leading `# `,
        `current_lines` = empty list.

   c. Else if `inside_fence` is false and the line matches a level-2 heading
      (starts with exactly `"## "` or is exactly `"##"`):
      - Flush the current accumulator (see step 7).
      - Set `current_level` = 2, `current_heading` = text after the leading `## `,
        `current_lines` = empty list.

   d. Else (plain content, or heading inside fence):
      - Append the line to `current_lines`.

   After all lines are processed, flush the final accumulator (step 7).

7. Flush accumulator sub-procedure.
   When flushing, if `current_heading` is set (i.e., a heading was seen):
   - Join `current_lines` with newline characters into a single string.
   - Trim leading and trailing blank lines from the joined string.
   - Produce a raw segment record: `{level, heading: current_heading, content: trimmed string}`.
   - Append to a raw segment list.
   If no heading has been set yet (content before first heading), and
   `current_lines` contains any non-blank line, this is a pre-heading content
   error — raise error "unexpected content before first heading".

8. Validate the node name.
   The first raw segment must have `level` = 1.
   Normalize its `heading` using `NormalizeName`.
   Also normalize the `logical_name` for comparison:
   - Strip any qualifier (parenthetical suffix) using `ExtractQualifier` logic:
     if a qualifier exists, remove it.
   - Take only the last path component (the part after the final `/`).
   - Apply `NormalizeName` to that component.
   If the two normalized strings do not match, raise error "node name does not match".

9. Build Section and Subsection records from raw segments.
   Group raw segments into sections:

   Iterate through all raw segments in order.
   Maintain `current_section` (a Section being assembled) and a result list of Sections.

   For each raw segment:
   - If `level` = 1:
     - If `current_section` exists, append it to result list.
     - Start a new Section:
       `heading` = segment heading
       `content` = segment content
       `subsections` = empty list
   - If `level` = 2:
     - If no `current_section` exists, this is unexpected; treat as orphaned
       (implementation may create a synthetic unnamed section or raise an error —
       the spec does not define this case; treat the subsection as belonging to
       the nearest preceding section; if none exists, skip).
     - Append a new Subsection to `current_section.subsections`:
       `heading` = segment heading
       `content` = segment content

   After iteration, if `current_section` exists, append it to result list.

10. Classify sections into ParsedNode fields.
    Iterate over result list of Sections.
    For each Section, normalize its `heading` with `NormalizeName`:

    - If normalized heading = `"public"`:
      - If `ParsedNode.public` is already set, raise error "duplicate public section".
      - Set `ParsedNode.public` = this Section.
      - Validate subsections for duplicates (step 11).

    - Else if normalized heading = `"agent"`:
      - Set `ParsedNode.agent` = this Section
        (spec does not define a "duplicate agent" error; treat as last-wins or
        first-wins; prefer first-wins for safety).

    - Else if this is the first section (it is the name section):
      - Set `ParsedNode.name_section` = this Section.

    - Else:
      - Append to `ParsedNode.private`.

    Note: the name section is always the first Section in the result list
    (validated in step 8). Process it separately before classifying the rest.

11. Validate subsection uniqueness within `# Public`.
    For each pair of Subsections in `public.subsections`:
    - Normalize both headings with `NormalizeName`.
    - If any two normalized headings are equal, raise error "duplicate subsection".

12. Return the completed `ParsedNode` record.

---

## Heading detection rules (detail for step 6)

A line is a level-1 heading if:
- It starts with the characters `#` followed by a space, then the heading text; OR
- It is exactly the single character `#` (empty heading).

A line is a level-2 heading if:
- It starts with exactly `##` followed by a space, then the heading text; OR
- It is exactly `##`.

A line starting with `###` or deeper is NOT structural — it is treated as content.

---

## Fenced code block detection (detail for step 5)

A fence-start or fence-end line is any line where the first non-whitespace
characters are three or more consecutive backticks (`` ``` ``) or three or
more consecutive tildes (`~~~`).

`inside_fence` starts as false. Each time such a line is encountered,
toggle `inside_fence`. When `inside_fence` is true, heading detection
is suppressed for all lines until the fence closes.
