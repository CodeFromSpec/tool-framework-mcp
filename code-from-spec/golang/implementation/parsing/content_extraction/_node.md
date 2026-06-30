---
depends_on:
  - SPEC/golang/implementation/oslayer(interface)
output: internal/parsing/content_extraction.go
---

# SPEC/golang/implementation/parsing/content_extraction

Implements content extraction helpers used by chain
hashing, chain assembly, and cache population. These
functions produce the exact text that participates in
the chain hash — hash and delivery never diverge.

# Agent

Implement the content extraction helpers as a Go file
in package `parsing`.

## Ownership

This file declares and implements:
- Functions: `ExtractBlock`, `FormatSection`,
  `ConcatenateSubsections`, `ExtractAgentContent`,
  `ReadFileContent`

The following exist in other files of this package and
can be used but must not be redeclared:
- Error sentinels — declared in `errors.go`.
- Types (`Node`, `NodeSection`, `NodeSubsection`,
  `NodeFrontmatter`, `CfsReference`,
  etc.) — declared in other files.
- Functions (`ParseNode`, `NormalizeText`,
  `CfsReferenceFromName`, etc.) — declared in other
  files.

To avoid name collisions with other files in this
package, all identifiers you declare beyond the ones
listed in the Ownership section (helper functions,
variables, types) must use the suffix `CE`.

## Logic

### ExtractBlock

```
func ExtractBlock(content []string) string
```

1. Remove leading blank lines from `content`
   (lines that are empty or contain only spaces and
   tabs, U+0020 and U+0009).
2. Remove trailing blank lines.
3. If nothing remains, return empty string.
4. Join remaining lines with `\n` and append exactly
   one `\n`.

### FormatSection

```
func FormatSection(rawHeading string, content []string) string
```

1. Let `head` = `rawHeading` with trailing whitespace
   (U+0020 and U+0009) removed, followed by `\n`.
2. Let `body` = `ExtractBlock(content)`.
3. Return concatenation of `head` and `body`.

### ConcatenateSubsections

```
func ConcatenateSubsections(subsections []*NodeSubsection) string
```

1. Let `result` = empty string.
2. For each subsection in `subsections`:
   a. Let `block` = `FormatSection(subsection.RawHeading,
      subsection.Content)`.
   b. If `result` is not empty and `block` is not empty,
      append `\n` to `result`.
   c. Append `block` to `result`.
3. Return `result`.

### ExtractAgentContent

```
func ExtractAgentContent(node *Node) string
```

1. If `node.Agent` is nil, return empty string.
2. Let `text` = `ExtractBlock(node.Agent.Content)`.
3. For each subsection in `node.Agent.Subsections`:
   a. Let `subBlock` = `FormatSection(
      subsection.RawHeading, subsection.Content)`.
   b. If `text` is not empty and `subBlock` is not
      empty, append `\n` to `text`.
   c. Append `subBlock` to `text`.
4. If `text` is empty, return empty string.
5. Return `text`.

### ReadFileContent

```
func ReadFileContent(cfsPath oslayer.CfsPath) (string, error)
```

1. Call `oslayer.OpenFile(cfsPath, "read", 30000)`.
   If it fails, propagate the error.
2. Let `lines` = empty list.
3. Loop:
   a. Call `handle.ReadLine()`.
   b. If `oslayer.ErrEndOfFile`, exit loop.
   c. Append the line to `lines`.
4. Call `handle.Close()`.
   (Call `handle.Close()` in error paths too before
   propagating.)
5. Let `text` = join `lines` with `\n`, append `\n`.
6. Return `text`.

## Go-specific guidance

- The package name is `parsing` (same package as
  `node_parsing.go`, `text_normalization.go`, etc.).
- Use `strings.TrimRight` with `" \t"` for removing
  trailing whitespace from headings.
- Use the `oslayer` package for `OpenFile`, `CfsPath`,
  `ErrEndOfFile`.
- A blank line is a line that, after trimming spaces
  and tabs, is empty.
- Wrap errors with `fmt.Errorf` using `%w`.
