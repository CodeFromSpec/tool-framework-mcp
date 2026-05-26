<!-- code-from-spec: ROOT/functional/utils/node_parsing@sivRqA3XAKgYX0ZwyK4RlFqhIPU -->

# node_parsing

Parses a spec node file into structured sections and subsections,
given a logical name. Relies on `logical_names` for path resolution,
`file_reader` for sequential reading, and `name_normalization` for
heading comparison.

---

## Data Structures

```
record Subsection
  heading: string       -- the raw heading text (after stripping the "##" marker)
  content: string       -- all lines between this subsection heading and the next structural heading, trimmed

record Section
  heading: string       -- the raw heading text (after stripping the "#" marker)
  content: string       -- lines directly under this section heading (before any subsection), trimmed
  subsections: list of Subsection

record ParsedNode
  name_section: Section         -- the first "# <name>" section (always present)
  public: optional Section      -- the "# Public" section, if present
  agent: optional Section       -- the "# Agent" section, if present
  private: list of Section      -- all other level-1 sections that are not the name, public, or agent sections
```

---

## Functions

---

### function ParseNode(logical_name) -> ParsedNode

**Parameters**
- `logical_name` — string; a ROOT/ logical name identifying the node to parse.

**Returns**
- a `ParsedNode` record with all sections populated.

**Errors**
- `"unexpected content before first heading"` — the file body (after frontmatter) has non-blank content before the first level-1 heading.
- `"node name does not match"` — the first level-1 heading does not match the logical name after normalization.
- `"duplicate public section"` — more than one `# Public` section is found.
- `"duplicate subsection"` — two level-2 headings within `# Public` normalize to the same text.

**Steps**

1. Call `ResolvePath(logical_name)` from `logical_names` to get the file path.
   If resolution raises an error, propagate it.

2. Call `OpenFileReader(file_path)` from `file_reader` to open the file.
   If the file is unreadable, raise error `"cannot read file: <file_path>"`.

3. Skip the frontmatter block:
   - Read lines one at a time.
   - Look for the first line that is exactly `"---"`. This marks the start of the frontmatter.
     If no such line is found before the end of file, the file has no frontmatter; treat all lines as body.
   - If a `"---"` start was found, continue reading until a second line that is exactly `"---"`.
     This marks the end of the frontmatter. All subsequent lines are the body.
   - Discard all frontmatter lines. The reader is now positioned at the first body line.

4. Collect all remaining lines from the reader into a list called `body_lines`.
   Normalize each line: convert CRLF to LF (already done by `ReadLine`).

5. Track a boolean `inside_fenced_block`, initially false.
   Track `pending_lines` (accumulator for content not yet assigned to a section), initially empty.
   Track `sections` as an ordered list of raw section records, initially empty.
   Track `current_section` (optional), initially absent.
   Track `current_subsection` (optional), initially absent.

   For each line in `body_lines`:

   a. If the line starts with exactly three backticks (` ``` `) — toggle `inside_fenced_block`.
      Append the line to the appropriate accumulator and continue to the next line.
      (Headings inside fenced blocks are not structural.)

   b. If `inside_fenced_block` is true, append the line to the appropriate accumulator and continue.

   c. If the line starts with `"## "` (two hashes followed by a space):
      - Extract the heading text: everything after `"## "`.
      - If `current_subsection` is present:
        - Close it: set its `content` to the trimmed accumulator, append it to `current_section.subsections`.
      - If `current_section` is absent:
        - Append the line to `pending_lines` and continue.
          (Level-2 headings before any level-1 heading are treated as content.)
      - Start a new `current_subsection` with the extracted heading and an empty content accumulator.
      - Continue to the next line.

   d. If the line starts with `"# "` (one hash followed by a space):
      - Extract the heading text: everything after `"# "`.
      - If `current_subsection` is present:
        - Close it: set its `content` to the trimmed accumulator, append it to `current_section.subsections`.
        - Clear `current_subsection`.
      - If `current_section` is present:
        - Close it: set its `content` to the trimmed content accumulator, append it to `sections`.
      - If `current_section` is absent and `pending_lines` has any non-blank line:
        - Raise error `"unexpected content before first heading"`.
      - Start a new `current_section` with the extracted heading, empty content accumulator, and empty subsections list.
      - Continue to the next line.

   e. Otherwise, the line is plain content:
      - If `current_subsection` is present, append the line to the subsection content accumulator.
      - Else if `current_section` is present, append the line to the section content accumulator.
      - Else append the line to `pending_lines`.

6. After processing all lines:
   - If `current_subsection` is present:
     - Close it: set its `content` to the trimmed accumulator, append it to `current_section.subsections`.
   - If `current_section` is present:
     - Close it: set its `content` to the trimmed content accumulator, append it to `sections`.
   - If `sections` is empty and `pending_lines` has any non-blank line:
     - Raise error `"unexpected content before first heading"`.

7. If `sections` is empty, raise error `"node name does not match"`.
   (A valid node must have at least one level-1 heading.)

8. Take the first element of `sections` as the candidate name section.
   Normalize its heading using `NormalizeName`.
   Derive the expected name from `logical_name`:
   - Strip any qualifier (text inside parentheses at the end).
   - Take the last path segment after the final `/`.
   - Apply `NormalizeName` to it.
   If the normalized heading does not equal the normalized expected name,
   raise error `"node name does not match"`.

9. Set `name_section` to the first element of `sections`.
   Initialize `public_section` as absent, `agent_section` as absent, `private_sections` as empty list.

10. For each remaining section in `sections` (all except the first):
    - Normalize the heading using `NormalizeName`.
    - If the normalized heading equals `"public"`:
      - If `public_section` is already present, raise error `"duplicate public section"`.
      - Check for duplicate subsections within this section:
        - Collect the normalized heading of each subsection.
        - If any two normalize to the same text, raise error `"duplicate subsection"`.
      - Set `public_section` to this section.
    - Else if the normalized heading equals `"agent"`:
      - Set `agent_section` to this section.
    - Else:
      - Append this section to `private_sections`.

11. Build and return a `ParsedNode` record:
    - `name_section` = the first section (from step 9)
    - `public` = `public_section` (may be absent)
    - `agent` = `agent_section` (may be absent)
    - `private` = `private_sections`

---

## Helper: Trimming content accumulators

When "trimming" a content accumulator (a list of lines) to produce the `content` string:

1. Remove all leading blank lines from the list.
2. Remove all trailing blank lines from the list.
3. Join the remaining lines with newline (`"\n"`).

A blank line is one that contains only whitespace characters.

---

## Helper: Detecting fenced code block boundaries

A line toggles the `inside_fenced_block` flag if and only if:
- After stripping leading whitespace, it starts with exactly three backticks (` ``` `).
- The check is per line, so a line starting with four or more backticks also qualifies
  as a fence boundary (matching the CommonMark fenced code block rule).

For simplicity, tildes (`~~~`) are not treated as fence delimiters in this implementation
unless the spec is extended to require it.

---

## Dependency summary

| Dependency | Function used |
|---|---|
| `ROOT/functional/utils/logical_names` | `ResolvePath(logical_name)` |
| `ROOT/functional/utils/file_reader` | `OpenFileReader(file_path)`, `ReadLine(reader)` |
| `ROOT/functional/utils/name_normalization` | `NormalizeName(raw_string)` |
