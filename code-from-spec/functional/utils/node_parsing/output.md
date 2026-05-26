<!-- code-from-spec: ROOT/functional/utils/node_parsing@R71DvYscAVs0Lp2XpuQu4Qh5QT0 -->

# node_parsing

## Data Structures

```
record Subsection
  heading: string          -- raw heading text (after stripping the leading "##" and whitespace)
  content: string          -- trimmed body text belonging to this subsection

record Section
  heading: string          -- raw heading text (after stripping the leading "#" and whitespace)
  content: string          -- trimmed body text that belongs directly to this section
                           -- (before any level-2 sub-heading)
  subsections: list of Subsection

record ParsedNode
  name_section: Section    -- the mandatory first "# <name>" section
  public:   optional Section  -- the "# Public" section, if present
  agent:    optional Section  -- the "# Agent" section, if present
  private:  list of Section   -- any remaining top-level sections
```

## Functions

---

### ParseNode(logical_name) -> ParsedNode

**Purpose:** Resolve the node file, skip its frontmatter, and parse the
remaining Markdown body into a structured `ParsedNode`.

**Errors:**
- `"unexpected content before first heading"` — the body (after the frontmatter)
  contains non-blank text before the first `#` heading.
- `"node name does not match"` — the first heading, when normalized, does not
  match `logical_name` after normalization.
- `"duplicate public section"` — more than one `# Public` section is found.
- `"duplicate subsection"` — two `##` headings inside `# Public` normalize to
  the same string.

**Steps:**

1. Call `ResolvePath(logical_name)` from `logical_names` to obtain `file_path`.

2. Call `OpenFileReader(file_path)`.
   If the file cannot be opened, raise the error from `OpenFileReader`.

3. **Skip frontmatter.**
   Read lines one at a time until the first line that is exactly `"---"`.
   That line is the opening delimiter. Continue reading lines until the next
   line that is exactly `"---"`. That second `"---"` is the closing delimiter.
   Discard all lines read so far (frontmatter).
   If the file has no opening `"---"`, treat the entire file as body
   (no frontmatter to skip).

4. **Read the body.**
   Read all remaining lines from the reader into a list called `body_lines`.

5. **Check for content before the first heading.**
   Iterate over `body_lines`. For each line, before encountering any line that
   starts with `"#"` (outside a fenced code block — see step 7 for the
   fenced-block rule):
     - If the line is non-blank, raise error `"unexpected content before first heading"`.

6. **Partition the body into raw sections.**
   Walk `body_lines` maintaining a `fence_open` flag (initially false) and
   building a list of raw sections.

   A **raw section** is a record:
   ```
   record RawSection
     level:   integer   -- 1 or 2
     heading: string    -- text after the leading "#" or "##" stripped of whitespace
     lines:   list of string
   ```

   For each line in `body_lines`:

   a. If the line starts with ` ``` ` (three or more backticks), toggle `fence_open`.
      Append the line to the current raw section's `lines` and continue.

   b. If `fence_open` is true, append the line to the current raw section's
      `lines` and continue (do not treat it as a structural heading).

   c. If the line matches `"## <text>"` (starts with `"## "` or is exactly `"##"`):
      Close the current raw section (if any). Start a new raw section with
      `level = 2` and `heading = <text>` (trimmed). Its `lines` list is empty.

   d. If the line matches `"# <text>"` (starts with `"# "` or is exactly `"#"`),
      AND does NOT match `"## "` (i.e., it is a true level-1 heading):
      Close the current raw section (if any). Start a new raw section with
      `level = 1` and `heading = <text>` (trimmed). Its `lines` list is empty.

   e. Otherwise, if a current raw section is open, append the line to its `lines`.
      If no section is open yet, skip (these are blank lines before the first heading,
      already validated in step 5).

   After processing all lines, close the last open raw section.

7. **Validate the first heading.**
   The first raw section must have `level = 1`.
   Apply `NormalizeName` to its `heading`.
   Apply `NormalizeName` to `logical_name` (strip any qualifier first using
   `ExtractQualifier`; use only the final path segment — the part after the
   last `"/"` in the unqualified name).
   If the two normalized strings are not equal, raise error `"node name does not match"`.

8. **Build Section records from raw sections.**
   Convert each raw section into either a `Section` or a `Subsection`:

   For a raw section, compute `content`:
     - Join its `lines` with newline characters.
     - Trim leading and trailing blank lines (lines that are empty or contain
       only whitespace).

9. **Assemble ParsedNode.**
   Group the raw sections hierarchically:

   a. The first raw section (level 1) becomes `name_section`. It can have no
      level-2 children because a `##` before any other `#` would be orphaned
      — treat such orphaned `##` sections as belonging to the immediately
      preceding `#` section (which in this case is `name_section`).

   b. Walk the remaining raw sections in order. For each:
      - If `level = 1`: finalize the previous top-level section (if any) and
        start a new one.
      - If `level = 2`: the current heading belongs as a `Subsection` of the
        most recently opened level-1 section. If no level-1 section is open
        yet (orphaned `##`), attach it to `name_section`.

   c. For each level-1 section, collect its level-2 children as `subsections`.
      The level-1 section's own `content` is built from the lines that appear
      after the `#` heading and before the first `##` child.

   d. Classify each completed top-level section:
      - Normalize the heading with `NormalizeName`.
      - If it equals `"public"`:
        - If `public` is already set, raise error `"duplicate public section"`.
        - Before assigning, validate that no two subsections within it have the
          same normalized heading; if they do, raise error `"duplicate subsection"`.
        - Set `parsed_node.public` to this section.
      - If it equals `"agent"`:
        - Set `parsed_node.agent` to this section.
      - Otherwise: append to `parsed_node.private`.

10. Return `parsed_node`.

---

## Invariants and Contracts

- Only `#` (level-1) and `##` (level-2) headings are structural. Level-3 and
  deeper headings (`###`, `####`, …) are treated as ordinary content lines.
- Headings inside fenced code blocks (delimited by lines starting with three or
  more backticks) are not structural; they are content lines of the enclosing section.
- Leading and trailing blank lines in a section's or subsection's `content` are
  trimmed before returning.
- Heading comparison (for matching `"public"`, `"agent"`, and for the node-name
  check) always uses `NormalizeName` from `ROOT/functional/utils/name_normalization`.
