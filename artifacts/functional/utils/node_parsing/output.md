<!-- code-from-spec: ROOT/functional/utils/node_parsing@PENDING -->

## Data structures

```
record Subsection
  heading: string
  content: string

record Section
  heading: string
  content: string
  subsections: list of Subsection

record ParsedNode
  name_section: Section
  public: optional Section
  agent: optional Section
  private: list of Section
```

## Functions

### ParseNode(logical_name) -> ParsedNode

1. Resolve the file path from logical_name using logical_names.

2. Open the file using file_reader.

3. Skip the frontmatter block.
   If the file starts with "---", skip all lines until the next "---"
   (inclusive). The remaining lines are the body.

4. Set up a tracking flag for fenced code blocks (initially outside).
   Set current_section to empty. Set sections to an empty list.

5. For each line in the body:
   a. If the line starts with "```", toggle the fenced-code-block flag.
   b. If inside a fenced code block, append the line to the current
      section's content and continue to the next line.
   c. If the line starts with "# " (level-1 heading):
      - If current_section is not empty, finalize it and append to sections.
      - Start a new section with heading set to the text after "# ".
   d. If the line starts with "## " (level-2 heading):
      - Append the line as a new subsection of the current section.
        The subsection heading is the text after "## ".
   e. Otherwise (including level-3+ headings), append the line to the
      current section's content (or current subsection's content if one
      is active).

6. After all lines are processed, finalize the last section and append
   to sections.

7. If there is any non-blank content before the first level-1 heading,
   raise error "unexpected content before first heading".

8. Normalize the first section's heading using name_normalization.
   Normalize the logical_name the same way.
   If they do not match, raise error "node name does not match".

9. Set name_section to the first section.

10. Walk through the remaining sections. For each section:
    a. Normalize the heading.
    b. If it normalizes to "public":
       - If public is already set, raise error "duplicate public section".
       - Check all subsection headings within this section. Normalize each.
         If any two normalize to the same text,
         raise error "duplicate subsection".
       - Set public to this section.
    c. If it normalizes to "agent":
       - Set agent to this section.
    d. Otherwise:
       - Append to the private list.

11. For each section and subsection, trim leading and trailing blank
    lines from content.

12. Return the ParsedNode record with name_section, public, agent,
    and private.
