---
depends_on:
  - ARTIFACT/golang/interfaces/spec_tree/validate
  - ARTIFACT/golang/interfaces/os/file
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/utils/text_normalization
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/parsing/node_parsing
output: internal/spectreevalidate/spectreevalidate_test.go
---

# SPEC/golang/tests/spec_tree/validate

# Agent

## Test setup guidance

`SpecTreeValidate` receives `SpecTreeValidateInput`
entries built by the caller. Each entry has:
- `LogicalName`: a string like `"ROOT/a"`.
- `Frontmatter`: a `*frontmatter.Frontmatter` struct.
- `Node`: a `*parsenode.Node` struct.

The `Node` must be constructed to match what `NodeParse`
would produce:
- `NameSection.Heading`: normalized form of the logical
  name (e.g. `"root/a"` for `"ROOT/a"`). Use
  `textnormalization.NormalizeText` or hardcode the
  lowercase form.
- `NameSection.RawHeading`: the original heading line
  (e.g. `"# ROOT/a"`).
- `NameSection.Content`: `[]string` (can be empty).
- `Public`: if present, a `*parsenode.NodeSection` with
  `Heading: "public"`, `RawHeading: "# Public"`,
  `Content: []string{...}`, and `Subsections` as needed.
- `Agent`: if present, similar structure with
  `Heading: "agent"`.
- Subsection headings are also normalized (e.g.
  `"interface"` not `"Interface"`).

For tests that validate external files, use `testChdir`
and create files on disk.

## Test cases

### Happy path

#### Valid leaf node passes all checks

Setup:
- entries: SPEC (intermediate, has children SPEC/a and
  SPEC/b), SPEC/a (leaf, heading = "spec/a",
  depends_on = ["SPEC/b"], output = "internal/out.go"),
  SPEC/b (leaf, heading = "spec/b").
- all_dirs: ["code-from-spec", "code-from-spec/a",
  "code-from-spec/b"].

Actions:
1. Call SpecTreeValidate(entries, all_dirs).

Expected: No format errors.

#### Valid intermediate node passes all checks

Setup:
- entries: SPEC (intermediate, heading = "spec", public
  present with empty content, no frontmatter fields, no
  agent), SPEC/a (leaf, heading = "spec/a").
- all_dirs: ["code-from-spec", "code-from-spec/a"].

Actions:
1. Call SpecTreeValidate(entries, all_dirs).

Expected: No format errors.

#### Leaf with no frontmatter fields

Setup:
- entries: SPEC (heading = "spec"), SPEC/a (leaf,
  heading = "spec/a", empty frontmatter).
- all_dirs: ["code-from-spec", "code-from-spec/a"].

Actions:
1. Call SpecTreeValidate(entries, all_dirs).

Expected: No format errors.

### name_heading

#### Heading matches logical name

Setup:
- entries: SPEC (heading = "spec"), SPEC/a
  (heading = "spec/a").

Actions:
1. Call SpecTreeValidate.

Expected: No name_heading error.

#### Heading does not match logical name

Setup:
- entries: SPEC (heading = "spec"), SPEC/a
  (heading = "spec/wrong").

Actions:
1. Call SpecTreeValidate.

Expected: FormatError { Node: "SPEC/a",
Rule: "name_heading" }.

### leaf_only_fields

#### Intermediate node with depends_on

Setup:
- SPEC, SPEC/a (intermediate, has child SPEC/a/b,
  depends_on = ["SPEC/b"]), SPEC/a/b (leaf).

Expected: FormatError { Node: "SPEC/a",
Rule: "leaf_only_fields" }.

#### Intermediate node with output

Setup:
- SPEC, SPEC/a (intermediate, has child SPEC/a/b,
  output = "x.go"), SPEC/a/b (leaf).

Expected: FormatError { Node: "SPEC/a",
Rule: "leaf_only_fields" }.

#### Intermediate node with input

Setup:
- SPEC, SPEC/a (intermediate, has child SPEC/a/b,
  input = "ARTIFACT/c"), SPEC/a/b (leaf).

Expected: FormatError { Node: "SPEC/a",
Rule: "leaf_only_fields" }.

#### Intermediate node with multiple restricted fields

Setup:
- SPEC, SPEC/a (intermediate, has child SPEC/a/b,
  depends_on = ["SPEC/b"], output = "x.go"),
  SPEC/a/b (leaf).

Expected: Two FormatErrors with
Rule = "leaf_only_fields" for SPEC/a.

### leaf_only_agent

#### Intermediate node with agent section

Setup:
- SPEC, SPEC/a (intermediate, has child SPEC/a/b,
  agent present with content), SPEC/a/b (leaf).

Expected: FormatError { Node: "SPEC/a",
Rule: "leaf_only_agent" }.

#### Leaf node with agent section — no error

Setup:
- SPEC (heading = "spec"), SPEC/a (leaf,
  heading = "spec/a", agent present with content).

Expected: No leaf_only_agent error.

### dependency_targets

#### depends_on targets non-existent SPEC node

Setup:
- SPEC, SPEC/a (leaf, depends_on = ["SPEC/missing"]).

Expected: FormatError { Node: "SPEC/a",
Rule: "dependency_targets" }.

#### depends_on targets ancestor

Setup:
- SPEC, SPEC/a (intermediate, has child SPEC/a/b),
  SPEC/a/b (leaf, depends_on = ["SPEC"]).

Expected: FormatError { Node: "SPEC/a/b",
Rule: "dependency_targets" }.

#### depends_on targets descendant

Setup:
- SPEC, SPEC/a (intermediate, has child SPEC/a/b,
  depends_on = ["SPEC/a/b"]), SPEC/a/b (leaf).

Expected: FormatError { Node: "SPEC/a",
Rule: "dependency_targets" }.

#### depends_on targets self

Setup:
- SPEC, SPEC/a (leaf, depends_on = ["SPEC/a"]).

Expected: FormatError { Node: "SPEC/a",
Rule: "dependency_targets" }.

#### depends_on with valid SPEC qualifier

Setup:
- SPEC, SPEC/a (leaf), SPEC/b (leaf,
  depends_on = ["SPEC/a(interface)"]).

Expected: No dependency_targets error.

#### depends_on with valid ARTIFACT reference

Setup:
- SPEC, SPEC/a (leaf, output = "lib.go"), SPEC/b
  (leaf, depends_on = ["ARTIFACT/a"]).

Expected: No dependency_targets error.

#### depends_on with non-existent ARTIFACT reference

Setup:
- SPEC, SPEC/a (leaf,
  depends_on = ["ARTIFACT/missing"]).

Expected: FormatError { Node: "SPEC/a",
Rule: "dependency_targets" }.

#### depends_on with valid EXTERNAL reference

Setup:
- SPEC, SPEC/a (leaf,
  depends_on = ["EXTERNAL/proto/api.proto"]).
- Create "proto/api.proto" on disk.

Expected: No dependency_targets error.

#### depends_on with non-existent EXTERNAL file

Setup:
- SPEC, SPEC/a (leaf,
  depends_on = ["EXTERNAL/nonexistent.txt"]).
- Do not create the file.

Expected: FormatError { Node: "SPEC/a",
Rule: "dependency_targets" }.

#### depends_on with unrecognized prefix

Setup:
- SPEC, SPEC/a (leaf,
  depends_on = ["UNKNOWN/something"]).

Expected: FormatError { Node: "SPEC/a",
Rule: "dependency_targets" }.

#### Multiple invalid depends_on — one error per entry

Setup:
- SPEC, SPEC/a (leaf,
  depends_on = ["SPEC/missing", "SPEC/also_missing"]).

Expected: Two FormatErrors with
Rule = "dependency_targets" for SPEC/a.

### input_target

#### Valid ARTIFACT input reference

Setup:
- SPEC, SPEC/a (leaf, output = "a.go"), SPEC/b (leaf,
  input = "ARTIFACT/a").

Expected: No input_target error.

#### Valid EXTERNAL input reference

Setup:
- SPEC, SPEC/a (leaf,
  input = "EXTERNAL/docs/spec.yaml").
- Create "docs/spec.yaml" on disk.

Expected: No input_target error.

#### Input with unsupported prefix

Setup:
- SPEC, SPEC/a (leaf, input = "SPEC/something").

Expected: FormatError { Node: "SPEC/a",
Rule: "input_target" }.

#### Input references non-existent artifact

Setup:
- SPEC, SPEC/a (leaf, input = "ARTIFACT/missing").

Expected: FormatError { Node: "SPEC/a",
Rule: "input_target" }.

#### Input references non-existent EXTERNAL file

Setup:
- SPEC, SPEC/a (leaf,
  input = "EXTERNAL/nonexistent.txt").
- Do not create the file.

Expected: FormatError { Node: "SPEC/a",
Rule: "input_target" }.

### missing_node_md

#### Subdirectory without _node.md

Setup:
- entries: SPEC, SPEC/a (leaf).
- all_dirs: ["code-from-spec", "code-from-spec/a",
  "code-from-spec/b"].

Expected: FormatError { Node: "code-from-spec/b",
Rule: "missing_node_md" }.

#### _-prefixed dir under code-from-spec — no error

Setup:
- entries: SPEC.
- all_dirs: ["code-from-spec",
  "code-from-spec/_rules"].

Expected: No missing_node_md error.

#### All subdirectories have _node.md — no error

Setup:
- entries: SPEC, SPEC/a (leaf), SPEC/b (leaf).
- all_dirs: ["code-from-spec", "code-from-spec/a",
  "code-from-spec/b"].

Expected: No missing_node_md error.

### output_paths

#### Valid output path

Setup:
- SPEC, SPEC/a (leaf, output = "internal/x.go").

Expected: No output_paths error.

#### Output path with traversal

Setup:
- SPEC, SPEC/a (leaf, output = "../../etc/passwd").

Expected: FormatError { Node: "SPEC/a",
Rule: "output_paths" }.

#### Output path with backslash

Setup:
- SPEC, SPEC/a (leaf, output = "internal\\x.go").

Expected: FormatError { Node: "SPEC/a",
Rule: "output_paths" }.

### public_subsection_required

#### Public with content before first subsection

Setup:
- SPEC, SPEC/a (leaf, public present with content =
  ["Some loose content."], subsections =
  [{heading: "interface", raw_heading: "## Interface",
  content: ["Types."]}]).

Expected: FormatError { Node: "SPEC/a",
Rule: "public_subsection_required", Detail: "content
in # Public must be under a ## subsection" }.

#### Public with only blank lines before subsection — no error

Setup:
- SPEC, SPEC/a (leaf, public present with content =
  ["", "  ", ""], subsections =
  [{heading: "interface", raw_heading: "## Interface",
  content: ["Types."]}]).

Expected: No public_subsection_required error.

#### Public with content and no subsections

Setup:
- SPEC, SPEC/a (leaf, public present with content =
  ["Some content."], subsections = []).

Expected: FormatError { Node: "SPEC/a",
Rule: "public_subsection_required" }.

#### Public with only subsections — no error

Setup:
- SPEC, SPEC/a (leaf, public present with content = [],
  subsections = [{heading: "interface",
  raw_heading: "## Interface",
  content: ["Types."]}]).

Expected: No public_subsection_required error.

#### No public section — skip

Setup:
- SPEC, SPEC/a (leaf, public absent).

Expected: No public_subsection_required error.

### duplicate_subsections

#### Unique subsection headings — no error

Setup:
- SPEC, SPEC/a (leaf, public with subsections
  [{heading: "interface"}, {heading: "context"}]).

Expected: No duplicate_subsections error.

#### Duplicate subsection headings

Setup:
- SPEC, SPEC/a (leaf, public with subsections
  [{heading: "interface", content: ["First."]},
  {heading: "interface", content: ["Second."]}]).

Expected: One FormatError with
Rule = "duplicate_subsections" for SPEC/a.

#### Three identical subsection headings

Setup:
- SPEC, SPEC/a (leaf, public with subsections
  [{heading: "interface"}, {heading: "interface"},
  {heading: "interface"}]).

Expected: Two FormatErrors with
Rule = "duplicate_subsections" for SPEC/a.

#### No public section — skip

Setup:
- SPEC, SPEC/a (leaf, public absent).

Expected: No duplicate_subsections error.

### Cross-cutting

#### Collects multiple errors from different rules

Setup:
- SPEC, SPEC/a (leaf, heading = "spec/wrong",
  depends_on = ["SPEC/missing"], public with duplicate
  subsections).

Expected: At least three FormatErrors: name_heading,
dependency_targets, duplicate_subsections.

#### Empty input list

Setup:
- entries: empty list, all_dirs: [].

Expected: No format errors.

## Go-specific guidance

- The package name is `spectreevalidate_test` (external
  test package).
- Use `t.TempDir()` for isolation.
- Use `testChdir` helper for EXTERNAL file tests.
- Build SpecTreeValidateInput records directly — no
  file I/O except for EXTERNAL reference tests.
