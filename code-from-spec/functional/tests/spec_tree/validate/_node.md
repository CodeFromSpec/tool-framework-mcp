---
depends_on:
  - ROOT/functional/logic/spec_tree/validate(interface)
outputs:
  - id: format_validation_tests
    path: code-from-spec/functional/tests/spec_tree/validate/output.md
---

# ROOT/functional/tests/spec_tree/validate

Test cases for the spec tree validation component.

# Public

## Test cases

### Happy path

#### Valid leaf node passes all checks

Input: ROOT (intermediate, has children ROOT/a and
ROOT/b), ROOT/a (leaf, node.name_section.heading =
"ROOT/a", depends_on = ["ROOT/b"], outputs = [{id:
"out", path: "internal/out.go"}]), ROOT/b (leaf,
node.name_section.heading = "ROOT/b"). Call
SpecTreeValidate. Expect no format errors.

#### Valid intermediate node passes all checks

Input: ROOT (intermediate, node.name_section.heading =
"ROOT", node.public present with empty content, no
frontmatter fields, no agent section), ROOT/a (leaf,
node.name_section.heading = "ROOT/a"). Call
SpecTreeValidate. Expect no format errors.

#### Leaf with no frontmatter fields

Input: ROOT (node.name_section.heading = "ROOT"),
ROOT/a (leaf, node.name_section.heading = "ROOT/a",
empty frontmatter). Call SpecTreeValidate. Expect no
format errors.

### name_heading

#### Heading matches logical name

Input: ROOT (node.name_section.heading = "ROOT"),
ROOT/a (node.name_section.heading = "ROOT/a"). Call
SpecTreeValidate. Expect no name_heading error.

#### Heading does not match logical name

Input: ROOT (node.name_section.heading = "ROOT"),
ROOT/a (node.name_section.heading = "ROOT/wrong").
Call SpecTreeValidate. Expect a FormatError with rule
= "name_heading" for ROOT/a.

### leaf_only_fields

#### Intermediate node with depends_on

Input: ROOT, ROOT/a (intermediate, has child
ROOT/a/b, depends_on = ["ROOT/b"]), ROOT/a/b (leaf).
Call SpecTreeValidate. Expect a FormatError with rule
= "leaf_only_fields" for ROOT/a.

#### Intermediate node with outputs

Input: ROOT, ROOT/a (intermediate, has child
ROOT/a/b, outputs = [{id: "x", path: "x.go"}]),
ROOT/a/b (leaf). Call SpecTreeValidate. Expect a
FormatError with rule = "leaf_only_fields" for ROOT/a.

#### Intermediate node with input

Input: ROOT, ROOT/a (intermediate, has child
ROOT/a/b, input = "ARTIFACT/c(id)"), ROOT/a/b (leaf).
Call SpecTreeValidate. Expect a FormatError with rule
= "leaf_only_fields" for ROOT/a.

#### Intermediate node with external

Input: ROOT, ROOT/a (intermediate, has child
ROOT/a/b, external = [{path: "some/file.txt"}]),
ROOT/a/b (leaf). Call SpecTreeValidate. Expect a
FormatError with rule = "leaf_only_fields" for ROOT/a.

#### Intermediate node with multiple restricted fields

Input: ROOT, ROOT/a (intermediate, has child
ROOT/a/b, depends_on = ["ROOT/b"], outputs = [{id:
"x", path: "x.go"}]), ROOT/a/b (leaf). Call
SpecTreeValidate. Expect two FormatErrors with rule =
"leaf_only_fields" for ROOT/a (one per field).

### leaf_only_agent

#### Intermediate node with agent section

Input: ROOT, ROOT/a (intermediate, has child
ROOT/a/b, node.agent present with content =
["Agent instructions."]), ROOT/a/b (leaf). Call
SpecTreeValidate. Expect a FormatError with rule =
"leaf_only_agent" for ROOT/a.

#### Leaf node with agent section — no error

Input: ROOT (node.name_section.heading = "ROOT"),
ROOT/a (leaf, node.name_section.heading = "ROOT/a",
node.agent present with content = ["Agent
instructions."]). Call SpecTreeValidate. Expect no
leaf_only_agent error.

### dependency_targets

#### depends_on targets non-existent ROOT node

Input: ROOT, ROOT/a (leaf, depends_on =
["ROOT/missing"]). Call SpecTreeValidate. Expect a
FormatError with rule = "dependency_targets" for
ROOT/a.

#### depends_on targets ancestor

Input: ROOT, ROOT/a (intermediate, has child
ROOT/a/b), ROOT/a/b (leaf, depends_on = ["ROOT"]).
Call SpecTreeValidate. Expect a FormatError with rule
= "dependency_targets" for ROOT/a/b.

#### depends_on targets descendant

Input: ROOT, ROOT/a (intermediate, has child
ROOT/a/b, depends_on = ["ROOT/a/b"]), ROOT/a/b
(leaf). Call SpecTreeValidate. Expect a FormatError
with rule = "dependency_targets" for ROOT/a.

#### depends_on targets self

Input: ROOT, ROOT/a (leaf, depends_on = ["ROOT/a"]).
Call SpecTreeValidate. Expect a FormatError with rule
= "dependency_targets" for ROOT/a.

#### depends_on with valid ROOT qualifier

Input: ROOT, ROOT/a (leaf), ROOT/b (leaf, depends_on
= ["ROOT/a(interface)"]). Call SpecTreeValidate.
Expect no dependency_targets error (qualifier
stripped, ROOT/a exists and is not
ancestor/descendant/self).

#### depends_on with valid ARTIFACT reference

Input: ROOT, ROOT/a (leaf, outputs = [{id: "lib",
path: "lib.go"}]), ROOT/b (leaf, depends_on =
["ARTIFACT/a(lib)"]). Call SpecTreeValidate. Expect
no dependency_targets error.

#### depends_on with non-existent ARTIFACT reference

Input: ROOT, ROOT/a (leaf, depends_on =
["ARTIFACT/missing(id)"]). Call SpecTreeValidate.
Expect a FormatError with rule = "dependency_targets"
for ROOT/a.

#### Multiple invalid depends_on — one error per entry

Input: ROOT, ROOT/a (leaf, depends_on =
["ROOT/missing", "ROOT/also_missing"]). Call
SpecTreeValidate. Expect two FormatErrors with rule =
"dependency_targets" for ROOT/a.

### input_target

#### Valid input reference

Input: ROOT, ROOT/a (leaf, outputs = [{id: "out",
path: "a.go"}]), ROOT/b (leaf, input =
"ARTIFACT/a(out)"). Call SpecTreeValidate. Expect no
input_target error.

#### Input not starting with ARTIFACT/

Input: ROOT, ROOT/a (leaf, input = "ROOT/something").
Call SpecTreeValidate. Expect a FormatError with rule
= "input_target" for ROOT/a.

#### Input references non-existent artifact

Input: ROOT, ROOT/a (leaf, input =
"ARTIFACT/missing(id)"). Call SpecTreeValidate.
Expect a FormatError with rule = "input_target" for
ROOT/a.

### external_files

Fragment hashes use SHA-1 encoded as base64url (RFC
4648 §5, no padding) — always 27 characters. The input
to SHA-1 is the lines in the declared range, read with
`FileReadLine` (which normalizes CRLF to LF and strips
terminators), each with `\n` (LF) appended — including
the last line. Tests that need a "correct hash" must
compute it using this algorithm.

#### External file exists — no fragments

Input: ROOT, ROOT/a (leaf, external = [{path:
"some/file.txt"}]). Create "some/file.txt" on disk
with content "hello\n". Call SpecTreeValidate. Expect
no external_files error.

#### External file does not exist

Input: ROOT, ROOT/a (leaf, external = [{path:
"nonexistent.txt"}]). Do not create the file. Call
SpecTreeValidate. Expect a FormatError with rule =
"external_files" for ROOT/a.

#### Fragment with valid hash

Input: ROOT, ROOT/a (leaf, external = [{path:
"f.txt", fragments: [{lines: "1-3", hash: <correct
hash>}]}]). Create "f.txt" with 5 lines: "alpha",
"bravo", "charlie", "delta", "echo" (one per line).
Compute the correct hash from lines 1-3 per the rule
above. Call SpecTreeValidate. Expect no external_files
error.

#### Fragment with invalid hash

Input: ROOT, ROOT/a (leaf, external = [{path:
"f.txt", fragments: [{lines: "1-3", hash:
"wrong_______________________"}]}]). Create "f.txt"
with 5 lines: "alpha", "bravo", "charlie", "delta",
"echo". Call SpecTreeValidate. Expect a FormatError
with rule = "external_files" for ROOT/a.

#### Fragment with invalid range format

Input: ROOT, ROOT/a (leaf, external = [{path:
"f.txt", fragments: [{lines: "abc", hash: "x"}]}]).
Create "f.txt" with content "hello\n". Call
SpecTreeValidate. Expect a FormatError with rule =
"external_files" for ROOT/a.

#### Fragment with start > end

Input: ROOT, ROOT/a (leaf, external = [{path:
"f.txt", fragments: [{lines: "5-3", hash: "x"}]}]).
Create "f.txt" with 5 lines: "alpha", "bravo",
"charlie", "delta", "echo". Call SpecTreeValidate.
Expect a FormatError with rule = "external_files" for
ROOT/a.

#### Fragment with start < 1

Input: ROOT, ROOT/a (leaf, external = [{path:
"f.txt", fragments: [{lines: "0-3", hash: "x"}]}]).
Create "f.txt" with 5 lines: "alpha", "bravo",
"charlie", "delta", "echo". Call SpecTreeValidate.
Expect a FormatError with rule = "external_files" for
ROOT/a.

#### Fragment out of range

Input: ROOT, ROOT/a (leaf, external = [{path:
"f.txt", fragments: [{lines: "1-100", hash:
"x"}]}]). Create "f.txt" with 5 lines: "alpha",
"bravo", "charlie", "delta", "echo". Call
SpecTreeValidate. Expect a FormatError with rule =
"external_files" for ROOT/a indicating fragment out
of range.

### output_paths

#### Valid output path

Input: ROOT, ROOT/a (leaf, outputs = [{id: "x",
path: "internal/x.go"}]). Call SpecTreeValidate.
Expect no output_paths error.

#### Output path with traversal

Input: ROOT, ROOT/a (leaf, outputs = [{id: "x",
path: "../../etc/passwd"}]). Call SpecTreeValidate.
Expect a FormatError with rule = "output_paths" for
ROOT/a.

#### Output path with backslash

Input: ROOT, ROOT/a (leaf, outputs = [{id: "x",
path: "internal\\x.go"}]). Call SpecTreeValidate.
Expect a FormatError with rule = "output_paths" for
ROOT/a.

### duplicate_subsections

#### Unique subsection headings — no error

Input: ROOT, ROOT/a (leaf, node.public present with
subsections [{heading: "interface", raw_heading:
"## Interface", content: ["Types."]}, {heading:
"context", raw_heading: "## Context", content:
["Background."]}]). Call SpecTreeValidate. Expect no
duplicate_subsections error.

#### Duplicate subsection headings

Input: ROOT, ROOT/a (leaf, node.public present with
subsections [{heading: "interface", raw_heading:
"## Interface", content: ["First."]}, {heading:
"interface", raw_heading: "## Interface", content:
["Second."]}]). Call SpecTreeValidate. Expect one
FormatError with rule = "duplicate_subsections" for
ROOT/a (the second occurrence).

#### Three identical subsection headings

Input: ROOT, ROOT/a (leaf, node.public present with
subsections [{heading: "interface", raw_heading:
"## Interface", content: ["First."]}, {heading:
"interface", raw_heading: "## Interface", content:
["Second."]}, {heading: "interface", raw_heading:
"## Interface", content: ["Third."]}]). Call
SpecTreeValidate. Expect two FormatErrors with rule =
"duplicate_subsections" for ROOT/a (second and third
occurrences).

#### No public section — skip

Input: ROOT, ROOT/a (leaf, node.public absent). Call
SpecTreeValidate. Expect no duplicate_subsections
error.

### Cross-cutting

#### Collects multiple errors from different rules

Input: ROOT, ROOT/a (leaf, node.name_section.heading
= "ROOT/wrong", depends_on = ["ROOT/missing"],
node.public present with subsections [{heading:
"interface", raw_heading: "## Interface", content:
["First."]}, {heading: "interface", raw_heading:
"## Interface", content: ["Second."]}]). Call
SpecTreeValidate. Expect at least three FormatErrors:
one with rule = "name_heading", one with rule =
"dependency_targets", one with rule =
"duplicate_subsections".

#### Empty input list

Input: empty list. Call SpecTreeValidate. Expect no
format errors.

# Agent

Generate a test specification document listing each test
case with its setup, actions, and expected outcome.

## Rules

- Use the function name from the interface:
  `SpecTreeValidate`.
- Use the record names from the interface:
  `SpecTreeValidateInput`, `FormatError`.
- Describe tests in terms of the functional interface —
  use function names, record names, and rule names from
  the spec.
- Each test case has: a description, setup (input data
  as list of SpecTreeValidateInput), actions (function
  call), and expected outcome.
- Input is always a list of `SpecTreeValidateInput` — no
  file I/O in tests, except for external_files tests
  which require files on disk.
- For external_files tests that need files on disk,
  describe what files to create and their content.
