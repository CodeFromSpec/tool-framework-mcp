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

Create a file with all fields: depends_on, external (with
fragments), input, and outputs. Call ParseFrontmatter.

Expect depends_on contains the listed dependencies.
External has one entry with path and one fragment containing
description, lines, and hash. Input matches the specified
value. Outputs has two entries each with id and path.

#### Parses frontmatter with only outputs

Create a file with only outputs in frontmatter. Call
ParseFrontmatter.

Expect depends_on is empty, external is empty, input is
empty. Outputs has one entry with the correct id and path.
No error.

#### Parses frontmatter with only depends_on

Create a file with only depends_on in frontmatter. Call
ParseFrontmatter.

Expect depends_on contains the listed values. External is
empty, input is empty, outputs is empty. No error.

#### Parses frontmatter with external and fragments

Create a file with external entries including fragments.
Call ParseFrontmatter.

Expect external has two entries. First entry has path and
two fragments with correct description, lines, and hash.
Second entry has path only, no fragments.

#### Parses frontmatter with input field

Create a file with only the input field. Call
ParseFrontmatter.

Expect input matches the specified value. All other fields
are empty. No error.

#### Ignores unknown frontmatter fields

Create a file with known fields plus extra unknown fields.
Call ParseFrontmatter.

Expect no error. Known fields parsed correctly. Unknown
fields ignored.

#### File with no frontmatter returns empty result

Create a file with no frontmatter delimiters at all. Call
ParseFrontmatter.

Expect no error. Result has all fields empty.

### Edge cases

#### Empty frontmatter

Create a file with empty frontmatter (opening and closing
delimiters with nothing between). Call ParseFrontmatter.

Expect no error. Result has all fields empty.

#### File with only frontmatter, nothing after

Create a file with frontmatter and no body content after
the closing delimiter. Call ParseFrontmatter.

Expect no error. Fields parsed correctly.

### Failure cases

#### File does not exist

Call ParseFrontmatter with a non-existent path. Expect
"read error".

#### Malformed YAML in frontmatter

Create a file with invalid YAML between frontmatter
delimiters. Call ParseFrontmatter.

Expect "frontmatter parse error".

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Describe tests in terms of the functional interface —
  use function names and error names from the interface,
  not language-specific constructs.
- Each test case has: a description, setup (what files to
  create and with what content), actions (what functions
  to call), and expected outcome.
- Do not prescribe how to create test files or assert
  results — those are implementation details for the
  language layer.
