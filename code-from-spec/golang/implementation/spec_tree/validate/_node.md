---
depends_on:
  - SPEC/golang/implementation/oslayer(interface)
  - SPEC/golang/implementation/parsing(interface)
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
type FormatError struct {
	Node   string
	Rule   string
	Detail string
}

func SpecTreeValidate(entries []parsing.Node, allDirs []string) []FormatError
```

Takes the full set of discovered nodes (each containing
its reference, frontmatter, and parsed body), plus a
list of all subdirectory paths found under
`code-from-spec/`. Returns a list of format errors
(empty if all nodes are valid).

# Agent

Implement the spec tree validation as a Go package.

## Logic

1. Initialize `errors` as an empty list of FormatError.

2. Build `known_logical_names` as an empty set of strings.
   For each entry in entries:
     Add entry.Reference.LogicalName to `known_logical_names`.
     If entry.Frontmatter.Output is not nil:
       Derive the artifact logical name by stripping
       the `SPEC/` prefix from entry.Reference.LogicalName and
       prepending `ARTIFACT/`.
       Add the artifact logical name to
       `known_logical_names`.

3. For each entry in entries, determine `has_children`:
   `has_children` is true if any other entry in entries
   has a Reference.LogicalName that starts with this
   entry's Reference.LogicalName followed by `"/"`.

4. For each entry in entries, run the per-entry rules
   below. Collect all errors — do not stop at the first.

### Rule: name_heading (per entry)

   Normalize entry.Reference.LogicalName using
   parsing.NormalizeText. Normalize
   entry.name_section.heading using
   parsing.NormalizeText. If the two normalized values are not
   equal: Append FormatError:
     node: entry.Reference.LogicalName
     rule: "name_heading"
     detail: "first heading does not match the node
     logical name"

### Rule: leaf_only_fields (per entry)

   If `has_children` is true:
     If entry.Frontmatter.depends_on is non-empty:
       Append FormatError with rule "leaf_only_fields",
       detail "depends_on is only permitted on leaf
       nodes".
     If entry.Frontmatter.Input is not nil:
       Append FormatError with rule "leaf_only_fields",
       detail "input is only permitted on leaf nodes".
     If entry.Frontmatter.Output is not nil:
       Append FormatError with rule "leaf_only_fields",
       detail "output is only permitted on leaf nodes".

### Rule: leaf_only_agent (per entry)

   If `has_children` is true and entry.agent is
   present: Append FormatError with rule
   "leaf_only_agent", detail "# Agent section is only
   permitted on leaf nodes".

### Rule: dependency_targets (per entry)

   For each dep in entry.Frontmatter.depends_on:

     If dep starts with "SPEC/":
       Call parsing.CfsReferenceFromName(dep). If it fails:
         error "depends_on entry cannot be parsed: <dep>"
         Continue to next dep.
       Let `ref` be the result.
       If ref.LogicalName is not in `known_logical_names`:
         error "depends_on references unknown SPEC
         node: <dep>"
       Else if ref.LogicalName equals entry.Reference.LogicalName:
         error "depends_on must not reference the node
         itself: <dep>"
       Else if ref.LogicalName followed by "/" is a prefix of
       entry.Reference.LogicalName:
         error "depends_on must not reference an
         ancestor: <dep>"
       Else if entry.Reference.LogicalName followed by "/" is a
       prefix of ref.LogicalName:
         error "depends_on must not reference a
         descendant: <dep>"

     Else if dep starts with "ARTIFACT/":
       If dep is not in `known_logical_names`:
         error "depends_on references unknown
         ARTIFACT: <dep>"

     Else if dep starts with "EXTERNAL/":
       Let relative = dep with "EXTERNAL/" prefix
       removed.
       Let cfs_path = oslayer.CfsPath(relative).
       Attempt oslayer.OpenFile(cfs_path, "read", 30000).
       If OpenFile raises any error:
         error "depends_on references unreadable
         EXTERNAL file: <dep>"
       Else: Call handle.Close() on the returned handle.

     Else:
       error "depends_on entry has unrecognized
       prefix: <dep>"

### Rule: input_target (per entry)

   If entry.Frontmatter.Input is not nil:
     Let inp = *entry.Frontmatter.Input.

     If inp starts with "SPEC/":
       Call parsing.CfsReferenceFromName(inp). If it fails:
         error "input entry cannot be parsed: <inp>"
       Else:
         Let `ref` be the result.
         If ref.LogicalName is not in `known_logical_names`:
           error "input references unknown SPEC
           node: <inp>"

     Else if inp starts with "ARTIFACT/":
       If inp is not in `known_logical_names`:
         error "input references unknown ARTIFACT:
         <inp>"

     Else if inp starts with "EXTERNAL/":
       Let relative = inp with "EXTERNAL/" prefix
       removed.
       Let cfs_path = oslayer.CfsPath(relative).
       Attempt oslayer.OpenFile(cfs_path, "read", 30000).
       If OpenFile raises any error:
         error "input references unreadable EXTERNAL
         file: <inp>"
       Else: Call handle.Close() on the returned handle.

     Else:
       error "input must start with SPEC/, ARTIFACT/,
       or EXTERNAL/"

### Rule: output_paths (per entry)

   If entry.Frontmatter.Output is not nil:
     Call oslayer.ValidateCfsPath(*entry.Frontmatter.Output).
     If ValidateCfsPath raises any error:
       Append FormatError with rule "output_paths",
       detail "output path is invalid: <error message>".

### Rule: public_subsection_required (per entry)

   If entry.public is present:
     For each line in entry.public.content:
       If the line is not blank (contains at least one
       non-whitespace character):
         Append FormatError with rule
         "public_subsection_required", detail "content
         in # Public must be under a ## subsection".
         Break — report at most one error per node.

### Rule: duplicate_subsections (per entry)

   If entry.public is present and
   entry.public.subsections is non-empty:
     Initialize `seen_headings` as an empty set.
     For each subsection in
     entry.public.subsections:
       Let normalized = parsing.NormalizeText(
       subsection.heading).
       If normalized is already in `seen_headings`:
         Append FormatError with rule
         "duplicate_subsections", detail "duplicate ##
         subsection heading in # Public:
         <subsection.raw_heading>".
       Else: Add normalized to `seen_headings`.

5. After the per-entry loop, run the global rule:

### Rule: missing_node_md (global)

   For each dir in all_dirs:
     If dir equals "code-from-spec/" or dir equals
     "code-from-spec": Skip.
     Remove the "code-from-spec/" prefix from dir.
     Split the remainder by "/". If any segment starts
     with ".": Skip.
     Derive the expected logical name from dir: remove
     the "code-from-spec/" prefix, prepend "SPEC/".
     For example, dir "code-from-spec/root/a" yields
     "SPEC/root/a".
     Check whether any entry in entries has a
     logical_name equal to the expected logical name.
     If no such entry exists: Append FormatError with
     node = dir, rule = "missing_node_md",
     detail = "subdirectory has no _node.md".

Return `errors`.

## Go-specific guidance

- Use the `oslayer` package for `OpenFile`, `.Close()`
  (only for EXTERNAL existence checks), `ValidateCfsPath`,
  and `CfsPath`.
- Use the `parsing` package for `NormalizeText`,
  `CfsReferenceFromName` (only for SPEC references in
  dependency_targets), `NodeFrontmatter`, and `Node`.
  Use `strings.HasPrefix` for ARTIFACT/ and EXTERNAL/
  classification.
- The package name should be `spectreevalidate`.
- `FormatError` is the only exported struct in this
  package.
- The function never returns an error — all problems
  are collected as FormatError entries in the returned
  list.
