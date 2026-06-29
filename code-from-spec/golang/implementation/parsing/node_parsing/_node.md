---
depends_on:
  - SPEC/golang/dependencies/goccy-go-yaml
  - SPEC/golang/dependencies/yuin-goldmark
  - SPEC/golang/implementation/oslayer(interface)
output: internal/parsing/node_parsing.go
---

# SPEC/golang/implementation/parsing/node_parsing

Opens a spec node file, extracts frontmatter and body,
and returns a structured representation. The file is
opened and read once.

# Agent

Implement the types and function listed in the
Ownership section as a Go file in package `parsing`.

## Ownership

This file declares and implements:
- Types: `NodeFrontmatter`, `NodeSubsection`,
  `NodeSection`, `Node`
- Function: `ParseNode`

The following exist in other files of this package and
can be used but must not be redeclared:
- Functions: `NormalizeText` — declared in
  `text_normalization.go`.
- Functions: `CfsReferenceFromName` — declared in
  `logical_names.go`.
- Types: `CfsReference`, `CfsNodeTypeSpec` — declared
  in `logical_names.go`.
- Types and functions from the oslayer package
  (`CfsPath`, `OpenFile`, etc.) — imported externally.
- Error sentinels — declared in `errors.go`.

All unexported helpers must use the suffix `NP`
(e.g. `extractHeadingNP`, `buildSectionsNP`,
`parseYamlNP`). This is mandatory to avoid name
collisions with other files in the package.

## Logic

### Step 1 — Validate logical name

- Call CfsReferenceFromName(logical_name). If it fails,
  raise ErrNotASpecReference. Let `ref` be the result.
- If ref.NodeType is not CfsNodeTypeSpec, raise
  ErrNotASpecReference.
- If ref.Qualifier is not nil, raise ErrHasQualifier.

### Step 2 — Read file

- Call oslayer.OpenFile(oslayer.CfsPath(ref.Path),
  "read", 30000). If it fails, raise ErrFileUnreadable.
- Read all lines using handle.ReadLine() in a loop
  until ErrEndOfFile. Collect all lines. Call
  handle.Close().
- Join all lines with `\n` and append a trailing `\n`.
  Let `source` be the resulting byte slice.

### Step 3 — Extract frontmatter

- If `source` does not start with `---\n`:
    Set frontmatter = nil. Let `body` = `source`.
    Skip to step 4.

- Find the next occurrence of `\n---\n` after the
  first line. If not found, raise ErrMalformedYAML.

- Extract the text between the opening `---\n` and the
  closing `\n---\n` as `yaml_text`.

- Let `body` = everything after the closing `---\n`.

- If `yaml_text` is empty (nothing between delimiters):
    Set frontmatter = nil.
    Skip to step 4.

- Parse `yaml_text` as YAML. If parsing fails, raise
  ErrMalformedYAML.

- From the parsed YAML, extract the following fields,
  ignoring all other keys:
  - depends_on: list of strings. If absent or null,
    use nil.
  - input: *string. If absent or null, use nil.
  - output: *string. If absent or null, use nil.

- Build a NodeFrontmatter record with the extracted
  fields. Set frontmatter to a pointer to this record.

### Step 4 — Parse body with goldmark

- Parse `body` with goldmark:
  ```
  md := goldmark.New()
  doc := md.Parser().Parse(text.NewReader(body))
  ```

### Step 5 — Collect structural headings

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

### Step 6 — Build sections

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

### Step 7 — Return result

Return Node with reference = *ref, frontmatter,
name_section, public, agent, private.

## Go-specific guidance

- Use `github.com/goccy/go-yaml` for YAML unmarshalling.
  Define an unexported struct with `yaml` tags to map
  YAML keys to Go fields, then convert to the exported
  NodeFrontmatter type.
- Use `goldmark.New()` and `md.Parser().Parse(
  text.NewReader(body))` for body parsing.
- Use direct child iteration
  (`doc.FirstChild()` / `NextSibling()`) to collect
  headings. Check `n.Kind() == ast.KindHeading` and
  cast to `*ast.Heading` to read `Level`.
- Use `NormalizeText` from this package for heading
  comparisons.
- Use `CfsReferenceFromName` and `CfsNodeTypeSpec`
  from this package for validation.
- Use the `oslayer` package for `OpenFile`,
  `.ReadLine()`, `.Close()`, and `CfsPath`.
- Split content by `\n` using `strings.Split` on the
  string cast of the byte slice range.
- Error wrapping: wrap all errors with `fmt.Errorf`
  using `%w` so callers can match with `errors.Is()`.
- The package name should be `parsing`.

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

### File reading via oslayer package

goldmark needs `[]byte` but the `oslayer` package reads
line by line. The implementation reads all lines via
handle.ReadLine() loop, joins with `\n`, and converts to
`[]byte`. CRLF normalization is handled by the oslayer
package — no manual normalization needed.

### Content remains []string

The public interface keeps `Content []string` for
compatibility with chainhash and load_chain, which
process content line by line. Internally, the byte
range from the source is split into lines.

### Unified frontmatter and body parsing

Frontmatter extraction and body parsing were previously
separate packages. Unified into a single function
(`ParseNode`) that opens the file once, extracts
frontmatter from the YAML block, and parses the body
with goldmark. Eliminates double file I/O and
simplifies the public API — callers get both frontmatter
and structured body from one call.
