<!-- code-from-spec: ROOT/functional/logic/spec_tree/validate@qPy0qXmxWDr_PSVglIRsIl_Ha7w -->

# SpecTreeValidate

## Data Structures

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

---

## function SpecTreeValidate(entries: list of SpecTreeValidateInput) -> list of FormatError

### Overview

Validates all discovered nodes against format rules. Returns all
format errors found across all entries. Validation does not stop
at the first error — every rule is applied to every entry, and
all errors are collected.

### Step 1 — Build the known names set

1. Create an empty set: `known_names`.
2. For each entry in entries:
   - Add `entry.logical_name` to `known_names`.
3. For each entry in entries:
   - For each output in `entry.frontmatter.outputs`:
     - Strip `ROOT/` from `entry.logical_name` to get the bare path.
       Example: `ROOT/a/b` → `a/b`.
     - Construct artifact name: `"ARTIFACT/" + bare path + "(" + output.id + ")"`.
       Example: `ARTIFACT/a/b(foo)`.
     - Add the constructed artifact name to `known_names`.

### Step 2 — Validate each entry

4. Create an empty list: `errors`.
5. For each entry in entries:
   - Determine whether the entry has children:
     - A node has children if any other entry's `logical_name`
       starts with `entry.logical_name + "/"`.
     - Set `has_children` to true or false accordingly.
   - Run all validation rules below against this entry.
     Append any rule violations to `errors`.
6. Return `errors`.

---

## Validation Rules

### Rule: name_heading

Rule name: `"name_heading"`.

1. Apply `NormalizeText` to `node.name_section.heading` → `normalized_heading`.
2. Apply `NormalizeText` to `entry.logical_name` → `normalized_name`.
3. If `normalized_heading` does not equal `normalized_name`:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"name_heading"`
     - detail: `"name section heading \"<heading>\" does not match logical name \"<logical_name>\""`

---

### Rule: leaf_only_fields

Rule name: `"leaf_only_fields"`.

Only applies when `has_children` is true.

1. If `has_children` is false, skip this rule entirely.
2. If `entry.frontmatter.depends_on` is non-empty:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"leaf_only_fields"`
     - detail: `"depends_on is only permitted on leaf nodes"`
3. If `entry.frontmatter.external` is non-empty:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"leaf_only_fields"`
     - detail: `"external is only permitted on leaf nodes"`
4. If `entry.frontmatter.input` is non-empty:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"leaf_only_fields"`
     - detail: `"input is only permitted on leaf nodes"`
5. If `entry.frontmatter.outputs` is non-empty:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"leaf_only_fields"`
     - detail: `"outputs is only permitted on leaf nodes"`

---

### Rule: leaf_only_agent

Rule name: `"leaf_only_agent"`.

1. If `has_children` is false, skip this rule entirely.
2. If `entry.node.agent` is present:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"leaf_only_agent"`
     - detail: `"# Agent section is only permitted on leaf nodes"`

---

### Rule: dependency_targets

Rule name: `"dependency_targets"`.

For each `ref` in `entry.frontmatter.depends_on`:

  **If `ref` starts with `"ROOT/":`**

  1. Find the bare logical name:
     - If `ref` contains `"("`, take the substring before the first `"("` → `bare_name`.
     - Otherwise, `bare_name` = `ref`.
  2. If `bare_name` does not exist in `known_names`:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"dependency_targets"`
       - detail: `"depends_on refers to unknown node \"<ref>\""`
     - Continue to next entry.
  3. If `bare_name` equals `entry.logical_name`:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"dependency_targets"`
       - detail: `"depends_on refers to the node itself: \"<ref>\""`
     - Continue to next entry.
  4. If `bare_name + "/"` is a prefix of `entry.logical_name`:
     (i.e., `entry.logical_name` starts with `bare_name + "/"`)
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"dependency_targets"`
       - detail: `"depends_on refers to an ancestor node: \"<ref>\""`
     - Continue to next entry.
  5. If `entry.logical_name + "/"` is a prefix of `bare_name`:
     (i.e., `bare_name` starts with `entry.logical_name + "/"`)
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"dependency_targets"`
       - detail: `"depends_on refers to a descendant node: \"<ref>\""`
     - Continue to next entry.

  **If `ref` starts with `"ARTIFACT/":`**

  1. If `ref` does not exist in `known_names`:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"dependency_targets"`
       - detail: `"depends_on refers to unknown artifact \"<ref>\""`
     - Continue to next entry.

  **Otherwise (ref starts with neither `ROOT/` nor `ARTIFACT/`):**

  1. Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"dependency_targets"`
     - detail: `"depends_on entry has unrecognized prefix: \"<ref>\""`

---

### Rule: input_target

Rule name: `"input_target"`.

1. If `entry.frontmatter.input` is empty, skip this rule.
2. If `entry.frontmatter.input` does not start with `"ARTIFACT/"`:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"input_target"`
     - detail: `"input must start with ARTIFACT/, got \"<input>\""`
   - Stop processing this rule for this entry.
3. If `entry.frontmatter.input` does not exist in `known_names`:
   - Append a FormatError:
     - node: `entry.logical_name`
     - rule: `"input_target"`
     - detail: `"input refers to unknown artifact \"<input>\""`

---

### Rule: external_files

Rule name: `"external_files"`.

For each `ext` in `entry.frontmatter.external`:

  **Step 1 — Verify existence.**

  1. Create a PathCfs with `value` = `ext.path`.
  2. Call `FileOpen(path_cfs)`.
     - If it raises an error (invalid path, file does not exist, not readable):
       - Append a FormatError:
         - node: `entry.logical_name`
         - rule: `"external_files"`
         - detail: `"external file cannot be opened: \"<ext.path>\""`
       - Skip to the next external entry.
     - If it succeeds, call `FileClose(reader)` immediately.

  **Step 2 — Verify fragments.**

  3. If `ext.fragments` is empty or absent, continue to the next external entry.
  4. For each `fragment` in `ext.fragments`:

     a. Parse `fragment.lines` as `"<start>-<end>"`.
        - If the format is invalid, or `start < 1`, or `start > end`:
          - Append a FormatError:
            - node: `entry.logical_name`
            - rule: `"external_files"`
            - detail: `"fragment has invalid lines range \"<fragment.lines>\" in \"<ext.path>\""`
          - Skip to the next fragment.

     b. Call `FileOpen(path_cfs)`.
        - If it raises an error:
          - Append a FormatError:
            - node: `entry.logical_name`
            - rule: `"external_files"`
            - detail: `"external file cannot be opened for fragment read: \"<ext.path>\""`
          - Skip to the next fragment.

     c. Call `FileSkipLines(reader, start - 1)` to skip the first `start - 1` lines.

     d. Set `line_count` = `end - start + 1`.
        Set `read_lines` = empty list.
        Repeat `line_count` times:
        - Call `FileReadLine(reader)`.
          - If it raises "end of file":
            - Call `FileClose(reader)`.
            - Append a FormatError:
              - node: `entry.logical_name`
              - rule: `"external_files"`
              - detail: `"fragment out of range: lines <start>-<end> in \"<ext.path>\""`
            - Skip to the next fragment.
          - Otherwise append the returned line to `read_lines`.

     e. Call `FileClose(reader)`.

     f. Join `read_lines` with `"\n"` (LF) → `content`.

     g. Compute SHA-1 of `content` (UTF-8 encoded).
        Encode the 20-byte digest as base64url (RFC 4648 §5, no padding) → `computed_hash`.

     h. If `computed_hash` does not equal `fragment.hash`:
        - Append a FormatError:
          - node: `entry.logical_name`
          - rule: `"external_files"`
          - detail: `"fragment hash mismatch for lines <start>-<end> in \"<ext.path>\": expected \"<fragment.hash>\", got \"<computed_hash>\""`

---

### Rule: output_paths

Rule name: `"output_paths"`.

For each `output` in `entry.frontmatter.outputs`:

  1. Call `PathValidateCfs(output.path)`.
     - If it raises an error:
       - Append a FormatError:
         - node: `entry.logical_name`
         - rule: `"output_paths"`
         - detail: `"output path \"<output.path>\" is invalid: <validation error message>"`

---

### Rule: duplicate_subsections

Rule name: `"duplicate_subsections"`.

1. If `entry.node.public` is absent, skip this rule.
2. If `entry.node.public.subsections` is empty, skip this rule.
3. Create an empty set: `seen_headings`.
4. For each `subsection` in `entry.node.public.subsections`:
   - Apply `NormalizeText` to `subsection.heading` → `normalized`.
   - If `normalized` is already in `seen_headings`:
     - Append a FormatError:
       - node: `entry.logical_name`
       - rule: `"duplicate_subsections"`
       - detail: `"duplicate public subsection heading \"<subsection.heading>\""`
   - Otherwise add `normalized` to `seen_headings`.
