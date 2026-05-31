<!-- code-from-spec: ROOT/functional/logic/spec_tree/validate@gnJ1WFNFg2b24lrSLtgQ3G-lmYs -->

# SpecTreeValidate

## Records

record SpecTreeValidateInput
  logical_name: string
  frontmatter: Frontmatter
  node: Node

record FormatError
  node: string
  rule: string
  detail: string


## Functions

---

function SpecTreeValidate(entries: list of SpecTreeValidateInput) -> list of FormatError

  1. Build the known logical names set.
     For each entry in entries:
       - Add entry.logical_name to the set.
       - If entry.frontmatter.outputs is non-empty:
           For each output in entry.frontmatter.outputs:
             - Strip "ROOT/" from entry.logical_name to get the suffix.
             - Construct artifact name: "ARTIFACT/" + suffix + "(" + output.id + ")".
             - Add this artifact name to the set.

  2. Initialize errors as an empty list.

  3. For each entry in entries:

     a. Determine whether entry has children:
        A node has children if any other entry in entries has a logical_name
        that starts with entry.logical_name followed by "/".

     b. Run all validation rules below, collecting errors into the errors list.
        Do not stop at the first error — run all rules for all entries.

  4. Return errors.

---

### Rule: name_heading

function ValidateNameHeading(entry: SpecTreeValidateInput, errors: list of FormatError)

  1. Compute normalized_heading = NormalizeText(entry.node.name_section.heading).
  2. Compute normalized_name   = NormalizeText(entry.logical_name).
  3. If normalized_heading does not equal normalized_name:
       Append FormatError:
         node:   entry.logical_name
         rule:   "name_heading"
         detail: "name section heading <normalized_heading> does not match logical name <normalized_name>"

---

### Rule: leaf_only_fields

function ValidateLeafOnlyFields(entry: SpecTreeValidateInput, has_children: boolean, errors: list of FormatError)

  1. If has_children is false, return immediately (leaf node — no violation possible).

  2. If entry.frontmatter.depends_on is non-empty:
       Append FormatError:
         node:   entry.logical_name
         rule:   "leaf_only_fields"
         detail: "non-leaf node has depends_on"

  3. If entry.frontmatter.external is non-empty:
       Append FormatError:
         node:   entry.logical_name
         rule:   "leaf_only_fields"
         detail: "non-leaf node has external"

  4. If entry.frontmatter.input is non-empty:
       Append FormatError:
         node:   entry.logical_name
         rule:   "leaf_only_fields"
         detail: "non-leaf node has input"

  5. If entry.frontmatter.outputs is non-empty:
       Append FormatError:
         node:   entry.logical_name
         rule:   "leaf_only_fields"
         detail: "non-leaf node has outputs"

---

### Rule: leaf_only_agent

function ValidateLeafOnlyAgent(entry: SpecTreeValidateInput, has_children: boolean, errors: list of FormatError)

  1. If has_children is false, return immediately.

  2. If entry.node.agent is present:
       Append FormatError:
         node:   entry.logical_name
         rule:   "leaf_only_agent"
         detail: "non-leaf node has an Agent section"

---

### Rule: dependency_targets

function ValidateDependencyTargets(entry: SpecTreeValidateInput, known_names: set of string, errors: list of FormatError)

  1. For each dep in entry.frontmatter.depends_on:

     a. If dep starts with "ROOT/":
          i.  bare_name = LogicalNameStripQualifier(dep).
          ii. If bare_name is not in known_names:
                Append FormatError:
                  node:   entry.logical_name
                  rule:   "dependency_targets"
                  detail: "depends_on target <dep> does not exist"
                Continue to next dep.
          iii. If bare_name equals entry.logical_name:
                 Append FormatError:
                   node:   entry.logical_name
                   rule:   "dependency_targets"
                   detail: "depends_on target <dep> points to the node itself"
                 Continue to next dep.
          iv.  If entry.logical_name starts with bare_name + "/":
                 Append FormatError:
                   node:   entry.logical_name
                   rule:   "dependency_targets"
                   detail: "depends_on target <dep> is an ancestor of this node"
                 Continue to next dep.
          v.   If bare_name starts with entry.logical_name + "/":
                 Append FormatError:
                   node:   entry.logical_name
                   rule:   "dependency_targets"
                   detail: "depends_on target <dep> is a descendant of this node"
                 Continue to next dep.

     b. Else if dep starts with "ARTIFACT/":
          i.  If dep is not in known_names:
                Append FormatError:
                  node:   entry.logical_name
                  rule:   "dependency_targets"
                  detail: "depends_on target <dep> does not exist"

     c. Else:
          Append FormatError:
            node:   entry.logical_name
            rule:   "dependency_targets"
            detail: "depends_on entry <dep> has unrecognized prefix (expected ROOT/ or ARTIFACT/)"

---

### Rule: input_target

function ValidateInputTarget(entry: SpecTreeValidateInput, known_names: set of string, errors: list of FormatError)

  1. If entry.frontmatter.input is empty, return immediately.

  2. If entry.frontmatter.input does not start with "ARTIFACT/":
       Append FormatError:
         node:   entry.logical_name
         rule:   "input_target"
         detail: "input <entry.frontmatter.input> must start with ARTIFACT/"
       Return.

  3. If entry.frontmatter.input is not in known_names:
       Append FormatError:
         node:   entry.logical_name
         rule:   "input_target"
         detail: "input target <entry.frontmatter.input> does not exist"

---

### Rule: external_files

function ValidateExternalFiles(entry: SpecTreeValidateInput, errors: list of FormatError)

  1. For each ext in entry.frontmatter.external:

     a. Create cfs_path as PathCfs with value = ext.path.

     b. Step 1 — Verify existence:
          Call FileOpen(cfs_path).
          If FileOpen fails (any error):
            Append FormatError:
              node:   entry.logical_name
              rule:   "external_files"
              detail: "external file <ext.path> cannot be opened: <error>"
            Continue to next ext entry (skip Step 2).
          Call FileClose on the opened reader.

     c. Step 2 — Verify fragments:
          If ext.fragments is absent or empty, continue to next ext entry.

          For each fragment in ext.fragments:

            i.  Parse fragment.lines as "start-end":
                  Split on "-" to get start_str and end_str.
                  Parse both as integers.
                  If parsing fails, or start < 1, or start > end:
                    Append FormatError:
                      node:   entry.logical_name
                      rule:   "external_files"
                      detail: "external file <ext.path> fragment has invalid lines range <fragment.lines>"
                    Continue to next fragment.

            ii. Open the file again with FileOpen(cfs_path).
                  If FileOpen fails:
                    Append FormatError:
                      node:   entry.logical_name
                      rule:   "external_files"
                      detail: "external file <ext.path> cannot be opened for fragment verification: <error>"
                    Continue to next fragment.

            iii. Call FileSkipLines(reader, start - 1) to skip the first start-1 lines.

            iv. Read end - start + 1 lines using FileReadLine, appending "\n" after each line.
                  If FileReadLine raises "end of file" before all lines are read:
                    Call FileClose(reader).
                    Append FormatError:
                      node:   entry.logical_name
                      rule:   "external_files"
                      detail: "external file <ext.path> fragment lines <fragment.lines> out of range"
                    Continue to next fragment.

            v.  Call FileClose(reader).

            vi. Join all read lines (each already has "\n" appended) into a single content string.

            vii. Compute SHA-1 of the content string (UTF-8 encoded bytes).

            viii. Encode the 20-byte SHA-1 digest as base64url (RFC 4648 §5, no padding) → 27 characters.

            ix. If computed hash does not equal fragment.hash:
                  Append FormatError:
                    node:   entry.logical_name
                    rule:   "external_files"
                    detail: "external file <ext.path> fragment <fragment.lines> hash mismatch: expected <fragment.hash>, got <computed>"

---

### Rule: output_paths

function ValidateOutputPaths(entry: SpecTreeValidateInput, errors: list of FormatError)

  1. For each output in entry.frontmatter.outputs:
       Call PathValidateCfs(output.path).
       If PathValidateCfs raises any error:
         Append FormatError:
           node:   entry.logical_name
           rule:   "output_paths"
           detail: "output path <output.path> is invalid: <error>"

---

### Rule: duplicate_subsections

function ValidateDuplicateSubsections(entry: SpecTreeValidateInput, errors: list of FormatError)

  1. If entry.node.public is absent, return immediately.

  2. If entry.node.public.subsections is empty, return immediately.

  3. Initialize seen_headings as an empty set of strings.

  4. For each subsection in entry.node.public.subsections:
       normalized = NormalizeText(subsection.heading).
       If normalized is already in seen_headings:
         Append FormatError:
           node:   entry.logical_name
           rule:   "duplicate_subsections"
           detail: "duplicate public subsection heading <subsection.raw_heading>"
       Else:
         Add normalized to seen_headings.
