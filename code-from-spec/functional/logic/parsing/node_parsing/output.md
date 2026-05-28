<!-- code-from-spec: ROOT/functional/logic/parsing/node_parsing@B0fYdwRkD6IeO2alo4a4qC4BSX0 -->

# node_parsing

## Records

```
record NodeSubsection
  heading: string        (normalized)
  content: string        (raw markdown, leading/trailing blank lines trimmed)

record NodeSection
  heading: string        (normalized)
  content: string        (raw markdown, leading/trailing blank lines trimmed)
  subsections: list of NodeSubsection

record Node
  name_section: NodeSection
  public: optional NodeSection
  agent: optional NodeSection
  private: list of NodeSection
```

---

## function NodeParse(logical_name: string) -> Node

**Errors:**
- `"not a ROOT reference"` — the logical name does not start with `ROOT/`.
- `"has qualifier"` — the logical name contains a parenthetical qualifier.
- path errors — propagated from `LogicalNameToPath` and `FileOpen`.
- `"file unreadable"` — the file cannot be opened or read.
- `"unexpected content before first heading"` — non-blank content appears before the first level-1 heading, no level-1 heading exists at all, or the frontmatter closing `---` is never found.
- `"node name does not match"` — the first heading does not match the logical name after normalization.
- `"duplicate public section"` — more than one `# Public` section exists.
- `"duplicate agent section"` — more than one `# Agent` section exists.
- `"duplicate subsection"` — two `##` headings within the same section normalize to the same text.

---

### Steps

1. If `LogicalNameIsArtifact(logical_name)` returns true,
   raise error `"not a ROOT reference"`.

2. If `LogicalNameHasQualifier(logical_name)` returns true,
   raise error `"has qualifier"`.

3. Call `LogicalNameToPath(logical_name)` to get the file path.
   Propagate any errors from that call.

4. Call `FileOpen(cfs_path)` to open the file for reading.
   If the file cannot be opened, raise error `"file unreadable"`.
   From this point forward, `FileClose` must be called before
   returning or raising any error.

5. **Skip frontmatter:**
   Read the first line using `FileReadLine`.
   If end of file is raised immediately, the file is empty —
     close the reader, raise error `"unexpected content before first heading"`.
   If the first line is exactly `"---"`:
     Read lines one by one until a line that is exactly `"---"` is found.
     If end of file is reached before finding the closing `"---"`,
       close the reader, raise error `"unexpected content before first heading"`.
   Otherwise (first line is not `"---"`):
     Treat the first line as the first body line (carry it forward into step 6).

6. **Parse body lines into a flat line list:**
   Continue reading lines with `FileReadLine` until end of file.
   Collect all lines (including any first body line carried from step 5).
   Close the reader with `FileClose` when end of file is reached.

7. **Tokenize lines into events (heading or content), respecting fenced code blocks:**

   Initialize:
   - `in_fence` = false
   - `fence_char` = absent
   - `fence_length` = 0

   For each line:

     If `in_fence` is true:
       Check if the line is a closing fence:
         A closing fence is a line consisting only of `fence_length` or more
         consecutive `fence_char` characters (optionally followed by whitespace only).
       If the line is a closing fence:
         Set `in_fence` = false.
       Emit the line as a content event regardless (the fence line itself is content).

     If `in_fence` is false:
       Check if the line is an opening fence:
         An opening fence is a line that starts with 3 or more consecutive
         backtick (`` ` ``) characters or 3 or more consecutive tilde (`~`)
         characters, optionally followed by a language tag or whitespace.
       If the line is an opening fence:
         Set `in_fence` = true.
         Set `fence_char` to the repeated character (`` ` `` or `~`).
         Set `fence_length` to the number of leading repeated characters.
         Emit the line as a content event.
       Otherwise:
         Attempt to parse the line as an ATX heading:
           An ATX heading line starts with one or more `#` characters
           immediately followed by at least one space character.
           Count the leading `#` characters to determine `level`.
           Extract heading text as everything after the leading `# ` prefix.
           Trim leading and trailing whitespace from the heading text.
           Strip optional closing `#` sequence:
             If the trimmed text ends with one or more `#` characters
             preceded by at least one space, remove the closing `#` sequence
             and trim again.
           Normalize the heading text using `NormalizeText`.
         If the line is a valid ATX heading (level 1 or 2):
           Emit a heading event with `level` and normalized `heading_text`.
         Otherwise:
           Emit the line as a content event.
           (Level-3+ headings become content events.)

8. **Build sections from the event stream:**

   Initialize:
   - `sections` = empty list (will hold NodeSection records in order)
   - `current_section` = absent
   - `current_subsection` = absent
   - `current_section_content_lines` = empty list
   - `current_subsection_content_lines` = empty list
   - `found_non_blank_before_first_heading` = false

   Helper — **FlushSubsection:**
     If `current_subsection` is not absent:
       Set `current_subsection.content` = trim blank lines from
         `current_subsection_content_lines` joined with newlines.
       Append `current_subsection` to `current_section.subsections`.
       Set `current_subsection` = absent.
       Set `current_subsection_content_lines` = empty list.

   Helper — **FlushSection:**
     Call FlushSubsection.
     If `current_section` is not absent:
       Set `current_section.content` = trim blank lines from
         `current_section_content_lines` joined with newlines.
       Append `current_section` to `sections`.
       Set `current_section` = absent.
       Set `current_section_content_lines` = empty list.

   For each event:

     If the event is a **level-1 heading**:
       Call FlushSection.
       Create a new NodeSection with:
         `heading` = the normalized heading text
         `content` = `""` (will be set on flush)
         `subsections` = empty list
       Set `current_section` = the new section.
       Set `current_section_content_lines` = empty list.

     If the event is a **level-2 heading**:
       If `current_section` is absent:
         If no level-1 heading has been seen yet,
           treat this line as content (add to pre-heading buffer for step 9 check).
         Else treat as content in the current section.
       Otherwise:
         Call FlushSubsection.
         Check for duplicate subsection:
           For each existing subsection in `current_section.subsections`,
           if its `heading` equals the new normalized heading text,
           raise error `"duplicate subsection"`.
         Create a new NodeSubsection with:
           `heading` = the normalized heading text
           `content` = `""` (will be set on flush)
         Set `current_subsection` = the new subsection.
         Set `current_subsection_content_lines` = empty list.

     If the event is a **content line**:
       If `current_section` is absent:
         If the line is not blank,
           set `found_non_blank_before_first_heading` = true.
       Else if `current_subsection` is not absent:
         Append the line to `current_subsection_content_lines`.
       Else:
         Append the line to `current_section_content_lines`.

   After processing all events, call FlushSection.

9. **Validate pre-heading content:**
   If `found_non_blank_before_first_heading` is true,
   raise error `"unexpected content before first heading"`.

10. **Validate that at least one section was found:**
    If `sections` is empty,
    raise error `"unexpected content before first heading"`.

11. **Classify sections:**

    Initialize:
    - `name_section` = absent
    - `public_section` = absent
    - `agent_section` = absent
    - `private_sections` = empty list

    Normalize the logical name itself using `NormalizeText` to get
    `normalized_logical_name`.

    For each section in `sections` (in order):

      If `name_section` is absent (this is the first section):
        If `section.heading` does not equal `normalized_logical_name`,
          raise error `"node name does not match"`.
        Set `name_section` = this section.

      Else if `section.heading` equals `"public"`:
        If `public_section` is not absent,
          raise error `"duplicate public section"`.
        Set `public_section` = this section.

      Else if `section.heading` equals `"agent"`:
        If `agent_section` is not absent,
          raise error `"duplicate agent section"`.
        Set `agent_section` = this section.

      Else:
        Append this section to `private_sections`.

12. **Return** a Node record:
    - `name_section` = `name_section`
    - `public` = `public_section` (absent if no `# Public` section was found)
    - `agent` = `agent_section` (absent if no `# Agent` section was found)
    - `private` = `private_sections` (in order of appearance)

---

## Helper — TrimBlankLines(lines: list of strings) -> string

1. Remove leading entries from `lines` that are blank (empty or whitespace-only).
2. Remove trailing entries from `lines` that are blank.
3. Join the remaining entries with newline characters.
4. Return the resulting string.
   If no entries remain, return `""`.
