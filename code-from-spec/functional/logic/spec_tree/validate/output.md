<!-- code-from-spec: ROOT/functional/logic/spec_tree/validate@Kiq-tzjPdJ0lj-yJtehxYDZLfYo -->

## Namespace

    namespace: spectreevalidate

## Records

record SpecTreeValidateInput
  logical_name: string
  frontmatter: frontmatter.Frontmatter
  node: parsenode.Node

record FormatError
  node: string
  rule: string
  detail: string

## Functions

function SpecTreeValidate(entries: list of SpecTreeValidateInput) -> list of FormatError

  1. Build the known logical names set.

     For each entry in entries:
       Add entry.logical_name to the set.
       If entry.frontmatter.output is non-empty:
         Strip the "ROOT/" prefix from entry.logical_name, prepend "ARTIFACT/" to get the artifact logical name.
         Add the artifact logical name to the set.

  2. Determine which nodes have children.

     For each entry, a node has children if any other entry's logical_name starts with
     this entry's logical_name followed by "/".

  3. For each entry, run all validation rules below.
     Collect all errors — do not stop at the first.

     Rule: name_heading
       Apply NormalizeText to both node.name_section.heading and entry.logical_name.
       If they do not match, add a FormatError with:
         node: entry.logical_name
         rule: "name_heading"
         detail: describing the mismatch

     Rule: leaf_only_fields
       If the entry has children:
         If frontmatter.depends_on is non-empty, add a FormatError with rule "leaf_only_fields".
         If frontmatter.external is non-empty, add a FormatError with rule "leaf_only_fields".
         If frontmatter.input is non-empty, add a FormatError with rule "leaf_only_fields".
         If frontmatter.output is non-empty, add a FormatError with rule "leaf_only_fields".
       Report one error per non-empty field.

     Rule: leaf_only_agent
       If the entry has children and node.agent is present:
         Add a FormatError with:
           node: entry.logical_name
           rule: "leaf_only_agent"
           detail: describing that only leaf nodes may have an Agent section

     Rule: dependency_targets
       For each entry in frontmatter.depends_on:
         If the entry starts with "ROOT/":
           Call LogicalNameStripQualifier to get the bare logical name.
           If the bare logical name is not in the known logical names set:
             Add a FormatError with rule "dependency_targets" and detail describing the unknown reference.
           Else if the bare logical name equals entry.logical_name:
             Add a FormatError with rule "dependency_targets" and detail describing self-reference.
           Else if entry.logical_name starts with bare_name followed by "/":
             Add a FormatError with rule "dependency_targets" and detail describing ancestor reference.
           Else if bare_name starts with entry.logical_name followed by "/":
             Add a FormatError with rule "dependency_targets" and detail describing descendant reference.

         If the entry starts with "ARTIFACT/":
           Call LogicalNameStripQualifier to get the bare reference (defensive).
           If the bare reference is not in the known logical names set:
             Add a FormatError with rule "dependency_targets" and detail describing the unknown artifact reference.

         Report one error per invalid entry.

     Rule: input_target
       If frontmatter.input is non-empty:
         If frontmatter.input does not start with "ARTIFACT/":
           Add a FormatError with:
             node: entry.logical_name
             rule: "input_target"
             detail: describing that input must be an ARTIFACT/ reference
         Else:
           Call LogicalNameStripQualifier on frontmatter.input to get the bare reference (defensive).
           If the bare reference is not in the known logical names set:
             Add a FormatError with:
               node: entry.logical_name
               rule: "input_target"
               detail: describing the unknown artifact reference

     Rule: external_files
       For each external entry in frontmatter.external:
         Create a PathCfs with external_entry.path as its value.
         Call FileOpen with that PathCfs.
         If FileOpen fails (invalid path, file does not exist, or not readable):
           Add a FormatError with:
             node: entry.logical_name
             rule: "external_files"
             detail: describing the path and the failure reason
           Skip to the next external entry.
         If FileOpen succeeds:
           Call FileClose immediately.

     Rule: output_paths
       If frontmatter.output is non-empty:
         Call PathValidateCfs on frontmatter.output.
         If it fails:
           Add a FormatError with:
             node: entry.logical_name
             rule: "output_paths"
             detail: describing the validation failure

     Rule: duplicate_subsections
       If node.public is absent, skip this rule.
       If node.public is present and has subsections:
         Create a set of seen headings.
         For each subsection in node.public.subsections:
           Apply NormalizeText to the subsection heading.
           If the normalized heading is already in the seen set:
             Add a FormatError with:
               node: entry.logical_name
               rule: "duplicate_subsections"
               detail: describing the duplicate heading
           Else:
             Add the normalized heading to the seen set.

  4. Return the collected list of FormatError records.
     Return an empty list if no errors were found.
