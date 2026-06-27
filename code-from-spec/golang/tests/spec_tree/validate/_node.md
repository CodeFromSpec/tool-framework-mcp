---
depends_on:
  - ARTIFACT/golang/interfaces/spec_tree/validate
  - ARTIFACT/golang/interfaces/os/file
  - ARTIFACT/golang/interfaces/os/path_utils
  - ARTIFACT/golang/interfaces/utils/text_normalization
  - ARTIFACT/golang/interfaces/parsing/frontmatter
  - ARTIFACT/golang/interfaces/parsing/node_parsing
input: ARTIFACT/functional/tests/spec_tree/validate
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
