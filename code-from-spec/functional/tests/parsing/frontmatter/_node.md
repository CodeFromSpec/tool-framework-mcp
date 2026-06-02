---
depends_on:
  - ROOT/functional/logic/parsing/frontmatter(interface)
output: code-from-spec/functional/tests/parsing/frontmatter/output.md
---

# ROOT/functional/tests/parsing/frontmatter

Test cases for the frontmatter component.

# Public

## Test cases

### Happy path

#### Parses complete frontmatter (all fields)

Create a file with all fields: `depends_on`, `external`,
`input`, and `output`. Call `FrontmatterParse`.

Expect `depends_on` contains the listed dependencies.
`external` has two entries each with `path`. `input`
matches the specified value. `output` matches the
specified path. No error.

#### Parses frontmatter with only output

Create a file with only `output` in frontmatter. Call
`FrontmatterParse`.

Expect `depends_on` is empty, `external` is empty,
`input` is empty. `output` matches the specified path.
No error.

#### Parses frontmatter with only depends_on

Create a file with only `depends_on` in frontmatter.
Call `FrontmatterParse`.

Expect `depends_on` contains the listed values. All
other fields are empty. No error.

#### Parses frontmatter with external entries

Create a file with multiple `external` entries. Call
`FrontmatterParse`.

Expect `external` has two entries each with correct
`path`. No error.

#### Parses frontmatter with input field

Create a file with only the `input` field. Call
`FrontmatterParse`.

Expect `input` matches the specified value. All other
fields are empty. No error.

#### Ignores unknown frontmatter fields

Create a file with known fields plus extra unknown
fields (e.g., `custom_field: value`). Call
`FrontmatterParse`.

Expect no error. Known fields parsed correctly. Unknown
fields ignored.

#### File with no frontmatter returns empty result

Create a file with no `---` delimiter at all — body
content only. Call `FrontmatterParse`.

Expect no error. Result has all fields empty.

### Edge cases

#### Empty frontmatter

Create a file with empty frontmatter (opening and
closing `---` with nothing between). Call
`FrontmatterParse`.

Expect no error. Result has all fields empty.

#### File with only frontmatter, nothing after

Create a file with frontmatter and no body content
after the closing `---`. Call `FrontmatterParse`.

Expect no error. Fields parsed correctly.

#### Delimiter with trailing whitespace is not recognized

Create a file where the first line is `---   ` (with
trailing spaces). Call `FrontmatterParse`.

Expect no error. Result has all fields empty — the
line is not recognized as a delimiter.

### Failure cases

#### File does not exist

Call `FrontmatterParse` with a `PathCfs` pointing to a
non-existent file. Expect error FileUnreadable.

#### Propagates path errors

Call `FrontmatterParse` with an invalid `PathCfs`
(e.g., `"../../outside"`). Expect error DirectoryTraversal (propagated from
FileReader/PathUtils via FileOpen).

#### Malformed YAML

Create a file with invalid YAML between frontmatter
delimiters. Call `FrontmatterParse`.

Expect error MalformedYAML.

#### Unclosed frontmatter block

Create a file that starts with `---` but has no closing
`---`. Call `FrontmatterParse`.

Expect error MalformedYAML.

#### Missing required field in external entry

Create a file with an `external` entry that has no
`path` field. Call `FrontmatterParse`.

Expect error MalformedYAML.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface:
  `FrontmatterParse`.
