<!-- code-from-spec: ROOT/functional/logic/parsing/node_parsing@8kEjeEDVtWWqlOvCfDsfpyi8XlU -->

# Data Records

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


# Functions

---

function NodeParse(logical_name: string) -> Node
  errors:
    - "not a ROOT reference": the logical name starts with ARTIFACT/ or does not start with ROOT/
    - "has qualifier": the logical name contains a parenthetical qualifier
    - "file unreadable": the file cannot be opened or read
    - "unexpected content before first heading": non-blank content appears before the first level-1 heading, or no level-1 heading exists
    - "node name does not match": the first heading text does not match the logical name after normalization
    - "duplicate public section": more than one level-1 heading normalizes to "public"
    - "duplicate agent section": more than one level-1 heading normalizes to "agent"
    - "duplicate subsection": two level-2 headings within # Public normalize to the same text

  1. If LogicalNameIsArtifact(logical_name) is true,
       raise error "not a ROOT reference".

  2. If LogicalNameHasQualifier(logical_name) is true,
       raise error "has qualifier".

  3. Call LogicalNameToPath(logical_name) to get cfs_path.
     If LogicalNameToPath raises a path error, propagate it.

  4. Call FileOpen(cfs_path) to get reader.
     If FileOpen raises an error, raise error "file unreadable".

  5. (Everything from here on must ensure FileClose(reader) is called before
     returning or raising any error.)

  6. Skip frontmatter:
     Read the first non-end-of-file line using FileReadLine.
     If end of file is raised immediately, close reader and raise
       "unexpected content before first heading".
     If that first line is exactly "---":
       Read lines with FileReadLine until a line is exactly "---".
       If end of file is reached before finding the closing "---",
         close reader and raise "unexpected content before first heading".
     Else:
       The first line is part of the body — hold it as the pending line
       to process in step 7 (do not discard it).

  7. Parse the body into raw sections.
     Maintain state:
       - current_section_level: integer (0 = none started yet)
       - current_heading: string
       - content_lines: list of strings (lines accumulated for current section or subsection)
       - sections: ordered list of records with fields:
           level: integer (1 or 2)
           heading: string (normalized)
           content_lines: list of strings
       - in_fenced_block: boolean, initially false
       - fence_char: string ("`" or "~"), set when entering a fenced block
       - fence_length: integer, set when entering a fenced block

     Process each line (including the held line from step 6, if any):

       a. Fenced code block detection (checked before heading detection):
          If in_fenced_block is false:
            If the line consists of three or more consecutive backtick characters
            (optionally followed by a language tag and nothing else structural),
            or three or more consecutive tilde characters (same rule):
              Set in_fenced_block to true.
              Set fence_char to the character used ("`" or "~").
              Set fence_length to the count of that character at the start of the line.
              Treat the line as content (append to content_lines).
              Continue to next line.
          If in_fenced_block is true:
            If the line starts with at least fence_length occurrences of fence_char,
            and contains no other non-space characters after the closing sequence:
              Set in_fenced_block to false.
            Treat the line as content (append to content_lines).
            Continue to next line.

       b. Heading detection (only when not in_fenced_block):
          Check if the line matches the ATX heading pattern:
            - Starts with one or more "#" characters.
            - Followed by at least one space.
            - Heading text is everything after the leading "# " prefix,
              trimmed of leading and trailing whitespace.
            - If the trimmed heading text ends with one or more "#" characters
              preceded by at least one space, strip that closing sequence to get
              the final heading text.
          If the line does NOT match this pattern, treat it as content
          (append to content_lines) and continue to next line.

          If the line matches:
            Extract level (count of leading "#") and raw heading text.
            Only level 1 and level 2 are structurally significant
            (level 3 and deeper are treated as content — append to content_lines
            and continue to next line).
            Compute normalized_heading = NormalizeText(raw heading text).

            If level is 1:
              Flush current content:
                If there is an open section or subsection (current_section_level > 0),
                save a record with the current heading and content_lines to sections.
              Start a new level-1 section:
                Set current_section_level = 1.
                Set current_heading = normalized_heading.
                Reset content_lines to empty.

            If level is 2:
              Flush current content:
                If current_section_level > 0,
                save a record with the current level, heading, and content_lines to sections.
              Start a new level-2 subsection:
                Set current_section_level = 2.
                Set current_heading = normalized_heading.
                Reset content_lines to empty.

       c. If the line is content (not a structural heading, not inside a fenced
          block opening/closing check above):
          Append the line to content_lines.

     After all lines are processed:
       Flush remaining content_lines as a final record if current_section_level > 0.

  8. Validate and classify sections from the sections list built in step 7.

     Before any level-1 heading is encountered, if any non-blank line was
     accumulated in content_lines (which would have been flushed as a floating
     record), raise error "unexpected content before first heading".
     Also, if the sections list is empty (no level-1 heading found at all),
     raise error "unexpected content before first heading".

     Initialize:
       - node_name_section: absent
       - public_section: absent
       - agent_section: absent
       - private_sections: empty list
       - seen_subsection_headings: empty set (for duplicate detection within Public)

     Process records in order:

       For each level-1 record:

         If node_name_section is absent:
           This is the name section.
           Compute expected = NormalizeText(logical_name).
           If record.heading does not equal expected,
             close reader and raise error "node name does not match".
           Set node_name_section to a NodeSection with:
             heading = record.heading
             content = trimmed content (trim leading/trailing blank lines from content_lines)
             subsections = empty list (level-2 records that follow this section
               before the next level-1 are handled below)

         Else if record.heading equals "public":
           If public_section is already present,
             close reader and raise error "duplicate public section".
           Set public_section to a NodeSection with heading = "public",
             content = trimmed content, subsections to be filled from subsequent level-2 records.

         Else if record.heading equals "agent":
           If agent_section is already present,
             close reader and raise error "duplicate agent section".
           Set agent_section to a NodeSection with heading = "agent",
             content = trimmed content, subsections = empty list.

         Else:
           Append a NodeSection to private_sections with
             heading = record.heading,
             content = trimmed content,
             subsections = empty list.

       For each level-2 record:
         Determine which level-1 section it belongs to (it follows the most
         recently started level-1 section in the ordered sections list).
         If the owning section is the public section:
           If seen_subsection_headings already contains record.heading,
             close reader and raise error "duplicate subsection".
           Add record.heading to seen_subsection_headings.
           Append a NodeSubsection with
             heading = record.heading,
             content = trimmed content
           to public_section.subsections.
         Else:
           Level-2 records outside of Public should not appear as separate
           records — they are treated as content in step 7b. (This case does
           not arise given the parsing rules above.)

  9. If node_name_section is absent (no level-1 heading was found),
     close reader and raise error "unexpected content before first heading".

  10. Call FileClose(reader).

  11. Return a Node with:
        name_section = node_name_section
        public = public_section (absent if none found)
        agent = agent_section (absent if none found)
        private = private_sections


# Helper: Trim Leading and Trailing Blank Lines

function TrimBlankLines(lines: list of strings) -> string

  1. Remove leading lines that are blank (empty or whitespace-only).
  2. Remove trailing lines that are blank (empty or whitespace-only).
  3. Join the remaining lines with newline characters.
  4. Return the resulting string.
     If no lines remain, return "".


# Helper: ATX Heading Parse

function ParseAtxHeading(line: string) -> optional record with fields (level: integer, text: string)

  1. Count the number of leading "#" characters. Call this count level.
     If level is 0, return absent (not a heading).

  2. If the character at position level is not a space,
     return absent (e.g., "#Foo" has no space, not a heading).

  3. Extract raw_text = everything after the leading "#" characters and the
     single required space, trimmed of leading and trailing whitespace.

  4. Check for closing hash sequence:
     If raw_text ends with one or more "#" characters and the character
     immediately before that trailing sequence is a space,
       strip the trailing space(s) and "#" characters to get raw_text.
       (e.g., "## Foo ##" -> text is "Foo")
     Else leave raw_text as is.

  5. Return record with level = level, text = raw_text.


# Helper: Fenced Code Block Detection

function IsFenceOpener(line: string) -> optional record with fields (fence_char: string, fence_length: integer)

  1. Let trimmed = the line with trailing whitespace removed.
  2. Count the leading backtick characters. If count >= 3 and the
     remaining characters (after the backticks) do not contain any backtick,
     return record with fence_char = "`", fence_length = count.
  3. Count the leading tilde characters. If count >= 3 and the remaining
     characters (after the tildes) do not contain any tilde,
     return record with fence_char = "~", fence_length = count.
  4. Return absent.

function IsFenceCloser(line: string, fence_char: string, fence_length: integer) -> boolean

  1. Let trimmed = the line with trailing whitespace removed.
  2. Count the leading occurrences of fence_char in trimmed. Call this count.
  3. If count >= fence_length and trimmed contains no other characters
     beyond those leading fence_char characters, return true.
  4. Return false.
