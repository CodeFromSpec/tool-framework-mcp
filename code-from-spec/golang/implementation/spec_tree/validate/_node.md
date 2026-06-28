---
depends_on:
  - SPEC/golang/implementation/os/file/impl
  - SPEC/golang/implementation/os/list_files
  - SPEC/golang/implementation/os/path_utils
  - SPEC/golang/implementation/parsing/frontmatter
  - SPEC/golang/implementation/parsing/node_parsing
  - SPEC/golang/implementation/utils/logical_names
  - SPEC/golang/implementation/utils/text_normalization
output: internal/spectreevalidate/spectreevalidate.go
---

# SPEC/golang/implementation/spec_tree/validate

Linter for the spec tree. Receives discovered nodes with
their parsed frontmatter and body, checks structural
rules, and reports all violations found.

# Public

## Package

`package spectreevalidate`

## Import

`import "github.com/CodeFromSpec/tool-framework-mcp/v5/internal/spectreevalidate"`

## Interface

```go
type SpecTreeValidateInput struct {
	LogicalName string
	Frontmatter *frontmatter.Frontmatter
	Node        *parsenode.Node
}

type FormatError struct {
	Node   string
	Rule   string
	Detail string
}

func SpecTreeValidate(entries []*SpecTreeValidateInput, allDirs []string) []FormatError
```

Takes the full set of discovered nodes with their parsed
frontmatter and body, plus a list of all subdirectory
paths found under `code-from-spec/`. Returns a list of
format errors (empty if all nodes are valid).

# Agent

Implement the spec tree validation as a Go package.

## Logic

1. Initialize `errors` as an empty list of FormatError.

2. Build `known_logical_names` as an empty set of strings.
   For each entry in entries:
     Add entry.logical_name to `known_logical_names`.
     If entry.frontmatter.output is non-empty:
       Derive the artifact logical name by stripping
       the `SPEC/` prefix from entry.logical_name and
       prepending `ARTIFACT/`.
       Add the artifact logical name to
       `known_logical_names`.

3. For each entry in entries, determine `has_children`:
   `has_children` is true if any other entry in entries
   has a logical_name that starts with this entry's
   logical_name followed by `"/"`.

4. For each entry in entries, run all validation rules
   below. Collect all errors — do not stop at the first.

### Rule: name_heading

   Normalize entry.logical_name using NormalizeText.
   Normalize entry.node.name_section.heading using
   NormalizeText. If the two normalized values are not
   equal: Append FormatError:
     node: entry.logical_name
     rule: "name_heading"
     detail: "first heading does not match the node
     logical name"

### Rule: leaf_only_fields

   If `has_children` is true:
     If entry.frontmatter.depends_on is non-empty:
       Append FormatError with rule "leaf_only_fields",
       detail "depends_on is only permitted on leaf
       nodes".
     If entry.frontmatter.input is non-empty:
       Append FormatError with rule "leaf_only_fields",
       detail "input is only permitted on leaf nodes".
     If entry.frontmatter.output is non-empty:
       Append FormatError with rule "leaf_only_fields",
       detail "output is only permitted on leaf nodes".

### Rule: leaf_only_agent

   If `has_children` is true and entry.node.agent is
   present: Append FormatError with rule
   "leaf_only_agent", detail "# Agent section is only
   permitted on leaf nodes".

### Rule: dependency_targets

   For each dep in entry.frontmatter.depends_on:

     If dep starts with "SPEC/":
       Call LogicalNameParse(dep). If it fails:
         error "depends_on entry cannot be parsed: <dep>"
         Continue to next dep.
       Let `ln` be the result.
       If ln.Name is not in `known_logical_names`:
         error "depends_on references unknown SPEC
         node: <dep>"
       Else if ln.Name equals entry.logical_name:
         error "depends_on must not reference the node
         itself: <dep>"
       Else if ln.Name followed by "/" is a prefix of
       entry.logical_name:
         error "depends_on must not reference an
         ancestor: <dep>"
       Else if entry.logical_name followed by "/" is a
       prefix of ln.Name:
         error "depends_on must not reference a
         descendant: <dep>"

     Else if dep starts with "ARTIFACT/":
       If dep is not in `known_logical_names`:
         error "depends_on references unknown
         ARTIFACT: <dep>"

     Else if dep starts with "EXTERNAL/":
       Let relative = dep with "EXTERNAL/" prefix
       removed.
       Let cfs_path = PathCfs{Value: relative}.
       Attempt FileOpen(cfs_path, "read", 30000).
       If FileOpen raises any error:
         error "depends_on references unreadable
         EXTERNAL file: <dep>"
       Else: Call FileClose on the returned handle.

     Else:
       error "depends_on entry has unrecognized
       prefix: <dep>"

### Rule: input_target

   If entry.frontmatter.input is non-empty:
     Let inp = entry.frontmatter.input.

     If inp starts with "SPEC/":
       Call LogicalNameParse(inp). If it fails:
         error "input entry cannot be parsed: <inp>"
       Else:
         Let `ln` be the result.
         If ln.Name is not in `known_logical_names`:
           error "input references unknown SPEC
           node: <inp>"

     Else if inp starts with "ARTIFACT/":
       If inp is not in `known_logical_names`:
         error "input references unknown ARTIFACT:
         <inp>"

     Else if inp starts with "EXTERNAL/":
       Let relative = inp with "EXTERNAL/" prefix
       removed.
       Let cfs_path = PathCfs{Value: relative}.
       Attempt FileOpen(cfs_path, "read", 30000).
       If FileOpen raises any error:
         error "input references unreadable EXTERNAL
         file: <inp>"
       Else: Call FileClose on the returned handle.

     Else:
       error "input must start with SPEC/, ARTIFACT/,
       or EXTERNAL/"

### Rule: missing_node_md

   For each dir in all_dirs:
     If dir equals "code-from-spec/" or dir equals
     "code-from-spec": Skip.
     Derive the first path segment after
     "code-from-spec/" in dir. If that first segment
     starts with ".": Skip.
     Let expected_node_path = dir + "/_node.md"
       (normalized to use forward slashes, no trailing
       slash on dir).
     Check whether any entry in entries has a file path
     equal to expected_node_path. If no such entry
     exists: Append FormatError with node = dir,
     rule = "missing_node_md", detail = "subdirectory
     has no _node.md".

### Rule: output_paths

   If entry.frontmatter.output is non-empty:
     Call PathValidateCfs(entry.frontmatter.output).
     If PathValidateCfs raises any error:
       Append FormatError with rule "output_paths",
       detail "output path is invalid: <error message>".

### Rule: public_subsection_required

   If entry.node.public is present:
     For each line in entry.node.public.content:
       If the line is not blank (contains at least one
       non-whitespace character):
         Append FormatError with rule
         "public_subsection_required", detail "content
         in # Public must be under a ## subsection".
         Break — report at most one error per node.

### Rule: duplicate_subsections

   If entry.node.public is present and
   entry.node.public.subsections is non-empty:
     Initialize `seen_headings` as an empty set.
     For each subsection in
     entry.node.public.subsections:
       Let normalized = NormalizeText(
       subsection.heading).
       If normalized is already in `seen_headings`:
         Append FormatError with rule
         "duplicate_subsections", detail "duplicate ##
         subsection heading in # Public:
         <subsection.raw_heading>".
       Else: Add normalized to `seen_headings`.

5. Return `errors`.

## Go-specific guidance

- Use the `file` package for `FileOpen`, `FileReadLine`,
  `FileSkipLines`, `FileClose`.
- Use the `pathutils` package for `PathValidateCfs` and
  `PathCfs`.
- Use the `textnormalization` package for `NormalizeText`.
- Use the `logicalnames` package for `LogicalNameParse`
  (only for SPEC references in dependency_targets).
  Use `strings.HasPrefix` for ARTIFACT/ and EXTERNAL/
  classification.
- Use the `frontmatter` package for the `Frontmatter`
  record.
- Use the `parsenode` package for the `Node` record.
- The package name should be `spectreevalidate`.
- `SpecTreeValidateInput` and `FormatError` are exported
  structs in this package.
- The function never returns an error — all problems
  are collected as FormatError entries in the returned
  list.
