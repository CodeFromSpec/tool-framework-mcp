<!-- code-from-spec: ROOT/functional/logic/spec_tree/validate@qAQi94ztqt3OtsuVhLj6kPPFJIw -->

function SpecTreeValidate(entries: list of SpecTreeValidateInput) -> list of FormatError

  1. Build a set `known_names` of all logical names.
     For each entry in entries:
       Add entry.logical_name to known_names.
       For each output in entry.frontmatter.outputs:
         Strip "ROOT/" from entry.logical_name to get the suffix.
         Construct artifact_name = "ARTIFACT/" + suffix + "(" + output.id + ")".
         Add artifact_name to known_names.

  2. Initialize `errors` as an empty list of FormatError.

  3. For each entry in entries:
     Determine has_children:
       Set has_children to false.
       For each other_entry in entries:
         If other_entry.logical_name starts with entry.logical_name + "/":
           Set has_children to true.
           Break.

     Run all rules below, appending to errors as needed.

     --- Rule: name_heading ---

     Compute expected = NormalizeText(entry.logical_name).
     Compute actual   = entry.node.name_section.heading.
     If actual does not equal expected:
       Append FormatError
         node:   entry.logical_name
         rule:   "name_heading"
         detail: "name section heading <actual> does not match logical name <entry.logical_name>"

     --- Rule: leaf_only_fields ---

     If has_children is true:
       If entry.frontmatter.depends_on is non-empty:
         Append FormatError
           node:   entry.logical_name
           rule:   "leaf_only_fields"
           detail: "non-leaf node has depends_on"
       If entry.frontmatter.external is non-empty:
         Append FormatError
           node:   entry.logical_name
           rule:   "leaf_only_fields"
           detail: "non-leaf node has external"
       If entry.frontmatter.input is non-empty:
         Append FormatError
           node:   entry.logical_name
           rule:   "leaf_only_fields"
           detail: "non-leaf node has input"
       If entry.frontmatter.outputs is non-empty:
         Append FormatError
           node:   entry.logical_name
           rule:   "leaf_only_fields"
           detail: "non-leaf node has outputs"

     --- Rule: leaf_only_agent ---

     If has_children is true and entry.node.agent is present:
       Append FormatError
         node:   entry.logical_name
         rule:   "leaf_only_agent"
         detail: "non-leaf node has an Agent section"

     --- Rule: dependency_targets ---

     For each dep in entry.frontmatter.depends_on:
       If dep starts with "ROOT/":
         Set bare = LogicalNameStripQualifier(dep).
         If bare does not exist in known_names:
           Append FormatError
             node:   entry.logical_name
             rule:   "dependency_targets"
             detail: "depends_on target <dep> does not exist"
           Continue to next dep.
         If bare equals entry.logical_name:
           Append FormatError
             node:   entry.logical_name
             rule:   "dependency_targets"
             detail: "depends_on target <dep> refers to the node itself"
           Continue to next dep.
         If bare + "/" is a prefix of entry.logical_name:
           Append FormatError
             node:   entry.logical_name
             rule:   "dependency_targets"
             detail: "depends_on target <dep> is an ancestor of the node"
           Continue to next dep.
         If entry.logical_name + "/" is a prefix of bare:
           Append FormatError
             node:   entry.logical_name
             rule:   "dependency_targets"
             detail: "depends_on target <dep> is a descendant of the node"
           Continue to next dep.
       Else if dep starts with "ARTIFACT/":
         If dep does not exist in known_names:
           Append FormatError
             node:   entry.logical_name
             rule:   "dependency_targets"
             detail: "depends_on target <dep> does not exist"
       Else:
         Append FormatError
           node:   entry.logical_name
           rule:   "dependency_targets"
           detail: "depends_on entry <dep> is not a ROOT/ or ARTIFACT/ reference"

     --- Rule: input_target ---

     If entry.frontmatter.input is non-empty:
       If entry.frontmatter.input does not start with "ARTIFACT/":
         Append FormatError
           node:   entry.logical_name
           rule:   "input_target"
           detail: "input <entry.frontmatter.input> is not an ARTIFACT/ reference"
       Else if entry.frontmatter.input does not exist in known_names:
         Append FormatError
           node:   entry.logical_name
           rule:   "input_target"
           detail: "input target <entry.frontmatter.input> does not exist"

     --- Rule: external_files ---

     For each ext in entry.frontmatter.external:
       Create cfs_path as PathCfs with value = ext.path.
       Call FileOpen(cfs_path).
       If FileOpen raises any error:
         Append FormatError
           node:   entry.logical_name
           rule:   "external_files"
           detail: "external file <ext.path> cannot be opened"
         Continue to next ext.
       Call FileClose(reader).

     --- Rule: output_paths ---

     For each output in entry.frontmatter.outputs:
       Call PathValidateCfs(output.path).
       If PathValidateCfs raises any error:
         Append FormatError
           node:   entry.logical_name
           rule:   "output_paths"
           detail: "output path <output.path> is invalid: <error message>"

     --- Rule: duplicate_subsections ---

     If entry.node.public is present and entry.node.public.subsections is non-empty:
       Initialize seen_headings as an empty set of strings.
       For each subsection in entry.node.public.subsections:
         Set normalized = NormalizeText(subsection.heading).
         If normalized exists in seen_headings:
           Append FormatError
             node:   entry.logical_name
             rule:   "duplicate_subsections"
             detail: "duplicate Public subsection heading <subsection.raw_heading>"
         Else:
           Add normalized to seen_headings.

  4. Return errors.
