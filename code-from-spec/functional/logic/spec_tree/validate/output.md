<!-- code-from-spec: ROOT/functional/logic/spec_tree/validate@Jr15NVy-8fAe2bZG_DkSFMIpDew -->

namespace: spectreevalidate

record SpecTreeValidateInput
  logical_name: string
  frontmatter: frontmatter.Frontmatter
  node: parsenode.Node

record FormatError
  node: string
  rule: string
  detail: string

function SpecTreeValidate(entries: list of SpecTreeValidateInput, all_dirs: list of string) -> list of FormatError

  1. Initialize errors as an empty list of FormatError.

  2. Build known_names as an empty set of strings.
     For each entry in entries:
       Add entry.logical_name to known_names.
       If entry.frontmatter.output is non-empty:
         Strip the "SPEC/" prefix from entry.logical_name to get bare_path.
         Construct artifact_name = "ARTIFACT/" + bare_path.
         Add artifact_name to known_names.

  3. For each entry in entries:

     a. Determine has_children:
        Set has_children = false.
        For each other_entry in entries:
          If other_entry.logical_name starts with (entry.logical_name + "/"):
            Set has_children = true.
            Break.

     b. Run rule name_heading:
        Compute normalized_heading = NormalizeText(entry.node.name_section.heading).
        Compute normalized_name = NormalizeText(entry.logical_name).
        If normalized_heading does not equal normalized_name:
          Append FormatError(node=entry.logical_name, rule="name_heading",
            detail="name section heading does not match logical name") to errors.

     c. Run rule leaf_only_fields:
        If has_children is true:
          If entry.frontmatter.depends_on is non-empty:
            Append FormatError(node=entry.logical_name, rule="leaf_only_fields",
              detail="depends_on is only permitted on leaf nodes") to errors.
          If entry.frontmatter.input is non-empty:
            Append FormatError(node=entry.logical_name, rule="leaf_only_fields",
              detail="input is only permitted on leaf nodes") to errors.
          If entry.frontmatter.output is non-empty:
            Append FormatError(node=entry.logical_name, rule="leaf_only_fields",
              detail="output is only permitted on leaf nodes") to errors.

     d. Run rule leaf_only_agent:
        If has_children is true and entry.node.agent is present:
          Append FormatError(node=entry.logical_name, rule="leaf_only_agent",
            detail="# Agent section is only permitted on leaf nodes") to errors.

     e. Run rule dependency_targets:
        For each dep in entry.frontmatter.depends_on:
          If LogicalNameIsSpec(dep) is true:
            Set bare = LogicalNameStripQualifier(dep).
            If bare does not exist in known_names:
              Append FormatError(node=entry.logical_name, rule="dependency_targets",
                detail="depends_on target <dep> does not exist") to errors.
              Continue to next dep.
            If bare equals entry.logical_name:
              Append FormatError(node=entry.logical_name, rule="dependency_targets",
                detail="depends_on target <dep> points to the node itself") to errors.
              Continue to next dep.
            If (bare + "/") is a prefix of entry.logical_name:
              Append FormatError(node=entry.logical_name, rule="dependency_targets",
                detail="depends_on target <dep> points to an ancestor") to errors.
              Continue to next dep.
            If (entry.logical_name + "/") is a prefix of bare:
              Append FormatError(node=entry.logical_name, rule="dependency_targets",
                detail="depends_on target <dep> points to a descendant") to errors.
              Continue to next dep.
          Else if LogicalNameIsArtifact(dep) is true:
            Set bare = LogicalNameStripQualifier(dep).
            If bare does not exist in known_names:
              Append FormatError(node=entry.logical_name, rule="dependency_targets",
                detail="depends_on target <dep> does not exist") to errors.
          Else if LogicalNameIsExternal(dep) is true:
            Set ext_path_cfs = LogicalNameExternalToPath(dep).
            Attempt FileOpen(ext_path_cfs):
              If FileOpen raises FileUnreadable or any PathUtils error:
                Append FormatError(node=entry.logical_name, rule="dependency_targets",
                  detail="depends_on external target <dep> is not readable") to errors.
              Else:
                Call FileClose on the opened reader.
          Else:
            Append FormatError(node=entry.logical_name, rule="dependency_targets",
              detail="depends_on entry <dep> has an unrecognized prefix") to errors.

     f. Run rule input_target:
        If entry.frontmatter.input is non-empty:
          Set inp = entry.frontmatter.input.
          If LogicalNameIsArtifact(inp) is true:
            Set bare = LogicalNameStripQualifier(inp).
            If bare does not exist in known_names:
              Append FormatError(node=entry.logical_name, rule="input_target",
                detail="input target <inp> does not exist") to errors.
          Else if LogicalNameIsExternal(inp) is true:
            Set ext_path_cfs = LogicalNameExternalToPath(inp).
            Attempt FileOpen(ext_path_cfs):
              If FileOpen raises FileUnreadable or any PathUtils error:
                Append FormatError(node=entry.logical_name, rule="input_target",
                  detail="input external target <inp> is not readable") to errors.
              Else:
                Call FileClose on the opened reader.
          Else:
            Append FormatError(node=entry.logical_name, rule="input_target",
              detail="input must start with ARTIFACT/ or EXTERNAL/") to errors.

     g. Run rule output_paths:
        If entry.frontmatter.output is non-empty:
          Attempt PathValidateCfs(entry.frontmatter.output):
            If PathValidateCfs raises any error:
              Append FormatError(node=entry.logical_name, rule="output_paths",
                detail="output path is invalid: <error detail>") to errors.

     h. Run rule public_subsection_required:
        If entry.node.public is present:
          For each line in entry.node.public.content:
            If line is not blank (contains at least one non-whitespace character):
              Append FormatError(node=entry.logical_name, rule="public_subsection_required",
                detail="content in # Public must be under a ## subsection") to errors.
              Break.

     i. Run rule duplicate_subsections:
        If entry.node.public is present and entry.node.public.subsections is non-empty:
          Initialize seen_headings as an empty set of strings.
          For each subsection in entry.node.public.subsections:
            Set normalized = NormalizeText(subsection.heading).
            If normalized exists in seen_headings:
              Append FormatError(node=entry.logical_name, rule="duplicate_subsections",
                detail="duplicate ## subsection heading <subsection.raw_heading> in # Public") to errors.
            Else:
              Add normalized to seen_headings.

  4. Run rule missing_node_md:
     For each dir_path in all_dirs:
       If dir_path equals "code-from-spec/":
         Continue to next dir_path.
       Compute the first path segment after "code-from-spec/" from dir_path.
       If that first segment starts with "_":
         Continue to next dir_path.
       Construct expected_file = dir_path + "_node.md".
         (If dir_path does not end with "/", append "/" before "_node.md".)
       Check whether any entry in entries has a file path equal to expected_file.
         (To get an entry's file path: call LogicalNameToPath(entry.logical_name).value.)
       If no such entry exists:
         Append FormatError(node=dir_path, rule="missing_node_md",
           detail="subdirectory has no _node.md") to errors.

  5. Return errors.
