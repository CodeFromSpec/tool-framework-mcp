<!-- code-from-spec: SPEC/functional/logic/spec_tree/validate@BSjXyPWhMWzzaGvPfEGU2xWV2yE -->

## Namespace
    namespace: spectreevalidate


## Interface

```
record SpecTreeValidateInput
  logical_name: string
  frontmatter: frontmatter.Frontmatter
  node: parsenode.Node

record FormatError
  node: string
  rule: string
  detail: string

function SpecTreeValidate(entries: list of SpecTreeValidateInput, all_dirs: list of string) -> list of FormatError
```


## SpecTreeValidate

function SpecTreeValidate(entries, all_dirs) -> list of FormatError

  1. Initialize an empty list `errors`.

  2. Build `known_names` set:
     For each entry in entries:
       Add entry.logical_name to known_names.
       If entry.frontmatter.output is non-empty:
         Strip the "SPEC/" prefix from entry.logical_name to get the suffix.
         Construct artifact_name = "ARTIFACT/" + suffix.
         Add artifact_name to known_names.

  3. Build `has_children` mapping:
     For each entry in entries:
       Set has_children[entry.logical_name] = false.
     For each entry A in entries:
       For each entry B in entries:
         If A.logical_name is not equal to B.logical_name:
           If B.logical_name starts with A.logical_name + "/":
             Set has_children[A.logical_name] = true.

  4. For each entry in entries:
     Run all rules below, appending to errors. Do not stop at the first error.

     ### Rule: name_heading
     Compute normalized_heading = NormalizeText(entry.node.name_section.heading).
     Compute normalized_name = NormalizeText(entry.logical_name).
     If normalized_heading is not equal to normalized_name:
       Append FormatError with:
         node = entry.logical_name
         rule = "name_heading"
         detail = "first heading <entry.node.name_section.raw_heading> does not match node logical name <entry.logical_name>"

     ### Rule: leaf_only_fields
     If has_children[entry.logical_name] is true:
       If entry.frontmatter.depends_on is non-empty:
         Append FormatError with:
           node = entry.logical_name
           rule = "leaf_only_fields"
           detail = "field depends_on is only permitted on leaf nodes"
       If entry.frontmatter.input is non-empty:
         Append FormatError with:
           node = entry.logical_name
           rule = "leaf_only_fields"
           detail = "field input is only permitted on leaf nodes"
       If entry.frontmatter.output is non-empty:
         Append FormatError with:
           node = entry.logical_name
           rule = "leaf_only_fields"
           detail = "field output is only permitted on leaf nodes"

     ### Rule: leaf_only_agent
     If has_children[entry.logical_name] is true:
       If entry.node.agent is present:
         Append FormatError with:
           node = entry.logical_name
           rule = "leaf_only_agent"
           detail = "# Agent section is only permitted on leaf nodes"

     ### Rule: dependency_targets
     For each dep in entry.frontmatter.depends_on:
       If LogicalNameIsSpec(dep) is true:
         Set bare_name = LogicalNameStripQualifier(dep).
         If bare_name is not in known_names:
           Append FormatError with:
             node = entry.logical_name
             rule = "dependency_targets"
             detail = "depends_on target <dep> does not exist"
         Else if bare_name is equal to entry.logical_name:
           Append FormatError with:
             node = entry.logical_name
             rule = "dependency_targets"
             detail = "depends_on target <dep> refers to the node itself"
         Else if entry.logical_name starts with bare_name + "/":
           Append FormatError with:
             node = entry.logical_name
             rule = "dependency_targets"
             detail = "depends_on target <dep> is an ancestor of this node"
         Else if bare_name starts with entry.logical_name + "/":
           Append FormatError with:
             node = entry.logical_name
             rule = "dependency_targets"
             detail = "depends_on target <dep> is a descendant of this node"
       Else if LogicalNameIsArtifact(dep) is true:
         Set bare_ref = LogicalNameStripQualifier(dep).
         If bare_ref is not in known_names:
           Append FormatError with:
             node = entry.logical_name
             rule = "dependency_targets"
             detail = "depends_on target <dep> does not exist"
       Else if LogicalNameIsExternal(dep) is true:
         Set cfs_path = LogicalNameExternalToPath(dep).
         Attempt FileOpen(cfs_path):
           If it succeeds:
             Call FileClose on the reader immediately.
           If it raises FileUnreadable or any error:
             Append FormatError with:
               node = entry.logical_name
               rule = "dependency_targets"
               detail = "depends_on external target <dep> is not readable"
       Else:
         Append FormatError with:
           node = entry.logical_name
           rule = "dependency_targets"
           detail = "depends_on entry <dep> has an unrecognized prefix"

     ### Rule: input_target
     If entry.frontmatter.input is non-empty:
       Set inp = entry.frontmatter.input.
       If LogicalNameIsArtifact(inp) is true:
         Set bare_ref = LogicalNameStripQualifier(inp).
         If bare_ref is not in known_names:
           Append FormatError with:
             node = entry.logical_name
             rule = "input_target"
             detail = "input target <inp> does not exist"
       Else if LogicalNameIsExternal(inp) is true:
         Set cfs_path = LogicalNameExternalToPath(inp).
         Attempt FileOpen(cfs_path):
           If it succeeds:
             Call FileClose on the reader immediately.
           If it raises FileUnreadable or any error:
             Append FormatError with:
               node = entry.logical_name
               rule = "input_target"
               detail = "input external target <inp> is not readable"
       Else:
         Append FormatError with:
           node = entry.logical_name
           rule = "input_target"
           detail = "input field must start with ARTIFACT/ or EXTERNAL/"

     ### Rule: output_paths
     If entry.frontmatter.output is non-empty:
       Attempt PathValidateCfs(entry.frontmatter.output):
         If it raises any error:
           Append FormatError with:
             node = entry.logical_name
             rule = "output_paths"
             detail = "output path <entry.frontmatter.output> is invalid: <error message>"

     ### Rule: public_subsection_required
     If entry.node.public is present:
       For each line in entry.node.public.content:
         If the line is not blank (contains at least one non-whitespace character):
           Append FormatError with:
             node = entry.logical_name
             rule = "public_subsection_required"
             detail = "content in # Public must be under a ## subsection"
           Break — report only one error per node for this rule.

     ### Rule: duplicate_subsections
     If entry.node.public is present and entry.node.public.subsections is non-empty:
       Initialize empty set `seen_headings`.
       For each subsection in entry.node.public.subsections:
         Set norm = NormalizeText(subsection.heading).
         If norm is in seen_headings:
           Append FormatError with:
             node = entry.logical_name
             rule = "duplicate_subsections"
             detail = "duplicate ## subsection heading <subsection.raw_heading> in # Public"
         Else:
           Add norm to seen_headings.

  5. Run missing_node_md check (outside the per-entry loop):

     ### Rule: missing_node_md
     Build `known_paths` set:
       For each entry in entries:
         Convert entry.logical_name to its _node.md path using LogicalNameToPath.
         Add the resulting PathCfs value to known_paths.
     For each dir_path in all_dirs:
       If dir_path is equal to "code-from-spec":
         Continue to next.
       If dir_path is equal to "code-from-spec/":
         Continue to next.
       Compute relative_part = the portion of dir_path after "code-from-spec/".
       If relative_part is empty:
         Continue to next.
       Set first_segment = the first path component of relative_part
         (i.e., the substring before the first "/" or the entire string if no "/").
       If first_segment starts with "_":
         Continue to next.
       Set candidate_path = dir_path + "/_node.md".
       If candidate_path (as a PathCfs value) is not in known_paths:
         Append FormatError with:
           node = dir_path
           rule = "missing_node_md"
           detail = "subdirectory has no _node.md"

  6. Return errors.
