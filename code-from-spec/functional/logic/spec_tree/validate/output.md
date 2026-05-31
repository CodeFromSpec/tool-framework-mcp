<!-- code-from-spec: ROOT/functional/logic/spec_tree/validate@HgvoV8hPeySU34qX2OyVONFvd8Q -->

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

## Functions

---

function SpecTreeValidate(entries: list of SpecTreeValidateInput) -> list of FormatError

  1. Build the known logical names set (a set of strings):
     For each entry in entries:
       Add entry.logical_name to the set.
       If entry.frontmatter.outputs is non-empty:
         For each output in entry.frontmatter.outputs:
           Strip the "ROOT/" prefix from entry.logical_name to get the remainder.
           Construct the artifact name: "ARTIFACT/" + remainder + "(" + output.id + ")".
           Add the artifact name to the set.

  2. Initialize errors as an empty list of FormatError.

  3. For each entry in entries:

     a. Determine if the entry has children:
        has_children = false
        prefix = entry.logical_name + "/"
        For each other_entry in entries:
          If other_entry.logical_name starts with prefix:
            has_children = true
            break

     b. Run all validation rules on this entry. Collect all errors.

        --- Rule: name_heading ---

        normalized_heading = NormalizeText(entry.node.name_section.heading)
        normalized_name    = NormalizeText(entry.logical_name)
        If normalized_heading does not equal normalized_name:
          Append FormatError:
            node   = entry.logical_name
            rule   = "name_heading"
            detail = "first section heading <normalized_heading> does not match
                      logical name <normalized_name>"

        --- Rule: leaf_only_fields ---

        If has_children is true:
          If entry.frontmatter.depends_on is non-empty:
            Append FormatError:
              node   = entry.logical_name
              rule   = "leaf_only_fields"
              detail = "depends_on is only permitted on leaf nodes"
          If entry.frontmatter.external is non-empty:
            Append FormatError:
              node   = entry.logical_name
              rule   = "leaf_only_fields"
              detail = "external is only permitted on leaf nodes"
          If entry.frontmatter.input is non-empty:
            Append FormatError:
              node   = entry.logical_name
              rule   = "leaf_only_fields"
              detail = "input is only permitted on leaf nodes"
          If entry.frontmatter.outputs is non-empty:
            Append FormatError:
              node   = entry.logical_name
              rule   = "leaf_only_fields"
              detail = "outputs is only permitted on leaf nodes"

        --- Rule: leaf_only_agent ---

        If has_children is true and entry.node.agent is present:
          Append FormatError:
            node   = entry.logical_name
            rule   = "leaf_only_agent"
            detail = "# Agent section is only permitted on leaf nodes"

        --- Rule: dependency_targets ---

        For each dep in entry.frontmatter.depends_on:
          If dep starts with "ROOT/":
            bare = LogicalNameStripQualifier(dep)
            If bare is not in the known logical names set:
              Append FormatError:
                node   = entry.logical_name
                rule   = "dependency_targets"
                detail = "depends_on references unknown node <dep>"
            Else if bare equals entry.logical_name:
              Append FormatError:
                node   = entry.logical_name
                rule   = "dependency_targets"
                detail = "depends_on references the node itself: <dep>"
            Else if (bare + "/") is a prefix of entry.logical_name:
              Append FormatError:
                node   = entry.logical_name
                rule   = "dependency_targets"
                detail = "depends_on references an ancestor: <dep>"
            Else if (entry.logical_name + "/") is a prefix of bare:
              Append FormatError:
                node   = entry.logical_name
                rule   = "dependency_targets"
                detail = "depends_on references a descendant: <dep>"
          Else if dep starts with "ARTIFACT/":
            If dep is not in the known logical names set:
              Append FormatError:
                node   = entry.logical_name
                rule   = "dependency_targets"
                detail = "depends_on references unknown artifact <dep>"
          Else:
            Append FormatError:
              node   = entry.logical_name
              rule   = "dependency_targets"
              detail = "depends_on entry has unrecognized prefix: <dep>"

        --- Rule: input_target ---

        If entry.frontmatter.input is non-empty:
          If entry.frontmatter.input does not start with "ARTIFACT/":
            Append FormatError:
              node   = entry.logical_name
              rule   = "input_target"
              detail = "input must be an ARTIFACT/ reference, got: <entry.frontmatter.input>"
          Else:
            If entry.frontmatter.input is not in the known logical names set:
              Append FormatError:
                node   = entry.logical_name
                rule   = "input_target"
                detail = "input references unknown artifact: <entry.frontmatter.input>"

        --- Rule: external_files ---

        For each ext in entry.frontmatter.external:
          cfs_path = PathCfs with value = ext.path

          Step 1 — Verify existence:
            Call FileOpen(cfs_path).
            If FileOpen raises any error:
              Append FormatError:
                node   = entry.logical_name
                rule   = "external_files"
                detail = "external file cannot be opened: <ext.path>"
              Skip to next external entry.
            Call FileClose(reader).

          Step 2 — Verify fragments:
            If ext.fragments is present and non-empty:
              For each fragment in ext.fragments:
                Parse fragment.lines as "<start>-<end>".
                If the format is invalid:
                  Append FormatError:
                    node   = entry.logical_name
                    rule   = "external_files"
                    detail = "invalid lines format in fragment for <ext.path>: <fragment.lines>"
                  Skip to next fragment.
                Convert start and end to integers.
                If start < 1 or start > end:
                  Append FormatError:
                    node   = entry.logical_name
                    rule   = "external_files"
                    detail = "invalid line range in fragment for <ext.path>: <fragment.lines>"
                  Skip to next fragment.

                Call FileOpen(cfs_path).
                If FileOpen raises any error:
                  Append FormatError:
                    node   = entry.logical_name
                    rule   = "external_files"
                    detail = "external file cannot be opened for fragment read: <ext.path>"
                  Skip to next fragment.

                Call FileSkipLines(reader, start - 1).
                line_count = end - start + 1
                content    = ""
                read_ok    = true
                For i from 1 to line_count:
                  Call FileReadLine(reader).
                  If FileReadLine raises "end of file":
                    Call FileClose(reader).
                    Append FormatError:
                      node   = entry.logical_name
                      rule   = "external_files"
                      detail = "fragment out of range for <ext.path>: <fragment.lines>"
                    read_ok = false
                    break
                  Append the returned line + "\n" to content.
                If read_ok is false:
                  Skip to next fragment.
                Call FileClose(reader).

                Compute SHA-1 of content (UTF-8 bytes).
                Encode the 20-byte digest as base64url (RFC 4648 §5, no padding) — 27 characters.
                If the result does not equal fragment.hash:
                  Append FormatError:
                    node   = entry.logical_name
                    rule   = "external_files"
                    detail = "fragment hash mismatch for <ext.path> lines <fragment.lines>:
                              expected <fragment.hash>, got <computed_hash>"

        --- Rule: output_paths ---

        For each output in entry.frontmatter.outputs:
          Call PathValidateCfs(output.path).
          If PathValidateCfs raises any error:
            Append FormatError:
              node   = entry.logical_name
              rule   = "output_paths"
              detail = "invalid output path <output.path>: <error detail>"

        --- Rule: duplicate_subsections ---

        If entry.node.public is present:
          seen_headings = empty set of strings
          For each subsection in entry.node.public.subsections:
            normalized = NormalizeText(subsection.heading)
            If normalized is in seen_headings:
              Append FormatError:
                node   = entry.logical_name
                rule   = "duplicate_subsections"
                detail = "duplicate ## subsection heading in # Public: <subsection.raw_heading>"
            Else:
              Add normalized to seen_headings.

  4. Return errors.
```
