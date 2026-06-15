<!-- code-from-spec: SPEC/functional/logic/parsing/node_parsing@LSRGDOT3oscqosSBbn8E0ila5aA -->

namespace: parsenode

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


function NodeParse(logical_name: string) -> Node
  errors:
    - NotASpecReference
    - HasQualifier
    - FileUnreadable
    - UnexpectedContentBeforeFirstHeading
    - NodeNameDoesNotMatch
    - DuplicatePublicSection
    - DuplicateAgentSection
    - DuplicatePrivateSection
    - UnrecognizedSection
    - DuplicateSubsection
    - (FileReader.*)

  1. If LogicalNameIsSpec(logical_name) is false,
       raise error "not a SPEC reference".

  2. If LogicalNameHasQualifier(logical_name) is true,
       raise error "has qualifier".

  3. Let cfs_path = LogicalNameToPath(logical_name).

  4. Let reader = FileOpen(cfs_path).
       If FileOpen raises FileUnreadable or any PathUtils error,
         raise error "file unreadable".

  5. Skip frontmatter:
       Let first_line = FileReadLine(reader).
         If first_line raises EndOfFile, go to step 6 with empty body.
       If first_line is exactly "---":
         Loop:
           Let line = FileReadLine(reader).
           If line raises EndOfFile,
             call FileClose(reader),
             raise error "unexpected content before first heading".
           If line is exactly "---", stop the loop.
       Else:
         Treat first_line as the first body line (do not discard it).

  6. Parse the body into sections:
       Let name_section = absent.
       Let public_section = absent.
       Let agent_section = absent.
       Let private_section = absent.
       Let current_section = absent.
       Let current_subsection = absent.
       Let in_fence = false.
       Let fence_char = absent.
       Let fence_width = 0.

       For each line from the file (starting after frontmatter),
       and including first_line if it was not the frontmatter marker:

         a. Fenced code block tracking:
              Let stripped = line with leading/trailing whitespace removed.
              If in_fence is false:
                If stripped starts with "```" or "~~~":
                  Count leading backtick or tilde characters.
                  Let fence_char = that character.
                  Let fence_width = that count.
                  Set in_fence = true.
                  Append line to current content (see below).
                  Continue to next line.
              Else (in_fence is true):
                Check if stripped consists entirely of fence_char characters
                and its length >= fence_width.
                  If so, set in_fence = false, fence_char = absent, fence_width = 0.
                Append line to current content (see below).
                Continue to next line.

         b. Heading recognition (only when in_fence is false):
              If line matches the ATX heading pattern:
                Count leading "#" characters. Let level = that count.
                Let text_part = everything after the leading "# " (hashes + one space).
                Trim text_part of leading and trailing whitespace.
                If text_part ends with one or more "#" characters preceded by at least one space:
                  Strip the trailing "#" sequence and any preceding whitespace.
                Let raw_heading = the original line.
                Let heading = NormalizeText(text_part).

                If level = 1:
                  Finalize current_subsection into current_section if present.
                  Finalize current_section into the result record if present.
                  Classify heading:
                    If name_section is absent:
                      Let expected = NormalizeText(logical_name).
                      If heading != expected,
                        call FileClose(reader),
                        raise error "node name does not match".
                      Start name_section with raw_heading = raw_heading, heading = heading,
                        content = empty list, subsections = empty list.
                      Set current_section = name_section.
                    Else if heading = "public":
                      If public_section is not absent,
                        call FileClose(reader),
                        raise error "duplicate public section".
                      Start a new section: heading = heading, raw_heading = raw_heading,
                        content = empty list, subsections = empty list.
                      Set public_section = that section, current_section = that section.
                    Else if heading = "agent":
                      If agent_section is not absent,
                        call FileClose(reader),
                        raise error "duplicate agent section".
                      Start a new section: heading = heading, raw_heading = raw_heading,
                        content = empty list, subsections = empty list.
                      Set agent_section = that section, current_section = that section.
                    Else if heading = "private":
                      If private_section is not absent,
                        call FileClose(reader),
                        raise error "duplicate private section".
                      Start a new section: heading = heading, raw_heading = raw_heading,
                        content = empty list, subsections = empty list.
                      Set private_section = that section, current_section = that section.
                    Else:
                      call FileClose(reader),
                      raise error "unrecognized section".

                Else if level = 2:
                  If current_section is absent:
                    Treat line as content (no active section yet).
                    Continue to next line.
                  Finalize current_subsection into current_section if present.
                  Check if any existing subsection in current_section.subsections
                    has heading = heading.
                    If so,
                      call FileClose(reader),
                      raise error "duplicate subsection".
                  Start a new subsection: heading = heading, raw_heading = raw_heading,
                    content = empty list.
                  Set current_subsection = that subsection.

                Else (level >= 3):
                  Append line to current content (see below).

              Else (not a heading):
                Append line to current content (see below).

              Continue to next line.

         c. Appending to current content:
              If current_subsection is not absent:
                Append line to current_subsection.content.
              Else if current_section is not absent:
                Append line to current_section.content.
              Else:
                If line is not blank (contains non-whitespace characters):
                  call FileClose(reader),
                  raise error "unexpected content before first heading".
                (blank lines before the first heading are silently discarded)

       When EndOfFile is raised:
         Finalize current_subsection into current_section if present.
         Finalize current_section into the result record if present.

  7. If name_section is still absent (file had no level-1 heading or all
     content was blank before any heading):
       call FileClose(reader),
       raise error "unexpected content before first heading".

  8. Call FileClose(reader).

  9. Return a Node record:
       name_section = name_section
       public = public_section (absent if not found)
       agent = agent_section (absent if not found)
       private = private_section (absent if not found)


helper: ATX heading pattern
  A line matches if:
    It starts with one or more "#" characters,
    followed by at least one space character,
    followed by any remaining text (possibly empty after trimming).
  Lines starting with "#" not followed by a space do not match.
  Lines that are exactly one or more "#" characters with no following text do not match.

helper: finalize current_subsection into current_section
  Append current_subsection to current_section.subsections.
  Set current_subsection = absent.

helper: finalize current_section into the result record
  The section reference was already assigned to the appropriate slot
  (name_section, public_section, agent_section, or private_section)
  when the section started, so no additional assignment is needed.
  Set current_section = absent.
