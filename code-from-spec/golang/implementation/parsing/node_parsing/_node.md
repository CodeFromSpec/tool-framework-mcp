---
depends_on:
  - SPEC/golang/dependencies/yuin-goldmark
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/utils/logical_names
  - SPEC/golang/implementation/utils/text_normalization
output: internal/parsenode/parsenode.go
---

# SPEC/golang/implementation/parsing/node_parsing

Parses the body of a spec node file into a structured
representation of its sections and subsections using
goldmark for CommonMark parsing.

# Public

## Package

`package parsenode`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/parsenode"`

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

`Heading` is the normalized form (after `NormalizeText`),
used for comparisons and lookups. `RawHeading` is the
original heading line as it appears in the file (including
`#` prefix and closing `##` if present), preserved for
hashing.

`Content` is a list of lines between the heading and
the next structural heading (or end of file). Lines do
not include line terminators.

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

# Agent

Implement the node parsing component as a Go package.

## Logic

### Validate logical name

- Call LogicalNameParse(logical_name). If it fails,
  raise ErrNotASpecReference. Let `ln` be the result.
- If ln.Type is not NodeTypeSpec, raise
  ErrNotASpecReference.
- If ln.Qualifier is not nil, raise ErrHasQualifier.

### Read file

- Call FileOpen(PathCfs{Value: ln.Path}, "read",
  30000). If it fails, raise ErrFileUnreadable.
- Read all lines using FileReadLine in a loop until
  ErrEndOfFile. Collect all lines. Call FileClose.
- Join all lines with `\n` and append a trailing `\n`.
  Let `source` be the resulting byte slice.

### Skip frontmatter

- If `source` starts with `---\n`:
    Find the next occurrence of `\n---\n` after the
    first line. If not found, raise
    ErrUnexpectedContentBeforeFirstHeading.
    Let `body` = everything after the closing `---\n`.
- Else: let `body` = `source`.

### Parse with goldmark

- Parse `body` with goldmark:
  ```
  md := goldmark.New()
  doc := md.Parser().Parse(text.NewReader(body))
  ```

### Collect structural headings

Iterate the direct children of `doc`. For each child
that is `*ast.Heading` with Level 1 or 2:

- **Heading text**: concatenate `*ast.Text` segments
  from the heading's inline children. Let `text_part`
  be the result.

- **Line boundaries**: `Lines().At(0)` covers only the
  heading text content (e.g. `Foo` in `## Foo ##`),
  not the `#` prefix or closing `##`. To recover the
  full raw line:
  - `lineStart`: scan backward from
    `Lines().At(0).Start` to find the preceding `\n`
    (or start of body).
  - `lineEnd`: scan forward from
    `Lines().At(0).Stop` to find the next `\n` (or
    end of body).

- **Raw heading**: `string(body[lineStart:lineEnd])`
  with trailing whitespace (spaces, tabs) removed.

- **Normalized heading**: NormalizeText(text_part).

- **Content lines**: the range
  `body[lineEnd+1 : nextHeadingLineStart]`, where
  `nextHeadingLineStart` is the `lineStart` of the
  next structural heading, or `len(body)` if last.
  If `lineEnd` is at end of body, content is empty.
  Split by `\n` into a list of strings. Remove
  trailing empty string if present.

- Record: level, normalized heading, raw heading,
  content lines.

Headings with Level 3+ are NOT collected — they are
part of the content between structural headings.

If there is non-blank content in `body` before the
first structural heading, raise
ErrUnexpectedContentBeforeFirstHeading.

### Build sections

Process the collected heading records in order.

Let name_section, public, agent, private = absent.
Let current_section = absent.
Let current_subsection = absent.

For each record:

- **Level 1**: finalize current_subsection into
  current_section if present. Finalize
  current_section if present. Then classify:
  - If name_section is absent: compare normalized
    heading with NormalizeText(logical_name). If
    mismatch, raise ErrNodeNameDoesNotMatch. Start
    name_section. Set current_section.
  - `"public"`: if already set, raise
    ErrDuplicatePublicSection. Start section, set
    public and current_section.
  - `"agent"`: if already set, raise
    ErrDuplicateAgentSection. Start section, set
    agent and current_section.
  - `"private"`: if already set, raise
    ErrDuplicatePrivateSection. Start section, set
    private and current_section.
  - Anything else: raise ErrUnrecognizedSection.

- **Level 2**: if current_section is absent, raise
  ErrUnexpectedContentBeforeFirstHeading. Finalize
  current_subsection if present. Check for duplicate
  normalized heading in current_section's
  subsections — if found, raise
  ErrDuplicateSubsection. Start new subsection.
  Set current_subsection.

After all records: finalize current_subsection and
current_section.

A section's Content is the content lines collected
with its level-1 heading (lines between the heading
and the first subsection or next level-1 heading).
Each subsection gets its own content lines.

If name_section is absent, raise
ErrUnexpectedContentBeforeFirstHeading.

Return Node with name_section, public, agent, private.

## Go-specific guidance

- Use `goldmark.New()` and `md.Parser().Parse(
  text.NewReader(body))` for parsing.
- Use direct child iteration
  (`doc.FirstChild()` / `NextSibling()`) to collect
  headings. Check `n.Kind() == ast.KindHeading` and
  cast to `*ast.Heading` to read `Level`.
- Use `textnormalization.NormalizeText` for heading
  comparisons.
- Use `logicalnames.LogicalNameParse` for validation.
  Use `logicalnames.NodeTypeSpec` for type comparison.
- Use the `file` package for `FileOpen`,
  `FileReadLine`, `FileClose`.
- Use the `pathutils` package for `PathCfs`.
- Split content by `\n` using `strings.Split` on the
  string cast of the byte slice range.
- The package name should be `parsenode`.

# Private

## Decisions

### Migrated from manual parsing to goldmark

The previous implementation parsed markdown line by
line, manually tracking fenced code blocks, ATX heading
patterns, and closing `##` sequences. goldmark handles
all of this correctly as a CommonMark-compliant parser.
The migration eliminates fence tracking, heading regex,
and closing hash stripping with no loss of
functionality.

### File reading via file package

goldmark needs `[]byte` but the `file` package reads
line by line. The implementation reads all lines via
FileReadLine loop, joins with `\n`, and converts to
`[]byte`. CRLF normalization is handled by the file
package — no manual normalization needed.

### Content remains []string

The public interface keeps `Content []string` for
compatibility with chainhash and load_chain, which
process content line by line. Internally, the byte
range from the source is split into lines.
