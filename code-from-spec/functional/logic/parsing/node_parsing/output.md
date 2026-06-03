<!-- code-from-spec: ROOT/functional/logic/parsing/node_parsing@HXvIqhI-KPH4zXQyPw0qj9uWh_4 -->

## Namespace

    namespace: parsenode

## Records

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
  private: list of NodeSection

## Functions

function NodeParse(logical_name: string) -> Node
  errors:
    - NotARootReference: the logical name does not start with ROOT/.
    - HasQualifier: the logical name contains a parenthetical qualifier.
    - FileUnreadable: the file cannot be opened or read.
    - UnexpectedContentBeforeFirstHeading: file body has non-blank content
      before the first level-1 heading, or has no level-1 heading at all.
      Blank lines before the first heading are not an error.
    - NodeNameDoesNotMatch: the first heading does not match the logical
      name after normalization.
    - DuplicatePublicSection: more than one Public section exists.
    - DuplicateAgentSection: more than one Agent section exists.
    - DuplicateSubsection: two level-2 headings within the same section
      normalize to the same text.
    - (FileReader.*): propagated from FileOpen.

  1. If LogicalNameIsArtifact(logical_name) is true, raise "not a ROOT reference".

  2. If LogicalNameHasQualifier(logical_name) is true, raise "has qualifier".

  3. Call LogicalNameToPath(logical_name) to get the file path.

  4. Call FileOpen(file_path).
     If it fails, raise "file unreadable".
     Store result as reader.

  5. Wrap all remaining steps so that FileClose(reader) is called
     when done, whether parsing succeeds or fails.

  6. Skip frontmatter:
     Read the first line from reader.
     If it is exactly "---":
       Read and discard lines until a line that is exactly "---" is found.
       If end of file is reached before finding the closing "---",
         raise "unexpected content before first heading".
     Else if the first line is not blank:
       Put the line back into consideration for parsing (treat it as the
       first body line).

  7. Parse the body into sections:

     Initialize:
       - current_section = absent
       - sections = empty list
       - found_name_section = false
       - public_count = 0
       - agent_count = 0
       - in_fence = false
       - fence_char = absent
       - fence_length = 0

     For each line read from reader (until EndOfFile):

       a. Fenced code block tracking:
          If in_fence is false:
            If line matches a fence opener (three or more consecutive
            backtick or tilde characters, optionally followed by a
            language tag):
              Set in_fence = true.
              Set fence_char to the opening character (backtick or tilde).
              Set fence_length to the number of leading fence characters.
              Treat line as content (append to current content).
              Continue to next line.
          If in_fence is true:
            If line consists of at least fence_length of fence_char
            (optionally followed by whitespace only):
              Set in_fence = false.
            Treat line as content (append to current content).
            Continue to next line.

       b. Heading recognition (only when in_fence is false):
          Count leading "#" characters on the line.
          If count >= 1 and the character immediately after the "#"s is a space:
            heading_level = count of leading "#" characters.
            heading_text = everything after the leading "#" characters and the space.
            Trim leading and trailing whitespace from heading_text.
            If heading_text ends with one or more "#" preceded by a space:
              Strip the trailing whitespace and closing "#" sequence.
              Trim trailing whitespace again.
            normalized = NormalizeText(heading_text).
          Else:
            Not a heading — treat line as content.

       c. If line is a level-1 heading:
          Finalize current_section (push to sections list if present).
          Create new_section with:
            raw_heading = the original line.
            heading = normalized.
            content = empty list.
            subsections = empty list.
          Set current_section = new_section.

          If found_name_section is false:
            Set found_name_section = true.
            expected = NormalizeText(logical_name).
            If normalized != expected, raise "node name does not match".
            Mark new_section as name_section.
          Else if normalized == "public":
            Increment public_count.
            If public_count > 1, raise "duplicate public section".
            Mark new_section as public_section.
          Else if normalized == "agent":
            Increment agent_count.
            If agent_count > 1, raise "duplicate agent section".
            Mark new_section as agent_section.
          Else:
            Mark new_section as private_section.

       d. If line is a level-2 heading:
          If current_section is absent:
            Treat line as content before first heading (see step e).
          Else:
            new_subsection with:
              raw_heading = original line.
              heading = normalized.
              content = empty list.
            Check if current_section.subsections already has an entry
            with the same normalized heading.
            If so, raise "duplicate subsection".
            Append new_subsection to current_section.subsections.
            Set current_subsection = new_subsection.

       e. If line is not a heading (or level 3+), treat as content:
          If current_section is absent:
            If line is not blank, raise "unexpected content before first heading".
            Else discard line (blank lines before first heading are allowed).
          Else if current_section has a current_subsection:
            Append line to current_subsection.content.
          Else:
            Append line to current_section.content.

  8. After reading all lines:
     Finalize current_section (push to sections if present).
     If found_name_section is false, raise "unexpected content before first heading".

  9. Assemble Node from collected sections:
     - name_section: the section marked as name_section.
     - public: the section marked as public_section, or absent.
     - agent: the section marked as agent_section, or absent.
     - private: list of sections marked as private_section, in order.

  10. Return Node.
