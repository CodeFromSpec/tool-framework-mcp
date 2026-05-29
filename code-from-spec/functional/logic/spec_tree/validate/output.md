<!-- code-from-spec: ROOT/functional/logic/spec_tree/validate@ra5MzlMeelJ9zurM92WAq_ApXt0 -->

# SpecTreeValidate

## Records

```
record SpecTreeValidateInput
  logical_name: string
  frontmatter: Frontmatter
  node: Node

record FormatError
  node: string
  rule: string
  detail: string
```

## Functions

---

### function SpecTreeValidate(entries: list of SpecTreeValidateInput) -> list of FormatError

Takes the full set of discovered nodes with their parsed frontmatter and body.
Returns a list of format errors (empty if all nodes are valid).

**Step 1 — Build the known logical names set.**

  1. Create an empty set called known_names.
  2. For each entry in entries:
     a. Add entry.logical_name to known_names.
  3. For each entry in entries:
     a. If entry.frontmatter.outputs is non-empty:
        For each output in entry.frontmatter.outputs:
          - Strip the "ROOT/" prefix from entry.logical_name to get the bare path.
          - Construct artifact_name as "ARTIFACT/" + bare_path + "(" + output.id + ")".
          - Add artifact_name to known_names.

**Step 2 — Collect errors across all entries.**

  4. Create an empty list called errors.
  5. For each entry in entries:
     a. Determine has_children:
        - Set has_children to false.
        - For each other_entry in entries where other_entry.logical_name is not equal to entry.logical_name:
          - If other_entry.logical_name starts with entry.logical_name + "/":
            - Set has_children to true.
            - Stop checking.
     b. Run all validation rules (see below) for this entry.
        Append any errors produced to errors.
  6. Return errors.

---

### Rule: name_heading

  1. Apply NormalizeText to entry.node.name_section.heading — call it normalized_heading.
  2. Apply NormalizeText to entry.logical_name — call it normalized_name.
  3. If normalized_heading is not equal to normalized_name:
     - Append FormatError:
         node:   entry.logical_name
         rule:   "name_heading"
         detail: "first heading <normalized_heading> does not match logical name <normalized_name>"

---

### Rule: leaf_only_fields

  1. If has_children is true:
     a. If entry.frontmatter.depends_on is non-empty:
        - Append FormatError:
            node:   entry.logical_name
            rule:   "leaf_only_fields"
            detail: "depends_on is only permitted on leaf nodes"
     b. If entry.frontmatter.external is non-empty:
        - Append FormatError:
            node:   entry.logical_name
            rule:   "leaf_only_fields"
            detail: "external is only permitted on leaf nodes"
     c. If entry.frontmatter.input is non-empty:
        - Append FormatError:
            node:   entry.logical_name
            rule:   "leaf_only_fields"
            detail: "input is only permitted on leaf nodes"
     d. If entry.frontmatter.outputs is non-empty:
        - Append FormatError:
            node:   entry.logical_name
            rule:   "leaf_only_fields"
            detail: "outputs is only permitted on leaf nodes"

---

### Rule: leaf_only_agent

  1. If has_children is true and entry.node.agent is present:
     - Append FormatError:
         node:   entry.logical_name
         rule:   "leaf_only_agent"
         detail: "# Agent section is only permitted on leaf nodes"

---

### Rule: dependency_targets

  1. For each dep in entry.frontmatter.depends_on:

     a. If dep starts with "ROOT/":
        - Call LogicalNameStripQualifier(dep) to get bare_name.
        - If bare_name is not in known_names:
          - Append FormatError:
              node:   entry.logical_name
              rule:   "dependency_targets"
              detail: "depends_on entry <dep> references unknown node <bare_name>"
          - Continue to next dep.
        - If bare_name equals entry.logical_name:
          - Append FormatError:
              node:   entry.logical_name
              rule:   "dependency_targets"
              detail: "depends_on entry <dep> points to the node itself"
          - Continue to next dep.
        - If entry.logical_name starts with bare_name + "/":
          - Append FormatError:
              node:   entry.logical_name
              rule:   "dependency_targets"
              detail: "depends_on entry <dep> points to an ancestor"
          - Continue to next dep.
        - If bare_name starts with entry.logical_name + "/":
          - Append FormatError:
              node:   entry.logical_name
              rule:   "dependency_targets"
              detail: "depends_on entry <dep> points to a descendant"
          - Continue to next dep.

     b. If dep starts with "ARTIFACT/":
        - If dep is not in known_names:
          - Append FormatError:
              node:   entry.logical_name
              rule:   "dependency_targets"
              detail: "depends_on entry <dep> references unknown artifact"

     c. If dep does not start with "ROOT/" or "ARTIFACT/":
        - Append FormatError:
            node:   entry.logical_name
            rule:   "dependency_targets"
            detail: "depends_on entry <dep> has unrecognized prefix"

---

### Rule: input_target

  1. If entry.frontmatter.input is empty, skip this rule.
  2. If entry.frontmatter.input does not start with "ARTIFACT/":
     - Append FormatError:
         node:   entry.logical_name
         rule:   "input_target"
         detail: "input must be an ARTIFACT/ reference, got <entry.frontmatter.input>"
     - Return (do not proceed to existence check).
  3. If entry.frontmatter.input is not in known_names:
     - Append FormatError:
         node:   entry.logical_name
         rule:   "input_target"
         detail: "input references unknown artifact <entry.frontmatter.input>"

---

### Rule: external_files

  1. For each ext in entry.frontmatter.external:

     a. Create a PathCfs with value = ext.path.

     **Step 1 — Verify existence.**
     b. Call FileOpen(path_cfs).
        If FileOpen raises an error:
          - Append FormatError:
              node:   entry.logical_name
              rule:   "external_files"
              detail: "external file <ext.path> cannot be opened: <error message>"
          - Continue to next ext entry.
        Call FileClose on the reader immediately.

     **Step 2 — Verify fragments.**
     c. If ext.fragments is absent or empty, continue to next ext entry.
     d. For each fragment in ext.fragments:

        i.  Parse fragment.lines as "<start>-<end>".
            If the format is invalid (not matching "<integer>-<integer>"):
              - Append FormatError:
                  node:   entry.logical_name
                  rule:   "external_files"
                  detail: "external file <ext.path> fragment has invalid lines format: <fragment.lines>"
              - Continue to next fragment.
            Parse start and end as integers.
            If start < 1:
              - Append FormatError:
                  node:   entry.logical_name
                  rule:   "external_files"
                  detail: "external file <ext.path> fragment start line must be >= 1, got <start>"
              - Continue to next fragment.
            If start > end:
              - Append FormatError:
                  node:   entry.logical_name
                  rule:   "external_files"
                  detail: "external file <ext.path> fragment start <start> exceeds end <end>"
              - Continue to next fragment.

        ii. Call FileOpen(path_cfs).
            If FileOpen raises an error:
              - Append FormatError:
                  node:   entry.logical_name
                  rule:   "external_files"
                  detail: "external file <ext.path> cannot be opened for fragment read: <error message>"
              - Continue to next fragment.

        iii. Call FileSkipLines(reader, start - 1) to skip to the start line.

        iv. Set line_count = end - start + 1.
            Create an empty list called read_lines.
            Repeat line_count times:
              - Call FileReadLine(reader).
                If "end of file" is raised:
                  - Call FileClose(reader).
                  - Append FormatError:
                      node:   entry.logical_name
                      rule:   "external_files"
                      detail: "external file <ext.path> fragment <fragment.lines> is out of range"
                  - Continue to next fragment.
                - Append the returned line to read_lines.

        v.  Call FileClose(reader).

        vi. Join read_lines with "\n" (LF) to form content.

        vii. Compute SHA-1 of content (UTF-8 encoded).
             Encode the 20-byte digest as base64url (RFC 4648 §5, no padding) — 27 characters.
             Call this computed_hash.

        viii. If computed_hash is not equal to fragment.hash:
              - Append FormatError:
                  node:   entry.logical_name
                  rule:   "external_files"
                  detail: "external file <ext.path> fragment <fragment.lines> hash mismatch: expected <fragment.hash>, got <computed_hash>"

---

### Rule: output_paths

  1. For each output in entry.frontmatter.outputs:
     a. Call PathValidateCfs(output.path).
        If PathValidateCfs raises an error:
          - Append FormatError:
              node:   entry.logical_name
              rule:   "output_paths"
              detail: "output path <output.path> is invalid: <error message>"

---

### Rule: duplicate_subsections

  1. If entry.node.public is absent, skip this rule.
  2. If entry.node.public.subsections is empty, skip this rule.
  3. Create an empty set called seen_headings.
  4. For each subsection in entry.node.public.subsections:
     a. Apply NormalizeText to subsection.heading — call it normalized_heading.
     b. If normalized_heading is already in seen_headings:
        - Append FormatError:
            node:   entry.logical_name
            rule:   "duplicate_subsections"
            detail: "duplicate ## subsection heading \"<subsection.heading>\" in # Public section"
     c. Else:
        - Add normalized_heading to seen_headings.
