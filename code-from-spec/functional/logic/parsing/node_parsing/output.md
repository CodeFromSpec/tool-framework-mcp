<!-- code-from-spec: ROOT/functional/logic/parsing/node_parsing@36TexO1_40jx92Ezlxt7YkdcPPU -->

namespace: parsenode

---

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
  private: optional NodeSection

---

function NodeParse(logical_name: string) -> Node

  1. If LogicalNameIsSpec(logical_name) is false,
     raise error "not a SPEC reference".

  2. If LogicalNameHasQualifier(logical_name) is true,
     raise error "has qualifier".

  3. Resolve cfs_path = LogicalNameToPath(logical_name).

  4. Open reader = FileOpen(cfs_path).
     If FileOpen raises FileUnreadable or any PathUtils error,
     raise error "file unreadable".

  5. Skip frontmatter:
     Read the first line with FileReadLine.
     If end of file, proceed to step 6 (empty body).
     If the first line is exactly "---":
       Read lines one at a time with FileReadLine.
       If end of file is reached before finding a line
       that is exactly "---", close reader with FileClose,
       raise error "unexpected content before first heading".
       When a line that is exactly "---" is found, frontmatter
       is consumed; proceed to step 6.
     Else the first line is not "---":
       Use this line as the first body line in step 6.

  6. Parse the body into sections.
     Initialize:
       sections_seen = empty record mapping string -> boolean
         (tracks: "public", "agent", "private")
       current_section = absent
       current_subsection = absent
       result_name_section = absent
       result_public = absent
       result_agent = absent
       result_private = absent
       inside_fence = false
       fence_char = absent
       fence_length = 0

     If a leftover first body line exists from step 5,
     process it before reading more lines.

     Repeat (reading lines with FileReadLine until EndOfFile):
       line = next line (leftover or FileReadLine)
       If EndOfFile is raised, finalize the current open
       section and subsection and exit the loop.

       Check fenced code block state:
         If inside_fence is false:
           If line matches a fence-open pattern
           (3 or more consecutive "`" or "~" characters,
           optionally followed by a language tag):
             Set inside_fence = true.
             Set fence_char to the opening character ("`" or "~").
             Set fence_length to the count of those characters.
             Append line to the appropriate content list
             (current_subsection content if present, else
             current_section content if present, else this
             line appears before any heading — treat as below).
             Continue to next line.
         Else (inside_fence is true):
           If line matches a fence-close pattern
           (at least fence_length consecutive fence_char
           characters, nothing else or only whitespace after):
             Set inside_fence = false.
             fence_char = absent, fence_length = 0.
           Append line to the appropriate content list.
           Continue to next line (do not parse as heading).

       Detect ATX heading:
         If line starts with one or more "#" characters
         followed by at least one space:
           Count leading "#" characters → heading_level.
           Extract heading_text = everything after the
           leading "#" characters and the space.
           Trim leading and trailing whitespace from heading_text.
           If heading_text ends with one or more "#" characters
           preceded by at least one space, strip that trailing
           sequence and re-trim.
           normalized = NormalizeText(heading_text).
           raw_heading_line = line (the original line as read).
         Else:
           The line is content (not a heading).
           Append line to current content target:
             If current_subsection is present, append to
             current_subsection.content.
             Else if current_section is present, append to
             current_section.content.
             Else:
               If line is not blank, close reader with FileClose,
               raise error "unexpected content before first heading".
               (Blank lines before the first heading are ignored.)
           Continue to next line.

       Handle heading_level = 1:
         Finalize current_subsection if present:
           Append current_subsection to current_section.subsections.
           Set current_subsection = absent.
         Finalize current_section if present:
           Store current_section in the appropriate result field.
         Set current_subsection = absent.

         If result_name_section is absent:
           The normalized logical name = NormalizeText(logical_name).
           If normalized != NormalizeText(logical_name):
             — (Already normalized; compare normalized heading text
               to normalized logical name.)
           If normalized (the heading's normalized text) !=
           NormalizeText(logical_name):
             Close reader with FileClose.
             raise error "node name does not match".
           Create new NodeSection with:
             heading = normalized
             raw_heading = raw_heading_line
             content = empty list
             subsections = empty list
           Set current_section = this new NodeSection.
           Set result_name_section = current_section.
         Else if normalized == "public":
           If "public" is in sections_seen, close reader with
           FileClose, raise error "duplicate public section".
           Mark sections_seen["public"] = true.
           Create new NodeSection with:
             heading = normalized
             raw_heading = raw_heading_line
             content = empty list
             subsections = empty list
           Set current_section = this new NodeSection.
           Set result_public = current_section.
         Else if normalized == "agent":
           If "agent" is in sections_seen, close reader with
           FileClose, raise error "duplicate agent section".
           Mark sections_seen["agent"] = true.
           Create new NodeSection with:
             heading = normalized
             raw_heading = raw_heading_line
             content = empty list
             subsections = empty list
           Set current_section = this new NodeSection.
           Set result_agent = current_section.
         Else if normalized == "private":
           If "private" is in sections_seen, close reader with
           FileClose, raise error "duplicate private section".
           Mark sections_seen["private"] = true.
           Create new NodeSection with:
             heading = normalized
             raw_heading = raw_heading_line
             content = empty list
             subsections = empty list
           Set current_section = this new NodeSection.
           Set result_private = current_section.
         Else:
           Close reader with FileClose.
           raise error "unrecognized section".

       Handle heading_level = 2:
         If current_section is absent:
           If normalized heading text is blank: treat as content.
           Else: close reader with FileClose,
           raise error "unexpected content before first heading".
         Finalize current_subsection if present:
           Append current_subsection to current_section.subsections.
         Check for duplicate subsection:
           For each existing subsection in current_section.subsections,
           if its heading == normalized, close reader with FileClose,
           raise error "duplicate subsection".
         Create new NodeSubsection with:
           heading = normalized
           raw_heading = raw_heading_line
           content = empty list
         Set current_subsection = this new NodeSubsection.

       Handle heading_level >= 3:
         Treat the line as content (not structural).
         Append line to current content target:
           If current_subsection is present, append to
           current_subsection.content.
           Else if current_section is present, append to
           current_section.content.
           Else: treat as pre-heading content (blank check applies).

     End of body loop.

  7. After loop ends, finalize open structures:
     If current_subsection is present:
       Append current_subsection to current_section.subsections.
     If current_section is present:
       This was already stored in the result field by reference
       at the time of creation, so no further action needed.
     If result_name_section is absent:
       Close reader with FileClose.
       raise error "unexpected content before first heading".

  8. Call FileClose(reader).

  9. Return Node with:
       name_section = result_name_section
       public = result_public (absent if no Public section found)
       agent = result_agent (absent if no Agent section found)
       private = result_private (absent if no Private section found)

---

Errors:
  - NotASpecReference: raised in step 1.
  - HasQualifier: raised in step 2.
  - FileUnreadable: raised in step 4.
  - UnexpectedContentBeforeFirstHeading: raised in step 5 or step 6
    when non-blank content appears before any level-1 heading,
    or no level-1 heading exists in the file at all.
  - NodeNameDoesNotMatch: raised in step 6 when the first level-1
    heading's normalized text does not match the normalized logical name.
  - DuplicatePublicSection: raised in step 6.
  - DuplicateAgentSection: raised in step 6.
  - DuplicatePrivateSection: raised in step 6.
  - UnrecognizedSection: raised in step 6.
  - DuplicateSubsection: raised in step 6 within level-2 handling.
  - FileReader.*: propagated from FileOpen in step 4.
