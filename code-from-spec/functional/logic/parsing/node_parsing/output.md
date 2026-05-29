<!-- code-from-spec: ROOT/functional/logic/parsing/node_parsing@WR6sFtqVI0Vmu8PElvEi_xj9us4 -->

# NodeParse

## Records

record NodeSubsection
  heading: string       -- normalized heading text (via NormalizeText)
  raw_heading: string   -- original line as read from file
  content: list of string

record NodeSection
  heading: string       -- normalized heading text (via NormalizeText)
  raw_heading: string   -- original line as read from file
  content: list of string
  subsections: list of NodeSubsection

record Node
  name_section: NodeSection
  public: optional NodeSection
  agent: optional NodeSection
  private: list of NodeSection

---

## function NodeParse(logical_name: string) -> Node

  Errors:
    - "not a ROOT reference": logical name does not start with ROOT/.
    - "has qualifier": logical name contains a parenthetical qualifier.
    - (path errors): propagated from LogicalNameToPath or FileOpen.
    - "file unreadable": file cannot be opened or read.
    - "unexpected content before first heading": non-blank content
      before the first level-1 heading, or no level-1 heading at all.
    - "node name does not match": first heading text does not match
      logical name after normalization.
    - "duplicate public section": more than one `# Public` heading found.
    - "duplicate agent section": more than one `# Agent` heading found.
    - "duplicate subsection": two `##` headings within the same section
      normalize to the same text.

  Steps:

  1. If LogicalNameIsArtifact(logical_name) is true,
       raise error "not a ROOT reference".

  2. If LogicalNameHasQualifier(logical_name) is true,
       raise error "has qualifier".

  3. Let cfs_path = LogicalNameToPath(logical_name).
     If LogicalNameToPath raises an error, propagate it.

  4. Let reader = FileOpen(cfs_path).
     If FileOpen raises an error, propagate it.

  5. Let close_and_raise be a helper that calls FileClose(reader)
     then raises the given error. Use this in all error paths below.

  6. Skip frontmatter:
     a. Read the first line from reader using FileReadLine.
        If "end of file" is raised, call close_and_raise
        "unexpected content before first heading".
     b. If the first line equals exactly "---":
          i. Read lines using FileReadLine until a line equals
             exactly "---". Discard all lines read.
         ii. If "end of file" is raised before finding the
             closing "---", call close_and_raise
             "unexpected content before first heading".
     c. Else (first line is not "---"):
          If the first line is blank, store it as the pending
          first non-frontmatter line. Otherwise store it as the
          pending first non-frontmatter line for use in step 7.

  7. Parse the body into sections:

     Initialize:
       - current_section: absent
       - current_subsection: absent
       - sections: empty list of NodeSection
         (each entry tagged with its role: name, public, agent, or private)
       - has_public: false
       - has_agent: false
       - found_first_heading: false
       - in_fence: false
       - fence_char: absent
       - fence_length: 0
       - pending_line: the line saved in step 6c, or absent

     Define a helper IsBlank(line):
       Returns true if the line contains only whitespace characters.

     Define a helper ParseHeading(line) -> record or absent:
       a. Count leading "#" characters; call this level.
          If level is 0, return absent.
       b. If the character immediately after the leading "#" sequence
          is not a space character, return absent (e.g. "#Foo" is not
          a heading).
       c. Let text = everything after the leading "<level># " prefix.
          Trim leading and trailing whitespace from text.
       d. If text ends with one or more "#" characters preceded by
          at least one space, strip that trailing sequence and trim again.
       e. Let normalized = NormalizeText(text).
       f. Return record with fields: level, text (raw extracted text),
          normalized, raw_line (the original line unchanged).

     Define a helper AppendContent(line):
       If current_subsection is present,
         append line to current_subsection.content.
       Else if current_section is present,
         append line to current_section.content.
       -- Lines before any heading are handled separately (see below).

     Define a helper OpenSubsection(raw_line, normalized):
       a. If current_subsection is present, finalize it:
            append current_subsection to current_section.subsections.
       b. Check current_section.subsections for any existing subsection
          whose heading equals normalized.
          If found, call close_and_raise "duplicate subsection".
       c. Set current_subsection = new NodeSubsection with
            heading = normalized,
            raw_heading = raw_line,
            content = empty list.

     Define a helper FinalizeSubsection:
       If current_subsection is present,
         append current_subsection to current_section.subsections.
       Set current_subsection = absent.

     Define a helper FinalizeSection:
       Call FinalizeSubsection.
       If current_section is present,
         append current_section to sections.
       Set current_section = absent.

     Define a helper OpenSection(raw_line, normalized):
       a. Call FinalizeSection to close any open section.
       b. Let new_section = NodeSection with
            heading = normalized,
            raw_heading = raw_line,
            content = empty list,
            subsections = empty list.
       c. Set current_section = new_section.

     Now process lines:
       Let pre_heading_lines = empty list.
         (Accumulates blank lines seen before the first level-1 heading.)

     Loop:
       a. If pending_line is present, let line = pending_line,
            clear pending_line. Otherwise let line = FileReadLine.
          If FileReadLine raises "end of file", go to step 8.

       b. Fence tracking (fenced code block detection):
          If in_fence is false:
            i.  Count leading backtick characters in line; call it bt.
                If bt >= 3 and the rest of the line (after the backticks)
                contains no backtick characters (allowing a language tag
                with no embedded backticks), set in_fence = true,
                fence_char = "`", fence_length = bt. Treat line as content.
            ii. Else count leading tilde characters; call it ti.
                If ti >= 3, set in_fence = true, fence_char = "~",
                fence_length = ti. Treat line as content.
            iii. If neither, continue to heading detection below.
          If in_fence is true:
            i.  Count leading characters equal to fence_char; call it cl.
                If cl >= fence_length and the remainder of the line
                (after those characters) contains no fence_char characters,
                set in_fence = false. Treat line as content.
            ii. Else treat line as content.
          For content lines (in_fence path), call AppendContent(line)
          and go to next iteration.

       c. If in_fence is false, attempt heading detection:
          Let h = ParseHeading(line).
          If h is absent, treat line as content:
            If found_first_heading is false,
              if IsBlank(line), append line to pre_heading_lines and continue.
              else call close_and_raise
                "unexpected content before first heading".
            Else call AppendContent(line) and continue.

       d. h is a valid heading. Process by level:

          If h.level == 1:
            If found_first_heading is false:
              -- This is the name section heading.
              Let expected = NormalizeText(logical_name).
              If h.normalized does not equal expected,
                call close_and_raise "node name does not match".
              Set found_first_heading = true.
              Call OpenSection(h.raw_line, h.normalized).
              -- Discard pre_heading_lines (blank lines before first heading
                 are not added to any section content).
            Else:
              -- Subsequent level-1 heading.
              Call OpenSection(h.raw_line, h.normalized).
              -- Classify the section:
              If h.normalized equals "public":
                If has_public is true,
                  call close_and_raise "duplicate public section".
                Set has_public = true.
              Else if h.normalized equals "agent":
                If has_agent is true,
                  call close_and_raise "duplicate agent section".
                Set has_agent = true.

          If h.level == 2:
            If found_first_heading is false,
              call close_and_raise "unexpected content before first heading".
            If current_section is absent,
              call close_and_raise "unexpected content before first heading".
            Call OpenSubsection(h.raw_line, h.normalized).

          If h.level >= 3:
            -- Deep headings are always content.
            If found_first_heading is false,
              call close_and_raise "unexpected content before first heading".
            Call AppendContent(line).

       Go to next iteration.

  8. End of file reached. Call FinalizeSection.

     If found_first_heading is false,
       raise error "unexpected content before first heading".
       (FileClose was already called via FinalizeSection path — but since
        we never opened a section, call FileClose(reader) first.)

  9. Call FileClose(reader).

  10. Assemble the Node record:
      a. The first entry in sections is the name_section.
      b. For each remaining entry in sections:
           If its heading equals "public", assign it to public.
           Else if its heading equals "agent", assign it to agent.
           Else append it to private.
      c. If no public section was found, public = absent.
         If no agent section was found, agent = absent.
      d. Return Node with name_section, public, agent, private
         (private preserves file order).
