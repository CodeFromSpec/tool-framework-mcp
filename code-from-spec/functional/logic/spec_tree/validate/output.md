<!-- code-from-spec: ROOT/functional/logic/spec_tree/validate@0tXQaFHwa74i2m2o3YWI4I1ROTU -->

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

### SpecTreeValidate(entries: list of SpecTreeValidateInput) -> list of FormatError

1. Build the known names set.
   - For each entry in entries:
     - Add entry.logical_name to the known names set.
     - If entry.frontmatter.outputs is non-empty:
       - For each output in entry.frontmatter.outputs:
         - Strip "ROOT/" from entry.logical_name to get the bare suffix.
         - Construct artifact name: "ARTIFACT/" + bare suffix + "(" + output.id + ")".
         - Add the constructed artifact name to the known names set.

2. Collect errors list (starts empty).

3. For each entry in entries:

   a. Determine if the entry has children.
      - has_children = false
      - For each other entry in entries:
        - If other.logical_name starts with entry.logical_name + "/":
          - Set has_children = true.
          - Break.

   b. Run rule: name_heading.
      - Normalize entry.logical_name using NormalizeText.
      - Compare with entry.node.name_section.heading (already normalized).
      - If they do not match:
        - Append FormatError with:
          - node: entry.logical_name
          - rule: "name_heading"
          - detail: "heading <entry.node.name_section.heading> does not match logical name <entry.logical_name>"

   c. Run rule: leaf_only_fields.
      - Only applies if has_children is true.
      - If entry.frontmatter.depends_on is non-empty:
        - Append FormatError with:
          - node: entry.logical_name
          - rule: "leaf_only_fields"
          - detail: "field depends_on is not permitted on non-leaf nodes"
      - If entry.frontmatter.external is non-empty:
        - Append FormatError with:
          - node: entry.logical_name
          - rule: "leaf_only_fields"
          - detail: "field external is not permitted on non-leaf nodes"
      - If entry.frontmatter.input is non-empty:
        - Append FormatError with:
          - node: entry.logical_name
          - rule: "leaf_only_fields"
          - detail: "field input is not permitted on non-leaf nodes"
      - If entry.frontmatter.outputs is non-empty:
        - Append FormatError with:
          - node: entry.logical_name
          - rule: "leaf_only_fields"
          - detail: "field outputs is not permitted on non-leaf nodes"

   d. Run rule: leaf_only_agent.
      - Only applies if has_children is true.
      - If entry.node.agent is present:
        - Append FormatError with:
          - node: entry.logical_name
          - rule: "leaf_only_agent"
          - detail: "# Agent section is not permitted on non-leaf nodes"

   e. Run rule: dependency_targets.
      - For each dep in entry.frontmatter.depends_on:
        - If dep starts with "ROOT/":
          - bare = LogicalNameStripQualifier(dep)
          - If bare is not in the known names set:
            - Append FormatError with:
              - node: entry.logical_name
              - rule: "dependency_targets"
              - detail: "depends_on target <dep> does not exist"
          - Else if bare equals entry.logical_name:
            - Append FormatError with:
              - node: entry.logical_name
              - rule: "dependency_targets"
              - detail: "depends_on target <dep> refers to the node itself"
          - Else if bare + "/" is a prefix of entry.logical_name:
            - Append FormatError with:
              - node: entry.logical_name
              - rule: "dependency_targets"
              - detail: "depends_on target <dep> is an ancestor of this node"
          - Else if entry.logical_name + "/" is a prefix of bare:
            - Append FormatError with:
              - node: entry.logical_name
              - rule: "dependency_targets"
              - detail: "depends_on target <dep> is a descendant of this node"
        - Else if dep starts with "ARTIFACT/":
          - If dep is not in the known names set:
            - Append FormatError with:
              - node: entry.logical_name
              - rule: "dependency_targets"
              - detail: "depends_on target <dep> does not exist"
        - Else:
          - Append FormatError with:
            - node: entry.logical_name
            - rule: "dependency_targets"
            - detail: "depends_on entry <dep> is not a valid ROOT/ or ARTIFACT/ reference"

   f. Run rule: input_target.
      - If entry.frontmatter.input is non-empty:
        - If entry.frontmatter.input does not start with "ARTIFACT/":
          - Append FormatError with:
            - node: entry.logical_name
            - rule: "input_target"
            - detail: "input <entry.frontmatter.input> must be an ARTIFACT/ reference"
        - Else if entry.frontmatter.input is not in the known names set:
          - Append FormatError with:
            - node: entry.logical_name
            - rule: "input_target"
            - detail: "input target <entry.frontmatter.input> does not exist"

   g. Run rule: external_files.
      - For each ext in entry.frontmatter.external:
        - cfs_path = PathCfs with value = ext.path

        - Step 1 — Verify existence.
          - Call FileOpen(cfs_path).
          - If FileOpen fails:
            - Append FormatError with:
              - node: entry.logical_name
              - rule: "external_files"
              - detail: "external file <ext.path> cannot be opened: <error>"
            - Continue to next external entry.
          - Call FileClose immediately.

        - Step 2 — Verify fragments.
          - If ext.fragments is present and non-empty:
            - For each fragment in ext.fragments:
              - Parse fragment.lines as "start-end" (both integers, 1-based, inclusive).
              - If the format is invalid, or start < 1, or start > end:
                - Append FormatError with:
                  - node: entry.logical_name
                  - rule: "external_files"
                  - detail: "external file <ext.path> fragment has invalid lines field: <fragment.lines>"
                - Continue to next fragment.
              - Call FileOpen(cfs_path).
              - If FileOpen fails:
                - Append FormatError with:
                  - node: entry.logical_name
                  - rule: "external_files"
                  - detail: "external file <ext.path> cannot be opened for fragment verification"
                - Continue to next fragment.
              - Call FileSkipLines(reader, start - 1).
              - Set content = empty string.
              - Set lines_to_read = end - start + 1.
              - Set read_ok = true.
              - For i from 1 to lines_to_read:
                - Call FileReadLine(reader).
                - If FileReadLine raises "end of file":
                  - Call FileClose(reader).
                  - Append FormatError with:
                    - node: entry.logical_name
                    - rule: "external_files"
                    - detail: "external file <ext.path> fragment lines <fragment.lines> is out of range"
                  - Set read_ok = false.
                  - Break.
                - Append the returned line + "\n" to content.
              - If read_ok is false:
                - Continue to next fragment.
              - Call FileClose(reader).
              - Compute SHA-1 hash of content (UTF-8 bytes).
              - Encode as base64url (RFC 4648 §5, no padding) — 27 characters.
              - If encoded hash does not equal fragment.hash:
                - Append FormatError with:
                  - node: entry.logical_name
                  - rule: "external_files"
                  - detail: "external file <ext.path> fragment lines <fragment.lines> hash mismatch: expected <fragment.hash>, got <computed hash>"

   h. Run rule: output_paths.
      - For each output in entry.frontmatter.outputs:
        - Call PathValidateCfs(output.path).
        - If PathValidateCfs raises an error:
          - Append FormatError with:
            - node: entry.logical_name
            - rule: "output_paths"
            - detail: "output path <output.path> is invalid: <error>"

   i. Run rule: duplicate_subsections.
      - If entry.node.public is absent, skip.
      - If entry.node.public.subsections is empty, skip.
      - seen_headings = empty set.
      - For each subsection in entry.node.public.subsections:
        - If subsection.heading is already in seen_headings:
          - Append FormatError with:
            - node: entry.logical_name
            - rule: "duplicate_subsections"
            - detail: "duplicate subsection heading <subsection.heading> in # Public"
        - Else:
          - Add subsection.heading to seen_headings.

4. Return the collected errors list.
```

## Error conditions

- If `entries` is empty, returns an empty list (no errors).
- If any node's name heading does not match its logical name, a `name_heading` error is reported.
- Non-leaf nodes with restricted fields produce one `leaf_only_fields` error per offending field.
- Non-leaf nodes with `# Agent` produce a `leaf_only_agent` error.
- Invalid, self-referential, ancestor, or descendant `depends_on` targets produce `dependency_targets` errors.
- `input` that is not an `ARTIFACT/` reference or does not exist produces `input_target` errors.
- Unreadable external files, bad fragment line ranges, or hash mismatches produce `external_files` errors.
- Output paths failing `PathValidateCfs` produce `output_paths` errors.
- Repeated `##` headings (normalized) within `# Public` produce `duplicate_subsections` errors.
