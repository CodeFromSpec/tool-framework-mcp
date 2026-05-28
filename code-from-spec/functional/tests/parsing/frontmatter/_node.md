---
depends_on:
  - ROOT/functional/logic/parsing/frontmatter(interface)
outputs:
  - id: frontmatter_tests
    path: code-from-spec/functional/tests/parsing/frontmatter/output.md
---

# ROOT/functional/tests/parsing/frontmatter

Test cases for the frontmatter component.

# Public

## Test cases

### Happy path

#### Parses complete frontmatter (all fields)

Create a file with all fields: `depends_on`, `external`
(with fragments), `input`, and `outputs`. Call
`FrontmatterParse`.

Expect `depends_on` contains the listed dependencies.
`external` has one entry with `path` and one fragment
containing `description`, `lines`, and `hash`. `input`
matches the specified value. `outputs` has two entries
each with `id` and `path`. No error.

#### Parses frontmatter with only outputs

Create a file with only `outputs` in frontmatter. Call
`FrontmatterParse`.

Expect `depends_on` is empty, `external` is empty,
`input` is empty. `outputs` has one entry with the
correct `id` and `path`. No error.

#### Parses frontmatter with only depends_on

Create a file with only `depends_on` in frontmatter.
Call `FrontmatterParse`.

Expect `depends_on` contains the listed values. All
other fields are empty. No error.

#### Parses frontmatter with external and fragments

Create a file with `external` entries including
fragments. Call `FrontmatterParse`.

Expect `external` has two entries. First entry has `path`
and two fragments with correct `description`, `lines`,
and `hash`. Second entry has `path` only, no fragments.
No error.

#### Parses frontmatter with input field

Create a file with only the `input` field. Call
`FrontmatterParse`.

Expect `input` matches the specified value. All other
fields are empty. No error.

#### Fragment without description

Create a file with an `external` entry containing a
fragment with `lines` and `hash` but no `description`.
Call `FrontmatterParse`.

Expect the fragment is parsed with `description` absent,
`lines` and `hash` correct. No error.

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
non-existent file. Expect "file unreadable".

#### Propagates path errors

Call `FrontmatterParse` with an invalid `PathCfs`
(e.g., `"../../outside"`). Expect error "directory
traversal" propagated from `FileOpen`.

#### Malformed YAML

Create a file with invalid YAML between frontmatter
delimiters. Call `FrontmatterParse`.

Expect "malformed YAML".

#### Unclosed frontmatter block

Create a file that starts with `---` but has no closing
`---`. Call `FrontmatterParse`.

Expect "malformed YAML".

#### Missing required field in external entry

Create a file with an `external` entry that has no
`path` field. Call `FrontmatterParse`.

Expect "malformed YAML".

#### Missing required field in fragment

Create a file with an `external` entry containing a
fragment that has `lines` but no `hash`. Call
`FrontmatterParse`.

Expect "malformed YAML".

#### Missing required field in output entry

Create a file with an `outputs` entry that has `id`
but no `path`. Call `FrontmatterParse`.

Expect "malformed YAML".

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface:
  `FrontmatterParse`.
