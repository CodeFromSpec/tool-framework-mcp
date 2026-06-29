---
depends_on:
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
  - SPEC/golang/implementation/spec_tree/validate
output: internal/spectreevalidate/spectreevalidate_test.go
---

# SPEC/golang/tests/spec_tree/validate

# Agent

## Test setup guidance

`SpecTreeValidate` receives `[]parsing.Node` entries
built by the caller. Each `parsing.Node` must be
constructed to match what `parsing.ParseNode` would
produce:
- `Reference.LogicalName`: a string like
  `"SPEC/root/a"`.
- `Reference.ParentName`: pointer to parent's logical
  name, or nil for root nodes.
- `Frontmatter`: a `*parsing.NodeFrontmatter` struct
  (nil if absent).
- `NameSection.Heading`: normalized form of the logical
  name (e.g. `"spec/root/a"` for `"SPEC/root/a"`). Use
  `parsing.NormalizeText` or hardcode the lowercase
  form.
- `NameSection.RawHeading`: the original heading line
  (e.g. `"# SPEC/root/a"`).
- `NameSection.Content`: `[]string` (can be empty).
- `Public`: if present, a `*parsing.NodeSection` with
  `Heading: "public"`, `RawHeading: "# Public"`,
  `Content: []string{...}`, and `Subsections` as needed.
- `Agent`: if present, similar structure with
  `Heading: "agent"`.
- Subsection headings are also normalized (e.g.
  `"interface"` not `"Interface"`).

In v5, there is no bare "SPEC" root node. Root nodes
are direct children of code-from-spec/ (e.g.
"SPEC/root"). Tests use "SPEC/root" as the root where
a tree hierarchy is needed.

For tests that validate external files, use `testChdir`
and create files on disk.

## Test cases

### Happy path

#### Valid leaf node passes all checks

Setup:
- entries: SPEC/root (intermediate, has children
  SPEC/root/a and SPEC/root/b), SPEC/root/a (leaf,
  heading = "spec/root/a",
  depends_on = ["SPEC/root/b"],
  output = "internal/out.go"),
  SPEC/root/b (leaf, heading = "spec/root/b").
- all_dirs: ["code-from-spec", "code-from-spec/root",
  "code-from-spec/root/a", "code-from-spec/root/b"].

Actions:
1. Call SpecTreeValidate(entries, all_dirs).

Expected: No format errors.

#### Valid intermediate node passes all checks

Setup:
- entries: SPEC/root (intermediate,
  heading = "spec/root", public present with empty
  content, no frontmatter fields, no agent),
  SPEC/root/a (leaf, heading = "spec/root/a").
- all_dirs: ["code-from-spec", "code-from-spec/root",
  "code-from-spec/root/a"].

Actions:
1. Call SpecTreeValidate(entries, all_dirs).

Expected: No format errors.

#### Leaf with no frontmatter fields

Setup:
- entries: SPEC/root (heading = "spec/root"),
  SPEC/root/a (leaf, heading = "spec/root/a", empty
  frontmatter).
- all_dirs: ["code-from-spec", "code-from-spec/root",
  "code-from-spec/root/a"].

Actions:
1. Call SpecTreeValidate(entries, all_dirs).

Expected: No format errors.

### name_heading

#### Heading matches logical name

Setup:
- entries: SPEC/root (heading = "spec/root"),
  SPEC/root/a (heading = "spec/root/a").

Actions:
1. Call SpecTreeValidate.

Expected: No name_heading error.

#### Heading does not match logical name

Setup:
- entries: SPEC/root (heading = "spec/root"),
  SPEC/root/a (heading = "spec/wrong").

Actions:
1. Call SpecTreeValidate.

Expected: FormatError { Node: "SPEC/root/a",
Rule: "name_heading" }.

### leaf_only_fields

#### Intermediate node with depends_on

Setup:
- SPEC/root, SPEC/root/a (intermediate, has child
  SPEC/root/a/b,
  depends_on = ["SPEC/root/b"]),
  SPEC/root/a/b (leaf).

Expected: FormatError { Node: "SPEC/root/a",
Rule: "leaf_only_fields" }.

#### Intermediate node with output

Setup:
- SPEC/root, SPEC/root/a (intermediate, has child
  SPEC/root/a/b, output = "x.go"),
  SPEC/root/a/b (leaf).

Expected: FormatError { Node: "SPEC/root/a",
Rule: "leaf_only_fields" }.

#### Intermediate node with input

Setup:
- SPEC/root, SPEC/root/a (intermediate, has child
  SPEC/root/a/b, input = "ARTIFACT/root/c"),
  SPEC/root/a/b (leaf).

Expected: FormatError { Node: "SPEC/root/a",
Rule: "leaf_only_fields" }.

#### Intermediate node with multiple restricted fields

Setup:
- SPEC/root, SPEC/root/a (intermediate, has child
  SPEC/root/a/b,
  depends_on = ["SPEC/root/b"], output = "x.go"),
  SPEC/root/a/b (leaf).

Expected: Two FormatErrors with
Rule = "leaf_only_fields" for SPEC/root/a.

### leaf_only_agent

#### Intermediate node with agent section

Setup:
- SPEC/root, SPEC/root/a (intermediate, has child
  SPEC/root/a/b, agent present with content),
  SPEC/root/a/b (leaf).

Expected: FormatError { Node: "SPEC/root/a",
Rule: "leaf_only_agent" }.

#### Leaf node with agent section — no error

Setup:
- SPEC/root (heading = "spec/root"), SPEC/root/a
  (leaf, heading = "spec/root/a", agent present with
  content).

Expected: No leaf_only_agent error.

### dependency_targets

#### depends_on targets non-existent SPEC node

Setup:
- SPEC/root, SPEC/root/a (leaf,
  depends_on = ["SPEC/root/missing"]).

Expected: FormatError { Node: "SPEC/root/a",
Rule: "dependency_targets" }.

#### depends_on targets ancestor

Setup:
- SPEC/root (intermediate, has child SPEC/root/a),
  SPEC/root/a (leaf, depends_on = ["SPEC/root"]).

Expected: FormatError { Node: "SPEC/root/a",
Rule: "dependency_targets" }.

#### depends_on targets descendant

Setup:
- SPEC/root, SPEC/root/a (intermediate, has child
  SPEC/root/a/b,
  depends_on = ["SPEC/root/a/b"]),
  SPEC/root/a/b (leaf).

Expected: FormatError { Node: "SPEC/root/a",
Rule: "dependency_targets" }.

#### depends_on targets self

Setup:
- SPEC/root, SPEC/root/a (leaf,
  depends_on = ["SPEC/root/a"]).

Expected: FormatError { Node: "SPEC/root/a",
Rule: "dependency_targets" }.

#### depends_on with valid SPEC qualifier

Setup:
- SPEC/root, SPEC/root/a (leaf), SPEC/root/b (leaf,
  depends_on = ["SPEC/root/a(interface)"]).

Expected: No dependency_targets error.

#### depends_on with valid ARTIFACT reference

Setup:
- SPEC/root, SPEC/root/a (leaf, output = "lib.go"),
  SPEC/root/b (leaf,
  depends_on = ["ARTIFACT/root/a"]).

Expected: No dependency_targets error.

#### depends_on with non-existent ARTIFACT reference

Setup:
- SPEC/root, SPEC/root/a (leaf,
  depends_on = ["ARTIFACT/root/missing"]).

Expected: FormatError { Node: "SPEC/root/a",
Rule: "dependency_targets" }.

#### depends_on with valid EXTERNAL reference

Setup:
- SPEC/root, SPEC/root/a (leaf,
  depends_on = ["EXTERNAL/proto/api.proto"]).
- Create "proto/api.proto" on disk.

Expected: No dependency_targets error.

#### depends_on with non-existent EXTERNAL file

Setup:
- SPEC/root, SPEC/root/a (leaf,
  depends_on = ["EXTERNAL/nonexistent.txt"]).
- Do not create the file.

Expected: FormatError { Node: "SPEC/root/a",
Rule: "dependency_targets" }.

#### depends_on with unrecognized prefix

Setup:
- SPEC/root, SPEC/root/a (leaf,
  depends_on = ["UNKNOWN/something"]).

Expected: FormatError { Node: "SPEC/root/a",
Rule: "dependency_targets" }.

#### Multiple invalid depends_on — one error per entry

Setup:
- SPEC/root, SPEC/root/a (leaf,
  depends_on = ["SPEC/root/missing",
  "SPEC/root/also_missing"]).

Expected: Two FormatErrors with
Rule = "dependency_targets" for SPEC/root/a.

### input_target

#### Valid ARTIFACT input reference

Setup:
- SPEC/root, SPEC/root/a (leaf, output = "a.go"),
  SPEC/root/b (leaf, input = "ARTIFACT/root/a").

Expected: No input_target error.

#### Valid EXTERNAL input reference

Setup:
- SPEC/root, SPEC/root/a (leaf,
  input = "EXTERNAL/docs/spec.yaml").
- Create "docs/spec.yaml" on disk.

Expected: No input_target error.

#### Valid SPEC input reference

Setup:
- SPEC/root, SPEC/root/a (leaf), SPEC/root/b (leaf,
  input = "SPEC/root/a").

Expected: No input_target error.

#### Valid SPEC input with qualifier

Setup:
- SPEC/root, SPEC/root/a (leaf), SPEC/root/b (leaf,
  input = "SPEC/root/a(acceptance-tests)").

Expected: No input_target error.

#### Input with non-existent SPEC reference

Setup:
- SPEC/root, SPEC/root/a (leaf,
  input = "SPEC/root/missing").

Expected: FormatError { Node: "SPEC/root/a",
Rule: "input_target" }.

#### Input with unsupported prefix

Setup:
- SPEC/root, SPEC/root/a (leaf,
  input = "UNKNOWN/something").

Expected: FormatError { Node: "SPEC/root/a",
Rule: "input_target" }.

#### Input references non-existent artifact

Setup:
- SPEC/root, SPEC/root/a (leaf,
  input = "ARTIFACT/root/missing").

Expected: FormatError { Node: "SPEC/root/a",
Rule: "input_target" }.

#### Input references non-existent EXTERNAL file

Setup:
- SPEC/root, SPEC/root/a (leaf,
  input = "EXTERNAL/nonexistent.txt").
- Do not create the file.

Expected: FormatError { Node: "SPEC/root/a",
Rule: "input_target" }.

### missing_node_md

#### Subdirectory without _node.md

Setup:
- entries: SPEC/root (root), SPEC/root/a (leaf).
- all_dirs: ["code-from-spec", "code-from-spec/root",
  "code-from-spec/root/a", "code-from-spec/root/b"].

Expected: FormatError { Node: "code-from-spec/root/b",
Rule: "missing_node_md" }.

#### .-prefixed dir under code-from-spec — no error

Setup:
- entries: SPEC/root (root).
- all_dirs: ["code-from-spec",
  "code-from-spec/.cache"].

Expected: No missing_node_md error.

#### .-prefixed dir deeper in tree — no error

Setup:
- entries: SPEC/root (root), SPEC/root/a (leaf).
- all_dirs: ["code-from-spec", "code-from-spec/root",
  "code-from-spec/root/a",
  "code-from-spec/root/a/.internal"].

Expected: No missing_node_md error.

#### All subdirectories have _node.md — no error

Setup:
- entries: SPEC/root (root), SPEC/root/a (leaf),
  SPEC/root/b (leaf).
- all_dirs: ["code-from-spec", "code-from-spec/root",
  "code-from-spec/root/a", "code-from-spec/root/b"].

Expected: No missing_node_md error.

### output_paths

#### Valid output path

Setup:
- SPEC/root, SPEC/root/a (leaf,
  output = "internal/x.go").

Expected: No output_paths error.

#### Output path with traversal

Setup:
- SPEC/root, SPEC/root/a (leaf,
  output = "../../etc/passwd").

Expected: FormatError { Node: "SPEC/root/a",
Rule: "output_paths" }.

#### Output path with backslash

Setup:
- SPEC/root, SPEC/root/a (leaf,
  output = "internal\\x.go").

Expected: FormatError { Node: "SPEC/root/a",
Rule: "output_paths" }.

### public_subsection_required

#### Public with content before first subsection

Setup:
- SPEC/root, SPEC/root/a (leaf, public present with
  content = ["Some loose content."], subsections =
  [{heading: "interface", raw_heading: "## Interface",
  content: ["Types."]}]).

Expected: FormatError { Node: "SPEC/root/a",
Rule: "public_subsection_required", Detail: "content
in # Public must be under a ## subsection" }.

#### Public with only blank lines before subsection — no error

Setup:
- SPEC/root, SPEC/root/a (leaf, public present with
  content = ["", "  ", ""], subsections =
  [{heading: "interface", raw_heading: "## Interface",
  content: ["Types."]}]).

Expected: No public_subsection_required error.

#### Public with content and no subsections

Setup:
- SPEC/root, SPEC/root/a (leaf, public present with
  content = ["Some content."], subsections = []).

Expected: FormatError { Node: "SPEC/root/a",
Rule: "public_subsection_required" }.

#### Public with only subsections — no error

Setup:
- SPEC/root, SPEC/root/a (leaf, public present with
  content = [], subsections = [{heading: "interface",
  raw_heading: "## Interface",
  content: ["Types."]}]).

Expected: No public_subsection_required error.

#### No public section — skip

Setup:
- SPEC/root, SPEC/root/a (leaf, public absent).

Expected: No public_subsection_required error.

### duplicate_subsections

#### Unique subsection headings — no error

Setup:
- SPEC/root, SPEC/root/a (leaf, public with
  subsections [{heading: "interface"},
  {heading: "context"}]).

Expected: No duplicate_subsections error.

#### Duplicate subsection headings

Setup:
- SPEC/root, SPEC/root/a (leaf, public with
  subsections [{heading: "interface",
  content: ["First."]}, {heading: "interface",
  content: ["Second."]}]).

Expected: One FormatError with
Rule = "duplicate_subsections" for SPEC/root/a.

#### Three identical subsection headings

Setup:
- SPEC/root, SPEC/root/a (leaf, public with
  subsections [{heading: "interface"},
  {heading: "interface"}, {heading: "interface"}]).

Expected: Two FormatErrors with
Rule = "duplicate_subsections" for SPEC/root/a.

#### No public section — skip

Setup:
- SPEC/root, SPEC/root/a (leaf, public absent).

Expected: No duplicate_subsections error.

### Cross-cutting

#### Collects multiple errors from different rules

Setup:
- SPEC/root, SPEC/root/a (leaf,
  heading = "spec/wrong",
  depends_on = ["SPEC/root/missing"], public with
  duplicate subsections).

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
- Build parsing.Node records directly — no file I/O
  except for EXTERNAL reference tests.
