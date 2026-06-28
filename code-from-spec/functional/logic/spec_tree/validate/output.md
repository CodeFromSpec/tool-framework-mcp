<!-- code-from-spec: SPEC/functional/logic/spec_tree/validate@-v8CEBUFbx54McPfOWjkyaTvtqg -->

## Namespace
    namespace: spectreevalidate

## Records

```
record SpecTreeValidateInput
  logical_name: string
  frontmatter: frontmatter.Frontmatter
  node: parsenode.Node

record FormatError
  node: string
  rule: string
  detail: string
```

---

## function SpecTreeValidate(entries: list of SpecTreeValidateInput, all_dirs: list of string) -> list of FormatError

1. Initialize `errors` as an empty list of FormatError.

2. Build `known_logical_names` as an empty set of strings.
   For each entry in entries:
     Add entry.logical_name to `known_logical_names`.
     If entry.frontmatter.output is non-empty:
       Derive the artifact logical name by stripping the `SPEC/` prefix from entry.logical_name
         and prepending `ARTIFACT/`.
         Example: `SPEC/a/b` → `ARTIFACT/a/b`.
       Add the artifact logical name to `known_logical_names`.

3. For each entry in entries, determine `has_children`:
   `has_children` is true if any other entry in entries has a logical_name
   that starts with this entry's logical_name followed by `"/"`.

4. For each entry in entries, run all validation rules below.
   Collect all errors — do not stop at the first error within or across entries.

---

### Rule: name_heading

   Normalize entry.logical_name using NormalizeText.
   Normalize entry.node.name_section.heading using NormalizeText.
   If the two normalized values are not equal:
     Append FormatError:
       node: entry.logical_name
       rule: "name_heading"
       detail: "first heading does not match the node logical name"

---

### Rule: leaf_only_fields

   If `has_children` is true:
     If entry.frontmatter.depends_on is non-empty:
       Append FormatError:
         node: entry.logical_name
         rule: "leaf_only_fields"
         detail: "depends_on is only permitted on leaf nodes"
     If entry.frontmatter.input is non-empty:
       Append FormatError:
         node: entry.logical_name
         rule: "leaf_only_fields"
         detail: "input is only permitted on leaf nodes"
     If entry.frontmatter.output is non-empty:
       Append FormatError:
         node: entry.logical_name
         rule: "leaf_only_fields"
         detail: "output is only permitted on leaf nodes"

---

### Rule: leaf_only_agent

   If `has_children` is true and entry.node.agent is present:
     Append FormatError:
       node: entry.logical_name
       rule: "leaf_only_agent"
       detail: "# Agent section is only permitted on leaf nodes"

---

### Rule: dependency_targets

   For each dep in entry.frontmatter.depends_on:

     If LogicalNameIsSpec(dep) is true:
       Let bare = LogicalNameStripQualifier(dep).
       If bare is not in `known_logical_names`:
         Append FormatError:
           node: entry.logical_name
           rule: "dependency_targets"
           detail: "depends_on references unknown SPEC node: <dep>"
       Else if bare equals entry.logical_name:
         Append FormatError:
           node: entry.logical_name
           rule: "dependency_targets"
           detail: "depends_on must not reference the node itself: <dep>"
       Else if bare followed by "/" is a prefix of entry.logical_name:
         Append FormatError:
           node: entry.logical_name
           rule: "dependency_targets"
           detail: "depends_on must not reference an ancestor: <dep>"
       Else if entry.logical_name followed by "/" is a prefix of bare:
         Append FormatError:
           node: entry.logical_name
           rule: "dependency_targets"
           detail: "depends_on must not reference a descendant: <dep>"

     Else if LogicalNameIsArtifact(dep) is true:
       Let bare = LogicalNameStripQualifier(dep).
       If bare is not in `known_logical_names`:
         Append FormatError:
           node: entry.logical_name
           rule: "dependency_targets"
           detail: "depends_on references unknown ARTIFACT: <dep>"

     Else if LogicalNameIsExternal(dep) is true:
       Let cfs_path = LogicalNameExternalToPath(dep).
       Attempt FileOpen(cfs_path, "read", 30000).
       If FileOpen raises any error:
         Append FormatError:
           node: entry.logical_name
           rule: "dependency_targets"
           detail: "depends_on references unreadable EXTERNAL file: <dep>"
       Else:
         Call FileClose on the returned handle.

     Else:
       Append FormatError:
         node: entry.logical_name
         rule: "dependency_targets"
         detail: "depends_on entry has unrecognized prefix: <dep>"

---

### Rule: input_target

   If entry.frontmatter.input is non-empty:
     Let inp = entry.frontmatter.input.

     If LogicalNameIsArtifact(inp) is true:
       Let bare = LogicalNameStripQualifier(inp).
       If bare is not in `known_logical_names`:
         Append FormatError:
           node: entry.logical_name
           rule: "input_target"
           detail: "input references unknown ARTIFACT: <inp>"

     Else if LogicalNameIsExternal(inp) is true:
       Let cfs_path = LogicalNameExternalToPath(inp).
       Attempt FileOpen(cfs_path, "read", 30000).
       If FileOpen raises any error:
         Append FormatError:
           node: entry.logical_name
           rule: "input_target"
           detail: "input references unreadable EXTERNAL file: <inp>"
       Else:
         Call FileClose on the returned handle.

     Else:
       Append FormatError:
         node: entry.logical_name
         rule: "input_target"
         detail: "input must start with ARTIFACT/ or EXTERNAL/"

---

### Rule: missing_node_md

   For each dir in all_dirs:
     If dir equals "code-from-spec/" or dir equals "code-from-spec":
       Skip.
     Derive the first path segment after "code-from-spec/" in dir.
       (i.e., the next component following the "code-from-spec/" prefix)
     If that first segment starts with "_":
       Skip.
     Let expected_node_path = dir + "/_node.md"
       (normalized to use forward slashes, no trailing slash on dir).
     Check whether any entry in entries has a file path equal to
       expected_node_path.
       (A node's file path is derived from its logical_name using
       LogicalNameToPath, which maps `SPEC/x/y` → `code-from-spec/x/y/_node.md`.)
     If no such entry exists:
       Append FormatError:
         node: dir
         rule: "missing_node_md"
         detail: "subdirectory has no _node.md"

---

### Rule: output_paths

   If entry.frontmatter.output is non-empty:
     Call PathValidateCfs(entry.frontmatter.output).
     If PathValidateCfs raises any error:
       Append FormatError:
         node: entry.logical_name
         rule: "output_paths"
         detail: "output path is invalid: <error message from PathValidateCfs>"

---

### Rule: public_subsection_required

   If entry.node.public is present:
     For each line in entry.node.public.content:
       If the line is not blank (contains at least one non-whitespace character):
         Append FormatError:
           node: entry.logical_name
           rule: "public_subsection_required"
           detail: "content in # Public must be under a ## subsection"
         Break — report at most one error for this node on this rule.

---

### Rule: duplicate_subsections

   If entry.node.public is present and entry.node.public.subsections is non-empty:
     Initialize `seen_headings` as an empty set of strings.
     For each subsection in entry.node.public.subsections:
       Let normalized = NormalizeText(subsection.heading).
       If normalized is already in `seen_headings`:
         Append FormatError:
           node: entry.logical_name
           rule: "duplicate_subsections"
           detail: "duplicate ## subsection heading in # Public: <subsection.raw_heading>"
       Else:
         Add normalized to `seen_headings`.

---

5. Return `errors`.
