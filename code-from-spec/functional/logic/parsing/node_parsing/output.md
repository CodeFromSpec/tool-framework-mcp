<!-- code-from-spec: ROOT/functional/logic/parsing/node_parsing@_Dn_gWFWBQEuUMEUCNntkUxpLY0 -->

# NodeParse

## Data Structures

record NodeSubsection
  heading: string          (normalized via NormalizeText)
  raw_heading: string      (original line as read from file)
  content: list of string  (lines as returned by FileReadLine)

record NodeSection
  heading: string          (normalized via NormalizeText)
  raw_heading: string      (original line as read from file)
  content: list of string  (lines before the first ## heading)
  subsections: list of NodeSubsection

record Node
  name_section: NodeSection
  public: optional NodeSection
  agent: optional NodeSection
  private: list of NodeSection


## Functions

---

function NodeParse(logical_name: string) -> Node

  errors:
    - NotARootReference: logical_name does not start with ROOT/
    - HasQualifier: logical_name contains a parenthetical qualifier
    - FileUnreadable: the file cannot be opened or read
    - UnexpectedContentBeforeFirstHeading: non-blank content before the
      first level-1 heading, or no level-1 heading found, or frontmatter
      is malformed (no closing ---)
    - NodeNameDoesNotMatch: first heading text does not match logical_name
      after normalization
    - DuplicatePublicSection: more than one # Public heading found
    - DuplicateAgentSection: more than one # Agent heading found
    - DuplicateSubsection: two ## headings within the same section
      normalize to the same text
    - (FileReader.*): propagated from FileOpen

  Steps:

  1. If LogicalNameIsArtifact(logical_name) is true,
       raise error "not a ROOT reference".

  2. If LogicalNameHasQualifier(logical_name) is true,
       raise error "has qualifier".

  3. Call LogicalNameToPath(logical_name) to get file_path.
     Propagate any errors from LogicalNameToPath.

  4. Call FileOpen(file_path) to get reader.
     If FileUnreadable is raised, propagate it.
     Propagate any other FileReader errors.

  5. Skip frontmatter:
     a. Read the first line using FileReadLine.
        If EndOfFile is raised, close reader and
        raise "unexpected content before first heading".
     b. If the first line is exactly "---":
          Read lines one by one using FileReadLine.
          For each line:
            If it is exactly "---", stop — frontmatter skipped.
            If EndOfFile is raised, close reader and
            raise "unexpected content before first heading".
        Else:
          The first line is the first body line; hold it for step 6.

  6. Parse the body into sections.
     Maintain the following state:
       - sections_seen: record tracking whether name_section,
         public, and agent have been encountered (all start false)
       - current_section: optional — the NodeSection being built
         (absent until the first # heading is encountered)
       - current_subsection: optional — the NodeSubsection being
         built (absent until the first ## heading within a section)
       - private_sections: ordered list of NodeSections
       - in_fence: boolean, false initially
       - fence_char: string ("`" or "~"), set when in_fence becomes true
       - fence_length: integer, the number of fence characters on the
         opening line

     Process lines one at a time (starting from the held line in step 5,
     if any, then reading subsequent lines via FileReadLine until
     EndOfFile):

     For each line:

       a. Fence tracking (check before heading recognition):
          If in_fence is false:
            Count the number of leading backtick characters (`) in line.
            If count >= 3 and the rest of the line (after those backticks)
            has no backtick characters:
              Set in_fence = true, fence_char = "`", fence_length = count.
              Treat line as content (do not interpret as heading).
            Else:
              Count the number of leading tilde characters (~) in line.
              If count >= 3 and the rest of the line (after those tildes)
              has no tilde characters:
                Set in_fence = true, fence_char = "~", fence_length = count.
                Treat line as content (do not interpret as heading).
          Else (in_fence is true):
            Count the number of leading fence_char characters in line.
            If count >= fence_length and the rest of the line contains
            only whitespace (or is empty):
              Set in_fence = false.
            Treat line as content regardless.

       b. If in_fence is true (set in step a or already true before step a),
          treat line as content — go to step f (append to current content).

       c. Attempt to parse line as an ATX heading:
          - Count leading "#" characters; call this level.
          - If level == 0, line is not a heading — go to step f.
          - If the character after the leading "#" characters is not a space,
            line is not a heading — go to step f.
          - Extract heading_text: everything after the "# " prefix (level
            hashes + one space), trimmed of leading and trailing whitespace.
          - Strip optional closing "#" sequence: if heading_text ends with
            one or more "#" characters preceded by at least one space,
            remove those trailing "#" characters and trim again.
          - heading_text is now the extracted heading text.

       d. If level == 1 (a new section begins):
          - Finalize current state:
            If current_subsection is not absent:
              Append current_subsection to current_section.subsections.
              Set current_subsection to absent.
            If current_section is not absent:
              Store current_section in its appropriate place
              (see "storing a section" below).
          - Normalize heading_text with NormalizeText to get norm.
          - Compute expected_norm = NormalizeText(logical_name).
          - If name_section has not yet been seen:
              If norm != expected_norm, raise "node name does not match".
              Create a new NodeSection with heading = norm,
              raw_heading = line, content = [], subsections = [].
              Set current_section to this new section.
              Mark name_section as seen.
          - Else if norm == "public":
              If public has already been seen, raise "duplicate public section".
              Create a new NodeSection with heading = norm,
              raw_heading = line, content = [], subsections = [].
              Set current_section to this new section.
              Mark public as seen.
          - Else if norm == "agent":
              If agent has already been seen, raise "duplicate agent section".
              Create a new NodeSection with heading = norm,
              raw_heading = line, content = [], subsections = [].
              Set current_section to this new section.
              Mark agent as seen.
          - Else (private section):
              Create a new NodeSection with heading = norm,
              raw_heading = line, content = [], subsections = [].
              Set current_section to this new section.
          - Set current_subsection to absent.
          - Do not append line to any content list.

       e. If level == 2 (a new subsection begins within the current section):
          - If current_section is absent:
              Treat line as content that appears before any section;
              check if it is non-blank — if so, raise
              "unexpected content before first heading".
              If blank, discard.
              (In practice level-2 headings before any # heading are treated
              as non-blank content and trigger the error.)
          - Normalize heading_text with NormalizeText to get sub_norm.
          - Check if any existing subsection in current_section.subsections
            already has heading == sub_norm. If so, raise "duplicate subsection".
          - Also check if current_subsection (if not absent) has
            heading == sub_norm. If so, raise "duplicate subsection".
          - Finalize current_subsection:
            If current_subsection is not absent:
              Append current_subsection to current_section.subsections.
          - Create a new NodeSubsection with heading = sub_norm,
            raw_heading = line, content = [].
          - Set current_subsection to this new subsection.
          - Do not append line to any content list.

       f. Content line (level == 0, or level >= 3, or inside fence):
          - If current_section is absent:
              If line is blank, discard it.
              Else raise "unexpected content before first heading".
          - Else if current_subsection is not absent:
              Append line to current_subsection.content.
          - Else:
              Append line to current_section.content.

  7. After all lines are processed (EndOfFile reached):
     Finalize remaining state:
     - If current_subsection is not absent:
         Append current_subsection to current_section.subsections.
     - If current_section is not absent:
         Store current_section (see "storing a section" below).
     - If name_section was never seen (file had no level-1 heading,
       or only blank lines):
         Close reader and raise "unexpected content before first heading".

  8. Call FileClose(reader).
     (FileClose must be called in all cases — see error handling note.)

  9. Return Node with:
       name_section: the collected name section
       public: the collected public section (absent if none seen)
       agent: the collected agent section (absent if none seen)
       private: private_sections list in order of appearance


  ### Storing a section

  When finalizing current_section before starting a new section or at EOF:

  - If current_section.heading matches the name_section heading
    (i.e., this is the name section):
      Store it as name_section.
  - Else if current_section.heading == "public":
      Store it as the public section.
  - Else if current_section.heading == "agent":
      Store it as the agent section.
  - Else:
      Append it to private_sections.


  ### Error handling

  If any error is raised during steps 5–7, call FileClose(reader)
  before propagating the error. FileClose must be called in all
  exit paths — success or failure.


  ### Fence opening rule (detail)

  A fence opener is a line consisting of three or more consecutive
  backtick (`) characters or three or more consecutive tilde (~)
  characters, optionally followed by a language tag (any non-fence
  characters). The opening delimiter character must not appear in
  the language tag portion. The closer must use the same character
  as the opener and be at least as long as the opener.

  For the purposes of this function, the check is:
  - For backtick fences: line starts with 3 or more "`" and the
    remainder of the line contains no "`".
  - For tilde fences: line starts with 3 or more "~" and the
    remainder of the line contains no "~".
