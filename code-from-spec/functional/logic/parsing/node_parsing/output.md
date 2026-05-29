<!-- code-from-spec: ROOT/functional/logic/parsing/node_parsing@QKQ501PkuBVW5KlIOvSM4nTqPU4 -->

# node_parsing

## Records

```
record NodeSubsection
  heading:     string          -- normalized text (via NormalizeText)
  raw_heading: string          -- original line as read from file
  content:     list of string  -- lines with leading/trailing blank lines trimmed

record NodeSection
  heading:     string          -- normalized text (via NormalizeText)
  raw_heading: string          -- original line as read from file
  content:     list of string  -- lines before first ##, trimmed
  subsections: list of NodeSubsection

record Node
  name_section: NodeSection
  public:       optional NodeSection
  agent:        optional NodeSection
  private:      list of NodeSection  -- in order of appearance
```

---

## function NodeParse(logical_name: string) -> Node

  errors:
  - "not a ROOT reference": logical name does not start with ROOT/.
  - "has qualifier": logical name contains a parenthetical qualifier.
  - (path errors): propagated from FileOpen / LogicalNameToPath.
  - "file unreadable": the file cannot be opened or read.
  - "unexpected content before first heading": non-blank content appears
    before the first level-1 heading, or no level-1 heading exists at all.
  - "node name does not match": the first heading text does not match the
    logical name after normalization.
  - "duplicate public section": more than one # Public section exists.
  - "duplicate agent section": more than one # Agent section exists.
  - "duplicate subsection": two ## headings within the same section
    normalize to the same text.

### Steps

  1. If LogicalNameIsArtifact(logical_name) is true,
       raise error "not a ROOT reference".

  2. If LogicalNameHasQualifier(logical_name) is true,
       raise error "has qualifier".

  3. Call LogicalNameToPath(logical_name) to get file_path.
     Propagate any path errors.

  4. Call FileOpen(file_path) to get reader.
     If the file cannot be opened, raise error "file unreadable".

  5. Set node_result to absent (will be built up).
     Set error_to_raise to absent.
     Proceed to step 6 inside a context that always calls
     FileClose(reader) before returning, whether an error is
     raised or not.

  6. Skip the frontmatter:
     a. Call FileReadLine(reader) to get first_line.
        If "end of file" is raised, the file is empty —
          raise error "unexpected content before first heading".
     b. If first_line equals exactly "---":
          read lines with FileReadLine until a line equal to "---" is found.
          If "end of file" is raised before finding the closing "---",
            raise error "unexpected content before first heading".
     c. If first_line does not equal "---":
          treat first_line as the first body line
          (do not discard it — it will be processed in step 7).

  7. Parse the file body into sections.
     Maintain the following state:
     - pending_line: optional string (the unprocessed line from step 6c, if any)
     - current_section: optional NodeSection being built
     - current_subsection: optional NodeSubsection being built
     - sections_seen: list of NodeSection (completed sections, in order)
     - found_public: boolean, initially false
     - found_agent: boolean, initially false
     - in_fence: boolean, initially false
     - fence_char: string ("`" or "~"), set when in_fence becomes true
     - fence_min_length: integer, set when in_fence becomes true
     - normalized_logical_name: NormalizeText(logical_name)

     Read lines one at a time using FileReadLine.
     If a pending_line exists from step 6c, process it first
     before reading further lines.
     Continue until "end of file" is raised.

     For each line:

     a. Check for fenced code block boundaries (before heading detection):
        - If in_fence is false:
            Count the leading characters of line that are all "`" or all "~".
            If the count is >= 3 and the rest of the line (after those characters)
            is either empty or a language tag (no more fence characters):
              Set in_fence to true.
              Set fence_char to that character.
              Set fence_min_length to that count.
              Treat the line as content (go to step 7h).
        - If in_fence is true:
            Count the leading fence_char characters in line.
            If the count >= fence_min_length and the remainder of the line
            is blank (only whitespace):
              Set in_fence to false.
              Treat the line as content (go to step 7h).
            Otherwise, treat the line as content (go to step 7h).

     b. If in_fence is true, treat the line as content (go to step 7h).

     c. Detect ATX headings:
        Count the number of leading "#" characters in line — call it level.
        If level is 0, treat the line as content (go to step 7h).
        If the character immediately after the "#" characters is not a space,
          treat the line as content (go to step 7h).
        Extract heading_text: everything after the leading "#" characters and
          the single required space, then trim leading and trailing whitespace.
        Strip optional closing "#" sequence from heading_text:
          If heading_text ends with one or more "#" characters preceded by
          at least one space, remove the trailing whitespace and "#" characters.
          Trim the result again.
        Set raw_heading to line (the original line, unchanged).
        Set heading_norm to NormalizeText(heading_text).

     d. If level is 1, process a section boundary:
        i.   Finalize current_subsection (if any):
               Trim leading and trailing blank lines from current_subsection.content.
               Append current_subsection to current_section.subsections.
               Set current_subsection to absent.
        ii.  Finalize current_section (if any):
               Trim leading and trailing blank lines from current_section.content.
               Append current_section to sections_seen.
               Set current_section to absent.
        iii. Classify the new section by heading_norm:
               - If sections_seen is empty (this is the first section):
                   If heading_norm does not equal normalized_logical_name,
                     raise error "node name does not match".
               - Else if heading_norm equals "public":
                   If found_public is true, raise error "duplicate public section".
                   Set found_public to true.
               - Else if heading_norm equals "agent":
                   If found_agent is true, raise error "duplicate agent section".
                   Set found_agent to true.
               - Else:
                   (private section — no special checks)
        iv.  Create a new NodeSection:
               heading:     heading_norm
               raw_heading: raw_heading
               content:     empty list
               subsections: empty list
             Set current_section to this new section.

     e. If level is 2, process a subsection boundary:
        If current_section is absent,
          treat the line as content (go to step 7h).
        i.   Finalize current_subsection (if any):
               Trim leading and trailing blank lines from current_subsection.content.
               Append current_subsection to current_section.subsections.
               Set current_subsection to absent.
        ii.  Check for duplicate subsection:
               For each existing subsection in current_section.subsections,
                 if its heading equals heading_norm,
                   raise error "duplicate subsection".
        iii. Create a new NodeSubsection:
               heading:     heading_norm
               raw_heading: raw_heading
               content:     empty list
             Set current_subsection to this new section.

     f. If level >= 3:
          Treat the line as content (go to step 7h).

     g. (Heading processed — do not fall through to 7h for level 1 or 2.)
        Continue to the next line.

     h. Append the line to the content of the innermost active container:
        - If current_subsection is not absent, append to current_subsection.content.
        - Else if current_section is not absent, append to current_section.content.
        - Else (no section has started yet):
            If the line is non-blank, raise error
              "unexpected content before first heading".
            Otherwise, discard the blank line.

  8. After "end of file":
     a. Finalize current_subsection (if any):
          Trim leading and trailing blank lines from current_subsection.content.
          Append current_subsection to current_section.subsections.
     b. Finalize current_section (if any):
          Trim leading and trailing blank lines from current_section.content.
          Append current_section to sections_seen.
     c. If sections_seen is empty,
          raise error "unexpected content before first heading".

  9. Call FileClose(reader).

  10. Assemble the Node record from sections_seen:
      - The first entry in sections_seen is name_section.
      - For each remaining entry:
          if heading equals "public",  assign to node.public.
          if heading equals "agent",   assign to node.agent.
          otherwise, append to node.private (in order).
      - node.public and node.agent are absent if no such section was found.
      - node.private is an empty list if no private sections were found.

  11. Return the assembled Node.

---

## Helper: TrimBlankLines(lines: list of string) -> list of string

  1. Remove leading entries from lines that are blank (empty or whitespace-only).
  2. Remove trailing entries from lines that are blank (empty or whitespace-only).
  3. Return the trimmed list.

  (Used in place of "trim leading and trailing blank lines from <x>.content"
  in the steps above.)
