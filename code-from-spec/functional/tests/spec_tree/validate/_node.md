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

Input: ROOT (intermediate, has child), ROOT/a (leaf)
with correct heading, valid depends_on, valid outputs.
Call SpecTreeValidate. Expect no format errors.

#### Valid intermediate node passes all checks

Input: ROOT (intermediate), ROOT/a (leaf). ROOT has
only a name section and public section, no frontmatter
fields, no agent section. Call SpecTreeValidate. Expect
no format errors.

#### Leaf with no frontmatter fields

Input: ROOT, ROOT/a (leaf) with empty frontmatter and
correct heading. Call SpecTreeValidate. Expect no
format errors.

### name_heading

#### Heading matches logical name

Input: ROOT, ROOT/a where node.name_section.heading =
"ROOT/a". Call SpecTreeValidate. Expect no
name_heading error.

#### Heading does not match logical name

Input: ROOT, ROOT/a where node.name_section.heading =
"ROOT/wrong". Call SpecTreeValidate. Expect a
FormatError with rule = "name_heading".

### leaf_only_fields

#### Intermediate node with depends_on

Input: ROOT, ROOT/a (intermediate, has child ROOT/a/b),
ROOT/a/b. ROOT/a has depends_on = ["ROOT/b"]. Call
SpecTreeValidate. Expect a FormatError with rule =
"leaf_only_fields" for ROOT/a.

#### Intermediate node with outputs

Input: ROOT, ROOT/a (intermediate), ROOT/a/b. ROOT/a
has outputs = [{id: "x", path: "x.go"}]. Call
SpecTreeValidate. Expect a FormatError with rule =
"leaf_only_fields".

#### Intermediate node with input

Input: ROOT, ROOT/a (intermediate), ROOT/a/b. ROOT/a
has input = "ARTIFACT/c(id)". Call SpecTreeValidate.
Expect a FormatError with rule = "leaf_only_fields".

#### Intermediate node with external

Input: ROOT, ROOT/a (intermediate), ROOT/a/b. ROOT/a
has external = [{path: "some/file.txt"}]. Call
SpecTreeValidate. Expect a FormatError with rule =
"leaf_only_fields".

#### Intermediate node with multiple restricted fields

Input: ROOT, ROOT/a (intermediate), ROOT/a/b. ROOT/a
has depends_on and outputs both non-empty. Call
SpecTreeValidate. Expect two FormatErrors with rule =
"leaf_only_fields" (one per field).

### leaf_only_agent

#### Intermediate node with agent section

Input: ROOT, ROOT/a (intermediate), ROOT/a/b. ROOT/a
has node.agent present. Call SpecTreeValidate. Expect a
FormatError with rule = "leaf_only_agent".

#### Leaf node with agent section — no error

Input: ROOT, ROOT/a (leaf) with node.agent present.
Call SpecTreeValidate. Expect no leaf_only_agent error.

### dependency_targets

#### depends_on targets non-existent ROOT node

Input: ROOT, ROOT/a with depends_on = ["ROOT/missing"].
Call SpecTreeValidate. Expect a FormatError with rule =
"dependency_targets".

#### depends_on targets ancestor

Input: ROOT, ROOT/a, ROOT/a/b with depends_on =
["ROOT"]. Call SpecTreeValidate. Expect a FormatError
with rule = "dependency_targets" for ROOT/a/b.

#### depends_on targets descendant

Input: ROOT, ROOT/a with depends_on = ["ROOT/a/b"],
ROOT/a/b. Call SpecTreeValidate. Expect a FormatError
with rule = "dependency_targets" for ROOT/a.

#### depends_on targets self

Input: ROOT, ROOT/a with depends_on = ["ROOT/a"]. Call
SpecTreeValidate. Expect a FormatError with rule =
"dependency_targets".

#### depends_on with valid ROOT qualifier

Input: ROOT, ROOT/a, ROOT/b with depends_on =
["ROOT/a(interface)"]. Call SpecTreeValidate. Expect no
dependency_targets error (qualifier stripped, ROOT/a
exists and is not ancestor/descendant/self).

#### depends_on with valid ARTIFACT reference

Input: ROOT, ROOT/a with outputs = [{id: "lib",
path: "lib.go"}], ROOT/b with depends_on =
["ARTIFACT/a(lib)"]. Call SpecTreeValidate. Expect no
dependency_targets error.

#### depends_on with non-existent ARTIFACT reference

Input: ROOT, ROOT/a with depends_on =
["ARTIFACT/missing(id)"]. Call SpecTreeValidate. Expect
a FormatError with rule = "dependency_targets".

#### Multiple invalid depends_on — one error per entry

Input: ROOT, ROOT/a with depends_on = ["ROOT/missing",
"ROOT/also_missing"]. Call SpecTreeValidate. Expect two
FormatErrors with rule = "dependency_targets".

### input_target

#### Valid input reference

Input: ROOT, ROOT/a with outputs = [{id: "out",
path: "a.go"}], ROOT/b with input = "ARTIFACT/a(out)".
Call SpecTreeValidate. Expect no input_target error.

#### Input not starting with ARTIFACT/

Input: ROOT, ROOT/a with input = "ROOT/something". Call
SpecTreeValidate. Expect a FormatError with rule =
"input_target".

#### Input references non-existent artifact

Input: ROOT, ROOT/a with input = "ARTIFACT/missing(id)".
Call SpecTreeValidate. Expect a FormatError with rule =
"input_target".

### external_files

#### External file exists — no fragments

Input: ROOT, ROOT/a with external = [{path:
"some/file.txt"}]. Create the file "some/file.txt" on
disk. Call SpecTreeValidate. Expect no external_files
error.

#### External file does not exist

Input: ROOT, ROOT/a with external = [{path:
"nonexistent.txt"}]. Do not create the file. Call
SpecTreeValidate. Expect a FormatError with rule =
"external_files".

#### Fragment with valid hash

Input: ROOT, ROOT/a with external = [{path: "f.txt",
fragments: [{lines: "1-3", hash: <correct hash>}]}].
Create "f.txt" with known content. Call
SpecTreeValidate. Expect no external_files error.

#### Fragment with invalid hash

Input: ROOT, ROOT/a with external = [{path: "f.txt",
fragments: [{lines: "1-3", hash: "wrong"}]}]. Create
"f.txt" with known content. Call SpecTreeValidate.
Expect a FormatError with rule = "external_files".

#### Fragment with invalid range format

Input: ROOT, ROOT/a with external = [{path: "f.txt",
fragments: [{lines: "abc", hash: "x"}]}]. Create
"f.txt". Call SpecTreeValidate. Expect a FormatError
with rule = "external_files".

#### Fragment with start > end

Input: ROOT, ROOT/a with external = [{path: "f.txt",
fragments: [{lines: "5-3", hash: "x"}]}]. Create
"f.txt". Call SpecTreeValidate. Expect a FormatError
with rule = "external_files".

#### Fragment with start < 1

Input: ROOT, ROOT/a with external = [{path: "f.txt",
fragments: [{lines: "0-3", hash: "x"}]}]. Create
"f.txt". Call SpecTreeValidate. Expect a FormatError
with rule = "external_files".

#### Fragment out of range

Input: ROOT, ROOT/a with external = [{path: "f.txt",
fragments: [{lines: "1-100", hash: "x"}]}]. Create
"f.txt" with only 5 lines. Call SpecTreeValidate.
Expect a FormatError with rule = "external_files"
indicating fragment out of range.

### output_paths

#### Valid output path

Input: ROOT, ROOT/a with outputs = [{id: "x",
path: "internal/x.go"}]. Call SpecTreeValidate. Expect
no output_paths error.

#### Output path with traversal

Input: ROOT, ROOT/a with outputs = [{id: "x",
path: "../../etc/passwd"}]. Call SpecTreeValidate.
Expect a FormatError with rule = "output_paths".

#### Output path with backslash

Input: ROOT, ROOT/a with outputs = [{id: "x",
path: "internal\\x.go"}]. Call SpecTreeValidate.
Expect a FormatError with rule = "output_paths".

### duplicate_subsections

#### Unique subsection headings — no error

Input: ROOT, ROOT/a with node.public containing
subsections "Interface" and "Context". Call
SpecTreeValidate. Expect no duplicate_subsections error.

#### Duplicate subsection headings

Input: ROOT, ROOT/a with node.public containing two
subsections both named "Interface". Call
SpecTreeValidate. Expect one FormatError with rule =
"duplicate_subsections" (the second occurrence).

#### Three identical subsection headings

Input: ROOT, ROOT/a with node.public containing three
subsections all named "Interface". Call
SpecTreeValidate. Expect two FormatErrors with rule =
"duplicate_subsections" (second and third occurrences).

#### No public section — skip

Input: ROOT, ROOT/a with node.public absent. Call
SpecTreeValidate. Expect no duplicate_subsections error.

### Cross-cutting

#### Collects multiple errors from different rules

Input: ROOT, ROOT/a with heading mismatch, invalid
depends_on, and duplicate subsections. Call
SpecTreeValidate. Expect at least three FormatErrors
from different rules.

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
