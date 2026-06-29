---
depends_on:
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing/frontmatter
output: internal/frontmatter/frontmatter_test.go
---

# SPEC/golang/tests/parsing/frontmatter

# Agent

## Test cases

### Happy path

#### Parses complete frontmatter (all fields)

Setup:
- Create a file with frontmatter containing
  depends_on (SPEC/, ARTIFACT/, EXTERNAL/ entries),
  input, and output.

Actions:
1. Call FrontmatterParse.

Expected:
- DependsOn contains all listed entries.
- Input matches. Output matches. No error.

#### Parses frontmatter with only output

Setup:
- File with only `output` in frontmatter.

Expected: DependsOn empty, Input empty, Output set.

#### Parses frontmatter with only depends_on

Setup:
- File with only `depends_on` in frontmatter.

Expected: DependsOn contains values, others empty.

#### Parses frontmatter with EXTERNAL/ in depends_on

Setup:
- File with `depends_on: ["EXTERNAL/proto/api.proto"]`.

Expected: DependsOn contains the EXTERNAL entry.

#### Parses frontmatter with input field

Setup:
- File with only `input` field.

Expected: Input set, others empty.

#### Ignores unknown frontmatter fields

Setup:
- File with known fields plus `custom_field: value`.

Expected: No error. Known fields correct. Unknown
ignored.

#### File with no frontmatter returns empty result

Setup:
- File with no `---` delimiter — body content only.

Expected: No error. All fields empty.

### Edge cases

#### Empty frontmatter

Setup:
- File with `---` then `---` with nothing between.

Expected: No error. All fields empty.

#### File with only frontmatter, nothing after

Setup:
- File with frontmatter and no body.

Expected: No error. Fields parsed correctly.

#### Delimiter with trailing whitespace is not recognized

Setup:
- File whose first line is `---   ` (trailing spaces).

Expected: No error. All fields empty — line not
recognized as delimiter.

### Failure cases

#### File does not exist

Actions:
1. Call FrontmatterParse with non-existent CfsPath.

Expected: Error `ErrFileUnreadable`.

#### Propagates path errors

Actions:
1. Call FrontmatterParse with invalid CfsPath
   (`"../../outside"`).

Expected: Error `oslayer.ErrDirectoryTraversal`.

#### Malformed YAML

Setup:
- File with invalid YAML between `---` delimiters.

Expected: Error `ErrMalformedYAML`.

#### Unclosed frontmatter block

Setup:
- File starts with `---` but no closing `---`.

Expected: Error `ErrMalformedYAML`.

#### Unknown field 'external' is silently ignored

Setup:
- File with `external: "some/ref"` plus `output`.

Expected: No error. `external` ignored. Output set.

## Go-specific guidance

- The package name is `frontmatter_test` (external test
  package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper to set the working directory.
