---
depends_on:
  - SPEC/golang/test/utils/chdir
  - SPEC/golang/implementation/parsing(interface)
  - SPEC/golang/implementation/parsing/content_extraction
  - SPEC/golang/implementation/oslayer(interface)
output: internal/parsing/content_extraction_test.go
---

# SPEC/golang/test/cases/parsing/content_extraction

# Agent

## Test setup guidance

`ExtractBlock`, `FormatSection`, `ConcatenateSubsections`,
and `ExtractAgentContent` are pure functions that operate
on in-memory data — no file I/O needed.

`ReadFileContent` reads from disk via `oslayer.OpenFile`,
so tests for it need `testutils.Chdir(t)` for isolation.

## Test cases

### ExtractBlock

#### Removes leading blank lines

Actions:
1. Call `parsing.ExtractBlock([]string{"", "  ", "hello"})`.

Expected: `"hello\n"`.

#### Removes trailing blank lines

Actions:
1. Call `parsing.ExtractBlock([]string{"hello", "", "  "})`.

Expected: `"hello\n"`.

#### Preserves interior blank lines

Actions:
1. Call `parsing.ExtractBlock([]string{"a", "", "b"})`.

Expected: `"a\n\nb\n"`.

#### Empty input returns empty string

Actions:
1. Call `parsing.ExtractBlock([]string{})`.

Expected: `""`.

#### All blank lines returns empty string

Actions:
1. Call `parsing.ExtractBlock([]string{"", "  ", "\t"})`.

Expected: `""`.

#### Ends with exactly one LF

Actions:
1. Call `parsing.ExtractBlock([]string{"hello"})`.

Expected: `"hello\n"`.

### FormatSection

#### Heading plus content

Actions:
1. Call `parsing.FormatSection("## Interface",
   []string{"content line"})`.

Expected: `"## Interface\ncontent line\n"`.

#### Strips trailing whitespace from heading

Actions:
1. Call `parsing.FormatSection("## Interface   ",
   []string{"content"})`.

Expected: `"## Interface\ncontent\n"`.

#### Empty content — heading only

Actions:
1. Call `parsing.FormatSection("## Interface",
   []string{})`.

Expected: `"## Interface\n"`.

### ConcatenateSubsections

#### Single subsection

Actions:
1. Build one `parsing.NodeSubsection` with
   RawHeading `"## A"`, Content `["line1"]`.
2. Call `parsing.ConcatenateSubsections(subs)`.

Expected: `"## A\nline1\n"`.

#### Multiple subsections separated by blank line

Actions:
1. Build two subsections: `"## A"` with `["a1"]` and
   `"## B"` with `["b1"]`.
2. Call `parsing.ConcatenateSubsections(subs)`.

Expected: `"## A\na1\n\n## B\nb1\n"`.

#### Empty subsection skipped in separation

Actions:
1. Build three subsections: `"## A"` with `["a1"]`,
   `"## B"` with `[]` (empty), `"## C"` with `["c1"]`.
2. Call `parsing.ConcatenateSubsections(subs)`.

Expected: `"## A\na1\n\n## C\nc1\n"` (B contributes
heading only, which is non-empty, so blank line is
added).

Wait — FormatSection with empty content returns just
the heading line. That's non-empty. So all three are
separated.

Expected: `"## A\na1\n\n## B\n\n## C\nc1\n"`.

### ExtractAgentContent

#### Agent with direct content and subsections

Actions:
1. Build a `parsing.Node` with Agent section:
   Content `["direct line"]`,
   Subsections with one entry `"## Logic"` /
   `["step 1"]`.
2. Call `parsing.ExtractAgentContent(node)`.

Expected: `"direct line\n\n## Logic\nstep 1\n"`.

#### Agent with no content — returns empty

Actions:
1. Build a `parsing.Node` with Agent nil.
2. Call `parsing.ExtractAgentContent(node)`.

Expected: `""`.

#### Agent with only blank content lines and no subsections

Actions:
1. Build a `parsing.Node` with Agent section:
   Content `["", "  "]`, Subsections empty.
2. Call `parsing.ExtractAgentContent(node)`.

Expected: `""`.

### ReadFileContent

#### Reads file content

Setup:
1. Use `testutils.Chdir(t)`.
2. Create a file at `test.txt` with content
   `"line1\nline2\n"`.

Actions:
1. Call `parsing.ReadFileContent("test.txt")`.

Expected: `"line1\nline2\n"`.

#### Returns error for missing file

Setup:
1. Use `testutils.Chdir(t)`.

Actions:
1. Call `parsing.ReadFileContent("missing.txt")`.

Expected: error wrapping `oslayer.ErrFileUnreadable`.

## Go-specific guidance

- The package name is `parsing_test` (external test
  package).
- Use `testutils.Chdir(t)` only for `ReadFileContent`
  tests.
- Build `parsing.NodeSubsection` and `parsing.Node`
  structs directly for in-memory tests.
- Use `errors.Is` for error sentinel checks.
