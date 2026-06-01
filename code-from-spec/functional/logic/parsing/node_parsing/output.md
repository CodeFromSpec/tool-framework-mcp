<!-- code-from-spec: ROOT/functional/logic/parsing/node_parsing@hqZu165vvnpWECCgxHHSUpzrBLE -->

# namespace: parsenode

---

## Records

record NodeSubsection
  heading:     string          -- normalized text (via NormalizeText)
  raw_heading: string          -- original line as read from file
  content:     list of string  -- lines as returned by FileReadLine

record NodeSection
  heading:     string          -- normalized text (via NormalizeText)
  raw_heading: string          -- original line as read from file
  content:     list of string  -- lines before the first ## heading
  subsections: list of NodeSubsection

record Node
  name_section: NodeSection
  public:       optional NodeSection
  agent:        optional NodeSection
  private:      list of NodeSection  -- in file order

---

## Functions

function NodeParse(logical_name: string) -> Node
  errors:
    - NotARootReference: the logical name does not start with ROOT/.
    - HasQualifier: the logical name contains a parenthetical qualifier.
    - FileUnreadable: the file cannot be opened or read.
    - UnexpectedContentBeforeFirstHeading: file body has non-blank
      content before the first level-1 heading, or has no level-1
      heading at all. Blank lines before the first heading are not
      an error.
    - NodeNameDoesNotMatch: the first heading does not match the
      logical name after normalization.
    - DuplicatePublicSection: more than one Public section exists.
    - DuplicateAgentSection: more than one Agent section exists.
    - DuplicateSubsection: two level-2 headings within the same
      section normalize to the same text.
    - (FileReader.*): propagated from FileOpen.

  1. If LogicalNameIsArtifact(logical_name) is true,
       raise error "not a ROOT reference".

  2. If LogicalNameHasQualifier(logical_name) is true,
       raise error "has qualifier".

  3. Call LogicalNameToPath(logical_name) to obtain the file path.
     If LogicalNameToPath raises an error, propagate it.

  4. Call FileOpen(file_path) to open the file for reading.
     If FileOpen raises FileUnreadable or any propagated error,
       call FileClose if a reader was created, then propagate the error.

  5. Skip frontmatter:
     a. Call FileReadLine to read the first line.
        If end of file is reached immediately, proceed to step 6 with
        no lines (empty body).
     b. If the first line is exactly "---":
          Repeatedly call FileReadLine and discard lines until a line
          that is exactly "---" is read (the closing delimiter).
          If end of file is reached before finding the closing "---",
            call FileClose, then raise error
            "unexpected content before first heading".
     c. If the first line is not "---":
          Treat it as the first body line (do not discard it); proceed
          with this line as the next line to classify in step 6.

  6. Parse the body into sections:

     Initialize:
       - reader state from step 4/5 (positioned after frontmatter)
       - pending_line: the un-consumed body line from step 5c (if any),
         otherwise absent
       - name_section: absent
       - public_section: absent
       - agent_section: absent
       - private_sections: empty list
       - current_section: absent
       - current_subsection: absent
       - in_code_fence: false
       - fence_char: absent
       - fence_length: 0

     Repeat until end of file:

       a. Obtain the next line:
            If pending_line is present, use it and clear pending_line.
            Otherwise call FileReadLine; if end of file, exit the loop.

       b. Check for code fence transitions:
            If in_code_fence is false:
              If the line matches a fenced code block opening
              (three or more consecutive backtick characters, or three
              or more consecutive tilde characters, optionally followed
              by a language tag and nothing else on the line before):
                Set in_code_fence to true.
                Set fence_char to the opening character (` or ~).
                Set fence_length to the count of leading fence characters.
                Treat the line as content (go to step f).
              Otherwise, continue to step c.
            If in_code_fence is true:
              If the line is a valid closing fence (consists of at least
              fence_length consecutive fence_char characters, with only
              optional trailing whitespace):
                Set in_code_fence to false.
                Clear fence_char and fence_length.
              Treat the line as content (go to step f) regardless of
              whether it closed the fence.

       c. (Only reached when in_code_fence is false.)
          Attempt to parse the line as an ATX heading:
            An ATX heading line starts with one or more "#" characters
            immediately followed by at least one space character.
            The heading level equals the number of leading "#" characters.
            The heading text is everything after the "# " prefix,
            trimmed of leading and trailing whitespace.
            If a closing "#" sequence is present (preceded by at least
            one space), strip it and trim again.
            Lines like "#Foo" (no space after "#") are NOT headings.
            If the line is not an ATX heading, treat as content (go to step f).

       d. If the line is a level-1 heading:
            Finalize current_subsection into current_section (if any).
            Finalize current_section into the appropriate slot (if any)
              using ClassifyAndStoreSection (see below).
            raw_heading := the original line.
            heading := NormalizeText(extracted heading text).
            Start a new section with raw_heading and heading,
              empty content, and empty subsections list.
            Set current_section to this new section.
            Set current_subsection to absent.
            Continue to next iteration (do not add heading line to content).

       e. If the line is a level-2 heading and current_section is present:
            Finalize current_subsection into current_section (if any).
            raw_heading := the original line.
            heading := NormalizeText(extracted heading text).
            Check that no existing subsection in current_section.subsections
              has the same heading; if so, raise error "duplicate subsection"
              (call FileClose first).
            Start a new subsection with raw_heading, heading, and empty content.
            Set current_subsection to this new subsection.
            Continue to next iteration (do not add heading line to content).

       f. Add the line to content:
            If current_subsection is present:
              Append the line to current_subsection.content.
            Else if current_section is present:
              Append the line to current_section.content.
            Else:
              If the line is not blank (not empty and not all whitespace):
                Call FileClose, then raise error
                "unexpected content before first heading".
              (Blank lines before the first heading are silently discarded.)

     After the loop:
       Finalize current_subsection into current_section (if any).
       Finalize current_section into the appropriate slot (if any)
         using ClassifyAndStoreSection.

  7. Validate the name section:
     If name_section is absent,
       call FileClose, then raise error
       "unexpected content before first heading" (no level-1 heading found).
     Compare name_section.heading with NormalizeText(logical_name):
       If they do not match,
         call FileClose, then raise error "node name does not match".

  8. Call FileClose.

  9. Return a Node record:
       name_section: name_section
       public:       public_section (may be absent)
       agent:        agent_section  (may be absent)
       private:      private_sections (in file order)


---

## Helper procedures

procedure ClassifyAndStoreSection(section: NodeSection)
  -- Assigns the section to name_section, public_section, agent_section,
  -- or private_sections based on its normalized heading.
  -- Raises errors for duplicate public/agent sections.
  -- This procedure is called immediately before starting a new section.

  1. If name_section is absent:
       Set name_section to section.
       Return.

  2. If section.heading equals "public":
       If public_section is already present:
         Call FileClose, then raise error "duplicate public section".
       Set public_section to section.
       Return.

  3. If section.heading equals "agent":
       If agent_section is already present:
         Call FileClose, then raise error "duplicate agent section".
       Set agent_section to section.
       Return.

  4. Otherwise:
       Append section to private_sections.


procedure FinalizeSubsection(current_subsection, current_section)
  -- Moves current_subsection into current_section.subsections if present.

  1. If current_subsection is absent, return.
  2. Append current_subsection to current_section.subsections.
  3. Set current_subsection to absent.


---

## ATX Heading parsing detail

function ParseAtxHeading(line: string) -> optional record (level: integer, text: string, raw: string)

  1. Count the number of leading "#" characters. Call this count level.
     If level is 0, return absent (not a heading).

  2. If the character immediately after the "#" characters is not a space,
     return absent (e.g., "#Foo" is not a heading).

  3. Extract the heading text: everything after the first space following
     the leading "#" characters, trimmed of leading and trailing whitespace.

  4. Check for an optional closing "#" sequence:
     If the trimmed text ends with one or more "#" characters preceded by
     at least one space, strip the trailing "#" sequence and trim again.

  5. Return a record with:
       level: count of leading "#" characters
       text:  the extracted and cleaned heading text (before NormalizeText)
       raw:   the original line unchanged


---

## Fenced code block detection detail

function IsFenceOpening(line: string) -> optional record (char: character, length: integer)

  1. Count the number of leading backtick (`) characters. If 3 or more,
     and no backtick appears in the rest of the line, return
     record (char: "`", length: count).

  2. Count the number of leading tilde (~) characters. If 3 or more,
     return record (char: "~", length: count).

  3. Otherwise return absent.

function IsFenceClosing(line: string, fence_char: character, fence_length: integer) -> boolean

  1. Trim trailing whitespace from the line.
  2. If every character in the trimmed line equals fence_char,
     and the length of the trimmed line is at least fence_length,
     return true.
  3. Otherwise return false.
