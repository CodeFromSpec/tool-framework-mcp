---
depends_on:
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/utils/logical_names
  - SPEC/golang/implementation/utils/text_normalization
output: internal/parsenode/parsenode.go
---

# SPEC/golang/implementation/parsing/node_parsing

Parses the body of a spec node file into a structured
representation of its sections and subsections.

# Public

## Package

`package parsenode`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v4/internal/parsenode"`

## Interface

```go
type NodeSubsection struct {
	Heading    string
	RawHeading string
	Content    []string
}

type NodeSection struct {
	Heading     string
	RawHeading  string
	Content     []string
	Subsections []*NodeSubsection
}

type Node struct {
	NameSection *NodeSection
	Public      *NodeSection // nil if absent
	Agent       *NodeSection // nil if absent
	Private     *NodeSection // nil if absent
}

func NodeParse(logicalName string) (*Node, error)
```

`heading` is the normalized form (after `NormalizeText`),
used for comparisons and lookups. `raw_heading` is the
original line as read from the file, preserved for
hashing.

### Errors

- `ErrNotASpecReference`
- `ErrHasQualifier`
- `ErrFileUnreadable`
- `ErrUnexpectedContentBeforeFirstHeading`
- `ErrNodeNameDoesNotMatch`
- `ErrDuplicatePublicSection`
- `ErrDuplicateAgentSection`
- `ErrDuplicatePrivateSection`
- `ErrUnrecognizedSection`
- `ErrDuplicateSubsection`
- Propagated errors from `file` package.

# Agent

Implement the node parsing component as a Go package.

## Logic

1. Call LogicalNameParse(logical_name).
   If it fails, raise error "not a SPEC reference".
   Let `ln` be the result.

2. If ln.Type is not NodeTypeSpec,
     raise error "not a SPEC reference".

3. If ln.Qualifier is not nil,
     raise error "has qualifier".

4. Let cfs_path = PathCfs{Value: ln.Path}.

5. Let reader = FileOpen(cfs_path, mode "read",
   timeout_ms 30000).
     If FileOpen raises FileUnreadable or any PathUtils
     error, raise error "file unreadable".

6. Skip frontmatter:
     Let first_line = FileReadLine(reader).
       If first_line raises EndOfFile, go to step 6
       with empty body.
     If first_line is exactly "---":
       Loop:
         Let line = FileReadLine(reader).
         If line raises EndOfFile,
           call FileClose(reader),
           raise error "unexpected content before
           first heading".
         If line is exactly "---", stop the loop.
     Else:
       Treat first_line as the first body line (do not
       discard it).

7. Parse the body into sections:
     Let name_section = absent.
     Let public_section = absent.
     Let agent_section = absent.
     Let private_section = absent.
     Let current_section = absent.
     Let current_subsection = absent.
     Let in_fence = false.
     Let fence_char = absent.
     Let fence_width = 0.

     For each line from the file (starting after
     frontmatter), and including first_line if it was
     not the frontmatter marker:

       a. Fenced code block tracking:
            Let stripped = line with leading/trailing
            whitespace removed.
            If in_fence is false:
              If stripped starts with "```" or "~~~":
                Count leading backtick or tilde chars.
                Let fence_char = that character.
                Let fence_width = that count.
                Set in_fence = true.
                Append line to current content (step c).
                Continue to next line.
            Else (in_fence is true):
              Check if stripped consists entirely of
              fence_char characters and its length >=
              fence_width.
                If so, set in_fence = false,
                fence_char = absent, fence_width = 0.
              Append line to current content (step c).
              Continue to next line.

       b. Heading recognition (only when in_fence is
          false):
            If line matches the ATX heading pattern:
              Count leading "#" characters. Let level =
              that count.
              Let text_part = everything after the
              leading "# " (hashes + one space).
              Trim text_part of leading and trailing
              whitespace.
              If text_part ends with one or more "#"
              characters preceded by at least one space:
                Strip the trailing "#" sequence and any
                preceding whitespace.
              Let raw_heading = the original line.
              Let heading = NormalizeText(text_part).

              If level = 1:
                Finalize current_subsection into
                current_section if present.
                Finalize current_section into the
                result record if present.
                Classify heading:
                  If name_section is absent:
                    Let expected =
                    NormalizeText(logical_name).
                    If heading != expected,
                      call FileClose(reader),
                      raise error "node name does not
                      match".
                    Start name_section.
                    Set current_section = name_section.
                  Else if heading = "public":
                    If public_section is not absent,
                      call FileClose(reader),
                      raise error "duplicate public
                      section".
                    Start new section.
                    Set public_section and
                    current_section.
                  Else if heading = "agent":
                    If agent_section is not absent,
                      call FileClose(reader),
                      raise error "duplicate agent
                      section".
                    Start new section.
                    Set agent_section and
                    current_section.
                  Else if heading = "private":
                    If private_section is not absent,
                      call FileClose(reader),
                      raise error "duplicate private
                      section".
                    Start new section.
                    Set private_section and
                    current_section.
                  Else:
                    call FileClose(reader),
                    raise error "unrecognized section".

              Else if level = 2:
                If current_section is absent:
                  Treat line as content.
                  Continue to next line.
                Finalize current_subsection into
                current_section if present.
                Check if any existing subsection in
                current_section.subsections has
                heading = heading.
                  If so,
                    call FileClose(reader),
                    raise error "duplicate subsection".
                Start a new subsection.
                Set current_subsection.

              Else (level >= 3):
                Append line to current content (step c).

            Else (not a heading):
              Append line to current content (step c).

       c. Appending to current content:
            If current_subsection is not absent:
              Append line to
              current_subsection.content.
            Else if current_section is not absent:
              Append line to
              current_section.content.
            Else:
              If line is not blank (contains
              non-whitespace characters):
                call FileClose(reader),
                raise error "unexpected content before
                first heading".

     When EndOfFile is raised:
       Finalize current_subsection into
       current_section if present.
       Finalize current_section into the result
       record if present.

8. If name_section is still absent:
     call FileClose(reader),
     raise error "unexpected content before first
     heading".

9. Call FileClose(reader).

10. Return Node record with name_section, public,
   agent, private.

### ATX heading pattern

A line matches if it starts with one or more "#"
characters, followed by at least one space character,
followed by any remaining text. Lines starting with "#"
not followed by a space do not match. Lines that are
exactly one or more "#" characters with no following
text do not match.

CommonMark allows optional closing "#" sequences:
`## Foo ##` has heading text `Foo`. If present, the
closing sequence must be preceded by at least one space.

## Go-specific guidance

- Use `textnormalization.NormalizeText` for all heading
  comparisons.
- Use `logicalnames.LogicalNameParse` to parse and
  validate the logical name. Access `ln.Path` for the
  resolved file path. Use `logicalnames.NodeTypeSpec`
  for type comparison.
- Use the `file` package for file I/O: `FileOpen`,
  `FileReadLine`, `FileClose`.
- The package name should be `parsenode`.
