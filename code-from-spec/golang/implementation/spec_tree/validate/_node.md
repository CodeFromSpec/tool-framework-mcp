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

`import "github.com/CodeFromSpec/tool-framework-mcp/v6/internal/spectreevalidate"`

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

2. Build `known_logical_names` as an empty set of strings,
   and `known_spec_nodes` as an empty list of strings.
   For each entry in entries:
     Add entry.Reference.LogicalName to `known_logical_names`.
     Append entry.Reference.LogicalName to `known_spec_nodes`.
     If entry.Frontmatter.Type is not nil:
       Let `relative` = entry.Reference.LogicalName with
       "SPEC/" prefix stripped.
       If *entry.Frontmatter.Type is "artifact":
         Add "ARTIFACT/" + relative to
         `known_logical_names`.
       If *entry.Frontmatter.Type is "verdict":
         Add "VERDICT/" + relative to
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

### Rule: leaf_only_type (per entry)

   If `has_children` is true and entry.Frontmatter.Type
   is not nil:
     Append FormatError with rule "leaf_only_type",
     detail "type is only permitted on leaf nodes".

### Rule: type_value (per entry)

   If entry.Frontmatter.Type is not nil and
   *entry.Frontmatter.Type is not "artifact" and
   *entry.Frontmatter.Type is not "verdict":
     Append FormatError with rule "type_value",
     detail "unrecognized type value:
     <*entry.Frontmatter.Type>".

### Rule: requires_type (per entry)

   If `has_children` is false and
   entry.Frontmatter.Type is nil:
     If entry.Frontmatter.Imports is non-empty:
       Append FormatError with rule "requires_type",
       detail "imports requires type".
     If entry.Frontmatter.Input is non-empty:
       Append FormatError with rule "requires_type",
       detail "input requires type".
     If entry.Frontmatter.Output is not nil:
       Append FormatError with rule "requires_type",
       detail "output requires type".
     If entry.Frontmatter.WaitOn is non-empty:
       Append FormatError with rule "requires_type",
       detail "wait_on requires type".
     If entry.agent is present:
       Append FormatError with rule "requires_type",
       detail "# Agent section requires type".

### Rule: leaf_only_fields (per entry)

   If `has_children` is true:
     If entry.Frontmatter.Imports is non-empty:
       Append FormatError with rule "leaf_only_fields",
       detail "imports is only permitted on leaf
       nodes".
     If entry.Frontmatter.Input is non-empty:
       Append FormatError with rule "leaf_only_fields",
       detail "input is only permitted on leaf nodes".
     If entry.Frontmatter.Output is not nil:
       Append FormatError with rule "leaf_only_fields",
       detail "output is only permitted on leaf nodes".
     If entry.Frontmatter.WaitOn is non-empty:
       Append FormatError with rule "leaf_only_fields",
       detail "wait_on is only permitted on leaf nodes".

### Rule: leaf_only_agent (per entry)

   If `has_children` is true and entry.agent is
   present: Append FormatError with rule
   "leaf_only_agent", detail "# Agent section is only
   permitted on leaf nodes".

### Rule: import_targets (per entry)

   First, expand globs: initialize `expandedImports`
   as an empty list. For each dep in
   entry.Frontmatter.Imports:
     If dep ends with `/*`:
       Call parsing.ExpandGlob(dep,
       known_spec_nodes, &entry.Reference.LogicalName).
       If it fails:
         error "imports has invalid glob: <dep>:
         <error message>"
         Continue to next dep.
       Append all results to `expandedImports`.
     Else:
       Append dep to `expandedImports`.

   Then, for each dep in `expandedImports`:

     If dep starts with "SPEC/":
       Call parsing.CfsReferenceFromName(dep). If it fails:
         error "imports entry cannot be parsed: <dep>"
         Continue to next dep.
       Let `ref` be the result.
       If ref.LogicalName is not in `known_logical_names`:
         error "imports references unknown SPEC
         node: <dep>"
       Else if ref.LogicalName equals entry.Reference.LogicalName:
         error "imports must not reference the node
         itself: <dep>"
       Else if ref.LogicalName followed by "/" is a prefix of
       entry.Reference.LogicalName:
         error "imports must not reference an
         ancestor: <dep>"
       Else if entry.Reference.LogicalName followed by "/" is a
       prefix of ref.LogicalName:
         error "imports must not reference a
         descendant: <dep>"

     Else if dep starts with "ARTIFACT/":
       If dep is not in `known_logical_names`:
         error "imports references unknown
         ARTIFACT: <dep>"

     Else if dep starts with "VERDICT/":
       error "imports must not reference a
       VERDICT: <dep>"

     Else if dep starts with "EXTERNAL/":
       Let relative = dep with "EXTERNAL/" prefix
       removed.
       Let cfs_path = oslayer.CfsPath(relative).
       Attempt oslayer.OpenFile(cfs_path, "read", 30000).
       If OpenFile raises any error:
         error "imports references unreadable
         EXTERNAL file: <dep>"
       Else: Call handle.Close() on the returned handle.

     Else:
       error "imports entry has unrecognized
       prefix: <dep>"

### Rule: input_target (per entry)

   First, expand globs: initialize `expandedInput`
   as an empty list. For each inp in
   entry.Frontmatter.Input:
     If inp ends with `/*`:
       Call parsing.ExpandGlob(inp,
       known_spec_nodes, &entry.Reference.LogicalName).
       If it fails:
         error "input has invalid glob: <inp>:
         <error message>"
         Continue to next inp.
       Append all results to `expandedInput`.
     Else:
       Append inp to `expandedInput`.

   Then, for each inp in `expandedInput`:

     If inp starts with "SPEC/":
       Call parsing.CfsReferenceFromName(inp). If it fails:
         error "input entry cannot be parsed: <inp>"
         Continue to next inp.
       Let `ref` be the result.
       If ref.LogicalName is not in `known_logical_names`:
         error "input references unknown SPEC
         node: <inp>"

     Else if inp starts with "ARTIFACT/":
       If inp is not in `known_logical_names`:
         error "input references unknown ARTIFACT:
         <inp>"

     Else if inp starts with "VERDICT/":
       error "input must not reference a VERDICT:
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
       error "input entry has unrecognized prefix: <inp>"

### Rule: wait_on_targets (per entry)

   First, expand globs: initialize `expandedWaitOn`
   as an empty list. For each wo in
   entry.Frontmatter.WaitOn:
     If wo ends with `/*`:
       Call parsing.ExpandGlob(wo,
       known_spec_nodes, &entry.Reference.LogicalName).
       If it fails:
         error "wait_on has invalid glob: <wo>:
         <error message>"
         Continue to next wo.
       Append all results to `expandedWaitOn`.
     Else:
       Append wo to `expandedWaitOn`.

   Then, for each wo in `expandedWaitOn`:

     If wo starts with "ARTIFACT/":
       If wo is not in `known_logical_names`:
         error "wait_on references unknown
         ARTIFACT: <wo>"

     Else if wo starts with "VERDICT/":
       If wo is not in `known_logical_names`:
         error "wait_on references unknown
         VERDICT: <wo>"

     Else:
       error "wait_on entry has unrecognized
       prefix: <wo>"

### Rule: output_paths (per entry)

   If entry.Frontmatter.Output is not nil:
     Call oslayer.ValidateStringIsCfsPath(*entry.Frontmatter.Output).
     If ValidateStringIsCfsPath raises any error:
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
  (only for EXTERNAL existence checks), `ValidateStringIsCfsPath`,
  and `CfsPath`.
- Use the `parsing` package for `NormalizeText`,
  `CfsReferenceFromName` (only for SPEC references in
  import_targets), `NodeFrontmatter`, and `Node`.
  Use `strings.HasPrefix` for ARTIFACT/ and EXTERNAL/
  classification.
- The package name should be `spectreevalidate`.
- `FormatError` is the only exported struct in this
  package.
- The function never returns an error — all problems
  are collected as FormatError entries in the returned
  list.
